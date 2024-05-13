package internal

import (
	"bytes"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/jtanza/post-pigeon/internal/model"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/acme/autocert"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
)

const maxFileSize = 15000

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

type Router struct {
	db          DB
	postManager PostManager
}

func NewRouter(db DB, postCreator PostManager) Router {
	return Router{db, postCreator}
}

func (r Router) Engine(logFile *os.File) *echo.Echo {
	e := echo.New()

	if strings.EqualFold(os.Getenv("POST_PIGEON_ENV"), "prod") {
		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist("post-pigeon.com", "www.post-pigeon.com")
		e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
		e.Pre(middleware.HTTPSRedirect())
	}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Output: logFile}))
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	e.Validator = &CustomValidator{validator: validator.New()}

	e.HTTPErrorHandler = customHTTPErrorHandler

	e.Static("/public", "./public")
	e.File("/", "public/index.html")
	e.File("/new", "public/new.html")
	e.File("/delete", "public/delete.html")
	e.File("/search/users", "public/user.html")

	e.GET("/posts/:uuid", r.getPost)
	e.POST("/posts", r.createPost)
	e.DELETE("/posts", r.deletePost)

	e.POST("/users", r.getUserFingerprint)
	e.GET("/users/:fingerprint", r.getUserPosts)

	return e
}

func (r Router) getPost(c echo.Context) error {
	id := c.Param("uuid")

	postContent, err := r.postManager.FetchPostContent(id)
	if err != nil {
		return err
	}
	if postContent == nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	return c.HTML(http.StatusOK, postContent.HTML)
}

func (r Router) createPost(c echo.Context) error {
	var request model.PostRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "One or more fields missing or incorrect")
	}

	body, err := readFile(c)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Empty body on request")
	}
	request.Body = body

	if dupe, err := r.postManager.IsDuplicate(request); err != nil {
		return err
	} else if dupe {
		return echo.NewHTTPError(http.StatusBadRequest, "Duplicate posts (same author, same title) are not allowed.")
	}

	uuid, err := r.postManager.CreatePost(request)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("posts/%s", uuid))
}

func (r Router) deletePost(c echo.Context) error {
	var request model.PostDeleteRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "One or more fields missing or incorrect")
	}

	if err := r.postManager.RemovePost(request); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/new")
}

func (r Router) getUserPosts(c echo.Context) error {
	id := c.Param("fingerprint")

	if exists, err := r.postManager.HasPosts(id); err != nil {
		return err
	} else if !exists {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	posts, err := r.postManager.GetAllUserPosts(id)
	if err != nil {
		return err
	}

	return c.HTML(http.StatusOK, posts)
}

func (r Router) getUserFingerprint(c echo.Context) error {
	var request model.UserRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "One or more fields missing or incorrect")
	}

	fingerprint, err := Fingerprint(request.PublicKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	if exists, err := r.postManager.HasPosts(fingerprint); err != nil {
		return err
	} else if !exists {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	posts, err := r.postManager.GetAllUserPosts(fingerprint)
	if err != nil {
		return err
	}

	return c.HTML(http.StatusOK, posts)
}

func customHTTPErrorHandler(e error, c echo.Context) {
	errorMessage := e.Error()
	code := http.StatusInternalServerError
	if he, ok := e.(*echo.HTTPError); ok {
		if he.Code == http.StatusNotFound {
			errorMessage = "What you are looking for does not exist"
		} else {
			errorMessage = fmt.Sprintf("%s", he.Message)
		}
		code = he.Code
	}
	c.Logger().Warn(e)

	h, err := errorHTML(errorMessage, code)
	if err != nil {
		c.Logger().Error(err)
	}

	if err = c.HTML(code, h); err != nil {
		c.Logger().Error(err)
	}
}

func errorHTML(e string, code int) (string, error) {
	t, err := template.New("error").ParseFiles("templates/error")
	if err != nil {
		return "", err
	}

	m := map[string]interface{}{
		"Error":  e,
		"Status": code,
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, m); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func readFile(c echo.Context) (string, error) {
	file, err := c.FormFile("body")
	if err != nil {
		return "", err
	}

	if file.Size >= maxFileSize {
		return "", fmt.Errorf("file size exceeds limit of %d bytes", maxFileSize)
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	b, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

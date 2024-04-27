package internal

import (
	"bytes"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/jtanza/post-pigeon/internal/model"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"html/template"
	"io"
	"net/http"
)

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

func (r Router) Engine() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	e.Validator = &CustomValidator{validator: validator.New()}

	e.HTTPErrorHandler = customHTTPErrorHandler

	e.Static("public", "./public")
	e.File("/", "public/index.html")
	e.File("/new", "public/new.html")
	e.File("/delete", "public/delete.html")

	e.GET("/posts/:uuid", r.getPost)
	e.POST("/posts", r.createPost)
	e.DELETE("/posts", r.deletePost)

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
		return err
	}

	body, err := readFile(c)
	if err != nil {
		return err
	}
	request.Body = body

	if dupe, err := r.postManager.IsDuplicate(request); err != nil {
		return err
	} else if dupe {
		return echo.NewHTTPError(http.StatusBadRequest, "Duplicate posts (same author, same content) are not allowed.")
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
		return err
	}

	if err := r.postManager.RemovePost(request); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/new")
}

func (r Router) getUserPosts(c echo.Context) error {
	id := c.Param("fingerprint")

	exists, err := r.postManager.HasPosts(id)
	if err != nil {
		return err
	}
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	posts, err := r.postManager.GetAllUserPosts(id)
	if err != nil {
		return err
	}

	return c.HTML(http.StatusOK, posts)
}

func customHTTPErrorHandler(e error, c echo.Context) {
	errorMessage := e.Error()
	code := http.StatusInternalServerError
	if he, ok := e.(*echo.HTTPError); ok {
		// TODO case for others
		if he.Code == http.StatusNotFound {
			errorMessage = "The page you are looking for does not exist"
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

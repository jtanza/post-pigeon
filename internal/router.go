package internal

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jtanza/post-pigeon/internal/model"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	e.Validator = &CustomValidator{validator: validator.New()}

	e.HTTPErrorHandler = customHTTPErrorHandler

	e.File("/new", "public/new.html")

	e.GET("/posts/:uuid", r.getPost)
	e.POST("/posts", r.createPost)
	e.DELETE("/posts/:uuid", r.deletePost)

	return e
}

func (r Router) getPost(c echo.Context) error {
	id := c.Param("uuid")
	postContent, err := r.postManager.FetchPostContent(id)
	if err != nil {
		return err
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

	uuid, err := r.postManager.CreatePost(request)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("posts/%s", uuid))
}

func (r Router) deletePost(c echo.Context) error {
	id := c.Param("uuid")
	if err := r.postManager.RemovePost(id); err != nil {
		return err
	}

	return c.JSON(http.StatusAccepted, id)
}

func customHTTPErrorHandler(e error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := e.(*echo.HTTPError); ok {
		code = he.Code
	}
	c.Logger().Error(e)

	h, err := errorHTML(e, code)
	if err != nil {
		c.Logger().Error(err)
	}

	if err = c.HTML(code, h); err != nil {
		c.Logger().Error(err)
	}
}

func errorHTML(e error, code int) (string, error) {
	t, err := template.New("error").ParseFiles("templates/error")
	if err != nil {
		return "", err
	}

	m := map[string]interface{}{
		"Error":  e.Error(),
		"Status": code,
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, m); err != nil {
		return "", err
	}

	return buf.String(), nil
}

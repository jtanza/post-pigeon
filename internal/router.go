package internal

import (
	"bytes"
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
	postCreator PostManager
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

	e.POST("/posts", r.createPost)
	e.DELETE("/posts/:uuid", r.deletePost)

	return e
}

func (r Router) createPost(c echo.Context) error {
	var request model.PostRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(request); err != nil {
		return err
	}

	uuid, err := r.postCreator.CreatePost(request)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusPermanentRedirect, FormatKeyPath(uuid))
	//return c.String(http.StatusOK, uuid)
}

func (r Router) deletePost(c echo.Context) error {
	id := c.Param("uuid")
	if err := r.postCreator.RemovePost(id); err != nil {
		return err
	}

	return c.JSON(http.StatusAccepted, id)
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	c.Logger().Error(err)

	h, err := errorHTML(err, code)
	if err != nil {
		c.Logger().Error(err)
	}

	c.HTML(code, h)
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

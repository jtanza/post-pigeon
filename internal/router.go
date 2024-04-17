package internal

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jtanza/post-pigeon/internal/model"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
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
	db          *gorm.DB
	postCreator PostCreator
}

func NewRouter(db *gorm.DB, postCreator PostCreator) Router {
	return Router{db, postCreator}
}

func (r Router) Engine() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Validator = &CustomValidator{validator: validator.New()}

	e.Static("public", "./public")

	e.File("/about", "public/about.html")
	e.File("/new", "public/new.html")

	e.POST("/posts", r.createPost)

	return e
}

func (r Router) createPost(c echo.Context) error {
	j := make(map[string]interface{})
	if err := json.NewDecoder(c.Request().Body).Decode(&j); err != nil {
		return err
	}
	log.Printf("Json Request: %+v", j)

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

	//return c.Redirect(http.StatusPermanentRedirect, resp)
	return c.String(http.StatusOK, uuid)
}

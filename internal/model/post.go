package model

import (
	"gorm.io/gorm"
)

type PostRequest struct {
	Title string `form:"title" validate:"required"`
	Body  string `form:"body" validate:"required"`
}

type Post struct {
	gorm.Model
	ID    int
	UUID  string
	Title string
}

type PostContent struct {
	gorm.Model
	ID       int
	PostUUID string
	HTML     string
}

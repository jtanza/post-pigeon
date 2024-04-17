package model

import (
	"gorm.io/gorm"
)

type PostRequest struct {
	Title string `json:"title" validate:"required"`
	Body  string `json:"body" validate:"required"`
}

type Post struct {
	gorm.Model
	ID    int
	UUID  string
	Title string
}

type PostLocation struct {
	gorm.Model
	ID       int
	PostUUID string
	S3       string
	URL      string
}

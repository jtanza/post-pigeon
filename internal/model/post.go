package model

import (
	"gorm.io/gorm"
)

type PostRequest struct {
	Title     string `form:"title" validate:"required"`
	Body      string `form:"body" validate:"required"`
	PublicKey string `form:"publickey" validate:"required"`
	Signature string `form:"signature" validate:"required"`
}

type PostDeleteRequest struct {
	UUID      string `form:"uuid" validate:"required"`
	Signature string `form:"signature" validate:"required"`
}

type Post struct {
	gorm.Model
	ID   int
	UUID string
}

type PostContent struct {
	gorm.Model
	ID       int
	PostUUID string
	Title    string
	HTML     string
	Message  string
	Key      string
}

package model

import (
	"gorm.io/gorm"
	"time"
)

type PostRequest struct {
	Title      string `form:"title" validate:"required"`
	Body       string
	PublicKey  string `form:"publickey" validate:"required"`
	Signature  string `form:"signature" validate:"required"`
	Expiration string `form:"expiration"`
}

type PostDeleteRequest struct {
	UUID      string `form:"uuid" validate:"required"`
	Signature string `form:"signature" validate:"required"`
}

type Post struct {
	gorm.Model
	ID          int
	UUID        string
	Key         string
	Fingerprint string
	ExpiresAt   *time.Time
}

type PostContent struct {
	gorm.Model
	ID       int
	PostUUID string
	Title    string
	HTML     string
	Message  string
}

type FullPost struct {
	UUID        string
	Key         string
	Fingerprint string
	Title       string
	HTML        string
	Message     string
	CreatedAt   time.Time
}

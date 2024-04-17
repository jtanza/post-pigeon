package internal

import (
	"bytes"
	"html/template"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/jtanza/post-pigeon/internal/model"
	"gorm.io/gorm"
)

type PostCreator struct {
	DB       *gorm.DB
	S3Client *s3.Client
}

func NewPostCreator(db *gorm.DB, s3Client *s3.Client) PostCreator {
	return PostCreator{db, s3Client}
}

func (r PostCreator) CreatePost(request model.PostRequest) (string, error) {
	html, err := r.toHTML(request)
	if err != nil {
		return "", err
	}

	postUUID := uuid.New().String()
	s3Url, err := UploadPost(r.S3Client, postUUID, html)
	if err != nil {
		return "", err
	}

	if err = StorePost(r.DB, postUUID, request, s3Url); err != nil {
		return "", err
	}

	return postUUID, err
}

func (r PostCreator) toHTML(request model.PostRequest) (string, error) {
	t, err := template.New("post").ParseFiles("templates/post")
	if err != nil {
		return "", err
	}

	m, err := parseRequest(request)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, m); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func parseRequest(request model.PostRequest) (map[string]any, error) {
	m := map[string]interface{}{
		"Title": request.Title,
		"Body":  request.Body,
	}

	return m, nil
}

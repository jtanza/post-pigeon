package internal

import (
	"bytes"
	"html/template"

	"github.com/google/uuid"
	"github.com/jtanza/post-pigeon/internal/model"
)

const namespace = "post-pigeon-namespace"

type PostManager struct {
	db       DB
	s3Client S3Client
}

func NewPostManager(db DB, s3Client S3Client) PostManager {
	return PostManager{db, s3Client}
}

func (r PostManager) CreatePost(request model.PostRequest) (string, error) {
	html, err := r.toHTML(request)
	if err != nil {
		return "", err
	}

	postUUID, err := generateDeterministicUUID(request)
	if err != nil {
		return "", err
	}

	s3Url, err := r.s3Client.UploadPostObject(postUUID, html)
	if err != nil {
		return "", err
	}

	if err = r.db.PersistPost(postUUID, request, s3Url); err != nil {
		return "", err
	}

	return postUUID, err
}

func (r PostManager) RemovePost(postUUID string) error {
	if err := r.s3Client.DeletePostObject(postUUID); err != nil {
		return err
	}

	return r.db.DeletePost(postUUID)
}

func (r PostManager) toHTML(request model.PostRequest) (string, error) {
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

func generateDeterministicUUID(request model.PostRequest) (string, error) {
	id, err := uuid.FromBytes([]byte(namespace)[:16])
	if err != nil {
		return "", err
	}

	return uuid.NewSHA1(id, []byte(request.Body)).String(), nil
}

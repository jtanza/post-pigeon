package internal

import (
	"bytes"
	"encoding/gob"
	"errors"
	"hash/fnv"
	"html/template"
	"time"

	"github.com/google/uuid"
	"github.com/jtanza/post-pigeon/internal/model"
)

const namespace = "post-pigeon-namespace"

type PostManager struct {
	db DB
}

func NewPostManager(db DB) PostManager {
	return PostManager{db}
}

func (r PostManager) CreatePost(request model.PostRequest) (string, error) {
	if err := ValidateSignature(request.PublicKey, request.Signature, request.Body); err != nil {
		return "", errors.New("could not validate signature")
	}

	html, err := r.toHTML(request)
	if err != nil {
		return "", err
	}

	postUUID, err := GenerateDeterministicUUID(request.PublicKey, request.Body)
	if err != nil {
		return "", err
	}

	if err = r.db.PersistPost(postUUID, request, html); err != nil {
		return "", err
	}

	return postUUID, err
}

func (r PostManager) RemovePost(request model.PostDeleteRequest) error {
	post, err := r.db.GetPost(request.UUID)
	if err != nil {
		return err
	}

	if len(post.Key) == 0 {
		// dont leak proof of a non-existent post
		return errors.New("could not verify signature")
	}

	content, err := r.db.GetPostContent(request.UUID)
	if err != nil {
		return err
	}

	// https://crypto.stackexchange.com/q/111536/116199
	if err = ValidateSignature(post.Key, request.Signature, content.Message); err != nil {
		return errors.New("could not validate signature")
	}

	return r.db.DeletePost(request)
}

func (r PostManager) FetchPostContent(postUUID string) (model.PostContent, error) {
	return r.db.GetPostContent(postUUID)
}

// we use the base64 encoded signature to produce the deterministic (version 5) uuid
func GenerateDeterministicUUID(key string, content string) (string, error) {
	id, err := uuid.FromBytes([]byte(namespace)[:16])
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode([]string{key, content}); err != nil {
		return "", err
	}

	h := fnv.New64()
	if _, err = h.Write(buf.Bytes()); err != nil {
		return "", err
	}

	return uuid.NewSHA1(id, h.Sum(nil)).String(), nil
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
		"Title":        request.Title,
		"Body":         request.Body,
		"CreationDate": time.Now().Format(time.DateOnly),
	}

	return m, nil
}

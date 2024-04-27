package internal

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/labstack/gommon/log"
	"hash/fnv"
	"html/template"
	"time"

	"github.com/google/uuid"
	"github.com/jtanza/post-pigeon/internal/model"
)

const namespace = "post-pigeon-namespace"

type PostManager struct {
	db    DB
	cache gcache.Cache
}

func NewPostManager(db DB, cache gcache.Cache) PostManager {
	return PostManager{db, cache}
}

func (r PostManager) CreatePost(request model.PostRequest) (string, error) {
	if err := ValidateSignature(request.PublicKey, request.Signature, request.Body); err != nil {
		return "", errors.New("could not validate signature")
	}

	m, err := parseRequest(request)
	if err != nil {
		return "", err
	}

	html, err := toHTML("post", m)
	if err != nil {
		return "", err
	}

	postUUID, err := GenerateDeterministicUUID(request.PublicKey, request.Body)
	if err != nil {
		return "", err
	}

	if err = r.db.PersistPost(postUUID, request, html, ParseExpiration(request.Expiration)); err != nil {
		return "", err
	}

	return postUUID, err
}

func (r PostManager) IsDuplicate(request model.PostRequest) (bool, error) {
	postUUID, err := GenerateDeterministicUUID(request.PublicKey, request.Body)
	if err != nil {
		return false, err
	}

	post, err := r.db.GetPost(postUUID)
	if err != nil {
		return false, err
	}

	return post != nil, nil
}

func (r PostManager) HasPosts(fingerprint string) (bool, error) {
	posts, err := r.db.GetUserPosts(fingerprint)
	if err != nil {
		return false, err
	}
	return len(posts) > 0, nil
}

func (r PostManager) RemovePost(request model.PostDeleteRequest) error {
	post, err := r.db.GetPost(request.UUID)
	if err != nil {
		return err
	}

	if post == nil {
		// dont leak proof of a non-existent post
		return errors.New("could not verify signature")
	}

	content, err := r.db.GetPostContent(post.UUID)
	if err != nil {
		return err
	}

	// https://crypto.stackexchange.com/q/111536/116199
	if err = ValidateSignature(post.Key, request.Signature, content.Message); err != nil {
		return errors.New("could not validate signature")
	}

	r.cache.Remove(post.UUID)
	return r.db.DeletePost(request)
}

func (r PostManager) FetchPostContent(postUUID string) (*model.PostContent, error) {
	if r.cache.Has(postUUID) {
		post, err := r.cache.Get(postUUID)
		if err != nil {
			log.Error(err)
		} else {
			log.Infof("serving post %s from cache. hit rate: %f", postUUID, r.cache.HitRate())
			return post.(*model.PostContent), nil
		}
	}

	post, err := r.db.GetPostContent(postUUID)
	if err != nil {
		return nil, err
	}
	if err = r.cache.Set(postUUID, post); err != nil {
		log.Error(err)
	}

	return post, nil
}

func (r PostManager) GetAllUserPosts(fingerprint string) (string, error) {
	posts, err := r.db.GetUserPosts(fingerprint)
	if err != nil {
		return "", err
	}

	data := make([]map[string]interface{}, 0)
	for _, p := range posts {
		m := map[string]interface{}{
			"Fingerprint": fingerprint,
			"UUID":        p.UUID,
			"Title":       p.Title,
			"Date":        p.CreatedAt.Format(time.DateOnly),
		}
		data = append(data, m)
	}

	return toHTML("posts", data)
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

func ParseExpiration(expirationRequest string) *time.Time {
	expiration := time.Now().UTC()
	switch expirationRequest {
	case "1 hour":
		expiration = expiration.Add(time.Hour)
	case "1 day":
		expiration = expiration.AddDate(0, 0, 1)
	case "1 month":
		expiration = expiration.AddDate(0, 1, 0)
	case "1 year":
		expiration = expiration.AddDate(1, 0, 0)
	default:
		return nil
	}
	return &expiration
}

func toHTML(templateName string, data any) (string, error) {
	t, err := template.New(templateName).ParseFiles(fmt.Sprintf("templates/%s", templateName))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func parseRequest(request model.PostRequest) (map[string]any, error) {
	fingerprint, err := Fingerprint(request.PublicKey)
	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{
		"Title":        request.Title,
		"Body":         request.Body,
		"Fingerprint":  fingerprint,
		"CreationDate": time.Now().Format(time.DateOnly),
	}

	return m, nil
}

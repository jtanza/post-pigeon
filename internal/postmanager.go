package internal

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/labstack/gommon/log"
	"github.com/microcosm-cc/bluemonday"
	"hash/fnv"
	"html/template"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jtanza/post-pigeon/internal/model"
)

type PostManager struct {
	db                 DB
	cache              gcache.Cache
	namespace          string
	markdownExtensions parser.Extensions
}

func NewPostManager(db DB, cache gcache.Cache) PostManager {
	namespace := os.Getenv("POST_PIGEON_NS")
	if len(namespace) == 0 {
		log.Fatal("unset namespace for app")
	}

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	return PostManager{db, cache, namespace, extensions}
}

func (pm PostManager) CreatePost(request model.PostRequest) (string, error) {
	if err := ValidateSignature(request.PublicKey, request.Signature, request.Body); err != nil {
		return "", errors.New("could not validate signature")
	}

	m, err := pm.formatRequestData(request)
	if err != nil {
		return "", err
	}

	renderedHTML, err := toHTML("post", m)
	if err != nil {
		return "", err
	}

	postUUID, err := GenerateDeterministicUUID(request.PublicKey, request.Title, pm.namespace)
	if err != nil {
		return "", err
	}

	if err = pm.db.PersistPost(postUUID, request, renderedHTML, ParseExpiration(request.Expiration)); err != nil {
		return "", err
	}

	return postUUID, err
}

func (pm PostManager) IsDuplicate(request model.PostRequest) (bool, error) {
	postUUID, err := GenerateDeterministicUUID(request.PublicKey, request.Title, pm.namespace)
	if err != nil {
		return false, err
	}

	post, err := pm.db.GetPost(postUUID)
	if err != nil {
		return false, err
	}

	return post != nil, nil
}

func (pm PostManager) HasPosts(fingerprint string) (bool, error) {
	posts, err := pm.db.GetUserPosts(fingerprint)
	if err != nil {
		return false, err
	}
	return len(posts) > 0, nil
}

// RemovePost will use the stored public key and post message to delete a post from the provided request
// iff the signatures can be verified using the stored key/message.
func (pm PostManager) RemovePost(request model.PostDeleteRequest) error {
	post, err := pm.db.GetPost(request.UUID)
	if err != nil {
		return err
	}

	if post == nil {
		// dont leak proof of a non-existent post
		return errors.New("could not verify signature")
	}

	content, err := pm.db.GetPostContent(post.UUID)
	if err != nil {
		return err
	}

	// https://crypto.stackexchange.com/q/111536/116199
	if err = ValidateSignature(post.Key, request.Signature, content.Message); err != nil {
		return errors.New("could not validate signature")
	}

	pm.cache.Remove(post.UUID)
	return pm.db.DeletePost(request)
}

func (pm PostManager) FetchPostContent(postUUID string) (*model.PostContent, error) {
	if pm.cache.Has(postUUID) {
		post, err := pm.cache.Get(postUUID)
		if err != nil {
			log.Error(err)
		} else {
			log.Infof("serving post %s from cache. hit rate: %f", postUUID, pm.cache.HitRate())
			return post.(*model.PostContent), nil
		}
	}

	post, err := pm.db.GetPostContent(postUUID)
	if err != nil {
		return nil, err
	}
	if err = pm.cache.Set(postUUID, post); err != nil {
		log.Error(err)
	}

	return post, nil
}

func (pm PostManager) GetAllUserPosts(fingerprint string) (string, error) {
	posts, err := pm.db.GetUserPosts(fingerprint)
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

// GenerateDeterministicUUID creates a deterministic (version 5) uuid from the provided key and title
func GenerateDeterministicUUID(key, title, namespace string) (string, error) {
	if len(key) == 0 || len(title) == 0 {
		return "", errors.New("invalid title or key")
	}

	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode([]string{key, title}); err != nil {
		return "", err
	}

	h := fnv.New64()
	if _, err := h.Write(buf.Bytes()); err != nil {
		return "", err
	}

	id, err := uuid.FromBytes([]byte(namespace)[:16])
	if err != nil {
		return "", err
	}

	return uuid.NewSHA1(id, h.Sum(nil)).String(), nil
}

// ParseExpiration converts common expiration times used on our frontend into actual time periods
// our post reaper will leverage to expire posts.
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

func (pm PostManager) formatRequestData(request model.PostRequest) (map[string]any, error) {
	fingerprint, err := Fingerprint(request.PublicKey)
	if err != nil {
		return nil, err
	}

	md := parser.NewWithExtensions(pm.markdownExtensions).Parse([]byte(request.Body))
	renderer := html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags | html.HrefTargetBlank})
	sanitizedBody := bluemonday.UGCPolicy().SanitizeBytes(markdown.Render(md, renderer))

	m := map[string]interface{}{
		"Title":        request.Title,
		"Body":         template.HTML(sanitizedBody),
		"Fingerprint":  fingerprint,
		"CreationDate": time.Now().Format(time.DateOnly),
	}

	return m, nil
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

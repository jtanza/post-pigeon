package internal

import (
	"errors"
	"log"
	"time"

	"github.com/jtanza/post-pigeon/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type DB struct {
	db *gorm.DB
}

func NewDB() DB {
	db, err := gorm.Open(sqlite.Open(createDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	return DB{db}
}

func (d DB) PersistPost(postUUID string, request model.PostRequest, html string, expiration *time.Time) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		fingerprint, err := Fingerprint(request.PublicKey)
		if err != nil {
			return err
		}

		post := model.Post{UUID: postUUID, Key: request.PublicKey, Fingerprint: fingerprint, ExpiresAt: expiration}
		if postResult := tx.Create(&post); postResult.Error != nil {
			return postResult.Error
		}

		postLocation := model.PostContent{
			PostUUID: postUUID,
			HTML:     html,
			Message:  request.Body,
			Title:    request.Title,
		}
		if postLocationResult := tx.Create(&postLocation); postLocationResult.Error != nil {
			return postLocationResult.Error
		}

		return nil
	})
}

func (d DB) DeletePost(postDeleteRequest model.PostDeleteRequest) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if postDelete := d.db.Where("uuid = ?", postDeleteRequest.UUID).Delete(&model.Post{}); postDelete.Error != nil {
			return postDelete.Error
		}

		if postContentDelete := d.db.Where("post_uuid = ?", postDeleteRequest.UUID).Delete(&model.PostContent{}); postContentDelete.Error != nil {
			return postContentDelete.Error
		}

		return nil
	})
}

func (d DB) GetPostContent(postUUID string) (*model.PostContent, error) {
	var postContent model.PostContent
	if postQuery := d.db.Where("post_uuid = ?", postUUID).First(&postContent); postQuery.Error != nil {
		if errors.Is(postQuery.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, postQuery.Error
	}
	return &postContent, nil
}

func (d DB) GetPost(postUUID string) (*model.Post, error) {
	var post model.Post
	if postQuery := d.db.Where("uuid = ?", postUUID).First(&post); postQuery.Error != nil {
		if errors.Is(postQuery.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, postQuery.Error
	}
	return &post, nil
}

func (d DB) GetUserPosts(fingerprint string) ([]model.FullPost, error) {
	var posts []model.FullPost
	if postQuery := d.db.Model(&model.Post{}).Select("post.UUID, post.Key, post.Fingerprint, post.created_at, post_content.Title, post_content.HTML, post_content.Message").Joins("left join post_content on post.uuid = post_content.post_uuid").Where("post.fingerprint = ?", fingerprint).Scan(&posts); postQuery.Error != nil {
		return nil, postQuery.Error
	}
	return posts, nil
}

func (d DB) DeleteExpiredPosts() (int64, error) {
	postQuery := d.db.Unscoped().Model(&model.Post{}).Where("expires_at <= datetime('now')").Delete(&model.Post{})
	if postQuery.Error != nil {
		return 0, postQuery.Error
	}
	return postQuery.RowsAffected, nil
}

func createDSN() string {
	return "file:postpigeon.db"
}

package internal

import (
	"log"

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

func (d DB) PersistPost(postUUID string, request model.PostRequest, html string) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		post := model.Post{
			UUID:  postUUID,
			Title: request.Title,
		}
		if postResult := tx.Create(&post); postResult.Error != nil {
			return postResult.Error
		}

		postLocation := model.PostContent{
			PostUUID: postUUID,
			HTML:     html,
		}
		if postLocationResult := tx.Create(&postLocation); postLocationResult.Error != nil {
			return postLocationResult.Error
		}

		return nil
	})
}

func (d DB) DeletePost(postUUID string) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if postDelete := d.db.Where("uuid = ?", postUUID).Delete(&model.Post{}); postDelete.Error != nil {
			return postDelete.Error
		}

		if postContentDelete := d.db.Where("post_uuid = ?", postUUID).Delete(&model.PostContent{}); postContentDelete.Error != nil {
			return postContentDelete.Error
		}

		return nil
	})
}

func (d DB) GetPostContent(postUUID string) (model.PostContent, error) {
	postContent := model.PostContent{}
	if postQuery := d.db.Where("post_uuid = ?", postUUID).First(&postContent); postQuery.Error != nil {
		return model.PostContent{}, postQuery.Error
	}
	return postContent, nil
}

func createDSN() string {
	// https://github.com/mattn/go-sqlite3?tab=readme-ov-file#dsn-examples
	// file:test.db?cache=shared&mode=memory
	// build user os.Getenv("DBUSER"), auth os.Getenv("DBPASS") etc
	return "file:postpigeon.db"
}

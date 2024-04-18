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

func (d DB) PersistPost(postUUID string, request model.PostRequest, s3Location string) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		post := model.Post{
			UUID:  postUUID,
			Title: request.Title,
		}
		if postResult := tx.Create(&post); postResult.Error != nil {
			return postResult.Error
		}

		postLocation := model.PostLocation{
			PostUUID: postUUID,
			S3:       s3Location,
			URL:      s3Location,
		}
		if postLocationResult := tx.Create(&postLocation); postLocationResult.Error != nil {
			return postLocationResult.Error
		}

		return nil
	})
}

func (d DB) DeletePost(postUUID string) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		post := model.Post{UUID: postUUID}
		if queryResult := d.db.First(&post); queryResult.Error != nil {
			return queryResult.Error
		}

		return d.db.Delete(&model.Post{}, post.ID).Error
	})
}

func createDSN() string {
	// https://github.com/mattn/go-sqlite3?tab=readme-ov-file#dsn-examples
	// file:test.db?cache=shared&mode=memory
	// build user os.Getenv("DBUSER"), auth os.Getenv("DBPASS") etc
	return "file:postpigeon.db"
}

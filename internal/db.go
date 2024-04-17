package internal

import (
	"log"

	"github.com/google/uuid"
	"github.com/jtanza/post-pigeon/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func OpenDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(createDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func StorePost(db *gorm.DB, request model.PostRequest, s3Location string) (string, error) {
	postUUID := uuid.New().String()

	return postUUID, db.Transaction(func(tx *gorm.DB) error {
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

func createDSN() string {
	// https://github.com/mattn/go-sqlite3?tab=readme-ov-file#dsn-examples
	// file:test.db?cache=shared&mode=memory
	// build user os.Getenv("DBUSER"), auth os.Getenv("DBPASS") etc
	return "file:postpigeon.db"
}

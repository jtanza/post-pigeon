package main

import (
	"github.com/bluele/gcache"
	"github.com/jtanza/post-pigeon/internal"
	"github.com/labstack/gommon/log"
	"time"
)

const cacheSize = 50

func main() {
	db := internal.NewDB()
	go postReaper(db)

	cache := gcache.New(cacheSize).LRU().Build()
	r := internal.NewRouter(db, internal.NewPostManager(db, cache)).Engine()
	r.Logger.Fatal(r.Start(":8080"))
}

func postReaper(db internal.DB) {
	for range time.Tick(time.Minute * 5) {
		deleted, err := db.DeleteExpiredPosts()
		if err != nil {
			log.Error(err)
		} else {
			log.Infof("deleted %d expired posts", deleted)
		}
	}
}

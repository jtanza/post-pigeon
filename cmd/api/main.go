package main

import (
	"fmt"
	"github.com/bluele/gcache"
	"github.com/jtanza/post-pigeon/internal"
	"github.com/labstack/gommon/log"
	"os"
	"time"
)

const cacheSize = 50

func main() {
	db := internal.NewDB()
	go postReaper(db)

	logFile, err := os.OpenFile("log/postpigeon.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("error opening file: %v", err))
	}
	defer logFile.Close()

	r := internal.NewRouter(db, internal.NewPostManager(db, gcache.New(cacheSize).LRU().Build())).Engine(logFile)
	r.Logger.Fatal(r.Start(os.Getenv("POSTPIGEON_WEB_PORT")))
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

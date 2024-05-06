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
	go reapPosts(db)

	logFile, err := os.OpenFile("log/postpigeon.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("error opening file: %v", err))
	}
	defer logFile.Close()

	cache := gcache.New(cacheSize).LRU().Build()
	r := internal.NewRouter(db, internal.NewPostManager(db, cache)).Engine(logFile)
	r.Logger.Fatal(r.StartAutoTLS(":443"))
	// redirects to 443
	r.Logger.Fatal(r.Start(":80"))
}

func reapPosts(db internal.DB) {
	for range time.Tick(time.Minute * 5) {
		deleted, err := db.DeleteExpiredPosts()
		if err != nil {
			log.Error(err)
		} else {
			log.Infof("deleted %d expired posts", deleted)
		}
	}
}

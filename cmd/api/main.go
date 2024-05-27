package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/jtanza/post-pigeon/internal"
	"github.com/labstack/gommon/log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const cacheSize = 50

func main() {
	logFile, err := os.OpenFile("log/postpigeon.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("error opening file: %v", err))
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	db := internal.NewDB()
	go reapPosts(db)

	cache := gcache.New(cacheSize).LRU().Build()
	r := internal.NewRouter(db, internal.NewPostManager(db, cache)).Engine(logFile)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start server
	go func() {
		if err = r.StartAutoTLS(":443"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("shutting down the server %s", err)
		}
		// redirects to 443 in prod
		log.Fatal(r.Start(":80"))
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-ctx.Done()
	log.Warn("interrupt received, shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = r.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
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

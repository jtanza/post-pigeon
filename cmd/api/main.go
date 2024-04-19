package main

import "github.com/jtanza/post-pigeon/internal"

func main() {
	db := internal.NewDB()
	s3 := internal.NewS3Client()

	r := internal.NewRouter(db, internal.NewPostManager(db, s3)).Engine()
	r.Logger.Fatal(r.Start(":8080"))
}

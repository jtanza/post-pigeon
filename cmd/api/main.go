package main

import "github.com/jtanza/post-pigeon/internal"

func main() {
	db := internal.OpenDB()

	r := internal.NewRouter(db, internal.NewPostCreator(db, internal.CreateS3Client())).Engine()

	r.Logger.Fatal(r.Start(":8080"))
}

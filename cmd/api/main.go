package main

import "github.com/jtanza/post-pigeon/internal"

func main() {
	db := internal.NewDB()
	r := internal.NewRouter(db, internal.NewPostManager(db)).Engine()
	r.Logger.Fatal(r.Start(":8080"))
}

# Post Pigeon

Post Pigeon is a web app that allows users to publish and share posts with others. 
It can be self-hosted, run locally or used at [post-pigeon.com](https://post-pigeon.com)

Checkout the [docs](https://post-pigeon.com) for lots of info on using the app and how it works.

## Running Locally

Running locally should be pretty straight forward as there's not much by way of extra dependencies. You'll
need Go and SQLite and that's basically it. 

```shell
$ git clone https://github.com/jtanza/post-pigeon.git && cd post-pigeon
```
Create your db
```shell
$ touch postpigeon.db
$ sqlite3 postpigeon.db < migrations/1_add_init_tables.up.sql
```
Set your SHA1 namespace
```shell
$ export POST_PIGEON_NS="whatever.youd.like"
```
Run the app
```shell
$ go run cmd/api/main.go
```
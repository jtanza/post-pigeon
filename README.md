# Post Pigeon

Post Pigeon is a web app that allows users to publish and share posts with others. 
It can be self-hosted, run locally or used at [post-pigeon.com](https://post-pigeon.com)

Checkout the [docs](https://post-pigeon.com) for info on using the app and how it works.

## Example

<img width="1436" alt="Screenshot 2024-05-08 at 8 51 51 PM" src="https://github.com/jtanza/post-pigeon/assets/10635096/d7495a7e-fcc5-4a9c-b8e7-5953812f7fff">


## Running Locally

Running locally should be pretty straight forward. You'll need Go and SQLite and that's basically it.

Clone the repo
```shell
$ git clone https://github.com/jtanza/post-pigeon.git && cd post-pigeon
```
Create your db
```shell
$ touch postpigeon.db
$ sqlite3 postpigeon.db < migrations/1_add_init_tables.up.sql
```
Set your SHA1 [namespace](https://github.com/jtanza/post-pigeon/blob/main/internal/postmanager.go#L174-L179)
```shell
$ export POST_PIGEON_NS="whatever.youd.like.uuid"
```
Run the app and point your browser to `localhost:80` 
```shell
$ go run cmd/api/main.go
```

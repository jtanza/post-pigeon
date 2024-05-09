# Post Pigeon

Post Pigeon is a web app that allows users to publish and share posts with others. 
It can be self-hosted, run locally or used at [post-pigeon.com](https://post-pigeon.com)

Checkout the [docs](https://post-pigeon.com) for info on using the app and how it works.

## Example

<img width="1435" alt="Screenshot 2024-05-08 at 7 23 43 PM" src="https://github.com/jtanza/post-pigeon/assets/10635096/f75e42f7-3a94-4089-b04e-44b6790468e1">


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
Set your SHA1 namespace
```shell
$ export POST_PIGEON_NS="whatever.youd.like.uuid"
```
Run the app
```shell
$ go run cmd/api/main.go
```

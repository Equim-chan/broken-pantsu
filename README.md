# Broken Pantsu
[![Travis](https://img.shields.io/travis/Equim-chan/broken-pantsu.svg)](https://travis-ci.org/Equim-chan/broken-pantsu)
[![Go Report Card](https://goreportcard.com/badge/github.com/Equim-chan/broken-pantsu)](https://goreportcard.com/report/github.com/Equim-chan/broken-pantsu)
* Inspired by [fakku.dating](https://fakku.dating/)
* Aimed to provide more stable connections between matched partners
* Designed for high concurrency and performance
* Raw WebSocket, instead of socket.io
* __Love, rather than sorry__

## Setup
```console
$ go get -u github.com/Equim-chan/broken-pantsu
$ $GOPATH/bin/broken-pantsu
```
or manually
```console
$ git clone git@github.com:Equim-chan/broken-pantsu.git
$ make
$ ./broken-pantsu
```
Build executable files for all platforms and archs
```console
$ make release -j4
```
> Running `make` with the `-j4` flag will cause it to run 4 compilation jobs concurrently which may significantly reduce build time. The number after `-j` can be changed to best suit the number of processor cores on your machine. If you run into problems running `make` with concurrency, try running it without the `-j4` flag. See the [GNU Make Documentation](https://www.gnu.org/software/make/manual/html_node/Parallel.html) for more information.

Config can be passed via environment variables. Example:
```console
$ BP_ADDR=:5543 BP_ROOT_PATH=../dist BP_QUEUE_CAP=100 ./broken-pantsu
```
List of environment variables:

| Field | Default Value | Comment |
| ----    | -------    | --- |
| BP_ADDR | localhost:56833 | Where the application listens to (56833 means "loved") |
| BP_ROOT_PATH | ./public | Where the static files are located. Relative path will be resolved into absolute path automatically |
| BP_QUEUE_CAP | 1000 | The capacity of `singleQueue` and `lovelornQueue` |
| BP_COOKIE_AGE | 168h | The age of cookie |
| BP_LOVELORN_AGE | 1h30m | The age of lovelorn pairs stored in redis |
| BP_REDIS_ADDR | localhost:6379 | The address of redis |
| BP_REDIS_PASS | (empty) | The password of redis |
| BP_REDIS_DB | 0 | The DB of redis |
| BP_QUIET | (empty) | Set any value to disable logging |

## Dependencies
We use [glide](https://github.com/Masterminds/glide) as package manager.
* [github.com/gorilla/websocket](https://github.com/gorilla/websocket)
* [github.com/satori/go.uuid](https://github.com/satori/go.uuid)
* [github.com/go-redis/redis](https://github.com/go-redis/redis)

## TODO
There are lots of things to do at the moment...

### Backend
* [x] recv -> InBoundMessage -> unpack -> process -> pack -> OutBoundMessage -> send
* [x] reject new connection from the same client when there is already one
* [x] enforce matching algorithm
* [ ] check for thread safety
* [ ] check for possible memory leak
* [x] setup travis CI
* [x] migrate build_all into Makefile
* [ ] add test suite

### Frontend
* [ ] complete the UI
* [x] add emoji support
* [ ] add "savelog" button
* [ ] add "switch partner" button
* [ ] add "quit" button
* [ ] check for XSS
* [x] separate HTML, CSS and JS
* [x] add browser-out-of-date warning
* [ ] add auto-reconnection after dc
* [ ] _add i18n support_

### Features
* [ ] display "your partner is typing"
* [ ] display and auto-refreash online users count
* [ ] display partner's nickname, avatar, likes, timezone
* [x] __ensure partners can still find each others after unexpected disconnection__

## License
[Apache-2.0](https://github.com/Equim-chan/broken-pantsu/blob/master/LICENSE)
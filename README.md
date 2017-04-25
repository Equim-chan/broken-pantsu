# Broken Pantsu
[![Travis](https://img.shields.io/travis/Equim-chan/broken-pantsu.svg)](https://travis-ci.org/Equim-chan/broken-pantsu)
[![Go Report Card](https://goreportcard.com/badge/github.com/Equim-chan/broken-pantsu)](https://goreportcard.com/report/github.com/Equim-chan/broken-pantsu)
* Inspired by [fakku.dating](https://fakku.dating/)
* Aimed to provide more stable connections between matched partners
* Designed for high concurrency and performance
* Raw WebSocket, instead of socket.io
* __Love, better than sorry__

## Setup
```bash
$ go get -u github.com/Equim-chan/broken-pantsu
```
or manually
```bash
$ git clone git@github.com:Equim-chan/broken-pantsu.git
$ make love       # "love, better than sorry" after all
$ ./broken-pantsu
```
Config can be passed via environment. Example:
```bash
$ BP_ADDR=:5543 BP_PUB_PATH=../dist BP_MAX_QUEUE_LEN=100 ./broken-pantsu
```

## Dependencies
* [github.com/gorilla/websocket](https://github.com/gorilla/websocket)
* [github.com/satori/go.uuid](https://github.com/satori/go.uuid)
* [github.com/go-redis/redis](https://github.com/go-redis/redis)

## TODO
There are lots of things to do at the moment...

### Backend
* [x] recv -> InBoundMessage -> unpack -> process -> pack -> OutBoundMessage -> send
* [ ] reject new connection from the same client when there is already one
* [ ] add session support (in age of 3 hours)
* [x] enforce matching algorithm
* [ ] check for thread safety
* [x] setup travis CI
* [ ] migrate build_all into Makefile
* [ ] add test suite

### Frontend
* [ ] complete the UI
* [x] add emoji support
* [ ] check for XSS
* [x] separate HTML, CSS and JS
* [x] add browser-out-of-date warning
* [ ] add savelog button
* [ ] _add auto-reconnection after dc_

### Features
* [ ] display "your partner is typing"
* [ ] display "√" for the partner has read
* [ ] display and auto-refreash online users count
* [ ] display partner's nickname, avatar, likes, timezone
* [ ] ensure partners can still find each others after unexpected disconnection

## License
[Apache-2.0](https://github.com/Equim-chan/broken-pantsu/blob/master/LICENSE)
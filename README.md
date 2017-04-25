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
```
or manually
```console
$ git clone git@github.com:Equim-chan/broken-pantsu.git
$ make love       # "love, rather than sorry" after all
$ ./broken-pantsu
```
Build executable files for all platforms and archs
```console
$ make release -j4
```
> Running `make` with the `-j4` flag will cause it to run 4 compilation jobs concurrently which may significantly reduce build time. The number after `-j` can be changed to best suit the number of processor cores on your machine. If you run into problems running `make` with concurrency, try running it without the `-j4` flag. See the [GNU Make Documentation](https://www.gnu.org/software/make/manual/html_node/Parallel.html) for more information.

Config can be passed via environment. Example:
```console
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
* [x] migrate build_all into Makefile
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
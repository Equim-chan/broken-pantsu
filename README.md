# Broken Pantsu
* Inspired by [fakku.dating](https://fakku.dating/)
* Aimed to provide more stable connections between matched partners
* Designed for high concurrency and performance
* Go, instead of Node.js; raw WebSocket, instead of socket.io
* __Love, better than sorry__

## Setup
```bash
$ git clone git@github.com:Equim-chan/broken-pantsu.git
$ make build
$ ./main
```

## Dependencies
* [github.com/gorilla/websocket](https://github.com/gorilla/websocket)

## TODO
There are lots of things to do at the moment...

### Backend
* [x] recv -> InBoundMessage -> unpack -> process -> pack -> OutBoundMessage -> send
* [ ] reject new connection from the same client when there is already one
* [ ] add session support (in age of 3 hours)
* [x] enforce matching algorithm
* [ ] check for thread safety
* [ ] setup travis CI
* [ ] migrate build_all into Makefile

### Frontend
* [ ] complete the UI
* [x] add emoji support
* [ ] check for XSS
* [x] separate HTML, CSS and JS
* [ ] add browser-out-of-date warning
* [ ] _add auto-reconnection after dc_

### Features
* [ ] display "your partner is typing"
* [ ] display "√" for the partner has read
* [ ] display and auto-refreash online users count
* [ ] display partner's nickname, avatar, likes, timezone
* [ ] ensure partners can still find each others after unexpected disconnection

## License
[Apache-2.0](https://github.com/Equim-chan/broken-pantsu/blob/master/LICENSE)
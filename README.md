# Broken Pantsu
* Inspired by [fakku.dating](https://fakku.dating/)
* Aimed to provide more stable connections between matched partners
* Designed for high concurrency and performance
* Go, instead of Node.js; raw WebSocket, instead of socket.io
* __Love, better than sorry__

## Dependencies
* [github.com/gorilla/websocket](https://github.com/gorilla/websocket)

## TODO
There are lots of things to do at the moment...

### Backend
* [ ] recv -> InBoundMessage -> unpack -> process -> pack -> OutBoundMessage -> send
* [ ] reject new connection from the same client when there is already one
* [ ] add session support (in age of 3 hours)
* [ ] enforce matching algorithm
* [ ] check for thread safety
* [ ] travis CI

### Frontend
* [ ] complete the UI
* [x] add emoji support
* [ ] check for XSS
* [ ] separate HTML, CSS and JS
* [ ] add browser-out-of-date warning
* [ ] _add auto-reconnection after dc_

### Features
* [ ] display "your partner is typing"
* [ ] display "√" for the partner has read
* [ ] display and auto-refreash online users count
* [ ] display partner's nickname, avatar, likes, timezone
* [ ] ensure partners can still find each others after unexpected disconnection

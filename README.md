# Broken Pantsu
* Inspired by [fakku.dating](https://fakku.dating/)
* Aimed to provide more stable connections between matched partners
* Designed for high concurrency and performance
* Go, instead of Node.js; raw WebSocket, instead of socket.io
* __Love, better than sorry__

## TODO
There are lots of things to do at the moment...

### Backend
* [ ] recv -> InBoundMessage -> unpack -> process -> pack -> OutBoundMessage -> send
* [ ] add session support (in age of 3 hours)
* [ ] enforce matching algorithm
* [ ] check for thread safety
* [ ] travis CI

### Frontend
* [ ] add emoji support
* [ ] complete the UI
* [ ] separate HTML, CSS and JS
* [ ] add browser-out-of-date warning
* [ ] _add auto-reconnection after dc_

### Features
* [ ] display "your partner is typing"
* [ ] display and auto-refreash online users count
* [ ] display partner's nickname, avatar, likes, timezone
* [ ] ensure partners can still find each others after unexpected disconnection
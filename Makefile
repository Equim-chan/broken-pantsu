VERSION := ${shell date -u +%y%m%d}

build:
	go get -u github.com/gorilla/websocket
	go build -ldflags "-X main.version=${VERSION}" main.go

build-all:
	go get -u github.com/gorilla/websocket
	./build_all

clean:
	rm -rf ./release

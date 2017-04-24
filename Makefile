VERSION := ${shell date -u +%y%m%d}

build: install-dep
	go build -ldflags "-X main.version=${VERSION}" -v -x -o broken-pantsu main.go client.go

all: install-dep
	./build_all

install-dep:
	go get -u -v github.com/gorilla/websocket
	go get -u -v github.com/satori/go.uuid
	go get -u -v github.com/go-redis/redis

clean:
	rm -rf ./release

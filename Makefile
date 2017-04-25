SOFTWARE := broken-pantsu
VERSION := ${shell date -u +%y%m%d}
RELEASE := ./release
SOURCE := main.go client.go util.go

love: install-dep
	go build -ldflags "-X main.version=${VERSION}" -v -x -o ${SOFTWARE} ${SOURCE}

all: install-dep
	./build_all

install-dep:
	go get -u -v github.com/gorilla/websocket
	go get -u -v github.com/satori/go.uuid
	go get -u -v github.com/go-redis/redis

clean:
	rm -rf ./release

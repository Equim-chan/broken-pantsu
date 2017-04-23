VERSION := ${shell date -u +%y%m%d}

build:
	go get -u -v github.com/gorilla/websocket
	go build -ldflags "-X main.version=${VERSION}" -v -x -o broken-pantsu main.go client.go

build-all:
	go get -u github.com/gorilla/websocket
	./build_all

clean:
	rm -rf ./release

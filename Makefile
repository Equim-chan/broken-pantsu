SOFTWARE := broken-pantsu
SOURCE := main.go client.go util.go

RELEASE := "$(shell realpath ./release)"
BUILD_TMP := "$(shell realpath ./tmp)"
PWD := "$(shell pwd)"
VERSION := $(shell date -u +%y%m%d)
LDFLAGS := "-X main.VERSION=$(VERSION) -s -w"
GCFLAGS := ""

OSES := linux darwin windows freebsd
ARCHS := amd64 386
ARMS := 5 6 7
SUM := sha1sum

love: install-dep
	env CGO_ENABLED=0 \
		go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
		-x \
		-o $(SOFTWARE) $(SOURCE)

all: all-build \
	clean-tmp \
	all-chksum

all-build: install-dep
	@mkdir -p $(RELEASE) ; \
	mkdir -p $(BUILD_TMP) ; \
	# General
	@for os in $(OSES); do \
		for arch in $(ARCHS); do \
			suffix="" ; \
			if [ "$$os" == "windows" ]; then \
				suffix=".exe" ; \
			fi; \
			env CGO_ENABLED=0 \
				GOOS=$$os GOARCH=$$arch \
				go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
				-x \
				-o $(BUILD_TMP)/$(SOFTWARE)_$${os}_$${arch}$${suffix} $(SOURCE); \
			cd $(BUILD_TMP) ; \
			tar -zcf \
				$(RELEASE)/$(SOFTWARE)-$${os}-$${arch}-$(VERSION).tar.gz \
				$(SOFTWARE)_$${os}_$${arch}$${suffix}; \
			cd $(PWD) ; \
		done ; \
	done ; \
	# ARM
	@for v in $(ARMS); do \
	env CGO_ENABLED=0 \
		GOOS=linux GOARCH=arm GOARM=$${v} \
		go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
		-x \
		-o $(BUILD_TMP)/$(SOFTWARE)_linux_arm$${v} $(SOURCE) ; \
	done ; \
	if hash upx 2>/dev/null; then \
		upx -9 $(SOFTWARE)_linux_arm* ; \
	fi ; \
	cd $(BUILD_TMP) ; \
	tar -zcf \
		$(RELEASE)/$(SOFTWARE)-linux-arm-$(VERSION).tar.gz \
		$(SOFTWARE)_linux_arm* ; \
	cd $(PWD) ; \
	#MIPS32LE
	@env CGO_ENABLED=0 \
		GOOS=linux GOARCH=mipsle \
		go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
		-x \
		-o $(BUILD_TMP)/$(SOFTWARE)_linux_mipsle $(SOURCE); \
	env CGO_ENABLED=0 \
		GOOS=linux GOARCH=mipsle \
		go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
		-x \
		-o $(BUILD_TMP)/$(SOFTWARE)_linux_mips $(SOURCE); \
	if hash upx 2>/dev/null; then \
		upx -9 ${SOFTWARE}_linux_mips* ; \
	fi ; \
	cd $(BUILD_TMP) ; \
	tar -zcf \
		$(RELEASE)/$(SOFTWARE)-linux-mipsle-$(VERSION).tar.gz \
		$(SOFTWARE)_linux_mipsle ; \
	tar -zcf \
		$(RELEASE)/$(SOFTWARE)-linux-mips-$(VERSION).tar.gz \
		$(SOFTWARE)_linux_mips ; \
	cd $(PWD)

all-chksum:
	cd $(RELEASE); $(SUM) *

install-dep:
	go get -u github.com/gorilla/websocket
	go get -u github.com/satori/go.uuid
	go get -u github.com/go-redis/redis

clean-tmp:
	rm -rf $(BUILD_TMP)

clean: clean-tmp
	rm -rf $(RELEASE)

.PHONY: love all-build all install-dep clean-tmp clean
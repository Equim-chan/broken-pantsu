SOFTWARE := broken-pantsu

RELEASE := "$(shell readlink -f ./release)"
BUILD_TMP := "$(shell readlink -f ./tmp)"
PWD := "$(shell pwd)"
VERSION := $(shell date -u +%y%m%d)
LDFLAGS := "-X main.VERSION=$(VERSION) -s -w"
GCFLAGS := ""

OSES := linux darwin windows freebsd
ARCHS := amd64 386
ARMS := 5 6 7
SUM := sha1sum

SHELL := /bin/bash
VERBOSE := true

love: setup-tmp \
	install-dep \
	gopath-spoof \
	build

	make clean-tmp

release: setup-tmp \
	setup-release \
	install-dep \
	gopath-spoof \
	build-linux-darwin-windows-freebsd \
	build-arms \
	build-mipsle

	make clean-tmp
	make release-chksum

build:
	@GOPATH=$(BUILD_TMP) ; \
	CGO_ENABLED=0 ; \
	if [ "$(VERBOSE)" == true ]; then \
		V_ARG="-x" ; \
	fi ; \
	go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
		$$V_ARG \
		-o $(SOFTWARE)

setup-release:
	mkdir -p $(RELEASE)

build-linux-darwin-windows-freebsd: setup-release gopath-spoof
	@GOPATH=$(BUILD_TMP) ; \
	CGO_ENABLED=0 ; \
	if [ "$(VERBOSE)" == true ]; then \
		V_ARG="-x" ; \
	fi ; \
	for os in $(OSES); do \
		for arch in $(ARCHS); do \
			suffix="" ; \
			if [ "$$os" == "windows" ]; then \
				suffix=".exe" ; \
			fi; \
			env GOOS=$$os GOARCH=$$arch \
				go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
				$$V_ARG \
				-o $(BUILD_TMP)/$(SOFTWARE)_$${os}_$${arch}$${suffix} ; \
			cd $(BUILD_TMP) ; \
			tar -zcf \
				$(RELEASE)/$(SOFTWARE)-$${os}-$${arch}-$(VERSION).tar.gz \
				$(SOFTWARE)_$${os}_$${arch}$${suffix} ; \
			cd $(PWD) ; \
		done ; \
	done

build-arms: setup-release gopath-spoof
	@GOPATH=$(BUILD_TMP) ; \
	CGO_ENABLED=0 ; \
	if [ "$(VERBOSE)" == true ]; then \
		V_ARG="-x" ; \
	fi ; \
	for v in $(ARMS); do \
	env	GOOS=linux GOARCH=arm GOARM=$${v} \
		go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
		$$V_ARG \
		-o $(BUILD_TMP)/$(SOFTWARE)_linux_arm$${v} ; \
	done ; \
	if hash upx 2>/dev/null; then \
		upx -9 $(SOFTWARE)_linux_arm* ; \
	fi ; \
	cd $(BUILD_TMP) ; \
	tar -zcf \
		$(RELEASE)/$(SOFTWARE)-linux-arm-$(VERSION).tar.gz \
		$(SOFTWARE)_linux_arm*

build-mipsle: setup-release gopath-spoof
	@GOPATH=$(BUILD_TMP) ; \
	CGO_ENABLED=0 ; \
	if [ "$(VERBOSE)" == true ]; then \
		V_ARG="-x" ; \
	fi ; \
	env GOOS=linux GOARCH=mipsle \
		go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
		$$V_ARG \
		-o $(BUILD_TMP)/$(SOFTWARE)_linux_mipsle; \
	env GOOS=linux GOARCH=mipsle \
		go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
		$$V_ARG \
		-o $(BUILD_TMP)/$(SOFTWARE)_linux_mips; \
	if hash upx 2>/dev/null; then \
		upx -9 ${SOFTWARE}_linux_mips* ; \
	fi ; \
	cd $(BUILD_TMP) ; \
	tar -zcf \
		$(RELEASE)/$(SOFTWARE)-linux-mipsle-$(VERSION).tar.gz \
		$(SOFTWARE)_linux_mipsle ; \
	tar -zcf \
		$(RELEASE)/$(SOFTWARE)-linux-mips-$(VERSION).tar.gz \
		$(SOFTWARE)_linux_mips

release-chksum:
	@echo
	cd $(RELEASE); $(SUM) *
	@echo

install-dep:
	@if ! type glide > /dev/null; then \
		go get -u github.com/Masterminds/glide ; \
	fi
	glide install

gopath-spoof: setup-tmp install-dep
	ln -s $(PWD)/vendor $(BUILD_TMP)/src

setup-tmp:
	rm -rf $(BUILD_TMP)
	mkdir -p $(BUILD_TMP)

clean-tmp:
	rm -rf $(BUILD_TMP)

clean-dep:
	rm -rf $(PWD)/vendor

clean: clean-tmp
	rm -f $(PWD)/$(SOFTWARE)
	rm -rf $(RELEASE)

.PHONY: love \
	release \
	build \
	setup-release \
	build-linux-darwin-windows-freebsd \
	build-arms \
	build-mipsle \
	release-chksum \
	install-dep \
	gopath-spoof \
	setup-tmp \
	clean-tmp \
	clean-dep \
	clean

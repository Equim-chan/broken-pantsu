# Copyright (c) 2017 Equim and other Broken Pantsu contributors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

SOFTWARE := broken-pantsu

READLINK := "$(shell if type greadlink > /dev/null 2>&1 ; then echo greadlink; else echo readlink; fi)"

RELEASE := "$(shell $(READLINK) -f ./release)"
BUILD_TMP := "$(shell $(READLINK) -f ./tmp)"
PWD := "$(shell pwd)"
VERSION := $(shell cat ./semver)+build$(shell date -u +%y%m%d`)
LDFLAGS := "-X main.VERSION=$(VERSION) -s -w"
GCFLAGS := ""

ARCHS := amd64 386
ARMS := 5 6 7
SUM := sha1sum

SHELL := /bin/bash
VERBOSE := -x

love: setup-tmp \
	install-dep \
	gopath-spoof \
	build

	make clean-tmp

release: setup-tmp \
	setup-release \
	install-dep \
	gopath-spoof \
	release-linux \
	release-darwin \
	release-windows \
	release-freebsd \
	release-arms \
	release-mipsle

	make clean-tmp
	make release-chksum

build:
	@GOPATH=$(BUILD_TMP) CGO_ENABLED=0 \
		go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
			$(VERBOSE) \
			-o $(SOFTWARE)

setup-release:
	mkdir -p $(RELEASE)

release-linux: setup-release gopath-spoof
	@for arch in $(ARCHS); do \
		GOPATH=$(BUILD_TMP) CGO_ENABLED=0 \
		GOOS=linux GOARCH=$${arch} \
			go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
				$(VERBOSE) \
				-o $(BUILD_TMP)/$(SOFTWARE)-linux-$${arch} ; \
		cd $(BUILD_TMP) ; \
		tar -zcf \
			$(RELEASE)/$(SOFTWARE)-linux-$${arch}-$(VERSION).tar.gz \
			$(SOFTWARE)-linux-$${arch} ; \
		cd $(PWD) ; \
	done ; \

release-darwin: setup-release gopath-spoof
	@for arch in $(ARCHS); do \
		GOPATH=$(BUILD_TMP) CGO_ENABLED=0 \
		GOOS=darwin GOARCH=$${arch} \
			go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
				$(VERBOSE) \
				-o $(BUILD_TMP)/$(SOFTWARE)-darwin-$${arch} ; \
		cd $(BUILD_TMP) ; \
		tar -zcf \
			$(RELEASE)/$(SOFTWARE)-darwin-$${arch}-$(VERSION).tar.gz \
			$(SOFTWARE)-darwin-$${arch} ; \
		cd $(PWD) ; \
	done ; \

release-windows: setup-release gopath-spoof
	@for arch in $(ARCHS); do \
		GOPATH=$(BUILD_TMP) CGO_ENABLED=0 \
		GOOS=windows GOARCH=$${arch} \
			go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
				$(VERBOSE) \
				-o $(BUILD_TMP)/$(SOFTWARE)-windows-$${arch}.exe ; \
		cd $(BUILD_TMP) ; \
		tar -zcf \
			$(RELEASE)/$(SOFTWARE)-windows-$${arch}-$(VERSION).tar.gz \
			$(SOFTWARE)-windows-$${arch}.exe ; \
		cd $(PWD) ; \
	done ; \

release-freebsd: setup-release gopath-spoof
	@for arch in $(ARCHS); do \
		GOPATH=$(BUILD_TMP) CGO_ENABLED=0 \
		GOOS=freebsd GOARCH=$${arch} \
			go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
				$(VERBOSE) \
				-o $(BUILD_TMP)/$(SOFTWARE)-freebsd-$${arch} ; \
		cd $(BUILD_TMP) ; \
		tar -zcf \
			$(RELEASE)/$(SOFTWARE)-freebsd-$${arch}-$(VERSION).tar.gz \
			$(SOFTWARE)-freebsd-$${arch} ; \
		cd $(PWD) ; \
	done ; \

release-arms: setup-release gopath-spoof
	@for v in $(ARMS); do \
	GOPATH=$(BUILD_TMP) CGO_ENABLED=0 \
	GOOS=linux GOARCH=arm GOARM=$${v} \
		go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
			$(VERBOSE) \
			-o $(BUILD_TMP)/$(SOFTWARE)-linux-arm$${v} ; \
	done ; \
	if hash upx 2>/dev/null; then \
		upx -9 $(SOFTWARE)-linux-arm* ; \
	fi ; \
	cd $(BUILD_TMP) ; \
	tar -zcf \
		$(RELEASE)/$(SOFTWARE)-linux-arm-$(VERSION).tar.gz \
		$(SOFTWARE)-linux-arm*

release-mipsle: setup-release gopath-spoof
	@GOPATH=$(BUILD_TMP) CGO_ENABLED=0 \
	GOOS=linux GOARCH=mipsle \
		go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
			$(VERBOSE) \
			-o $(BUILD_TMP)/$(SOFTWARE)-linux-mipsle ; \
	GOPATH=$(BUILD_TMP) CGO_ENABLED=0 \
	GOOS=linux GOARCH=mipsle \
		go build -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
			$(VERBOSE) \
			-o $(BUILD_TMP)/$(SOFTWARE)-linux-mips ; \
	if hash upx 2>/dev/null; then \
		upx -9 ${SOFTWARE}-linux-mips* ; \
	fi ; \
	cd $(BUILD_TMP) ; \
	tar -zcf \
		$(RELEASE)/$(SOFTWARE)-linux-mipsle-$(VERSION).tar.gz \
		$(SOFTWARE)-linux-mipsle ; \
	tar -zcf \
		$(RELEASE)/$(SOFTWARE)-linux-mips-$(VERSION).tar.gz \
		$(SOFTWARE)-linux-mips

release-chksum:
	@echo
	cd $(RELEASE); $(SUM) *
	@echo

install-dep:
	@if ! type glide > /dev/null 2>&1 ; then \
		if [ ! -d $$GOPATH/bin/glide ]; then \
			go get -u github.com/Masterminds/glide ; \
		fi ; \
		$$GOPATH/bin/glide install ; \
	else \
		glide install ; \
	fi

gopath-spoof: setup-tmp install-dep
	ln -s $(PWD)/vendor $(BUILD_TMP)/src

setup-tmp:
	make clean-tmp
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
	release-linux-darwin-windows-freebsd \
	release-arms \
	release-mipsle \
	release-chksum \
	install-dep \
	gopath-spoof \
	setup-tmp \
	clean-tmp \
	clean-dep \
	clean

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

SHELL := /bin/bash
READLINK := "$(shell if type greadlink > /dev/null 2>&1 ; then echo greadlink; else echo readlink; fi)"

RELEASE_PATH := "$(shell $(READLINK) -f ./release)"
BUILD_TMP := "$(shell $(READLINK) -f ./tmp)"
PWD := "$(shell pwd)"
VERSION := $(shell cat ./semver)+build$(shell date -u +%y%m%d)

GC := go build
LDFLAGS := "-X main.VERSION=$(VERSION) -s -w"
GCFLAGS := ""

SUM := sha1sum

ARCHS := amd64 386
ARMS := 5 6 7

ifdef V
VERBOSE := -x
endif

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
	release-openbsd \
	release-arms \
	release-mipsle

	make clean-tmp
	make release-chksum

build:
	@# use printf instead of echo so that colors can print in windows cmd
	@ \
	printf "    \x1b[32mGOOS=`go env GOOS` \x1b[33mGOARCH=`go env GOARCH` \x1b[36mGC\x1b[0m\n" ; \
	GOPATH=$(BUILD_TMP) CGO_ENABLED=0 \
		$(GC) -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
			$(VERBOSE) \
			-o $(SOFTWARE)

setup-release:
	mkdir -p $(RELEASE_PATH)

release-%: setup-release gopath-spoof
	@ \
	GOOS=$(subst release-,,$@) ; \
	for arch in $(ARCHS); do \
		printf "    \x1b[32mGOOS=$${GOOS} \x1b[33mGOARCH=$${arch} \x1b[36mGC\x1b[0m\n" ; \
		GOPATH=$(BUILD_TMP) CGO_ENABLED=0 GOARCH=$${arch} \
			$(GC) -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
				$(VERBOSE) \
				-o $(BUILD_TMP)/$(SOFTWARE)-$${GOOS}-$${arch} ; \
		cd $(BUILD_TMP) ; \
		tar -zcf \
			$(RELEASE_PATH)/$(SOFTWARE)-$${GOOS}-$${arch}-$(VERSION).tar.gz \
			$(SOFTWARE)-$${GOOS}-$${arch} ; \
		cd $(PWD) ; \
	done ; \

release-arms: setup-release gopath-spoof
	@ \
	for v in $(ARMS); do \
		printf "    \x1b[32mGOOS=linux \x1b[33mGOARCH=arm \x1b[35mGOARM=$${v} \x1b[36mGC\x1b[0m\n" ; \
		GOPATH=$(BUILD_TMP) CGO_ENABLED=0 \
		GOOS=linux GOARCH=arm GOARM=$${v} \
			$(GC) -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
				$(VERBOSE) \
				-o $(BUILD_TMP)/$(SOFTWARE)-linux-arm$${v} ; \
	done ; \
	if hash upx 2>/dev/null; then \
		upx -9 $(SOFTWARE)-linux-arm* ; \
	fi ; \
	cd $(BUILD_TMP) ; \
	tar -zcf \
		$(RELEASE_PATH)/$(SOFTWARE)-linux-arm-$(VERSION).tar.gz \
		$(SOFTWARE)-linux-arm*

release-mipsle: setup-release gopath-spoof
	@ \
	printf "    \x1b[32mGOOS=linux \x1b[33mGOARCH=mipsle \x1b[36mGC\x1b[0m\n" ; \
	GOPATH=$(BUILD_TMP) CGO_ENABLED=0 \
	GOOS=linux GOARCH=mipsle \
		$(GC) -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
			$(VERBOSE) \
			-o $(BUILD_TMP)/$(SOFTWARE)-linux-mipsle ; \
	printf "    \x1b[32mGOOS=linux \x1b[33mGOARCH=mips \x1b[36mGC\x1b[0m\n" ; \
	GOPATH=$(BUILD_TMP) CGO_ENABLED=0 \
	GOOS=linux GOARCH=mips \
		$(GC) -ldflags $(LDFLAGS) -gcflags $(GCFLAGS) \
			$(VERBOSE) \
			-o $(BUILD_TMP)/$(SOFTWARE)-linux-mips ; \
	if hash upx 2>/dev/null; then \
		upx -9 ${SOFTWARE}-linux-mips* ; \
	fi ; \
	cd $(BUILD_TMP) ; \
	tar -zcf \
		$(RELEASE_PATH)/$(SOFTWARE)-linux-mipsle-$(VERSION).tar.gz \
		$(SOFTWARE)-linux-mipsle ; \
	tar -zcf \
		$(RELEASE_PATH)/$(SOFTWARE)-linux-mips-$(VERSION).tar.gz \
		$(SOFTWARE)-linux-mips

release-chksum:
	@echo
	cd $(RELEASE_PATH); $(SUM) *
	@echo

install-dep:
	@ \
	if ! type glide > /dev/null 2>&1 ; then \
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
	rm -rf $(RELEASE_PATH)

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

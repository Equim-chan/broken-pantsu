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
READLINK := $(shell if type greadlink > /dev/null 2>&1 ; then echo greadlink; else echo readlink; fi)

RELEASE_PATH := "$(shell $(READLINK) -f ./release)"
BUILD_TMP := "$(shell $(READLINK) -f ./tmp)"
PWD := "$(shell pwd)"

ARCHS := amd64 386
ARMS := 5 6 7

VERSION := $(shell cat ./semver)+build$(shell date -u +%y%m%d)
LDFLAGS := -X main.VERSION=$(VERSION) -s -w
GCFLAGS :=
ifdef V
VERBOSE := -x
endif

OUT_FILENAME := "$(SOFTWARE)-$${goos}-$${goarch}$${goarm}$${suffix}"

# use printf instead of echo so that colors can print properly in windows cmd
PRINT := \
	if [ "$${goarm}" != "" ]; then \
		printf "    \x1b[32mGOOS=$${goos} \x1b[33mGOARCH=$${goarch} \x1b[35mGOARM=$${goarm} \x1b[1;36mGC\x1b[0m\n" ; \
	else \
		printf "    \x1b[32mGOOS=$${goos} \x1b[33mGOARCH=$${goarch} \x1b[1;36mGC\x1b[0m\n" ; \
	fi

GC := \
	$(PRINT) ; \
	GOPATH=$(BUILD_TMP) CGO_ENABLED=0 \
	GOOS=$${goos} GOARCH=$${goarch} GOARM=$${goarm} \
		go build \
			-ldflags "$(LDFLAGS)" -gcflags "$(GCFLAGS)" \
			$(VERBOSE) \
			-o $(BUILD_TMP)/$(OUT_FILENAME)

PACK := \
	cd $(BUILD_TMP) ; \
	tar -zcf \
		$(RELEASE_PATH)/"$(SOFTWARE)-$${goos}-$${goarch}$${goarm}-$(VERSION).tar.gz" \
		$(OUT_FILENAME) ; \
	cd $(PWD)

SUM := sha1sum

love: setup-tmp \
	install-dep \
	gopath-spoof \
	build-local

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
	release-mips

	make clean-tmp
	make release-chksum

build-local:
	@ \
	goos=`go env GOOS` goarch=`go env GOARCH` goarm=`go env GOARM` suffix=`go env GOEXE` ; \
	$(GC) ; \
	mv $(BUILD_TMP)/$(OUT_FILENAME) $(PWD)/

setup-release:
	rm -rf $(RELEASE_PATH)
	mkdir -p $(RELEASE_PATH)

release-%: setup-release gopath-spoof
	@ \
	goos=$(subst release-,,$@) ; \
	if [ "$${goos}" == "windows" ]; then \
		suffix=".exe" ; \
	fi ; \
	for goarch in $(ARCHS); do \
		$(GC) ; \
		$(PACK) ; \
	done

release-arms: setup-release gopath-spoof
	@ \
	goos=linux goarch=arm64 ; \
	$(GC) ; \
	$(PACK) ; \
	goarch=arm ; \
	for goarm in $(ARMS); do \
		$(GC) ; \
		$(PACK) ; \
	done

release-mips: setup-release gopath-spoof
	@ \
	goos=linux goarch=mipsle ; \
	$(GC) ; \
	$(PACK) ; \
	goos=linux goarch=mips ; \
	$(GC) ; \
	$(PACK)

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
	rm -f $(PWD)/"$(SOFTWARE)-`go env GOOS`-`go env GOARCH``go env GOARM``go env GOEXE`"
	rm -rf $(RELEASE_PATH)

.PHONY: love \
	release \
	build-local \
	setup-release \
	release-% \
	release-chksum \
	install-dep \
	gopath-spoof \
	setup-tmp \
	clean-tmp \
	clean-dep \
	clean

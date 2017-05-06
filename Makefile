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
MAKE := make --no-print-directory

PROJECT_ROOT := "$(shell pwd)"
RELEASE_PATH := $(PROJECT_ROOT)/release
BUILD_TMP := $(PROJECT_ROOT)/tmp
MOCK_GOPATH := $(BUILD_TMP)/gopath

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
	GOPATH=$(MOCK_GOPATH) CGO_ENABLED=0 \
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
	cd $(PROJECT_ROOT)

SUM := sha1sum

love: setup
	@ \
	goos=`go env GOOS` goarch=`go env GOARCH` goarm=`go env GOARM` suffix=`go env GOEXE` ; \
	$(GC) ; \
	mv $(BUILD_TMP)/$(OUT_FILENAME) $(PROJECT_ROOT)/
	@$(MAKE) clean-tmp

release: setup-release \
	release-linux \
	release-darwin \
	release-windows \
	release-freebsd \
	release-openbsd \
	release-arms \
	release-mips

	@$(MAKE) clean-tmp
	@$(MAKE) release-chksum

setup:
	@$(MAKE) clean-tmp
	mkdir -p $(BUILD_TMP)
	@$(MAKE) install-dep
	@$(MAKE) gopath-spoof

setup-release: setup
	rm -rf $(RELEASE_PATH)
	mkdir -p $(RELEASE_PATH)

install-dep:
	@ \
	if ! type glide > /dev/null 2>&1 ; then \
		if [ ! -d "$${GOPATH}/bin/glide" ]; then \
			go get -u github.com/Masterminds/glide ; \
		fi ; \
		"$${GOPATH}/bin/glide" install ; \
	else \
		glide install ; \
	fi

gopath-spoof:
	mkdir -p $(MOCK_GOPATH)
	ln -sf $(PROJECT_ROOT)/vendor $(MOCK_GOPATH)/src

release-%: setup-release
	@ \
	goos=$(subst release-,,$@) ; \
	if [ "$${goos}" == "windows" ]; then \
		suffix=".exe" ; \
	fi ; \
	for goarch in $(ARCHS); do \
		$(GC) ; \
		$(PACK) ; \
	done

release-arms: setup-release
	@ \
	goos=linux goarch=arm64 ; \
	$(GC) ; \
	$(PACK) ; \
	goarch=arm ; \
	for goarm in $(ARMS); do \
		$(GC) ; \
		$(PACK) ; \
	done

release-mips: setup-release
	@ \
	goos=linux goarch=mipsle ; \
	$(GC) ; \
	$(PACK) ; \
	goos=linux goarch=mips ; \
	$(GC) ; \
	$(PACK)

release-chksum:
	@ \
	cd $(RELEASE_PATH) ; \
	$(SUM) * > ./chksum.txt ; \
	echo ; \
	cat ./chksum.txt ; \
	cd $(PROJECT_ROOT)

clean-tmp:
	rm -rf $(BUILD_TMP)

clean-dep:
	rm -rf $(PROJECT_ROOT)/vendor

clean: clean-tmp
	rm -f $(PROJECT_ROOT)/"$(SOFTWARE)-$(shell go env GOOS)-$(shell go env GOARCH)$(shell go env GOARM)$(shell go env GOEXE)"
	rm -rf $(RELEASE_PATH)

.PHONY: love \
	release \
	setup \
	setup-release \
	install-dep \
	gopath-spoof \
	release-% \
	clean-tmp \
	clean-dep \
	clean

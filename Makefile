BINARY := hitmaker
BUILD_DIR ?= bin
INSTALL_DIR ?= $(HOME)/.local/bin
VERSION ?= $(shell node -p "require('./npm/package.json').version")
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build release-build release-check install-local test fmt vet clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./cmd/hitmaker

release-build:
	VERSION="$(VERSION)" scripts/build-release.sh

release-check: test release-build
	cd npm && npm pack --dry-run

install-local:
	HITMAKER_INSTALL_DIR="$(INSTALL_DIR)" scripts/install-local.sh

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)

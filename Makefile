BINARY := hitmaker
BUILD_DIR ?= bin
INSTALL_DIR ?= $(HOME)/.local/bin

.PHONY: build install-local test fmt vet clean

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/hitmaker

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

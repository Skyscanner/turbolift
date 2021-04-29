NAME=turbolift
ARCH=$(shell uname -m)
VERSION=0.0.1
ITERATION := 1

GOLANGCI_VERSION = 1.32.0

BIN_DIR := $(CURDIR)/bin

$(BIN_DIR)/golangci-lint: $(BIN_DIR)/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} $(BIN_DIR)/golangci-lint
$(BIN_DIR)/golangci-lint-${GOLANGCI_VERSION}:
	@curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | BINARY=golangci-lint bash -s -- v${GOLANGCI_VERSION}
	@mv $(BIN_DIR)/golangci-lint $@

mod:
	@go mod download
	@go mod tidy
.PHONY: mod

lint: $(BIN_DIR)/golangci-lint
	@echo "--- lint all the things"
	@$(BIN_DIR)/golangci-lint run ./...
.PHONY: lint

fix: $(BIN_DIR)/golangci-lint
	@echo "--- Formatting all the things"
	@$(BIN_DIR)/golangci-lint run --fix ./...
.PHONY: lint-fix

fmt: fix

build:
	@go build .
.PHONY: build

install:
	@go install .
.PHONY: install

clean:
	@rm -fr ./dist
.PHONY: clean

test: lint
	@echo "--- test all the things"
	@go test -race -cover ./...
.PHONY: test


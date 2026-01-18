GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet

BINARY_NAME=mtc

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)
SHELL  := /bin/bash

# build args - evaluated at build time
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags="-X 'github.com/lucho00cuba/mtc/version.VERSION=$(VERSION)' -X 'github.com/lucho00cuba/mtc/version.COMMIT=$(COMMIT)' -X 'github.com/lucho00cuba/mtc/version.DATE=$(DATE)'"

all: help

## Build:
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

build: ## Build your project for all supported platforms
	mkdir -p dist/
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		OUTPUT=dist/${BINARY_NAME}-$${GOOS}-$${GOARCH}; \
		if [ "$${GOOS}" = "windows" ]; then OUTPUT=$${OUTPUT}.exe; fi; \
		echo "🚀 Building $${GOOS}/$${GOARCH}... 🏗️"; \
		GOOS=$${GOOS} GOARCH=$${GOARCH} go build $(LDFLAGS) -trimpath -o $${OUTPUT} . || exit 1; \
	done
	@echo "✅ All builds completed successfully"

vendor: ## Copy of all packages needed to support builds and tests in the vendor directory
	@$(GOCMD) mod vendor

## Test:
test: ## Run tests
	@$(GOCMD) test ./... -v

test-coverage: clean-test ## Run tests with coverage report for internal packages
	@$(GOCMD) test ./... -coverprofile=coverage.out -covermode=atomic
	@$(GOCMD) tool cover -func=coverage.out
	@rm -f coverage.out

test-race: ## Run tests with race detection
	@$(GOCMD) test ./... -race -v

clean-test: ## Clean up test artifacts
	@rm -f coverage.out
	@find . -name "*.test" -type f -delete 2>/dev/null || true

## Lint:
lint: ## Run lint
	@$(GOCMD) run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run ./...

lint-fix: ## Run lint and auto-fix issues
	@$(GOCMD) run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --fix ./...

## Format:
format: ## Run format
	@$(GOCMD) fmt ./...

## Clean:
clean: clean-test ## Clean up test artifacts and build files
	@echo "🧹 Cleaning up..."
	@rm -rf dist/
	@echo "✅ Cleanup complete"

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
	        if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
	        else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
	        }' $(MAKEFILE_LIST)

.PHONY: all build test clean lint tools init generate docs

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=hephaestus

# Build parameters
BUILD_DIR=build
MAIN_PACKAGE=./cmd/init

all: test build

build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

test:
	$(GOTEST) -v ./...

test-race:
	$(GOTEST) -v -race ./...

test-cover:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

bench:
	$(GOTEST) -v -bench=. -benchmem ./...

clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

lint:
	golangci-lint run

tools:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) golang.org/x/tools/cmd/goimports@latest
	$(GOGET) github.com/golang/mock/mockgen@latest

init: tools
	cp config/config.example.yaml config/config.yaml
	git config core.hooksPath .githooks

tidy:
	$(GOMOD) tidy
	$(GOMOD) verify

generate:
	$(GOCMD) generate ./...

docs:
	$(GOCMD) doc -all > docs/api.txt

# Development helpers
fmt:
	gofmt -s -w .
	goimports -w .

check: lint test

# Docker targets
docker-build:
	docker build -t $(BINARY_NAME) .

docker-run:
	docker run $(BINARY_NAME)

# Release targets
release:
	goreleaser release --snapshot --rm-dist

# Dependency management
deps-update:
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Database migrations
migrate-up:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go down

# Development server
dev:
	go run $(MAIN_PACKAGE) -config config/config.yaml

# Help target
help:
	@echo "Available targets:"
	@echo "  all          : Run tests and build"
	@echo "  build        : Build the binary"
	@echo "  test         : Run tests"
	@echo "  test-race    : Run tests with race detector"
	@echo "  test-cover   : Run tests with coverage"
	@echo "  bench        : Run benchmarks"
	@echo "  clean        : Clean build artifacts"
	@echo "  lint         : Run linter"
	@echo "  tools        : Install development tools"
	@echo "  init         : Initialize development environment"
	@echo "  tidy         : Tidy and verify go.mod"
	@echo "  generate     : Run go generate"
	@echo "  docs         : Generate documentation"
	@echo "  fmt          : Format code"
	@echo "  check        : Run linter and tests"
	@echo "  docker-build : Build Docker image"
	@echo "  docker-run   : Run Docker container"
	@echo "  release      : Create a release"
	@echo "  deps-update  : Update dependencies"
	@echo "  dev          : Run development server" 
# Makefile for Hephaestus project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=hephaestus
VERSION ?= $(shell git describe --tags --always --dirty)

# Test parameters
COVERAGE_THRESHOLD=80
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Protobuf parameters
PROTOC=protoc
PROTO_DIR=proto
PROTO_GO_OUT=pkg/proto

# Docker parameters
DOCKER=docker
DOCKER_IMAGE=hephaestus
DOCKER_TAG=latest

# Directories
CMD_DIR=cmd
PKG_DIR=pkg
INTERNAL_DIR=internal
EXAMPLES_DIR=examples
TEST_DIR=test

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

.PHONY: all build clean test coverage deps proto docker-build docker-push run help generate-mocks docs check-coverage

all: check deps proto test build ## Build everything

build: test ## Build the binary
	$(GOBUILD) -o $(BINARY_NAME) $(LDFLAGS) ./$(CMD_DIR)/server

clean: ## Clean build files
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(COVERAGE_FILE)
	rm -f $(COVERAGE_HTML)

test: unit-test integration-test ## Run all tests

unit-test: ## Run unit tests
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) ./internal/... ./pkg/...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Checking test coverage..."
	@coverage=$$(go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $${coverage%.*} -lt $(COVERAGE_THRESHOLD) ]; then \
		echo "Test coverage is below $(COVERAGE_THRESHOLD)%. Current coverage: $$coverage%"; \
		exit 1; \
	fi

integration-test: ## Run integration tests
	$(GOTEST) -v -tags=integration ./test/integration/...

performance-test: ## Run performance tests
	$(GOTEST) -v -tags=performance ./test/performance/...

check-coverage: ## Check test coverage
	@go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}'

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy
	$(GOMOD) verify

proto: ## Generate Protocol Buffer code
	mkdir -p $(PROTO_GO_OUT)
	$(PROTOC) --go_out=$(PROTO_GO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_GO_OUT) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto

docker-build: test ## Build Docker image
	$(DOCKER) build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-push: ## Push Docker image
	$(DOCKER) push $(DOCKER_IMAGE):$(DOCKER_TAG)

run: ## Run the server
	$(GOBUILD) -o $(BINARY_NAME) $(LDFLAGS) ./$(CMD_DIR)/server
	./$(BINARY_NAME)

generate-mocks: ## Generate mock interfaces
	@for pkg in $$(go list ./internal/... ./pkg/...); do \
		if grep -q "interface {" $$(echo $$pkg | sed 's|github.com/HoyeonS/hephaestus/||')/**.go; then \
			mockgen -source=$$(echo $$pkg | sed 's|github.com/HoyeonS/hephaestus/||')/**.go \
				-destination=$$(echo $$pkg | sed 's|github.com/HoyeonS/hephaestus/||')/mocks/mock_$$(basename $$pkg).go \
				-package=mocks; \
		fi \
	done

docs: ## Generate API documentation
	swag init -g $(CMD_DIR)/server/main.go -o docs/swagger

lint: ## Run linters
	golangci-lint run --timeout=5m ./...

fmt: ## Format code
	gofmt -s -w .
	goimports -w .

vet: ## Run go vet
	$(GOCMD) vet ./...

check: fmt vet lint test ## Run all checks

bench: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

update-deps: ## Update dependencies
	$(GOGET) -u ./...
	$(GOMOD) tidy
	$(GOMOD) verify

install: test ## Install binary
	$(GOBUILD) -o $(GOPATH)/bin/$(BINARY_NAME) $(LDFLAGS) ./$(CMD_DIR)/server

uninstall: ## Uninstall binary
	rm -f $(GOPATH)/bin/$(BINARY_NAME)

# Development tools installation
tools: ## Install development tools
	$(GOGET) github.com/golang/mock/mockgen
	$(GOGET) github.com/swaggo/swag/cmd/swag
	$(GOGET) golang.org/x/tools/cmd/goimports
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) github.com/stretchr/testify
	$(GOGET) github.com/prometheus/client_golang/prometheus

# Help target
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# Version information
version: ## Display version information
	@echo "Version: $(VERSION)"

# Default target
.DEFAULT_GOAL := help 
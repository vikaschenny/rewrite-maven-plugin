# Makefile for rewrite-go
# A Go implementation of the OpenRewrite Maven plugin

# Variables
BINARY_NAME=rewrite-go
MAIN_PACKAGE=.
BUILD_DIR=build
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/$(BUILD_DIR)
GOFILES=$(wildcard *.go)

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)"

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: help build clean install test run dry-run discover deps fmt vet lint tidy

## help: Show this help message
help: Makefile
	@echo "Available commands:"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary
build: deps
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean

## install: Install the binary to GOPATH/bin
install: deps
	@echo "Installing $(BINARY_NAME)..."
	@go install $(LDFLAGS) $(MAIN_PACKAGE)
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

## run: Build and run the tool with default arguments
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME) run --verbose

## dry-run: Build and run the tool in dry-run mode
dry-run: build
	@echo "Running $(BINARY_NAME) in dry-run mode..."
	@$(BUILD_DIR)/$(BINARY_NAME) dry-run --verbose

## discover: Build and run recipe discovery
discover: build
	@echo "Discovering recipes..."
	@$(BUILD_DIR)/$(BINARY_NAME) discover

## deps: Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

## lint: Run golint (if available)
lint:
	@echo "Running golint..."
	@which golint > /dev/null 2>&1 || (echo "golint not installed, skipping..."; exit 0)
	@golint ./...

## tidy: Tidy go modules
tidy:
	@echo "Tidying modules..."
	@go mod tidy

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "All checks passed!"

## release-build: Build for multiple platforms
release-build: clean deps
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux amd64
	@echo "Building for Linux amd64..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	
	# Linux arm64
	@echo "Building for Linux arm64..."
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	
	# macOS amd64
	@echo "Building for macOS amd64..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	
	# macOS arm64 (Apple Silicon)
	@echo "Building for macOS arm64..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	
	# Windows amd64
	@echo "Building for Windows amd64..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	
	@echo "Release builds completed in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

## example: Run the tool on the example configuration
example: build
	@echo "Running example configuration..."
	@$(BUILD_DIR)/$(BINARY_NAME) run --config example-rewrite.yml --dry-run --verbose

## version: Show version information
version: build
	@$(BUILD_DIR)/$(BINARY_NAME) version

## docker-build: Build Docker image (if Dockerfile exists)
docker-build:
	@if [ -f Dockerfile ]; then \
		echo "Building Docker image..."; \
		docker build -t $(BINARY_NAME):$(VERSION) .; \
	else \
		echo "Dockerfile not found, skipping Docker build"; \
	fi

## all: Run checks and build
all: check build
	@echo "Build completed successfully!"

# Default target
default: help 
# DocChain Makefile

# Variables
BINARY_NAME=docchain
SERVER_BINARY=docchain-server
CLI_BINARY=docchain-cli
GO=go
GOFMT=gofmt
GOFILES=$(shell find . -name "*.go" -type f)
SERVER_MAIN=./cmd/server
CLI_MAIN=./cmd/cli

# Default target
.PHONY: all
all: clean fmt test build

# Build the application
.PHONY: build
build: build-server build-cli

# Build the server
.PHONY: build-server
build-server:
	@echo "Building server..."
	$(GO) build -o $(SERVER_BINARY) $(SERVER_MAIN)

# Build the CLI
.PHONY: build-cli
build-cli:
	@echo "Building CLI..."
	$(GO) build -o $(CLI_BINARY) $(CLI_MAIN)

# Run the server
.PHONY: run
run: build-server
	@echo "Running server..."
	./$(SERVER_BINARY)

# Run the CLI
.PHONY: cli
cli: build-cli
	@echo "Running CLI..."
	./$(CLI_BINARY) $(ARGS)

# Format the code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w $(GOFILES)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(SERVER_BINARY) $(CLI_BINARY) coverage.out coverage.html

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GO) mod download

# Update dependencies
.PHONY: update-deps
update-deps:
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy

# Help
.PHONY: help
help:
	@echo "DocChain Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build        Build the server and CLI"
	@echo "  make build-server Build only the server"
	@echo "  make build-cli    Build only the CLI"
	@echo "  make run          Run the server"
	@echo "  make cli ARGS=\"\" Run the CLI with arguments"
	@echo "  make test         Run tests"
	@echo "  make fmt          Format code"
	@echo "  make clean        Clean build artifacts"
	@echo "  make deps         Install dependencies"
	@echo "  make update-deps  Update dependencies"
	@echo "  make help         Show this help"

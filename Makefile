# Makefile for escope project

# Binary name
BINARY_NAME=escope

# Elasticsearch configuration (loaded from active host in local config file)
ACTIVE_HOST=$(shell grep "active_host:" ~/.escope.yaml | awk '{print $$2}')
PROD_ES_HOST=$(shell grep -A 4 "$(ACTIVE_HOST):" ~/.escope.yaml | grep "host:" | awk '{print $$2}')
PROD_ES_USER=$(shell grep -A 4 "$(ACTIVE_HOST):" ~/.escope.yaml | grep "username:" | awk '{print $$2}')
PROD_ES_PASS=$(shell grep -A 4 "$(ACTIVE_HOST):" ~/.escope.yaml | grep "password:" | awk '{print $$2}')

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod


.PHONY: all build clean deps fmt lint help install run test-commands

# Default target
all: clean build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

# Install dependencies
tidy:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "Code formatting complete"

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run
	@echo "Linting complete"

# Install the application
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Development mode - run with hot reload (requires air)
dev:
	@echo "Starting development server..."
	air

# Create release build
release: clean
	@echo "Creating release build..."
	CGO_ENABLED=0 $(GOBUILD) -a -installsuffix cgo -o $(BINARY_NAME) .
	@echo "Release build complete"


test-commands: build
	@echo "Testing all escope commands..."

	@echo "1. Testing root command (connection check)..."
	-./$(BINARY_NAME) --host $(PROD_ES_HOST) --username $(PROD_ES_USER) --password "$(PROD_ES_PASS)" --secure
	@echo ""
	@echo "2. Testing basic commands..."
	-./$(BINARY_NAME) --help
	@echo ""
	@echo "3. Testing config commands..."
	-./$(BINARY_NAME) config list
	@echo ""
	@echo "3b. Testing config current command..."
	-./$(BINARY_NAME) config current
	@echo ""
	@echo "3c. Testing config timeout command..."
	-./$(BINARY_NAME) config timeout
	@echo ""
	@echo "3d. Testing config get command..."
	-./$(BINARY_NAME) config get
	@echo ""
	@echo "3e. Testing config switch command..."
	-./$(BINARY_NAME) config switch $(ACTIVE_HOST)
	@echo ""
	@echo "4. Testing cluster command..."
	-./$(BINARY_NAME) cluster
	@echo ""
	@echo "5. Testing check command..."
	-./$(BINARY_NAME) check
	@echo ""
	@echo "5b. Testing check command with duration flag..."
	-./$(BINARY_NAME) check --duration 10s
	@echo ""
	@echo "5c. Testing check command with interval flag..."
	-./$(BINARY_NAME) check --duration 6s --interval 2s
	@echo ""
	@echo "6. Testing node command..."
	-./$(BINARY_NAME) node
	@echo ""
	@echo "7. Testing node gc command..."
	-./$(BINARY_NAME) node gc
	@echo ""
	@echo "7b. Testing node gc command with name flag..."
	-./$(BINARY_NAME) node gc --name="*"
	@echo ""
	@echo "8. Testing index command..."
	-./$(BINARY_NAME) index
	@echo ""
	@echo "8b. Testing index command with name flag..."
	-./$(BINARY_NAME) index --name="*"
	@echo ""
	@echo "8c. Testing index command with top flag..."
	-timeout 5s ./$(BINARY_NAME) index --name="*" --top
	@echo ""
	@echo "9. Testing index system command..."
	-./$(BINARY_NAME) index system
	@echo ""
	@echo "10. Testing index sort command..."
	-./$(BINARY_NAME) index sort size
	@echo ""
	@echo "11. Testing shard command..."
	-./$(BINARY_NAME) shard
	@echo ""
	@echo "12. Testing shard dist command..."
	-./$(BINARY_NAME) shard dist
	@echo ""
	@echo "13. Testing shard system command..."
	-./$(BINARY_NAME) shard system
	@echo ""
	@echo "14. Testing shard sort command..."
	-./$(BINARY_NAME) shard sort size
	@echo ""
	@echo "15. Testing segments command..."
	-./$(BINARY_NAME) segments
	@echo ""
	@echo "16. Testing lucene command..."
	-./$(BINARY_NAME) lucene
	@echo ""
	@echo "16b. Testing lucene command with name flag..."
	-./$(BINARY_NAME) lucene --name="*"
	@echo ""
	@echo "17. Testing node dist command..."
	-./$(BINARY_NAME) node dist
	@echo ""
	@echo "18. Testing termvectors command..."
	-./$(BINARY_NAME) termvectors test-index test-doc --fields content,title
	@echo ""
	@echo "19. Testing analyze command..."
	-./$(BINARY_NAME) analyze standard "Hello World Test"
	@echo ""
	@echo "19b. Testing analyze command with tokenizer type..."
	-./$(BINARY_NAME) analyze whitespace "Hello World Test" --type tokenizer
	@echo ""
	@echo "20. Testing upgrade command..."
	-./$(BINARY_NAME) upgrade
	@echo ""
	@echo "All commands tested!"

# Show help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  clean        - Clean build artifacts"
	@echo "  tidy         - Install dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code (requires golangci-lint)"
	@echo "  install      - Install the application"
	@echo "  run          - Build and run the application"
	@echo "  dev          - Run with hot reload (requires air)"
	@echo "  release      - Create release build"
	@echo "  test-commands - Test all commands (including upgrade)"
	@echo "  help         - Show this help message"

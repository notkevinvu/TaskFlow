.PHONY: help run build test clean deps

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download Go dependencies
	go mod download
	go mod tidy

run: ## Run the application
	go run cmd/server/main.go

build: ## Build the application
	go build -o bin/server cmd/server/main.go

test: ## Run unit tests
	go test -v -cover ./...

test-coverage: ## Run tests with coverage report
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

.DEFAULT_GOAL := help

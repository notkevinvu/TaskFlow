.PHONY: help run build test migrate-up migrate-down migrate-create clean deps

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

migrate-up: ## Run database migrations up
	migrate -path migrations -database "postgres://taskflow_user:taskflow_dev_password@localhost:5432/taskflow?sslmode=disable" up

migrate-down: ## Run database migrations down
	migrate -path migrations -database "postgres://taskflow_user:taskflow_dev_password@localhost:5432/taskflow?sslmode=disable" down

migrate-create: ## Create a new migration (usage: make migrate-create name=migration_name)
	migrate create -ext sql -dir migrations -seq $(name)

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	docker build -t taskflow-backend .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env taskflow-backend

.DEFAULT_GOAL := help

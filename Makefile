# Makefile for MongoDB Change Stream PoC

.PHONY: help build run-inserter run-watcher clean docker-up docker-down docker-logs

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build both applications
	@echo "Building inserter..."
	@go build -o bin/inserter ./inserter
	@echo "Building watcher..."
	@go build -o bin/watcher ./watcher
	@echo "Build complete!"

run-inserter: ## Run the data inserter application
	@echo "Starting inserter..."
	@go run inserter/main.go

run-watcher: ## Run the change stream watcher application
	@echo "Starting watcher..."
	@go run watcher/*.go

clean: ## Clean built binaries
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "Clean complete!"

reset-token: ## Reset the resume token to start fresh
	@echo "Resetting resume token..."
	@bash reset-resume-token.sh
	@echo "Resume token reset complete!"

health-check: ## Run health check on the watcher
	@echo "Checking watcher health..."
	@bash health-check.sh

check-metrics: ## Display the current metrics
	@echo "Current watcher metrics:"
	@cat watcher_metrics.json 2>/dev/null || echo "No metrics file found!"

deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	@go mod tidy
	@go mod download
	@echo "Dependencies updated!"

test: ## Run tests (when tests are added)
	@go test ./...

# Docker commands
docker-up: ## Start MongoDB with Docker Compose
	@echo "Starting MongoDB with Docker Compose..."
	@docker-compose up -d mongodb mongodb-setup
	@echo "Waiting for MongoDB to be ready..."
	@sleep 10
	@echo "MongoDB is ready! Connection string: mongodb://admin:password123@localhost:27017/changestream_poc?authSource=admin&replicaSet=rs0"

docker-down: ## Stop and remove Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose down -v
	@echo "Docker containers stopped and volumes removed!"

docker-logs: ## Show MongoDB logs
	@docker-compose logs -f mongodb

docker-build-apps: ## Build Go applications in Docker
	@echo "Building Go applications..."
	@docker-compose build watcher inserter

docker-run-apps: ## Run both applications in Docker
	@echo "Starting both applications in Docker..."
	@docker-compose up watcher inserter

docker-full: docker-up docker-run-apps ## Start MongoDB and run both applications in Docker

# Development with Docker
dev-with-docker: docker-up ## Start MongoDB in Docker and run Go apps locally
	@echo "MongoDB is running in Docker. You can now run the Go applications locally."
	@echo "Use 'make run-watcher' and 'make run-inserter' in separate terminals"
	@echo "Or use the VS Code launch configuration"

.DEFAULT_GOAL := help

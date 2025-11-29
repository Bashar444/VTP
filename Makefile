.PHONY: help setup build run stop logs clean test migrate docker-up docker-down

help:
	@echo "VTP Platform - Development Commands"
	@echo "===================================="
	@echo "make setup        - Install dependencies and set up local environment"
	@echo "make build        - Build the Go binary"
	@echo "make run          - Run the application locally (requires PostgreSQL running)"
	@echo "make docker-up    - Start all services with Docker Compose"
	@echo "make docker-down  - Stop all Docker services"
	@echo "make logs         - View Docker logs"
	@echo "make migrate      - Run database migrations"
	@echo "make test         - Run unit tests"
	@echo "make clean        - Clean build artifacts"
	@echo "make fmt          - Format Go code"
	@echo "make lint         - Run Go linter"

setup:
	@echo "Setting up VTP Platform..."
	@go mod download
	@go mod tidy
	@cp .env.example .env
	@echo "✓ Setup complete. Edit .env with your configuration."

build:
	@echo "Building VTP Platform..."
	@go build -o ./bin/vtp ./cmd/main.go
	@echo "✓ Build complete: ./bin/vtp"

run: build
	@echo "Starting VTP Platform..."
	@./bin/vtp

docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d
	@echo "✓ Services started:"
	@echo "  - API: http://localhost:8080"
	@echo "  - Mediasoup: http://localhost:3000"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Redis: localhost:6379"
	@echo "  - MinIO Console: http://localhost:9001"
	@echo "  - pgAdmin: http://localhost:5050"

docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down
	@echo "✓ Services stopped"

logs:
	@docker-compose logs -f

stop:
	@docker-compose stop

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf ./bin
	@go clean
	@echo "✓ Clean complete"

test:
	@echo "Running tests..."
	@go test ./...

migrate:
	@echo "Running database migrations..."
	@go run ./cmd/main.go migrate

fmt:
	@echo "Formatting code..."
	@go fmt ./...

lint:
	@echo "Running linter..."
	@go vet ./...

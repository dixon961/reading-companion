# Makefile for Interactive Reading Companion

# Default target
.PHONY: help
help:
	@echo "Interactive Reading Companion - Makefile"
	@echo "Available commands:"
	@echo "  build          - Build the entire project"
	@echo "  run            - Run the entire project"
	@echo "  stop           - Stop the project"
	@echo "  test           - Run all tests"
	@echo "  lint           - Run linters"
	@echo "  run-backend    - Run backend service"
	@echo "  run-frontend   - Run frontend service"
	@echo "  migrate-up     - Apply database migrations"
	@echo "  migrate-down   - Rollback database migrations"

# Build targets
.PHONY: build
build:
	@echo "Building the project..."
	docker compose build

# Generate targets
.PHONY: generate
generate:
	@echo "Generating code..."
	cd backend && sqlc generate

# Run targets
.PHONY: run
run:
	@echo "Running the project with Docker Compose..."
	docker compose up -d

.PHONY: stop
stop:
	@echo "Stopping the project..."
	docker compose down

# Production targets
.PHONY: build-prod
build-prod:
	@echo "Building production images..."
	docker compose -f docker-compose.prod.yml build

.PHONY: run-prod
run-prod:
	@echo "Running production environment..."
	docker compose -f docker-compose.prod.yml up -d

.PHONY: stop-prod
stop-prod:
	@echo "Stopping production environment..."
	docker compose -f docker-compose.prod.yml down

# Test targets
.PHONY: test
test:
	@echo "Running tests..."
	cd backend && go test ./...

# Lint targets
.PHONY: lint
lint:
	@echo "Running linters..."

# Backend targets
.PHONY: run-backend
run-backend:
	@echo "Running backend service..."
	cd backend && go run cmd/app/main.go

# Frontend targets
.PHONY: run-frontend
run-frontend:
	@echo "Running frontend service..."
	cd frontend && npm run dev

# Frontend build target
.PHONY: build-frontend
build-frontend:
	@echo "Building frontend..."
	cd frontend && npm run build

# Database migration targets
.PHONY: migrate-up
migrate-up:
	@echo "Applying database migrations..."
	migrate -path backend/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable" -verbose up

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back database migrations..."
	migrate -path backend/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable" -verbose down
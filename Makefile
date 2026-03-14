# Call Notes AI Service Makefile

.PHONY: all build run test test-coverage test-short lint vet check fmt clean clean-all \
        docker-up docker-down docker-clean docker-status docker-logs docker-build docker-run \
        migrate-up migrate-down migrate-down-one migrate-create migrate-version migrate-force \
        mock mock-clean deps deps-install trace-up trace-down verify setup help

BINARY_NAME=call-notes-ai-service
BUILD_DIR=bin
MIGRATION_DIR=internal/database/migrations
DOCKER_COMPOSE=deployment/dev/docker-compose.yml
DOCKER_IMAGE=call-notes-ai-service

DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= callnotes
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

GREEN  := \033[0;32m
YELLOW := \033[0;33m
RED    := \033[0;31m
NC     := \033[0m

all: lint vet test build

## ==================== Build & Run ====================

build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

run:
	@echo "$(GREEN)Running $(BINARY_NAME)...$(NC)"
	@go run ./cmd/api

run-dev:
	@echo "$(GREEN)Running $(BINARY_NAME) with hot reload...$(NC)"
	@air

## ==================== Testing ====================

test:
	@echo "$(GREEN)Running tests...$(NC)"
	@go test -v -race ./...

test-short:
	@echo "$(GREEN)Running tests (short)...$(NC)"
	@go test -race ./...

test-coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -1
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"

## ==================== Code Quality ====================

lint:
	@echo "$(GREEN)Running linter...$(NC)"
	@golangci-lint run ./...

vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	@go vet ./...

check: vet lint test
	@echo "$(GREEN)All checks passed$(NC)"

fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	@gofmt -s -w .
	@goimports -w . 2>/dev/null || true

verify:
	@echo "$(GREEN)Verifying dependencies...$(NC)"
	@go mod verify

## ==================== Clean ====================

clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

clean-all: clean mock-clean

## ==================== Docker ====================

docker-up:
	@echo "$(GREEN)Starting Docker containers...$(NC)"
	@docker-compose -f $(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Waiting for services to be ready...$(NC)"
	@sleep 5
	@make docker-status

docker-down:
	@echo "$(YELLOW)Stopping Docker containers...$(NC)"
	@docker-compose -f $(DOCKER_COMPOSE) down

docker-clean:
	@echo "$(RED)Stopping Docker containers and removing volumes...$(NC)"
	@docker-compose -f $(DOCKER_COMPOSE) down -v

docker-status:
	@echo "$(GREEN)Docker container status:$(NC)"
	@docker-compose -f $(DOCKER_COMPOSE) ps

docker-logs:
	@docker-compose -f $(DOCKER_COMPOSE) logs -f

docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	@docker build -t $(DOCKER_IMAGE):latest .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):latest$(NC)"

docker-run:
	@echo "$(GREEN)Running Docker container...$(NC)"
	@docker run --rm -p 8080:8080 -p 8081:8081 --env-file .env $(DOCKER_IMAGE):latest

## ==================== Database ====================

migrate-up:
	@echo "$(GREEN)Running migrations...$(NC)"
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up
	@echo "$(GREEN)Migrations complete$(NC)"

migrate-down:
	@echo "$(YELLOW)Rolling back all migrations...$(NC)"
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down -all

migrate-down-one:
	@echo "$(YELLOW)Rolling back one migration...$(NC)"
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down 1

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $$name

migrate-version:
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" version

migrate-force:
	@read -p "Enter version to force: " version; \
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" force $$version

## ==================== Tracing ====================

trace-up:
	@echo "$(GREEN)Starting Jaeger for local tracing...$(NC)"
	@docker run -d --name jaeger \
		-p 16686:16686 \
		-p 4317:4317 \
		jaegertracing/all-in-one:latest
	@echo "$(GREEN)Jaeger UI: http://localhost:16686$(NC)"

trace-down:
	@echo "$(YELLOW)Stopping Jaeger...$(NC)"
	@docker rm -f jaeger 2>/dev/null || true

## ==================== Mock Generation ====================

mock: mock-clean
	@echo "$(GREEN)Generating mocks...$(NC)"
	@echo "  Generating session mocks..."
	@mkdir -p internal/modules/session/mock
	@mockgen -source=internal/modules/session/repository.go -destination=internal/modules/session/mock/mock_repository.go -package=mock
	@mockgen -source=internal/modules/session/core.go -destination=internal/modules/session/mock/mock_core.go -package=mock
	@echo "  Generating extraction mocks..."
	@mkdir -p internal/modules/extraction/mock
	@mockgen -source=internal/modules/extraction/core.go -destination=internal/modules/extraction/mock/mock_core.go -package=mock
	@echo "  Generating database pool mock..."
	@mkdir -p pkg/database/mock
	@mockgen -source=pkg/database/pool.go -destination=pkg/database/mock/mock_pool.go -package=mock
	@echo "  Generating LLM client mocks..."
	@mkdir -p internal/services/llm/mock
	@mockgen -source=internal/services/llm/client.go -destination=internal/services/llm/mock/mock_client.go -package=mock
	@echo "  Generating Deepgram client mocks..."
	@mkdir -p internal/services/deepgram/mock
	@mockgen -source=internal/services/deepgram/client.go -destination=internal/services/deepgram/mock/mock_client.go -package=mock
	@echo "  Generating SFDC client mocks..."
	@mkdir -p internal/services/sfdc/mock
	@mockgen -source=internal/services/sfdc/client.go -destination=internal/services/sfdc/mock/mock_client.go -package=mock
	@echo "$(GREEN)Mocks generated successfully$(NC)"

mock-clean:
	@echo "$(YELLOW)Cleaning generated mocks...$(NC)"
	@rm -f internal/modules/session/mock/*.go
	@rm -f internal/modules/extraction/mock/*.go
	@rm -f internal/services/llm/mock/*.go
	@rm -f internal/services/deepgram/mock/*.go
	@rm -f internal/services/sfdc/mock/*.go
	@rm -f pkg/database/mock/*.go

## ==================== Dependencies ====================

deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@go mod download
	@go mod tidy

deps-install:
	@echo "$(GREEN)Installing development tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install go.uber.org/mock/mockgen@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "$(GREEN)Development tools installed$(NC)"

## ==================== Setup ====================

setup:
	@echo ""
	@echo "$(GREEN)============================================$(NC)"
	@echo "$(GREEN)  Call Notes AI Service - Setup$(NC)"
	@echo "$(GREEN)============================================$(NC)"
	@echo ""
	@echo "$(GREEN)[1/5] Installing development tools...$(NC)"
	@$(MAKE) deps-install --no-print-directory
	@echo ""
	@echo "$(GREEN)[2/5] Downloading dependencies...$(NC)"
	@$(MAKE) deps --no-print-directory
	@echo ""
	@echo "$(GREEN)[3/5] Starting infrastructure...$(NC)"
	@$(MAKE) docker-up --no-print-directory
	@echo ""
	@echo "$(GREEN)[4/5] Running database migrations...$(NC)"
	@sleep 5
	@$(MAKE) migrate-up --no-print-directory
	@echo ""
	@echo "$(GREEN)[5/5] Building service...$(NC)"
	@$(MAKE) build --no-print-directory
	@echo ""
	@echo "$(GREEN)============================================$(NC)"
	@echo "$(GREEN)  Setup complete!$(NC)"
	@echo "$(GREEN)============================================$(NC)"
	@echo ""
	@echo "$(GREEN)Next steps:$(NC)"
	@echo "  $(YELLOW)make run$(NC)   - Start the service (API :8080, Ops :8081)"
	@echo "  $(YELLOW)make test$(NC)  - Run all tests"
	@echo "  $(YELLOW)make help$(NC)  - Show all available commands"
	@echo ""
	@echo "$(GREEN)Quick test:$(NC)"
	@echo "  curl http://localhost:8081/health/live"
	@echo ""

## ==================== Help ====================

help:
	@echo ""
	@echo "$(GREEN)============================================$(NC)"
	@echo "$(GREEN)  Call Notes AI Service - Help$(NC)"
	@echo "$(GREEN)============================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Quick Start:$(NC)"
	@echo "  1. make setup    - First-time setup (tools + deps + docker + migrate + build)"
	@echo "  2. make run      - Start the service"
	@echo "  3. curl http://localhost:8081/health/live"
	@echo ""
	@echo "$(YELLOW)Build & Run:$(NC)"
	@echo "  make build         - Build binary"
	@echo "  make run           - Run application"
	@echo "  make run-dev       - Run with hot reload (requires air)"
	@echo ""
	@echo "$(YELLOW)Testing & Quality:$(NC)"
	@echo "  make test          - Run all tests (verbose, race detector)"
	@echo "  make test-short    - Quick test run"
	@echo "  make test-coverage - Tests with coverage report"
	@echo "  make lint          - Run golangci-lint"
	@echo "  make vet           - Run go vet"
	@echo "  make check         - Run vet + lint + test"
	@echo "  make fmt           - Format code (gofmt + goimports)"
	@echo ""
	@echo "$(YELLOW)Docker:$(NC)"
	@echo "  make docker-up     - Start Postgres + Redis + Kafka"
	@echo "  make docker-down   - Stop containers"
	@echo "  make docker-clean  - Stop + remove volumes"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run Docker container"
	@echo ""
	@echo "$(YELLOW)Database:$(NC)"
	@echo "  make migrate-up       - Run migrations"
	@echo "  make migrate-down     - Rollback all"
	@echo "  make migrate-down-one - Rollback one"
	@echo "  make migrate-create   - New migration"
	@echo "  make migrate-version  - Current version"
	@echo "  make migrate-force    - Force migration version"
	@echo ""
	@echo "$(YELLOW)Tracing:$(NC)"
	@echo "  make trace-up      - Start Jaeger (UI: http://localhost:16686)"
	@echo "  make trace-down    - Stop Jaeger"
	@echo ""
	@echo "$(YELLOW)Mocks:$(NC)"
	@echo "  make mock          - Generate mocks"
	@echo "  make mock-clean    - Clean mocks"
	@echo ""

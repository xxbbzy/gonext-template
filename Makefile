.PHONY: help init dev dev-backend dev-frontend lint lint-backend lint-frontend typecheck-frontend check test test-backend build build-backend build-frontend seed swagger gen-types migrate-up migrate-down new-migration new-module docker-up docker-down docker-build clean

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ===== Setup =====

init: ## Initialize project: install deps, create .env, prepare local dirs
	@echo "==> Initializing project..."
	@[ -f .env ] || cp .env.example .env
	@echo "==> Installing backend dependencies..."
	@cd backend && go mod download
	@echo "==> Installing frontend dependencies..."
	@cd frontend && npm install
	@echo "==> Creating data directory..."
	@mkdir -p data uploads
	@echo "==> Project initialized! Run 'make dev' to start."

# ===== Development =====

dev: ## Start both backend and frontend dev servers
	@echo "==> Starting development servers..."
	@make -j2 dev-backend dev-frontend

dev-backend: ## Start backend dev server
	@echo "==> Starting backend server..."
	@cd backend && go run ./cmd/server/

dev-frontend: ## Start frontend dev server
	@echo "==> Starting frontend server..."
	@cd frontend && npm run dev

# ===== Code Quality =====

lint: lint-backend lint-frontend ## Run all linters

lint-backend: ## Run Go linters
	@echo "==> Linting backend..."
	@cd backend && golangci-lint run ./...

lint-frontend: ## Run frontend linters
	@echo "==> Linting frontend..."
	@cd frontend && npm run lint

typecheck-frontend: ## Run frontend type checking
	@echo "==> Type checking frontend..."
	@cd frontend && npm run typecheck

check: lint typecheck-frontend test ## Run local quality gates (lint, typecheck, tests)

# ===== Testing =====

test: test-backend ## Run all tests

test-backend: ## Run backend tests
	@echo "==> Testing backend..."
	@cd backend && go test -v ./...

# ===== Build =====

build: build-backend build-frontend ## Build all for production

build-backend: ## Build backend binary
	@echo "==> Building backend..."
	@cd backend && CGO_ENABLED=1 go build -o ../bin/server ./cmd/server/

build-frontend: ## Build frontend for production
	@echo "==> Building frontend..."
	@cd frontend && npm run build

# ===== Database =====

seed: ## Seed database with sample data
	@echo "==> Seeding database..."
	@cd backend && go run ../scripts/seed/main.go

migrate-up: ## Run database migrations (up)
	@echo "==> Running migrations..."
	@migrate -path backend/migrations -database "$(DB_DSN)" up

migrate-down: ## Rollback last database migration
	@echo "==> Rolling back migration..."
	@migrate -path backend/migrations -database "$(DB_DSN)" down 1

new-migration: ## Create new migration files (usage: make new-migration name=xxx)
	@echo "==> Creating migration: $(name)"
	@migrate create -ext sql -dir backend/migrations -seq $(name)

# ===== Code Generation =====

swagger: ## Generate Swagger documentation
	@echo "==> Generating Swagger docs..."
	@cd backend && swag init -g cmd/server/main.go -o docs

gen-types: ## Generate TypeScript types from OpenAPI spec
	@echo "==> Generating TypeScript types..."
	@cd frontend && npx openapi-typescript ../api/openapi.yaml -o types/api.ts

# ===== Module Scaffolding =====

new-module: ## Generate new backend module (usage: make new-module name=xxx)
	@echo "==> Generating module: $(name)"
	@bash scripts/new-module.sh $(name)

# ===== Docker =====

docker-up: ## Start all services with Docker Compose
	@docker compose up -d

docker-down: ## Stop all Docker services
	@docker compose down

docker-build: ## Build Docker images
	@docker compose build

# ===== Cleanup =====

clean: ## Remove build artifacts
	@rm -rf bin/ data/ uploads/ backend/tmp/ frontend/.next/ frontend/node_modules/

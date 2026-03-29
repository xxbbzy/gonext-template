.PHONY: all help init dev dev-backend dev-frontend lint lint-backend lint-frontend typecheck-frontend check test test-backend test-frontend build build-backend build-frontend seed swagger gen gen-server gen-client gen-types migrate-up migrate-down new-migration new-module e2e docker-up docker-down docker-build clean

BACKEND_GO_CACHE := $(CURDIR)/backend/.cache/go-build
BACKEND_LINT_CACHE := $(CURDIR)/backend/.cache/golangci-lint

# Default target
all: check ## Run full validation pipeline

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ===== Setup =====

init: ## Initialize project: install deps, create .env, prepare local dirs
	@echo "==> Initializing project..."
	@[ -f .env ] || cp .env.example .env
	@if [ -f package.json ]; then npm install; fi
	@echo "==> Installing backend dependencies..."
	@cd backend && go mod download
	@echo "==> Installing frontend dependencies..."
	@cd frontend && npm install
	@echo "==> Creating data directory..."
	@mkdir -p data uploads
	@echo "==> Bootstrapping development database..."
	@cd backend && go run ./cmd/bootstrap
	@echo "==> Generating API artifacts..."
	@$(MAKE) gen
	@echo "==> Project initialized! Run 'make dev' to start."

# ===== Development =====

dev: ## Start both backend and frontend dev servers
	@echo "==> Starting development servers..."
	@make -j2 dev-backend dev-frontend

dev-backend: ## Start backend dev server
	@echo "==> Starting backend server..."
	@cd backend && go run github.com/air-verse/air@latest -c .air.toml

dev-frontend: ## Start frontend dev server
	@echo "==> Starting frontend server..."
	@cd frontend && npm run dev

# ===== Code Quality =====

lint: lint-backend lint-frontend ## Run all linters

lint-backend: ## Run Go linters
	@echo "==> Linting backend..."
	@cd backend && mkdir -p .cache/go-build .cache/golangci-lint && \
		GOCACHE=$(BACKEND_GO_CACHE) \
		GOLANGCI_LINT_CACHE=$(BACKEND_LINT_CACHE) \
		golangci-lint run ./...

lint-frontend: ## Run frontend linters
	@echo "==> Linting frontend..."
	@cd frontend && npm run lint

typecheck-frontend: ## Run frontend type checking
	@echo "==> Type checking frontend..."
	@cd frontend && npm run typecheck

check: lint typecheck-frontend test build ## Full validation pipeline
	@echo "✅ All checks passed"

# ===== Testing =====

test: test-backend test-frontend ## Run all tests

test-backend: ## Run backend tests
	@echo "==> Testing backend..."
	@cd backend && mkdir -p .cache/go-build && \
		GOCACHE=$(BACKEND_GO_CACHE) \
		go test -v ./...

test-frontend: ## Run frontend tests
	@echo "==> Testing frontend..."
	@cd frontend && npm test

# ===== Build =====

build: build-backend build-frontend ## Build all for production

build-backend: ## Build backend binary
	@echo "==> Building backend..."
	@cd backend && mkdir -p .cache/go-build && \
		GOCACHE=$(BACKEND_GO_CACHE) \
		CGO_ENABLED=1 go build -o ../bin/server ./cmd/server/

build-frontend: ## Build frontend for production
	@echo "==> Building frontend..."
	@cd frontend && npm run build

# ===== Database =====

seed: ## Seed database with sample data
	@echo "==> Seeding database..."
	@cd backend && go run ./cmd/seed/

migrate-up: ## Run database migrations (up)
	@echo "==> Running migrations..."
	@migrate -path backend/migrations -database "$(DB_DSN)" up

migrate-down: ## Rollback last database migration
	@echo "==> Rolling back migration..."
	@migrate -path backend/migrations -database "$(DB_DSN)" down 1

new-migration: ## Create new migration files (usage: make new-migration name=xxx)
	@echo "==> Creating migration: $(name)"
	@test -n "$(name)" || (echo "name is required" && exit 1)
	@timestamp=$$(date +"%Y%m%d%H%M%S"); \
	touch backend/migrations/$${timestamp}_$(name).up.sql backend/migrations/$${timestamp}_$(name).down.sql

# ===== Code Generation =====

swagger: ## Generate Swagger documentation
	@echo "==> Generating Swagger docs..."
	@cd backend && go run ../scripts/swagger/main.go

gen: gen-server gen-types swagger ## Generate all code from OpenAPI spec

gen-server: ## Generate Go server code from OpenAPI spec
	@echo "==> Generating Go server code..."
	@mkdir -p backend/internal/api
	@cd backend && GOPROXY=https://goproxy.cn,direct go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		-package api \
		-generate "types,gin,strict-server,spec" \
		-exclude-operation-ids "uploadFile,livenessCheck,readinessCheck" \
		-o internal/api/server.gen.go \
		../api/openapi.yaml

# ===== TypeScript Generation =====

gen-types: ## Generate TypeScript API types from OpenAPI spec
	@echo "==> Generating TypeScript types..."
	@cd frontend && npx openapi-typescript ../api/openapi.yaml -o types/api.ts

gen-client: gen-types ## Alias for gen-types (legacy)

# ===== Module Scaffolding =====

new-module: ## Generate new backend module (usage: make new-module name=xxx)
	@echo "==> Generating module: $(name)"
	@bash scripts/new-module.sh $(name)

# ===== E2E Testing =====

e2e: build-backend ## Run E2E smoke test (register → login → CRUD)
	@echo "==> Running E2E smoke test..."
	@bash scripts/e2e-smoke.sh

# ===== Docker =====

docker-up: ## Start all services with Docker Compose
	@docker compose up -d

docker-down: ## Stop all Docker services
	@docker compose down

docker-build: ## Build Docker images
	@docker compose build

# ===== Cleanup =====

clean: ## Remove build artifacts
	@rm -rf bin/ data/ uploads/ backend/tmp/ backend/.cache/ frontend/.next/ frontend/node_modules/
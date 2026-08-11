# =============================================================================
# Accessible Path - Makefile
# =============================================================================

.PHONY: help up down logs restart build test migrate clean ps

# Default target
help:
	@echo "Accessible Path - Available commands:"
	@echo ""
	@echo "Infrastructure:"
	@echo "  make up           - Start all services (dev)"
	@echo "  make down         - Stop all services"
	@echo "  make restart      - Restart all services"
	@echo "  make logs         - Follow logs of all services"
	@echo "  make ps           - Show running containers"
	@echo "  make build        - Build all Docker images"
	@echo ""
	@echo "Development:"
	@echo "  make test         - Run all tests"
	@echo "  make test-go      - Run Go tests"
	@echo "  make test-python  - Run Python tests"
	@echo "  make test-frontend - Run Frontend tests"
	@echo "  make lint         - Run linters"
	@echo ""
	@echo "Database:"
	@echo "  make migrate      - Run all migrations"
	@echo "  make migrate-create - Create new migration (usage: make migrate-create name=xxx)"
	@echo ""
	@echo "Production:"
	@echo "  make prod-up      - Start production stack"
	@echo "  make prod-down    - Stop production stack"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean        - Remove containers, volumes, images"

# -----------------------------------------------------------------------------
# Infrastructure (Development)
# -----------------------------------------------------------------------------

up:
	docker compose up -d

down:
	docker compose down

restart:
	docker compose restart

logs:
	docker compose logs -f

logs-%:
	docker compose logs -f $*

ps:
	docker compose ps

build:
	docker compose build --parallel

# -----------------------------------------------------------------------------
# Testing
# -----------------------------------------------------------------------------

test: test-go test-python test-frontend

test-go:
	@echo "Running Go tests..."
	@cd services/auth-service && go test -race -cover ./...
	@cd services/route-service && go test -race -cover ./...
	@cd services/notification-service && go test -race -cover ./...

test-python:
	@echo "Running Python tests..."
	@cd services/barrier-service && python -m pytest --cov=src --cov-report=term-missing
	@cd services/moderation-service && python -m pytest --cov=src --cov-report=term-missing
	@cd services/poi-service && python -m pytest --cov=src --cov-report=term-missing

test-frontend:
	@echo "Running Frontend tests..."
	@cd frontend && npm run test

# -----------------------------------------------------------------------------
# Linting
# -----------------------------------------------------------------------------

lint: lint-go lint-python lint-frontend

lint-go:
	@echo "Linting Go..."
	@cd services/auth-service && go vet ./... && staticcheck ./...
	@cd services/route-service && go vet ./... && staticcheck ./...
	@cd services/notification-service && go vet ./... && staticcheck ./...

lint-python:
	@echo "Linting Python..."
	@cd services/barrier-service && ruff check . && mypy src
	@cd services/moderation-service && ruff check . && mypy src
	@cd services/poi-service && ruff check . && mypy src

lint-frontend:
	@echo "Linting Frontend..."
	@cd frontend && npm run lint && npx tsc --noEmit

# -----------------------------------------------------------------------------
# Database Migrations
# -----------------------------------------------------------------------------

migrate:
	@echo "Running migrations for all services..."
	@cd services/auth-service && make migrate
	@cd services/barrier-service && make migrate
	@cd services/moderation-service && make migrate
	@cd services/poi-service && make migrate
	@cd services/notification-service && make migrate
	@cd services/route-service && make migrate

migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=migration_name"; exit 1; fi
	@cd services/barrier-service && alembic revision --autogenerate -m "$(name)"
	@cd services/moderation-service && alembic revision --autogenerate -m "$(name)"
	@cd services/poi-service && alembic revision --autogenerate -m "$(name)"

# -----------------------------------------------------------------------------
# Production
# -----------------------------------------------------------------------------

prod-up:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

prod-down:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml down

prod-logs:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml logs -f

prod-build:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml build --parallel

# -----------------------------------------------------------------------------
# Cleanup
# -----------------------------------------------------------------------------

clean:
	docker compose down -v --rmi all --remove-orphans
	docker system prune -f

clean-all:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml down -v --rmi all --remove-orphans
	docker system prune -af --volumes
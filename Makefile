.PHONY: all build test test-backend test-frontend smoke-test integration-test e2e-test clean dev up down logs

# Default target
all: test build

# Build all services
build:
	docker compose build

# Run all tests
test: test-backend test-frontend

# Run backend tests
test-backend:
	cd backend && go test -v -race ./...

# Run frontend tests
test-frontend:
	cd frontend && npm run test:run

# Run smoke tests (requires running services)
smoke-test:
	./scripts/smoke-test.sh

# Run smoke tests with verbose output
smoke-test-verbose:
	VERBOSE=true ./scripts/smoke-test.sh

# Run integration tests (starts Docker services)
integration-test: up
	@echo "Waiting for services to start..."
	@sleep 10
	./scripts/smoke-test.sh
	$(MAKE) down

# Start development environment
dev: up logs

# Start all services
up:
	docker compose up -d --build

# Stop all services
down:
	docker compose down

# Stop all services and remove volumes
down-clean:
	docker compose down -v

# View logs
logs:
	docker compose logs -f

# View logs for specific service
logs-backend:
	docker compose logs -f backend

logs-frontend:
	docker compose logs -f frontend

logs-db:
	docker compose logs -f db

# Run backend locally (for development)
run-backend:
	cd backend && go run ./cmd/server

# Generate self-signed certificates
generate-certs:
	./scripts/generate-self-signed-cert.sh

# Clean build artifacts
clean:
	cd backend && make clean
	cd frontend && rm -rf dist node_modules

# Database operations
db-shell:
	docker compose exec db psql -U ownu -d ownu

db-reset:
	docker compose down -v
	docker compose up -d db
	@echo "Waiting for database..."
	@sleep 5
	docker compose up -d backend

# Quick rebuild of a single service
rebuild-backend:
	docker compose build backend
	docker compose up -d backend

rebuild-frontend:
	docker compose build frontend
	docker compose up -d frontend

# Check service health
health:
	@echo "Backend health:"
	@curl -s http://localhost:8080/health | jq . || echo "Backend not responding"
	@echo "\nFrontend health:"
	@curl -s -k https://localhost/ -o /dev/null -w "HTTP %{http_code}\n" || echo "Frontend not responding"

# Run all quality checks
check: lint test smoke-test

# Run linters
lint:
	cd backend && go vet ./...
	cd frontend && npm run lint || true

# Format code
fmt:
	cd backend && go fmt ./...
	cd frontend && npm run lint:fix || true

# Help
help:
	@echo "OwnU Development Commands"
	@echo "========================"
	@echo ""
	@echo "Development:"
	@echo "  make dev              - Start all services and follow logs"
	@echo "  make up               - Start all services"
	@echo "  make down             - Stop all services"
	@echo "  make logs             - Follow all logs"
	@echo "  make rebuild-backend  - Rebuild and restart backend"
	@echo "  make rebuild-frontend - Rebuild and restart frontend"
	@echo ""
	@echo "Testing:"
	@echo "  make test             - Run all unit tests"
	@echo "  make test-backend     - Run backend tests"
	@echo "  make test-frontend    - Run frontend tests"
	@echo "  make smoke-test       - Run smoke tests (services must be running)"
	@echo "  make integration-test - Start services and run smoke tests"
	@echo ""
	@echo "Database:"
	@echo "  make db-shell         - Open PostgreSQL shell"
	@echo "  make db-reset         - Reset database (removes all data)"
	@echo ""
	@echo "Utilities:"
	@echo "  make health           - Check service health"
	@echo "  make generate-certs   - Generate self-signed SSL certificates"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make lint             - Run linters"
	@echo "  make fmt              - Format code"

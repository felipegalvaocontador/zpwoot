# zpwoot Makefile

.PHONY: help build run test clean deps docker-build docker-run migrate-up migrate-down kill ps-port down-clean down-cw-clean clean-volumes list-volumes swagger swagger-quick install-swag

# Variables
APP_NAME=zpwoot
BUILD_DIR=build
DOCKER_IMAGE=zpwoot:latest
DATABASE_URL=postgres://user:password@localhost:5432/zpwoot?sslmode=disable

# Build information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS = -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
deps: ## Install dependencies
	go mod download
	go mod tidy

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) cmd/zpwoot/main.go

build-release: ## Build the application for release
	@echo "Building $(APP_NAME) for release..."
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -s -w" -o $(BUILD_DIR)/$(APP_NAME) cmd/zpwoot/main.go

version: ## Show version information
	@go run -ldflags "$(LDFLAGS)" cmd/zpwoot/main.go -version

run: ## Run the application (local development)
	@echo "🚀 Running $(APP_NAME) in local mode..."
	go run cmd/zpwoot/main.go

run-build: build ## Build and run the application
	@echo "🚀 Running built $(APP_NAME)..."
	./$(BUILD_DIR)/$(APP_NAME)

run-docker: ## Run the application with Docker environment variables
	@echo "Running $(APP_NAME) with Docker configuration..."
	@if [ -f .env.docker ]; then \
		export $$(cat .env.docker | grep -v '^#' | xargs) && go run cmd/zpwoot/main.go; \
	else \
		echo "Error: .env.docker file not found"; \
		exit 1; \
	fi

dev: ## Run in development mode with hot reload (requires air)
	@echo "🚀 Starting development server with hot reload..."
	@echo "📁 Working directory: $(shell pwd)"
	@echo "🔥 Air will watch for changes and automatically rebuild..."
	@echo "📝 Config file: .air.toml"
	@echo ""
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "❌ Air not found. Installing..."; \
		$(MAKE) install-air; \
		air; \
	fi

dev-init: ## Initialize Air configuration
	@echo "🔧 Initializing Air configuration..."
	@if [ -f .air.toml ]; then \
		echo "⚠️  .air.toml already exists. Backing up to .air.toml.backup"; \
		cp .air.toml .air.toml.backup; \
	fi
	air init
	@echo "✅ Air configuration initialized!"

dev-clean: ## Clean Air temporary files
	@echo "🧹 Cleaning Air temporary files..."
	@rm -rf tmp/
	@rm -f .air.toml.backup
	@echo "✅ Air temporary files cleaned!"

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

swagger: install-swag ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	swag init -g cmd/zpwoot/main.go -o docs/swagger --parseDependency --parseInternal
	@echo "✅ Swagger docs generated at docs/swagger/"

swagger-serve: swagger ## Generate docs and serve Swagger documentation locally
	@echo "Starting Swagger UI server..."
	@echo "📖 Swagger UI will be available at: http://localhost:8080/swagger/"
	@echo "🚀 Starting zpwoot server..."
	go run cmd/zpwoot/main.go

swagger-quick: ## Quick install swag and generate docs
	@echo "🚀 Quick Swagger setup..."
	go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g cmd/zpwoot/main.go -o docs/swagger --parseDependency --parseInternal
	@echo "✅ Swagger docs generated at docs/swagger/"

swagger-test: swagger ## Generate docs and test Swagger endpoint
	@echo "🧪 Testing Swagger documentation..."
	@echo "📖 Generating and starting server..."
	@go run cmd/zpwoot/main.go &
	@sleep 3
	@echo "🔍 Testing Swagger endpoints..."
	@curl -s http://localhost:8080/swagger/index.html > /dev/null && echo "✅ Swagger UI is accessible" || echo "❌ Swagger UI failed"
	@curl -s http://localhost:8080/swagger/doc.json > /dev/null && echo "✅ Swagger JSON is accessible" || echo "❌ Swagger JSON failed"
	@curl -s http://localhost:8080/health | jq . && echo "✅ Health endpoint working" || echo "❌ Health endpoint failed"
	@pkill -f "go run cmd/zpwoot/main.go" || true

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

kill: ## Kill processes running on port 8080
	@echo "Killing processes on port 8080..."
	@if command -v lsof >/dev/null 2>&1; then \
		pids=$$(lsof -ti:8080 2>/dev/null); \
		if [ -n "$$pids" ]; then \
			echo "Found processes: $$pids"; \
			echo "$$pids" | xargs kill -9; \
			echo "Processes killed successfully!"; \
		else \
			echo "No processes found on port 8080"; \
		fi; \
	elif command -v netstat >/dev/null 2>&1; then \
		pids=$$(netstat -tlnp 2>/dev/null | grep :8080 | awk '{print $$7}' | cut -d/ -f1 | grep -v '^-$$'); \
		if [ -n "$$pids" ]; then \
			echo "Found processes: $$pids"; \
			echo "$$pids" | xargs kill -9; \
			echo "Processes killed successfully!"; \
		else \
			echo "No processes found on port 8080"; \
		fi; \
	else \
		echo "Neither lsof nor netstat found. Cannot kill processes."; \
		exit 1; \
	fi

ps-port: ## Show processes running on port 8080
	@echo "Checking processes on port 8080..."
	@if command -v lsof >/dev/null 2>&1; then \
		lsof -i:8080 || echo "No processes found on port 8080"; \
	elif command -v netstat >/dev/null 2>&1; then \
		netstat -tlnp | grep :8080 || echo "No processes found on port 8080"; \
	else \
		echo "Neither lsof nor netstat found. Cannot check processes."; \
	fi

# Database
migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	@go run cmd/zpwoot/main.go -migrate-up || echo "Note: Migrations are automatically run on application startup"

migrate-down: ## Run database migrations down (rollback last migration)
	@echo "Rolling back last migration..."
	@go run cmd/zpwoot/main.go -migrate-down

migrate-status: ## Show migration status
	@echo "Checking migration status..."
	@go run cmd/zpwoot/main.go -migrate-status

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@echo "Creating migration: $(NAME)"
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@NEXT_VERSION=$$(ls internal/infra/db/migrations/ | grep -E '^[0-9]+_' | sed 's/_.*//' | sort -n | tail -1 | awk '{print $$1 + 1}') && \
	if [ -z "$$NEXT_VERSION" ]; then NEXT_VERSION=1; fi && \
	printf "%03d" $$NEXT_VERSION > /tmp/version && \
	VERSION=$$(cat /tmp/version) && \
	echo "Creating migration files for version $$VERSION..." && \
	touch "internal/infra/db/migrations/$${VERSION}_$(NAME).up.sql" && \
	touch "internal/infra/db/migrations/$${VERSION}_$(NAME).down.sql" && \
	echo "-- Migration: $(NAME)" > "internal/infra/db/migrations/$${VERSION}_$(NAME).up.sql" && \
	echo "-- Add your migration SQL here" >> "internal/infra/db/migrations/$${VERSION}_$(NAME).up.sql" && \
	echo "" >> "internal/infra/db/migrations/$${VERSION}_$(NAME).up.sql" && \
	echo "-- Migration: $(NAME) (rollback)" > "internal/infra/db/migrations/$${VERSION}_$(NAME).down.sql" && \
	echo "-- Add your rollback SQL here" >> "internal/infra/db/migrations/$${VERSION}_$(NAME).down.sql" && \
	echo "" >> "internal/infra/db/migrations/$${VERSION}_$(NAME).down.sql" && \
	echo "Created migration files:" && \
	echo "  internal/infra/db/migrations/$${VERSION}_$(NAME).up.sql" && \
	echo "  internal/infra/db/migrations/$${VERSION}_$(NAME).down.sql"

# Docker
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE)

docker-compose-up: ## Start services with docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose up -d

docker-compose-down: ## Stop services with docker-compose
	@echo "Stopping services with docker-compose..."
	docker-compose down

# Development Environment Services
up: ## Start main development services (PostgreSQL, Redis, DbGate, etc.)
	@echo "🚀 Starting zpwoot main services..."
	docker compose -f docker-compose.dev.yml up -d
	@echo "✅ Main services started!"
	@echo "📊 DbGate: http://localhost:3000"
	@echo "🔴 Redis Commander: http://localhost:8081"
	@echo "🪝 Webhook Tester: http://localhost:8090"

down: ## Stop main development services (keeps volumes)
	@echo "🛑 Stopping zpwoot main services..."
	docker compose -f docker-compose.dev.yml down
	@echo "✅ Main services stopped!"
	@echo "💾 Volumes preserved. Use 'make down-clean' to remove volumes too."

down-clean: ## Stop main development services and remove volumes
	@echo "🛑 Stopping zpwoot main services and removing volumes..."
	docker compose -f docker-compose.dev.yml down -v
	@echo "✅ Main services stopped and volumes removed!"
	@echo "⚠️  All data has been permanently deleted!"

up-cw: ## Start Chatwoot services
	@echo "💬 Starting Chatwoot services..."
	docker compose -f docker-compose.chatwoot.yml up -d
	@echo "✅ Chatwoot services started!"
	@echo "💬 Chatwoot: http://localhost:3001"
	@echo ""
	@echo "⏳ Chatwoot may take a few minutes to initialize..."
	@echo "📋 Check logs with: make logs-cw"

down-cw: ## Stop Chatwoot services (keeps volumes)
	@echo "🛑 Stopping Chatwoot services..."
	docker compose -f docker-compose.chatwoot.yml down
	@echo "✅ Chatwoot services stopped!"
	@echo "💾 Volumes preserved. Use 'make down-cw-clean' to remove volumes too."

down-cw-clean: ## Stop Chatwoot services and remove volumes
	@echo "🛑 Stopping Chatwoot services and removing volumes..."
	docker compose -f docker-compose.chatwoot.yml down -v
	@echo "✅ Chatwoot services stopped and volumes removed!"
	@echo "⚠️  All Chatwoot data has been permanently deleted!"

logs-cw: ## Show Chatwoot logs
	@echo "📋 Showing logs for Chatwoot services..."
	docker compose -f docker-compose.chatwoot.yml logs -f

ps-services: ## Show status of all development containers
	@echo "📊 Development services status:"
	@echo "==============================="
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "(zpwoot|NAMES)"

clean-services: ## Stop all services and remove volumes (DESTRUCTIVE)
	@echo "🧹 Cleaning up all development services and volumes..."
	@echo "⚠️  This will permanently delete ALL data!"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	docker compose -f docker-compose.dev.yml down -v
	docker compose -f docker-compose.chatwoot.yml down -v
	@echo "✅ Cleanup complete - all data permanently deleted!"

clean-volumes: ## Remove only the volumes (without stopping services)
	@echo "🧹 Removing development volumes..."
	@echo "⚠️  This will permanently delete ALL data!"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	docker volume rm zpwoot_postgres_data zpwoot_redis_data zpwoot_chatwoot_postgres_data zpwoot_chatwoot_redis_data zpwoot_chatwoot_storage zpwoot_chatwoot_public 2>/dev/null || true
	@echo "✅ Volumes removed!"

list-volumes: ## List all project volumes and their sizes
	@echo "📊 zpwoot Development Volumes:"
	@echo "=============================="
	@docker volume ls --filter name=zpwoot --format "table {{.Name}}\t{{.Driver}}\t{{.Scope}}" 2>/dev/null || echo "No volumes found"
	@echo ""
	@echo "💾 Volume sizes:"
	@docker system df -v | grep -E "(zpwoot|VOLUME NAME)" || echo "No volume size info available"

restart-services: ## Restart main development services
	@echo "🔄 Restarting main services..."
	docker compose -f docker-compose.dev.yml restart
	@echo "✅ Main services restarted!"

restart-cw: ## Restart Chatwoot services
	@echo "🔄 Restarting Chatwoot services..."
	docker compose -f docker-compose.chatwoot.yml restart
	@echo "✅ Chatwoot services restarted!"

urls: ## Show all service URLs
	@echo "🌐 Development Service URLs:"
	@echo "============================"
	@echo "📊 DbGate (Database Admin): http://localhost:3000"
	@echo "💬 Chatwoot (Customer Support): http://localhost:3001"
	@echo "🔴 Redis Commander: http://localhost:8081"
	@echo "🪝 Webhook Tester: http://localhost:8090"
	@echo ""
	@echo "🐘 PostgreSQL: localhost:5432"
	@echo "🔴 Redis: localhost:6379"

# Linting and formatting
fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

# Security
security-check: ## Run security checks
	@echo "Running security checks..."
	gosec ./...

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	godoc -http=:6060

# Installation helpers
install-swag: ## Install swag tool for Swagger generation
	@echo "Checking if swag is installed..."
	@which swag > /dev/null 2>&1 || { \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		echo "✅ swag installed successfully"; \
	}

install-air: ## Install Air for hot reload
	@echo "📦 Installing Air for hot reload..."
	@if command -v air >/dev/null 2>&1; then \
		echo "✅ Air is already installed"; \
		air -v; \
	else \
		echo "Installing Air..."; \
		go install github.com/air-verse/air@latest; \
		echo "✅ Air installed successfully"; \
		air -v; \
	fi

install-tools: install-swag install-air ## Install development tools
	@echo "📦 Installing development tools..."
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
	@echo "✅ All development tools installed!"

# Environment setup
setup: deps install-tools ## Setup development environment
	@echo "Setting up development environment..."
	cp .env.example .env
	@echo "Please edit .env file with your configuration"

# Production
build-prod: ## Build for production
	@echo "Building for production..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(APP_NAME) cmd/zpwoot/main.go

# Health checks
health: ## Check application health
	@echo "Checking application health..."
	curl -f http://localhost:8080/health || exit 1

# Logs
logs: ## Show application logs (for docker-compose)
	docker-compose logs -f zpwoot

# Database operations
db-reset: migrate-down migrate-up ## Reset database

db-seed: ## Seed database with sample data
	@echo "Seeding database..."
	@go run cmd/zpwoot/main.go -seed

# Backup and restore
backup: ## Backup database
	@echo "Backing up database..."
	pg_dump $(DATABASE_URL) > backup_$(shell date +%Y%m%d_%H%M%S).sql

restore: ## Restore database from backup (usage: make restore BACKUP=backup_file.sql)
	@echo "Restoring database from $(BACKUP)..."
	psql $(DATABASE_URL) < $(BACKUP)

.PHONY: help dev build run test lint \
       migrate-up migrate-fresh migrate-refresh seed \
       docker-all docker-db docker-cache docker-routing docker-down docker-logs \
       check-services clean

# ─── Colors ──────────────────────────────────────────
GREEN  := \033[0;32m
YELLOW := \033[0;33m
CYAN   := \033[0;36m
RED    := \033[0;31m
DIM    := \033[0;90m
RESET  := \033[0m

help: ## Show this help
	@echo "$(CYAN)UB-Mager API$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'

# ─── Development ─────────────────────────────────────

dev: ## Run in development mode
	@echo "$(YELLOW)Starting dev server...$(RESET)"
	go run ./cmd/server

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build: ## Build the binary
	@echo "$(YELLOW)Building $(VERSION)...$(RESET)"
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/server ./cmd/server
	@echo "$(GREEN)Build complete: bin/server ($(VERSION))$(RESET)"

run: build ## Build and run
	./bin/server

test: ## Run tests
	go test -v -race -cover ./...

lint: ## Run linter
	golangci-lint run ./...

# ─── Database ────────────────────────────────────────

migrate-up: ## Run auto-migration (create/alter tables)
	go run ./cmd/migrate up

migrate-fresh: ## Drop all tables and re-migrate from scratch
	go run ./cmd/migrate fresh

migrate-refresh: ## Fresh migration + seed (full reset)
	go run ./cmd/migrate fresh
	@echo ""
	go run ./cmd/seed

seed: ## Seed database with test data
	go run ./cmd/seed

# ─── Docker — pick what you need ─────────────────────
#
# Already have PostgreSQL/Redis running? Just start what's missing:
#   make docker-routing        → OSRM only
#   make docker-db             → PostgreSQL + pgAdmin only
#   make docker-cache          → Redis only
#   make docker-all            → Everything
#

docker-all: ## Start ALL services (PostgreSQL, Redis, OSRM, pgAdmin)
	docker compose --profile full up -d
	@printf "$(GREEN)All services started$(RESET)\n"

docker-db: ## Start PostgreSQL + pgAdmin only
	docker compose --profile db up -d
	@printf "$(GREEN)PostgreSQL + pgAdmin started$(RESET)\n"

docker-cache: ## Start Redis only
	docker compose --profile cache up -d
	@printf "$(GREEN)Redis started$(RESET)\n"

docker-routing: ## Start OSRM only
	docker compose --profile routing up -d
	@printf "$(GREEN)OSRM started on :5000$(RESET)\n"

docker-down: ## Stop all Docker services
	docker compose --profile full down
	@printf "$(GREEN)Services stopped$(RESET)\n"

docker-logs: ## Show Docker logs
	docker compose --profile full logs -f

check-services: ## Check all required services are running
	@printf "\033[0;36mChecking services...\033[0m\n"
	@for pair in \
		"postgres_server|ubmager-postgres:PostgreSQL" \
		"wardayagate-redis|ubmager-redis:Redis" \
		"pgadmin_web|ubmager-pgadmin:pgAdmin" \
		"ubmager-osrm:OSRM"; \
	do \
		names=$${pair%%:*}; \
		label=$${pair##*:}; \
		found=0; \
		IFS='|'; for name in $$names; do \
			status=$$(docker ps --filter "name=$$name" --format "{{.Status}}" 2>/dev/null | head -1); \
			if [ -n "$$status" ]; then \
				printf "  \033[0;32m✓\033[0m %-12s %s \033[0;90m(%s)\033[0m\n" "$$label" "$$status" "$$name"; \
				found=1; \
				break; \
			fi; \
		done; \
		if [ "$$found" = "0" ]; then \
			printf "  \033[0;31m✗\033[0m %-12s NOT RUNNING\n" "$$label"; \
		fi; \
	done

# ─── Cleanup ─────────────────────────────────────────

clean: ## Clean build artifacts
	rm -rf bin/ tmp/
	@echo "$(GREEN)Cleaned$(RESET)"

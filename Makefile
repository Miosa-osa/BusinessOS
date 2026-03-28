# =============================================================================
# BusinessOS — Project Makefile
# Usage: make <target>
# Run `make help` to see all available targets.
# =============================================================================

.DEFAULT_GOAL := help
SHELL         := /bin/bash

# Colours for terminal output
BOLD  := \033[1m
RESET := \033[0m
GREEN := \033[32m
CYAN  := \033[36m
YELLOW := \033[33m
RED := \033[31m

.PHONY: help
help: ## Show this help message
	@printf '$(BOLD)BusinessOS — available targets:$(RESET)\n\n'
	@printf '$(YELLOW)Setup:$(RESET)\n'
	@awk 'BEGIN {FS = ":.*##"} /^setup|^up|^down|^restart/ { printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@printf '\n$(YELLOW)Development:$(RESET)\n'
	@awk 'BEGIN {FS = ":.*##"} /^dev|^build|^rebuild/ { printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@printf '\n$(YELLOW)Testing:$(RESET)\n'
	@awk 'BEGIN {FS = ":.*##"} /^test|^weaver/ { printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@printf '\n$(YELLOW)Logs & Debugging:$(RESET)\n'
	@awk 'BEGIN {FS = ":.*##"} /^logs|^status|^debug|^profile|^shell/ { printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@printf '\n$(YELLOW)Database:$(RESET)\n'
	@awk 'BEGIN {FS = ":.*##"} /^db-/ { printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@printf '\n$(YELLOW)Data Layer (bos):$(RESET)\n'
	@awk 'BEGIN {FS = ":.*##"} /^bos-/ { printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@printf '\n$(YELLOW)Cleanup:$(RESET)\n'
	@awk 'BEGIN {FS = ":.*##"} /^clean|^urls/ { printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@printf '\n'

# =============================================================================
# Setup
# =============================================================================

.PHONY: setup
setup: ## First-time setup: copy .env files, pull images, start stack
	@echo ""
	@printf '$(BOLD)Setting up BusinessOS...$(RESET)\n'
	@cp -n .env.example .env 2>/dev/null && \
		printf '  $(GREEN)Created$(RESET) .env from .env.example\n' || \
		printf '  .env already exists — skipping\n'
	@cp -n desktop/backend-go/.env.example desktop/backend-go/.env 2>/dev/null && \
		printf '  $(GREEN)Created$(RESET) desktop/backend-go/.env\n' || \
		printf '  desktop/backend-go/.env already exists — skipping\n'
	@echo ""
	@printf '$(BOLD)Pulling base images (DB, cache, observability)...$(RESET)\n'
	@docker compose pull postgres redis otel-collector jaeger
	@echo ""
	@printf '$(BOLD)Starting infrastructure services...$(RESET)\n'
	@docker compose up -d postgres redis
	@printf 'Waiting for postgres + redis to be healthy...\n'
	@bash scripts/wait-healthy.sh postgres redis
	@echo ""
	@printf '$(BOLD)Building and starting application services...$(RESET)\n'
	@docker compose up -d --build backend
	@bash scripts/wait-healthy.sh backend
	@docker compose up -d --build frontend
	@echo ""
	@bash scripts/print-urls.sh

.PHONY: dev
dev: ## Start all services (build if needed), follow logs
	@docker compose up -d --build
	@docker compose logs -f

.PHONY: up
up: ## Start all services in the background
	@docker compose up -d

.PHONY: down
down: ## Stop all services (preserves volumes)
	@docker compose down

.PHONY: restart
restart: ## Restart all services
	@docker compose restart

# =============================================================================
# Build
# =============================================================================

.PHONY: build
build: ## Build all Docker images
	@docker compose build

.PHONY: build-backend
build-backend: ## Build only the backend image
	@docker compose build backend

.PHONY: build-frontend
build-frontend: ## Build only the frontend image
	@docker compose build frontend

.PHONY: rebuild
rebuild: ## Force rebuild all images (no cache)
	@docker compose build --no-cache

# =============================================================================
# Logs & Status
# =============================================================================

.PHONY: logs
logs: ## Follow logs from all services
	@docker compose logs -f

.PHONY: logs-backend
logs-backend: ## Follow backend logs only
	@docker compose logs -f backend

.PHONY: logs-frontend
logs-frontend: ## Follow frontend logs only
	@docker compose logs -f frontend

.PHONY: logs-db
logs-db: ## Follow postgres logs only
	@docker compose logs -f postgres

.PHONY: status
status: ## Show service health status
	@docker compose ps

.PHONY: debug
debug: ## Open a bash shell in the running backend container
	@docker compose exec backend bash

.PHONY: profile
profile: ## Show CPU/memory usage of running services
	@docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}"

# =============================================================================
# Testing
# =============================================================================

.PHONY: test
test: test-backend test-frontend ## Run all tests (unit + frontend)

.PHONY: test-backend
test-backend: ## Run Go backend tests
	@echo ""
	@printf '$(BOLD)Running Go tests...$(RESET)\n'
	@cd desktop/backend-go && go test ./... -count=1

# =============================================================================
# bos — BusinessOS Data Layer CLI
# =============================================================================

.PHONY: bos-build
bos-build: ## Build bos CLI (Rust)
	@echo ""
	@printf '$(BOLD)Building bos CLI...$(RESET)\n'
	@cd bos && cargo build --release
	@printf '  $(GREEN)Built$(RESET) bos/target/release/bos\n'

.PHONY: bos-test
bos-test: ## Run bos tests
	@echo ""
	@printf '$(BOLD)Running bos tests...$(RESET)\n'
	@cd bos && cargo test
	@printf '  $(GREEN)All tests passed$(RESET)\n'

.PHONY: bos-validate
bos-validate: bos-build ## Validate ontology mappings
	@echo ""
	@printf '$(BOLD)Validating ontology mappings...$(RESET)\n'
	@./bos/target/release/bos ontology construct \
		--mapping ./ontology-mappings.json \
		--output /tmp/bos-validate-check/
	@printf '  $(GREEN)Ontology mappings valid$(RESET)\n'

.PHONY: bos-generate
bos-generate: bos-build ## Generate SPARQL CONSTRUCT queries from ontology mappings
	@echo ""
	@printf '$(BOLD)Generating SPARQL CONSTRUCT queries...$(RESET)\n'
	@mkdir -p desktop/backend-go/bos/queries
	@./bos/target/release/bos ontology construct \
		--mapping ./ontology-mappings.json \
		--output ./desktop/backend-go/bos/queries/
	@printf '  $(GREEN)Generated $$(ls desktop/backend-go/bos/queries/*.rq 2>/dev/null | wc -l | tr -d " ") CONSTRUCT queries$(RESET)\n'

.PHONY: bos-export
bos-export: bos-build ## Export ontology mappings as RDF
	@echo ""
	@printf '$(BOLD)Exporting ontology as RDF...$(RESET)\n'
	@mkdir -p desktop/backend-go/bos/artifacts
	@./bos/target/release/bos ontology export \
		--mapping ./ontology-mappings.json \
		--output ./desktop/backend-go/bos/artifacts/ontology.ttl \
		--format ttl
	@printf '  $(GREEN)Exported to desktop/backend-go/bos/artifacts/ontology.ttl$(RESET)\n'

.PHONY: bos-execute
bos-execute: bos-build ## Execute CONSTRUCT queries against PostgreSQL (requires DATABASE_URL)
	@test -n "$(DATABASE_URL)" || (echo "ERROR: DATABASE_URL not set" && exit 1)
	@echo ""
	@printf '$(BOLD)Executing CONSTRUCT queries against PostgreSQL...$(RESET)\n'
	@./bos/target/release/bos ontology execute \
		--mapping ./ontology-mappings.json \
		--database "$(DATABASE_URL)" \
		--format ttl

.PHONY: test-frontend
test-frontend: ## Run SvelteKit frontend tests
	@echo ""
	@printf '$(BOLD)Running frontend tests...$(RESET)\n'
	@cd frontend && npm test

.PHONY: test-backend-verbose
test-backend-verbose: ## Run Go tests with verbose output
	@cd desktop/backend-go && go test ./... -count=1 -v

.PHONY: weaver-live-check
weaver-live-check: ## Run semconv tests with Weaver live-check OTEL export
	@echo ""
	@printf '$(BOLD)Running Weaver live-check...$(RESET)\n'
	@cd desktop/backend-go && WEAVER_LIVE_CHECK=true go test ./internal/semconv/... -v -count=1

# =============================================================================
# Database
# =============================================================================

.PHONY: db-shell
db-shell: ## Open a psql shell inside the postgres container
	@docker compose exec postgres psql -U $${POSTGRES_USER:-postgres} -d $${POSTGRES_DB:-business_os}

.PHONY: db-migrate
db-migrate: ## Re-apply init.sql against the running postgres container
	@docker compose exec -T postgres psql \
		-U $${POSTGRES_USER:-postgres} \
		-d $${POSTGRES_DB:-business_os} \
		< desktop/backend-go/internal/database/init.sql
	@printf '$(GREEN)Migration applied$(RESET)\n'

.PHONY: db-seed
db-seed: ## Run seed data against the running postgres container
	@docker compose exec -T postgres psql \
		-U $${POSTGRES_USER:-postgres} \
		-d $${POSTGRES_DB:-business_os} \
		< desktop/backend-go/scripts/seed/seed.sql
	@printf '$(GREEN)Seed data loaded$(RESET)\n'

# =============================================================================
# Cleanup
# =============================================================================

.PHONY: clean
clean: ## Stop containers and remove volumes (DESTROYS all local data)
	@printf '$(BOLD)Removing containers and volumes...$(RESET)\n'
	@docker compose down -v
	@printf '$(GREEN)Done$(RESET)\n'

.PHONY: clean-images
clean-images: ## Remove locally built BusinessOS images
	@docker rmi businessos-backend:local businessos-frontend:local 2>/dev/null || true
	@printf '$(GREEN)Local images removed$(RESET)\n'

# =============================================================================
# Shortcuts
# =============================================================================

.PHONY: shell-backend
shell-backend: ## Open a shell inside the running backend container
	@docker compose exec backend sh

.PHONY: shell-frontend
shell-frontend: ## Open a shell inside the running frontend container
	@docker compose exec frontend sh

.PHONY: urls
urls: ## Print service URLs
	@bash scripts/print-urls.sh

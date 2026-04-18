# SDA Framework — Unified Interface
# Usage: make <target>

SHELL := /bin/bash
.DEFAULT_GOAL := help

# Directories
ROOT_DIR := $(shell pwd)
SERVICES_DIR := $(ROOT_DIR)/services
DEPLOY_DIR := $(ROOT_DIR)/deploy

# Go
GOBIN := $(ROOT_DIR)/bin
GO_SERVICES := $(shell ls -d $(SERVICES_DIR)/*/go.mod 2>/dev/null | xargs -I{} dirname {} | xargs -I{} basename {})

# Build info (injected via -ldflags)
GIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS_BASE := -s -w \
	-X github.com/Camionerou/rag-saldivia/pkg/server.GitSHA=$(GIT_SHA) \
	-X github.com/Camionerou/rag-saldivia/pkg/server.BuildTime=$(BUILD_TIME)

export GOBIN

.PHONY: help dev stop test lint build migrate deploy new-service clean versions

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ── Development ──────────────────────────────────────────────────────────

dev: ## Start infra only (run Go services on host)
	docker compose -f $(DEPLOY_DIR)/docker-compose.dev.yml up

dev-full: ## Start infra + all Go services in Docker
	docker compose -f $(DEPLOY_DIR)/docker-compose.dev.yml --profile full up --build

dev-gpu: ## Start infra + SGLang model servers (requires NVIDIA GPU)
	docker compose -f $(DEPLOY_DIR)/docker-compose.dev.yml --profile gpu up

dev-services: ## Start all Go services on host (requires infra running)
	@echo "Starting all Go services..."
	@ENV_COMMON="POSTGRES_TENANT_URL=postgres://sda:sda_dev@localhost:5432/sda_tenant_dev?sslmode=disable \
		POSTGRES_PLATFORM_URL=postgres://sda:sda_dev@localhost:5432/sda_platform?sslmode=disable \
		REDIS_URL=localhost:6379 NATS_URL=nats://localhost:4222 TENANT_SLUG=dev \
		JWT_PUBLIC_KEY=LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUNvd0JRWURLMlZ3QXlFQVpMSmkrZmtPbitKUllNQmc4VkVBTkh2bXRzZUxQK3JmRFdFUStZL3ZIU0E9Ci0tLS0tRU5EIFBVQkxJQyBLRVktLS0tLQo= \
		JWT_PRIVATE_KEY=LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1DNENBUUF3QlFZREsyVndCQ0lFSUZvSXFxYU1BcjVjYnZFSE9Rc2g0cnVQTUUzeCtRSkVlVDByNnkxQ2tjMmgKLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQo="; \
	env $$ENV_COMMON ERP_PORT=8013 nohup go run ./services/erp/cmd/... > /tmp/sda-erp.log 2>&1 & \
	env $$ENV_COMMON APP_PORT=8020 TENANT_ID=dev SCAN_MODE=passive PROMETHEUS_URL=http://localhost:9090 DOCKER_PROXY_URL=http://localhost:2375 PLATFORM_TENANT_SLUG=platform INGEST_STAGING_DIR=/tmp/ingest-staging SMTP_HOST=localhost SMTP_PORT=1025 SMTP_FROM=noreply@sda.local WS_ALLOWED_ORIGINS="http://localhost:3000" NOTIFICATION_SERVICE_URL=http://localhost:8020 BIGBROTHER_SERVICE_URL=http://localhost:8020 ERP_SERVICE_URL=http://localhost:8013 nohup go run ./services/app/cmd > /tmp/sda-app.log 2>&1 & \
	echo "All services starting. Logs in /tmp/sda-*.log" && echo "Run 'make status' to check."

dev-frontend: ## Start Next.js frontend in dev/HMR mode (LOCAL laptop @ localhost only — remote IP access fails to hydrate)
	@cd apps/web && NEXT_PUBLIC_API_URL= NEXT_PUBLIC_TENANT_SLUG=dev nohup bun run dev > /tmp/sda-frontend.log 2>&1 &
	@echo "Frontend (dev mode) starting on :3000. Log: /tmp/sda-frontend.log"
	@echo "NOTE: dev mode does not hydrate over remote IP. Workstation/test envs"
	@echo "      use the Docker 'web' service via docker compose (see deploy.yml)."

dev-all: ## Start everything: infra + services + frontend
	@$(MAKE) dev &
	@sleep 15
	@$(MAKE) dev-services
	@sleep 5
	@$(MAKE) dev-frontend
	@sleep 10
	@$(MAKE) status

stop: ## Stop all services (Docker + Go + frontend)
	@echo "Stopping Go services..."
	@pkill -f "go run ./services/" 2>/dev/null || true
	@pkill -f "go-build.*sda" 2>/dev/null || true
	@echo "Stopping frontend..."
	@pkill -f "bun.*dev" 2>/dev/null || true
	@pkill -f "next-server" 2>/dev/null || true
	@echo "Stopping Docker..."
	@docker compose -f $(DEPLOY_DIR)/docker-compose.dev.yml --profile full --profile gpu down 2>/dev/null || true
	@echo "All stopped."

# ── Build ────────────────────────────────────────────────────────────────

build: ## Build all Go services
	@for svc in $(GO_SERVICES); do \
		echo "Building $$svc..."; \
		ver=$$(cat $(SERVICES_DIR)/$$svc/VERSION 2>/dev/null | tr -d '[:space:]' || echo "dev"); \
		cd $(SERVICES_DIR)/$$svc && go build \
			-ldflags '$(LDFLAGS_BASE) -X github.com/Camionerou/rag-saldivia/pkg/server.Version='"$$ver" \
			-o $(GOBIN)/$$svc ./cmd || exit 1; \
	done
	@echo "All services built → $(GOBIN)/ (sha: $(GIT_SHA))"

build-%: ## Build a specific service (e.g., make build-auth)
	@ver=$$(cat $(SERVICES_DIR)/$*/VERSION 2>/dev/null | tr -d '[:space:]' || echo "dev"); \
	cd $(SERVICES_DIR)/$* && go build \
		-ldflags '$(LDFLAGS_BASE) -X github.com/Camionerou/rag-saldivia/pkg/server.Version='"$$ver" \
		-o $(GOBIN)/$* ./cmd

# ── Testing ──────────────────────────────────────────────────────────────

test: ## Run all Go tests
	go test ./pkg/... -count=1
	@for svc in app erp; do \
		echo "▸ testing services/$$svc"; \
		(cd services/$$svc && go test ./... -count=1) || exit 1; \
	done
	@for tool in cli mcp pkg; do \
		echo "▸ testing tools/$$tool"; \
		(cd tools/$$tool && go test ./... -count=1) || exit 1; \
	done

test-%: ## Run tests for a specific service (e.g., make test-auth)
	cd $(SERVICES_DIR)/$* && go test ./... -count=1 -v

test-coverage: ## Run tests with coverage report
	go test \
		github.com/Camionerou/rag-saldivia/pkg/... \
		github.com/Camionerou/rag-saldivia/services/app/... \
		github.com/Camionerou/rag-saldivia/services/erp/... \
		github.com/Camionerou/rag-saldivia/tools/cli/... \
		github.com/Camionerou/rag-saldivia/tools/mcp/... \
		github.com/Camionerou/rag-saldivia/tools/pkg/... \
		-count=1 -coverprofile=coverage.out
	go tool cover -html=coverage.out -o cover.html
	@echo "Coverage report → cover.html"

test-integration: ## Run integration tests (requires Docker)
	go test \
		github.com/Camionerou/rag-saldivia/services/app/... \
		github.com/Camionerou/rag-saldivia/services/erp/... \
		-tags=integration -count=1 -v

test-frontend: ## Run frontend tests
	cd apps/web && bun test

test-e2e: ## Run E2E tests (Playwright)
	cd apps/web && bunx playwright test

test-workstation: ## Run smoke tests (API + E2E) against the workstation
	@bash scripts/test-workstation.sh

test-workstation-api: ## Run only API smoke against workstation
	@bash scripts/test-workstation.sh api

test-workstation-e2e: ## Run only E2E smoke against workstation
	@bash scripts/test-workstation.sh e2e

test-storage: ## Run storage tests (requires MinIO running)
	cd $(ROOT_DIR)/pkg && go test ./storage/... -v -count=1

test-guardrails: ## Run guardrails tests
	cd $(ROOT_DIR)/pkg && go test ./guardrails/... -v -count=1

test-extractor: ## Run extractor tests (Python, no GPU needed)
	cd $(SERVICES_DIR)/extractor && .venv/bin/python -m pytest tests/ -v

test-all: test test-frontend test-e2e ## Run all test suites

# ── Linting ──────────────────────────────────────────────────────────────

lint: ## Lint all Go code
	@echo "::group::Linting pkg/"
	@cd $(ROOT_DIR) && golangci-lint run ./pkg/...
	@echo "::endgroup::"
	@for svc in $(GO_SERVICES); do \
		echo "▸ linting services/$$svc"; \
		cd $(SERVICES_DIR)/$$svc && golangci-lint run ./... || exit 1; \
	done
	@for tool in tools/*/; do \
		[ -f "$(ROOT_DIR)/$$tool/go.mod" ] || continue; \
		echo "▸ linting $$tool"; \
		cd $(ROOT_DIR)/$$tool && golangci-lint run ./... || exit 1; \
	done

lint-%: ## Lint a specific service
	cd $(SERVICES_DIR)/$* && golangci-lint run ./...

lint-frontend: ## Lint frontend code
	cd apps/web && bun run lint

# ── Code Generation ─────────────────────────────────────────────────────

sqlc: ## Generate Go code from SQL queries (all services)
	@for svc in $(GO_SERVICES); do \
		if [ -f "$(SERVICES_DIR)/$$svc/db/sqlc.yaml" ]; then \
			echo "sqlc generate → $$svc"; \
			cd $(SERVICES_DIR)/$$svc/db && sqlc generate || exit 1; \
		fi; \
	done

sqlc-%: ## Generate sqlc for a specific service
	cd $(SERVICES_DIR)/$*/db && sqlc generate

# ── Events codegen (Plan 26 spine) ───────────────────────────────────────

events-gen: ## Regenerate Go/TS/Markdown from pkg/events/spec/*.cue
	@cd $(ROOT_DIR)/tools/eventsgen && go run . \
		-spec $(ROOT_DIR)/pkg/events/spec \
		-out-go $(ROOT_DIR)/pkg/events/gen \
		-out-ts $(ROOT_DIR)/apps/web/src/lib/events/gen \
		-out-docs $(ROOT_DIR)/docs/events

events-validate: ## Verify generated events match spec (used by CI)
	@tmpdir=$$(mktemp -d); \
	cd $(ROOT_DIR)/tools/eventsgen && go run . \
		-spec $(ROOT_DIR)/pkg/events/spec \
		-out-go $$tmpdir/go -out-ts $$tmpdir/ts -out-docs $$tmpdir/docs; \
	diff -r $$tmpdir/go $(ROOT_DIR)/pkg/events/gen >/dev/null || { echo "pkg/events/gen out of date — run 'make events-gen'"; rm -rf $$tmpdir; exit 1; }; \
	diff -r $$tmpdir/ts $(ROOT_DIR)/apps/web/src/lib/events/gen --exclude=envelope.ts >/dev/null || { echo "apps/web/src/lib/events/gen out of date"; rm -rf $$tmpdir; exit 1; }; \
	diff -r $$tmpdir/docs $(ROOT_DIR)/docs/events --exclude=README.md >/dev/null || { echo "docs/events out of date"; rm -rf $$tmpdir; exit 1; }; \
	rm -rf $$tmpdir; \
	echo "events: generated files match specs"

# ── Database ─────────────────────────────────────────────────────────────

migrate: ## Run database migrations (platform + tenant)
	$(DEPLOY_DIR)/scripts/migrate.sh

seed: ## Seed development data (users, roles, tenant)
	$(DEPLOY_DIR)/scripts/seed.sh

migrate-seed: migrate seed ## Run migrations + seed in one step

# ── Deploy ───────────────────────────────────────────────────────────────

deploy-gen: ## Generate Traefik/Cloudflare configs from templates + .env
	@bash $(DEPLOY_DIR)/scripts/gen-config.sh

deploy-preflight: ## Run pre-deploy validation checks
	@bash $(DEPLOY_DIR)/scripts/preflight.sh

deploy-dev: ## Start development stack (no gen — dev uses Docker Compose env substitution)
	docker compose -f $(DEPLOY_DIR)/docker-compose.dev.yml up -d

deploy-prod: deploy-preflight deploy-gen ## Build + deploy + verify all services
	bash $(DEPLOY_DIR)/scripts/deploy.sh

deploy-stop: ## Stop all running services
	docker compose -f $(DEPLOY_DIR)/docker-compose.prod.yml down 2>/dev/null || true
	docker compose -f $(DEPLOY_DIR)/docker-compose.dev.yml down 2>/dev/null || true

deploy: ## Deploy all services to production (legacy — use deploy-prod)
	$(GOBIN)/sda deploy --all

deploy-%: ## Deploy a specific service
	$(GOBIN)/sda deploy $*

rollback-%: ## Rollback a specific service
	$(GOBIN)/sda rollback $*

versions: ## Show running vs expected service versions
	@GIT_SHA=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	echo ""; \
	echo "Current HEAD: $$GIT_SHA"; \
	echo ""; \
	printf "%-20s %-10s %-10s %-22s %s\n" "SERVICE" "VERSION" "GIT SHA" "BUILD TIME" "STATUS"; \
	echo "────────────────────────────────────────────────────────────────────────────────"; \
	for entry in \
		"8013:erp" "8020:app"; do \
		port=$$(echo $$entry | cut -d: -f1); \
		name=$$(echo $$entry | cut -d: -f2); \
		info=$$(curl -sf --max-time 2 http://localhost:$$port/v1/info 2>/dev/null || echo ""); \
		if [ -z "$$info" ]; then \
			printf "%-20s %-10s %-10s %-22s \033[31mDOWN\033[0m\n" "$$name" "-" "-" "-"; \
		else \
			ver=$$(echo "$$info" | grep -o '"version":"[^"]*"' | cut -d'"' -f4); \
			sha=$$(echo "$$info" | grep -o '"git_sha":"[^"]*"' | cut -d'"' -f4); \
			btime=$$(echo "$$info" | grep -o '"build_time":"[^"]*"' | cut -d'"' -f4); \
			if [ "$$sha" = "$$GIT_SHA" ]; then \
				status="\033[32mMATCH\033[0m"; \
			elif [ "$$sha" = "unknown" ] || [ -z "$$sha" ]; then \
				status="\033[33mNO INFO\033[0m"; \
			else \
				status="\033[31mSTALE\033[0m"; \
			fi; \
			printf "%-20s %-10s %-10s %-22s $$status\n" "$$name" "$$ver" "$$sha" "$$btime"; \
		fi; \
	done

# ── Scaffolding ──────────────────────────────────────────────────────────

new-service: ## Create a new service (make new-service NAME=billing)
ifndef NAME
	$(error NAME is required. Usage: make new-service NAME=billing)
endif
	@echo "Creating service: $(NAME)"
	@cp -r $(SERVICES_DIR)/.scaffold $(SERVICES_DIR)/$(NAME)
	@find $(SERVICES_DIR)/$(NAME) -type f -exec sed -i 's/scaffold/$(NAME)/g' {} +
	@echo "0.1.0" > $(SERVICES_DIR)/$(NAME)/VERSION
	@echo "Service created → services/$(NAME)/"
	@echo "Next: initialize go.mod and add to go.work"

# ── Security ─────────────────────────────────────────────────────────────

security: ## Run security scans
	gosec ./services/... ./pkg/...
	trivy fs --scanners vuln .

# ── Cleanup ──────────────────────────────────────────────────────────────

clean: ## Remove build artifacts
	rm -rf $(GOBIN) coverage.out cover.html
	@for svc in $(GO_SERVICES); do \
		rm -rf $(SERVICES_DIR)/$$svc/tmp; \
	done

# ── Status ───────────────────────────────────────────────────────────────

status: ## Full system status — infra, services, frontend, GPU
	@echo ""
	@echo "╔══════════════════════════════════════════════════════════╗"
	@echo "║               SDA Framework — System Status             ║"
	@echo "╚══════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "── Infrastructure (Docker) ──────────────────────────────────"
	@docker ps --format "  {{.Names}}\t{{.Status}}" 2>/dev/null | column -t -s$$'\t' || echo "  Docker not running"
	@echo ""
	@echo "── Go Services ─────────────────────────────────────────────"
	@for entry in \
		"8013:sda-erp" \
		"8020:sda-app"; do \
		port=$$(echo $$entry | cut -d: -f1); \
		name=$$(echo $$entry | cut -d: -f2); \
		code=$$(curl -s --max-time 1 -o /dev/null -w "%{http_code}" http://localhost:$$port/health 2>/dev/null); \
		if [ "$$code" = "200" ]; then \
			printf "  %-20s :%-5s \033[32mUP\033[0m\n" "$$name" "$$port"; \
		else \
			printf "  %-20s :%-5s \033[31mDOWN\033[0m\n" "$$name" "$$port"; \
		fi; \
	done
	@echo ""
	@echo "── Frontend ────────────────────────────────────────────────"
	@code=$$(curl -s --max-time 2 -o /dev/null -w "%{http_code}" http://localhost:3000 2>/dev/null); \
	if [ "$$code" = "200" ]; then \
		printf "  %-20s :%-5s \033[32mUP\033[0m\n" "next.js" "3000"; \
	else \
		printf "  %-20s :%-5s \033[31mDOWN\033[0m\n" "next.js" "3000"; \
	fi
	@echo ""
	@echo "── GPU ─────────────────────────────────────────────────────"
	@nvidia-smi --query-gpu=name,memory.used,memory.total --format=csv,noheader 2>/dev/null || echo "  No GPU detected"
	@echo ""

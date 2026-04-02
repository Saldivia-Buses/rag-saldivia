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

export GOBIN

.PHONY: help dev stop test lint build proto migrate deploy new-service clean versions

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ── Development ──────────────────────────────────────────────────────────

dev: ## Start infra only (run Go services on host)
	docker compose -f $(DEPLOY_DIR)/docker-compose.dev.yml up

dev-full: ## Start infra + all Go services in Docker
	docker compose -f $(DEPLOY_DIR)/docker-compose.dev.yml --profile full up --build

stop: ## Stop all services
	docker compose -f $(DEPLOY_DIR)/docker-compose.dev.yml --profile full down

# ── Build ────────────────────────────────────────────────────────────────

build: ## Build all Go services
	@for svc in $(GO_SERVICES); do \
		echo "Building $$svc..."; \
		cd $(SERVICES_DIR)/$$svc && go build -o $(GOBIN)/$$svc ./cmd/... || exit 1; \
	done
	@echo "All services built → $(GOBIN)/"

build-%: ## Build a specific service (e.g., make build-auth)
	cd $(SERVICES_DIR)/$* && go build -o $(GOBIN)/$* ./cmd/...

# ── Testing ──────────────────────────────────────────────────────────────

test: ## Run all Go tests
	go test ./services/... ./pkg/... ./tools/... -count=1

test-%: ## Run tests for a specific service (e.g., make test-auth)
	cd $(SERVICES_DIR)/$* && go test ./... -count=1 -v

test-coverage: ## Run tests with coverage report
	go test ./services/... ./pkg/... ./tools/... -count=1 -coverprofile=coverage.out
	go tool cover -html=coverage.out -o cover.html
	@echo "Coverage report → cover.html"

test-integration: ## Run integration tests (requires Docker)
	go test ./services/... -tags=integration -count=1 -v

test-frontend: ## Run frontend tests
	cd apps/web && bun test

test-e2e: ## Run E2E tests (Playwright)
	cd apps/web && bunx playwright test

test-all: test test-frontend test-e2e ## Run all test suites

# ── Linting ──────────────────────────────────────────────────────────────

lint: ## Lint all Go code
	golangci-lint run ./services/... ./pkg/... ./tools/...

lint-%: ## Lint a specific service
	cd $(SERVICES_DIR)/$* && golangci-lint run ./...

lint-frontend: ## Lint frontend code
	cd apps/web && bun run lint

# ── Code Generation ─────────────────────────────────────────────────────

proto: ## Generate gRPC code from proto files
	@echo "Generating protobuf code..."
	buf generate proto/

sqlc: ## Generate Go code from SQL queries (all services)
	@for svc in $(GO_SERVICES); do \
		if [ -f "$(SERVICES_DIR)/$$svc/db/sqlc.yaml" ]; then \
			echo "sqlc generate → $$svc"; \
			cd $(SERVICES_DIR)/$$svc/db && sqlc generate || exit 1; \
		fi; \
	done

sqlc-%: ## Generate sqlc for a specific service
	cd $(SERVICES_DIR)/$*/db && sqlc generate

# ── Database ─────────────────────────────────────────────────────────────

migrate: ## Run migrations for all tenants
	@echo "Running migrations..."
	$(GOBIN)/sda db migrate --tenant all

migrate-%: ## Run migrations for a specific service
	$(GOBIN)/sda db migrate --service $*

seed: ## Seed development data
	$(GOBIN)/sda db seed

# ── Deploy ───────────────────────────────────────────────────────────────

deploy: ## Deploy all services to production
	$(GOBIN)/sda deploy --all

deploy-%: ## Deploy a specific service
	$(GOBIN)/sda deploy $*

rollback-%: ## Rollback a specific service
	$(GOBIN)/sda rollback $*

versions: ## Show running vs available versions
	$(GOBIN)/sda versions

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

status: ## Show status of all services
	@docker compose -f $(DEPLOY_DIR)/docker-compose.dev.yml ps 2>/dev/null || echo "No services running"
	@echo ""
	@echo "GPU:"
	@nvidia-smi --query-gpu=name,memory.used,memory.total --format=csv,noheader 2>/dev/null || echo "  No GPU detected"

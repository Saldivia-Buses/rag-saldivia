# RAG Saldivia — Unified Interface
# Usage: make <target> [ARGS]

SHELL := /bin/bash
.DEFAULT_GOAL := help

SALDIVIA_ROOT := $(shell pwd)
BLUEPRINT_DIR := $(SALDIVIA_ROOT)/vendor/rag-blueprint
COMPOSE_DIR := $(BLUEPRINT_DIR)/deploy/compose
PROFILE ?= workstation-1gpu
BLUEPRINT_VERSION ?= 2.5.0

export SALDIVIA_ROOT

.PHONY: help setup deploy stop restart status health ingest query test test-unit test-coverage test-e2e test-e2e-brev test-backend test-stress patch-check patch-create clean validate show-env watch cli

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Clone blueprint, apply patches, build images
	@./scripts/setup.sh $(BLUEPRINT_VERSION)

deploy: ## Start services (PROFILE=workstation-1gpu)
	@./scripts/deploy.sh $(PROFILE)

stop: ## Stop all services
	@cd $(COMPOSE_DIR) && docker compose --env-file .env.merged \
		-f docker-compose-rag-server.yaml \
		-f $(SALDIVIA_ROOT)/config/compose-overrides.yaml down

restart: ## Stop and redeploy (PROFILE=workstation-1gpu)
	$(MAKE) stop && $(MAKE) deploy PROFILE=$(PROFILE)

status: ## Show GPU, Docker, and Milvus status
	@echo "=== GPU ===" && nvidia-smi --query-gpu=index,memory.used,memory.total --format=csv,noheader 2>/dev/null || echo "No GPU"
	@echo ""
	@echo "=== Docker ===" && docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' 2>/dev/null | head -20
	@echo ""
	@echo "=== RAG Health ===" && curl -sf http://localhost:8081/health 2>/dev/null | python3 -m json.tool 2>/dev/null || echo "RAG server not responding"


health: ## Run health check on all services
	@bash $(SALDIVIA_ROOT)/scripts/health_check.sh
ingest: ## Smart ingest PDFs (DOCS=path COLLECTION=name)
	@python3 $(SALDIVIA_ROOT)/scripts/smart_ingest.py \
		$(or $(COLLECTION),tecpia) \
		$(or $(DOCS),$(error DOCS is required. Usage: make ingest DOCS=~/docs/pdfs/))

query: ## Crossdoc query (Q="question")
	@python3 $(SALDIVIA_ROOT)/scripts/crossdoc_client.py \
		$(or $(Q),$(error Q is required. Usage: make query Q="your question here"))

## ── Testing ──────────────────────────────────────────────────────────────────
test: test-unit test-backend ## Run unit + component + backend tests (no E2E — E2E requires app running)

test-unit: ## Run Vitest unit + component tests (frontend)
	cd services/sda-frontend && npm run test

test-coverage: ## Run Vitest con reporte de coverage (falla si <80%)
	cd services/sda-frontend && npm run test:coverage
	@echo "Coverage report: services/sda-frontend/coverage/index.html"

test-e2e: ## Run Playwright E2E tests (build + preview + tests)
	cd services/sda-frontend && npm run build && npm run preview & \
	sleep 5 && npx playwright test; \
	kill %1

test-e2e-brev: ## Run E2E tests against Brev instance (BREV_URL=https://...)
	cd services/sda-frontend && PLAYWRIGHT_BASE_URL=$(BREV_URL) npx playwright test

test-backend: ## Run Python pytest tests (saldivia SDK)
	uv run pytest saldivia/tests/ -v

test-stress: ## Run HTTP stress test against running gateway
	@python3 $(SALDIVIA_ROOT)/scripts/stress_test.py

patch-check: ## Validate patches without applying (dry-run)
	@cd $(BLUEPRINT_DIR) && \
	for patch in $(SALDIVIA_ROOT)/patches/frontend/patches/*.patch; do \
		[ -f "$$patch" ] || continue; \
		if git apply --check "$$patch" 2>/dev/null; then \
			echo "OK: $$(basename $$patch)"; \
		else \
			echo "FAIL: $$(basename $$patch)"; \
		fi; \
	done

patch-create: ## Generate patches from current blueprint changes
	@cd $(BLUEPRINT_DIR) && \
	git diff --cached > /tmp/saldivia-patches.patch && \
	echo "Patch saved to /tmp/saldivia-patches.patch"

validate: ## Validate config for PROFILE (PROFILE=workstation-1gpu)
	@python3 -c "from saldivia.config import ConfigLoader, validate_config; \
		c = ConfigLoader('config').load('$(PROFILE)'); \
		errors = validate_config(c); \
		print('OK' if not errors else '\n'.join(errors))"

show-env: ## Show generated env vars for PROFILE
	@python3 -c "from saldivia.config import ConfigLoader; \
		env = ConfigLoader('config').generate_env('$(PROFILE)'); \
		print('\n'.join(f'{k}={v}' for k,v in sorted(env.items())))"

watch: ## Watch folder for auto-ingest (COLLECTION=name)
	python -m saldivia.watch ./watch $(COLLECTION)

cli: ## Run CLI command (ARGS="command [options]")
	python -m cli.main $(ARGS)

clean: ## Remove blueprint clone and build artifacts
	@echo "This will delete the blueprint/ directory. Are you sure? (Ctrl+C to cancel)"
	@read -r
	@rm -rf $(BLUEPRINT_DIR)
	@echo "Cleaned."

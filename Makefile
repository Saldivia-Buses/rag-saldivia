# RAG Saldivia — Unified Interface
# Usage: make <target> [ARGS]

SHELL := /bin/bash
.DEFAULT_GOAL := help

SALDIVIA_ROOT := $(shell pwd)
BLUEPRINT_DIR := $(SALDIVIA_ROOT)/blueprint
COMPOSE_DIR := $(BLUEPRINT_DIR)/deploy/compose
PROFILE ?= brev-2gpu
BLUEPRINT_VERSION ?= 2.5.0

export SALDIVIA_ROOT

.PHONY: help setup deploy stop status ingest query test patch-check patch-create clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Clone blueprint, apply patches, build images
	@./scripts/setup.sh $(BLUEPRINT_VERSION)

deploy: ## Start services (PROFILE=brev-2gpu|workstation-1gpu)
	@./scripts/deploy.sh $(PROFILE)

stop: ## Stop all services
	@cd $(COMPOSE_DIR) && docker compose --env-file .env.merged \
		-f docker-compose-rag-server.yaml \
		-f $(SALDIVIA_ROOT)/config/compose-overrides.yaml down

status: ## Show GPU, Docker, and Milvus status
	@echo "=== GPU ===" && nvidia-smi --query-gpu=index,memory.used,memory.total --format=csv,noheader 2>/dev/null || echo "No GPU"
	@echo ""
	@echo "=== Docker ===" && docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' 2>/dev/null | head -20
	@echo ""
	@echo "=== RAG Health ===" && curl -sf http://localhost:8081/health 2>/dev/null | python3 -m json.tool 2>/dev/null || echo "RAG server not responding"

ingest: ## Smart ingest PDFs (DOCS=path COLLECTION=name)
	@python3 $(SALDIVIA_ROOT)/scripts/smart_ingest.py \
		$(or $(COLLECTION),tecpia) \
		$(or $(DOCS),$(error DOCS is required. Usage: make ingest DOCS=~/docs/pdfs/))

query: ## Crossdoc query (Q="question")
	@python3 $(SALDIVIA_ROOT)/scripts/crossdoc_client.py \
		$(or $(Q),$(error Q is required. Usage: make query Q="your question here"))

test: ## Run stress test
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

clean: ## Remove blueprint clone and build artifacts
	@echo "This will delete the blueprint/ directory. Are you sure? (Ctrl+C to cancel)"
	@read -r
	@rm -rf $(BLUEPRINT_DIR)
	@echo "Cleaned."

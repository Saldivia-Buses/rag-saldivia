# RAG Saldivia

Production-ready overlay on NVIDIA RAG Blueprint v2.5.0 with authentication, RBAC, multi-collection support, SvelteKit 5 frontend, and multiple deployment profiles.

[![CI](https://img.shields.io/badge/CI-passing-brightgreen)](https://github.com/Camionerou/rag-saldivia)
[![Coverage](https://img.shields.io/badge/coverage-80%25-green)](https://github.com/Camionerou/rag-saldivia)

## What it is

RAG Saldivia extends the NVIDIA RAG Blueprint v2.5.0 with authentication (JWT + RBAC), multi-collection vector storage, a modern SvelteKit 5 frontend, Python CLI/SDK, and support for 2-GPU, 1-GPU (with dynamic mode switching), and cloud-only deployments.

## Architecture

```
User
  |
  | HTTPS (JWT cookie)
  v
SDA Frontend (port 3000, SvelteKit 5 BFF)
  |
  | HTTP Bearer token
  v
Auth Gateway (port 9000, FastAPI)
  |
  | Proxy with RBAC
  v
RAG Server (port 8081, NVIDIA Blueprint)
  |
  v
Milvus (vector DB) + NIMs (embed, rerank, OCR) + LLM (Nemotron-3 or external API)
```

## Quick Start

```bash
# 1. Clone
git clone git@github.com:Camionerou/rag-saldivia.git && cd rag-saldivia
cp .env.example .env.local  # Add your NGC_API_KEY, JWT_SECRET, etc.

# 2. Bootstrap (instala Docker, NVIDIA Container Toolkit, Node.js, pnpm si faltan)
sudo ./scripts/bootstrap.sh

# 3. Deploy
make deploy PROFILE=workstation-1gpu # 1-GPU (RunPod / workstation local)
# OR
make deploy PROFILE=full-cloud       # sin GPU, todo via API

# 3. Verify
make status

# 4. Ingest documents
make ingest DOCS=~/docs/pdfs/ COLLECTION=my-collection

# 5. Query
make query Q="What is the main topic of the documents?"
```

## Documentation

| Doc | Description |
|-----|-------------|
| [Architecture](docs/architecture.md) | Service map, request flow, design decisions |
| [Development Workflow](docs/development-workflow.md) | How to contribute and build features |
| [Testing](docs/testing.md) | How to run and write tests |
| [Deployment](docs/deployment.md) | Profiles, Brev, environment variables |
| [Contributing](docs/contributing.md) | Code conventions, commits, PRs |

## Roadmap

| Phase | Status | Description |
|-------|--------|-------------|
| 1 | ✅ Done | Foundation: Auth Gateway, JWT, RBAC, SQLite AuthDB, SvelteKit 5 BFF |
| 2 | ✅ Done | Chat Pro: SSE streaming, chat history, basic UI |
| 3 | ✅ Done | Collections: CRUD, detail view, CollectionCard grid, delete modal |
| 4 | ✅ Done | Upload: Drag-and-drop, multipart upload, BFF → gateway proxy |
| 5 | ✅ Done | Crossdoc: Client-side decomposition, 4-phase pipeline, progress UI |
| 6 | 🚧 In Progress | Testing: Vitest unit tests, component tests, Playwright E2E |
| 7 | 📋 Planned | Admin Panel: User management, API keys, audit log |
| 8 | 📋 Planned | MCP Integration: Claude Desktop integration, 6 tools |
| 9 | 📋 Planned | Smart Ingest: Tier system, deadlock detection, auto-resume |
| 10+ | 📋 Planned | See docs/superpowers/specs/ for upcoming phases |

## What It Solves

| Problem | Solution |
|---------|----------|
| Reranker kills cross-doc diversity | Client-side crossdoc decomposition |
| NV-Ingest deadlocks | Smart ingest with serial batching and auto-split |
| Generic prompts | Custom Envie persona with strict instructions |
| GPU_CAGRA causes VRAM pressure | HNSW on CPU + hybrid search |
| Milvus wastes 3.7 GB VRAM | GPU memory pool disabled |

## Deployment Profiles

| Profile | Hardware | LLM | Use Case |
|---------|----------|-----|----------|
| `brev-2gpu` | 2x RTX PRO 6000 Blackwell | Nemotron-3 (local) | Production on Brev |
| `workstation-1gpu` | 1x GPU (≥98 GB VRAM) | External API | Development workstation |
| `full-cloud` | No GPU | External API | Cloud-only |

## 1-GPU Mode

The `workstation-1gpu` profile uses a mode manager to switch between QUERY and INGEST modes:

| Mode | VRAM | Active |
|------|------|--------|
| QUERY | ~46 GB | NIMs (embed + rerank) only |
| INGEST | ~90 GB | NIMs + VLM (Qwen3-VL-8B) |

The mode manager monitors the Redis ingestion queue and automatically loads/unloads the VLM to free memory. When idle for 5 minutes, it switches back to QUERY mode.

## CLI Usage

```bash
# Collections
rag-saldivia collections list
rag-saldivia collections create my-collection
rag-saldivia collections stats my-collection

# Ingestion queue
rag-saldivia ingest add /path/to/docs/ --collection my-collection
rag-saldivia ingest queue
rag-saldivia ingest watch /watch/dir/ --collection my-collection

# Platform status
rag-saldivia status
```

## Structure

```
saldivia/        — Python SDK (ConfigLoader, ProviderClient, ModeManager, etc.)
cli/             — Click CLI (collections, ingest, status, mcp)
config/          — YAML config files + profiles (brev-2gpu, workstation-1gpu, full-cloud)
services/        — Additional docker services (mode-manager, openrouter-proxy)
patches/         — Frontend patches and new files
scripts/         — deploy.sh, smart_ingest.py, crossdoc_client.py, stress_test.py
docs/            — Documentation (architecture, workflow, testing, deployment)
```

## License

MIT

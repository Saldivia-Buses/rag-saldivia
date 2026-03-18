# RAG Saldivia

Reproducible customization overlay for the [NVIDIA RAG Blueprint](https://github.com/NVIDIA/rag-blueprint) v2.5.0.

## Quick Start

```bash
# 1. Clone this repo
git clone git@github.com:enzosaldivia/rag-saldivia.git
cd rag-saldivia

# 2. Create your secrets file
cp .env.example .env.local
# Edit .env.local with your NGC_API_KEY and other secrets

# 3. Setup (clone blueprint, apply patches, build images)
make setup

# 4. Deploy
make deploy PROFILE=brev-2gpu       # 2-GPU Brev instance
make deploy PROFILE=workstation-1gpu # 1-GPU workstation (LLM via API)

# 5. Use
make ingest DOCS=~/docs/pdfs/       # Ingest documents
make query Q="your question"        # Cross-document query
make status                          # Check health
```

## What It Solves

| Problem | Solution |
|---------|----------|
| Reranker kills cross-doc diversity | Client-side crossdoc decomposition |
| NV-Ingest deadlocks | Smart ingest with serial batching and auto-split |
| Generic prompts | Custom Envie persona with strict instructions |
| GPU_CAGRA causes VRAM pressure | HNSW on CPU + hybrid search |
| Milvus wastes 3.7 GB VRAM | GPU memory pool disabled |

## Deployment Profiles

- **`brev-2gpu`**: 2x RTX PRO 6000 Blackwell. All local.
- **`workstation-1gpu`**: 1x GPU for NIMs + VLM. LLM via NVIDIA API / OpenRouter.
- **`full-cloud`**: No GPU. All services via external APIs.

## 1-GPU Mode

The `workstation-1gpu` profile runs on a single GPU (~98 GB VRAM) using a dynamic mode manager:

| Mode | VRAM | Active |
|------|------|--------|
| QUERY | ~46 GB | NIMs (embed + rerank) only |
| INGEST | ~90 GB | NIMs + VLM (Qwen3-VL-8B) |

The mode manager monitors the Redis ingestion queue and loads/unloads the VLM automatically.
When idle for 5 minutes, it switches back to QUERY mode to free memory.

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

## MCP Integration

The MCP server exposes 6 tools for AI assistant integration (Claude, etc.):

```bash
# Start MCP server
make mcp
# or
python -m saldivia.mcp_server
```

Tools available: `search_documents`, `ask_question`, `list_collections`,
`collection_stats`, `ingest_document`, `ingestion_status`.

Add to Claude Desktop config:
```json
{
  "mcpServers": {
    "rag-saldivia": {
      "command": "python",
      "args": ["-m", "saldivia.mcp_server"],
      "cwd": "/path/to/rag-saldivia"
    }
  }
}
```

## Make Targets

```
make setup            — Clone blueprint, apply patches, build images
make deploy           — Start services (PROFILE=brev-2gpu|workstation-1gpu|full-cloud)
make stop             — Stop all services
make status           — GPU, Docker, health status
make validate         — Validate config for PROFILE
make show-env         — Show generated env vars for PROFILE
make ingest           — Smart PDF ingestion (DOCS=path COLLECTION=name)
make query            — Cross-document query (Q="question")
make mcp              — Start MCP server
make watch            — Watch folder for auto-ingest (COLLECTION=name)
make cli              — Run CLI command (ARGS="...")
make test             — Stress test
```

## Structure

```
saldivia/        — Python SDK (ConfigLoader, ProviderClient, ModeManager, etc.)
cli/             — Click CLI (collections, ingest, status, mcp)
config/          — YAML config files + profiles (brev-2gpu, workstation-1gpu, full-cloud)
services/        — Additional docker services (mode-manager, openrouter-proxy)
patches/         — Frontend patches and new files
scripts/         — deploy.sh, smart_ingest.py, crossdoc_client.py, stress_test.py
docs/            — Problems and solutions documentation
```

See `docs/problems-and-solutions.md` for detailed documentation of each problem and fix.

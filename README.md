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
- **`workstation-1gpu`**: 1x GPU for NIMs + VLM. LLM via OpenRouter/NVIDIA API.

## Make Targets

```
make setup    — Clone blueprint, apply patches, build images
make deploy   — Start services with profile
make stop     — Stop all services
make status   — GPU, Docker, health status
make ingest   — Smart PDF ingestion
make query    — Cross-document query (CLI)
make test     — Stress test
```

## Structure

```
config/          — .env overrides, prompt.yaml, milvus.yaml, compose overrides, profiles
patches/         — Frontend patches and new files
scripts/         — setup.sh, deploy.sh, smart_ingest.py, crossdoc_client.py, stress_test.py
docs/            — Problems and solutions documentation
```

See `docs/problems-and-solutions.md` for detailed documentation of each problem and fix.

# Architecture

## What it is

RAG Saldivia is a production-ready overlay on top of NVIDIA RAG Blueprint v2.5.0. It extends the blueprint with:

- **Authentication and RBAC**: JWT-based authentication, role-based access control (admin, area_manager, user), API key management
- **Multi-collection support**: Isolated vector collections per use case, with per-collection CRUD and stats
- **SvelteKit 5 BFF Frontend**: Modern, type-safe frontend with server-side rendering, SSE streaming, and Svelte 5 runes
- **Python CLI and SDK**: Command-line interface and programmatic access to all platform features
- **Multiple deployment profiles**: Support for 2-GPU, 1-GPU (with dynamic mode switching), and cloud-only deployments

The overlay does not fork the blueprint. The blueprint is included as a git submodule
in `vendor/rag-blueprint/` for reference. The deploy workflow clones the blueprint into
`blueprint/` via `make setup` and applies patches from `patches/`.

## Service Map

| Port | Service | Role |
|------|---------|------|
| 3000 | SDA Frontend | SvelteKit 5 BFF, handles auth cookies, SSE streaming, file uploads |
| 8081 | RAG Server | NVIDIA Blueprint core: /health, /generate (SSE), /search, /chat/completions, /configuration, /summary |
| 8082 | Ingestor Server | Document ingestion pipeline (PDF, DOCX, images) with VLM captioning: /documents (POST/GET/PATCH/DELETE), /status, /collections |
| 9000 | Auth Gateway | FastAPI gateway: JWT validation, RBAC, proxy to RAG Server |
| 19530 | Milvus | Vector database with hybrid search (HNSW on CPU) |
| (internal) | NIMs | Embed (nvidia/llama-nemotron-embed-1b-v2, port 9080), rerank (nvidia/llama-nemotron-rerank-1b-v2, port 1976), OCR |
| (internal) | LLM | External API (NVIDIA API via nvidia-api provider) |
| (internal) | Redis | Ingestion queue for mode manager coordination |

> **Nota de red interna:** El container `auth-gateway` escucha en el puerto interno 8090 (`GATEWAY_PORT`).
> El host expone el puerto 9000 (mapping `9000:8090`). Dentro de la red Docker,
> el frontend BFF se conecta a `http://auth-gateway:8090`.

## Request Flow

```
User
  |
  | HTTPS (JWT cookie)
  v
SDA Frontend (port 3000)
  |
  | HTTP API Key (Bearer, extracted from JWT by BFF)
  v
Auth Gateway (port 9000)
  |
  | Validate API key hash + check RBAC
  | Note: JWT is internal to the BFF; the gateway only sees the API key
  | Proxy request with user context
  v
RAG Server (port 8081)
  |
  | /search or /generate
  v
Milvus (vector DB)
  |
  | Retrieve top-K chunks (hybrid search)
  v
NIMs (embed, rerank)
  |
  | Embed query, rerank results
  v
LLM (nvidia/llama-3.3-nemotron-super-49b-v1.5 or external API)
  |
  | Generate answer with citations
  v
Response
  |
  | Stream via SSE
  v
User
```

## Deployment Profiles

| Profile | Hardware | LLM | Use Case |
|---------|----------|-----|----------|
| `workstation-1gpu` | 1x RTX PRO 6000 Blackwell (96 GB VRAM) | External (NVIDIA API via nvidia-api provider; OpenRouter for crossdoc) | Production — physical workstation Ubuntu 24.04 |

## 1-GPU Mode

The `workstation-1gpu` profile runs on a single GPU with ~98 GB VRAM using a dynamic mode manager that switches between two modes:

| Mode | VRAM Usage | Active Services | Purpose |
|------|------------|-----------------|---------|
| QUERY | ~46 GB | NIMs (embed + rerank) only | Handle user queries in real-time |
| INGEST | ~90 GB | NIMs + VLM (Qwen3-VL-8B) | Process document uploads with VLM captioning |

**Mode Manager Logic:**
1. Monitor Redis ingestion queue every 10 seconds
2. If documents in queue and mode is QUERY → switch to INGEST (load VLM)
3. Process documents in batches (serial, no parallel to avoid deadlock)
4. If queue empty for 5 minutes → switch back to QUERY (unload VLM)

This ensures the system can ingest documents with VLM captioning while keeping query latency low during interactive use.

## Key Architectural Decisions

### 1. Client-side crossdoc decomposition → avoids reranker diversity loss

**Problem:** The reranker collapses diversity when cross-document queries hit the VDB with a single broad query. Top-K results skew to the most relevant document, ignoring others.

**Solution:** Decompose the query into sub-queries on the client (SvelteKit BFF), send one sub-query per document, then synthesize results. This bypasses the reranker's diversity collapse and ensures all documents contribute.

**Tradeoff:** More network round-trips (N sub-queries instead of 1), but better answer quality for multi-document questions.

### 2. HNSW on CPU + hybrid search → avoids GPU_CAGRA VRAM pressure

**Problem:** GPU_CAGRA indexing consumes 3-5 GB VRAM and provides minimal latency benefit for our workload (queries are infrequent, not real-time search).

**Solution:** Use HNSW index on CPU with hybrid search (vector + BM25). This frees ~4 GB VRAM for LLM context while maintaining acceptable retrieval latency (<200ms).

**Tradeoff:** Slightly slower search (50-100ms penalty), but enables larger LLM context and smaller GPU requirements.

### 3. Milvus GPU memory pool disabled → saves 3.7 GB VRAM

**Problem:** Milvus allocates a 3.7 GB GPU memory pool at startup, even when GPU indexing/search is disabled.

**Solution:** Set `MILVUS_GPU_MEMORY_POOL_SIZE=0` in environment. This disables the pool and frees 3.7 GB VRAM.

**Tradeoff:** None. The pool is unused when GPU search is disabled, so disabling it is pure savings.

### 4. Smart ingest with serial batching → avoids NV-Ingest deadlocks

**Problem:** NV-Ingest hangs on batch ingestion when documents are processed in parallel. Root cause: internal resource contention in the Triton inference server.

**Solution:** Implement a tier system (tiny/small/medium/large based on page count), process documents serially within each tier, with adaptive timeout and deadlock detection. Resume from last successful document on failure.

**Tradeoff:** Slower ingestion (serial vs parallel), but 100% reliability and automatic recovery.


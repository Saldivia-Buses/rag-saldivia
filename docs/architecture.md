# Architecture

## What it is

RAG Saldivia is a production-ready overlay on top of NVIDIA RAG Blueprint v2.5.0. It extends the blueprint with:

- **Authentication and RBAC**: JWT-based authentication, role-based access control (admin, user, guest), API key management
- **Multi-collection support**: Isolated vector collections per use case, with per-collection CRUD and stats
- **SvelteKit 5 BFF Frontend**: Modern, type-safe frontend with server-side rendering, SSE streaming, and Svelte 5 runes
- **Python CLI and SDK**: Command-line interface and programmatic access to all platform features
- **Multiple deployment profiles**: Support for 2-GPU, 1-GPU (with dynamic mode switching), and cloud-only deployments

The overlay is designed as a symlink-based wrapper that does not fork the blueprint. It applies patches, adds services, and orchestrates the existing blueprint components.

## Service Map

| Port | Service | Role |
|------|---------|------|
| 3000 | SDA Frontend | SvelteKit 5 BFF, handles auth cookies, SSE streaming, file uploads |
| 8081 | RAG Server | NVIDIA Blueprint core: /generate, /search, /documents endpoints |
| 8082 | NV-Ingest | Document ingestion pipeline (PDF, DOCX, images) with VLM captioning |
| 9000 | Auth Gateway | FastAPI gateway: JWT validation, RBAC, proxy to RAG Server |
| (internal) | Milvus | Vector database with hybrid search (HNSW on CPU) |
| (internal) | NIMs | Embed (nv-embedqa-e5-v5), rerank (nv-rerankqa-mistral-4b-v3), OCR |
| (internal) | LLM | Nemotron-3-Super-120B-A12B (2-GPU) or external API (1-GPU, full-cloud) |
| (internal) | Redis | Ingestion queue for mode manager coordination |

## Request Flow

```
User
  |
  | HTTPS (JWT cookie)
  v
SDA Frontend (port 3000)
  |
  | HTTP Bearer token (extracted from JWT)
  v
Auth Gateway (port 9000)
  |
  | Validate JWT + check RBAC
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
LLM (Nemotron-3 or external API)
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
| `brev-2gpu` | 2x RTX PRO 6000 Blackwell (196 GB total VRAM) | Nemotron-3-Super-120B-A12B (local) | Production on Brev cloud |
| `workstation-1gpu` | 1x GPU (≥98 GB VRAM) | External (NVIDIA API or OpenRouter) | Development workstation with mode switching |
| `full-cloud` | No GPU | External (NVIDIA API or OpenRouter) | Cloud-only, all services via API |

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


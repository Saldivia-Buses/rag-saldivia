# Deployment

## Deployment Profiles

RAG Saldivia supports three deployment profiles, each optimized for different hardware and use cases:

| Profile | Hardware | LLM | NIMs | VLM | Use Case |
|---------|----------|-----|------|-----|----------|
| `brev-2gpu` | 2x RTX PRO 6000 Blackwell (196 GB total VRAM) | Nemotron-3-Super-120B-A12B (local, GPU 1) | Local (GPU 0) | Qwen3-VL-8B (local, GPU 1) | Production on Brev cloud |
| `workstation-1gpu` | 1x GPU (≥98 GB VRAM) | External (NVIDIA API or OpenRouter) | Local (GPU 0) | Qwen3-VL-8B (local, GPU 0, INGEST mode only) | Development workstation with mode switching |
| `full-cloud` | No GPU (CPU only) | External (NVIDIA API or OpenRouter) | External (NVIDIA API) | External (NVIDIA API) | Cloud-only, minimal infra |

**Profile selection:**
- Brev instance with 2 GPUs → `brev-2gpu`
- Local workstation with 1 GPU → `workstation-1gpu`
- Cloud VM with no GPU → `full-cloud`

## Deploying to Brev

### 1. SSH into Brev instance

```bash
ssh nvidia-enterprise-rag-deb106
```

### 2. Pull latest changes

```bash
cd ~/rag-saldivia
git pull origin main
```

### 3. Deploy with profile

```bash
make deploy PROFILE=brev-2gpu
```

This will:
1. Load profile from `config/profiles/brev-2gpu.yaml`
2. Merge with `.env.saldivia` and `.env.local`
3. Generate `.env.merged` for docker-compose
4. Apply patches to blueprint frontend
5. Start all services (RAG Server, Milvus, NIMs, Auth Gateway, SDA Frontend)
6. Wait for health checks (30s timeout)

### 4. Verify deployment

```bash
make status
```

Expected output:

```
=== GPU ===
0, 89123 MiB, 98304 MiB
1, 67234 MiB, 98304 MiB

=== Docker ===
NAMES                       STATUS              PORTS
sda-frontend                Up 2 minutes        0.0.0.0:3000->3000/tcp
auth-gateway                Up 2 minutes        0.0.0.0:9000->9000/tcp
rag-server                  Up 2 minutes        0.0.0.0:8081->8081/tcp
milvus-standalone           Up 2 minutes        0.0.0.0:19530->19530/tcp
...

=== RAG Health ===
{
  "status": "healthy",
  "services": {
    "milvus": "ok",
    "nims": "ok",
    "llm": "ok"
  }
}
```

## Environment Variables

RAG Saldivia uses a three-layer environment variable system:

1. **Blueprint defaults** — built into NVIDIA RAG Blueprint
2. **Saldivia overrides** — `config/.env.saldivia` (tracked in git)
3. **Secrets** — `.env.local` (NOT tracked in git)

### config/.env.saldivia (tracked)

| Variable | Purpose | Example Value |
|----------|---------|---------------|
| `APP_VECTORSTORE_SEARCHTYPE` | Vector search type | `hybrid` |
| `APP_VECTORSTORE_RANKER_TYPE` | Hybrid search ranker | `rrf` |
| `APP_VECTORSTORE_INDEXTYPE` | Milvus index type | `HNSW` |
| `APP_VECTORSTORE_ENABLEGPUSEARCH` | Use GPU for search | `False` |
| `APP_VECTORSTORE_ENABLEGPUINDEX` | Use GPU for indexing | `False` |
| `VECTOR_DB_TOPK` | Vector DB top-K | `100` |
| `APP_RETRIEVER_TOPK` | Retriever top-K (after rerank) | `25` |
| `APP_NVINGEST_EXTRACTIMAGES` | Extract images from PDFs | `True` |
| `APP_NVINGEST_EXTRACTTABLES` | Extract tables from PDFs | `True` |
| `APP_NVINGEST_EXTRACTCHARTS` | Extract charts from PDFs | `True` |
| `ENABLE_NV_INGEST_PARALLEL_BATCH_MODE` | Parallel ingestion (disabled to avoid deadlock) | `False` |
| `APP_VLM_MAX_TOKENS` | VLM caption max tokens | `1024` |
| `APP_VLM_TEMPERATURE` | VLM caption temperature | `0.3` |
| `APP_NVINGEST_CAPTIONMODELNAME` | VLM model name | `qwen3-vl-8b` |
| `ENABLE_VLM_INFERENCE` | Enable VLM inference in RAG Server | `false` |
| `ENABLE_SOURCE_METADATA` | Include source metadata in responses | `true` |
| `FILTER_THINK_TOKENS` | Filter <think> tokens from responses | `true` |
| `ENABLE_QUERY_DECOMPOSITION` | Server-side query decomposition (disabled, handled in frontend) | `false` |
| `ENABLE_QUERYREWRITER` | Query rewriter | `False` |
| `ENABLE_CITATIONS` | Citations in responses | `False` |
| `CONVERSATION_HISTORY` | Number of turns to include in context | `0` |
| `PROMPT_CONFIG_FILE` | Custom prompt config | `/opt/rag-saldivia/prompt.yaml` |
| `SDA_ORIGIN` | Frontend origin (for CORS) | `https://sda.tecpia.local` |

### .env.local (NOT tracked)

**Critical:** These variables contain secrets and must be set manually.

| Variable | Purpose | How to Generate |
|----------|---------|-----------------|
| `JWT_SECRET` | Secret for signing JWT tokens | `python3 -c "import secrets; print(secrets.token_hex(32))"` |
| `SYSTEM_API_KEY` | Admin API key for Auth Gateway | Copy from `data/auth.db` after first run |
| `NGC_API_KEY` | NVIDIA NGC API key | Get from https://ngc.nvidia.com/ |
| `OPENROUTER_API_KEY` | OpenRouter API key (for external LLM) | Get from https://openrouter.ai/ |

**Setup:**

```bash
cd ~/rag-saldivia
cp .env.example .env.local
nano .env.local  # Edit with your secrets
```

**Verify:**

```bash
make show-env PROFILE=brev-2gpu | grep JWT_SECRET
# Should show your secret (not empty)
```

## Makefile Reference

| Target | Description | Example |
|--------|-------------|---------|
| `make setup` | Clone blueprint, apply patches, build images | `make setup` |
| `make deploy` | Start services with profile | `make deploy PROFILE=brev-2gpu` |
| `make stop` | Stop all services | `make stop` |
| `make status` | Show GPU, Docker, and RAG server health | `make status` |
| `make health` | Run health check on all services | `make health` |
| `make validate` | Validate config for profile | `make validate PROFILE=brev-2gpu` |
| `make show-env` | Show merged env vars for profile | `make show-env PROFILE=brev-2gpu` |
| `make ingest` | Smart ingest PDFs | `make ingest DOCS=~/docs/pdfs/ COLLECTION=tecpia` |
| `make query` | Cross-document query | `make query Q="What is RAG?"` |
| `make test` | Run unit + backend tests (pytest + Vitest) | `make test` |
| `make test-stress` | Run HTTP stress test against running gateway | `make test-stress` |
| `make test-unit` | Run unit tests (pytest) | `make test-unit` |
| `make test-backend` | Run backend tests (pytest) | `make test-backend` |
| `make test-coverage` | Run tests with coverage report | `make test-coverage` |
| `make test-e2e` | Run E2E tests (Playwright) | `make test-e2e` |
| `make mcp` | Start MCP server | `make mcp` |
| `make watch` | Watch folder for auto-ingest | `make watch COLLECTION=tecpia` |
| `make cli` | Run CLI command | `make cli ARGS="collections list"` |
| `make patch-check` | Validate patches (dry-run) | `make patch-check` |
| `make patch-create` | Generate patches from blueprint changes | `make patch-create` |
| `make clean` | Remove blueprint clone and artifacts | `make clean` |

## Port Table

| Port | Service | Protocol | Public |
|------|---------|----------|--------|
| 3000 | SDA Frontend | HTTP | Yes (Caddy reverse proxy) |
| 8081 | RAG Server | HTTP | No (internal only, via gateway) |
| 8082 | Ingestor Server | HTTP | No (internal only) |
| 9000 | Auth Gateway | HTTP | No (internal only, accessed by frontend BFF) |
| 19530 | Milvus | gRPC | No (internal only) |

**Production setup:**

```
Internet
  |
  | HTTPS
  v
Caddy (reverse proxy, port 443)
  |
  | HTTP (localhost:3000)
  v
SDA Frontend
```

## Known Gotchas

### 1. Docker network connect fails silently

**Symptom:** `docker network connect --alias` runs without error but container is not on network.

**Root cause:** Container is already connected to the network. Docker does not overwrite aliases.

**Fix:** Disconnect first, then reconnect:

```bash
docker network disconnect rag-net my-container || true
docker network connect --alias my-alias rag-net my-container
```

**Context:** Happens in `scripts/deploy.sh` when redeploying without `make stop`.

### 2. PYTHONPATH not defined + set -u = crash

**Symptom:** Deployment script crashes with "PYTHONPATH: unbound variable".

**Root cause:** `set -u` (exit on undefined variable) treats empty `PYTHONPATH` as an error.

**Fix:** Use `${PYTHONPATH:-}` (default to empty string if undefined):

```bash
export PYTHONPATH="${PYTHONPATH:-}:$(pwd)/saldivia"
```

**Context:** Happens in `scripts/deploy.sh` on fresh Brev instances.

### 3. Milvus downtime on redeploy

**Symptom:** Milvus takes 30-60s to restart, causing ingestion and query failures.

**Root cause:** Docker stops containers gracefully (10s timeout), then forcefully. Milvus needs time to flush data.

**Fix:** Check Milvus health before starting dependent services:

```bash
until curl -sf http://localhost:9091/healthz; do
  echo "Waiting for Milvus..."
  sleep 5
done
```

**Context:** Added to `scripts/deploy.sh` in commit bdd2af4.

## 1-GPU Mode Switching

The `workstation-1gpu` profile uses a mode manager to switch between QUERY and INGEST modes.

### Mode Manager Logic

1. **Monitor:** Check Redis ingestion queue every 10 seconds
2. **Switch to INGEST:** If queue has documents and mode is QUERY
   - Load VLM (Qwen3-VL-8B, consumes ~44 GB additional VRAM)
   - Note: both NIMs remain loaded (no unload step)
   - Process documents serially (no parallel to avoid deadlock)
3. **Switch to QUERY:** If queue empty for 5 minutes
   - Unload VLM (frees ~44 GB VRAM)
   - NIMs remain loaded
   - Ready for user queries

### VRAM Usage Table

| Mode | NIMs (embed + rerank) | VLM (Qwen3-VL-8B) | Total VRAM |
|------|-----------------------|-------------------|------------|
| QUERY | ~46 GB | — | ~46 GB |
| INGEST | ~46 GB (embed + rerank) | ~44 GB | ~90 GB |

**Note:** Reranker NIM (~42 GB) and embed NIM (~4 GB) both stay loaded in both modes. Total NIMs footprint: ~46 GB.

### Manual Mode Control

The mode manager runs as a Docker service (`mode-manager`). To manually control modes:

```bash
# Check current mode
docker exec mode-manager cat /tmp/current_mode

# Force switch to INGEST (for testing)
docker exec mode-manager python3 -c "from saldivia.mode_manager import switch_to_ingest; switch_to_ingest()"

# Force switch to QUERY
docker exec mode-manager python3 -c "from saldivia.mode_manager import switch_to_query; switch_to_query()"
```

### Disabling Mode Manager

To disable mode switching (keep both NIMs and VLM loaded):

1. Edit `config/profiles/workstation-1gpu.yaml`
2. Set `mode_manager.enabled: false`
3. Redeploy: `make deploy PROFILE=workstation-1gpu`

**Tradeoff:** Uses ~90 GB VRAM permanently, but no ingestion delay.


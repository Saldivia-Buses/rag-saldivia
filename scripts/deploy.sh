#!/usr/bin/env bash
set -euo pipefail

# RAG Saldivia — Deploy Script
# Starts services with compose + overrides for the selected profile.
#
# Usage:
#   ./scripts/deploy.sh [PROFILE]
#   ./scripts/deploy.sh brev-2gpu       # default
#   ./scripts/deploy.sh workstation-1gpu

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SALDIVIA_ROOT="$(dirname "$SCRIPT_DIR")"
BLUEPRINT_DIR="${SALDIVIA_ROOT}/vendor/rag-blueprint"

# Activate uv venv if present (contains httpx, pyyaml, etc.)
if [ -f "${SALDIVIA_ROOT}/.venv/bin/activate" ]; then
    source "${SALDIVIA_ROOT}/.venv/bin/activate"
fi
PROFILE="${1:-workstation-1gpu}"
COMPOSE_DIR="${BLUEPRINT_DIR}/deploy/compose"

log() { echo "[$(date +%H:%M:%S)] $*"; }
err() { echo "[$(date +%H:%M:%S)] ERROR: $*" >&2; exit 1; }
warn() { echo "[$(date +%H:%M:%S)] WARN: $*" >&2; }

[ -d "$BLUEPRINT_DIR" ] || err "Blueprint not found. Run: make setup"
[ -f "${SALDIVIA_ROOT}/config/profiles/${PROFILE}.yaml" ] || err "Unknown profile: ${PROFILE}"

# --- Step 1: Generate .env.merged = .env.saldivia + Python-config + .env.local + runtime ---
export PYTHONPATH="${PYTHONPATH:-}:${SALDIVIA_ROOT}"

GPU_COUNT=$(nvidia-smi -L 2>/dev/null | wc -l || echo "0")
log "Detected ${GPU_COUNT} GPU(s)"

ENV_FILE="${COMPOSE_DIR}/.env.merged"
log "Generating .env.merged (profile: ${PROFILE})..."

# Base: Saldivia env overrides (HNSW, hybrid search, VLM config, NV-Ingest, features).
# These are the 23+ vars that Blueprint doesn't know about.
cp "${SALDIVIA_ROOT}/config/.env.saldivia" "$ENV_FILE"
echo "" >> "$ENV_FILE"

# Override: model names and endpoints from YAML config (profile-aware).
# generate_env() outputs ~15 vars (APP_LLM_MODELNAME, APP_EMBEDDINGS_*, etc.)
# that override any matching vars from .env.saldivia above.
python3 -c "
from saldivia.config import ConfigLoader
loader = ConfigLoader('${SALDIVIA_ROOT}/config')
env = loader.generate_env(profile='${PROFILE}')
for k, v in sorted(env.items()):
    print(f'{k}={v}')
" >> "$ENV_FILE" || err "Failed to generate env from config"

# Validate configuration before proceeding
log "Validating configuration..."
python3 -c "
from saldivia.config import ConfigLoader, validate_config
loader = ConfigLoader('${SALDIVIA_ROOT}/config')
config = loader.load('${PROFILE}')
errors = validate_config(config)
if errors:
    print('Config validation errors:')
    for e in errors: print(f'  - {e}')
    exit(1)
print('  Config OK')
" || err "Config validation failed for profile: ${PROFILE}"

# Override: local secrets (API keys, passwords) — last wins.
if [ -f "${SALDIVIA_ROOT}/.env.local" ]; then
    echo "" >> "$ENV_FILE"
    cat "${SALDIVIA_ROOT}/.env.local" >> "$ENV_FILE"
fi

# Runtime paths — must be absolute (${PWD}-based paths break inside Docker containers).
# These override the placeholder PROMPT_CONFIG_FILE from .env.saldivia.
echo "" >> "$ENV_FILE"
echo "SALDIVIA_ROOT=${SALDIVIA_ROOT}" >> "$ENV_FILE"
echo "PROMPT_CONFIG_FILE=${SALDIVIA_ROOT}/config/prompt.yaml" >> "$ENV_FILE"

# --- Step 2: Build compose files ---
COMPOSE_FILES="-f ${COMPOSE_DIR}/docker-compose-rag-server.yaml"
COMPOSE_FILES="$COMPOSE_FILES -f ${SALDIVIA_ROOT}/config/compose-platform-services.yaml"

# Add optional services based on profile config
if python3 -c "from saldivia.config import ConfigLoader; c = ConfigLoader('${SALDIVIA_ROOT}/config').load('${PROFILE}'); exit(0 if c.get('services',{}).get('llm',{}).get('provider') == 'openrouter-proxy' else 1)" 2>/dev/null; then
    COMPOSE_FILES="$COMPOSE_FILES -f ${SALDIVIA_ROOT}/config/compose-openrouter-proxy.yaml"
    log "  OpenRouter proxy enabled"
fi

if python3 -c "from saldivia.config import ConfigLoader; c = ConfigLoader('${SALDIVIA_ROOT}/config').load('${PROFILE}'); exit(0 if c.get('guardrails',{}).get('enabled') else 1)" 2>/dev/null; then
    COMPOSE_FILES="$COMPOSE_FILES -f ${SALDIVIA_ROOT}/config/compose-guardrails-cloud.yaml"
    log "  Guardrails enabled"
fi

if python3 -c "from saldivia.config import ConfigLoader; c = ConfigLoader('${SALDIVIA_ROOT}/config').load('${PROFILE}'); exit(0 if c.get('observability',{}).get('enabled') else 1)" 2>/dev/null; then
    COMPOSE_FILES="$COMPOSE_FILES -f ${COMPOSE_DIR}/observability.yaml"
    log "  Observability enabled"
fi

log "Compose files: $COMPOSE_FILES"
# compose-overrides.yaml: overrides rag-frontend image to nginx:alpine (available locally)
# to prevent Docker from pulling nvcr.io/nvstaging/blueprint/rag-frontend:2.5.0 from NGC.
# rag-frontend runs harmlessly on port 8090; our SDA frontend is on port 3000.
COMPOSE_FILES="$COMPOSE_FILES -f ${SALDIVIA_ROOT}/config/compose-overrides.yaml"
SCALE_ARGS=""

# --- Step 2b: Ensure nvidia-rag network exists ---
# compose-platform-services.yaml and docker-compose-rag-server.yaml both declare
# the default network as name: nvidia-rag. Create it upfront to avoid "external network not found"
# errors when the two compose files are merged and Docker sees conflicting declarations.
log "Ensuring nvidia-rag network exists..."
docker network inspect nvidia-rag > /dev/null 2>&1 \
    || docker network create nvidia-rag \
    && log "  nvidia-rag network ready"

# --- Step 3: Start services ---
# Build platform service images first (mode-manager, ingestion-worker, auth-gateway).
# Done separately so Docker BuildKit does not try to resolve blueprint build contexts.
log "Building platform service images..."
# CACHE_BUST = git hash to invalidate Docker layer cache when code changes.
# Forces Docker to re-run COPY . . even if npm ci layer is cached.
CACHE_BUST=$(git -C "$SALDIVIA_ROOT" rev-parse --short HEAD 2>/dev/null || date +%s)
log "  Cache bust: ${CACHE_BUST}"
# --project-name must match the CWD-derived name used by 'docker compose up' below
# (which runs from $COMPOSE_DIR whose basename is 'compose')
SALDIVIA_ROOT="$SALDIVIA_ROOT" docker compose \
    --project-name compose \
    --env-file "$ENV_FILE" \
    -f "${SALDIVIA_ROOT}/config/compose-platform-services.yaml" \
    build --build-arg CACHE_BUST="${CACHE_BUST}" 2>&1 | tail -5

cd "$COMPOSE_DIR"
log "Starting services..."
docker compose --env-file "$ENV_FILE" $COMPOSE_FILES up -d --force-recreate --no-build $SCALE_ARGS 2>&1 | tail -10

# --- Step 3b: Ensure Milvus stack has restart policy ---
# Milvus is managed by the blueprint's vectordb.yaml (separate compose project).
# That file lacks restart policies, so we apply them after every deploy.
for container in milvus-standalone milvus-etcd milvus-minio; do
    if docker ps -q --filter "name=^${container}$" | grep -q .; then
        docker update --restart unless-stopped "$container" > /dev/null \
            && log "  restart policy: ${container}" \
            || warn "  failed to set restart on ${container}"
    fi
done

# --- Step 4: Flush Redis (clear orphaned NV-Ingest tasks) ---
log "Flushing Redis..."
REDIS_CONTAINER=$(docker ps --filter "name=redis" --format '{{.Names}}' | head -1)
if [ -n "$REDIS_CONTAINER" ]; then
    docker exec "$REDIS_CONTAINER" redis-cli FLUSHALL > /dev/null 2>&1 || warn "Redis flush failed"
else
    warn "Redis container not found"
fi

# --- Step 5: Apply NV-Ingest vlm.py runtime patch ---
# The NV-Ingest container hardcodes max_tokens=512 in vlm.py.
# This sed command changes it to 1024 inside the running container.
log "Applying NV-Ingest vlm.py max_tokens patch..."
sleep 5  # Wait for container to be ready
NVINGEST_CONTAINER=$(docker ps --filter "name=nv-ingest" --format '{{.Names}}' | head -1)
if [ -n "$NVINGEST_CONTAINER" ]; then
    docker exec "$NVINGEST_CONTAINER" \
        sed -i 's/max_tokens=512/max_tokens=1024/g' \
        /usr/lib/python3.10/dist-packages/nv_ingest/util/nim/vlm.py 2>/dev/null \
        && log "  vlm.py patched (max_tokens=1024)" \
        || warn "  vlm.py patch failed — captioning may use 512 tokens"
else
    warn "NV-Ingest container not found. vlm.py patch skipped."
fi

# --- Step 5b: Patch langchain_milvus for Milvus 2.6+ compatibility ---
# Milvus 2.6+ rejects output_fields containing "sparse" (BM25 field) even in DENSE search mode.
# Root cause: _remove_forbidden_fields() only strips BM25 fields when builtin_func is set,
# but in DENSE mode builtin_func=None → "sparse" stays in output_fields → Milvus rejects it.
# Fix: patch 4 search methods to exclude "sparse" from output_fields.
# NOTE: This patches the 'else' branch (enable_dynamic_field=False path), NOT the 'if' branch
# that was incorrectly patched before.
log "Patching langchain_milvus for Milvus 2.6+ compatibility..."
RAG_CONTAINER=$(docker ps --filter "name=rag-server" --filter "status=running" --format '{{.Names}}' | head -1)
if [ -n "$RAG_CONTAINER" ]; then
    # Write patch script locally and docker cp it in — avoids shell escaping hell.
    cat > /tmp/patch_langchain_milvus.py << 'PYEOF'
import shutil, os, sys

path = "/workspace/.venv/lib/python3.13/site-packages/langchain_milvus/vectorstores/milvus.py"

if not os.path.exists(path):
    print(f"File not found: {path}")
    sys.exit(1)

with open(path) as f:
    content = f.read()

old = "output_fields = self._remove_forbidden_fields(self.fields[:])"
new = "output_fields = [f for f in self._remove_forbidden_fields(self.fields[:]) if f != 'sparse']"

count = content.count(old)
if count == 0:
    print("Pattern not found — already patched or different langchain_milvus version")
    sys.exit(0)

content = content.replace(old, new)
with open(path, "w") as f:
    f.write(content)

# Clear __pycache__ to force Python to recompile
cache_dir = os.path.dirname(path) + "/__pycache__"
if os.path.exists(cache_dir):
    shutil.rmtree(cache_dir)

print(f"Patched {count} occurrence(s) in milvus.py. __pycache__ cleared.")
PYEOF

    docker cp /tmp/patch_langchain_milvus.py "${RAG_CONTAINER}:/tmp/patch_langchain_milvus.py" 2>/dev/null \
        && docker exec "$RAG_CONTAINER" python3 /tmp/patch_langchain_milvus.py \
        && log "  langchain_milvus patched (sparse excluded from output_fields)" \
        || warn "  langchain_milvus patch failed — RAG queries may fail on hybrid collections"

    # Restart rag-server so it reloads the patched Python code
    log "  Restarting rag-server to reload patched langchain_milvus..."
    docker restart "$RAG_CONTAINER" > /dev/null
    sleep 10
else
    warn "rag-server container not found. langchain_milvus patch skipped."
fi

# --- Step 6: Connect Nemotron network alias (brev-2gpu only) ---
# docker network connect --alias fails silently if the container is already on the network
# (even without the alias). Must disconnect first to force re-attach with alias.
if [ "$PROFILE" = "brev-2gpu" ]; then
    log "Connecting Nemotron-3-Super network alias..."
    NETWORK=$(docker network ls --filter "name=nvidia-rag" --format '{{.Name}}' | head -1)
    if [ -n "$NETWORK" ]; then
        docker network disconnect "$NETWORK" nemotron3-super 2>/dev/null || true
        sleep 1
        docker network connect --alias nim-llm "$NETWORK" nemotron3-super \
            && log "  nim-llm alias connected" \
            || warn "  Network alias failed — LLM may not be reachable as nim-llm"
    else
        warn "nvidia-rag network not found — is the Blueprint running?"
    fi
fi

# --- Step 7: Health checks ---
log "Waiting for services..."
MAX_WAIT=120
ELAPSED=0

check_service() {
    local name=$1 url=$2
    if curl -sf "$url" > /dev/null 2>&1; then
        log "  OK: ${name}"
        return 0
    fi
    return 1
}

while [ $ELAPSED -lt $MAX_WAIT ]; do
    RAG_OK=false
    INGEST_OK=false

    check_service "RAG Server" "http://localhost:8081/health" && RAG_OK=true
    check_service "Ingestor" "http://localhost:8082/health" && INGEST_OK=true

    if $RAG_OK; then
        log "Core services healthy (RAG Server up)."
        if $INGEST_OK; then
            log "  Ingestor also up."
        else
            warn "  Ingestor (8082) not up — ingestion unavailable. Start separately if needed."
        fi
        break
    fi

    sleep 10
    ELAPSED=$((ELAPSED + 10))
    log "  Waiting... (${ELAPSED}s/${MAX_WAIT}s)"
done

if [ $ELAPSED -ge $MAX_WAIT ]; then
    err "RAG Server did not become healthy within ${MAX_WAIT}s. Check: docker ps && docker logs rag-server"
fi

log "Deploy complete. Profile: ${PROFILE}"
log "  RAG Server:   http://localhost:8081"
log "  Ingestor:     http://localhost:8082"
log "  Frontend:     http://localhost:3000"
log "  Auth Gateway: http://localhost:9000  (API entry point with RBAC)"

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
BLUEPRINT_DIR="${SALDIVIA_ROOT}/blueprint"
PROFILE="${1:-brev-2gpu}"
COMPOSE_DIR="${BLUEPRINT_DIR}/deploy/compose"

log() { echo "[$(date +%H:%M:%S)] $*"; }
err() { echo "[$(date +%H:%M:%S)] ERROR: $*" >&2; exit 1; }
warn() { echo "[$(date +%H:%M:%S)] WARN: $*" >&2; }

[ -d "$BLUEPRINT_DIR" ] || err "Blueprint not found. Run: make setup"
[ -f "${SALDIVIA_ROOT}/config/profiles/${PROFILE}.yaml" ] || err "Unknown profile: ${PROFILE}"

# --- Step 1: Generate env from YAML config ---
export PYTHONPATH="${PYTHONPATH:-}:${SALDIVIA_ROOT}"

GPU_COUNT=$(nvidia-smi -L 2>/dev/null | wc -l || echo "0")
log "Detected ${GPU_COUNT} GPU(s)"

ENV_FILE="${COMPOSE_DIR}/.env.merged"
log "Generating environment from config (profile: ${PROFILE})..."
python3 -c "
from saldivia.config import ConfigLoader
loader = ConfigLoader('${SALDIVIA_ROOT}/config')
loader.write_env_file('${ENV_FILE}', profile='${PROFILE}')
print('Generated ${ENV_FILE}')
"

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

# Local secrets override everything
if [ -f "${SALDIVIA_ROOT}/.env.local" ]; then
    echo "" >> "$ENV_FILE"
    cat "${SALDIVIA_ROOT}/.env.local" >> "$ENV_FILE"
fi

# Set SALDIVIA_ROOT for compose-overrides.yaml volume mounts
echo "SALDIVIA_ROOT=${SALDIVIA_ROOT}" >> "$ENV_FILE"

# --- Step 2: Build compose files ---
COMPOSE_FILES="-f ${COMPOSE_DIR}/docker-compose-rag-server.yaml"
COMPOSE_FILES="$COMPOSE_FILES -f ${SALDIVIA_ROOT}/config/compose-overrides.yaml"
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
SCALE_ARGS=""

if [ "$PROFILE" = "workstation-1gpu" ]; then
    SCALE_ARGS="--scale nemotron3-super=0"
    log "Workstation profile: nemotron3-super disabled"
fi

# --- Step 3: Start services ---
cd "$COMPOSE_DIR"
log "Starting services..."
docker compose --env-file "$ENV_FILE" $COMPOSE_FILES up -d --force-recreate $SCALE_ARGS 2>&1 | tail -10

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

# --- Step 6: Connect Nemotron network alias (brev-2gpu only) ---
if [ "$PROFILE" = "brev-2gpu" ]; then
    log "Connecting Nemotron-3-Super network alias..."
    NETWORK=$(docker network ls --filter "name=nvidia-rag" --format '{{.Name}}' | head -1)
    if [ -n "$NETWORK" ]; then
        docker network connect --alias nim-llm "$NETWORK" nemotron3-super 2>/dev/null \
            && log "  nim-llm alias connected" \
            || warn "  Network alias already exists or container not found"
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

    if $RAG_OK && $INGEST_OK; then
        log "All services healthy."
        break
    fi

    sleep 10
    ELAPSED=$((ELAPSED + 10))
    log "  Waiting... (${ELAPSED}s/${MAX_WAIT}s)"
done

if [ $ELAPSED -ge $MAX_WAIT ]; then
    warn "Some services may not be ready yet. Check: docker ps"
fi

log "Deploy complete. Profile: ${PROFILE}"
log "  RAG Server: http://localhost:8081"
log "  Ingestor:   http://localhost:8082"
log "  Frontend:   http://localhost:8090"

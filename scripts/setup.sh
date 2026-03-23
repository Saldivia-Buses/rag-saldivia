#!/usr/bin/env bash
set -euo pipefail

# RAG Saldivia — Setup Script
# Inicializa el submodule del blueprint, aplica patches, y construye imágenes Docker.
#
# Uso:
#   ./scripts/setup.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SALDIVIA_ROOT="$(dirname "$SCRIPT_DIR")"
BLUEPRINT_DIR="${SALDIVIA_ROOT}/vendor/rag-blueprint"

log() { echo "[$(date +%H:%M:%S)] $*"; }
err() { echo "[$(date +%H:%M:%S)] ERROR: $*" >&2; exit 1; }
warn() { echo "[$(date +%H:%M:%S)] WARN: $*" >&2; }

# --- Step 1: Inicializar submodule del blueprint ---
if [ -f "${BLUEPRINT_DIR}/README.md" ]; then
    log "Blueprint submodule ya inicializado (vendor/rag-blueprint)."
else
    log "Inicializando submodule vendor/rag-blueprint..."
    cd "$SALDIVIA_ROOT"
    GIT_TERMINAL_PROMPT=0 git submodule update --init --recursive \
        || err "Falló la inicialización del submodule. Verificar conexión a GitHub."
    log "Submodule inicializado."
fi

cd "$BLUEPRINT_DIR"

# --- Step 2: Validate frontend patches (dry-run) ---
PATCH_DIR="${SALDIVIA_ROOT}/patches/frontend/patches"
if [ -d "$PATCH_DIR" ]; then
    log "Validating frontend patches (blueprint React UI — optional)..."
    PATCHES_OK=true
    for patch in "$PATCH_DIR"/*.patch; do
        [ -f "$patch" ] || continue
        if ! git apply --check "$patch" 2>/dev/null; then
            warn "Patch no aplica: $(basename "$patch") (blueprint version mismatch — skipping)"
            PATCHES_OK=false
        else
            log "  OK: $(basename "$patch")"
        fi
    done

    # --- Step 3: Apply frontend patches ---
    if [ "$PATCHES_OK" = true ]; then
        log "Applying frontend patches..."
        for patch in "$PATCH_DIR"/*.patch; do
            [ -f "$patch" ] || continue
            git apply "$patch"
            log "  Applied: $(basename "$patch")"
        done
    else
        warn "Skipping patch application (SDA SvelteKit frontend no las necesita)"
    fi
else
    log "No frontend patches found. Skipping."
fi

# --- Step 4: Copy new frontend files ---
NEW_FILES_DIR="${SALDIVIA_ROOT}/patches/frontend/new"
if [ -d "$NEW_FILES_DIR" ]; then
    log "Copying new frontend files..."

    # SaldiviaSection.tsx -> frontend/src/components/settings/
    if [ -f "$NEW_FILES_DIR/SaldiviaSection.tsx" ]; then
        cp "$NEW_FILES_DIR/SaldiviaSection.tsx" \
           "$BLUEPRINT_DIR/frontend/src/components/settings/SaldiviaSection.tsx"
        log "  Copied: SaldiviaSection.tsx"
    fi

    # Crossdoc hooks -> frontend/src/hooks/
    for hook in useCrossdocStream.ts useCrossdocDecompose.ts; do
        if [ -f "$NEW_FILES_DIR/$hook" ]; then
            cp "$NEW_FILES_DIR/$hook" "$BLUEPRINT_DIR/frontend/src/hooks/$hook"
            log "  Copied: $hook"
        fi
    done
fi

# --- Step 5: Preserve blueprint's default .env ---
cd "$BLUEPRINT_DIR/deploy/compose"
if [ -f ".env" ] && [ ! -f ".env.blueprint" ]; then
    cp .env .env.blueprint
    log "Saved blueprint defaults to .env.blueprint"
fi

# --- Step 6: Build Docker images ---
log "Building Docker images (ingestor-server, rag-server, rag-frontend)..."

# Merge env: blueprint defaults < saldivia overrides < local secrets
cp .env.blueprint .env.build 2>/dev/null || : > .env.build
echo "" >> .env.build
cat "${SALDIVIA_ROOT}/config/.env.saldivia" >> .env.build
if [ -f "${SALDIVIA_ROOT}/.env.local" ]; then
    echo "" >> .env.build
    cat "${SALDIVIA_ROOT}/.env.local" >> .env.build
else
    warn "No .env.local found. NGC_API_KEY may be missing for image pulls."
fi

docker compose --env-file .env.build -f docker-compose-rag-server.yaml build \
    rag-server rag-frontend 2>&1 | tail -5
docker compose --env-file .env.build -f docker-compose-ingestor-server.yaml build \
    2>&1 | tail -5

rm -f .env.build
log "Setup complete. Run: make deploy PROFILE=workstation-1gpu"

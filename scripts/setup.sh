#!/usr/bin/env bash
set -euo pipefail

# RAG Saldivia — Setup Script
# Clones the NVIDIA RAG Blueprint, applies patches, copies config.
#
# Usage:
#   ./scripts/setup.sh [BLUEPRINT_VERSION]
#   ./scripts/setup.sh 2.5.0

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SALDIVIA_ROOT="$(dirname "$SCRIPT_DIR")"
BLUEPRINT_DIR="${SALDIVIA_ROOT}/blueprint"
BLUEPRINT_VERSION="${1:-2.5.0}"
BLUEPRINT_REPO="https://github.com/NVIDIA/rag-blueprint.git"
BLUEPRINT_BRANCH="release-v${BLUEPRINT_VERSION}"

log() { echo "[$(date +%H:%M:%S)] $*"; }
err() { echo "[$(date +%H:%M:%S)] ERROR: $*" >&2; exit 1; }
warn() { echo "[$(date +%H:%M:%S)] WARN: $*" >&2; }

# --- Step 1: Clone blueprint ---
if [ -d "$BLUEPRINT_DIR" ]; then
    log "Blueprint directory exists. Checking version..."
    cd "$BLUEPRINT_DIR"
    CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    if [ "$CURRENT_BRANCH" != "$BLUEPRINT_BRANCH" ]; then
        err "Blueprint is on branch '$CURRENT_BRANCH', expected '$BLUEPRINT_BRANCH'. Delete blueprint/ and re-run."
    fi
    log "Blueprint already cloned at correct version."
else
    log "Cloning NVIDIA RAG Blueprint ${BLUEPRINT_VERSION}..."
    git clone --branch "$BLUEPRINT_BRANCH" --depth 1 "$BLUEPRINT_REPO" "$BLUEPRINT_DIR" \
        || err "Failed to clone blueprint. Check version and network."
fi

cd "$BLUEPRINT_DIR"

# --- Step 2: Validate frontend patches (dry-run) ---
PATCH_DIR="${SALDIVIA_ROOT}/patches/frontend/patches"
if [ -d "$PATCH_DIR" ]; then
    log "Validating frontend patches..."
    for patch in "$PATCH_DIR"/*.patch; do
        [ -f "$patch" ] || continue
        if ! git apply --check "$patch" 2>/dev/null; then
            err "Patch failed dry-run: $(basename "$patch"). Blueprint version mismatch?"
        fi
        log "  OK: $(basename "$patch")"
    done

    # --- Step 3: Apply frontend patches ---
    log "Applying frontend patches..."
    for patch in "$PATCH_DIR"/*.patch; do
        [ -f "$patch" ] || continue
        git apply "$patch"
        log "  Applied: $(basename "$patch")"
    done
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
    ingestor-server rag-server rag-frontend 2>&1 | tail -5

rm -f .env.build
log "Setup complete. Run: make deploy PROFILE=brev-2gpu"

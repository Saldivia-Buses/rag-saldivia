#!/usr/bin/env bash
# Smoke test for Phase 1 infra: MinIO + SGLang instances.
# Usage: ./deploy/scripts/smoke-test-infra.sh [--gpu]
#
# Without --gpu: only tests MinIO
# With --gpu: also tests SGLang instances (requires running GPU containers)

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

pass() { echo -e "${GREEN}✓ $1${NC}"; }
fail() { echo -e "${RED}✗ $1${NC}"; exit 1; }
skip() { echo -e "${YELLOW}⊘ $1${NC}"; }

MINIO_ENDPOINT="${STORAGE_ENDPOINT:-http://localhost:9000}"
SGLANG_OCR_URL="${SGLANG_OCR_URL:-http://localhost:8100}"
SGLANG_VISION_URL="${SGLANG_VISION_URL:-http://localhost:8101}"

echo "=== SDA Infrastructure Smoke Test ==="
echo ""

# ── MinIO ──────────────────────────────────────────────────────────────

echo "--- MinIO (${MINIO_ENDPOINT}) ---"

if curl -sf "${MINIO_ENDPOINT}/minio/health/live" > /dev/null 2>&1; then
    pass "MinIO is alive"
else
    fail "MinIO not reachable at ${MINIO_ENDPOINT}"
fi

# Test put/get via Go test
if command -v go &> /dev/null; then
    cd "$(dirname "$0")/../.." # repo root
    if (cd services/app && go test ./internal/rag/ingest/storage/...) -count=1 -run TestPutGetDelete > /dev/null 2>&1; then
        pass "Storage put/get/delete works"
    else
        fail "Storage test failed"
    fi
else
    skip "Go not available, skipping storage test"
fi

# ── SGLang (only with --gpu flag) ──────────────────────────────────────

if [[ "${1:-}" == "--gpu" ]]; then
    echo ""
    echo "--- SGLang OCR (${SGLANG_OCR_URL}) ---"

    if curl -sf "${SGLANG_OCR_URL}/health" > /dev/null 2>&1; then
        pass "sglang-ocr is healthy"
    else
        fail "sglang-ocr not reachable"
    fi

    echo "--- SGLang Vision (${SGLANG_VISION_URL}) ---"

    if curl -sf "${SGLANG_VISION_URL}/health" > /dev/null 2>&1; then
        pass "sglang-vision is healthy"
    else
        fail "sglang-vision not reachable"
    fi
else
    echo ""
    skip "SGLang tests skipped (run with --gpu to test model servers)"
fi

echo ""
echo -e "${GREEN}=== All checks passed ===${NC}"

#!/bin/bash
# record-deploy.sh — Record a deploy event in the platform DB via API.
#
# Calls POST /v1/platform/deploys with a short-lived service JWT.
#
# Usage: record-deploy.sh --version <ref> --sha <sha> --status <status>
# Exit code 0 = recorded. Exit code 1 = failed (non-fatal, deploy still succeeded).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

VERSION=""
SHA=""
STATUS=""
PLATFORM_URL="http://localhost:8006"

# ── Parse args ──────────────────────────────────────────────────────────

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      [[ "$2" =~ ^v?[0-9][a-zA-Z0-9._-]*$ ]] || { echo "ERROR: invalid --version format"; exit 1; }
      VERSION="$2"; shift 2 ;;
    --sha)
      [[ "$2" =~ ^[a-f0-9]{7,40}$ ]] || { echo "ERROR: invalid --sha format"; exit 1; }
      SHA="$2"; shift 2 ;;
    --status)
      [[ "$2" =~ ^(success|failed|rollback)$ ]] || { echo "ERROR: --status must be success, failed, or rollback"; exit 1; }
      STATUS="$2"; shift 2 ;;
    *) echo "ERROR: unknown argument: $1"; exit 1 ;;
  esac
done

if [ -z "$VERSION" ] || [ -z "$SHA" ] || [ -z "$STATUS" ]; then
  echo "ERROR: --version, --sha, and --status are required"
  exit 1
fi

# ── Get service token ──────────────────────────────────────────────────

TOKEN=""
if [ -x "$SCRIPT_DIR/get-service-token.sh" ]; then
  TOKEN=$(bash "$SCRIPT_DIR/get-service-token.sh" deploy 2>/dev/null || echo "")
fi

if [ -z "$TOKEN" ]; then
  echo "WARN: could not obtain service token — skipping deploy recording" >&2
  echo "Deploy was successful but not recorded in platform DB."
  exit 0  # non-fatal: the deploy itself succeeded
fi

# ── Record deploy ──────────────────────────────────────────────────────

# Build JSON payload safely (jq --arg escapes special characters)
PAYLOAD=$(jq -n \
  --arg svc "sda" \
  --arg ver_to "$SHA" \
  --arg status "$STATUS" \
  --arg notes "ref=${VERSION} sha=${SHA}" \
  '{"service":$svc,"version_from":"","version_to":$ver_to,"status":$status,"notes":$notes}')

RESPONSE=$(curl -sf --max-time 10 \
  -X POST "${PLATFORM_URL}/v1/platform/deploys" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD" 2>&1) || {
  echo "WARN: failed to record deploy: $RESPONSE" >&2
  echo "Deploy was successful but not recorded in platform DB."
  exit 0  # non-fatal
}

echo "Deploy recorded: ${VERSION} (${SHA}) status=${STATUS}"

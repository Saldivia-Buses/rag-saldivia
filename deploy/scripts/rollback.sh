#!/bin/bash
# rollback.sh — Restore previous service versions after failed deploy.
#
# SECURITY: Does NOT use source to load the versions file (DS2).
# Parses KEY=value lines with strict regex validation.
#
# Usage: rollback.sh <versions-file>
# Exit code 0 = rollback successful. Exit code 1 = invalid input.

set -euo pipefail

VERSIONS_FILE="${1:-}"

if [ -z "$VERSIONS_FILE" ]; then
  echo "ERROR: usage: rollback.sh <versions-file>"
  exit 1
fi

if [ ! -f "$VERSIONS_FILE" ]; then
  echo "ERROR: versions file not found: $VERSIONS_FILE"
  exit 1
fi

# ── Safe parser: only accept KEY=value lines matching expected format ──

declare -A VERSIONS
PARSED=0

while IFS='=' read -r key value || [ -n "$key" ]; do
  # Skip empty lines and comments
  [[ -z "$key" || "$key" =~ ^[[:space:]]*# ]] && continue

  # Strip whitespace
  key=$(echo "$key" | tr -d '[:space:]')
  value=$(echo "$value" | tr -d '[:space:]')

  # Validate key format: UPPERCASE_VERSION
  if [[ ! "$key" =~ ^[A-Z][A-Z0-9_]*_VERSION$ ]]; then
    echo "WARN: skipping invalid key: $key" >&2
    continue
  fi

  # Validate value format: semver (X.Y.Z) or git SHA (7-40 hex chars)
  if [[ ! "$value" =~ ^([0-9]+\.[0-9]+\.[0-9]+|[a-f0-9]{7,40})$ ]]; then
    echo "ERROR: invalid version for $key: $value"
    exit 1
  fi

  VERSIONS[$key]="$value"
  ((PARSED++))
done < "$VERSIONS_FILE"

if [ "$PARSED" -eq 0 ]; then
  echo "ERROR: no valid version entries found in $VERSIONS_FILE"
  exit 1
fi

echo "Parsed $PARSED version(s) from rollback file"

# ── Export validated versions and re-deploy ────────────────────────────

for key in "${!VERSIONS[@]}"; do
  export "$key=${VERSIONS[$key]}"
done

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEPLOY_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Rolling back with versions:"
for key in $(echo "${!VERSIONS[@]}" | tr ' ' '\n' | sort); do
  echo "  $key=${VERSIONS[$key]}"
done

docker compose -f "$DEPLOY_DIR/docker-compose.prod.yml" up -d --pull always

echo ""
echo "Rollback complete. Versions restored from $VERSIONS_FILE"

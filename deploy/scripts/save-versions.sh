#!/bin/bash
# save-versions.sh — Capture current service versions for rollback.
#
# Queries /v1/info on each running Go service and outputs KEY=value lines
# with validated version format (semver or SHA). Output is safe for use
# as docker compose --env-file input.
#
# Usage: save-versions.sh > rollback.env
# Exit code 0 = at least one version captured. Exit code 1 = total failure.

set -euo pipefail

# Service name → port mapping (must match docker-compose)
declare -A SERVICES=(
  [erp]=8013
  [app]=8020
)

HOST="localhost"
CAPTURED=0

for svc in $(echo "${!SERVICES[@]}" | tr ' ' '\n' | sort); do
  port="${SERVICES[$svc]}"

  # Query /v1/info and extract version + git_sha
  raw=$(curl -sf --max-time 5 "http://${HOST}:${port}/v1/info" 2>/dev/null || echo "")

  if [ -z "$raw" ]; then
    echo "WARN: ${svc} (:${port}) unreachable" >&2
    continue
  fi

  ver=$(echo "$raw" | jq -re '.git_sha // .version' 2>/dev/null || echo "")

  if [ -z "$ver" ]; then
    echo "WARN: ${svc} (:${port}) returned unparseable JSON: ${raw:0:80}" >&2
    continue
  fi

  # Normalize key: lowercase service name → uppercase with underscores
  KEY=$(echo "${svc}" | tr '[:lower:]-' '[:upper:]_')_VERSION

  # Validate version format: semver (X.Y.Z) or git SHA (7-40 hex chars)
  if [[ "$ver" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || [[ "$ver" =~ ^[a-f0-9]{7,40}$ ]]; then
    echo "${KEY}=${ver}"
    ((CAPTURED++))
  else
    echo "WARN: ${svc} returned invalid version format: ${ver}" >&2
  fi
done

if [ "$CAPTURED" -eq 0 ]; then
  echo "ERROR: no valid versions captured — cannot create rollback file" >&2
  exit 1
fi

echo "Captured ${CAPTURED} service version(s)" >&2

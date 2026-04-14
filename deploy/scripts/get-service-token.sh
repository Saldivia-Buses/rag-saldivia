#!/bin/bash
# get-service-token.sh — Obtain a short-lived JWT for machine-to-machine calls.
#
# Attempts to get a service token from the auth service using service account
# credentials stored in Docker secrets or environment variable.
#
# Usage: get-service-token.sh <service-name>
# Output: JWT token string on stdout.
# Exit code 0 = token obtained. Exit code 1 = failed.

set -euo pipefail

SERVICE="${1:-}"

if [ -z "$SERVICE" ]; then
  echo "ERROR: usage: get-service-token.sh <service-name>" >&2
  exit 1
fi

# Validate service name format
if [[ ! "$SERVICE" =~ ^[a-z][a-z0-9-]*$ ]]; then
  echo "ERROR: invalid service name format: $SERVICE" >&2
  exit 1
fi

AUTH_URL="http://localhost:8001"

# Read secret from Docker secret file first, fall back to env var
SECRET=""
if [ -f /run/secrets/service_account_key ]; then
  SECRET=$(cat /run/secrets/service_account_key)
elif [ -n "${SERVICE_ACCOUNT_KEY:-}" ]; then
  SECRET="$SERVICE_ACCOUNT_KEY"
else
  echo "ERROR: no service account key found (checked /run/secrets/service_account_key and \$SERVICE_ACCOUNT_KEY)" >&2
  exit 1
fi

if [ -z "$SECRET" ]; then
  echo "ERROR: service account key is empty" >&2
  exit 1
fi

# Build JSON payload safely (jq --arg escapes quotes, backslashes, control chars)
PAYLOAD=$(jq -n --arg svc "$SERVICE" --arg key "$SECRET" \
  '{"service":$svc,"key":$key}')

# Request short-lived token from auth service
TOKEN=$(curl -sf --max-time 10 \
  -X POST "${AUTH_URL}/v1/auth/service-token" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD" 2>/dev/null \
  | jq -re '.token' 2>/dev/null) || {
  echo "ERROR: failed to obtain service token from ${AUTH_URL}" >&2
  exit 1
}

if [ -z "$TOKEN" ]; then
  echo "ERROR: auth service returned empty token" >&2
  exit 1
fi

echo "$TOKEN"

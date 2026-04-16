#!/bin/bash
# health-check.sh — Parametrized health check for SDA services.
#
# Checks /health endpoint on every Go service and reports results.
# Retries with backoff until --timeout is reached.
#
# Usage: health-check.sh --env dev|prod [--timeout 120]
# Exit code 0 = all healthy. Exit code 1 = at least one unhealthy.

set -euo pipefail

# ── Defaults ────────────────────────────────────────────────────────────

ENV=""
TIMEOUT=120
INTERVAL=5

# Service name → port mapping (must match docker-compose)
declare -A SERVICES=(
  [auth]=8001
  [ws]=8002
  [chat]=8003
  [agent]=8004
  [notification]=8005
  [platform]=8006
  [ingest]=8007
  [feedback]=8008
  [traces]=8009
  [search]=8010
  [astro]=8011
  [bigbrother]=8012
  [erp]=8013
  [healthwatch]=8014
)

# ── Parse args ──────────────────────────────────────────────────────────

while [[ $# -gt 0 ]]; do
  case "$1" in
    --env)
      [[ "$2" =~ ^(dev|prod)$ ]] || { echo "ERROR: --env must be dev or prod"; exit 1; }
      ENV="$2"; shift 2 ;;
    --timeout)
      [[ "$2" =~ ^[0-9]+$ ]] || { echo "ERROR: --timeout must be a number"; exit 1; }
      TIMEOUT="$2"; shift 2 ;;
    *) echo "ERROR: unknown argument: $1"; exit 1 ;;
  esac
done

if [ -z "$ENV" ]; then
  echo "ERROR: --env is required (dev or prod)"
  exit 1
fi

# ── Health check loop ──────────────────────────────────────────────────

echo "═══ SDA Health Check (env=$ENV, timeout=${TIMEOUT}s) ═══"
echo ""

HOST="localhost"
DEADLINE=$((SECONDS + TIMEOUT))
PASS=0
FAIL=0
UNHEALTHY=()

# Give services time to start on first check
sleep 3

while [ $SECONDS -lt $DEADLINE ]; do
  PASS=0
  FAIL=0
  UNHEALTHY=()

  for svc in "${!SERVICES[@]}"; do
    port="${SERVICES[$svc]}"
    if curl -sf --max-time 5 "http://${HOST}:${port}/health" > /dev/null 2>&1; then
      ((PASS++))
    else
      ((FAIL++))
      UNHEALTHY+=("$svc:$port")
    fi
  done

  if [ "$FAIL" -eq 0 ]; then
    echo "  All ${PASS} services healthy after $((SECONDS))s"
    echo ""
    echo "RESULT: OK"
    exit 0
  fi

  echo "  Waiting... ${PASS} healthy, ${FAIL} unhealthy: ${UNHEALTHY[*]}"
  sleep "$INTERVAL"
done

# Timeout reached — report final state
echo ""
echo "═══ Final Status (TIMEOUT after ${TIMEOUT}s) ═══"
for svc in $(echo "${!SERVICES[@]}" | tr ' ' '\n' | sort); do
  port="${SERVICES[$svc]}"
  if curl -sf --max-time 5 "http://${HOST}:${port}/health" > /dev/null 2>&1; then
    echo "  ✓ $svc (:$port)"
  else
    echo "  ✗ $svc (:$port)"
  fi
done
echo ""
echo "RESULT: FAIL — ${FAIL} service(s) did not become healthy within ${TIMEOUT}s"
exit 1

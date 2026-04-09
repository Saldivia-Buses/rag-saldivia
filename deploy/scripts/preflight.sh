#!/bin/bash
# preflight.sh — Pre-deploy validation for SDA Framework.
#
# Runs 13 checks to verify the system is ready for deployment.
# Exit code 0 = all pass. Exit code 1 = at least one failure.
#
# Usage: deploy/scripts/preflight.sh
# Or:    make deploy-preflight

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEPLOY_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

PASS=0
WARN=0
FAIL=0

check() {
    local name="$1"
    local result="$2" # "ok", "warn", or "fail"
    local msg="${3:-}"

    case "$result" in
        ok)   echo "  ✓ $name"; ((PASS++)) ;;
        warn) echo "  ⚠ $name: $msg"; ((WARN++)) ;;
        fail) echo "  ✗ $name: $msg"; ((FAIL++)) ;;
    esac
}

echo "═══ SDA Framework — Preflight Checks ═══"
echo ""

# 1. .env exists
if [ -f "$DEPLOY_DIR/.env" ]; then
    check ".env file exists" "ok"
    source "$DEPLOY_DIR/.env"
else
    check ".env file exists" "fail" "run: cp deploy/.env.example deploy/.env"
    # Can't continue without .env
    echo ""
    echo "RESULT: FAIL (1 critical missing)"
    exit 1
fi

# 2. SDA_DOMAIN set and not localhost in production
if [ "${SDA_ENV:-development}" = "production" ] && [ "${SDA_DOMAIN:-localhost}" = "localhost" ]; then
    check "SDA_DOMAIN configured for production" "fail" "SDA_DOMAIN cannot be 'localhost' in production"
elif [ -n "${SDA_DOMAIN:-}" ]; then
    check "SDA_DOMAIN configured (${SDA_DOMAIN})" "ok"
else
    check "SDA_DOMAIN configured" "warn" "using default 'localhost'"
fi

# 3. SDA_ACME_EMAIL not default in production
if [ "${SDA_ENV:-development}" = "production" ] && [ "${SDA_ACME_EMAIL:-admin@example.com}" = "admin@example.com" ]; then
    check "SDA_ACME_EMAIL configured for production" "fail" "set a real email for Let's Encrypt"
elif [ -n "${SDA_ACME_EMAIL:-}" ]; then
    check "SDA_ACME_EMAIL configured" "ok"
else
    check "SDA_ACME_EMAIL configured" "warn" "using default (OK for dev)"
fi

# 4. Docker daemon running
if command -v docker &>/dev/null && docker info &>/dev/null 2>&1; then
    check "Docker daemon running" "ok"
else
    check "Docker daemon running" "fail" "start Docker or install it"
fi

# 5. Docker Compose v2
if command -v docker &>/dev/null && docker compose version &>/dev/null 2>&1; then
    local_compose_ver=$(docker compose version --short 2>/dev/null || echo "unknown")
    check "Docker Compose v2 ($local_compose_ver)" "ok"
else
    check "Docker Compose v2 available" "fail" "install docker-compose-plugin"
fi

# 6. GPU available (if configured)
if [ -n "${SDA_GPU_DEVICES:-}" ] && [ "$SDA_GPU_DEVICES" != "" ]; then
    if command -v nvidia-smi &>/dev/null && nvidia-smi &>/dev/null 2>&1; then
        gpu_name=$(nvidia-smi --query-gpu=name --format=csv,noheader 2>/dev/null | head -1)
        check "GPU available ($gpu_name)" "ok"
    else
        check "GPU available" "warn" "SDA_GPU_DEVICES=$SDA_GPU_DEVICES but nvidia-smi not found"
    fi
else
    check "GPU configured" "warn" "no GPU configured (SDA_GPU_DEVICES empty)"
fi

# 7. Ports 80/443 free
for port in 80 443; do
    if command -v ss &>/dev/null; then
        if ss -tlnp 2>/dev/null | grep -q ":$port "; then
            check "Port $port free" "warn" "port $port already in use"
        else
            check "Port $port free" "ok"
        fi
    else
        check "Port $port free" "warn" "cannot check (ss not available)"
    fi
done

# 8. Secrets exist
secrets_ok=true
for secret in jwt-private.pem jwt-public.pem; do
    if [ ! -f "$DEPLOY_DIR/secrets/dynamic/$secret" ]; then
        check "Secret $secret exists" "fail" "run: deploy/scripts/gen-jwt-keys.sh"
        secrets_ok=false
    fi
done
if [ "$secrets_ok" = true ]; then
    check "JWT secrets exist" "ok"
fi

# 9. Cloudflare credentials (production only)
if [ "${SDA_ENV:-development}" = "production" ]; then
    if [ -f "$DEPLOY_DIR/secrets/cloudflared-credentials.json" ]; then
        check "Cloudflare tunnel credentials" "ok"
    else
        check "Cloudflare tunnel credentials" "fail" "missing deploy/secrets/cloudflared-credentials.json"
    fi
fi

# 10. Generated configs exist (not stale .tmpl)
configs_ok=true
for config in traefik/traefik.prod.yml traefik/dynamic/prod.yml cloudflare/config.yml; do
    if [ ! -f "$DEPLOY_DIR/$config" ]; then
        check "Generated config $config" "fail" "run: make deploy-gen"
        configs_ok=false
    fi
done
if [ "$configs_ok" = true ]; then
    check "Generated configs up to date" "ok"
fi

# 11. Disk space > 10GB free
if command -v df &>/dev/null; then
    free_gb=$(df -BG --output=avail / 2>/dev/null | tail -1 | tr -d ' G' || echo "0")
    if [ "$free_gb" -gt 10 ] 2>/dev/null; then
        check "Disk space (${free_gb}GB free)" "ok"
    else
        check "Disk space" "warn" "only ${free_gb}GB free (recommend > 10GB)"
    fi
fi

# 12. RAM > 4GB available
if command -v free &>/dev/null; then
    free_mb=$(free -m 2>/dev/null | awk '/^Mem:/{print $7}' || echo "0")
    if [ "$free_mb" -gt 4096 ] 2>/dev/null; then
        check "RAM available (${free_mb}MB free)" "ok"
    elif [ "$free_mb" -gt 2048 ] 2>/dev/null; then
        check "RAM available" "warn" "only ${free_mb}MB free (recommend > 4GB)"
    else
        check "RAM available" "fail" "only ${free_mb}MB free (need > 2GB)"
    fi
fi

# 13. Docker images pullable (spot check)
if command -v docker &>/dev/null && docker info &>/dev/null 2>&1; then
    if docker image ls --format '{{.Repository}}' 2>/dev/null | grep -q "sda-auth\|postgres\|redis" ; then
        check "Docker images available (local)" "ok"
    else
        check "Docker images available" "warn" "no SDA images found locally — first deploy will build"
    fi
fi

# Summary
echo ""
echo "═══════════════════════════════════════"
echo "  PASS: $PASS  WARN: $WARN  FAIL: $FAIL"
echo "═══════════════════════════════════════"

if [ "$FAIL" -gt 0 ]; then
    echo ""
    echo "RESULT: FAIL — fix the issues above before deploying."
    exit 1
else
    echo ""
    echo "RESULT: OK — ready to deploy."
    exit 0
fi

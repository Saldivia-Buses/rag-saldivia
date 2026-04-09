#!/bin/bash
# gen-config.sh — Generate Traefik and Cloudflare configs from templates.
#
# Reads deploy/.env, computes derived variables, and runs envsubst
# on .yml.tmpl templates to produce final .yml configs.
#
# Usage: deploy/scripts/gen-config.sh
# Or:    make deploy-gen

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEPLOY_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_DIR="$(cd "$DEPLOY_DIR/.." && pwd)"

# Load .env
ENV_FILE="$DEPLOY_DIR/.env"
if [ ! -f "$ENV_FILE" ]; then
    echo "ERROR: $ENV_FILE not found. Run: cp deploy/.env.example deploy/.env"
    exit 1
fi
# Source .env but don't override already-set env vars
while IFS='=' read -r key val; do
    # Skip comments and empty lines
    [[ "$key" =~ ^#.*$ || -z "$key" ]] && continue
    # Strip quotes
    val="${val%\"}" && val="${val#\"}"
    val="${val%\'}" && val="${val#\'}"
    # Only set if not already defined
    if [ -z "${!key:-}" ]; then
        export "$key=$val"
    fi
done < "$ENV_FILE"

# Auto-detect host IP
source "$SCRIPT_DIR/detect-host.sh"

# Compute derived variables
export SDA_DOMAIN="${SDA_DOMAIN:-localhost}"
export SDA_DOMAIN_ESCAPED
SDA_DOMAIN_ESCAPED=$(echo "$SDA_DOMAIN" | sed 's/\./\\./g')
export SDA_SMTP_FROM="${SMTP_FROM:-noreply@${SDA_DOMAIN}}"
export SDA_WS_ORIGINS="https://*.${SDA_DOMAIN}"
export SDA_CORS_REGEX="^https://[a-z0-9-]+\\.${SDA_DOMAIN_ESCAPED}\$"
export SDA_TENANT_SLUG="${SDA_TENANT_SLUG:-dev}"
export SDA_ACME_EMAIL="${SDA_ACME_EMAIL:-admin@example.com}"
export SDA_GPU_DEVICES="${SDA_GPU_DEVICES:-0}"
export SDA_ENV="${SDA_ENV:-development}"

# Explicit variable list — only substitute SDA_* variables.
# Prevents replacing Go template syntax ({{ }}) in Prometheus alerts.
VARS='$SDA_DOMAIN $SDA_DOMAIN_ESCAPED $SDA_ACME_EMAIL $SDA_TENANT_SLUG
$SDA_GPU_DEVICES $SDA_ENV $SDA_SMTP_FROM $SDA_WS_ORIGINS $SDA_CORS_REGEX $SDA_HOST_IP'

# Generate configs from templates
generate() {
    local tmpl="$1"
    local out="${tmpl%.tmpl}"
    if [ ! -f "$tmpl" ]; then
        echo "  SKIP: $tmpl (not found)"
        return
    fi
    envsubst "$VARS" < "$tmpl" > "$out"
    echo "  OK: $out"
}

echo "Generating configs for domain: $SDA_DOMAIN (env: $SDA_ENV)"
echo ""

generate "$DEPLOY_DIR/traefik/traefik.prod.yml.tmpl"
generate "$DEPLOY_DIR/traefik/dynamic/prod.yml.tmpl"
generate "$DEPLOY_DIR/cloudflare/config.yml.tmpl"

echo ""
echo "Done. Host IP: $SDA_HOST_IP"

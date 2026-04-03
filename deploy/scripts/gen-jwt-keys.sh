#!/usr/bin/env bash
# Generate Ed25519 keypair for JWT signing/verification.
# Auth service uses the private key (signing). All other services use the public key (verification).
# Output: base64-encoded PEM strings suitable for env vars (JWT_PRIVATE_KEY, JWT_PUBLIC_KEY).
set -euo pipefail

OUTDIR="${1:-$(dirname "$0")/../secrets/dynamic}"
mkdir -p "$OUTDIR"

PRIV="$OUTDIR/jwt-private.pem"
PUB="$OUTDIR/jwt-public.pem"

if [ -f "$PRIV" ] && [ -f "$PUB" ]; then
    echo "Keys already exist at $OUTDIR — skipping generation."
    echo "Delete them first if you want to regenerate."
else
    openssl genpkey -algorithm Ed25519 -out "$PRIV" 2>/dev/null
    openssl pkey -in "$PRIV" -pubout -out "$PUB" 2>/dev/null
    echo "Generated Ed25519 keypair:"
    echo "  Private: $PRIV"
    echo "  Public:  $PUB"
fi

echo ""
echo "Base64 values for .env / docker-compose:"
echo "JWT_PRIVATE_KEY=$(base64 -w0 "$PRIV")"
echo "JWT_PUBLIC_KEY=$(base64 -w0 "$PUB")"

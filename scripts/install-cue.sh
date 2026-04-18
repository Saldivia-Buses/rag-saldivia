#!/usr/bin/env bash
# install-cue.sh — install CUE binary with SHA256 verification.
#
# CUE is the schema language for services/app/internal/events/spec (Plan 26). The Go eventsgen
# tool embeds cuelang.org/go as a library, but a local `cue` binary is useful
# for editor tooling, `cue vet`, and interactive debugging.
#
# Usage: scripts/install-cue.sh [install_dir]
#   default install_dir: $HOME/.local/bin
#
# Called by CI (with actions/cache) and by developers manually.
set -euo pipefail

CUE_VERSION="${CUE_VERSION:-v0.14.2}"
INSTALL_DIR="${1:-$HOME/.local/bin}"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "unsupported arch: $ARCH"; exit 1 ;;
esac

# Checksums (obtain from https://github.com/cue-lang/cue/releases — keep in sync
# with CUE_VERSION). A mismatch fails the install.
declare -A SHA256=(
    [v0.14.2-linux-amd64]="de6bcd5a601ca53dde09de8e5f884a4bad23c91306f04ca6a6ce11dab48a8307"
)

KEY="${CUE_VERSION}-${OS}-${ARCH}"
EXPECTED_SHA="${SHA256[$KEY]:-}"
if [[ -z "$EXPECTED_SHA" ]]; then
    echo "no checksum for $KEY — update scripts/install-cue.sh"
    exit 1
fi

if command -v cue >/dev/null 2>&1 && cue version 2>&1 | grep -q "$CUE_VERSION"; then
    echo "cue $CUE_VERSION already installed at $(command -v cue)"
    exit 0
fi

URL="https://github.com/cue-lang/cue/releases/download/${CUE_VERSION}/cue_${CUE_VERSION}_${OS}_${ARCH}.tar.gz"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

echo "downloading $URL"
curl -sSL -o "$TMPDIR/cue.tgz" "$URL"

ACTUAL_SHA="$(sha256sum "$TMPDIR/cue.tgz" | awk '{print $1}')"
if [[ "$ACTUAL_SHA" != "$EXPECTED_SHA" ]]; then
    echo "checksum mismatch: expected $EXPECTED_SHA, got $ACTUAL_SHA"
    exit 1
fi
echo "checksum OK"

tar -xzf "$TMPDIR/cue.tgz" -C "$TMPDIR" cue
mkdir -p "$INSTALL_DIR"
install -m 755 "$TMPDIR/cue" "$INSTALL_DIR/cue"

echo "installed: $("$INSTALL_DIR/cue" version | head -1) → $INSTALL_DIR/cue"
echo "add to PATH if needed: export PATH=\"$INSTALL_DIR:\$PATH\""

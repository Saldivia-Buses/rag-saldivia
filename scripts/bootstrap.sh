#!/usr/bin/env bash
set -euo pipefail

# RAG Saldivia — Bootstrap Script
# Verifica e instala todas las dependencias necesarias, luego corre setup.sh
#
# Uso:
#   ./scripts/bootstrap.sh [BLUEPRINT_VERSION]
#
# Requiere: git, curl, apt (Ubuntu 22.04+)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SALDIVIA_ROOT="$(dirname "$SCRIPT_DIR")"
BLUEPRINT_VERSION="${1:-2.5.0}"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log()  { echo -e "${GREEN}[bootstrap]${NC} $*"; }
warn() { echo -e "${YELLOW}[bootstrap] WARN:${NC} $*"; }
err()  { echo -e "${RED}[bootstrap] ERROR:${NC} $*" >&2; exit 1; }
ok()   { echo -e "${GREEN}[bootstrap]${NC} ✅ $*"; }
miss() { echo -e "${YELLOW}[bootstrap]${NC} ⬇️  Instalando: $*"; }

[ "$(id -u)" = "0" ] || err "Correr como root (o con sudo)"

# --- Chequear Ubuntu ---
. /etc/os-release 2>/dev/null || true
if [[ "${ID:-}" != "ubuntu" ]]; then
    warn "Este script está optimizado para Ubuntu. Continuando de todas formas..."
fi

log "=== RAG Saldivia Bootstrap ==="
log "Verificando dependencias..."

# -------------------------------------------------------
# 1. Git
# -------------------------------------------------------
if command -v git &>/dev/null; then
    ok "git $(git --version | awk '{print $3}')"
else
    miss "git"
    apt-get update -qq && apt-get install -y -qq git
    ok "git instalado"
fi

# -------------------------------------------------------
# 2. curl (necesario para installs siguientes)
# -------------------------------------------------------
if command -v curl &>/dev/null; then
    ok "curl"
else
    miss "curl"
    apt-get update -qq && apt-get install -y -qq curl ca-certificates gnupg
fi

# -------------------------------------------------------
# 3. Docker CLI
# -------------------------------------------------------
if command -v docker &>/dev/null && docker info &>/dev/null 2>&1; then
    ok "docker $(docker --version | awk '{print $3}' | tr -d ',')"
else
    # En RunPod, el daemon corre en el host y se expone via socket
    # Solo necesitamos instalar el CLI + compose plugin
    if [ -S /var/run/docker.sock ]; then
        log "Socket Docker del host detectado (/var/run/docker.sock)"
        if ! command -v docker &>/dev/null; then
            miss "Docker CLI"
            apt-get update -qq
            apt-get install -y -qq ca-certificates curl gnupg
            install -m 0755 -d /etc/apt/keyrings
            curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
                | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
            chmod a+r /etc/apt/keyrings/docker.gpg
            echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" \
                > /etc/apt/sources.list.d/docker.list
            apt-get update -qq
            apt-get install -y -qq docker-ce-cli docker-buildx-plugin docker-compose-plugin
        fi
        ok "docker $(docker --version | awk '{print $3}' | tr -d ',') (via host socket)"
    else
        err "No hay socket Docker ni daemon disponible. Este entorno no soporta Docker."
    fi
fi

# -------------------------------------------------------
# 4. GPU accesible en Docker
# -------------------------------------------------------
if docker run --rm --gpus all ubuntu:22.04 nvidia-smi &>/dev/null 2>&1; then
    ok "GPU accesible en Docker"
else
    warn "GPU no accesible directamente en Docker (puede funcionar igual si el host tiene NVIDIA runtime)"
fi

# -------------------------------------------------------
# 6. Node.js 20+
# -------------------------------------------------------
NODE_OK=false
if command -v node &>/dev/null; then
    NODE_VER=$(node --version | tr -d 'v' | cut -d. -f1)
    if [ "$NODE_VER" -ge 20 ]; then
        ok "node $(node --version)"
        NODE_OK=true
    else
        warn "Node.js $(node --version) es muy viejo (necesita 20+). Actualizando..."
    fi
fi

if [ "$NODE_OK" = false ]; then
    miss "Node.js 20"
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
    apt-get install -y -qq nodejs
    ok "node $(node --version)"
fi

# -------------------------------------------------------
# 7. pnpm
# -------------------------------------------------------
if command -v pnpm &>/dev/null; then
    ok "pnpm $(pnpm --version)"
else
    miss "pnpm"
    npm install -g pnpm --quiet
    ok "pnpm $(pnpm --version)"
fi

# -------------------------------------------------------
# 8. uv (Python package manager)
# -------------------------------------------------------
if command -v uv &>/dev/null; then
    ok "uv $(uv --version)"
else
    miss "uv"
    curl -LsSf https://astral.sh/uv/install.sh | sh
    export PATH="$HOME/.cargo/bin:$PATH"
    ok "uv instalado"
fi

# -------------------------------------------------------
# 9. Verificar GPU visible
# -------------------------------------------------------
if nvidia-smi &>/dev/null 2>&1; then
    GPU=$(nvidia-smi --query-gpu=name,memory.total --format=csv,noheader | head -1)
    ok "GPU: $GPU"
else
    err "nvidia-smi falló — ¿los drivers están instalados?"
fi

echo ""
log "=== Todas las dependencias OK. Corriendo setup.sh... ==="
echo ""

exec "$SCRIPT_DIR/setup.sh" "$BLUEPRINT_VERSION"

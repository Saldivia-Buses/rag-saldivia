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
# 3. Docker
# -------------------------------------------------------
if command -v docker &>/dev/null; then
    ok "docker $(docker --version | awk '{print $3}' | tr -d ',')"
else
    miss "Docker CE"
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
    apt-get install -y -qq \
        docker-ce docker-ce-cli containerd.io \
        docker-buildx-plugin docker-compose-plugin
    ok "Docker instalado"
fi

# -------------------------------------------------------
# 4. Docker daemon corriendo
# -------------------------------------------------------
if docker info &>/dev/null 2>&1; then
    ok "Docker daemon corriendo"
else
    log "Iniciando Docker daemon..."
    # En RunPod no hay systemd — usar dockerd directo en background
    if command -v systemctl &>/dev/null && systemctl is-active docker &>/dev/null; then
        ok "Docker ya activo via systemd"
    else
        dockerd &>/var/log/dockerd.log &
        DOCKERD_PID=$!
        log "Docker daemon iniciado (PID $DOCKERD_PID), esperando..."
        for i in $(seq 1 30); do
            sleep 1
            docker info &>/dev/null 2>&1 && break
            [ $i -eq 30 ] && err "Docker daemon no arrancó en 30s. Ver /var/log/dockerd.log"
        done
        ok "Docker daemon listo"
    fi
fi

# -------------------------------------------------------
# 5. NVIDIA Container Toolkit
# -------------------------------------------------------
if docker run --rm --gpus all nvidia/cuda:12.0.0-base-ubuntu22.04 nvidia-smi &>/dev/null 2>&1; then
    ok "NVIDIA Container Toolkit (GPU accesible en Docker)"
else
    miss "NVIDIA Container Toolkit"
    curl -fsSL https://nvidia.github.io/libnvidia-container/gpgkey \
        | gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg
    curl -s -L https://nvidia.github.io/libnvidia-container/stable/deb/nvidia-container-toolkit.list \
        | sed 's#deb https://#deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] https://#g' \
        > /etc/apt/sources.list.d/nvidia-container-toolkit.list
    apt-get update -qq
    apt-get install -y -qq nvidia-container-toolkit
    nvidia-ctk runtime configure --runtime=docker
    # Reiniciar dockerd para que tome la config de nvidia
    pkill dockerd 2>/dev/null || true
    sleep 2
    dockerd &>/var/log/dockerd.log &
    sleep 5
    ok "NVIDIA Container Toolkit instalado"
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

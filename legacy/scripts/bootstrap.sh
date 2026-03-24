#!/usr/bin/env bash
set -euo pipefail

# RAG Saldivia — Bootstrap Script
# Instala todas las dependencias del sistema desde cero en Ubuntu 24.04.
# Idempotente: puede correrse múltiples veces sin romper nada.
#
# Uso:
#   ./scripts/bootstrap.sh
#   NVIDIA_DRIVER_VERSION=560 ./scripts/bootstrap.sh
#
# Requiere: curl, Ubuntu 24.04 (bare-metal o VM con GPU física)

# ---------------------------------------------------------------------------
# Configuración
# ---------------------------------------------------------------------------
NVIDIA_DRIVER_VERSION="${NVIDIA_DRIVER_VERSION:-570}"

# ---------------------------------------------------------------------------
# Helpers de logging
# ---------------------------------------------------------------------------
log() { echo "[$(date +%H:%M:%S)] $*"; }
err() { echo "[$(date +%H:%M:%S)] ERROR: $*" >&2; exit 1; }
ok()  { echo "[$(date +%H:%M:%S)] OK: $*"; }
skip(){ echo "[$(date +%H:%M:%S)] SKIP: $* (ya instalado)"; }

# ---------------------------------------------------------------------------
# Verificar Ubuntu 24.04
# ---------------------------------------------------------------------------
log "=== RAG Saldivia Bootstrap — Ubuntu 24.04 ==="

if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    if [[ "${ID:-}" != "ubuntu" ]]; then
        err "Este script requiere Ubuntu. Sistema detectado: ${ID:-desconocido}"
    fi
    if [[ "${VERSION_ID:-}" != "24.04" ]]; then
        log "ADVERTENCIA: Script optimizado para Ubuntu 24.04. Versión detectada: ${VERSION_ID:-desconocida}"
    fi
else
    err "No se puede detectar el sistema operativo (/etc/os-release no existe)"
fi

# ---------------------------------------------------------------------------
# Verificar que NO corre como root (usa sudo internamente)
# ---------------------------------------------------------------------------
if [[ "$(id -u)" = "0" ]]; then
    err "No correr como root. El script usa sudo internamente donde necesario."
fi

# Verificar que sudo está disponible
command -v sudo &>/dev/null || err "sudo no encontrado. Instalarlo primero."

log "Usuario: $(whoami)"
log "NVIDIA_DRIVER_VERSION: ${NVIDIA_DRIVER_VERSION}"

# ---------------------------------------------------------------------------
# 1. Actualizar índice de paquetes y dependencias base
# ---------------------------------------------------------------------------
log "--- Actualizando índice de paquetes ---"
sudo apt-get update -qq
sudo apt-get install -y -qq \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    software-properties-common \
    apt-transport-https
ok "Dependencias base instaladas"

# ---------------------------------------------------------------------------
# 2. NVIDIA Driver
# ---------------------------------------------------------------------------
log "--- Verificando NVIDIA Driver ${NVIDIA_DRIVER_VERSION} ---"

if dpkg -l 2>/dev/null | grep -q "nvidia-driver-${NVIDIA_DRIVER_VERSION}"; then
    skip "nvidia-driver-${NVIDIA_DRIVER_VERSION}"
else
    log "Instalando nvidia-driver-${NVIDIA_DRIVER_VERSION}..."
    # Agregar el repositorio de drivers NVIDIA (ubuntu-drivers-common)
    sudo add-apt-repository -y ppa:graphics-drivers/ppa 2>/dev/null || true
    sudo apt-get update -qq
    sudo apt-get install -y -qq "nvidia-driver-${NVIDIA_DRIVER_VERSION}"
    ok "nvidia-driver-${NVIDIA_DRIVER_VERSION} instalado (requiere reboot para activarse)"
fi

# ---------------------------------------------------------------------------
# 3. NVIDIA CUDA Toolkit
# ---------------------------------------------------------------------------
log "--- Verificando NVIDIA CUDA Toolkit ---"

if dpkg -l 2>/dev/null | grep -q "nvidia-cuda-toolkit"; then
    skip "nvidia-cuda-toolkit"
else
    log "Instalando nvidia-cuda-toolkit..."
    sudo apt-get install -y -qq nvidia-cuda-toolkit
    ok "nvidia-cuda-toolkit instalado"
fi

# ---------------------------------------------------------------------------
# 4. NVIDIA Container Toolkit (GPU access en Docker)
# ---------------------------------------------------------------------------
log "--- Verificando NVIDIA Container Toolkit ---"

if dpkg -l 2>/dev/null | grep -q "nvidia-container-toolkit"; then
    skip "nvidia-container-toolkit"
else
    log "Configurando repositorio NVIDIA Container Toolkit..."
    # Repositorio oficial de NVIDIA
    curl -fsSL https://nvidia.github.io/libnvidia-container/gpgkey \
        | sudo gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg
    curl -s -L https://nvidia.github.io/libnvidia-container/stable/deb/nvidia-container-toolkit.list \
        | sed 's#deb https://#deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] https://#g' \
        | sudo tee /etc/apt/sources.list.d/nvidia-container-toolkit.list > /dev/null
    sudo apt-get update -qq
    log "Instalando nvidia-container-toolkit..."
    sudo apt-get install -y -qq nvidia-container-toolkit
    ok "nvidia-container-toolkit instalado"
fi

# ---------------------------------------------------------------------------
# 5. Docker Engine completo
# ---------------------------------------------------------------------------
log "--- Verificando Docker Engine ---"

if command -v docker &>/dev/null; then
    DOCKER_VER=$(docker --version | awk '{print $3}' | tr -d ',')
    skip "docker ${DOCKER_VER}"
else
    log "Instalando Docker Engine desde repositorio oficial..."
    # Agregar GPG key oficial de Docker
    sudo install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
        | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    sudo chmod a+r /etc/apt/keyrings/docker.gpg

    # Agregar repositorio
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "${VERSION_CODENAME}") stable" \
        | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

    sudo apt-get update -qq
    sudo apt-get install -y -qq \
        docker-ce \
        docker-ce-cli \
        containerd.io \
        docker-buildx-plugin \
        docker-compose-plugin

    # Iniciar y habilitar el daemon
    sudo systemctl enable docker
    sudo systemctl start docker

    ok "docker $(docker --version | awk '{print $3}' | tr -d ',') instalado"
fi

# Configurar Docker runtime para NVIDIA (si el toolkit está presente)
if command -v nvidia-ctk &>/dev/null; then
    log "Configurando NVIDIA runtime en Docker..."
    sudo nvidia-ctk runtime configure --runtime=docker
    sudo systemctl restart docker
    ok "NVIDIA runtime configurado en Docker"
fi

# ---------------------------------------------------------------------------
# 6. Configurar Docker para usuario no-root
# ---------------------------------------------------------------------------
log "--- Configurando Docker para usuario no-root ---"

CURRENT_USER="$(whoami)"
if groups "${CURRENT_USER}" | grep -q "\bdocker\b"; then
    skip "usuario ${CURRENT_USER} ya está en el grupo docker"
else
    log "Agregando ${CURRENT_USER} al grupo docker..."
    sudo usermod -aG docker "${CURRENT_USER}"
    ok "Usuario ${CURRENT_USER} agregado al grupo docker (efectivo al reabrir sesión)"
fi

# ---------------------------------------------------------------------------
# 7. Node.js 20 LTS
# ---------------------------------------------------------------------------
log "--- Verificando Node.js 20 LTS ---"

NODE_OK=false
if command -v node &>/dev/null; then
    NODE_VER=$(node --version 2>/dev/null | tr -d 'v' | cut -d. -f1)
    if [[ "${NODE_VER}" -ge 20 ]]; then
        skip "node $(node --version)"
        NODE_OK=true
    else
        log "Node.js $(node --version) es muy viejo (necesita 20+). Actualizando..."
    fi
fi

if [[ "${NODE_OK}" = false ]]; then
    log "Instalando Node.js 20 LTS via NodeSource..."
    curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
    sudo apt-get install -y -qq nodejs
    ok "node $(node --version) instalado"
fi

# ---------------------------------------------------------------------------
# 8. pnpm
# ---------------------------------------------------------------------------
log "--- Verificando pnpm ---"

if command -v pnpm &>/dev/null; then
    skip "pnpm $(pnpm --version)"
else
    log "Instalando pnpm via npm..."
    sudo npm install -g pnpm --quiet
    ok "pnpm $(pnpm --version) instalado"
fi

# ---------------------------------------------------------------------------
# 9. uv (Python package manager)
# ---------------------------------------------------------------------------
log "--- Verificando uv ---"

if command -v uv &>/dev/null; then
    skip "uv $(uv --version)"
else
    log "Instalando uv via astral.sh..."
    curl -LsSf https://astral.sh/uv/install.sh | sh
    # uv instala en ~/.local/bin
    export PATH="${HOME}/.local/bin:${PATH}"
    if command -v uv &>/dev/null; then
        ok "uv $(uv --version) instalado"
    else
        ok "uv instalado en ~/.local/bin (agregar al PATH para usar en esta sesión)"
    fi
fi

# ---------------------------------------------------------------------------
# Resumen final
# ---------------------------------------------------------------------------
echo ""
log "=== Bootstrap completo ==="
log "Cerrar sesión y volver a abrir para aplicar cambios de grupo Docker."
log ""
log "Próximo paso: ./scripts/setup.sh"

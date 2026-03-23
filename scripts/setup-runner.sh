#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# setup-runner.sh — Instala y configura un GitHub Actions self-hosted runner
#
# Uso:
#   GITHUB_REPO=Camionerou/rag-saldivia \
#   RUNNER_TOKEN=<token> \
#   RUNNER_LABELS=self-hosted,workstation-1 \
#   ./scripts/setup-runner.sh
#
# El token se obtiene en:
#   https://github.com/<owner>/<repo>/settings/actions/runners/new
# =============================================================================

# ---------------------------------------------------------------------------
# Variables configurables
# ---------------------------------------------------------------------------
GITHUB_REPO="${GITHUB_REPO:-}"                         # ej: Camionerou/rag-saldivia
RUNNER_TOKEN="${RUNNER_TOKEN:-}"                        # token de GitHub Actions
RUNNER_NAME="${RUNNER_NAME:-$(hostname)}"               # default: hostname de la máquina
RUNNER_LABELS="${RUNNER_LABELS:-self-hosted,workstation-1}"
RUNNER_VERSION="2.321.0"                               # pin a versión específica
RUNNER_DIR="${RUNNER_DIR:-${HOME}/actions-runner}"

RUNNER_ARCH="x64"
RUNNER_OS="linux"
RUNNER_TARBALL="actions-runner-${RUNNER_OS}-${RUNNER_ARCH}-${RUNNER_VERSION}.tar.gz"
RUNNER_URL="https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/${RUNNER_TARBALL}"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
log()  { echo "[setup-runner] $*"; }
die()  { echo "[setup-runner] ERROR: $*" >&2; exit 1; }

# ---------------------------------------------------------------------------
# 1. Verificar variables requeridas
# ---------------------------------------------------------------------------
log "Verificando variables de entorno..."

[[ -z "${GITHUB_REPO}" ]] && die "GITHUB_REPO no está definido. Ej: Camionerou/rag-saldivia"
[[ -z "${RUNNER_TOKEN}" ]] && die "RUNNER_TOKEN no está definido. Obtené uno en: https://github.com/${GITHUB_REPO}/settings/actions/runners/new"

log "Repo:   ${GITHUB_REPO}"
log "Runner: ${RUNNER_NAME}"
log "Labels: ${RUNNER_LABELS}"
log "Dir:    ${RUNNER_DIR}"

# ---------------------------------------------------------------------------
# 2. Verificar dependencias del sistema
# ---------------------------------------------------------------------------
log "Verificando dependencias..."
for cmd in curl tar sha256sum sudo; do
    command -v "${cmd}" >/dev/null 2>&1 || die "Comando requerido no encontrado: ${cmd}"
done

# ---------------------------------------------------------------------------
# 3. Crear directorio de instalación
# ---------------------------------------------------------------------------
log "Creando directorio ${RUNNER_DIR}..."
mkdir -p "${RUNNER_DIR}"
cd "${RUNNER_DIR}"

# ---------------------------------------------------------------------------
# 4. Descargar el runner
# ---------------------------------------------------------------------------
if [[ -f "${RUNNER_TARBALL}" ]]; then
    log "Tarball ya existe, omitiendo descarga: ${RUNNER_TARBALL}"
else
    log "Descargando GitHub Actions runner v${RUNNER_VERSION}..."
    curl -fsSL -o "${RUNNER_TARBALL}" "${RUNNER_URL}"
    log "Descarga completada."
fi

# ---------------------------------------------------------------------------
# 5. Verificar hash SHA256
# ---------------------------------------------------------------------------
log "Verificando integridad del tarball..."
HASH_URL="https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/actions-runner-${RUNNER_OS}-${RUNNER_ARCH}-${RUNNER_VERSION}.tar.gz.sha256"
EXPECTED_HASH=$(curl -fsSL "${HASH_URL}" | awk '{print $1}')
ACTUAL_HASH=$(sha256sum "${RUNNER_TARBALL}" | awk '{print $1}')

if [[ "${EXPECTED_HASH}" != "${ACTUAL_HASH}" ]]; then
    rm -f "${RUNNER_TARBALL}"
    die "Verificación SHA256 fallida. Tarball eliminado por seguridad."
fi
log "SHA256 verificado correctamente."

# ---------------------------------------------------------------------------
# 6. Descomprimir
# ---------------------------------------------------------------------------
log "Descomprimiendo runner..."
tar -xzf "${RUNNER_TARBALL}"
log "Descompresión completada."

# ---------------------------------------------------------------------------
# 7. Configurar el runner
# ---------------------------------------------------------------------------
log "Configurando runner..."
./config.sh \
    --url "https://github.com/${GITHUB_REPO}" \
    --token "${RUNNER_TOKEN}" \
    --name "${RUNNER_NAME}" \
    --labels "${RUNNER_LABELS}" \
    --unattended

log "Runner configurado."

# ---------------------------------------------------------------------------
# 8. Instalar y arrancar como servicio systemd
# ---------------------------------------------------------------------------
log "Instalando servicio systemd..."
sudo ./svc.sh install

log "Iniciando servicio..."
sudo ./svc.sh start

# ---------------------------------------------------------------------------
# 9. Verificar estado
# ---------------------------------------------------------------------------
log "Verificando estado del servicio..."
sudo ./svc.sh status

log ""
log "============================================================"
log "Runner instalado y corriendo correctamente."
log "  Nombre:  ${RUNNER_NAME}"
log "  Labels:  ${RUNNER_LABELS}"
log "  Repo:    https://github.com/${GITHUB_REPO}"
log "  Dir:     ${RUNNER_DIR}"
log ""
log "Para ver los logs del servicio:"
log "  sudo journalctl -u actions.runner.*.service -f"
log ""
log "Para desinstalar:"
log "  cd ${RUNNER_DIR} && sudo ./svc.sh stop && sudo ./svc.sh uninstall && ./config.sh remove --token <token>"
log "============================================================"

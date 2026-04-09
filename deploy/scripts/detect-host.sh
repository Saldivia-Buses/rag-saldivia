#!/bin/bash
# detect-host.sh — Auto-detect host IP for Docker networking.
# WSL2, native Linux, and Docker Desktop have different requirements.
#
# Output: exports SDA_HOST_IP with the correct value.
# Usage: source deploy/scripts/detect-host.sh

set -euo pipefail

detect_host_ip() {
    # WSL2: extract host IP from /etc/resolv.conf
    if grep -qi microsoft /proc/version 2>/dev/null; then
        local wsl_ip
        wsl_ip=$(grep -m1 nameserver /etc/resolv.conf | awk '{print $2}')
        if [ -n "$wsl_ip" ]; then
            echo "$wsl_ip"
            return
        fi
    fi

    # Docker Desktop (macOS/Windows): host-gateway works natively
    if docker info 2>/dev/null | grep -q "Operating System: Docker Desktop"; then
        echo "host-gateway"
        return
    fi

    # Native Linux: use docker0 bridge IP
    if command -v ip &>/dev/null; then
        local docker0_ip
        docker0_ip=$(ip addr show docker0 2>/dev/null | grep 'inet ' | awk '{print $2}' | cut -d/ -f1)
        if [ -n "$docker0_ip" ]; then
            echo "$docker0_ip"
            return
        fi
    fi

    # Fallback
    echo "host-gateway"
}

# Only set if not already defined
if [ -z "${SDA_HOST_IP:-}" ]; then
    export SDA_HOST_IP
    SDA_HOST_IP=$(detect_host_ip)
fi

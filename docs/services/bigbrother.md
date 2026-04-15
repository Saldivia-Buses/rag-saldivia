---
title: Service: bigbrother
audience: ai
last_reviewed: 2026-04-15
related:
  - ../packages/plc.md
  - ../packages/remote.md
  - ../packages/crypto.md
  - ../packages/approval.md
  - ../packages/audit.md
---

## Purpose

Network and OT (operational technology) intelligence: ARP-scans the LAN to
discover devices, fingerprints them, talks Modbus/OPC-UA to PLCs, runs
SSH/WinRM commands on hosts, and stores credentials in an envelope-encrypted
vault gated by two-person approval. Read this before changing scan logic, PLC
register access, the credential vault, or the approval workflow.

**Container exception:** the Dockerfile is **Alpine**, not distroless. ARP
scanning needs `NET_RAW` + `NET_ADMIN` capabilities and `libpcap`, which
distroless cannot provide. Documented as the only allowed deviation in the
invariant checks.

## Endpoints

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/health` | none | Liveness + postgres/nats/scanner/redis check |
| GET | `/v1/bigbrother/devices` | JWT + `bigbrother.read` | List discovered devices |
| GET | `/v1/bigbrother/devices/{id}` | JWT + `bigbrother.read` | Device detail |
| GET | `/v1/bigbrother/topology` | JWT + `bigbrother.read` | Layered topology view |
| GET | `/v1/bigbrother/events` | JWT + `bigbrother.read` | Recent scan + control events |
| GET | `/v1/bigbrother/stats` | JWT + `bigbrother.read` | Aggregated counts |
| GET | `/v1/bigbrother/devices/{id}/registers` | JWT + `bigbrother.plc.read` | Read PLC registers |
| POST | `/v1/bigbrother/devices/{id}/registers/{addr}` | JWT + `bigbrother.plc.write` | Write a PLC register (may require approval) |
| POST | `/v1/bigbrother/devices/{id}/exec` | JWT + `bigbrother.exec` | SSH/WinRM exec from allowlist |
| POST | `/v1/bigbrother/devices/{id}/registers/{addr}/approve` | JWT + `bigbrother.admin` | Two-person approve a pending PLC write |
| GET | `/v1/bigbrother/credentials` | JWT + `bigbrother.admin` | List credential metadata |
| POST | `/v1/bigbrother/credentials` | JWT + `bigbrother.admin` | Store encrypted credential |
| DELETE | `/v1/bigbrother/credentials/{id}` | JWT + `bigbrother.admin` | Revoke credential |
| POST | `/v1/bigbrother/scan` | JWT + `bigbrother.admin` | Trigger an immediate ARP sweep |
| POST | `/v1/bigbrother/scan/mode` | JWT + `bigbrother.admin` | Switch scan mode (passive/active) |

Routes assembled in `services/bigbrother/internal/handler/devices.go:63`.
All `FailOpen=false` — security beats availability for this service.

## NATS events

| Subject | Direction | Trigger |
|---|---|---|
| `tenant.{slug}.bigbrother.device.discovered` | pub | New device on scan (`services/bigbrother/internal/service/scanner.go:61`) |
| `tenant.{slug}.bigbrother.scan.completed` | pub | Scan loop iteration finished |
| `tenant.{slug}.bigbrother.plc.value_changed` | pub | PLC register written or polled change |
| `tenant.{slug}.bigbrother.plc.approval_requested` | pub | PLC write awaiting second approver (`services/bigbrother/internal/service/plc.go:186`) |

Bigbrother does not subscribe.

## Env vars

| Name | Required | Default | Purpose |
|---|---|---|---|
| `BIGBROTHER_PORT` | no | `8012` | HTTP listener port |
| `POSTGRES_TENANT_URL` | yes | — | Devices, events, credentials, pending writes |
| `JWT_PUBLIC_KEY` | yes | — | Ed25519 public key |
| `NATS_URL` | no | `nats://localhost:4222` | Event publishing |
| `REDIS_URL` | no | `localhost:6379` | Token blacklist + write rate limit |
| `TENANT_SLUG` | no | `dev` | NATS subject prefix |
| `SCAN_MODE` | no | `passive` | `passive` (ARP listen) or `active` (probe) |
| `LAN_INTERFACE` | no | — | NIC for ARP scanner; empty falls back to stub |
| `BB_KEK_PATH` | no | `/run/secrets/bb_kek` | Key-encryption key for credential vault |
| `EVENT_RETENTION_DAYS` | no | `90` | Cleanup cutoff |

## Dependencies

- **PostgreSQL tenant** — `bb_devices`, `bb_events`, `bb_credentials`,
  `bb_pending_writes`. Hourly cleanup goroutine
  (`services/bigbrother/cmd/main.go:158`).
- **Redis** — token blacklist + per-write rate limit on credential ops.
- **NATS** — fan-out for discovery and PLC events.
- **LAN interface** — ARP scanner on a real NIC; stub scanner otherwise
  (WSL2 dev).
- **Docker secret `bb_kek`** — KEK feeding `pkg/crypto.Encryptor` (envelope
  encryption with AAD) for the credential vault.
- **Audit writer** — `pkg/audit` records every credential and PLC write.

## Permissions used

- `bigbrother.read` — devices, topology, events, stats.
- `bigbrother.plc.read` — read PLC registers.
- `bigbrother.plc.write` — write PLC registers (subject to safety tier and
  approval).
- `bigbrother.exec` — remote command execution.
- `bigbrother.admin` — credentials, approvals, manual scans, scan mode.

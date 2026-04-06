# Docker Secrets for Production

This directory holds secret files consumed by Docker Compose via
`docker-compose.prod.yml`. Files are mounted as `/run/secrets/*`
inside containers.

## Required files

| File | Used by | Content |
|---|---|---|
| `jwt-private.pem` | auth only | Ed25519 private key (PEM) |
| `jwt-public.pem` | all Go services | Ed25519 public key (PEM) |
| `db-platform-url` | platform, feedback, traces | PostgreSQL connection URL |
| `db-tenant-template-url` | auth, chat, search, ingest, notification | PostgreSQL connection URL template |
| `redis-password` | redis | Password string |
| `encryption-master-key` | auth | 32-byte AES-256 key (base64) |

## NATS per-service credentials

NATS uses per-service auth (not a shared token). Each service connects
with its own username/password defined in `deploy/nats/nats-server.conf`.
Passwords are provided as environment variables to the NATS container:

| Variable | NATS user | Service |
|---|---|---|
| `AUTH_NATS_PASS` | auth | Auth Service |
| `CHAT_NATS_PASS` | chat | Chat Service |
| `WS_NATS_PASS` | ws | WebSocket Hub |
| `NOTIF_NATS_PASS` | notification | Notification Service |
| `INGEST_NATS_PASS` | ingest | Ingest Service |
| `FEEDBACK_NATS_PASS` | feedback | Feedback Service |
| `AGENT_NATS_PASS` | agent | Agent Service |
| `TRACES_NATS_PASS` | traces | Traces Service |
| `PLATFORM_NATS_PASS` | platform | Platform Service |
| `EXTRACTOR_NATS_PASS` | extractor | Extractor Service |

Provide these via a `.env` file or export them before running
`docker compose up`. Generate random passwords:

```bash
for svc in AUTH CHAT WS NOTIF INGEST FEEDBACK AGENT TRACES PLATFORM EXTRACTOR; do
  echo "${svc}_NATS_PASS=$(openssl rand -base64 24)"
done > .env.nats
```

## PostgreSQL SSL

Connection URLs in production MUST include `sslmode=require`:

```
postgres://sda:password@postgres:5432/sda_platform?sslmode=require
```

The tenant resolver (`pkg/tenant/resolver.go`) enforces `sslmode=require`
as a fallback if no sslmode is specified in the URL. Dev URLs with
`sslmode=disable` are respected — the fallback only applies when the
parameter is absent.

If PostgreSQL ever moves off the local host, upgrade to `sslmode=verify-full`
with CA certificate pinning.

## Generation

JWT keys: `deploy/scripts/gen-jwt-keys.sh`

Files in `dynamic/` are gitignored and generated at deploy time.

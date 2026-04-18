# Docker Secrets for Production

This directory holds secret files consumed by Docker Compose via
`docker-compose.prod.yml`. Files are mounted as `/run/secrets/*`
inside containers.

## Required files

| File | Used by | Content |
|---|---|---|
| `jwt-private.pem` | app (auth module signs) | Ed25519 private key (PEM) |
| `jwt-public.pem` | app + erp | Ed25519 public key (PEM) |
| `db-platform-url` | app (platform + feedback + traces + healthwatch modules) | PostgreSQL connection URL |
| `db-tenant-template-url` | app (auth + chat + search + ingest + notification modules) | PostgreSQL connection URL template |
| `redis-password` | redis | Password string |
| `encryption-master-key` | app (auth module) | 32-byte AES-256 key (base64) |

## NATS per-service credentials

NATS uses per-service auth (not a shared token). Each service connects
with its own username/password defined in `deploy/nats/nats-server.conf`.
Passwords are provided as environment variables to the NATS container.

Post ADR 025 the monolith collapses the 10 pre-fusion Go services into a
single `app` user; only three credentials remain:

| Variable | NATS user | Service |
|---|---|---|
| `APP_NATS_PASS` | app | The consolidated Go monolith (core + ops + rag + realtime) |
| `ERP_NATS_PASS` | erp | ERP (still a standalone service) |
| `EXTRACTOR_NATS_PASS` | extractor | Python OCR / vision pipeline |

Provide these via a `.env` file or export them before running
`docker compose up`. Generate random passwords:

```bash
for svc in APP ERP EXTRACTOR; do
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

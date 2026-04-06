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
| `nats-token` | nats, all Go services | Shared NATS auth token |
| `encryption-master-key` | auth | 32-byte AES-256 key (base64) |

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

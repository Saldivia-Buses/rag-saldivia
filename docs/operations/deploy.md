---
title: Operations: Deploy
audience: ai
last_reviewed: 2026-04-15
related:
  - ../flows/deploy-pipeline.md
  - ./runbook.md
  - ./monitoring.md
---

## Purpose

How to ship a build of SDA Framework to the inhouse workstation: pipeline
inputs, secrets layout, environment variables, DNS/TLS, Dockerfile rules,
and the rollback procedure. Read before any deploy or rollback.

## Pipeline (GitHub Actions)

`.github/workflows/deploy.yml` builds per-service images on push to a release
tag, pushes to GHCR, and SSHes to the workstation to run
`bash deploy/scripts/deploy.sh`. The script honors the service list from the
workflow input (default: all). Build info (`GIT_SHA`, `BUILD_TIME`) is
injected at compile time and exposed at `/v1/info` on every Go service.

## Preflight checks

Run these before tagging a release. Stop on first failure.

| # | Command | Pass criteria |
|---|---------|---------------|
| 1 | `make build` | All services compile |
| 2 | `make test` | All packages green |
| 3 | `make lint` | No findings |
| 4 | `bash .claude/hooks/check-invariants.sh` | 35/35 pass |
| 5 | `docker compose -f deploy/docker-compose.prod.yml config --quiet` | Valid |
| 6 | `git status --short` | Clean working tree |

## Docker secrets (`deploy/secrets/dynamic/`)

All files are gitignored. Generated locally on the workstation, never copied
out. See `deploy/secrets/README.md` for ownership.

| File | Used by | Source |
|------|---------|--------|
| `jwt-private.pem` | auth | `bash deploy/scripts/gen-jwt-keys.sh` |
| `jwt-public.pem` | all Go services | same script |
| `db-platform-url` | platform, feedback, traces | manual write |
| `db-tenant-template-url` | auth, chat, search, ingest, notification, astro, erp, bigbrother | manual write (template w/ `{slug}`) |
| `db-platform-password` | postgres-platform | manual write |
| `redis-password` | redis | `openssl rand -base64 32` |
| `encryption-master-key` | auth | `openssl rand -base64 32` |
| `bb-kek` | bigbrother | `openssl rand -base64 32` |

## Required environment variables (`deploy/.env`)

Copy from `deploy/.env.example`. Required:

| Variable | Description |
|----------|-------------|
| `SDA_DOMAIN` | Base domain — controls Traefik routing and CORS |
| `SDA_ACME_EMAIL` | Let's Encrypt contact email |
| `SDA_TENANT_SLUG` | Active tenant slug |
| `POSTGRES_PASSWORD` | Platform DB password |
| `{SVC}_NATS_PASS` | One per service: AUTH, WS, CHAT, AGENT, SEARCH, TRACES, NOTIF, PLATFORM, INGEST, FEEDBACK, ASTRO, BIGBROTHER, ERP, EXTRACTOR — must match `deploy/nats/nats-server.conf` |
| `SMTP_HOST` | Notification SMTP server |
| `EPHE_DATA_PATH` | Host path to Swiss Ephemeris data (astro only) |
| `LAN_INTERFACE`, `LAN_SUBNET` | macvlan config (bigbrother only) |

Optional: `SGLANG_LLM_URL`, `SGLANG_OCR_URL`, `SGLANG_VISION_URL`,
`SMTP_PORT`, `SMTP_FROM`, `GRAFANA_ADMIN_PASSWORD`, `GIT_SHA`, `BUILD_TIME`.

Generate per-service NATS passwords at once:

```bash
for svc in AUTH CHAT WS NOTIF INGEST FEEDBACK AGENT TRACES PLATFORM \
           EXTRACTOR ASTRO BIGBROTHER ERP; do
  echo "${svc}_NATS_PASS=$(openssl rand -base64 24)"
done
```

## DNS and TLS

Traefik uses Cloudflare DNS-01 challenge for Let's Encrypt
(`deploy/traefik/traefik.prod.yml.tmpl`). Required:

- DNS A/CNAME for `{SDA_DOMAIN}` and `platform.{SDA_DOMAIN}` → host IP.
- `CLOUDFLARE_DNS_API_TOKEN` available to the Traefik container.
- `SDA_ACME_EMAIL` set.

HTTP redirects to HTTPS automatically. The Traefik dashboard is disabled in
production.

## Dockerfile rules (enforced by invariants)

- **Multi-stage build** — at least 2 `FROM` directives. Build stage compiles,
  runtime stage copies the binary only.
- **Non-root user** — must declare `USER` (e.g., `nonroot`, `appuser`,
  `nobody`, or UID `65534`) before `ENTRYPOINT`.
- Service ports are internal only (Traefik is the sole public ingress on
  443). See port table in `docs/services/README.md`.

## Deploy

```bash
# Full stack
bash deploy/scripts/deploy.sh

# Specific services (rebuild only the ones listed)
bash deploy/scripts/deploy.sh auth chat erp

# Fast rebuild using Docker layer cache
bash deploy/scripts/deploy.sh --cache auth
```

After deploy, `make versions` must show `MATCH` for every service. `STALE`
means the running container is on an older `GIT_SHA` than the source — rebuild
without cache.

## Rollback

### Image rollback (preferred)

Edit `deploy/.env`, set `{SVC}_VERSION=2.0.x` to a previous tag, then:

```bash
docker compose -f deploy/docker-compose.prod.yml up -d {service}
```

Compose pulls the pinned image and restarts only that container.

### Migration rollback (manual)

Migrations are forward-only. To revert:

```bash
psql "$PLATFORM_DB_URL" -f db/platform/migrations/{NNNN}_*.down.sql
psql "$PLATFORM_DB_URL" -c "DELETE FROM schema_migrations \
  WHERE filename = '{NNNN}_*.up.sql'"
```

Repeat against `$TENANT_DB_URL` if the migration touches `db/tenant/`.

### Restore from backup

See [backup-restore.md](./backup-restore.md).

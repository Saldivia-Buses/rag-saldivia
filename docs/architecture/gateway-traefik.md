---
title: Gateway (Traefik)
audience: ai
last_reviewed: 2026-04-15
related:
  - ../operations/deploy.md
  - multi-tenancy.md
  - auth-jwt.md
---

This document describes how external HTTP traffic reaches SDA services.
Read it before adding a new service route, changing TLS or DNS, or
modifying the Cloudflare Tunnel configuration — all inbound traffic flows
through this seam, and the tenant slug is derived here.

## Edge topology

```
client -> Cloudflare DNS (*.${SDA_DOMAIN})
       -> Cloudflare Tunnel (cloudflared, no inbound ports)
       -> Traefik :80 (inhouse workstation)
       -> service:port (Docker network)
```

Cloudflare Tunnel terminates TLS at the edge and connects outbound from
the workstation to Cloudflare's network — the workstation never opens a
port to the internet. Configuration template:
`deploy/cloudflare/config.yml.tmpl`. In production the Cloudflare ACME
provider also issues an internal cert that Traefik serves on `:443`
(`deploy/traefik/traefik.prod.yml.tmpl:15`).

## Production routing (Docker labels)

In `traefik.prod.yml.tmpl` the file provider is intentionally disabled —
the Docker label provider on each service definition is the **sole**
source of routing rules (`traefik.prod.yml.tmpl:23`). Each service in
`deploy/docker-compose.prod.yml` carries its own labels declaring router
rule, entrypoint (`websecure`), TLS resolver (`cloudflare`), and
middlewares. Add a new service by labelling it; do not edit a central file.

## Dev routing (file provider)

In dev the Go services run on the host (`go run ./cmd/...`); only Traefik
runs in Docker. `deploy/traefik/dynamic/dev.yml` maps URL prefixes to
`http://host.docker.internal:{port}` for each service:

| Prefix                | Service        | Port |
|-----------------------|----------------|------|
| `/v1/auth`, `/v1/modules` | auth        | 8001 |
| `/ws`                 | ws             | 8002 |
| `/v1/chat`            | chat           | 8003 |
| `/v1/agent`           | agent          | 8004 |
| `/v1/notifications`   | notification   | 8005 |
| `/v1/platform`, `/v1/flags` | platform | 8006 |
| `/v1/ingest`          | ingest         | 8007 |
| `/v1/feedback`        | feedback       | 8008 |
| `/v1/traces`          | traces         | 8009 |
| `/v1/search`          | search         | 8010 |
| `/v1/astro`           | astro          | 8011 |
| `/v1/erp`             | erp            | 8013 |
| `/v1/healthwatch`     | healthwatch    | 8014 |

Two dev-only middlewares ship in the same file: `dev-cors` (allow
`localhost:3000/3001`, credentialed) and `dev-tenant` (injects fixed
`X-Tenant-Slug: dev` so handlers can resolve a tenant without DNS).

## Tenant slug derivation

In production Traefik extracts the slug from the request subdomain and
sets `X-Tenant-Slug`. The auth middleware then strips any spoofed copy and
re-injects only the value derived from the verified JWT — and finally
**cross-validates** them (`pkg/middleware/auth.go:115`). Mismatch → 403.
This double check is the reason a user holding a saldivia-tenant JWT
cannot use it against `acme.{domain}`.

## Edge defence

`crowdsec` runs alongside Traefik in production
(`deploy/docker-compose.prod.yml:118`) with the `crowdsecurity/traefik`
and `crowdsecurity/http-cve` collections. It tails Traefik's access log
(`traefik_logs:/var/log/traefik:ro`) to detect scanners and HTTP CVEs.
Acquisition rules: `deploy/crowdsec/acquis.yaml`.

## Internal traffic

Service-to-service calls inside the `sda-backend` Docker network bypass
Traefik (e.g. agent → search at `http://search:8010`). The same JWT travels
in the `Authorization` header so the receiving service runs the same
middleware stack.

## What you must never do

- Bind a service port to `0.0.0.0` on the workstation — only Traefik should
  be reachable from the LAN. Use Docker network aliases instead.
- Re-enable the Traefik file provider in production
  (`traefik.prod.yml.tmpl:23` explains why).
- Trust `X-Tenant-Slug` in a handler before the auth middleware has run.

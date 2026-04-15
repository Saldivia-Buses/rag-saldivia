---
title: Service: astro
audience: ai
last_reviewed: 2026-04-15
related:
  - ../packages/llm.md
  - ../packages/traces.md
  - ../packages/cache.md
  - ../architecture/multi-tenancy.md
---

## Purpose

Astro Super Agent: 55+ astrology techniques (natal, transits, progressions,
returns, profections, eclipses, zodiacal releasing, etc.) plus business
intelligence dashboards (cashflow, hiring calendar, team compatibility) and a
streaming SSE query endpoint backed by an LLM. Built with **CGO + Swiss
Ephemeris** (`libswe.a` + ephemeris files at `EPHE_PATH`) — `make build-astro`
requires `CGO_ENABLED=1`. Read this when changing astro endpoints, the
intelligence layer, the chart cache, or the weekly alert cron.

## Endpoints

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/health` | none | Liveness + postgres/redis check |
| POST | `/v1/astro/natal` | JWT + `astro.read` | Natal chart calculation |
| POST | `/v1/astro/transits` | JWT + `astro.read` | Transits over a natal |
| POST | `/v1/astro/{technique}` | JWT + `astro.read` | 30+ technique endpoints — see `services/astro/cmd/main.go:120` |
| POST | `/v1/astro/synastry` | JWT + `astro.read` | Two-chart compatibility |
| POST | `/v1/astro/composite` | JWT + `astro.read` | Composite midpoint chart |
| POST | `/v1/astro/employee-screening` | JWT + `astro.read` | HR screening helper |
| GET | `/v1/astro/contacts` | JWT + `astro.read` | List saved contacts |
| POST | `/v1/astro/contacts` | JWT + `astro.write` | Create contact |
| PUT | `/v1/astro/contacts/{id}` | JWT + `astro.write` | Update contact |
| DELETE | `/v1/astro/contacts/{id}` | JWT + `astro.write` | Delete contact |
| GET/POST/PATCH/DELETE | `/v1/astro/sessions[/...]` | JWT + `astro.read`/`write` | Conversation sessions |
| GET | `/v1/astro/predictions[/stats]` | JWT + `astro.read` | Prediction history |
| POST | `/v1/astro/predictions` | JWT + `astro.write` | Record a prediction |
| PATCH | `/v1/astro/predictions/{id}/verify` | JWT + `astro.write` | Mark verified/missed |
| POST | `/v1/astro/feedback` | JWT + `astro.write` | Submit user feedback |
| POST | `/v1/astro/query` | JWT + `astro.read` (5/min) | SSE-streamed LLM query |
| GET | `/v1/astro/business/*` | JWT + `astro.business` | Dashboards (cashflow, risk, forecast, team, hiring, mercury-rx) |

Read endpoints use `FailOpen=true` (available during Redis outage); write
endpoints use `FailOpen=false`. Per-route timeouts: 5min read, 30s write,
2min business (`services/astro/cmd/main.go:118`).

## NATS events

| Subject | Direction | Trigger |
|---|---|---|
| `tenant.{slug}.traces.start` | pub | LLM query started (`pkg/traces.Publisher`) |
| `tenant.{slug}.traces.end` | pub | LLM query ended with cost |
| `tenant.{slug}.traces.event` | pub | Per-step events from intelligence engine |
| `tenant.{slug}.feedback.{category}` | pub | User feedback submissions |

Astro does not subscribe.

## Env vars

| Name | Required | Default | Purpose |
|---|---|---|---|
| `ASTRO_PORT` | no | `8011` | HTTP listener port |
| `EPHE_PATH` | no | `/ephe` | Swiss Ephemeris data files (CGO) |
| `POSTGRES_TENANT_URL` | no | — | Tenant DB (contacts, sessions, predictions); empty disables persistence |
| `JWT_PUBLIC_KEY` | yes | — | Ed25519 public key |
| `NATS_URL` | no | `nats://localhost:4222` | Trace + feedback events |
| `REDIS_URL` | no | `localhost:6379` | Token blacklist |
| `SGLANG_LLM_URL` | no | — | LLM endpoint for `/v1/astro/query` (empty disables LLM) |
| `SGLANG_LLM_MODEL` | no | — | Model name |
| `LLM_API_KEY` | no | — | Bearer token if using cloud LLM |
| `TENANT_SLUG` | no | `saldivia` | NATS subject prefix |

## Dependencies

- **PostgreSQL tenant** (optional) — contacts, sessions, predictions, feedback.
- **Redis** — token blacklist + chart cache (`services/astro/internal/cache`).
- **NATS** — trace and feedback publishing via `pkg/traces`.
- **LLM** (SGLang or cloud) — only the SSE query endpoint and intelligence
  engine.
- **Swiss Ephemeris (CGO)** — embedded via `libswe.a`; planetary positions
  computed in-process. Container runs distroless but ships the static lib.

## Permissions used

- `astro.read` — all calculation, contact-list, session-list, prediction-list
  endpoints, plus the SSE query.
- `astro.write` — contact/session/prediction/feedback mutations.
- `astro.business` — `/v1/astro/business/*` dashboards.

A weekly cron (Mondays ~09:00 UTC, `services/astro/cmd/main.go:222`) is
scaffolded for SA/DP urgency scans but currently only logs intent.

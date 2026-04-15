---
title: Service: notification
audience: ai
last_reviewed: 2026-04-15
related:
  - ../packages/nats.md
  - ../architecture/websocket-hub.md
  - ./auth.md
  - ./feedback.md
---

## Purpose

In-app + email notifications. Subscribes to `tenant.*.notify.>` and persists
each event as a per-user notification (auth login/lockout, chat new messages,
agent results, platform lifecycle, etc.). Holds user delivery preferences,
serves the inbox API, exposes an admin-only `/send` for platform-driven
emails, and runs an internal Alertmanager webhook so monitoring alerts get
emailed without leaving the cluster. Read this when changing the consumer
contract, mailer config, or admin send-path.

## Endpoints

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/health` | none | Liveness + postgres/nats/redis check |
| GET | `/v1/notifications/` | JWT | List my notifications (filterable, paginated) |
| GET | `/v1/notifications/count` | JWT | Unread count for badge |
| POST | `/v1/notifications/read-all` | JWT | Mark all read |
| PATCH | `/v1/notifications/{notificationID}/read` | JWT | Mark one read |
| GET | `/v1/notifications/preferences` | JWT | Get my channel preferences |
| PUT | `/v1/notifications/preferences` | JWT | Update preferences |
| POST | `/v1/notifications/send` | JWT (admin role) | Send a notification (in-app or email) — blocks arbitrary relay |
| POST | `/internal/webhook/alert` | shared token | Receive Prometheus Alertmanager webhooks (`services/notification/cmd/main.go:127`) |

Routes assembled in
`services/notification/internal/handler/notification.go:39`. The internal
webhook is conditionally registered only when both `POSTGRES_PLATFORM_URL`
and a 32+ byte `ALERTMANAGER_WEBHOOK_TOKEN` are present.

## NATS events

| Subject | Direction | Trigger |
|---|---|---|
| `tenant.*.notify.>` | sub | Stream consumer with filter `tenant.*.notify.>` (`services/notification/internal/service/consumer.go:54`) |

Notification does not publish — it sinks events into PostgreSQL and SMTP.
Tenant slug is parsed from `msg.Subject()` (never payload) — see
`services/notification/internal/service/consumer_test.go` for the
spoofing-prevention test.

## Env vars

| Name | Required | Default | Purpose |
|---|---|---|---|
| `NOTIFICATION_PORT` | no | `8005` | HTTP listener port |
| `POSTGRES_TENANT_URL` | yes | — | Notifications + preferences |
| `POSTGRES_PLATFORM_URL` | no | — | Alert persistence (enables webhook) |
| `JWT_PUBLIC_KEY` | yes | — | Ed25519 public key |
| `NATS_URL` | no | `nats://localhost:4222` | Subscriber |
| `REDIS_URL` | no | `localhost:6379` | Token blacklist |
| `SMTP_HOST` | no | `localhost` | SMTP host |
| `SMTP_PORT` | no | `1025` | SMTP port |
| `SMTP_FROM` | no | `noreply@sda.local` | From address |
| `ALERTMANAGER_WEBHOOK_TOKEN` | no | — | Shared secret for `/internal/webhook/alert` (≥32 bytes; also accepts `/run/secrets/alertmanager_webhook_token`) |
| `ALERT_CRITICAL_EMAIL` | no | — | Where to email critical alerts |

## Dependencies

- **PostgreSQL tenant** — `notifications`, `notification_preferences`.
- **PostgreSQL platform** — alerts table (only when webhook is configured).
- **NATS JetStream** — subscriber on `tenant.*.notify.>`.
- **SMTP server** — `service.SMTPMailer`.
- **Redis** — token blacklist.
- No outbound calls to other application services.

## Permissions used

No `RequirePermission` middleware. Authorization is header-based
(`requireUserID`, `requireAdmin` —
`services/notification/internal/handler/notification.go:204`):

- `requireUserID` for personal notification routes.
- `requireAdmin` (checks `X-User-Role == admin`) for `POST /send`.
- Bearer-token verification for the internal webhook is enforced inside the
  handler, not via JWT.

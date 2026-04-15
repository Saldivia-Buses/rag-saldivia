---
title: Flow: Self-Healing Triage
audience: ai
last_reviewed: 2026-04-15
related:
  - ../services/healthwatch.md
  - ./deploy-pipeline.md
---

## Purpose

How HealthWatch turns raw signals into scrubbed summaries, then how the
nightly Sonnet triage opens GitHub issues and how Alertmanager pushes
real-time alerts to the notification service. Read this before changing
collectors, the `Summary` API, the daily workflow, the alertmanager
routes, or the post-deploy auto-close logic. Architecture (data model,
retention) lives in `services/healthwatch.md`.

## Steps

1. HealthWatch (`HEALTHWATCH_PORT=8014`) wires three collectors in
   `services/healthwatch/cmd/main.go`: `collector.Service` (HTTP
   `/health`), `collector.Prometheus` (metrics + alerts), and
   `collector.Docker` (container + host stats via the docker-socket
   proxy, never the raw socket).
2. `Service.Summary`
   (`services/healthwatch/internal/service/healthwatch.go:97`) calls
   `ServiceStatuses`, `ActiveAlerts`, and `collectInfra`, computes the
   overall status, and asynchronously persists a snapshot via
   `persistSnapshots` (best-effort, `context.WithoutCancel`).
3. The HTTP layer exposes `GET /v1/healthwatch/summary` at
   `services/healthwatch/internal/handler/healthwatch.go:60`,
   authenticated by JWT and gated to platform admins; output is the
   scrubbed `HealthSummary` (no raw errors, IPs, or credentials —
   invariant `M3`).
4. The nightly `Daily Health Triage` workflow
   (`.github/workflows/daily-triage.yml`) fires at `0 22 * * *` on the
   `workstation-1` runner and calls
   `deploy/scripts/get-service-token.sh healthwatch` to mint a service
   JWT (`DS4`).
5. The workflow `curl`s `http://localhost:8014/v1/healthwatch/summary`
   into `/tmp/health-summary.json` and validates it with
   `jq -e '.overall_status'`; the file is the only channel — health data
   is never interpolated into a prompt string (`DS6`).
6. `anthropics/claude-code-action` runs Sonnet 4.6 with a fixed prompt
   that reads the JSON file and emits a `findings[]` array plus an
   `executive_summary`, with severity `critical|high|medium|low|info`.
7. The "Create GitHub issues" step parses the AI output via `jq` (with
   regex/Python fallbacks for fenced JSON) and pipes
   `severity == critical or high` findings into `gh api ... /issues`
   with labels `severity:{level}`, `auto-triage`, `service:{name}`.
8. The "Send email summary" step builds the payload with `jq -n
   --arg body ...` and POSTs to
   `http://notification:8005/v1/notifications/send`; failure is
   non-fatal (`echo ::warning::`), the issues are still authoritative.
9. In real time, Prometheus alerts hit Alertmanager
   (`deploy/observability/alertmanager/config.yml`); the `default` and
   `critical` receivers POST to
   `http://notification:8005/internal/webhook/alert` with bearer auth
   from the `alertmanager_webhook_token` Docker secret.
10. After every successful deploy, `Post-Deploy Verification`
    (`.github/workflows/post-deploy.yml`) re-checks health and only
    closes `auto-triage` issues whose `service:{name}` label appears in
    the live `healthy` list — never bulk-closes (`2M`).

## Invariants

- HealthWatch responses are scrubbed at the source; no log lines, IPs,
  or stack traces leave `Service.Summary` (invariant `M3`).
- The triage workflow uses `ANTHROPIC_API_KEY_TRIAGE` (separate from
  developer keys, invariant `M5`); rotate it independently.
- AI output is always written to a file and parsed with `jq`/`python`;
  it is never spliced into shell strings (`DS6`).
- Docker stats come through the socket proxy, never the raw
  `/var/run/docker.sock` (`DS5`).
- Critical Alertmanager routes use `group_wait: 10s` /
  `repeat_interval: 1h`; warnings batch every 4h. Inhibit rules
  suppress warnings while a critical for the same service is firing.
- Auto-close MUST verify the specific service is healthy; never close an
  `auto-triage` issue blindly on deploy success.

## Failure modes

- `Health summary is not valid JSON` — HealthWatch returned non-JSON
  or 5xx; the workflow exits 1, no issues are filed; check
  `services/healthwatch/internal/service/healthwatch.go:97` and
  Prometheus reachability.
- `Could not parse triage output as JSON` (warning) — the model emitted
  prose or Markdown; the step exits 0 to avoid breaking the workflow,
  fix the prompt or the parser.
- Email step warning — notification service is missing the `/send`
  endpoint or down; issues still land via `gh api` so the loop is
  resilient.
- Alert delivered but no notification — check Alertmanager bearer
  secret (`/run/secrets/alertmanager_webhook_token`) and the
  notification service `/internal/webhook/alert` handler.
- Auto-close skipped (warning "Could not fetch service health") —
  service token expired or HealthWatch down; rerun the workflow once
  HealthWatch recovers, no issue state is lost.
- Issue without a `service:` label — the post-deploy script logs and
  skips it (`post-deploy.yml`); fix the triage prompt to always emit
  `service`.

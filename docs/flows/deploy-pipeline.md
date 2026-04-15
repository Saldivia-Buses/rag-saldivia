---
title: Flow: Deploy Pipeline
audience: ai
last_reviewed: 2026-04-15
related:
  - ../operations/deploy.md
---

## Purpose

End-to-end sequence of `Deploy SDA` (`.github/workflows/deploy.yml`):
discover services, validate the host, build images on GitHub-hosted
runners, deploy on the workstation, smoke-test, and roll back on
failure. Read this before changing any deploy job, helper script, or
the rollback contract. Operational details (secrets, runners, domains)
are in `operations/deploy.md`.

## Steps

1. `discover` walks `services/*/Dockerfile` (excluding `.scaffold` and
   `extractor`) and emits a JSON matrix consumed by every downstream job
   so adding a new service requires no workflow edit
   (`.github/workflows/deploy.yml:24`).
2. `preflight` runs on the `workstation-1` self-hosted runner and
   executes `deploy/scripts/preflight.sh` — 13 checks for `.env`,
   Docker, GPU, free ports, JWT secrets, Cloudflare credentials,
   generated configs, disk, and RAM; any `fail` aborts the workflow.
3. `build` fans out across the matrix on `ubuntu-latest`, logs into
   GHCR with the workflow token, and builds each image tagged with
   `${GITHUB_SHA}` plus build metadata
   (`.github/workflows/deploy.yml:57`).
4. `generate-deploy-env` writes one `{SERVICE}_VERSION=${SHA}` line per
   discovered service to `deploy/.env.deploy` and uploads it as a
   workflow artifact for the deploy jobs (line 97).
5. `deploy-dev` (when `inputs.environment=dev` or a tag) downloads the
   artifact, runs `docker compose --env-file deploy/.env.deploy -f
   deploy/docker-compose.dev.yml up -d --pull always` on the `dev-pc`
   runner, then `deploy/scripts/health-check.sh --env dev --timeout 120`.
6. `deploy-prod` requires the `production` GitHub environment (manual
   approval); on the workstation runner it captures rollback state
   first via `deploy/scripts/save-versions.sh` into a `chmod 600`
   tempfile referenced by `ROLLBACK_FILE` (line 175).
7. The job pulls the SHA-pinned images and runs `docker compose ... up -d
   --pull always` against `docker-compose.prod.yml`, then
   `deploy/scripts/health-check.sh --env prod --timeout 120` polls every
   service's `/health` until pass or timeout.
8. On smoke-test failure the same job runs
   `deploy/scripts/rollback.sh "$ROLLBACK_FILE"` and immediately calls
   `deploy/scripts/record-deploy.sh --status rollback` so the platform
   DB reflects the abort (lines 195–207).
9. `release` (depends on `deploy-prod` success) records the deploy with
   `record-deploy.sh --status success` and emits a `::notice::` line
   linking the version + SHA (line 213).
10. `Post-Deploy Verification` (`.github/workflows/post-deploy.yml`)
    waits 60 s, re-runs the prod health check, fetches per-service
    status from `/v1/healthwatch/services` using
    `deploy/scripts/get-service-token.sh healthwatch`, and auto-closes
    `auto-triage` issues whose `service:{name}` label is now healthy.

## Invariants

- Builds always run on GitHub-hosted runners (`ubuntu-latest`); the
  self-hosted workstation is read-only for build-related jobs (`DS1`).
- Image tags are immutable SHAs — `latest` is never deployed; the
  `.env.deploy` artifact is the single source of which SHAs went out.
- Rollback state is captured BEFORE the new `up -d`; if `save-versions`
  fails the deploy aborts (line 178), so there is always a known-good
  target to roll back to.
- Health check failure during `deploy-prod` triggers the rollback step
  in the SAME job (`ROLLBACK_FILE` is `GITHUB_ENV`-scoped to the job).
- `concurrency: deploy-production` with `cancel-in-progress: false`
  serializes deploys; never override this — half-applied deploys are
  worse than queued ones.

## Failure modes

- `preflight` fails — read its output: missing `.env`, expired JWT
  keys, port already bound, low disk/RAM, or stale generated configs;
  fix locally and retry.
- `build` fails for one service — matrix `fail-fast: false` lets others
  finish; rerun the failing matrix slot after fixing the Dockerfile or
  context.
- `health-check.sh` times out — `deploy/scripts/health-check.sh:69`
  loops every 5 s; the failure block prints which `svc:port` did not
  return 200, check that container's logs.
- Rollback runs but issue persists — `rollback.sh` only restores
  images; data migrations are not reverted, fix forward.
- `release` skipped — `deploy-prod` failed; no success record is
  written, so dashboards stay on the previous version.
- `post-deploy.yml` cannot fetch healthwatch — the service account
  token is missing or HealthWatch is down; auto-close skips silently
  rather than mass-closing issues.

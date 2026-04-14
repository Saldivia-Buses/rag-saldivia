---
name: Plan 23 S3 review (tasks 2.4-2.6)
description: Review of commits 43224715, e89da92b, 7f797d04 — deploy workflow, circuit-breaker scripts, prod compose dependency ordering
type: project
---

Session 3 review of Plan 23 (deploy pipeline tasks 2.4-2.6). CAMBIOS REQUERIDOS.

Blocker 1: deploy-prod skipped when workflow_dispatch uses environment=prod.
`deploy-prod` has `needs: [deploy-dev, ...]` but `deploy-dev` has an `if` that is false for prod-only dispatch. GitHub Actions skips downstream when a needs-dep is skipped (unless `always()` used). Fix: remove `deploy-dev` from `deploy-prod` needs, or add `if: always()` on `deploy-prod` + check `needs.deploy-dev.result != 'failure'`.

Blocker 2: `services_list` expression interpolated into shell run block (generate-deploy-env step:111). DS6 violation — a malicious service directory name could inject shell.

Must-fix 1: SECRET injected raw into JSON body in get-service-token.sh:48 via shell interpolation. If secret contains `"` or `\n` the JSON is malformed or injectable. Use `--data-urlencode` / `jq --arg` instead.

Must-fix 2: record-deploy.sh hardcodes `"service": "sda-platform"` regardless of which service was deployed. Misleading in deploy_log.

Must-fix 3: rollback does not run health-check after re-deploy, so a failed rollback looks like success.

**Why:** Deploy pipeline is the safety net for prod — if the prod-only path silently skips, the entire circuit-breaker is dead on the most common prod deploy path.
**How to apply:** Flag any GitHub Actions `needs` + `if` combination where the needs-dep may be conditionally skipped.

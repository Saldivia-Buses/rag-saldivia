---
title: Convention: Git Workflow
audience: ai
last_reviewed: 2026-04-15
related:
  - ./testing.md
  - ./error-handling.md
  - ../operations/deploy.md
---

Branching, commit format, and PR rules. CI enforces commit format and gates listed below; pre-commit hook (`bash .claude/hooks/check-invariants.sh`) blocks commits that violate structural invariants.

## Branches

DO branch off `2.0.5` (current development branch). Long-lived branches: `main` (production-protected), `2.0.5` (active dev). Feature branches are short-lived (1–3 days).

DO name branches by intent: `feat/<scope>-<short-desc>`, `fix/<scope>-<short-desc>`, `refactor/<scope>-<short-desc>`, `docs/<scope>`.

DON'T commit directly to `main` or `2.0.5`. Every change goes through a PR.

## Commit format

`type(scope): description` — lowercase, imperative, no trailing period.

Allowed types: `feat`, `fix`, `refactor`, `test`, `docs`, `ci`, `chore`. Plus `perf` and `revert` accepted by commitlint.

Allowed scopes: `auth`, `chat`, `agent`, `search`, `astro`, `traces`, `ingest`, `extractor`, `notification`, `ws`, `platform`, `feedback`, `web`, `cli`, `mcp`, `infra`, `pkg`, `docs`, `deploy`, `bigbrother`, `healthwatch`, `erp`.

Examples:
- `feat(auth): add totp-based mfa flow`
- `fix(chat): handle empty message array on session fork`
- `refactor(rag): extract collection resolver to pkg/tenant`

DO keep subject ≤ 72 chars. Body explains the *why* (problem) and *how* (solution) — a reviewer should understand both from `git show` alone, without opening files.

DO use `!` after type for breaking changes: `feat(auth)!: rotate JWT signing keys`.

DON'T use vague subjects like `fix bug`, `update code`, `wip`. CI blocks them via commitlint.

## One concern per commit

DO split unrelated changes into separate commits. A drive-by lint fix during a feature commit is its own commit (`chore(scope): fix lint`).

DO commit after each working file when running an agent — never accumulate >3 file changes per commit during agent execution.

## Pull Requests

1. Branch off `2.0.5` → implement → run `make test && make lint && make build` locally.
2. Update docs **in the same PR** as the code (every PR-affected feature has a doc update; see Documentation rule below).
3. Push branch → CI runs (4 gates) + AI review (3 passes).
4. CI green + AI review without critical/high findings → squash merge.
5. Post-merge: version bump + changelog + Docker image build via release workflow.

DO use squash merge for feature branches. DON'T use merge commits or rebase merges into protected branches.

DON'T mark a PR ready for review with unresolved TODO/FIXME in the diff.

## Documentation in the same PR

| Change | Doc to update |
|---|---|
| New endpoint | OpenAPI spec for the service |
| New service | Service README + `docs/services/<name>.md` + `CLAUDE.md` |
| New module | Module manifest YAML + spec |
| Architectural decision | New ADR under `docs/decisions/` |
| Convention change | This file + `docs/bible.md` |

PR body must include a checklist confirming docs are updated when the diff touches behaviour.

## Hooks and signing

DON'T pass `--no-verify` to `git commit` or `git push`. The pre-commit invariant hook and pre-push checks must always run. If a hook fails, fix the underlying problem and create a new commit — never amend or bypass.

DON'T use `git commit --amend` after a failed hook unless the hook actually completed and you are intentionally rewriting the previous commit. Hooks that exit non-zero block the commit; the next commit is fresh.

## CI gates

Gates run sequentially; later gates depend on earlier ones.

1. **Verify** — commitlint, `go build`, `go vet` (parallel inside the gate)
2. **Test** — `make test`
3. **Security** — gosec + trivy
4. **Docker** — build all service images (matrix from `services/*/Dockerfile`)

AI review (3 parallel passes on every PR): Quality (Opus), Security (Opus), Dependencies (Sonnet). Critical/high findings block merge. Address by either fixing or replying with justification — don't ignore.

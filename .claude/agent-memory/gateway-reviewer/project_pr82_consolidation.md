---
name: PR #82 Plan 07 Consolidation review
description: Foundation refactoring -- centralized migrations, pkg/config, NATS standardization. Blockers: event type injection still unfixed, duplicate migration files.
type: project
---

PR #82 consolidates migrations from per-service dirs into `db/tenant/` (8 files) and `db/platform/` (4 files), adds `pkg/config.Env()/MustEnv()`, and standardizes NATS with `Connect()` factory and exported `IsValidSubjectToken`.

Blockers:
1. NATS Notify STILL does not validate `parsed.Type` -- same gap from PR #37/#52. Dots are intentional but wildcards/spaces/control chars are not blocked. Need `IsValidEventType` with `^[a-zA-Z0-9_][a-zA-Z0-9_.-]*$`.
2. Old migration files in `services/*/db/migrations/` not removed. Dual existence with `db/` creates maintenance hazard. Recommended: symlinks or sqlc path overrides.

Must-fix: migrate.sh has no transaction wrapping (002's ADD CONSTRAINT is not idempotent on retry), check-then-act race condition, dead `../services` volume mount in docker-compose, SQL interpolation in shell (use psql -v variables).

**Why:** This is the foundation for Plan 07. NATS event type injection has been open since PR #37 and this PR claims to "standardize" NATS publishing, so it must close the gap.

**How to apply:** When reviewing future NATS PRs, verify event types are validated. When reviewing migration changes, verify single source of truth.

---
name: PR #131 Plan 16 Backend Polish
description: pkg/server bootstrap package + 12 main.go migrations + Dockerfiles + compose + tests, blockers: Timeout kills WS/SSE, chat compose broken, agent Dockerfile missing modules runtime copy
type: project
---

PR #131: Plan 16 Backend Polish for 2.0.5. feat/plan16-polish -> 2.0.5. 6 phases, 45 files, +2078 -1781.

**CAMBIOS REQUERIDOS** -- 4 blockers, 6 must-fix.

Key blockers:
- B1/B2: `server.New()` applies `middleware.Timeout(30s)` unconditionally, killing WS connections and capping astro SSE at 30s. Needs WithNoTimeout/WithTimeout option.
- B3: `db_tenant_template_url` listed under `expose:` instead of `secrets:` in chat prod compose -- Docker Compose won't start chat.
- B4: Agent Dockerfile copies `modules/` to builder but not runtime stage -- agent starts with zero module tools in Docker.

Must-fix:
- Duplicate REDIS_URL in WS prod compose
- Astro logs raw NATS URL with credentials (all others use RedactURL)
- `go-chi/chi/v5` missing from `pkg/go.mod` (works via workspace, but architecturally wrong)
- WS blacklist initialized but never wired to upgrade handler (revoked JWT can connect)
- Run/RunWithWriteTimeout code duplication
- Scaffold template not updated to use server.New()

**Why:** pkg/server centralizes boilerplate correctly but the unconditional 30s timeout breaks long-lived connections. Compose errors would prevent prod deploy.

**How to apply:** The timeout issue requires an API change to server.New() before WS/astro can use it. The compose and Dockerfile issues are simple fixes.

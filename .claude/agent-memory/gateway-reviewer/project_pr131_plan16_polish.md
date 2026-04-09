---
name: PR #131 Plan 16 Backend Polish
description: pkg/server bootstrap package + 12 main.go migrations + Dockerfiles + compose + tests. Re-review found 1 remaining blocker (astro WriteTimeout) and 1 unfixed must-fix (WS blacklist). 7/9 original findings fixed.
type: project
---

PR #131: Plan 16 Backend Polish for 2.0.5. feat/plan16-polish -> 2.0.5. 7 commits after fix commit.

**Re-review: CAMBIOS REQUERIDOS** -- 1 remaining blocker, 1 unfixed must-fix.

Fixed (7/9):
- B3: chat compose secret -- FIXED
- B4: agent Dockerfile modules/ copy -- FIXED
- M1: duplicate REDIS_URL ws -- FIXED
- M2: astro RedactURL -- FIXED
- M3: chi in pkg/go.mod -- FIXED
- M5: Run/RunWithWriteTimeout dedup -- FIXED
- M6: scaffold uses server.New() -- FIXED

Remaining:
- B1/B2 partially fixed: chi timeout disabled via WithTimeout(0) for WS+astro, but astro uses app.Run() (30s WriteTimeout) instead of RunWithWriteTimeout(0). SSE endpoint /v1/astro/query and 5-min route timeout on read endpoints will be killed at 30s by HTTP WriteTimeout.
- M4 unfixed: WS blacklist initialized but never passed to handler.NewWS(). Revoked JWTs can still connect to WebSocket.

**Why:** The astro WriteTimeout mismatch means SSE streaming and long calculations (up to 5 min) fail after 30s in production Docker. The WS blacklist gap means logout doesn't revoke WS connections.

**How to apply:** astro main.go needs `app.RunWithWriteTimeout(0)` instead of `app.Run()`. WS handler.NewWS needs blacklist parameter + check in Upgrade handler before JWT verify or after.

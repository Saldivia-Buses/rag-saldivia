---
name: PR #126 Astro Plan 12 review
description: Massive 10k LOC PR adding 44 techniques, intelligence layer, business module, sessions, predictions, quality system to astro service
type: project
---

PR #126 reviewed 2026-04-08. 79 files, 10,232 insertions.

Key blockers found:
1. sqlc queries not generated -- sessions/predictions/feedback/usage handlers reference 13+ queries that don't exist in queries.sql.go (won't compile)
2. Line 621 astro.go uses `ctx.Brief` but variable is `fullCtx` (compile error)
3. GetMessages handler checks tenant_id but not user_id (user isolation gap)
4. UpdateSession swallows DB errors (silent failures)
5. No prompt injection guardrails on user query to LLM
6. No category/outcome validation before DB (500 instead of 400)

**Why:** Plan 12 adds intelligence layer + business module + session management. Architecture is clean but sqlc generation was not run and several runtime bugs exist.

**How to apply:** These findings should be fixed before merge. The sqlc queries.sql needs 13+ queries added and regenerated. The compile errors are hard blockers.

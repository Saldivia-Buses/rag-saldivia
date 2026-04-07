---
name: PR #88 Plan 08 Phase 2 review
description: PR review of 9 security depth fixes -- rate limiting, ownership checks, JTI auto-gen, FailOpen, cachedModelConfig, ReadHeaderTimeout
type: project
---

Reviewed PR #88 (Plan 08 Phase 2) on 2026-04-05. Status: CAMBIOS REQUERIDOS.

2 blockers:
1. Chat integration test expects `ErrNotOwner` but `GetSession` with SQL-level filtering returns `ErrSessionNotFound` -- test will fail
2. Chat handler test `TestAddMessage_ValidRole_Success` iterates over {user,assistant,system} expecting 201, but handler now blocks system role with 403 -- test will fail

5 corrections:
- FailOpen comment says "(default)" but Go bool zero is false -- misleading docs
- `cachedModelConfig` cache-hit path silently ignores DB error when fetching API key -- empty key causes confusing downstream failures
- `golang.org/x/time` is indirect in go.mod but directly imported by ratelimit.go
- Token bucket semantics differ from "N requests per window" label (burst + refill allows ~2x advertised rate)
- WS service WriteTimeout 30s could kill WebSocket connections (coder/websocket hijacks but deadline persists)

Pattern: ownership checks at SQL level (WHERE id AND user_id) is the right design, but sentinel errors must match. When SQL returns no rows for both "not found" and "wrong owner", don't declare a separate ErrNotOwner unless the service layer distinguishes the cases.

**Why:** tracks blockers for Plan 08 Phase 2 merge
**How to apply:** verify test fixes before approving. The token bucket semantics should be documented even if not changed.

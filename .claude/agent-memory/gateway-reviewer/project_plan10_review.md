---
name: Plan 10 Backend Polish review
description: Review of Plan 10 (backend polish), 3 blockers (llm interface mismatch, WS JWT expiry undocumented), 7 must-fix items (platform blacklist gap, multi-tenant blacklist+enckey, partial failure semantics lost, OpenAPI scope creep)
type: project
---

Plan 10 reviewed 2026-04-05. CAMBIOS REQUERIDOS.

Key findings:
- B1: `llm.ChatClient` interface uses signatures (`ChatRequest`, `SimplePrompt(ctx, sys, user)`) that don't match actual code
- B3: WS mutations store raw JWT in Client struct; mutations stop working after 15min token expiry -- needs documented strategy
- M2: Platform `requirePlatformAdmin` does NOT check token blacklist -- revoked admin tokens remain valid
- M3: Multi-tenant auth `resolveService()` doesn't pass blacklist or encryption key to per-tenant service
- M4: Batch insert changes partial-failure to all-or-nothing (behavioral change not documented)
- SC1: OpenAPI (Fase 8) is scope creep -- adds new deps + endpoints, not "polish"

**Why:** Plan is the last backend round before frontend. Security gaps (M2, M3) would ship unfixed. Interface mismatch (B1) would cause compile errors.

**How to apply:** B1 must be fixed before implementation starts (interface won't compile). M2+M3 should be added to Fases 1-2. SC1 should be reconsidered or moved to separate plan.

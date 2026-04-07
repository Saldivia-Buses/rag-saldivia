---
name: Plan 02 rewrite context
description: Plan 02 Full Wiring rewritten 2026-04-03 — original 11 phases reduced to 6, 70% was already implemented
type: project
---

Plan 02 (Full Wiring) was rewritten on 2026-04-03. The original plan had 11 phases covering everything from "frontend makes no fetch" to Cloudflare Tunnel. After audit, 70% was already implemented.

The rewritten plan has 6 phases covering only what remains:
1. Backend PATCH /v1/auth/me (auth service — new endpoint)
2. Settings Profile wire to backend (replace hardcoded data)
3. System Settings localStorage persistence (language, timezone)
4. Search Command dynamic list (from enabled modules)
5. WebSocket chat.message invalidation (React Query)
6. Notification Preferences UI (backend already exists)

**Why:** Enzo requested the rewrite after a frontend audit showed most wiring was done.
**How to apply:** When referencing Plan 02, always use the rewritten version. The original scope (CLI, MCP, OpenAPI, gRPC, OTel, Cloudflare) is either done or belongs in other plans.

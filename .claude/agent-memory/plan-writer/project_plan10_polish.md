---
name: Plan 10 backend polish
description: Backend polish plan written 2026-04-05, 9 phases covering security residual (H1/H2/M6), batch insert, interfaces, WS mutations via gRPC, OpenAPI baseline, hot reload
type: project
---

Plan 10 covers the last round of backend work before going to frontend.

**Why:** closes 3 audit HIGHs + 1 MEDIUM, adds batch insert performance, DX improvements (OpenAPI, interfaces, Routes(), hot reload), and gRPC expansion (Chat server + WS mutations).

**How to apply:** this plan should be fully implemented before starting any frontend plan (Plan 11+). Phases 1-2 are security and should be done first. Phases 3-5 are independent. Phase 6 (Chat gRPC + WS mutations) is the largest item. Phase 8 (OpenAPI) is baseline only (auth, chat, search).

Key decisions:
- FailOpen: true for non-auth services (availability > security)
- Pool tuning: keep at 4 conns/pool, add env var override
- OpenAPI: only 3 services as baseline, evaluate before expanding
- Auth Routes(): split into PublicRoutes() + ProtectedRoutes()

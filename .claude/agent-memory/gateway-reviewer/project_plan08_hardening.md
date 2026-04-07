---
name: Plan 08 Backend Hardening review
description: Review of 52-finding hardening plan -- critical issues with factual errors, scope creep (gRPC should be separate), missing Traefik config updates
type: project
---

Reviewed Plan 08 (backend hardening) on 2026-04-05. Status: CAMBIOS REQUERIDOS.

Key findings:
- C1 audit partially wrong: 4 services already have audit (auth, chat, search, notification). Missing: ingest, platform, agent, feedback, traces
- C2 rate limit partially wrong: Traefik prod already has 100req/s/IP rate limit. Plan adds application-level granularity
- Fase 4 (gRPC 16-20h) is scope creep -- feature, not hardening. Should be Plan 09
- Missing critical: Traefik configs still reference deprecated rag service on 8004, missing routes for agent/search/traces
- Missing: docker-compose.prod.yml has rag service definition (non-existent Dockerfile), missing agent/search/traces
- Fases 2 and 3 can run in parallel (dependency graph is over-constrained)
- H5 (cache service structs) should be higher priority
- M9/L4/L5 are in wrong phases

**Why:** establishes the review record for Plan 08 hardening decisions
**How to apply:** reference when implementing Plan 08 or reviewing related PRs

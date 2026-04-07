---
name: Plan 09 gRPC review
description: Plan 09 gRPC inter-service communication review -- 7 bloqueantes (compile errors in interceptor), 8 must-fix (parity gaps with HTTP auth)
type: project
---

Plan 09 gRPC reviewed 2026-04-05. Result: CAMBIOS REQUERIDOS.

**Why:** The plan proposes gRPC for 2 of 4 inter-service HTTP calls (agent->search, agent->ingest). Scope is correct and pragmatic.

**Bloqueantes (7):** All compile errors in proposed code:
- `sdajwt.NewContext()` does not exist (use `tenant.WithInfo()`)
- `tenant.NewContext()` does not exist (use `tenant.WithInfo()`)
- `claims.TenantSlug` field does not exist (use `claims.Slug`)
- `sdajwt.Verify(token, pub)` arg order reversed (should be `pub, token`)
- `grpc.DialContext` deprecated in v1.63 (use `grpc.NewClient`)
- `h.svc.Query()` does not exist (use `h.svc.SearchDocuments()`)
- `derefString()` and `toProto()` referenced but never defined

**Must-fix (8):** Feature parity gaps:
- gRPC interceptor missing role/permissions injection (breaks RBAC)
- No blacklist check (revoked tokens accepted via gRPC)
- No MFA-pending rejection
- Error message leaks internal details
- No `JWTFromIncomingContext` helper for gRPC-to-gRPC chaining

**How to apply:** When this plan gets implementation PR, verify all 7 compile errors are fixed and all 8 parity gaps are addressed.

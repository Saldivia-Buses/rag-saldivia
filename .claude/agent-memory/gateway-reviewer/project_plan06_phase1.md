---
name: Plan 06 Phase 1 review
description: PR #64 reviewed 2026-04-05 -- infra for GPU model serving (SGLang) and object storage (MinIO + pkg/storage). Blockers: missing go.mod deps, missing Makefile targets.
type: project
---

PR #64 reviewed 2026-04-05. Plan 06 Phase 1: SGLang (sglang-ocr :8100, sglang-vision :8101) + MinIO + pkg/storage.

**Blockers found:**
- pkg/go.mod missing aws-sdk-go-v2 deps -- code won't compile
- Makefile missing dev-gpu and test-storage targets

**Must-fix:**
- EnsureBucket falls through to CreateBucket on any error (should check types.NotFound)
- Put() missing ContentType support -- will matter for Phase 2
- Get() doesn't distinguish not-found from other errors (needs sentinel ErrNotFound)

**Why:** This is the foundation for all subsequent Plan 06 phases. Getting the Store interface right now avoids breaking changes later.

**How to apply:** When reviewing Phase 2+ PRs, verify they import pkg/storage correctly and use ErrNotFound for control flow.

---
name: PR #98 Agent gRPC Client for Search
description: Plan 09 Phase 4 -- agent-side gRPC client for search_documents tool. APPROVED with 3 must-fixes.
type: project
---

PR #98 reviewed 2026-04-05. Result: APPROVED.

**Must-fix (no blockers):**
1. `grpc_search.go:67` -- `json.Marshal(resp)` silently drops error on proto message; use `protojson.Marshal` instead. `encoding/json` on a proto message can return null silently.
2. `grpc_search.go:47` -- unmarshal error leaks Go error text to LLM context. Use generic message.
3. `executor.go:103` -- `ExecuteConfirmed` hardcodes HTTP path, bypasses gRPC even if wired. Not a bug today (search has RequiresConfirmation=false) but needs the same nil-check.

**What's correct:**
- JWT forwarding: `sdagrpc.ForwardJWT(ctx, jwt)` → metadata "authorization" → interceptor strips Bearer prefix → Ed25519 verify. Full round-trip works.
- Fallback to HTTP is transparent -- nil check on grpcSearch, caller unaware.
- CollectionId optional: `if p.CollectionID != ""` sets proto pointer correctly. Server reads `if req.CollectionId != nil`.
- Non-fatal gRPC init in main.go is the right pattern.
- Permission check is server-side (grpc.go:29-40 from PR #97 fix), client does not need to re-check.

**Why:** protojson vs encoding/json matters for proto messages -- a silent null would cause the LLM to hallucinate results.
**How to apply:** Any time a proto message is serialized to send to the LLM, use protojson not encoding/json.

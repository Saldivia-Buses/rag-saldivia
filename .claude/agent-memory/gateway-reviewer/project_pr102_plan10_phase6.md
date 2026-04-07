---
name: PR #102 Plan 10 Phase 6 review
description: Chat gRPC server + WS Hub mutations, CAMBIOS REQUERIDOS, 2 blockers (gRPC error leak, JWT expiry unhandled), 4 must-fix (UserID fallback escalation, role not validated, limit ignored, context.Background() no timeout)
type: project
---

PR #102 reviewed 2026-04-05. CAMBIOS REQUERIDOS.

Key findings:
- B1: `mutations.Handle()` sends `err.Error()` to WS client — full gRPC status string and transport errors leak to browser. Fix: map gRPC codes to clean messages.
- B2: `client.JWT` is set at WS upgrade and never refreshed. After 15min access token expires, all mutations silently fail with gRPC Unauthenticated (which leaks per B1). Needs either token-refresh mutation (Option A) or structured `token_expired` error event (Option B, simpler).
- M1: All 7 gRPC RPCs fall back to `req.UserId` when context is empty — privilege escalation if interceptor is misconfigured. Remove fallback, fail hard with `codes.Unauthenticated`.
- M2: `AddMessage` passes `req.Role` to service without validating it — WS client can inject `system` role messages.
- M3: `ListMessages` ignores `req.Limit` from proto, hardcodes 100. Use req.Limit with cap.
- M4: `dispatch()` uses `context.Background()` — no timeout, can block forever. Use `context.WithTimeout(10s)`.

What was good: protojson.Marshal (correct, per PR #98 feedback), graceful nil fallback for Mutations, dual-listener pattern, shared service layer, GracefulStop ordering.

**Why:** B1+B2 are user-visible failures (browser receives internal gRPC errors; mutations break after 15min). M1 is a security gap — spoofed UserID accepted if interceptor fails silently.

**How to apply:** B1 must be fixed alongside B2 (fixing B2 without B1 still leaks auth errors to client). M1 fix is a single line per RPC — high leverage.

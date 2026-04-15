---
title: Convention: Error Handling
audience: ai
last_reviewed: 2026-04-15
related:
  - ./go.md
  - ./logging.md
  - ./security.md
  - ../packages/httperr.md
---

How errors flow through SDA services. Three layers: repository (raw infra errors) → service (sentinel + wrap) → handler (translate to `httperr.Error` and JSON response).

## Wrap with context

DO wrap every returned error with `fmt.Errorf("<verb> <noun>: %w", err)`. The verb describes what was being attempted, never just `"%w"`.

- `return nil, fmt.Errorf("get user by email: %w", err)`
- `return fmt.Errorf("publish notification event: %w", err)`

DON'T return `err` unmodified across package boundaries — the call site loses traceability.

DO use `%w` (not `%v`) so callers can `errors.Is` / `errors.As` upstream.

## Sentinel errors

DO declare sentinel errors as package-level `var ErrXxx = errors.New("...")` for distinguishable failure modes the caller may want to handle differently.

- `ErrInvalidCredentials`, `ErrAccountLocked`, `ErrInvalidRefreshToken` — see `services/auth/internal/service/`.
- `ErrInvalidToken`, `ErrMissingClaim`, `ErrInvalidKey` — see `pkg/jwt/jwt.go:25`.

DO check with `errors.Is(err, service.ErrXxx)` in handlers (`services/auth/internal/handler/auth.go:145`).

## HTTP response format

This is an architectural invariant. Every error response from an HTTP API is JSON.

DO use `pkg/httperr` for every error response from a handler:
- `httperr.WriteError(w, r, httperr.InvalidInput("email is required"))`
- See `pkg/httperr/httperr.go:94` and `services/auth/internal/handler/auth.go:118`.

Available constructors: `Internal(cause)`, `NotFound(resource)`, `InvalidInput(msg)`, `InvalidID(id)`, `Conflict(msg)`, `Unauthorized(msg)`, `Forbidden(msg)`, `Wrap(cause, code, msg, status)`.

DON'T call `http.Error(w, "msg", 400)` from a handler. It writes plain text. The pre-commit invariant scans for this pattern.

DON'T hand-roll JSON error bodies — use `httperr.WriteError`. It sets `Content-Type`, picks the status, logs at the right severity, and emits the canonical schema `{"error":"...","code":"..."}`.

## Translating errors at the boundary

DO map service-layer errors to HTTP statuses inside the handler:

```
switch {
case errors.Is(err, service.ErrInvalidCredentials):
    httperr.WriteError(w, r, httperr.Unauthorized("invalid email or password"))
case errors.Is(err, service.ErrAccountLocked):
    httperr.WriteError(w, r, httperr.Wrap(nil, httperr.CodeForbidden, "too many attempts", http.StatusTooManyRequests))
default:
    httperr.WriteError(w, r, httperr.Internal(err))
}
```

(Pattern from `services/auth/internal/handler/auth.go:144`.)

DO hide the underlying error from the response body. `httperr.Internal(cause)` logs the cause with full context but tells the client only `"internal error"`.

## Never swallow errors

DON'T write `_ = err` to discard a returned error. Either handle it, log it, or return it.

DON'T write `if err != nil { return nil, nil }` — silently dropping errors makes bugs invisible.

The only acceptable swallowing is in `defer` cleanup paths where the original error has already been captured and the cleanup error is genuinely best-effort:
- `defer func() { _ = rows.Close() }()` after the main error has already returned.

## NATS event publishing

DO publish a NATS event for every successful write. The WebSocket Hub depends on this — without the event, the frontend never updates.

DO publish *after* the DB transaction commits, never before. If the publish fails after the commit, log at error level and continue — the write happened, the event is the secondary effect.

DON'T let a NATS publish failure roll back the transaction. The DB is the source of truth; events are a derived projection.

## Panics

DON'T panic in request paths. Recover middleware exists (`pkg/middleware`) but reaching it is a bug.

DO panic in `init()` or service bootstrap when the process cannot operate (missing private key, malformed config). Failure at boot is preferable to running with broken invariants.

## Logging vs returning

DO log at the layer that decides what to do about the error. Repositories and pure logic return errors; handlers and bootstrap log them.

See [logging](./logging.md) for slog patterns, structured fields, and PII rules.

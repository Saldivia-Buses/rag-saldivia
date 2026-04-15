---
title: Package: pkg/approval
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./audit.md
  - ./plc.md
---

## Purpose

Generic two-person approval pattern for operations that need a second user to
authorize before execution (PLC critical writes, tenant deletion, data purge).
Defines types and a `Store` interface — the consuming service supplies its own
storage implementation backed by its own table (e.g., BigBrother uses
`bb_pending_writes`). Import this when an action MUST NOT execute on a single
person's say-so.

## Public API

Source: `pkg/approval/approval.go:14`

| Symbol | Kind | Description |
|--------|------|-------------|
| `Status` | type | Enum: `pending`, `approved`, `expired`, `rejected` |
| `StatusPending`/`StatusApproved`/`StatusExpired`/`StatusRejected` | const | Status values |
| `PendingAction` | struct | One pending request: ID, TenantID, ResourceID, Action, Payload, Requestor, ApprovedBy, ExpiresAt |
| `PendingAction.IsExpired()` | method | True if `time.Now() > ExpiresAt` |
| `CreateRequest` | struct | Inputs for a new pending action (TenantID, ResourceID, Action, Payload, RequestorID, TTL) |
| `CreateRequest.Validate()` | method | Required-field check |
| `Store` | interface | `Create`, `Approve`, `Reject`, `Get`, `CleanExpired` |
| `ErrNotFound`/`ErrExpired`/`ErrSelfApprove`/`ErrAlreadyExists`/`ErrAlreadyHandled` | var | Sentinel errors |

## Usage

```go
req := approval.CreateRequest{
    TenantID:    tid, ResourceID: registerID, Action: "plc_write",
    Payload:     payload, RequestorID: userID, TTL: 30 * time.Minute,
}
if err := req.Validate(); err != nil { return err }
pending, err := store.Create(ctx, req)
// later, second user approves:
done, err := store.Approve(ctx, pending.ID, otherUserID)
```

## Invariants

- `Store.Approve` MUST be atomic and reject when `requestor_id == approver_id`
  (returns `ErrSelfApprove`). Implementations use `UPDATE ... WHERE status='pending'`
  to guarantee single-winner semantics.
- `Store.Create` MUST enforce one pending action per resource via a partial
  unique index on `status='pending'`.
- Caller is responsible for periodic `CleanExpired` (cron or background loop).
- `PendingAction.Payload` is opaque bytes — the consumer interprets it.

## Importers

None in production code yet — `services/bigbrother/internal/service/plc.go:18`
mentions the package in a doc comment but does not import it. `Store`
implementations live in consuming services.

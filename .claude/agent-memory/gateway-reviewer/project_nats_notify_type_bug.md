---
name: NATS Notify event type not validated
description: publisher.Notify validates tenant slug but not the event type field, allowing subject injection via type
type: project
---

In PR #52 review, found that `pkg/nats/publisher.go` Notify method validates the tenant slug via `isValidSubjectToken` but does not validate `parsed.Type` before building the NATS subject `tenant.{slug}.notify.{type}`. An event type containing NATS special characters could create invalid or malicious subjects.

**Why:** The Broadcast method was fixed in a previous review to validate both slug and channel. The Notify method was missed.

**How to apply:** When reviewing NATS publishing code, always check that ALL tokens used in subject construction are validated via `isValidSubjectToken`.

---
name: NATS channel validation gap
description: Publisher.Broadcast does not validate the channel parameter before concatenating into NATS subject -- subject injection risk found during PR #37 review
type: project
---

Publisher.Broadcast (pkg/nats/publisher.go) validates tenantSlug with isValidSubjectToken but does NOT validate the channel parameter. The channel is concatenated directly into the NATS subject: `"tenant." + tenantSlug + "." + channel`. A channel value containing dots, wildcards (`*`, `>`), or whitespace would create malformed or dangerously broad NATS subjects.

Similarly, Notify does not validate the event Type field which also gets interpolated into subjects. Dots are intentional (e.g., "chat.new_message") but wildcards are not validated against.

**Why:** Found during PR #37 test review. The tests correctly validate slug injection but miss channel injection entirely.
**How to apply:** Track as a production bug fix. When reviewing future NATS-related PRs, check that ALL user-influenced components of NATS subjects are validated.

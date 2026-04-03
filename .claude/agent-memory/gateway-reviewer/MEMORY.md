# Gateway Reviewer Memory

- [First review context](project_first_review.md) -- PR #34 Platform Service was the first review, established patterns
- [NATS channel validation gap](project_nats_channel_bug.md) -- Broadcast does not validate channel param, subject injection risk
- [NATS Notify type not validated](project_nats_notify_type_bug.md) -- Notify validates slug but not event type, found in PR #52
- [Auth service multi-tenant WIP](project_auth_single_tenant.md) -- PR #54 adds dual-mode, blockers: slug-as-tenantID, missing Traefik header for login

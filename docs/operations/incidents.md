---
title: Operations: Incident Response
audience: ai
last_reviewed: 2026-04-15
related:
  - ./runbook.md
  - ./monitoring.md
  - ./backup-restore.md
---

## Purpose

How to respond when production breaks: severity ladder, on-call workflow,
communication, and the postmortem template. The runbook tells you *how* to
fix specific failures; this doc tells you *what to do as a human* during the
incident.

## Severity ladder

| Sev | Definition | Examples | Response time |
|-----|------------|----------|---------------|
| **Sev-1** | Data integrity loss or cross-tenant leak | User from tenant A sees tenant B data; backup restore fails verification | Stop traffic immediately |
| **Sev-2** | Full outage of core flow | Login broken; chat unable to send; WS not accepting upgrades | Within 15 min |
| **Sev-3** | Degraded experience | One service down; high latency; one tenant affected | Within 1h |
| **Sev-4** | Cosmetic or non-blocking | Slow dashboard; intermittent error in low-traffic endpoint | Next business day |

A Sev-1 always overrides everything. If in doubt, treat as one severity
higher than your initial guess.

## On-call

The current owner is Enzo. Escalation path:

1. Incident detected (dashboard, alert, user report).
2. Owner acknowledges within 15 min for Sev-1/2.
3. If owner unreachable for 30 min: pause public traffic at Cloudflare,
   document the action.
4. Resume on-call recovery.

## Response workflow

### 1. Detect and classify

- Pull up Grafana service overview and the affected dashboard.
- Read [monitoring.md](./monitoring.md) watch-list to confirm severity.
- Check `make versions` and last commit — recent deploy?

### 2. Stabilize

Order of preference:

1. **Roll back** to the last known good image — see
   [deploy.md](./deploy.md) Rollback section.
2. **Restart** the affected service if logs indicate a transient state.
3. **Reduce blast radius** — if one tenant is affected, disable that tenant
   in the platform DB; if one feature, feature-flag it off.
4. **Cut traffic** — last resort. Pause the public hostname at Cloudflare.

For a suspected Sev-1 cross-tenant leak: stop the offending service, do
**not** restart, audit the query, fix, deploy, then verify before reopening
traffic.

### 3. Diagnose

Use [runbook.md](./runbook.md) to bisect. Capture:

- Time of first failure (Loki query timestamp).
- Affected services, endpoints, tenants.
- Last commit before the failure.
- Logs around the first failure (save to `docs/artifacts/`).

### 4. Fix and verify

- Apply the fix (rollback, hotfix, config change).
- Verify with the relevant health endpoint and a smoke flow
  (login → chat → write → WS push).
- Confirm no recurrence over a 15-minute observation window.

### 5. Communicate

- During: short status updates ("auth login failing, rolling back"). No
  speculation.
- After: a one-line resolution note ("auth login restored at HH:MM via
  rollback to image 2.0.4").

## Postmortem

Write within 48h of resolution for any Sev-1 or Sev-2. Save to
`docs/artifacts/postmortem-{YYYY-MM-DD}-{slug}.md`. Use this template:

```markdown
# Postmortem — {short title}

**Date:** {YYYY-MM-DD}
**Severity:** Sev-{1|2|3|4}
**Duration:** {start time} → {end time} ({minutes})
**Author:** {name}

## Summary
{2-3 sentences on what broke and what users saw}

## Timeline (UTC)
- HH:MM — first error in logs
- HH:MM — alert fired
- HH:MM — owner acknowledged
- HH:MM — root cause identified
- HH:MM — fix applied
- HH:MM — verified resolved

## Root cause
{The technical reason, with file:line references. Use 5-whys if not obvious.}

## Impact
- Tenants affected: {list or "all"}
- Endpoints affected: {list}
- Data integrity impact: {none | <description>}

## What went well
- {list}

## What went poorly
- {list}

## Action items
| # | Action | Owner | Type | Due |
|---|--------|-------|------|-----|
| 1 | {description} | {name} | prevent / detect / respond | {date} |

## Detection gap
{If the alert fired late or never, what monitor needs to be added.}
```

## Action item discipline

Every postmortem produces at least one action item. Each is filed as a
GitHub issue with the `incident` label. Action items are reviewed monthly —
overdue items are escalated to Enzo.

## Never do

- Skip the postmortem because "everyone knows what happened".
- Blame an individual — postmortems are blameless. Fix the system.
- Apply a hotfix without writing it down — undocumented hotfixes are how
  invariants drift.
- Leave a workaround in place without an action item to remove it.

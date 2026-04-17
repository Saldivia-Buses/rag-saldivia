# ADR 021 — Reduce before add: consolidation is the default

**Status:** accepted
**Date:** 2026-04-16
**Deciders:** Enzo Saldivia

## Context

The codebase currently holds ~89 k lines of Go split across 15 services and 32
packages, running on a **single vertical-scale workstation** (Threadripper Pro 9975WX,
256 GB DDR5 ECC, 8 TB NVMe, 1× RTX PRO 6000 Blackwell, 96 GB VRAM). It is developed
and operated by **one person**.

The shape was inherited from an earlier assumption of horizontal scale and a
marketable multi-tenant SaaS. Neither matches the current reality:

- Network latency between services is artificial — they all run on the same kernel.
- "Multi-tenant" scaffolding exists (DB-per-tenant, tenant-namespaced NATS subjects,
  tenant-aware middleware in every service) but the real customer is Saldivia.
- Each new service adds boilerplate (main.go, Dockerfile, VERSION, health probes,
  JWT wiring, migrations, config, logging) proportional to service count, paid every
  deploy and every refactor.
- Most shared `pkg/*` packages have 1–3 importers; several have zero.

A one-dev team cannot honestly maintain 15 services plus 32 packages while shipping
new behavior. The complexity tax is paid; the capability it buys is hypothetical.

## Decision

**Consolidate, delete, and unify is the default direction.** Every decision — every
PR, every new feature, every refactor — starts with the question:

> Can this be done by removing or merging something, instead of adding?

Consequences operationalized across the harness:

- `CLAUDE.md` principle **#0: Reduce before adding** — elevated above all others.
- `continuous-improvement` skill's primary axis is **consolidation** (delete → inline → merge → question).
- `code-review` skill requires an explicit **simplicity vote** per review, and adds a
  `bloat` category for findings.
- `backend-go` skill blocks new services/packages unless they can defend a real
  process boundary (independent release cadence, runtime, or scaling profile).

The concrete north star is:

- 15 services → **3–5** (grouped by domain, not noun).
- 32 packages → **~10** (delete 0-importer ones; inline 1–2 importer ones).
- 89 k LOC → **whatever expresses the product with fewer lines.**
- Multi-tenant scaffolding: **question every instance**; if it doesn't support a
  real product requirement, remove it.

## Consequences

**Positive**

- Less surface area to maintain, debug, deploy.
- Faster iteration: fewer moving parts per feature.
- Easier onboarding (for future collaborators and for Claude).
- Better use of the box: less serialization/RPC overhead, more monolith-friendly
  Postgres tuning.
- Closer alignment between the project's complexity and its team size (one person).

**Negative**

- Short-term refactor cost: consolidation is not free.
- Loses optionality for horizontal scaling. Acceptable because horizontal scaling
  is not a near-term need — the box is vertical.
- Some abstractions that looked defensible in isolation won't survive the
  "does this justify a process boundary?" test. That's the point.

**Neutral**

- Supersedes the implicit microservices-first posture of ADRs 013 and 017 without
  formally deprecating them (those still apply to genuine process-boundary needs).

## Alternatives considered

1. **Continue adding features on the current shape.**
   Rejected. Complexity tax compounds; one dev cannot keep pace.

2. **Microservices for future horizontal scale.**
   Rejected. There is no cluster. The workstation is the scale plan. Premature
   distribution is the most expensive form of premature optimization here.

3. **Big-bang rewrite to a monolith.**
   Rejected. Too risky. The consolidation bias is a direction, not a weekend project.
   Merges and deletes happen one area at a time, as the `continuous-improvement`
   loop picks them up.

4. **Keep multi-tenant because "maybe SaaS later".**
   Rejected as a rationale. If and when a real SaaS plan materializes, multi-tenant
   scaffolding can be reintroduced with measured scope. Carrying it indefinitely
   against a hypothetical future is the exact sort of cost this ADR is about.

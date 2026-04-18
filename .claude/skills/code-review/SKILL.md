---
name: code-review
description: Use when reviewing a diff, a PR, or your own changes before merge. Defines the severity taxonomy (blocking/critical/high/medium/low), the category axis (security/correctness/performance/simplicity/architecture), the per-scope checklist (backend-go, frontend-next, auth-security, database, deploy-ops, rag-pipeline), and how to dispatch parallel reviewer subagents for wide diffs. Every finding produces either a fix or an explicit waiver — never "noted".
---

# code-review

Scope: any diff about to land — a PR, a commit, a staged change, or the result of
your own work before declaring done.

## The principle

Code reviews on this project are load-bearing. A missed critical finding in auth,
tenant isolation, NATS namespacing, or migrations causes outages or data leaks.
The review is where bugs die. Treat it accordingly.

Every finding ends as one of:

- **Fixed** in the same PR.
- **Tracked** in `docs/plans/` with a concrete owner and deadline.
- **Waived** with a written rationale captured in the PR description.

Never "noted", "future work", "we'll see", or "out of scope".

## Severity taxonomy

| Severity | Meaning | Disposition |
|---|---|---|
| **blocking** | Merge is blocked until fixed. Breaks an invariant, production-impacting, or violates a contract. | Fix before merge. |
| **critical** | Security hole, data loss risk, silent cross-tenant leak, lost write, auth bypass. | Fix before merge. No exceptions. |
| **high** | Correctness bug under realistic load, missing index on tenant query, unhandled error in a write path. | Fix in the same PR. |
| **medium** | Logic smell, risky refactor, missing test for a new branch, unclear naming in a hotspot file. | Fix in the same PR or track. |
| **low** | Style, micro-refactor, comment nit. | Fix if cheap, otherwise skip. |

**Blocking** and **critical** overlap; use blocking for invariant breaks and
critical for security/data correctness.

## Category axis

Every finding has a category. A single issue can have multiple (e.g. security +
correctness).

- **security** — auth bypass, RBAC hole, secret leak, tenant boundary violation,
  unsafe deserialization, SQL/NoSQL injection, log exposure of PII.
- **correctness** — wrong result, race, unhandled error, resource leak, missing
  context cancellation.
- **performance** — N+1 query, unbounded allocation, missing index, chatty RPC.
- **simplicity** — code that does the job but with more machinery than needed.
  Simpler alternative exists with the same behavior.
- **bloat** — net-positive LOC without net-positive behavior. New package / service /
  abstraction that isn't justified by a concrete current need.
- **architecture** — couples layers incorrectly, bypasses a shared package,
  introduces a new pattern where an existing one fits.
- **testability** — no test covers the new branch, test mocks where it must hit real infra.

### Simplicity vote (mandatory)

Every review ends with an explicit **"what could this PR have deleted?"** answer.
If the reviewer can't find at least one simplification the PR missed, say so —
but that's the rare case, not the default. The project's direction is reduction;
reviews reinforce it.

## Per-scope checklist

Apply the invariant checklist for each scope the diff touches. Do not skip.

### backend-go
- [ ] Every new tenant query includes `WHERE tenant_id = $1`.
- [ ] Every public handler is behind the right middleware (JWT, tenant, rate limit).
- [ ] Errors are wrapped with `%w` and rendered via `pkg/httperr` at the edge.
- [ ] `ctx` is the first parameter on every I/O-doing function and actually propagated.
- [ ] No new pool opened inline; uses `pkg/tenant.Resolve`.
- [ ] Race detector passes (`go test -race`).
- [ ] Table-driven tests, testify, real DB/NATS via testcontainers for integration.
- [ ] No log of secrets, JWTs, or PII.
- [ ] Any `pkg/*` change: importers grepped, migration path considered for breaking changes.

### frontend-next
- [ ] Server Components by default; `"use client"` only where strictly needed.
- [ ] No polling introduced — real-time uses the WS hub.
- [ ] Auth guard present on new `(core)/` routes.
- [ ] No JWT parsing client-side.
- [ ] Tokens (CSS vars) used for all colors/spacing; no hardcoded hex.

### auth-security
- [ ] No identity read from headers/body.
- [ ] NATS subjects built via `pkg/nats.Subject()`, tenant-namespaced.
- [ ] Role checks live in the service layer, not handlers.
- [ ] New endpoint has an explicit role requirement or is explicitly public.
- [ ] Two-tenant isolation test exists for any new tenant-scoped code path.

### database
- [ ] `.up.sql` has a matching `.down.sql`; numbering contiguous.
- [ ] Foreign-key columns indexed.
- [ ] Tenant queries have `tenant_id` as the leading column of their index.
- [ ] sqlc regenerated and committed; generated files not hand-edited.
- [ ] Non-null column additions follow the 4-step zero-downtime pattern.

### deploy-ops
- [ ] `VERSION` bumped for every service whose binary changed.
- [ ] No `:latest` pins.
- [ ] `HEALTHCHECK` present on new containers.
- [ ] Migrations run as a separate step, not in service startup.

### rag-pipeline
- [ ] No new vector-store call site introduced.
- [ ] Tree-search contract preserved; no shortcut lookup in `search`.
- [ ] Crossdoc phase count unchanged.
- [ ] Tier logic in `smart_ingest.py` not altered without benchmark evidence.

## The review flow

1. **Read the PR description.** If the PR doesn't state the intent, stop — ask
   for it. Reviewing without intent is guessing.
2. **Map the diff by scope.** What does this touch: backend-go? frontend? auth?
   Multiple? The applicable checklists are the union.
3. **Parallel research for wide diffs.** If the diff crosses ≥ 3 scopes, dispatch
   the `parallel-research` flow: one Explore agent per scope, each applying its
   own checklist. Synthesize.
4. **For each file changed:** walk the checklist. Record findings with severity +
   category + file:line.
5. **Prioritize simplicity.** If the change is correct but complex, propose the
   simpler alternative. Simple + functional is the target. More code is worse code
   unless it buys real capability.
6. **Report.** Group findings by severity, highest first. For each: what,
   why, suggested fix or waiver.

## Report template

```
## Review — <PR title>

### Blocking (N)
- [security, auth-security] pkg/jwt/verify.go:47 — accepts `none` alg when header is missing.
  Fix: reject unsigned tokens; assert alg in {EdDSA}.

### Critical (N)
- [security, tenant-isolation] services/app/internal/realtime/chat/handler/chat.go:118 — list endpoint omits tenant_id.
  Fix: use ListChatsByTenant, add two-tenant integration test.

### High (N)
- …

### Medium (N)
- …

### Low / nits (N)
- …

### Simplicity wins
- services/app/internal/realtime/chat/service/chat.go:60 — two helpers unused once the branch collapses.
  Drop them; net -40 LOC, same behavior.

### Waivers
- [architecture] pkg/llm/client.go: new adapter duplicates 30 LOC from pkg/remote.
  Waived: adapter is temporary, will be unified in plan24. Owner: Enzo. ETA: 2.0.7.
```

## Dispatching reviewer subagents

For large PRs (> ~400 LOC changed or ≥ 3 scopes), parallelize:

```
Agent(Explore,  "backend-go checklist on services/chat + pkg/chi changes — report findings")
Agent(Explore,  "auth-security checklist on the diff; focus on tenant boundary + JWT")
Agent(general-purpose, "test coverage audit: what branches added in this PR have no test?")
```

Each returns a scoped report. You synthesize, assign severity, and decide.

Do **not** delegate severity assignment to a subagent. That is a judgment call
that needs the whole picture.

## Anti-patterns

- "LGTM" without walking the checklist.
- Accepting "out of scope" for a `critical`/`blocking` finding.
- Filing a "follow-up" ticket for a security finding.
- Review that only catches style — means the substance was skipped.
- Approving your own code as a shortcut. Self-review is allowed, but hold yourself
  to the same checklist and be explicit in the PR description.
- Waiving a `bloat` finding with "we'll simplify later". "Later" is a waiver with
  no owner and no date. If it's not worth fixing now, it's not worth tracking.

## Hard rejects (no discussion)

These block merge on sight. No negotiation, no "follow-up":

- **Reintroduces pool-tenant** — new `tenant_id` column, `WHERE tenant_id`, new
  `pkg/tenant` call site, tenant-namespaced subject. ADR 022 supersedes all of those.
- **New service or package that can't defend a process boundary** (see `backend-go` §
  "Before you create anything new"). Put it inside an existing service/package.
- **Polling instead of WS subscription** on anything tenant-scoped in the frontend.
- **Credentials in the repo** — password, key, token, `.env` with real values.
- **`log.Printf` / `fmt.Println` in production paths** — must be `slog.*Context`.
- **Client-side JWT parsing** — identity is server-side only.
- **Supabase added for backend-adjacent needs** — auth, tenants, data flow through the gateway.

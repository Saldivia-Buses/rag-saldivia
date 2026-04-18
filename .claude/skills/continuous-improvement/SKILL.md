---
name: continuous-improvement
description: Default mode when no task is set — walks ADR 027 top-down, picks the first un-ticked un-blocked item, ships. Phase 0 gates always win over any Phase 2+ work. Karpathy-style auto-research within the phase ordering.
---

# continuous-improvement

Scope: the whole codebase, one item per session, picked against ADR 027.

## The operating mode

This project runs on continuous improvement **against a fixed checklist**
(ADR 027). Every session:

1. Read **ADR 027** (`docs/decisions/027-mvp-success-criteria.md`) top-down.
2. Find the first un-ticked item whose dependencies are ticked (Phase 0 gates
   always first; Phase N+1 only if Phase N is green or the item is waived).
3. Ship it. Tick the ADR in the same PR.

You are not here to maintain the status quo. You are here to iterate the
codebase toward the ADR 026 north star (SDA replaces Histrix + the AI layer
that follows).

## The hard rule — phase precedence

| When | Do |
|---|---|
| A Phase 0 gate is failing (ghost rows, tool perms, prod drift) | **That's the task.** No Phase 2+ work ships. |
| A Phase 1 parity item is failing AND Phase 0 is green | Take a Phase 1 item. |
| Phase 1 is complete (or every remaining item is waived) | Phase 2 is open. |
| All earlier phases are green or waived | Pick from the next phase. |

**"I found a nice refactor" does not override this.** The ADR 027 walk is
the source of truth for what to work on; personal preference is not.

### Primary bias inside a phase: parity first, consolidation second

Within whatever phase is currently active:

- **Phase 0 work:** correctness + integrity. Find the root cause, fix it,
  add the check that catches it next time. See `systematic-debugging` and
  `migration-health`.
- **Phase 1 work:** replicate a Histrix capability. Start by reading the
  relevant XML-form in `.intranet-scrape/`. See `htx-parity`.
- **Phase 2+ work:** build the SDA-differential capability. The skills
  `agent-tools`, `prompt-layers`, `background-agents` each own a slice
  of this.

Within any phase's work, **consolidation remains a first-class axis**:

- **Delete** before refactor — if code isn't reachable, it goes.
- **Inline** before abstract — a package with 1–2 importers is usually
  worse than inlining.
- **Merge** before split — adjacent services (same domain, same data)
  over HTTP on the same kernel should be one service, not two.
- **Question** before accept — every surviving piece of complexity must
  earn its keep this session.

A session that ends with `−800 LOC + Phase 1 tick + same tests green`
beats a session that adds a feature without reducing or ticks nothing.

## The loop

```
1. PICK an area (service, package, flow, layer, doc).
2. MEASURE it — what is actually there, what is the baseline.
3. HUNT for improvements across axes (see below).
4. SCORE each finding: impact vs. cost.
5. EXECUTE the top 1–3 items that pass the bar.
6. VERIFY — tests green, invariants intact, measurable win.
7. RECORD — commit, update ADR if architectural, next session picks the next area.
```

Stop the loop when: nothing in the current area clears the bar, or the user
redirects. Move to the next area next session.

### Automating the loop

Claude Code supports two primitives for running this on autopilot:

- **`/loop`** — run a slash command on a recurring interval, or let the model self-pace.
  Example: `/loop 20m /continuous-improvement` after setting up a prompt that invokes this skill.
- **`routines`** (research preview) — a prompt + repo + connectors bundled to run on
  a schedule, via API call, or on a webhook. Useful for nightly "pick an area, find 3
  wins, open a PR" workflows.

Use them for off-hours iteration; keep the bar (`impact/cost ≥ 3×`) the same.
A scheduled session that ships nothing is a successful session.

## Picking work

**Default: walk ADR 027 top-down.** That's the primary picker.

Secondary signals apply only as tie-breakers **within** a phase, or when
ADR 027 items in the active phase are all blocked on external input
(rare but possible):

- **Hotspot file** — `git log --since="3 months ago" --name-only --pretty=format: | sort | uniq -c | sort -rn | head -20`
- **Skipped area** — no commits in 60+ days; silent rot risk.
- **User complaint** — anything recently flagged as annoying / slow.
- **Friction you hit** — a slow dev command or tangled flow surfaced
  during the current session.
- **Test-gap signal** — missing integration test for a Phase 0 / Phase 1
  path.

On any new session without a user task:

1. `git log --oneline -20` — what just shipped.
2. Read ADR 027 + project_vision memory — where we are.
3. `grep "\[ \]" docs/decisions/027-*.md` — remaining items.
4. Pick the first un-ticked item whose dependencies (earlier phase)
   are all ticked or waived.
5. If the chosen item is hard-blocked on external input (e.g. a cutover
   date), record the block and move to the next un-blocked item
   **in the same phase**. Don't jump phases.

## Axes of improvement

For the chosen area, scan every axis:

| Axis | Ask |
|---|---|
| **Consolidation** | Can this service/package/module merge with a neighbor? Can this package be inlined? |
| **Dead code** | Anything unreachable, unused, or superseded and left behind? Delete on sight. |
| **Simplicity** | Can this be shorter with the same behavior? Delete, don't add. |
| **Correctness** | Does it do the right thing in the edge cases I can name? |
| **Performance** | Is there an obvious N+1, unbounded allocation, missing index, chatty RPC? |
| **Clarity** | Would a new engineer understand this in 60 seconds? |
| **Tests** | Is the core contract covered by a test that fails when the contract breaks? |
| **Invariants** | Does this area still obey the hard rules? |
| **Architecture** | Does this bypass a shared package, duplicate a pattern, diverge from neighbors? |
| **Docs** | Does CLAUDE.md or the right skill still describe this area accurately? |
| **DX** | Is there a command, helper, or log line that would save time on every touch? |

Most findings land on *simplicity* and *clarity*. Those are the cheapest wins and
they compound.

## Impact / cost scoring

Before you change anything, score each finding — roughly, not with a spreadsheet.

- **Impact**: how many future changes benefit? How many bugs does this prevent?
  How much faster does the system get? How much code disappears?
- **Cost**: lines changed, blast radius, review effort, risk of regression.

Bar to execute:

- `impact / cost ≥ 3×` for in-session execution.
- `1–3×` → add to a short list, execute next session if still on top.
- `< 1×` → drop it. Don't hoard TODOs.

## What "simple + functional = excellent" looks like

- Deleting 40 lines and keeping behavior = win.
- Replacing a custom helper with a standard library call = win.
- Collapsing three nearly-identical handlers into one with a parameter = win.
- Removing a feature flag that has been `true` for six months = win.
- Adding a missing index on a hot tenant query = win.
- Rewriting a tangled function into one pass = win.

What does **not** count:

- Renaming for aesthetics with no readability gain.
- Refactoring to hit a pattern the rest of the repo doesn't use.
- Adding an abstraction in case you need it later.
- Touching a file just to bump test coverage without a real scenario.

## Guardrails

1. **One area per session.** Don't touch five services in one pass; the review
   burden explodes and regressions hide.
2. **Improvements ship with evidence.** A perf claim has a before/after measurement.
   A correctness claim has a new test. A simplicity claim has a LOC delta and an
   unchanged behavior test.
3. **Architectural wins update the ADR** — use the `decisions` skill.
4. **Don't drift.** If the user arrives with a task mid-session, drop the improvement
   work and switch. Improvement is the default, not the priority.
5. **Every commit is atomic.** "Refactor X" is one commit. "Refactor X and also
   fix Y and rename Z" is three.

## No stopping rule

A session ends when **the work meets the bar**, not when you get tired of it.
The bar is: the area you picked is closer to the target the harness describes —
smaller, simpler, more correct, more consolidated — and the evidence is on disk.

If you hit a blocker (unclear requirement, external dep, need for user input),
record the blocker explicitly in the session end template, then pick a *different*
area the same session. Don't coast to a stop because the first pick stalled.

When the user has not set a task, the default order of operations every session:

1. Read `CLAUDE.md` → confirm invariants are current.
2. Check the **anti-pattern hunt list** below — run one grep, pick the category
   with the most hits, that's your area.
3. Pick items that clear `impact/cost ≥ 3×` and ship them.
4. Record what shipped; if nothing clears the bar in the chosen area, move on to
   the next category and try again. A session should not end with zero shipped
   work unless every anti-pattern category is empty — which it is not.

## Anti-pattern hunt list

Known categories of avoidable complexity / inefficiency in this codebase. When
picking an area, start here. Each entry has a **detection command** so you can
measure progress objectively.

| Category | Detection | Bias | Audited count |
|---|---|---|---|
| Tenant-awareness residual (ADR 022) | `rg -c "tenant_id\|pkg/tenant\|TenantSlug" services/ pkg/` | **Delete on sight.** | **~2,480 hits** |
| Dead `pkg/*` (0 importers) | `for p in pkg/*/; do n=$(basename $p); c=$(rg -l "pkg/$n" services/ pkg/ \| grep -v "/$n/" \| wc -l); echo "$n $c"; done \| awk '$2==0'` | **Delete.** | 3 confirmed (approval, featureflags, cache) |
| Main.go boilerplate | diff `services/*/cmd/main.go` pairwise | **Extract `pkg/server.Bootstrap`.** | 14 services, ~300-400 LOC dup each |
| `panic()` in request paths | `rg -n "panic\(" services/ -g '!*_test.go'` | **Replace with httperr.** | 2 confirmed |
| Goroutines without shutdown | `rg -nB1 "go func\(" services/ -g '!*_test.go'` | **Add ctx + select.** | 6 confirmed |
| Client-side fetching on auth routes | `rg -n 'useEffect.*fetch\|useQuery.*fetch' apps/web/src/app/\(core\)` | **Move to Server Component.** | — |
| Polling | `rg -n "setInterval\|refetchInterval" apps/web/` | **Switch to WS hub subscription.** | — |
| Supabase auth calls | `rg -n "supabase\.auth\." apps/web/` | **Route through Go auth gateway.** | **14 violations** |
| `"use client"` without state/effects | manual review of `components/ui/*.tsx` | **Drop the directive.** | **68 candidates** |
| Sync-imported heavy libs | `rg -n "import \* as (THREE\|Y\|math\|faceapi)" apps/web/` | **Dynamic import.** | 49 (THREE=46) |
| No `use cache` in data fetches | `rg -n "'use cache'\|cacheTag\|updateTag" apps/web/src/` | **Add per ADR (Next 16).** | **0 uses today** |
| `:latest` Docker tags | `rg -n ':latest' deploy/` | **Pin a version.** | 5 confirmed |
| Missing HEALTHCHECK on Go Dockerfiles | `grep -L HEALTHCHECK services/*/Dockerfile` | **Add `/readyz` probe.** | 11 services |
| `sleep N` for sync in scripts | `rg -n "^\s*sleep " Makefile deploy/` | **Bounded retry loop on `/readyz`.** | 3 confirmed |
| `log.Printf` / `fmt.Println` in prod services | `rg -n "log\.Printf\|fmt\.Println" services/ pkg/ -g '!*_test.go'` | **`slog.*Context`.** | 0 in services, 15+ in tools/cli (tolerable) |
| Hardcoded hex colors in TSX | `rg -n "#[0-9a-fA-F]{3,6}" apps/web/src/components/` | **Use Tailwind tokens.** | — |
| Two migration trees (ADR 022) | `ls db/platform/migrations/ db/tenant/migrations/` | **Merge into one.** | 2 trees exist |

The list grows. When you find a new recurring anti-pattern, add a row — the hunt
list is a living artifact.

## Session end protocol

A session ends only when the bar is met (see "No stopping rule"). When it ends,
you emit the session-end block **before** leaving the conversation. No trailing
summaries in chat replace this block — it's the artifact.

### Required block

```
Area:       pkg/jwt
Findings:   5  (2 shipped, 2 backlog, 1 dropped)
Shipped:
  - Removed dead Validator interface (−38 LOC, behavior unchanged; test preserved).
  - Collapsed Parse + ParseWithLeeway into one function (−20 LOC).
Backlog (next):
  - Index on tokens(user_id, jti) — needs a benchmark first.
  - Integration test for expiry edge case.
Dropped:
  - Rename parseClaims → decodeClaims (no readability gain).
Evidence:   go test ./pkg/jwt/... passed; −58 LOC net; no public API changed.
ADR delta: none (no architectural decision touched).
```

### What MUST land where

- **Commit message** — derived from `Shipped`. One commit per shipped item.
  The commit body quotes the evidence line ("go test ./pkg/jwt/... passed,
  −58 LOC net").
- **PR description** — the whole block, verbatim. Reviewers see intent + evidence.
- **ADR update** — only if `ADR delta` is not `none`. Use the `decisions` skill.
- **Next session target** — the `Backlog (next)` items are the candidates.
  Write them in the present imperative so the next session can act without
  re-reading context.

### What NEVER lingers

- No `TODO:` comments introduced by this session. If it's worth doing, it's on
  the Backlog; if not, it's Dropped. `TODO` in source is a third category that
  does not exist here.
- No "we'll clean this up later" anywhere. That IS the Backlog.
- No shipped item without evidence. Perf claims have a number; correctness
  claims have a test; simplicity claims have a LOC delta.
- No architectural change that doesn't update the relevant ADR in the same PR.

### When the bar isn't met

If nothing clears `impact/cost ≥ 3×` in the chosen area, the session still ends
with a block — shipped count 0, area moved to a different one explored that
same session, or blocker recorded with a reason specific enough that next
session can pick up without re-reading context.

A zero-shipped session is only valid after confirming the anti-pattern hunt
list is clear or blocked. Don't claim "nothing to do" without grep output.

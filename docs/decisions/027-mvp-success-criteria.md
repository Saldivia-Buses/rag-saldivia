# ADR 027 — MVP success criteria, phased

**Status:** accepted
**Date:** 2026-04-18
**Deciders:** Enzo Saldivia
**Refines:** ADR 026 (SDA replaces Histrix) — turns the four phases into
verifiable checks.
**Supersedes:** implicit "we'll know when we know" approach to readiness.

## Context

ADR 026 names the phases but not the gate. Every past session picked its
next task on vibes — sometimes the right vibes (fusion pilot), sometimes
not (pretty frontend while 200K invoice lines were missing). A concrete,
verifiable "is this phase done?" checklist removes the guesswork.

The checklist below is the **contract**. When a session ends, the shipped
work must be traceable to exactly one of these items. When a session
starts without a specific task, the operator (human or agent) reads this
list top-down, picks the first un-ticked item that is not blocked, and
works on it. Phase 0 always wins over any later phase that's not
blocked-on.

## Decision

Every item below is a binary check. Items can be waived only by a new ADR
that justifies the waiver. The waiver ADR number goes in the item.

### Phase 0 — Transversal (every session verifies these are still green)

- [ ] **Migration integrity**: for every row in `erp_migration_table_progress`
      for the latest prod run, `rows_read == rows_written + rows_skipped`.
      Today: fails (214K ghost rows).
- [ ] **Migrators with `rows_written=0` and `rows_read>0`**: zero.
      Today: fails (5 migrators, including FACDETAL → erp_invoice_lines =
      198K skipped).
- [ ] **Every migrated table has at least one sqlc query** (no dead-end
      writes). Today: fails (16 orphan tables per backend audit).
- [ ] **Every agent tool declares a capability** and is rejected at dispatch
      time when the user lacks the permission. Today: not implemented.
- [ ] **Workstation SHA == `main` HEAD**. A drift-detect script runs and
      passes. Today: fails (workstation runs pre-fusion 14 services).

### Phase 1 — Histrix parity + shutdown

#### Data migration

- [ ] Every legacy Histrix table in `.intranet-scrape/db-tables.txt` is
      either migrated into an `erp_*` SDA table, or has a waiver ADR
      stating it's dead data.
- [ ] Full migration run reports `status=completed` **with zero ghost
      rows** and zero `rows_written=0` migrators.
- [ ] Replay mode: a migrator run against current Histrix data produces
      no new SDA rows (idempotent quiescence).
- [ ] `erp_legacy_archive` contents are accessible via a read endpoint;
      any forensic check against archive data is a SQL query, not a JSONB
      dig.

#### UI parity

- [ ] Every XML-form in `.intranet-scrape/xml-forms/` has either:
      - (a) an `apps/web/src/app/**/page.tsx` equivalent reachable from
        the nav, OR
      - (b) a waiver entry (form X is dead / covered by form Y / out of
        scope) in `docs/parity/waivers.md`.
- [ ] Every page is reachable from a user's nav based on their role
      (no dead routes, no "admin only" screens the employee needs).
- [ ] Pagination, filtering, sort are wired for every page with more
      than ~50 rows in Histrix usage logs. Today: fails (hardcoded
      `page=1` across administración).

#### Operational parity

- [ ] Every critical Histrix report has an SDA equivalent (list in
      `docs/parity/reports.md`).
- [ ] All Histrix batch/cron jobs the company relies on have SDA
      equivalents or explicit retirement.
- [ ] Seamless-day test: one real employee runs their full work day in
      SDA without opening Histrix. Written record of what was missing.

#### Cutover readiness

- [ ] Cutover runbook in `docs/cutover/` with rollback plan.
- [ ] Shutdown date set and agreed.
- [ ] Histrix in read-only mode for N days before cutover; zero mismatch
      between Histrix and SDA reports during that window.

### Phase 2 — The SDA layer

#### Chat as UI

- [ ] Tool coverage: for the top-20 ERP write actions (identified from
      Histrix audit logs), the agent can execute each via chat and the
      resulting data matches what the UI would have produced.
- [ ] Permissions enforced: the agent refuses actions outside the user's
      role, tested via an integration test per role.
- [ ] Chat history survives context compression without losing the
      "who the user is" wiring.

#### Hierarchical prompts

- [ ] `services/app/internal/rag/agent/prompts/` contains:
      - `system.md` (the company overview, jargon, policies)
      - one `area.md` per area (`ingenieria.md`, `compras.md`, ...)
      - `user_template.md` (the editable starting point for per-user)
- [ ] Per-user prompts are stored in a DB table, editable via UI and by
      the user through chat commands.
- [ ] Every agent invocation assembles
      `system.md + area.md + user.md + recent_memories` before the user
      message, with explicit token budgets per layer.

#### Memories

- [ ] `erp_memories_global` and `erp_memories_user` (or equivalent)
      exist with time, source, user, area, content fields.
- [ ] Memory curator agent runs on a schedule; writes are idempotent
      (same conversation → same memories).
- [ ] Memories are consulted in the RAG retrieval step and ranked with
      the rest of the tree results.

#### Tree-RAG with ACL

- [ ] Collections are first-class: `erp_collections` (or equivalent)
      with (id, name, area, role_required).
- [ ] Every tree node belongs to exactly one collection.
- [ ] Retrieval filters collections by the user's role before scoring.
      Tested with a cross-area denial test.

### Phase 3 — Background agents + data hoarding

- [ ] Mail agent running as a durable NATS consumer; last-seen message
      ID persisted; replay after a 24h outage is clean.
- [ ] WhatsApp agent connected to the internal numbers; messages stored
      with sender, channel, timestamp, text, attachments.
- [ ] Memory curator agent running on a schedule; writes appear in the
      memories tables; no duplicates over re-runs.
- [ ] Analytics: at least one trained prediction endpoint live (e.g.
      stock below-min prediction); predictions tested against holdout
      data.

### Phase 4 — Differential UX

- [ ] Dashboard-personal page where the user adds, removes, and arranges
      widgets; state persists per-user.
- [ ] Widget library covers: KPI card, chart (line/bar), table with
      filter, chat shortcut, routine runner, tree explorer.
- [ ] Routine builder: user composes a named sequence of tool calls;
      executes by name from chat or UI.

## Consequences

**Positive**

- Sessions have a deterministic next-task picker: walk the list, take
  the first un-ticked item that is not blocked on an earlier phase.
- PR descriptions become honest: "ticks Phase 1 §UI parity item 2".
  Reviews gain signal.
- Disagreements about priority resolve against the list, not against
  personal preference.
- Misaligned sessions are blocked by the list itself: no one can ship
  a Phase 4 widget while Phase 0 integrity fails.

**Negative**

- The list is long. Deliberately so — the scope is honest.
- The list will grow as the .intranet-scrape audit surfaces more items.
  That's expected; each addition is an ADR or a line in a parity doc.
- "Waiver ADR" overhead is real. Intended: the friction of writing a
  waiver forces the thought.

**Neutral**

- This is an operational / process ADR, not architectural. It does not
  change any code directly; it changes which code gets written next.

## Alternatives considered

1. **Keep the list informal (issues or a Notion page).** Rejected.
   Anything off-repo desyncs within weeks. The harness (repo docs +
   skills + CLAUDE.md) is the single source of truth — ADR 027 belongs
   in that same place.

2. **Add phases to CLAUDE.md directly.** Rejected. CLAUDE.md is the
   constitution; ADR 027 is the current body of checks. Body can
   evolve (new items, waivers, strike-throughs) without touching
   the constitution on every tick.

3. **One giant checklist, no phases.** Rejected. Phases encode
   dependency: Phase 2 without Phase 1 is a toy. The ordering is
   itself a design decision.

## How to use this ADR

- **Session start**: read top-down, find the first un-ticked item whose
  dependencies (earlier phase) are ticked. That's your candidate.
- **Shipped work**: update this ADR in the same PR, ticking the item;
  no separate "docs" PR.
- **Waivers**: write a dedicated ADR (`0XX-waive-<form>.md` or similar)
  and reference it inline here. Waivers are rare and deliberate.
- **New items**: append to the right phase section. Keep items
  verifiable — if it's not binary-checkable, reshape until it is.

## Open items

- Populate `docs/parity/waivers.md` and `docs/parity/reports.md` (empty
  today).
- Define the drift-detect script (`make check-prod-drift` target?) for
  the Phase 0 "workstation SHA == main HEAD" check.
- Identify the top-20 ERP write actions from Histrix audit logs for the
  Phase 2 chat-coverage item.

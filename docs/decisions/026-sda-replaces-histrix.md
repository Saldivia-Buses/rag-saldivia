# ADR 026 — SDA replaces Histrix (no wrapper, no integration)

**Status:** accepted
**Date:** 2026-04-18
**Deciders:** Enzo Saldivia
**Refines:** project vision; corrects ambiguous wording in earlier memory.
**Relates to:** ADR 016 (tree-RAG), ADR 021 (reduce before add), ADR 022
(silo), ADR 025 (5 modules in 1 binary).

## Context

Earlier memory and commit messages described SDA as an "asistente empresarial
with RAG + agents and **integration with the legacy ERP (Histrix)**". That
framing was wrong and led past sessions to treat Histrix as a permanent
upstream to pull from — building readers, mappers, RAG layers *on top* of a
system that's assumed to keep running.

The actual product intent, surfaced in session 2026-04-18, is the opposite:
**SDA replaces Histrix.** Once SDA covers the capability surface, the Histrix
server gets powered off. Nobody at the company should notice except that the
new UI is better.

The scrape under `.intranet-scrape/` is the concrete yardstick:

- **676 tables** in the legacy MySQL (`db-tables.txt`).
- **434 XML-forms** (one form ≈ one CRUD screen) in `xml-forms/`.
- Full backend PHP + frontend JS + menu maps captured.

That's the "parity surface" SDA must cover before the switch.

## Decision

**SDA is a full replacement of Histrix, plus a new capability layer Histrix
could never provide.** The product is structured in four sequential phases,
gated by a transversal phase 0.

### Phase 0 — Transversal (always, blocks everything else)

- **Data integrity = religion.** Zero row loss in migration.
  `rows_read == rows_written + rows_skipped` for every migrator. Ghost
  rows are blockers, not footnotes.
- **Tool security from day one.** Every agent tool declares capabilities and
  is checked against the user's permissions before execution. Retrofit is
  not allowed; security is primitive, not feature.
- **Prod = source of truth.** If the workstation runs a different SHA than
  `main`, closing that gap preempts new work.

### Phase 1 — Histrix parity + shutdown

The gating phase. SDA does not exist as a product until this ships.

- Every one of the 434 XML-forms has an SDA equivalent, or an explicit waiver
  (recorded per-form as an ADR or a line in a scoped document).
- Zero ghost rows in `erp_migration_table_progress`.
- Seamless cutover: a real employee can do a full work day in SDA and never
  open Histrix.
- Critical Histrix reports are reproducible in SDA.
- Final migration runs in replay mode (HTX canonical, SDA mirrored) until
  the shutdown window, then SDA becomes canonical.

### Phase 2 — The SDA layer (what Histrix never was)

- **Chat as first-class UI.** The agent is the user's representative:
  "cargame esta factura" triggers the same write paths the UI would trigger,
  scoped to the user's permissions. Chat ↔ UI capability parity.
- **Hierarchical prompts.** `system.md` (company-wide: jargon, org shape,
  policies) → `area.md` (per-area context) → `user.md` (editable by the
  employee) → memories (global + per-user, curated by an agent).
- **Tree-RAG with per-collection ACL.** Engineering does not see purchasing
  trees unless they have the role. Trees are grouped by area; collections
  are the unit of access control.
- **Granular tool permission model.** Each tool declares a capability; the
  agent runtime enforces the user's permission set before dispatch. Refines
  ADR 014 (JWT identity) and the auth-security skill's RBAC.

### Phase 3 — Background agents + data hoarding

Data is the strategic asset; we capture everything.

- **Mail agent:** parses company mail, indexes by date and thread, extracts
  entities and attachments, feeds them into the RAG tree.
- **WhatsApp agent:** ingests the company's internal WhatsApp numbers.
- **Memory curator agent:** decides what survives long-term (global + per-user
  memories). Runs in the background, not on every message.
- **Analytics + prediction:** models trained on the accumulated data to
  forecast stock, demand, cash flow, etc.

### Phase 4 — Differential UX

- **Personal dashboards.** Every employee builds their own. There is **no**
  global dashboard. Widgets are composable; persistence is per-user.
- **Personal routines.** A recipe of actions the user runs frequently,
  triggerable by one tap or one chat line.

## Consequences

**Positive**

- The north star becomes verifiable. "Is SDA done?" decomposes into Phase 1
  gate checks, Phase 2 capability checks, etc. (see ADR 027).
- Misaligned work stops sooner. A session that builds a pretty `/inicio`
  dashboard while `erp_invoice_lines` is 0 rows is wrong under this ADR; it
  fails the Phase 0 / Phase 1 gate before it ships.
- Histrix-as-upstream assumptions can be retired. Code paths that read live
  from Histrix are migration-tooling only, not product surface.
- `.intranet-scrape` becomes a first-class artifact: it is the parity
  contract, consulted before any "new ERP feature" is designed.

**Negative**

- The explicit scope (434 forms + 676 tables) is sobering. It forces honest
  scheduling instead of the earlier "consolidate the pkg/ dir" busywork.
- No shortcut path exists for Phase 2+ that skips parity. Product value
  lives on top of a correctly migrated base.
- Several earlier memories and CLAUDE.md lines used integration wording;
  they are corrected by this ADR and become superseded context.

**Neutral**

- ADR 022 (silo), ADR 023 (one container), ADR 024 (frontend in container),
  ADR 025 (5 internal modules) all remain valid — this is a product
  direction ADR, they are architecture.
- ADR 016 (tree-RAG no vectors) stands, but per-collection ACL is a
  refinement surfaced here; it will get its own ADR when the implementation
  lands.

## Alternatives considered

1. **Keep Histrix alive, SDA is a sidecar.** Rejected. That's the earlier
   implicit model, and it traps SDA in "integration" scope forever: every
   new capability has to round-trip through the legacy system. Also doubles
   the ops cost permanently.

2. **Freeze Histrix now, migrate piecemeal over years.** Rejected. Without
   a shutdown date, parity never gets prioritized and SDA stays "pretty,
   but not enough". A defined cutover forces Phase 1 discipline.

3. **Build SDA as a totally new ERP, walk away from the Histrix data.**
   Rejected. The company's operational history *is* in those 22M+ rows —
   throwing them away is not an option. Zero-loss migration is a
   hard constraint.

## Open items

- Which XML-forms are genuine replace vs "we stopped using that" — needs
  a sweep of `.intranet-scrape/xml-forms/` against actual Histrix usage
  logs. Output becomes a waiver list referenced by ADR 027's checklist.
- Concrete shutdown criteria and date — depend on Phase 1 progress. Track
  in ADR 027.
- Prompt-layers, agent-tools, background-agents architectures each need
  their own ADR when the first implementation session touches them.

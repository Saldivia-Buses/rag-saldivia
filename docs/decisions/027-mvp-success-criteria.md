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

**Owners:** `migration-health` (integrity + orphan tables), `agent-tools` (tool perm), `deploy-ops` (prod drift).

- [x] **Migration integrity**: for every row in `erp_migration_table_progress`
      for the latest prod run,
      `rows_read == rows_written + rows_skipped + rows_duplicate`.
      Shipped 2026-04-18 (2.0.6 commits `fee8ab14` + `dad0e7a6`): `rows_duplicate`
      counter on `erp_migration_table_progress` + invariant reporter in
      `printReport`. Empirical check — new run `0eb3af71` on saldivia_bench:
      22 tables completed, canonical ghost-row query returns zero, invariant
      holds table-by-table.
- [x] **Migrators with `rows_written=0` and `rows_read>0`**: zero except waivers.
      Shipped 2026-04-18 (2.0.6 commits `8522fb15` + `8ea0a790`): PreloadDomain
      inside every `Build*Index`, `BuildRegMovimIndex` moved from `AddSetupHook`
      to `AddAfterTableHook("IVACOMPRAS")`, and FICHADADIA's clock_in/out now
      encode as `*time.Time` (not `*string`) for the TIMESTAMPTZ target. Empirical
      recovery on run `0eb3af71`: FACDETAL 191,051 rows, FICHADADIA 932,665,
      RHDESCUENTOS 16,809, RRHH_ADICIONALES 1,902 — **1,142,427 rows** previously
      silently dropped. One no-op remains: REMDETAL (5K), waived as `W-001` in
      `docs/parity/waivers.md` pending a REMITOINT migrator.
- [x] **Every migrated table has at least one sqlc query** (no dead-end
      writes). Shipped 2026-04-18 (2.0.6 commits to follow): added 5 read
      queries for populated orphans (`erp_bom_history` 14.7M rows,
      `erp_legacy_archive` 7.5M, `erp_training_attendees` 5.9K,
      `erp_inspection_templates` 2K, `erp_medical_visits_log` 59). The
      remaining 9 orphans are waived — 4 are CLI-internal bookkeeping
      (`erp_legacy_mapping` + the three `erp_migration_*`, waiver `W-002`)
      and 5 are empty Phase 1 UI placeholders (`erp_article_photos`,
      `erp_communication_recipients`, `erp_sequences`, `erp_survey_questions`,
      `erp_unit_photos`, waiver `W-003`). Canonical check now runs green
      against `docs/parity/waivers.md` entries.
- [ ] **Every agent tool declares a capability** and is rejected at dispatch
      time when the user lacks the permission. Today: not implemented.
- [x] **Workstation SHA == `main` HEAD**. A drift-detect script runs and
      passes. Shipped 2026-04-18 (2.0.7): `make check-prod-drift` target
      compares `/opt/saldivia/repo` HEAD vs `origin/main` and checks that
      running containers' compose source lives under `/opt/saldivia/`.
      Workstation moved off the stale GitHub-runner workspace onto the
      canonical `/opt/saldivia/deploy/` compose and the post-fusion shape
      (`{erp, web, postgres, redis, nats, traefik, mailpit, minio}`); the
      pre-fusion 14-service stack is gone. The monolith `app` service is
      not in compose yet — that wiring is itself the next Phase 0/ADR 025
      work item, not part of this tick. Pre-transition backup:
      `/var/backups/saldivia/postgres/pre-deploy-transition-20260418_045110.sql.gz`
      (2.7 GB) + orphaned `deploy_postgres_platform_data` volume archived to
      `/var/backups/saldivia/archives/deploy_postgres_platform_data_orphaned-20260418_045507.tar.gz`.

### Phase 1 — Histrix parity + shutdown

**Sub-section order inside Phase 1** (pick items top-down, same rule as phases): Data migration → UI parity → Operational parity → Cutover readiness. UI parity work stalls if Data migration has an open ghost-row or missing-table item.

#### Data migration

**Owners:** `migration-health` (integrity, replay), `htx-parity` (table coverage), `backend-go` (archive read endpoint).

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

**Owners:** `htx-parity` (form coverage, waivers), `frontend-next` (page implementation, nav, pagination).

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

**Owners:** `htx-parity` (reports + seamless-day), `backend-go` (batch/cron equivalents).

- [ ] Every critical Histrix report has an SDA equivalent (list in
      `docs/parity/reports.md`).
- [ ] All Histrix batch/cron jobs the company relies on have SDA
      equivalents or explicit retirement.
- [ ] Seamless-day test: one real employee runs their full work day in
      SDA without opening Histrix. Written record of what was missing.

#### Cutover readiness

**Owners:** `deploy-ops` (runbook, rollback, shutdown date), `migration-health` (read-only/replay quiescence).

- [ ] Cutover runbook in `docs/cutover/` with rollback plan.
- [ ] Shutdown date set and agreed.
- [ ] Histrix in read-only mode for N days before cutover; zero mismatch
      between Histrix and SDA reports during that window.

### Phase 2 — The SDA layer

#### Chat as UI

**Owners:** `agent-tools` (tool registry, perm tests), `rag-pipeline` (dispatcher + streaming), `auth-security` (capability model).

- [ ] Tool coverage: for the top-20 ERP write actions (identified from
      Histrix audit logs), the agent can execute each via chat and the
      resulting data matches what the UI would have produced.
- [ ] Permissions enforced: the agent refuses actions outside the user's
      role, tested via an integration test per role.
- [ ] Chat history survives context compression without losing the
      "who the user is" wiring.

#### Hierarchical prompts

**Owners:** `prompt-layers` (layer design, budgets, assembly, logging), `database` (per-user prompt table).

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

**Owners:** `prompt-layers` (retrieval + assembly side, schema requirements — see that skill for the full column list), `background-agents` (memory curator), `database` (migrations).

- [ ] `erp_memories_global` and `erp_memories_user` (or equivalent)
      exist with time, source, user, area, content fields.
- [ ] Memory curator agent runs on a schedule; writes are idempotent
      (same conversation → same memories).
- [ ] Memories are consulted in the RAG retrieval step and ranked with
      the rest of the tree results.

#### Tree-RAG with ACL

**Owners:** `rag-pipeline` (collection filter in retrieval), `auth-security` (role → collection mapping), `database` (collection table).

- [ ] Collections are first-class: `erp_collections` (or equivalent)
      with (id, name, area, role_required).
- [ ] Every tree node belongs to exactly one collection.
- [ ] Retrieval filters collections by the user's role before scoring.
      Tested with a cross-area denial test.

### Phase 3 — Background agents + data hoarding

**Owners:** `background-agents` (every item). Sub-owners: `prompt-layers` (memory curator output side), `rag-pipeline` (mail/WhatsApp tree integration), `database` (new tables).

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

**Owners:** `frontend-next` (dashboard + widgets + routine UI), `backend-go` (persistence + routine-run API), `agent-tools` (routines that execute tool sequences).

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
- Identify the top-20 ERP write actions from Histrix audit logs for the
  Phase 2 chat-coverage item.
- Harden prod deploy leftovers the app-in-compose work surfaced:
  (a) ~~redis auth~~ — shipped in 2.0.7: `config.EnvOrFile` reads
  `REDIS_PASSWORD` / `REDIS_PASSWORD_FILE`, `security.InitBlacklist` takes
  `*redis.Options`, both `app` and `erp` pass Addr + Password to redis client
  + blacklist, prod compose mounts `redis_password` secret on both services.
  (b) prod compose `healthcheck:` for `app` and `erp` uses `wget` but the runtime
  base is `distroless/static-debian12` (no shell, no wget) — the healthcheck
  never actually runs. Either add a `--healthcheck` self-probe to the Go
  binaries or ship a tiny static probe alongside.
  (c) `db_tenant_template_url` secret is mounted for erp and meant for app,
  but both services read `POSTGRES_TENANT_URL` env directly — the secret
  is dead weight unless plumbed through an init step.

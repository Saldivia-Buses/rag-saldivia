# Histrix parity — waivers

Each entry is either (a) a Histrix XML-form that does NOT get an SDA
equivalent (with justification) or (b) a legacy table that does NOT get a
migrator (dead data, unrecoverable, superseded by a different feature). ADR
027 Phase 1 UI-parity item cites this file; the data-migration item does
too when a specific legacy table is knowingly skipped.

Format per waiver:

- **Scope**: which form / table / migrator the waiver covers.
- **Why**: the concrete reason (not "out of scope" — always a specific fact).
- **Blast radius**: the rows / users / workflows affected and whether anything
  downstream depends on this.
- **Revisit**: the condition that would reopen the waiver.

---

## W-002 — CLI-internal migration tables (no app-layer query)

**Scope**: `erp_legacy_mapping`, `erp_migration_runs`,
`erp_migration_table_progress`, `erp_migration_validation_issues`.

**Why**: These four tables are bookkeeping for the `sda migrate-legacy` CLI
(see `tools/cli/internal/migration/`). They are written by the CLI, read by
the CLI (via `Mapper.Resolve*`, `Orchestrator.Resume`, pre-migration
validators). No app service needs them — the "sqlc query" check assumes
app-facing reads, which is the wrong shape for these.

**Blast radius**: zero — the data is correctly read by the tool that
owns it. If an admin dashboard wants to surface "last migration run
status" it can lift the existing `erp_migration_table_progress` shape
into a new sqlc query; until then, `docker exec ... psql` is fine.

**Revisit**: if an admin UI for migration status enters scope.

## W-003 — Phase 1 UI placeholder tables (empty, awaiting write + read)

**Scope**: `erp_article_photos`, `erp_communication_recipients`,
`erp_sequences`, `erp_survey_questions`, `erp_unit_photos`.

**Why**: These tables exist in the tenant schema but have zero rows on
the prod saldivia DB — no migrator writes them yet and no app handler
reads them. The Phase 0 "no dead-end writes" check is satisfied by the
absence of writes; queries become relevant in Phase 1 when the paired
UI (stock article sheet, communications flow, auto-number sequences,
survey engine, unit-card gallery) lands.

**Blast radius**: zero — empty tables can't silently drop data.

**Revisit**: when the matching Phase 1 XML-form area (see
`.intranet-scrape/xml-forms/`) is implemented:
- `articulos/` → erp_article_photos
- `comunicaciones/` → erp_communication_recipients
- admin UI for sequence counters → erp_sequences
- `calidad/` surveys → erp_survey_questions
- `produccion/` unit card → erp_unit_photos

The paired PR adds the read queries (and write queries if the UI edits
the rows), and strikes the row from this waiver.

## W-001 — REMDETAL (`erp_invoice_lines`) silent drop on fresh migration

**Scope**: the `NewDeliveryNoteLineMigrator` in
`tools/cli/internal/migration/migrators.go`. 5,125 legacy rows.

**Why**: REMDETAL's `idRemito` column is documented in code as "FK to REMITO"
but the MySQL schema (`.intranet-scrape/db-schema.sql:15503-15510`) shows
REMITO has PK `(numero, puesto)` with no `idRemito` column at all. The
actual parent is `REMITOINT` (the "remitos internos" table at line 15520,
PK `idRemito`), which today has no migrator. `BuildRemitoIndex` correctly
detects REMITO lacks idRemito and no-ops — but then every REMDETAL row
fails to resolve its parent and is skipped. The 5K rows go through the
archive-skips path (when `--archive-skips` is on) or are dropped otherwise.

The operational context is that REMDETAL lines cover internal workshop
material issuance (piezas entregadas al taller). They are referenced by
warehouse ops for reconciling depot outflows against production consumption.

**Blast radius**: 5,125 historical delivery-note lines. No live feature in
SDA reads them yet (the warehouse reconciliation UI is Phase 1 work that
hasn't landed). The archive keeps the raw data for forensic queries.

**Revisit**: when the warehouse reconciliation page enters Phase 1 scope
(`stock` domain, `almacen/*` XML-forms). The fix is to add a
`NewInternalDeliveryNoteMigrator` for REMITOINT, wire a
`BuildRemitoIntIndex`, and repoint `NewDeliveryNoteLineMigrator` at the
new index. Tracked in ADR 027 as a Phase 1 data-migration follow-up.

## W-004 — Histrix intranet infrastructure tables (HTX*)

**Scope**: 31 tables owned by the Histrix intranet platform itself
(menus, mail, calendar, chat, prefs, media library, auth plumbing,
logging, record-level ACLs, etc.) — not business data.

```
HTXACCESSLOG, HTXADDRESSBOOK, HTXCALENDAR, HTXCHAT, HTXFONTS, HTXLOG,
HTXMAIL, HTXMENU, HTXMESSAGES, HTXNEWS, HTXNOTIFUSER, HTXOPTIONS,
HTXPREFS, HTXPRINTERS, HTXPROFILEDIR, HTXTHEME, HTX_ACCESS_TOKEN,
HTX_ATTRIBUTES, HTX_ATTRIBUTE_OPTIONS, HTX_CRONTAB, HTX_EMPRESAS,
HTX_MEDIA, HTX_NOTIFICATION_TOKEN, HTX_OPENIDS, HTX_RECORD_AUTH,
HTX_SUBSYSTEM_AUTH, HTX_TABLE_ATTRIBUTE, HTX_TABLE_TAGS, HTX_TAGS,
HTX_URLS, HTX_USER_PROFILE
```

**Why**: These back Histrix-as-a-CMS, not Saldivia-as-a-company. SDA
replaces the surface they power:

- `HTXMENU`, `HTXPROFILEDIR`, `HTX_SUBSYSTEM_AUTH`, `HTX_RECORD_AUTH`,
  `HTX_URLS`, `HTX_ACCESS_TOKEN`, `HTX_OPENIDS`, `HTX_USER_PROFILE`,
  `HTXPREFS`, `HTXOPTIONS`, `HTXTHEME`, `HTXFONTS`, `HTXPRINTERS` →
  SDA's own auth, routing, RBAC, per-user preferences.
- `HTXMAIL`, `HTXADDRESSBOOK`, `HTXNEWS`, `HTXNOTIFUSER`, `HTX_MEDIA`,
  `HTXNOTIFUSER`, `HTX_NOTIFICATION_TOKEN` → replaced by
  `erp_communications` (migration 029) + the Phase 3 mail / WhatsApp
  ingest agents (ADR 027 §Phase 3).
- `HTXCHAT`, `HTXMESSAGES` → replaced by SDA chat sessions
  (migration 003 chat tables) and the Phase 2 agent.
- `HTXCALENDAR` → replaced by `erp_communications` events
  (migration 029 is the successor; calendar-event UX is Phase 1 admin).
- `HTXACCESSLOG`, `HTXLOG` → replaced by `audit_log` (migration 001) +
  structured slog streamed to OTel.
- `HTX_ATTRIBUTES`, `HTX_ATTRIBUTE_OPTIONS`, `HTX_TABLE_ATTRIBUTE`,
  `HTX_TABLE_TAGS`, `HTX_TAGS`, `HTX_CRONTAB`, `HTX_EMPRESAS` →
  platform-generic machinery; SDA is single-tenant (silo per ADR 022),
  schedules live in NATS consumers (per ADR 027 §Phase 3 background
  agents), tags are a RAG concern (Phase 2).

**Blast radius**: zero. None of these hold Saldivia business state.
Historical `HTXACCESSLOG` / `HTXLOG` rows are interesting only for the
retrospective "who touched what in Histrix on day X" question; that
stays on the Histrix box for as long as it runs and is orthogonal to
cutover readiness.

**Revisit**: never — these are cutover casualties by design. If a
forensic need surfaces for historical Histrix access logs, that's a
one-shot dump into `erp_legacy_archive`, not a new migrator.

## W-005 — `*_OLD` superseded Histrix tables

**Scope**: 5 schema-reboot leftovers inside Histrix itself:
`CPSENCAB_OLD`, `CPSMOVIM_OLD`, `CTBCONCE_OLD`, `CTBCUENT_OLD`,
`CTBPARAM_OLD`.

**Why**: The `*_OLD` suffix is Histrix's own convention for "we rebooted
this table — use the unsuffixed version". The current, in-use tables
(`CPSENCAB`, `CPSMOVIM`, `CTBCONCE`, `CTBCUENT`, `CTBPARAM`) appear in
the parity list separately — covering those (or waiving them) is the
real work. Migrating the `*_OLD` rows would import abandoned historical
snapshots that Histrix already decided were wrong.

**Blast radius**: zero for the live business. Rows in the `*_OLD`
tables are not referenced by any Histrix form or handler today;
retaining them would pull dead-weight into the SDA schema and inflate
Phase 0 integrity checks for no benefit.

**Revisit**: if a forensic need emerges, dump the `*_OLD` rows into
`erp_legacy_archive` and query there — same shape as the HTXLOG
archive path in W-004.

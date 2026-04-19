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

## ~~W-001~~ — REMDETAL (`erp_invoice_lines`) silent drop on fresh migration

**Status**: ✅ **Closed 2026-04-18** (2.0.8). Resolution shipped the exact
path the waiver documented: `NewInternalDeliveryNoteMigrator` writes
REMITOINT → `erp_invoices` (`tools/cli/internal/migration/migrators.go`),
`BuildRemitoIntIndex` populates the REMITOINT.idRemito → UUID index
(`tools/cli/internal/migration/mapper.go`), and
`NewDeliveryNoteLineMigrator` now resolves REMDETAL parents via
`ResolveByRemitoIntID` (with `ResolveByRemitoID` as defensive fallback
for snapshots where REMITO does carry idRemito).

**Evidence** (direct SQL against live Histrix, 2026-04-18):
- REMITO_has_idRemito = 0 → original root cause confirmed.
- REMITOINT = 746 rows → new parent migrator target.
- REMDETAL = 5,125 rows, every single one joins cleanly against
  REMITOINT.idRemito (orphans_from_REMITOINT = 0).
- Post-fix Phase 0 invariant: `rows_read = rows_written + rows_skipped
  + rows_duplicate` holds with 5,125 = 5,125 + 0 + 0 on a fresh run.

**Original scope (kept for history)**: the `NewDeliveryNoteLineMigrator`
in `tools/cli/internal/migration/migrators.go`. 5,125 legacy rows.

**Why** (original): REMDETAL's `idRemito` column is documented in code
as "FK to REMITO" but the MySQL schema
(`.intranet-scrape/db-schema.sql:15503-15510`) shows REMITO has PK
`(numero, puesto)` with no `idRemito` column at all. The actual parent
is `REMITOINT` (the "remitos internos" table at line 15520, PK
`idRemito`), which had no migrator until this fix. `BuildRemitoIndex`
correctly detected REMITO lacks idRemito and no-opped — but then every
REMDETAL row failed to resolve its parent and was skipped. The 5K rows
went through the archive-skips path (when `--archive-skips` was on) or
were dropped otherwise.

The operational context is that REMDETAL lines cover internal workshop
material issuance (piezas entregadas al taller). They are referenced by
warehouse ops for reconciling depot outflows against production consumption.

**Blast radius** (original): 5,125 historical delivery-note lines + 746
parent REMITOINT rows now both land on a fresh migration. The Phase 1
warehouse reconciliation UI (`stock` domain, `almacen/*` XML-forms) gets
a full dataset to read from when it lands.

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

## W-006 — zero-row uncovered Histrix tables (bulk waiver)

**Scope**: 225 tables from `.intranet-scrape/db-tables.txt` that have
no migrator/reader registered AND return `table_rows = 0` from
`information_schema.tables` on the live Histrix DB (`saldivia` schema
at `172.22.100.99`, query run 2026-04-18).

The list is pinned in `docs/parity/data-migration.md` (top-of-repo
reproducer regenerates it against the live DB).

By prefix family, the largest chunks are:

- 22× `STK_*` — abandoned stock subsystems (inspección detail,
  sub-rubros, parámetros, insumos).
- 17× `REG_*` — entity-extension tables never populated
  (certificados, comisiones, contactos grupos, categorías).
- 11× `GEN_*` — secondary catalogs never seeded beyond the 17 already
  migrated.
- 11× `CAJ_*` — treasury parametrization tables (numeradores, formas,
  conceptos aux) never used on this tenant.
- 9× `MANT_*` — maintenance subsystem tables (talleres, ajustes, tipos
  equipo) never populated.
- 8× `RH_*` — HR extended (cursos plan, docentes) never used.
- 7× `oauth_*` — Sabre OAuth server tables (library leftover — never
  configured).
- 8× calendar/cards tables (`calendars`, `calendarobjects`,
  `calendarsubscriptions`, `calendarchanges`, `addressbooks`,
  `addressbookchanges`, `cards`, `principals`) — Sabre CalDAV/CardDAV
  library leftover.
- 4× `HtxMailgun*` + assorted lowercase tables (`users`, `groupmembers`,
  `locks`, `tmp_mac`, `t_d_SwipeRecord`, `replicasql`) — third-party
  integration leftovers that Histrix ships with but Saldivia never
  activated.
- Remaining 130+ — individual dead tables (`VUNIDADES`, `VSTOCK`,
  various `TIPO*` and unsuffixed legacy variants like `CTBCONCE`,
  `CTBCUENT`, `CPSMOVIM`, `CARCHE`, `BCSMOVIM`, `CCTMOVIM` —
  confirming these are the dead-side of the schema reboot captured by
  W-005 at finer granularity).

**Why**: table_rows = 0 from MySQL's `information_schema` means the
table is either truly empty or has never had a row (the statistic is
updated on every write). A 0-row table has no data to migrate. If a
table was intended to hold data but never got any, that's the feature
never being used — the parity contract is "don't lose data", not "don't
lose feature surface that no one ever used".

**Blast radius**: zero row data to lose. Zero risk of corrupting an SDA
table that a later feature might need: the table shape is still
discoverable in `.intranet-scrape/db-schema.sql` if a future Phase 1
feature resurrects one of these subsystems — at which point the cost is
writing a migrator plus seed data, not data recovery.

**Revisit**: if a later Phase 1 session needs one of these subsystems
live (e.g. someone decides to activate Sabre calendar, or enable the
fleet GPS tables), strike that entry from W-006 in the same PR and
write the migrator + seed. Default answer for a PR that touches a
W-006 table: either bring rows along with a migrator, or keep it waived.

## W-007 — sub-15 K-row long tail (bulk waiver)

**Scope**: every table in `.intranet-scrape/db-tables.txt` with
`table_rows < 15000` from `information_schema.tables` on the live
Histrix DB (`saldivia` schema at `172.22.100.99`, query run
2026-04-19) that is NOT already covered by:

  (a) one of the waivers above (W-004 HTX infra, W-005 `*_OLD`,
      W-006 zero-row);
  (b) a registered migrator / reader under
      `tools/cli/internal/migration/` or `tools/cli/internal/legacy/`;
  (c) the Pareto-tail migrators landing in 2.0.11 (migrations `077`,
      `078`, `079`).

The exact list is regenerable from the reproducer at the bottom of
`docs/parity/data-migration.md` — it naturally shrinks as later
sessions promote individual tables out of the tail.

Rough shape (post-2.0.11): **~285 tables, ≤ 200 K live rows
combined**. Each individual table contributes less than 0.1 % of the
remaining Histrix row volume. Mean live count per waived table is
sub-500.

**Why**: row volume in the remaining uncovered gap is **extremely**
long-tail. Post-Grupo A (237 K), Grupo B (176 K), and W-008 industrial
telemetry (~130 K), the residue is ≤ 200 K rows spread across ~285
tables. Writing a bespoke migrator per table would take ~50+ sessions
at the current shape-audit + reader + migrator + sqlc pattern.
Collectively these tables are <1 % of the Histrix row budget; the
ROI on individual migrators is negative relative to shipping the
Phase 1 UI parity surfaces that actually consume the already-migrated
data.

**Blast radius**: the data is preserved in the read-only Histrix DB
— no deletion, no drift. If a Phase 1 UI surface resurrects one of
the waived tables, the escape hatch is "strike the table from W-007
and write its migrator in the same PR". The contract stays: default
is waive, specific-form ask lifts it.

**Revisit**: on every new Phase 1 UI PR, check the waived-table list
for names the form reads. The default — "if the form reads a W-007
table, close the waiver for that table before landing the form" —
is enforced by the `htx-parity` skill at form time.

## W-008 — industrial & telephony telemetry (EGX300EPE, EGX_300, TEL_LOG)

**Scope**: three legacy Histrix tables that back *monitoring*
dashboards, not operational business flows:

- `EGX300EPE` (79,376 rows live) — Schneider PowerLogic EGX300
  per-panel electrical readings (potencia activa / reactiva / aparente,
  demanda, intensidad por fase). Histrix form:
  `xml-forms/it/egx300epe.xml` ("Tableros electricos"), view-only
  consulta paginar=300.
- `EGX_300` (15,992 rows live) — older-shape EGX300 historic
  readings (3-phase intensity only, no PK). Histrix forms:
  `xml-forms/it/egx300.xml`, `egx_300_xfase.xml`, `egx_300_naves.xml`,
  all view-only.
- `TEL_LOG` (34,885 rows live) — Asterisk / VoIP call detail
  records (fecha, extension, numero, duracion). Histrix forms:
  `xml-forms/tel/tel_log_qry.xml` ("CONSULTA DE LLAMADAS",
  autoupdate=180s) and `tel/tel_promedio_llamadas_qry.xml`.

Combined: ~130 K rows across three tables that feed three IT /
telephony dashboards under `xml-forms/it/` and `xml-forms/tel/`.

**Why**: these are sensor / log streams from equipment outside the
ERP, not data the ERP owns. EGX300 rows come from a physical
Schneider power meter feeding a CSV ingester; TEL_LOG is the Asterisk
call log. Keeping them inside the SDA tenant DB as static rows mis-
represents the architecture — the correct Phase 3 answer is a
proper time-series / log sink (InfluxDB / Loki / similar) fed by the
background agents that ADR 027 Phase 3 covers, not a PostgreSQL
`erp_power_meter_readings` that silently freezes in time at cutover.

The Histrix consulta forms are viewer-only and cover the past — not
decision-driving, not write-back paths. Employees who need real-time
monitoring already use dedicated tools (Grafana, Schneider's own
front-end); the Histrix views were a convenience surface, not a
system of record.

**Blast radius**: zero. The raw rows remain in the read-only Histrix
DB. A future Phase 3 background-agent for industrial telemetry can
re-read them directly from MySQL and push into the right sink.

**Revisit**: when a Phase 3 industrial-monitoring agent enters scope
and we pick the time-series sink. At that point the correct move is
to write the ingester (read EGX/TEL_LOG every N seconds → sink), not
to migrate the historic rows into Postgres.

## W-009 — bcs_importacion_auto_ins bulk CSV/XLS ingest (§UI)

**Scope**: `bancos_local/bcsmovim_importacion_auto_ins.xml` — the
Histrix form that accepts a bank extract file (CSV / XLS) and bulk-
inserts one row per line into `BCS_IMPORTACION` (now
`erp_bank_imports`). Paired read-only qry + per-row toggle forms
(`bcs_importacion_qry.xml` and `bcsmovim_importacion_auto_mov_ins.xml`)
are **covered** by `/administracion/tesoreria/importaciones` in 2.0.13
— see `docs/parity/ui-parity.md`.

**Why**: SDA does not yet have a file-upload + server-side parser
pipeline. Histrix's implementation expects a CSV or proprietary XLS
layout per bank (format varies — Supervielle, Santander, Galicia
layouts all differ) and runs a PHP parser to map columns onto
`BCS_IMPORTACION` fields. Reproducing that would require: (a)
file-upload infra (MinIO bucket + pre-signed URLs, or direct
multipart/form-data into the Go service); (b) a per-bank format
registry with parsers; (c) error-row surfacing UI. That's a dedicated
cluster, not part of the XML-form replacement for the read/toggle
surface.

Meanwhile, the staging table (`erp_bank_imports`, 91,959 rows live) is
already populated through the existing Histrix ingester — which still
runs against the same `saldivia` DB, feeding rows that SDA now reads
and toggles via the covered forms.

**Blast radius**: zero for operators during Phase 1 — the existing
Histrix ingester keeps running, and the SDA page consumes its output.
No data loss: the rows still land in MySQL first and are mirrored into
`erp_bank_imports` by the migrator on each sync.

**Revisit**: when a file-ingest cluster enters scope (likely alongside
Phase 3 mail / document-ingestion agents, or an explicit "bank-import
UI" Phase 1 session). At that point the right move is to stand up the
upload surface + parser registry once and cover every per-bank layout,
not to port Histrix's one-off PHP handler. Strike this waiver in the
same PR that ships the first bank parser.

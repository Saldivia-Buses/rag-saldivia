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

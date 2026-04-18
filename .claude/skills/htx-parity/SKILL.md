---
name: htx-parity
description: Use when planning or implementing a Phase 1 capability — anything that replicates what Histrix does. The contract lives in .intranet-scrape/ (676 tables, ~4,500 XML-forms across 99 area groups). Every "new ERP feature" starts by reading the relevant XML-form first; every ERP feature ships with a tick on the ADR 027 Phase 1 UI-parity checklist.
---

# htx-parity

Scope: Phase 1 of ADR 027 (Histrix parity + shutdown). The parity
contract is `.intranet-scrape/`, committed at the repo root. Every ERP
capability in SDA must have a traceable line back to a Histrix artifact
in that directory, or an explicit waiver in `docs/parity/waivers.md`.

## What's in `.intranet-scrape/`

| Artifact | What it tells you |
|---|---|
| `db-schema.sql` (1 MB) | Full Histrix MySQL schema — DDL for all 676 tables |
| `db-tables.txt` | Bare list of 676 table names (easier to grep) |
| `xml-forms/` | **99 form-groups + 335 top-level forms = 434 top-level entries; ~4,500 XML files total** — each form ≈ one Histrix screen. The visual + logical contract of every CRUD surface |
| `php/` | Backend PHP. Read it to understand data flow / validations |
| `js/` | Frontend JS. Read it to understand client-side behavior |
| `intranet-all-menu-items.json` | The menu tree — what's in the nav, grouped by area |
| `intranet-all-xmls.json` | Index of every XML-form with metadata |
| `intranet-full-structure.json` | Page graph — what links to what |
| `intranet-menu-handlers.json` | Menu → form handler mapping |
| `intranet-principal-full.html` | Landing page HTML (navigation reference) |
| `intranet-xml-urls.json` | Full URL inventory — where each form was accessible |
| `intranet-xml-test-fetch.json` | Fetch test results (confirms forms were live at scrape time) |

## The parity methodology

When you start a Phase 1 capability, follow this loop:

### 1. Find the Histrix artifact

- Start from the user's verbal description ("cargar factura").
- Search `intranet-all-menu-items.json` for the matching menu entry
  (text match, area).
- That gives you the handler name, which points to one or more XML-forms
  under `xml-forms/<area>/<form>.xml`.
- `grep`-friendly entry: `rg -l "facturacion" .intranet-scrape/xml-forms/`.

### 2. Read the XML-form

A Histrix XML-form describes:

- **Fields** — column list, types, validations, defaults.
- **Actions** — what buttons exist, what each does (INSERT, UPDATE,
  DELETE, stored procedure call, a custom PHP handler).
- **Queries** — SELECTs that populate dropdowns and grids.
- **Permissions** — role gates inlined into the form.

Read the form end-to-end before writing a single line of SDA code.
Note:

- The source tables it writes to (match against SDA's `erp_*` via
  `erp_legacy_mapping`).
- The derived/computed fields.
- Validations that are "business rules" — those are the parity
  contract, not just the visible fields.

### 3. Check the SDA mapping

- `erp_legacy_mapping` answers: "which Histrix records have already
  been migrated to which SDA IDs".
- `services/app/db/queries/` (or `services/erp/db/queries/`) for the
  existing SDA CRUD against the target table.
- `services/app/internal/.../handler/` for existing HTTP routes.
- `apps/web/src/app/administracion/` (or `apps/web/src/app/<area>/`)
  for existing UI pages.

### 4. Write the parity diff

For every Histrix form, produce a short parity note in
`docs/parity/forms/<form-name>.md`:

```markdown
# Parity: <form-name>

**Histrix source:** .intranet-scrape/xml-forms/<area>/<form-name>.xml
**SDA equivalent:**
  - handler: services/app/internal/<module>/handler/<file>.go:<route>
  - UI: apps/web/src/app/<path>/page.tsx

**Field coverage**
| Histrix field | SDA column | Status |
|---|---|---|
| ... | ... | covered / partial / missing |

**Action coverage**
| Histrix action | SDA route | Status |
| INSERT | POST /v1/... | covered |

**Validation parity**
- <business rule 1>: <status>

**Waivers:** <none, or link to ADR>
```

### 5. Close the gap or waive

- If everything ticks "covered": the form is parity-complete. Update
  ADR 027 Phase 1 §UI parity.
- If gaps: implement them in the SDA handler + UI. Every missing field
  or action is a task; track in the parity note until done.
- If the form is dead / unused / superseded: write a waiver in
  `docs/parity/waivers.md` citing the reason and the Histrix audit-log
  evidence (last-used date).

## Shortcut: use the DB to prioritize

Not every Histrix form gets equal usage. Before committing to parity
for form X, check how many rows its primary source table actually has
in Histrix (via the migrated counts on the workstation):

```sql
-- count per Histrix source table, from migrated SDA data
SELECT legacy_table, SUM(rows_read) AS legacy_rows, SUM(rows_written) AS sda_rows
FROM erp_migration_table_progress
WHERE run_id = (SELECT id FROM erp_migration_runs ORDER BY started_at DESC LIMIT 1)
GROUP BY legacy_table ORDER BY legacy_rows DESC LIMIT 40;
```

High-usage tables → high-priority forms. A form against a 5-row config
table can wait.

## The waiver process

Waivers live in `docs/parity/waivers.md`. Each waiver entry:

```markdown
## <form-name>

- **Waived:** 2026-04-18
- **Reason:** dead feature — last accessed 2024-09, no business
  process uses it per <user>.
- **Migration:** data remains in erp_legacy_archive; if resurrected,
  new parity work required.
```

Waivers are rare and deliberate. The **default** is parity. When in
doubt, build.

## Reports parity

The company also has reports / dashboards in Histrix. Reports parity
lives in `docs/parity/reports.md` with the same shape:

- Histrix report name + SQL (if recoverable from `.intranet-scrape/`).
- SDA equivalent (endpoint, page, query).
- Sample data equivalence check.

## Cutover readiness

Before the Histrix shutdown:

- Every non-waived form is parity-complete.
- Every critical report runs in SDA.
- Seamless-day test: one real employee works a full day in SDA without
  opening Histrix. Log what was missing.
- Replay mode: migrator runs clean (no new SDA rows from current HTX
  data) for N consecutive days.
- Runbook in `docs/cutover/` covers rollback.

## Integration with ADR 027

This skill owns the Phase 1 UI-parity + reports-parity + seamless-day
checklist items. Every parity note landing ticks an item or a fragment
of an item. Waivers reference ADR 027's waiver clause.

## Don't

- Don't design a new ERP feature from scratch without reading the
  Histrix form first. The scrape is the contract.
- Don't treat `.intranet-scrape/` as read-only docs. When you find a
  gap in a form (e.g., a broken XML), note it — but fix parity, don't
  fork Histrix.
- Don't over-generalize. If Histrix has two forms that do nearly the
  same thing, SDA can consolidate — but record the consolidation
  decision as a parity waiver + ADR note, don't hide it.
- Don't skip the waiver write-up by "it's obvious that's dead". The
  friction of writing a waiver forces the check.

---
title: Package: pkg/export
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
---

## Purpose

CSV and Excel (XLSX) writers for tabular data. Both formats share the same
`Column` schema (header + key + format) and `Row` map, so handlers can offer
both download formats from one query result. Import this when an HTTP handler
needs to stream report data as CSV or XLSX (BI dashboards, accounting
reports, audit exports).

## Public API

Source: `pkg/export/export.go:2`

| Symbol | Kind | Description |
|--------|------|-------------|
| `Column` | struct | `Header`, `Key`, `Format` (`"text"`, `"number"`, `"currency"`, `"date"`, `"percent"`) |
| `Row` | type | `map[string]any` keyed by column key |
| `WriteCSV(w, columns, rows)` | func | Writes header + records using `encoding/csv` |
| `WriteExcel(w, sheetName, columns, rows)` | func | Writes XLSX via `excelize/v2` with bold header, auto-fit columns |

## Usage

```go
cols := []export.Column{
    {Header: "Date", Key: "date", Format: "date"},
    {Header: "Amount", Key: "amount", Format: "currency"},
}
rows := []export.Row{
    {"date": time.Now(), "amount": 1234.56},
}
if r.URL.Query().Get("format") == "xlsx" {
    w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    _ = export.WriteExcel(w, "Report", cols, rows)
}
```

## Invariants

- `formatValue` (`pkg/export/export.go:110`) renders all numeric formats with
  `%.2f` and `percent` multiplies by 100. Strings are passed through.
- `WriteExcel` deletes the default `Sheet1` if `sheetName != "Sheet1"`
  (`pkg/export/export.go:60`).
- Column widths are clamped to `[12, 40]` characters (`pkg/export/export.go:96`).
- Caller owns the `io.Writer`; this package does not set HTTP headers.

## Importers

`services/erp/internal/handler/analytics.go` (BI exports). Other dashboards
will follow the same pattern.

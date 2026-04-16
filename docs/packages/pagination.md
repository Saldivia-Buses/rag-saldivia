---
title: Package: pkg/pagination
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./httperr.md
---

## Purpose

Helpers for paginated list endpoints. Parses `?page=&page_size=` query
parameters with sensible defaults and bounds, computes SQL `OFFSET`/`LIMIT`,
and writes the standard `X-Page`, `X-Page-Size`, and `X-Total-Count` response
headers. Import this in every list handler so all SDA endpoints share the
same pagination contract.

## Public API

Source: `pkg/pagination/pagination.go:2`

| Symbol | Kind | Description |
|--------|------|-------------|
| `DefaultPage` | const | 1 |
| `DefaultPageSize` | const | 50 |
| `MaxPageSize` | const | 100 |
| `MaxPage` | const | 10000 (caps int32 overflow on `Offset()`) |
| `Params` | struct | `Page`, `PageSize` |
| `Params.Offset()` | method | `(Page-1) * PageSize` |
| `Params.Limit()` | method | `PageSize` |
| `Parse(r)` | func | Reads query params, applies defaults and caps |
| `SetHeaders(w, p, totalCount)` | func | Writes `X-Page`, `X-Page-Size`, optionally `X-Total-Count` (skip if `totalCount < 0`) |

## Usage

```go
p := pagination.Parse(r)
rows, err := q.ListThings(ctx, db.ListThingsParams{
    Limit: int32(p.Limit()), Offset: int32(p.Offset()),
})
total, _ := q.CountThings(ctx)
pagination.SetHeaders(w, p, int(total))
writeJSON(w, rows)
```

## Invariants

- `PageSize` is silently capped to `MaxPageSize` (100) — clients cannot
  request unbounded reads.
- `Page` is silently capped to `MaxPage` (10000) — beyond that, OFFSET would
  overflow int32 in PostgreSQL.
- Negative or zero `page`/`page_size` query values fall back to the defaults
  (`pkg/pagination/pagination.go:34`).
- Pass `-1` as `totalCount` to omit the `X-Total-Count` header (e.g., when
  computing it would be too expensive).

## Importers

`services/erp/internal/handler/*` (most handlers), `services/auth`, `chat`,
`platform`.

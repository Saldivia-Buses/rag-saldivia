# Gateway Review -- PR #89 Plan 08 Phase 3: DB & Query Hardening

**Fecha:** 2026-04-05
**Resultado:** CAMBIOS REQUERIDOS

## Bloqueantes

### B1. Pagination Offset overflow causes negative int32 cast
**Archivo:** `pkg/pagination/pagination.go:23`

`Offset()` returns `(p.Page - 1) * p.PageSize` as `int` (64-bit). All callers cast to `int32`:
```go
int32(pg.Offset())  // auth handler:375, chat handler:81, platform handler:92
```

A large `?page=21474837` with `page_size=100` produces offset `2147483600`, which fits int32. But `?page=21474838` produces `2147483700` which overflows int32, wrapping to a negative number. PostgreSQL rejects `OFFSET -N` with an error.

**Fix:** Add a max page guard in `Parse()`:
```go
const MaxPage = 10000 // or compute from MaxPageSize to stay within int32
if p.Page > MaxPage {
    p.Page = MaxPage
}
```
Or clamp `Offset()` to `math.MaxInt32`:
```go
func (p Params) Offset() int {
    off := (p.Page - 1) * p.PageSize
    if off > math.MaxInt32 { return math.MaxInt32 }
    return off
}
```

### B2. Stale sqlc models in 3 services
**Archivos:** `services/auth/internal/repository/models.go:91-103`, `services/chat/internal/repository/models.go:91-103`, `services/ingest/internal/repository/models.go:91-103`

The `FeedbackEvent` struct in auth, chat, and ingest is missing the `LatencyMs pgtype.Numeric` field that was added by migration 009. Only feedback and notification services were regenerated.

Although these services don't query `feedback_events` directly, stale models break `make sqlc` idempotency and will cause failures if anyone re-runs sqlc on those services individually.

**Fix:** Run `make sqlc` (or equivalent) for auth, chat, and ingest services to regenerate models.

## Debe corregirse

### D1. `pagination.SetHeaders` defined but never called
**Archivo:** `pkg/pagination/pagination.go:52`

The plan explicitly says: "Agregar header `X-Total-Count` para que el frontend sepa el total." The helper `SetHeaders()` is implemented but zero handlers call it. None of the paginated endpoints set `X-Page`, `X-Page-Size`, or `X-Total-Count` headers.

Without these headers, the frontend has no way to know whether there are more pages or the total count.

**Fix:** Call `pagination.SetHeaders(w, pg, -1)` in each paginated handler (auth ListUsers, chat ListSessions, platform ListTenants). Total count can be -1 initially (skips the header) until COUNT queries are added.

### D2. `ListMessages` pagination is misleading
**Archivo:** `services/chat/internal/handler/chat.go:184-185`

```go
pg := pagination.Parse(r)
messages, err := h.chatSvc.GetMessages(r.Context(), sessionID, int32(pg.Limit()))
```

`Parse(r)` reads both `?page=` and `?page_size=`, but only `pg.Limit()` is passed to `GetMessages`. The `page` parameter is silently ignored. The plan explicitly called for cursor-based pagination for messages (`WHERE created_at > $cursor`), but this was not implemented.

**Fix (minimum):** Don't call `pagination.Parse(r)` for messages if only limit is used. Instead, parse `page_size` directly or document that messages only support a limit parameter, not paging. This avoids API consumer confusion.

### D3. PerformancePercentiles still extracts from JSONB
**Archivo:** `services/feedback/db/queries/feedback.sql:31-34`

Migration 009 adds a generated column `latency_ms` to avoid JSONB extraction, but `PerformancePercentiles` still uses `(context->>'latency_ms')::numeric` instead of the new `latency_ms` column. The index is on `WHERE category = 'usage'` so it doesn't help `PerformancePercentiles` (which filters `category = 'performance'`), but the generated column is available for all rows and should be used.

**Fix:** Replace `(context->>'latency_ms')::numeric` with `latency_ms` in the query:
```sql
COALESCE(percentile_cont(0.50) WITHIN GROUP (ORDER BY latency_ms), 0)::float8 AS p50,
```

### D4. Generated column will reject inserts with non-numeric `latency_ms` in context
**Archivo:** `db/tenant/migrations/009_db_hardening.up.sql:8-9`

The expression `(context->>'latency_ms')::numeric` is STORED, meaning it is computed and stored on every INSERT/UPDATE. If any existing or future row has a `context` JSONB where `latency_ms` is a non-numeric string (e.g., `"N/A"` or `""`), the migration (for existing data) or the INSERT (for new data) will fail with a cast error.

**Fix:** Use a safer expression:
```sql
GENERATED ALWAYS AS (
    CASE WHEN context->>'latency_ms' ~ '^\d+\.?\d*$'
         THEN (context->>'latency_ms')::numeric
         ELSE NULL
    END
) STORED;
```
Or validate at the application layer that `latency_ms` is always numeric when present.

## Sugerencias

### S1. Plan items deferred from Phase 3
The following items from the plan's Fase 3 spec were not implemented in this PR:
- **M10**: Batch insert for document pages (`pgx.CopyFrom`) -- significant perf gain for multi-page PDFs
- **L5**: `DeleteJobByID` ownership check (`AND user_id = $2`)
- **L10**: Category constants in `pkg/feedback/categories.go`

These should be tracked for a follow-up PR or the plan should be updated to note deferral.

### S2. No unit tests for `pkg/pagination`
The pagination package has no `_test.go` file. Edge cases worth testing: page=0 (should default to 1), page_size=0 (default), page_size=200 (capped to 100), negative values, non-numeric strings.

### S3. `GetAllDocumentTrees` and `GetCollectionDocumentTrees` have no callers
These queries were paginated in the SQL but have zero callers in Go code. The search service uses raw SQL for tree loading (`services/search/internal/service/search.go:122`). Consider migrating the search service to use the sqlc-generated queries.

### S4. `PurgeOldNotifications` and `PurgeOldEvents` have no scheduled invocation
Both purge queries are defined but never called. They need a cron job or periodic goroutine to be effective.

## Lo que esta bien

- **LATERAL JOIN** for `ListActiveUsers` is clean and correct -- resolves the N+1 subquery correctly. The SQL matches the plan spec exactly.
- **Migration up/down symmetry** is correct: down reverses in the right order (index first, then column, then FK constraint).
- **Redundant index removal** (`idx_refresh_tokens_hash`) is correct -- the UNIQUE constraint on `token_hash` already creates an implicit index.
- **FK `documents.uploaded_by -> users(id)`** is correctly defined without ON DELETE CASCADE, which is appropriate (don't cascade-delete documents when a user is removed).
- **`pagination.Parse()`** handles edge cases well: invalid strings, zero, negative values all fall back to defaults. MaxPageSize cap prevents unreasonable queries.
- **`CountHistoricalUsage`** correctly changed from unbounded full-table scan to parameterized time filter.
- **`ListNotifications`** already caps at 100 in the service layer (`notification.go:91-92`).
- **Test mocks** (auth `auth_test.go:69`, chat `chat_test.go:53`, platform `platform_test.go:38`) all have correct signatures matching their respective interfaces. No compilation issues.
- **Generated sqlc code** matches the SQL queries for all services that were regenerated.

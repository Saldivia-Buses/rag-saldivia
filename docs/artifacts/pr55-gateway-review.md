# Gateway Review -- PR #55 Chat Enhancements (backend)

**Fecha:** 2026-04-03
**Resultado:** CAMBIOS REQUERIDOS

---

## Bloqueantes

### 1. [chat_test.go:95] Mock does not implement updated ChatService interface -- tests will not compile

The `ChatService` interface now requires:
```go
AddMessage(ctx, sessionID, userID, role, content string, thinking *string, sources, metadata []byte) (*service.Message, error)
```

But the mock at `services/chat/internal/handler/chat_test.go:95` still uses the old signature:
```go
func (m *mockChatService) AddMessage(_ context.Context, sessionID, userID, role, content string, sources, metadata []byte) (*service.Message, error)
```

Missing `thinking *string` parameter. This is a compile error -- `go test ./services/chat/...` fails.

**Fix:** Add `thinking *string` parameter to the mock:
```go
func (m *mockChatService) AddMessage(_ context.Context, sessionID, userID, role, content string, thinking *string, sources, metadata []byte) (*service.Message, error) {
```

### 2. [chat_integration_test.go:190,198,243,244] Integration test calls use old AddMessage signature

All `svc.AddMessage()` calls pass 7 args (old) instead of 8 (new). They need a `nil` for thinking inserted between `content` and `sources`:
```go
// Before:
svc.AddMessage(ctx, session.ID, "u-1", "user", "Hola", nil, nil)
// After:
svc.AddMessage(ctx, session.ID, "u-1", "user", "Hola", nil, nil, nil)
```

Four call sites: lines 190, 198, 243, 244.

### 3. [chat_integration_test.go:73-82] Integration test schema missing `thinking` column

The inline SQL migration in `setupTestDB` creates the `messages` table without the `thinking` column. When the service `Scan`s for `thinking`, pgx will fail with a column-not-found or scan-mismatch error.

**Fix:** Add `thinking TEXT,` between `content` and `sources` in the inline schema:
```sql
CREATE TABLE messages (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    thinking TEXT,
    sources JSONB,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

## Debe corregirse

### 4. [rag.go:70-86] Reasoning config leaks to Blueprint mode

In `GenerateStream`, when `r.cfg.APIKey == ""` (Blueprint mode), the `req.Reasoning` field is not cleared. The NVIDIA Blueprint `/v1/chat/completions` endpoint does not support a `reasoning` field -- at best it ignores it, at worst it returns a 400.

**Fix:** Clear `req.Reasoning` in the Blueprint branch (line ~74):
```go
if r.cfg.APIKey == "" {
    // Blueprint mode
    req.Reasoning = nil  // Blueprint does not support reasoning
    // ... existing collection logic
} else {
    // OpenRouter mode
    // ... existing logic
}
```

---

## Sugerencias

### 5. Migration is safe and idempotent

`ALTER TABLE messages ADD COLUMN IF NOT EXISTS thinking TEXT` is correct:
- `IF NOT EXISTS` makes it idempotent (safe to re-run).
- `TEXT` with no `NOT NULL` constraint defaults to `NULL` for existing rows -- zero data loss risk.
- The down migration (`DROP COLUMN IF EXISTS thinking`) is symmetric and correct.
- No data backfill needed -- `*string` (pointer) in Go maps `NULL` to `nil` cleanly.

### 6. Scan order is correct

The `INSERT ... RETURNING` at `chat.go:163-165` returns 8 columns: `id, session_id, role, content, thinking, sources, metadata, created_at`. The `.Scan()` on line 167 matches this order exactly with `&m.ID, &m.SessionID, &m.Role, &m.Content, &m.Thinking, &m.Sources, &m.Metadata, &m.CreatedAt`. Same for `GetMessages` SELECT at line 205.

The INSERT uses `$1-$6` for `(session_id, role, content, thinking, sources, metadata)` which is 6 columns -- matches the 6 values passed on line 166. Correct.

### 7. Thinking is properly optional

- `Thinking *string` with `json:"thinking,omitempty"` means it is `nil` when omitted from JSON input.
- Handler passes `req.Thinking` (which is `nil` when not in request body) directly to service.
- Service passes it as `$4` to PostgreSQL, which stores `NULL`.
- On read, pgx scans `NULL` into `*string` as `nil`.
- JSON marshal with `omitempty` omits it from response when `nil`.
- Full round-trip is clean.

### 8. Consider adding a test for the thinking field

The existing `TestAddMessage_ValidRole_Success` should be extended (or a new test added) that sends a request body with `"thinking":"some reasoning"` and verifies it appears in the response. This validates the full path through handler -> service -> response.

---

## Lo que esta bien

- Migration is safe, idempotent, and has a proper down migration.
- `*string` is the right Go type for a nullable TEXT column -- clean nil/NULL mapping.
- No changes to auth middleware, tenant isolation, or NATS events -- no security surface change.
- Scan column order matches SELECT/RETURNING column order in both AddMessage and GetMessages.
- Handler properly validates role and content before passing thinking through.
- `MaxBytesReader` limit on AddMessage request body is already in place.
- RAG `ReasoningCfg` struct is clean -- only exposes `effort` string, no arbitrary passthrough.
- Blueprint-specific fields are correctly cleared in OpenRouter mode (existing logic).

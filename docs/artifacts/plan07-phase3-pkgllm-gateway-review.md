# Gateway Review -- Plan 07 Phase 3: pkg/llm

**Fecha:** 2026-04-05
**PR:** #84 (feat/plan07-pkg-llm)
**Resultado:** CAMBIOS REQUERIDOS

## Bloqueantes

### B1: Ingest go.mod missing pkg dependency [services/ingest/go.mod]

The ingest service's `go.mod` does not have a `replace` directive or `require` entry for `github.com/Camionerou/rag-saldivia/pkg`, yet `services/ingest/internal/llm/client.go` imports `pkgllm "github.com/Camionerou/rag-saldivia/pkg/llm"`.

This compiles locally because `go.work` resolves the path, but it will **fail in Docker builds and CI** where `go.work` is not used.

**Fix:** Add to `services/ingest/go.mod`:
```
replace github.com/Camionerou/rag-saldivia/pkg => ../../pkg

require (
    github.com/Camionerou/rag-saldivia/pkg v0.0.0-00010101000000-000000000000
    ...
)
```
Then run `go mod tidy` from `services/ingest/`.

### B2: Plan verification criterion NOT met [services/agent/*, services/ingest/*]

The plan states:

> **Verificacion:** `grep -r "internal/llm" services/` retorna 0 resultados.

Current state: 4 files still import `internal/llm`:
- `services/agent/internal/service/agent.go`
- `services/agent/internal/handler/agent.go`
- `services/agent/cmd/main.go`
- `services/ingest/internal/tree/generate.go`

The PR chose backward-compatible re-export shims instead of direct migration. This is a reasonable intermediate step, but it contradicts the plan's explicit success criterion.

**Fix (pick one):**
1. **Complete the migration:** Update all 4 files to import `pkg/llm` directly and delete `services/agent/internal/llm/adapter.go` and `services/ingest/internal/llm/client.go` entirely. This is the plan's intent.
2. **Update the plan:** If shims are intentional (e.g., to reduce diff size), document this decision and adjust the verification criterion. But this defers tech debt that the consolidation plan was created to eliminate.

Option 1 is strongly recommended -- the re-export shims add complexity without value. The type alias `Adapter = pkgllm.Client` in agent just renames `Client` to `Adapter` for backward compat, but `Adapter` was the old name that was already wrong (it's a client, not an adapter).

## Debe corregirse

### D1: SimplePrompt hardcodes max_tokens=4096 for all callers [pkg/llm/client.go:159]

`SimplePrompt` hardcodes `maxTokens: 4096` which is:
- **Wasteful for search:** The tree navigation prompt returns comma-separated node IDs -- maybe 50-100 tokens. Sending `max_tokens=4096` wastes SGLang's KV-cache budget on reserved-but-unused output slots.
- **Potentially insufficient:** If a future caller needs more, they can't use `SimplePrompt` and must fall back to `Chat()`.
- **Inconsistent with agent:** Agent uses `8192`.

**Fix:** Add `maxTokens` parameter to `SimplePrompt`:
```go
func (c *Client) SimplePrompt(ctx context.Context, prompt string, temperature float64, maxTokens int) (string, error) {
    if maxTokens <= 0 {
        maxTokens = 4096 // default
    }
    ...
}
```
This is a one-line API change that makes the function properly flexible. The callers in ingest and search can pass 0 for default, and search's `navigateTrees` can pass a smaller value like 512.

### D2: Dead comment at end of search.go [services/search/internal/service/search.go:328]

```go
// llmChat is now handled by pkg/llm.Client.SimplePrompt — see line 191.
```

This is a change-log comment, not documentation. It provides no value to future readers and will become confusing after further edits shift line numbers.

**Fix:** Delete line 328.

### D3: Response body not drained on success path [pkg/llm/client.go:139-141]

After `json.NewDecoder(resp.Body).Decode(&raw)`, the body may not be fully consumed. This prevents HTTP/1.1 connection reuse since `net/http` requires the body to be fully read before the connection can be returned to the pool.

**Fix:** Add after the decode block:
```go
io.Copy(io.Discard, resp.Body)
```

This is low-priority for localhost SGLang but matters when the same client is pointed at cloud providers (OpenAI, Anthropic) where connection reuse saves TLS handshake latency.

## Sugerencias

### S1: No tests for pkg/llm

`pkg/llm/` has zero test files. Since this is now the **single LLM client for the entire system**, it needs at minimum:
- A unit test using `httptest.Server` that validates request format (headers, body structure, auth header presence/absence)
- A test for error handling (non-200 responses, empty choices, malformed JSON)
- A test that `SimplePrompt` sends only a user message with the correct max_tokens

Delegate to **test-writer** agent.

### S2: Consider functional options pattern

The plan specified `Chat(ctx, msgs, opts ...ChatOption)` but the PR uses positional params `Chat(ctx, msgs, tools, temperature, maxTokens)`. The current signature is fine for now but will need breaking changes when adding `tool_choice`, `response_format`, `stop`, `top_p`, or `stream`. The functional options pattern (`WithTools(...)`, `WithTemperature(0.5)`, etc.) is the idiomatic Go approach for this.

Not blocking for this PR -- the current callers only need what's exposed. But flag this for Phase 6 (cmd/main.go standardization) where the API surface will be finalized.

### S3: Missing NewFromModelConfig / NewFromSlot

The plan specifies `NewFromModelConfig` and slot resolution integration. This appears to be deferred to a later phase. Acceptable as long as the plan tracks it.

### S4: Search service still uses hardcoded POSTGRES_TENANT_URL

`services/search/cmd/main.go:32` uses `env("POSTGRES_TENANT_URL", "")` with a single pgxpool. In a multi-tenant-per-database model, this should use `pkg/tenant.Resolver` to connect to the correct database per request. This is a pre-existing issue (flagged in PRs #77, #82) but worth noting since the search service is being modified in this PR.

### S5: multimodal.go still has its own HTTP client

`services/agent/internal/tools/multimodal.go` has a duplicated OpenAI-compatible HTTP client (lines 16-29, 43-95). The plan explicitly says this should be refactored to use `pkg/llm` but "No wirearlo en el agent -- eso es feature nueva, va en un plan separado." Confirming this is out of scope per the plan.

## Lo que esta bien

- **Client design is clean.** The `Client` struct is immutable after construction, thread-safe by design, and the comment documents this correctly.
- **otelhttp transport wrapping.** Trace propagation is built into the client at transport level -- this means every LLM call automatically participates in distributed traces without any caller effort.
- **LimitReader on error path.** `io.LimitReader(resp.Body, 4096)` prevents OOM on malformed error responses. This was a gap in previous implementations that is now fixed.
- **Type alias approach is correct for Go.** Using `type X = pkg.X` (alias, not definition) means JSON struct tags, method sets, and type assertions all work transparently. No behavioral change from the migration.
- **Error wrapping is consistent.** All errors use `fmt.Errorf("context: %w", err)` pattern per bible conventions.
- **API key is optional.** `if c.apiKey != ""` check allows the client to work both with SGLang (no auth) and cloud providers (API key auth).
- **The search service correctly calls SimplePrompt.** The old inline `llmChat` is fully replaced with `s.llmClient.SimplePrompt()` at line 191, and the search service now imports `pkg/llm` directly (not through a shim). This is the correct target state.

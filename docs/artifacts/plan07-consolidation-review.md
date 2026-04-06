# Plan 07 Review — Consolidacion: shared schema, config central, dedup, estandarizacion

> **Reviewer:** plan-writer agent
> **Date:** 2026-04-05
> **Plan reviewed:** `docs/plans/2.0.x-plan07-consolidation.md`
> **Verdict:** CONDITIONAL APPROVAL — 3 blockers, 7 recommendations

---

## Executive Summary

The plan correctly identifies real problems. The duplication counts are accurate
(verified against code). The proposed solutions are architecturally sound. However,
there are three issues that would break things if not addressed before execution,
and several design decisions that should be refined.

The plan is a good consolidation plan with one phase (Phase 7) that is clearly
a new feature, not refactoring. That phase should be split out.

---

## 1. Completeness — Does it cover all issues from the audit?

**Verified duplication counts (actual vs. claimed):**

| Item | Claimed | Actual | Status |
|------|---------|--------|--------|
| `env()` copies | 10 | 10 (auth, chat, feedback, ingest, notification, platform, search, traces, agent, ws) | Correct |
| `loadPublicKey()` copies | 9 | 9 (all except auth which has `loadJWTKeys()`) | Correct |
| LLM clients | 3 | 3 (agent/internal/llm, ingest/internal/llm, search/service inline) | Correct |
| NATS subject validation patterns | 3 | 3 (pkg/nats denylist, agent/service/traces.go regex, extractor/main.py regex) | Correct |
| `pkg/services/` empty dirs | claimed | Directory does not exist at all (already clean or never created) | False positive |
| `pkg/config` empty | claimed | `pkg/config/` does not exist at all | Correct (no directory, not empty) |
| Token blacklist not wired | claimed | Confirmed — `pkg/security/blacklist.go` exists with tests but zero imports from any service | Correct |
| `multimodal.go` not wired | claimed | Confirmed — exists at `services/agent/internal/tools/multimodal.go`, not in tool definitions or executor | Correct |

**Missing from the plan:**

1. **`000_deps.up.sql` stub files** — chat, ingest, and notification each have a
   `000_deps.up.sql` that creates a stub `users` table for sqlc FK resolution.
   The plan says "services maintain their `db/queries/` and `sqlc.yaml`" but does
   not address what happens to these stubs. When migrations move to `db/tenant/`,
   the sqlc compilation in each service still needs the full schema. This needs a
   concrete answer.

2. **`nc.Close()` vs `nc.Drain()` inconsistency** — The plan mentions it in Phase 4
   but doesn't list which services currently use `nc.Close()` vs `nc.Drain()`.
   Actual count: auth, chat, notification, ingest, feedback all use `defer nc.Close()`.
   Agent, traces use `defer nc.Drain()`. WS not checked but likely Close. This is
   7 services to change. The plan should enumerate them.

3. **Auth service has `loadJWTKeys()` not `loadPublicKey()`** — The plan says "9
   copies of `loadPublicKey()`" but auth uses a different function `loadJWTKeys()`
   that loads BOTH private and public keys. A shared `MustLoadPublicKey()` wouldn't
   replace auth's function. The plan needs to account for this: auth needs both
   `config.MustLoadPublicKey()` AND a separate `jwt.MustLoadPrivateKey()` (or similar).

4. **Feedback service** — Not listed in the `000_deps.up.sql` pattern but has its
   own migration file. Needs to be included in the centralized migration ordering.

5. **`nc.JetStream()` old API** — The plan correctly identifies traces using the
   old `nc.JetStream()` API but also the extractor's Python code uses the old
   `js.subscribe` pattern. The plan only says "Traces service migra de `nc.JetStream()`
   a `jetstream.New()`" — should also mention notification service which uses
   JetStream durables.

---

## 2. Phase Ordering — Dependencies correct?

The dependency graph is mostly correct. Analysis:

```
Phase 1 (migrations)     ← independent
Phase 4 (nats)           ← independent
Phase 5 (security)       ← independent
Phase 2 (pkg/config)     ← depends on Phase 1 (needs Platform DB schema in place)
Phase 3 (pkg/llm)        ← depends on Phase 2 (NewFromSlot)
Phase 6 (main.go)        ← depends on Phases 2, 3, 4
Phase 7 (multimodal)     ← depends on Phases 2, 3
```

**Correct.** The plan states this accurately.

**Parallelization opportunity missed:** Phase 3's `pkg/llm` extraction can be done
in TWO sub-phases:
- Phase 3a: Extract `pkg/llm` as a standalone package (no `NewFromSlot`). Just the
  `Client` struct with `NewClient(endpoint, model, apiKey)`. This depends on nothing.
- Phase 3b: Add `NewFromSlot(cfg *config.ModelConfig)` after Phase 2 lands.

This would allow Phase 3a to run in parallel with Phases 1, 4, 5, and Phase 2
would only block Phase 3b. **Recommendation: split Phase 3.**

---

## 3. Scope Creep Risk

**BLOCKER: Phase 7 is a new feature, not refactoring.**

The plan says: "Wire multimodal.go — refactorear para usar `pkg/llm` + slot
resolution via `pkg/config`, agregar a tool definitions del agent, agregar routing
especial en executor para tools internas (no HTTP)."

This is not consolidation. This is:
- Adding a new tool to the agent's tool definitions (feature)
- Adding internal tool routing to the executor (feature: new code path)
- Fixing the `userID` TODO (bug fix, fine in a refactor plan)

The multimodal refactoring to use `pkg/llm` is legitimate (dedup). The wiring into
the agent is a feature. **Split Phase 7 into:**
- Phase 7a (refactor): Refactor `multimodal.go` to use `pkg/llm` instead of inline HTTP
- Phase 7b (feature): Wire `analyze_image` into agent — move to a separate plan or
  a "post-consolidation" phase clearly marked as feature work

**Minor scope concern in Phase 2:** The `pkg/config` Resolver is enormous. The API
surface includes `Get`, `GetString`, `GetInt`, `GetFloat`, `GetBool`, `ResolveSlot`,
`LoadPipeline`, `GetActivePrompt`, `GetEnabledTools`, `GetRateLimits`, `InvalidateCache`,
`Env`, `MustEnv`. That's 12 methods. For a consolidation plan, the question is: do
ALL 12 methods need to land in this plan, or only the ones that replace existing
duplication (`Env`, `MustEnv`, `ResolveSlot`)? The pipeline/prompt/rate-limit
resolution methods are **new capabilities** that no service currently uses. They're
laying groundwork for when services start reading config from DB instead of env vars.

**Recommendation:** Phase 2 should be split:
- Phase 2a: `Env()`, `MustEnv()` (replaces 10 copies — pure dedup)
- Phase 2b: `ResolveSlot()`, `LoadPipeline()`, `GetActivePrompt()`, etc. (new
  capability — can land now as code, but wiring services to use it is a separate
  concern)

This keeps the "consolidation" claim honest.

---

## 4. Feasibility — Will it break things?

**BLOCKER: Migration renumbering is dangerous.**

The plan proposes renumbering migrations:
```
001_auth.up.sql          <- was services/auth/db/migrations/001_init.up.sql
002_chat.up.sql          <- was services/chat/db/migrations/001_init.up.sql
003_ingest.up.sql        <- was services/ingest/db/migrations/001_init.up.sql + 002
```

Problem: if ANY existing database has already run these migrations, the migration
tool (currently a bash script `deploy/scripts/migrate.sh` running `psql -f`) will
try to re-run them with new names. Since the script uses `IF NOT EXISTS` on CREATE
TABLE statements, it might not error — but this is fragile and undocumented.

**The current `migrate.sh` doesn't track migration state** — it just runs all `.up.sql`
files in order via `psql`. There is no `schema_migrations` table. This means:
- On a fresh DB: works fine
- On an existing DB: `CREATE TABLE IF NOT EXISTS` prevents errors but `INSERT` seeds
  might duplicate or conflict

**Recommendation:** Before centralizing, add proper migration tracking (even a simple
`schema_migrations` table with applied file hashes). Or explicitly document that this
plan requires a fresh DB rebuild, which is acceptable for dev/staging but must be
called out.

**`nc.Close()` to `nc.Drain()` migration risk:** Changing from `Close` to `Drain`
changes shutdown behavior — `Drain` waits for in-flight messages to complete.
Services with JetStream consumers (notification, traces, feedback) should use `Drain`.
Services that are pure publishers (auth, chat) could use either. This is safe but
the plan should note that `Drain` blocks during shutdown, and services should still
have a shutdown timeout (which they do via `signal.NotifyContext`).

---

## 5. The Shared Schema Decision

**The decision to centralize migrations in `db/` is correct** for this project.

Rationale:
- All tenant services share the same PostgreSQL instance per tenant
- FK dependencies already exist (chat, ingest, notification reference `users` from auth)
- The `000_deps.up.sql` stubs prove the current model is already broken — services
  need to know about other services' tables
- `migrate.sh` already hardcodes the ordering: `auth chat notification ingest`

**However, there is a sqlc concern.** Each service has its own `sqlc.yaml` pointing
to its own `db/queries/` and `db/migrations/` for schema inference. When migrations
move to `db/tenant/migrations/`, every service's `sqlc.yaml` must be updated to
point to the centralized schema directory. The plan mentions this implicitly
("services maintain their `db/queries/` and `sqlc.yaml`") but should be explicit:
every `sqlc.yaml` gets a `schema` path change.

**Current sqlc.yaml pattern (example from chat):**
```yaml
schema: "db/migrations/"    # needs to become "../../db/tenant/migrations/"
queries: "db/queries/"      # stays the same
```

The plan should list every `sqlc.yaml` that needs updating. Count: auth, chat,
ingest, notification, feedback (tenant DB), platform (platform DB), traces (platform DB),
search (tenant DB). That's 8 files.

**Additionally:** The `000_deps.up.sql` stubs can be removed entirely once the
centralized schema includes the auth migration first. But sqlc still needs the full
schema visible. This works if `sqlc.yaml` points to `../../db/tenant/migrations/`
(which includes the auth migration that creates the real `users` table).

---

## 6. The `pkg/config` Resolver

**API surface: appropriately sized for the long term, too big for this plan.**

The Resolver does three distinct things:
1. **Env helpers** (`Env`, `MustEnv`) — trivial, zero dependencies, replace 10 copies
2. **Config resolution** (`Get`, `GetString`, etc.) — needs Platform DB + Redis
3. **Domain-specific resolution** (`ResolveSlot`, `LoadPipeline`, `GetActivePrompt`,
   `GetEnabledTools`, `GetRateLimits`) — needs Platform DB + knowledge of agent_config
   schema

Recommendations:
- Items 1 and 2 should be separate from each other. `Env`/`MustEnv` are pure functions
  with no dependencies — they can live in `pkg/config/env.go` and land immediately.
- The Resolver struct can land with a minimal surface: `Get` + `ResolveSlot` + cache.
  Other methods can be added when a consumer actually needs them.

**Redis caching is the right approach.** Config doesn't change frequently, and multiple
services reading the same slot config should not each hit the Platform DB. TTL-based
cache with explicit invalidation via NATS lifecycle events is sound.

**One concern:** The Resolver takes `*pgxpool.Pool` for Platform DB. Services that
currently only connect to Tenant DB (chat, search, ingest) would need a NEW database
connection to Platform DB just to resolve config. This is a significant infrastructure
change — every service now needs `POSTGRES_PLATFORM_URL`. The plan should:
1. Make Platform DB connection optional in the Resolver (graceful degradation to env vars)
2. Or acknowledge that this env var must be added to all services' Docker configs

---

## 7. The `pkg/llm` Extraction — Circular Dependency

**`NewFromSlot` does NOT create a circular dependency.** Here is the actual dependency
graph:

```
pkg/config   imports: pgxpool, redis
pkg/llm      imports: pkg/config (only for ModelConfig type and NewFromSlot)
services/*   imports: pkg/llm, pkg/config
```

There is no cycle. `pkg/config` does not import `pkg/llm`.

**However, there is a coupling concern.** If `pkg/llm` imports `pkg/config`, then
any service that uses `pkg/llm.NewClient()` directly (e.g., for a quick test without
a Platform DB) still transitively imports `pgxpool` and `redis` via `pkg/config`.
This bloats the dependency tree unnecessarily.

**Recommendation:** `NewFromSlot` should be a standalone function, not in `pkg/llm`.
Instead, put it in `pkg/config` or in a thin adapter:

```go
// In pkg/config/llm.go
func NewLLMClient(mc *ModelConfig) *llm.Client {
    return llm.NewClient(mc.Endpoint, mc.ModelID, mc.APIKey)
}
```

Or keep `NewFromSlot` in `pkg/llm` but only import the `ModelConfig` struct, which
should be defined in a types-only package (e.g., `pkg/config/types.go`) with no
heavy dependencies. This is a refinement, not a blocker.

---

## 8. Missing Concerns

### Tests

**The plan has ZERO test requirements.** This is a significant gap for a refactoring
plan. Every phase should include:

- Phase 1: `make migrate` on a fresh DB succeeds. `make sqlc` generates clean. All
  existing tests still pass.
- Phase 2: Unit tests for `Env()`, `MustEnv()`. Integration test for Resolver with
  a real Platform DB (testcontainers). Test cascaded resolution: tenant > plan > global.
- Phase 3: Unit test for `Client.Chat()` with httptest server. Test that
  `SimplePrompt` works. Verify `grep -r "internal/llm" services/` = 0.
- Phase 4: Test `natspub.Connect()` returns a valid connection. Test
  `IsValidSubjectToken` with edge cases. Verify `grep -r "nats.Connect(" services/` = 0.
- Phase 5: Integration test: login > get token > logout > use token > 401.
- Phase 6: All existing tests still pass. Visual diff of all `cmd/main.go` files.
- Phase 7: `analyze_image` tool test (if wiring lands here).

**Recommendation: add test requirements to every phase.**

### Python Extractor

The plan says "Borrar `_SAFE_SUBJECT_RE` de `extractor/main.py` (documentar el regex
canonico)." But the Python extractor cannot import Go packages. It will need its own
copy of the validation regex. The plan should say:

> "Replace `_SAFE_SUBJECT_RE` in `extractor/main.py` with a documented canonical
> regex that matches `pkg/nats.IsValidSubjectToken`. Add a comment referencing the
> Go package as the source of truth."

One copy in Python is acceptable — it's a different language. Zero copies is
impossible. The plan should acknowledge this rather than implying the Python
duplication can be eliminated.

### `make sqlc` impact

Centralizing migrations changes every `sqlc.yaml` schema path. After Phase 1, every
service's `make sqlc` (or global `make sqlc`) must still generate correct code. This
should be an explicit verification step in Phase 1.

### `deploy/scripts/migrate.sh`

The plan says "update db-init script" but the actual script is `deploy/scripts/migrate.sh`.
It should reference the exact file path.

---

## 9. Verification Criteria

| Phase | Criteria in Plan | Measurable? | Gap |
|-------|-----------------|-------------|-----|
| 1 | `make migrate` works, services compile, db-init works | Partially — "db-init" does not exist, it's `migrate.sh` | Fix script name |
| 2 | Test Resolver with Platform DB | Too vague — what assertions? | Add specific test scenarios |
| 3 | `grep -r "internal/llm" services/` = 0 | Yes, measurable | Good |
| 4 | `grep -r "nats.Connect(" services/` = 0 | Yes, measurable | Good |
| 5 | Integration test: login > logout > use token > 401 | Yes, measurable | Good |
| 6 | Visual diff of cmd/main.go | Subjective | Add "all follow the same 11-step order" as checklist |
| 7 | `analyze_image` works, `pkg/services/` gone | `pkg/services/` already doesn't exist | Remove false target |

**Success metrics table at the bottom is good** — all are grep-able or countable.
But it's missing:
- `make test` passes after each phase
- `make sqlc` generates clean after Phase 1
- `make build` compiles all services after each phase
- No new env vars required unless explicitly listed

---

## Blockers (must fix before execution)

1. **Phase 7 scope creep** — Wiring multimodal into the agent is a feature, not a
   refactoring. Split Phase 7 into refactor (use `pkg/llm`) and feature (wire into
   agent). The feature part belongs in a separate plan or is explicitly marked.

2. **Migration renumbering without tracking** — The current `migrate.sh` has no
   state tracking. Renumbering files risks double-applying on existing databases.
   Either add migration tracking or document that this requires a fresh DB.

3. **`pkg/services/` false positive** — The plan says "Borrar `pkg/services/`
   (directorios vacios)" but this directory does not exist. Remove this from the
   plan to avoid confusion during execution.

---

## Recommendations (should fix)

1. **Split Phase 3** into 3a (extract `pkg/llm`, no config dependency) and 3b
   (add `NewFromSlot` after Phase 2). Enables more parallelism.

2. **Split Phase 2** into 2a (`Env`/`MustEnv` — pure dedup) and 2b (Resolver with
   DB/Redis — new capability). Keeps the consolidation claim honest.

3. **Add test requirements** to every phase. A refactoring plan without tests is a
   regression plan.

4. **Address sqlc.yaml paths** explicitly in Phase 1. List all 8 files that need
   `schema` path changes.

5. **Fix the Python extractor story** — Acknowledge that one copy of the validation
   regex in Python is necessary and acceptable. Document the canonical regex.

6. **Make Platform DB connection optional** in the Resolver — services that don't
   need business config should not need a Platform DB connection. Graceful fallback
   to env vars.

7. **Account for auth's `loadJWTKeys()`** — It loads both private and public keys.
   `pkg/jwt.MustLoadPublicKey()` doesn't replace it. Add `pkg/jwt.MustLoadPrivateKey()`
   or a combined `MustLoadKeyPair()`.

---

## Verdict

**CONDITIONAL APPROVAL** — fix the 3 blockers, then this plan is ready to execute.
The recommendations are improvements that would make execution smoother but are not
hard blockers.

The plan is well-structured, correctly identifies real problems, proposes the right
abstractions, and has a sound dependency graph. The main issues are one phase that's
secretly a feature, migration safety on existing databases, and missing test coverage.
Fix those and ship it.

# CLAUDE.md

## Session Start Protocol (MANDATORY)

On EVERY new session, before doing anything:
1. SessionStart hook runs automatically → shows recent commits, modified files, branch state
2. Read `docs/bible.md` — permanent conventions
3. Check `MEMORY.md` — feedback from past sessions
4. If task involves a critical flow → read the relevant section of `docs/CRITICAL_FLOWS.md`

## Code Intelligence (MANDATORY)

Two MCP servers provide codebase intelligence. Use them — they prevent regressions.

### CodeGraphContext (code graph — auto-updates on file changes)

| When | Tool | Example |
|------|------|---------|
| **Before editing a function** | `analyze_code_relationships(find_callers, "FunctionName")` | See who calls it — if you change the signature, these break |
| **Before editing `pkg/*`** | `analyze_code_relationships(find_importers, "pkg/tenant")` | Every service importing this package |
| **Exploring code** | `find_code("query keyword")` | Find functions, types, or patterns |
| **Checking blast radius** | `analyze_code_relationships(find_all_callers, "FunctionName")` | Full transitive caller tree |
| **Finding dead code** | `find_dead_code(repo_path="/home/enzo/rag-saldivia")` | Unused functions to clean up |
| **Complexity check** | `find_most_complex_functions()` | Top 10 most complex functions |
| **After renaming** | `find_code("OldName")` | Verify no references remain |

Graph stats: ~270 functions, ~108 files, ~83 modules. Watch enabled — auto-updates on every file change.

### Repowise (documentation engine)

| When | Tool |
|------|------|
| **First call on every task** | `get_overview()` |
| **Before reading/editing a file** | `get_context(targets=["path/to/file"])` |
| **Before changing hotspot files** | `get_risk(targets=["path/to/file"])` |
| **Before architectural changes** | `get_why(query="why X over Y")` |
| **Finding code** | `search_codebase(query="authentication flow")` |
| **After code changes** | `update_decision_records(action="list")` then create/update |
| **Tracing dependencies** | `get_dependency_path(source="src/auth", target="src/db")` |

## Regression Protection

Before editing ANY Go file in `services/` or `pkg/`:
1. `git log --oneline -5 -- {file}` — was it recently changed?
2. If changed in last 48h: read the diff first (`git diff HEAD~5..HEAD -- {file}`)
3. `analyze_code_relationships(find_callers, "FunctionName")` — who calls this?
4. `get_context()` and `get_risk()` for hotspot files
5. If changing a function signature → `find_all_callers` to see full blast radius
6. If changing `pkg/*` → `find_importers` to check every consumer

## CRITICAL INVARIANTS (7 rules that MUST NOT break)

These are the SDA equivalent of astro-v2's domain invariants. Violations block commits.

1. **Tenant isolation at every layer** — JWT claim tenant == request tenant. Every sqlc query for tenant data includes `tenant_id` in WHERE. No hardcoded tenant IDs anywhere. `tenant.Resolver` provides per-tenant DB pools — never share a pool across tenants.

2. **JWT is the single source of identity** — UserID, TenantID, Slug, Role all come from JWT claims. Services verify JWT locally with ed25519 public key. Never trust client-supplied identity headers without middleware validation.

3. **NATS subjects are tenant-namespaced** — ALL events follow `tenant.{slug}.{service}.{entity}[.{action}]` format. Never publish without slug prefix. Consumers use `tenant.*.{service}.>` wildcard.

4. **Every write publishes a NATS event** — so the WebSocket Hub can push real-time updates. No polling in frontend. If you add a mutation endpoint, you MUST publish the corresponding NATS event.

5. **Migration pairs are always complete** — every `.up.sql` has a matching `.down.sql`. Numbers are sequential with no gaps. sqlc generated code must match the current schema.

6. **Service structure is uniform** — every Go service has `cmd/main.go`, `VERSION` (valid semver), `Dockerfile` (multi-stage, non-root), `README.md`. Every service is registered in `go.work`.

7. **Error responses are JSON** — all `http.Error` calls in handlers must use JSON format: `` http.Error(w, `{"error":"msg"}`, code) `` or `writeJSON()`. Never plain text error responses on API endpoints.

Run `bash .claude/hooks/check-invariants.sh` to verify all checks.

## Known Staleness Risks

- **`docs/CRITICAL_FLOWS.md` line numbers** drift on every edit — use function names to locate code, not line numbers
- **Repowise index** can go stale — invariant check warns if >3 days old
- **sqlc generated code** can drift — invariant check compares query vs generated timestamps

## File-Specific Guards

**READ `docs/CRITICAL_FLOWS.md` BEFORE modifying these files:**

| If you're touching... | Read Flow # |
|----------------------|-------------|
| `services/auth/internal/` | Flow 1: Auth (login → JWT → refresh) |
| `pkg/tenant/`, `pkg/middleware/`, `pkg/database/` | Flow 2: Multi-tenant routing |
| `services/agent/internal/`, `services/chat/internal/` | Flow 3: Chat + Agent pipeline |
| `services/ingest/internal/`, `services/extractor/` | Flow 4: Document ingestion |
| `services/ws/internal/`, `pkg/nats/` | Flow 5: WebSocket real-time |
| `deploy/`, `.github/workflows/deploy.yml` | Flow 6: Deploy Pipeline |
| `services/healthwatch/`, `.github/workflows/daily-triage.yml` | Flow 7: Self-Healing Loop |

## Hooks (automated)

| Hook | Event | What it does |
|------|-------|-------------|
| Session briefing | `SessionStart` | Shows last 10 commits, modified files in 48h, uncommitted changes, service versions |
| Pre-edit regression | `PreToolUse` on Edit/Write | Alerts if a Go file was modified in recent commits — read diff first |
| Pre-commit invariants | `PreToolUse` on `git commit` | Runs 26 invariant checks — **blocks commit if any fail** (exit 2) |
| Stop verification | `Stop` | Haiku verifies claims have evidence before session ends |

## Skills (`.claude/skills/superpowers/`)

14 specialized skills. Also available as `/command` slash commands.

| Priority | Skill | When |
|----------|-------|------|
| ALWAYS | `using-superpowers` | Session start — establishes framework |
| Before code | `brainstorming` | New feature, service, or module |
| Before code | `writing-plans` | Clear spec, multi-step task |
| During code | `test-driven-development` | EVERY implementation — test first |
| During code | `systematic-debugging` | Bug, failure, unexpected behavior |
| During code | `subagent-driven-development` | Dispatching per-task agents |
| During code | `dispatching-parallel-agents` | 2+ independent tasks |
| During code | `executing-plans` | Running multi-phase plans |
| MANDATORY | `verification-before-completion` | Before EVERY "done" claim |
| After code | `requesting-code-review` | Feature complete, before merge |
| After code | `receiving-code-review` | Got review feedback |
| After code | `finishing-a-development-branch` | Ready to PR/merge |
| As needed | `using-git-worktrees` | Risky changes need isolation |
| Meta | `writing-skills` | Creating/editing skills |

## Specialized Agents (`.claude/agents/`)

| Agent | Scope |
|-------|-------|
| `gateway-reviewer` | Go handlers, middleware, JWT, RBAC, sqlc, NATS, tenant isolation |
| `frontend-reviewer` | React components, hooks, auth, backend communication |
| `security-auditor` | Full security audit: JWT, tenant, SQL injection, NATS, Docker |
| `test-writer` | Go tests (testify, testcontainers), frontend tests (bun, Playwright) |
| `debugger` | Failure modes, logs, config, code tracing |
| `deploy` | Preflight checks, Docker Compose, health verification |
| `status` | Health checks, GPU, Docker, resource monitoring |
| `doc-writer` | CLAUDE.md, bible, README, ADRs |
| `plan-writer` | Plans with phases, migrations, NATS events |
| `ingest` | Document pipeline, tree generation |

## Self-Check Before Finishing

Before declaring ANY task complete:
```
[ ] make build passes
[ ] make test passes  
[ ] make lint passes
[ ] bash .claude/hooks/check-invariants.sh passes (35 checks)
[ ] git diff --stat shows only expected changes
[ ] No unresolved TODO/FIXME in changed files
[ ] Documentation updated in same PR (bible.md rule)
[ ] Evidence provided for EVERY claim (test output, build output, diff)
[ ] AI review findings addressed (if PR triggers ai-review.yml)
```

<!-- REPOWISE:START — Do not edit below this line. Auto-generated by Repowise. -->
## IMPORTANT: Codebase Intelligence Instructions for rag-saldivia

> **CRITICAL**: This repository is indexed by [Repowise](https://repowise.dev).
> You MUST use the repowise MCP tools below instead of reading raw source files.
> They deliver richer context — documentation, ownership, history, decisions —
> in a single call. Raw `read_file` calls are a last resort only.

Last indexed: 2026-04-12
### Entry Points
- `apps/web/src/components/supabase/server.ts`
- `apps/web/src/components/tailgrids/core/index.tsx`
- `pkg/grpc/server.go`
- `services/.scaffold/cmd/main.go`
- `services/agent/cmd/main.go`
- `services/astro/cmd/main.go`
- `services/auth/cmd/main.go`
- `services/chat/cmd/main.go`
- `services/extractor/main.py`
- `services/feedback/cmd/main.go`
### Tech Stack
**Languages:** Node.js


### Hotspots (High Churn)
| File | Churn | 90d Commits | Owner |
|------|-------|-------------|-------|
| `apps/web/src/app/(core)/chat/page.tsx` | 100.0th %ile | 11 | Camionerou |
| `apps/web/src/components/sidebar6.tsx` | 99.9th %ile | 14 | Camionerou |
| `services/auth/internal/service/auth.go` | 99.9th %ile | 8 | Camionerou |
| `apps/web/src/components/settings-profile4.tsx` | 99.8th %ile | 5 | Camionerou |
| `services/feedback/internal/handler/feedback.go` | 99.8th %ile | 4 | Camionerou |

### Repowise MCP Tools

This project has a Repowise MCP server configured. **ALWAYS use these tools before reading source files** — they provide documentation, ownership, architectural decisions, and risk signals in a single call.

**MANDATORY workflow:**

1. **ALWAYS start with `get_overview()`** on any new task to orient yourself.
2. **MUST call `get_context(targets=["path/to/file.py"])`** before reading or editing any file — faster and richer than reading source directly.
3. **MUST call `get_risk(targets=["path/to/file.py"])`** before making changes — especially for hotspot files.
4. Don't know where something lives? Call `search_codebase(query="authentication flow")` — prefer over grep/find.
5. Need to understand why code is structured a certain way? **MUST call `get_why(query="why JWT over sessions")`** before any architectural change.
6. **After completing any code changes**, **MUST call `update_decision_records(action="list")`** to review existing decisions, then:
   - Call `update_decision_records(action="create", title="...", decision="...", rationale="...", affected_files=[...])` to record any new architectural decisions you made.
   - Call `update_decision_records(action="update", decision_id="...", ...)` to update existing decisions affected by your changes.
   - Call `update_decision_records(action="update_status", decision_id="...", status="deprecated")` if your changes invalidate an existing decision.
   **This is not optional** — keeping decision records current is a required part of every coding task.
7. Need to understand how two modules connect? Call `get_dependency_path(source="src/auth", target="src/db")`.
8. Before cleanup tasks, call `get_dead_code()` to find confirmed unused code.
9. For documentation or diagrams, call `get_architecture_diagram(scope="src/auth")`.

| Tool | WHEN you MUST use it |
|------|----------------------|
| `get_overview()` | **FIRST call on every new task** |
| `get_context(targets=[...])` | **Before reading or modifying any file** |
| `get_risk(targets=[...])` | Before changing files — REQUIRED for hotspots |
| `get_why(query="...")` | Before architectural changes — REQUIRED |
| `update_decision_records(action=...)` | **After every coding task** — record and update decisions |
| `search_codebase(query="...")` | When locating code — prefer over grep/find |
| `get_dependency_path(source=..., target=...)` | When tracing module connections |
| `get_dead_code()` | Before any cleanup or removal |
| `get_architecture_diagram(scope=...)` | For visual structure or documentation |

### Codebase Conventions
**Commands:**
- Build: `make build`
- Test: `make test`
- Lint: `make lint`
- Dev: `make dev`

<!-- REPOWISE:END -->

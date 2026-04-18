---
name: prompt-layers
description: Use when editing or adding any prompt layer — company system prompt, area prompts, user prompts, or memory entries — or when changing how the agent assembles context for a conversation. Owns the Phase 2 "hierarchical prompts" gate of ADR 027.
---

# prompt-layers

Scope: the stacked prompt layers the agent uses before the user's
message in every conversation. Files under
`services/app/internal/rag/agent/prompts/` (to be created), the
`erp_memories_{global,user}` tables (to be created), and the
assembly logic in `services/app/internal/rag/agent/service/agent.go`.

## The four-layer stack

Every agent invocation composes the context in a fixed order, each
layer with its own editing rules and token budget:

```
┌──────────────────────────────────────────────┐ ← system layer (repo-owned)
│ system.md   company overview, jargon,        │
│             org shape, policies, non-nego-   │
│             tiables                          │
├──────────────────────────────────────────────┤ ← area layer (repo-owned, per area)
│ area/<slug>.md   area-specific context:      │
│                  typical workflows, SLA,     │
│                  terminology for this area   │
├──────────────────────────────────────────────┤ ← user layer (per-user, DB)
│ users.prompt     the employee's custom       │
│                  prompt, editable by them    │
├──────────────────────────────────────────────┤ ← memories (curated, DB)
│ memories_global  + memories_user             │
│ ranked by relevance to the query + recency   │
├──────────────────────────────────────────────┤ ← chat history (conversation so far)
│ recent_messages  up to T tokens              │
├──────────────────────────────────────────────┤ ← user query (this message)
│ user's current input                         │
└──────────────────────────────────────────────┘
```

## Layer ownership + editing

| Layer | Who edits | Where it lives | How it ships |
|---|---|---|---|
| **system.md** | dev (repo) | `services/app/internal/rag/agent/prompts/system.md` | commit + deploy |
| **area/\*.md** | dev (repo), reviewed by area lead | `services/app/internal/rag/agent/prompts/area/<slug>.md` | commit + deploy |
| **user prompt** | the user | DB `users.prompt` column (or equivalent) | UI or chat command |
| **memories_global** | memory curator agent + admin UI | DB `erp_memories_global` | background job |
| **memories_user** | memory curator agent + user themselves | DB `erp_memories_user` | background job + UI |

### File loading mechanism

`system.md` and `area/*.md` ship as part of the binary via `//go:embed`:

```go
//go:embed prompts/system.md prompts/area/*.md
var promptsFS embed.FS
```

This removes runtime disk-access fragility and makes the rollback story
trivial (revert the commit → redeploy → old prompts back). It also makes
the build fail loudly if a referenced file is missing, which is the
verification path for the token-budget + completeness tests below.

The DB-owned layers (`user prompt`, `memories_*`) load per-request via
the sqlc-generated repo layer — same pattern as any other tenant data.

### Build-time prompt validation

`make check-prompts` (to be added; same pattern as `make events-validate`)
runs a script that:

1. Walks the embedded prompts + every area slug in the `areas` config.
2. Counts tokens per layer (use a fixed tokenizer so the count is
   stable across machines — recommend `tiktoken` via a small Go wrapper,
   or a dev dependency call-out).
3. Fails with non-zero exit and a per-file line count if any layer
   exceeds its declared budget (see §Token budgets below).

CI runs this step before the Go build, so a prompt-over-budget PR cannot
merge. The validation test that enforces this lives under
`services/app/internal/rag/agent/prompts/prompts_test.go` (golden +
budget).

## Token budgets

The assembly caps each layer so none starves the others. Default
budgets (tunable via `agent_config`):

| Layer | Default budget |
|---|---|
| system.md | 1500 tokens |
| area/<slug>.md | 1000 tokens |
| user prompt | 500 tokens |
| memories (global + user, merged) | 1500 tokens |
| chat history | 6000 tokens |
| user query | 2000 tokens (hard cap) |
| **total before model response** | ~12 500 tokens |

Budgets are enforced by truncation on the **least important** chunk
first (oldest chat message, lowest-ranked memory). System + area never
truncate — if they exceed their budget, the build fails loudly (the
dev needs to edit the source).

## system.md — the company-wide prompt

One file. Rare changes (company evolution, major process shift).
Contents:

- Who the company is (Saldivia Buses), what it does, scale.
- Org shape (areas, roles, key systems).
- Jargon: "cochecarrocería", "cotizacion vs presupuesto", "OP",
  "cotizopciones", internal abbreviations. The agent must know these.
- Non-negotiables: agent NEVER promises a price/delivery date without
  a system read; NEVER mixes tenants; NEVER bypasses permissions; etc.
- Linkage: a short statement of what the system is (SDA, not Histrix),
  what the agent's role is (user's representative).

## area/<slug>.md — per-area context

One file per business area that has its own workflow quirks. Suggested
set (starting list, grow as needed):

```
area/
  ingenieria.md
  compras.md
  produccion.md
  pcp.md                # planning-control-production
  rrhh.md
  calidad.md
  mantenimiento.md
  ventas.md
  tesoreria.md
  contabilidad.md
  seguridad-laboral.md
```

Each contains:

- Typical workflows (step-by-step).
- Tables/views this area touches most.
- Common mistakes + how to prevent them.
- SLA expectations.
- Area-specific terminology refinement.

The agent picks the `area.md` from the user's assigned area (1:1 in
most cases; multi-area users get multiple `area.md`s merged, with the
budget divided).

## User prompt

Each user has a `prompt` field (start from a template, fully editable).
Stored in the DB (not the repo).

- UI: a text editor page "Personalizar mi asistente".
- Chat command: `"actualizame el prompt: <content>"` → tool call to
  update the user row.
- Default template `user_template.md` seeds new users. Editable via
  PR (repo-owned).

## Memories (Phase 2, not yet implemented)

Two tables:

```sql
erp_memories_global (
  id uuid,
  content text,
  source text,           -- "curator" | "admin"
  confidence int,        -- 0-100
  area text nullable,
  tags text[],
  created_at, updated_at
)

erp_memories_user (
  id uuid,
  user_id uuid,
  content text,
  source text,           -- "curator" | "user"
  area text nullable,
  tags text[],
  created_at, updated_at
)
```

### Split of responsibility with `background-agents`

`background-agents` skill owns the **write** side of memories (curator
agent scans chat logs, extracts candidate memories, writes rows).
`prompt-layers` (this skill) owns the **read** side:

| Operation | Owner | Where |
|---|---|---|
| Candidate extraction (LLM call) | `background-agents` | curator agent goroutine |
| Idempotency (content hash dedup) | `background-agents` | curator's write path |
| Confidence scoring + pending queue | `background-agents` | curator policy |
| Admin review UI | `background-agents` + `frontend-next` | admin page |
| Retrieval + ranking at query time | `prompt-layers` | agent's context assembly |
| Budget enforcement per layer | `prompt-layers` | assembly code |
| ACL filter (per-user visibility) | `prompt-layers` + `auth-security` | retrieval query |

The concrete interface between the two skills is the shared repo layer
(`sqlc`-generated queries on `erp_memories_*`). Curator does
`Upsert...`, retrieval does `List...ByRelevance`. Neither skill is
allowed to bypass the repo — no ad-hoc `db.Exec` writes from the
curator, no ad-hoc reads from the agent.

### Retrieval flow

At query time `prompt-layers` assembles the memories layer by:

1. Compute the query embedding (reuse the RAG embedding path).
2. Fetch top-K candidate memories filtered by:
   - ACL: user's area, global memories they have the role for.
   - Recency: last T days (configurable).
   - Tag match: soft boost if any tag matches query terms.
3. Rank by weighted (embedding similarity + recency + tag match).
4. Truncate to the per-layer token budget.
5. Log the selected memories for audit (alongside the rest of the
   assembled prompt).

Global memories below `confidence<80` stay in a pending queue (never
retrieved) until an admin approves them — that's the curator's
responsibility per `background-agents`.

## Assembly + logging

The agent logs every assembled prompt in full, with per-layer token
counts, to `trace_events` or a dedicated `agent_prompt_log` table.
This is the only way to debug "why did the agent hallucinate X" — it
answers "was X even in context?".

Hash the resulting prompt so conversations with identical context can
be deduplicated in audit views.

## Versioning + rollout

- **system.md** and **area/\*.md** are versioned in git — deploy = new
  version in container.
- A rollback is a git revert + deploy.
- When a change is risky (e.g., new policy), ship behind a feature
  flag that selects old vs new at the tenant level, for one release
  cycle.
- Memory schema changes go through `database` + `migration-health`.

## Testing

- **Golden tests**: a set of representative user queries + expected
  assembled prompt shape. Fails if a layer silently empties.
- **Budget test**: forces a system.md > budget → build fails with a
  clear message.
- **Memory ACL test**: a user without a capability does not see
  memories gated by that capability.

## Integration with ADR 027

This skill owns the Phase 2 "hierarchical prompts" items. Every
session that lands system.md / area.md / user.md infrastructure ticks
an item; the memory tables tick the "memories curator" item in
partnership with `background-agents`.

## Don't

- Don't put tenant-specific content in system.md. The repo is
  single-tenant at the code level (ADR 022) but the prompt is also
  single-tenant — no if-else.
- Don't let area.md grow into a manual. Operational context only; long
  docs go to the RAG tree where retrieval picks what's relevant.
- Don't let the user's prompt leak into system.md edits by conflation.
  Users who want a change to company-wide policy open an issue.
- Don't inject memories the user can't read. ACL applies before the
  prompt is assembled.
- Don't skip logging the assembled prompt. It's the only forensic
  artifact when the agent goes off-rails.

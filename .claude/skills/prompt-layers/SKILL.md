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

Curator agent (see `background-agents`) scans chat logs nightly,
extracts candidate memories, writes them. Global memories need admin
review before use (flagged `confidence<80` defaults to pending).

Retrieval: at query time, the top-K memories (by recency + tag match +
embedding similarity) are injected into the memories layer, up to the
budget.

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

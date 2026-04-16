---
title: AI: Specialized Agents
audience: ai
last_reviewed: 2026-04-15
related:
  - ./skills.md
  - ./hooks.md
---

## Purpose

Catalog of specialized Claude agents in `.claude/agents/`. Each agent is a
prompt-tuned reviewer or operator with a narrow scope. Invoke by name when
the task matches the scope. Detail (checklists, tool lists, output format)
lives in each agent's own file — link below.

## Agents

| Agent | Scope (one line) | Invoke when |
|-------|------------------|-------------|
| [gateway-reviewer](../../.claude/agents/gateway-reviewer.md) | Code review of Go microservices: handlers, middleware, JWT, RBAC, sqlc, NATS, tenant isolation | Changes in `services/*/internal/`, `pkg/middleware/`, `pkg/jwt/`, `pkg/nats/`, `pkg/tenant/` |
| [frontend-reviewer](../../.claude/agents/frontend-reviewer.md) | Code review of Next.js frontend: components, hooks, auth, backend communication | Changes in `apps/web/` |
| [security-auditor](../../.claude/agents/security-auditor.md) | Full security audit: auth surface, tenant isolation, SQL injection, NATS subjects, Docker, header spoofing | Before releases or on suspicion of vulnerability |
| [test-writer](../../.claude/agents/test-writer.md) | Write Go (testify, testcontainers) and frontend (bun, Playwright) tests; commits incrementally per file | New code without coverage; explicit "add tests for X" |
| [debugger](../../.claude/agents/debugger.md) | Root-cause debugging by failure-mode table → logs → config → code | Something broken or unexpected — never for review |
| [deploy](../../.claude/agents/deploy.md) | Run preflight (build/test/lint/compose-config), execute `deploy/scripts/deploy.sh`, verify versions | Promoting a build to production |
| [status](../../.claude/agents/status.md) | Snapshot health: container state, Go service `/health`, infra liveness, NATS, GPU, disk | "Is everything up?" — read-only check |
| [doc-writer](../../.claude/agents/doc-writer.md) | Update human-authored docs (CLAUDE.md, READMEs, architecture) by reading code first | Structural change, new convention, stale CLAUDE.md |
| [plan-writer](../../.claude/agents/plan-writer.md) | Author plans under `docs/plans/` with phases, migrations, NATS events, scope control | Non-trivial new feature, before implementation |
| [ingest](../../.claude/agents/ingest.md) | Drive document ingestion through `services/ingest/` or directly into the RAG Blueprint | Adding documents to a tenant's knowledge base |
| [doc-sync](../../.claude/agents/doc-sync.md) | Edit modular docs under `docs/` from a code→doc queue produced by the pre-commit hook; fail-closed at 200 lines | Invoked automatically by `.claude/hooks/doc-sync.sh` |

## Invocation

Agents run via the Claude Code subagent system. From the main session:

```
Use the {agent-name} agent to {task}.
```

Each agent has its own `tools`, `model`, and `permissionMode` set in its
frontmatter. Do not override unless necessary.

## Coordination

- Found a bug while reviewing → file it; do not fix in the review session.
- Need cross-cutting work → main session dispatches multiple agents in
  parallel (`dispatching-parallel-agents` skill).
- Output of review agents goes to `docs/artifacts/{context}-{agent}.md`.

## Adding or removing an agent

1. Create or delete the file under `.claude/agents/`.
2. Update the table above.
3. Update `CLAUDE.md` and `.claude/CLAUDE.md` agent tables.
4. The doc-sync hook routes `.claude/agents/*.md` changes to this file.

---
title: AI: Memory System
audience: ai
last_reviewed: 2026-04-15
related:
  - ./agents.md
  - ../README.md
---

## Purpose

How Claude persists context across sessions. Two layers: the global
`MEMORY.md` for cross-session feedback, and per-agent memory directories
for specialist context. Read this before saving anything to memory.

## Layers

### Global memory — `MEMORY.md`

Path: `~/.claude/projects/-home-enzo-rag-saldivia/memory/MEMORY.md`.

Loaded at session start. Holds short feedback notes that should influence
every session, indexed as bullet links to per-topic files in the same
directory (`feedback_*.md`, `project_*.md`, `reference_*.md`, `user_*.md`).

### Per-agent memory — `agent-memory/`

Path: `.claude/agent-memory/{agent-name}/`. One directory per agent that
opts in (currently: `debugger`, `doc-writer`, `frontend-reviewer`,
`gateway-reviewer`, `plan-writer`, `security-auditor`, `test-writer`).

Holds notes the agent wrote during prior runs — patterns it learned,
gotchas it hit, file paths worth remembering. Loaded only when that
agent is invoked.

## What to save

- **Persistent preferences** — "always commit incrementally", "use Spanish
  in plans".
- **Hard-won gotchas** — "pgtype.Timestamptz needs Postgres format, not
  RFC3339".
- **Stable project facts** — "tenant slug 'saldivia' merges old saldivia +
  saldivia_sa".
- **Decisions with rationale** — link to a `decisions/` ADR or paste the
  one-liner.

## What NOT to save

- **Version-specific entries** — anything that changes with the next
  release ("Plan 24 status: in progress", "running on 2.0.5"). These
  belong in `docs/CHANGELOG.md` or `docs/plans/`.
- **Code excerpts** — code lives in code; memory points to it by path.
- **Secrets** — never. Memory is a plaintext file.
- **One-shot context** — if it only matters for the current task, don't
  save it.

## Hygiene rule

Memory must be **true today**. When an entry stops being true, delete it
or rewrite it. Stale memory is worse than no memory because it actively
misleads the next session.

Audit triggers:
- Major release (e.g., 2.0.x → 2.1.x) — sweep version-specific entries.
- Convention change documented in `docs/conventions/` — reflect it in memory.
- Project rename or restructure — update paths.

## Adding an entry

1. Create `feedback_{slug}.md` (or `project_`, `reference_`, `user_`) in
   the memory directory.
2. Add a bullet to `MEMORY.md` linking it: ``- [Title](slug.md) — one-line description``.
3. Keep each file under ~30 lines. If it's longer, it's probably a doc,
   not a memory entry — file under `docs/` instead.

## Removing an entry

1. Delete the bullet from `MEMORY.md`.
2. Delete the slug file (no archiving — git is the archive).
3. If the rule still matters but is now permanent, move it to
   `docs/conventions/` or `docs/ai/` instead of memory.

## Subagent inheritance

A dispatched subagent inherits the global `MEMORY.md` plus its own
`agent-memory/{name}/` directory. It does **not** see the parent
session's transient context. Save anything the subagent must know to
its agent-memory directory before dispatch.

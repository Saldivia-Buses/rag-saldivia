---
title: AI: Skills
audience: ai
last_reviewed: 2026-04-15
related:
  - ./agents.md
  - ./hooks.md
---

## Purpose

Catalog of skills under `.claude/skills/`. Skills are reusable workflows
loaded into the Claude session. Each is self-describing — read the
linked file for its full instructions. This page is the index.

## Superpowers (`.claude/skills/superpowers/`)

The 14 core workflow skills. Priority indicates when in a task they belong.

| Priority | Skill | When |
|----------|-------|------|
| ALWAYS | [using-superpowers](../../.claude/skills/superpowers/using-superpowers.md) | Session start — establishes the framework |
| Before code | [brainstorming](../../.claude/skills/superpowers/brainstorming.md) | New service, module, or feature |
| Before code | [writing-plans](../../.claude/skills/superpowers/writing-plans.md) | Spec is clear, multi-step task |
| During code | [test-driven-development](../../.claude/skills/superpowers/test-driven-development.md) | Every implementation — test first |
| During code | [systematic-debugging](../../.claude/skills/superpowers/systematic-debugging.md) | Bug, failure, unexpected behavior |
| During code | [subagent-driven-development](../../.claude/skills/superpowers/subagent-driven-development.md) | Per-task subagent dispatch with TDD + review |
| During code | [dispatching-parallel-agents](../../.claude/skills/superpowers/dispatching-parallel-agents.md) | 2+ independent tasks |
| During code | [executing-plans](../../.claude/skills/superpowers/executing-plans.md) | Running multi-phase plan in session |
| MANDATORY | [verification-before-completion](../../.claude/skills/superpowers/verification-before-completion.md) | Before claiming any task done |
| After code | [requesting-code-review](../../.claude/skills/superpowers/requesting-code-review.md) | Feature complete, before merge |
| After code | [receiving-code-review](../../.claude/skills/superpowers/receiving-code-review.md) | Got review feedback |
| After code | [finishing-a-development-branch](../../.claude/skills/superpowers/finishing-a-development-branch.md) | Ready to PR / merge |
| As needed | [using-git-worktrees](../../.claude/skills/superpowers/using-git-worktrees.md) | Risky changes need isolation |
| Meta | [writing-skills](../../.claude/skills/superpowers/writing-skills.md) | Creating or editing a skill |

## Other skills

| Skill | What it does |
|-------|--------------|
| [architecture-diagram](../../.claude/skills/architecture-diagram/) | Generate a dark-themed standalone HTML+SVG architecture diagram |

## Invocation

Skills are referenced by name in prompts and as `/command` slash commands.
The session-start hook activates `using-superpowers` automatically; the
other skills load on demand when their trigger conditions match.

## Adding or editing a skill

1. Create or edit the file under `.claude/skills/`.
2. Update the table above (superpowers or other).
3. The doc-sync hook routes `.claude/skills/**/*.md` changes to this file.

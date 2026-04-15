---
title: AI: Hooks
audience: ai
last_reviewed: 2026-04-15
related:
  - ./invariants.md
  - ./agents.md
---

## Purpose

Catalog of automated hooks registered in `.claude/settings.json` and
implemented under `.claude/hooks/`. Hooks fire on Claude Code lifecycle
events. Read this to know what runs when, and what blocks vs warns.

## Registered hooks (`.claude/settings.json`)

| Event | Matcher | Script | Behavior |
|-------|---------|--------|----------|
| `SessionStart` | `startup` | [session-briefing.sh](../../.claude/hooks/session-briefing.sh) | Prints recent commits, files modified in 48h, uncommitted changes, service versions |
| `PreToolUse` | `Edit`, `Write` | [pre-edit-check.sh](../../.claude/hooks/pre-edit-check.sh) | Warns (never blocks) if the target file was modified in the last 5 commits — read the diff first |
| `PreToolUse` | `Bash` (filtered: `git commit*`) | [pre-commit-check.sh](../../.claude/hooks/pre-commit-check.sh) | Runs the 35 invariant checks; **exit 2 blocks the commit** |
| `PreToolUse` | `Bash` (filtered: `git commit*`) | [doc-sync.sh](../../.claude/hooks/doc-sync.sh) | Computes a code→doc queue from staged files, surfaces it to the session, blocks if any target doc exceeds 200 lines |
| `Stop` | — | [stop-verify.sh](../../.claude/hooks/stop-verify.sh) | Runs invariants again (non-blocking) and emits a verification context |
| `Stop` | — | Inline Haiku prompt | If code was changed, asks Claude to confirm evidence (build/test/lint output) was provided |

## Helper scripts (not registered)

| Script | Use |
|--------|-----|
| [check-invariants.sh](../../.claude/hooks/check-invariants.sh) | The 35 invariant checks themselves; called by the pre-commit and stop hooks. See [invariants.md](./invariants.md) |
| [smart-test.sh](../../.claude/hooks/smart-test.sh) | Maps changed files (via `test-file-mapping.json`) to relevant Go test packages and runs only those |
| [doc-check.sh](../../.claude/hooks/doc-check.sh) | Standalone post-commit warning when code changes ship without doc changes (bible rule #10) — currently not registered in `settings.json` |

## Blocking vs warning

- **Blocking (exit 2)**: `pre-commit-check.sh`, `doc-sync.sh` (only on
  doc overflow). The commit is rejected; fix and re-stage.
- **Warning (exit 0 with `additionalContext`)**: `pre-edit-check.sh`,
  `stop-verify.sh`, `doc-check.sh`. Surface info; don't block.

## Bypassing

`git commit --no-verify` skips both pre-commit hooks. Do not use unless
explicitly authorized — every invariant exists because it caught a real
regression.

## Adding or editing a hook

1. Create or edit the script under `.claude/hooks/`.
2. Register it in `.claude/settings.json` if it should run on a Claude
   Code event.
3. Update the table above.
4. The doc-sync hook routes `.claude/hooks/*.sh` changes to this file.

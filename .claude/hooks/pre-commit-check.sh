#!/usr/bin/env bash
# SDA Framework — Pre-Commit Invariant Check (Claude Code hook)
# Event: PreToolUse on Bash, filtered by `if: "Bash(git commit*)"`
# Runs the 19+ invariant checks. Exit 2 = block commit.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

# Run invariant checks
if ! bash "$ROOT/.claude/hooks/check-invariants.sh" 2>&1; then
    echo "INVARIANT CHECK FAILED — commit blocked. Fix violations first." >&2
    exit 2  # exit 2 = block the tool call
fi

exit 0

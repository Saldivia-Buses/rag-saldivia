#!/usr/bin/env bash
# SDA Framework — Stop Verification Hook
# Runs invariant checks when Claude stops. Non-blocking (exit 0).
# Provides results as context for Claude to self-assess.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

echo "Stop verification — running invariant checks..."
echo ""

# Run invariants (capture output, don't block)
RESULT=$(bash "$ROOT/.claude/hooks/check-invariants.sh" 2>&1) || true
PASS=$(echo "$RESULT" | grep -c "✓" 2>/dev/null || true)
PASS=${PASS:-0}
FAIL=$(echo "$RESULT" | grep -c "✗" 2>/dev/null || true)
FAIL=${FAIL:-0}

# Check for uncommitted changes
DIRTY=$(git -C "$ROOT" status --short 2>/dev/null | wc -l | tr -d ' ')

# Detect if source code was changed (not just CI/config files)
SOURCE_CHANGED=$(git -C "$ROOT" diff --name-only HEAD~1 2>/dev/null \
  | grep -cE '\.(go|ts|tsx|js|jsx|py)$' 2>/dev/null || true)
SOURCE_CHANGED=${SOURCE_CHANGED:-0}

if [ "$SOURCE_CHANGED" -gt 0 ]; then
  VERIFY_MSG="Source code was changed — verify make build/test/lint passed and show evidence."
else
  VERIFY_MSG="Only config/CI files changed — syntax validation is sufficient (no build/test/lint needed)."
fi

cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "Stop",
    "additionalContext": "Stop verification: $PASS invariants passed, $FAIL failed. $DIRTY uncommitted files. $VERIFY_MSG"
  }
}
EOF

exit 0

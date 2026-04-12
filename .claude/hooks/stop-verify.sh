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
PASS=$(echo "$RESULT" | grep -c "✓" || echo 0)
FAIL=$(echo "$RESULT" | grep -c "✗" || echo 0)

# Check for uncommitted changes
DIRTY=$(git -C "$ROOT" status --short 2>/dev/null | wc -l)

cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "Stop",
    "additionalContext": "Stop verification: $PASS invariants passed, $FAIL failed. $DIRTY uncommitted files. If you made code changes, verify make build/test/lint passed and show evidence."
  }
}
EOF

exit 0

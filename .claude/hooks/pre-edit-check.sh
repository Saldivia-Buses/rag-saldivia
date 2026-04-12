#!/usr/bin/env bash
# SDA Framework — Pre-Edit Regression Check (Claude Code hook)
# Event: PreToolUse on Edit|Write
# Input: JSON on stdin with tool_input.file_path
# Output: informational warning on stdout (never blocks)
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

# Read hook input from stdin (Claude Code protocol)
INPUT=$(cat)
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty' 2>/dev/null || true)
[ -z "$FILE" ] && exit 0

# Check Go files, SQL, proto, deploy, and frontend config
case "$FILE" in
    */services/*/internal/*.go|*/pkg/*.go) ;;
    */db/queries/*.sql|*/db/migrations/*.sql) ;;
    */proto/*.proto) ;;
    */deploy/*.yml|*/deploy/*.yaml|*/deploy/*.sh) ;;
    */apps/web/src/*.ts|*/apps/web/src/*.tsx) ;;
    *) exit 0 ;;
esac

REL_FILE="${FILE#$ROOT/}"

# Check if file was modified in last 5 commits
RECENT=$(git -C "$ROOT" log --oneline -5 --follow -- "$REL_FILE" 2>/dev/null | head -5)
if [ -n "$RECENT" ]; then
    # Output goes to Claude as additionalContext
    cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "additionalContext": "WARNING: $REL_FILE was recently modified.\nRecent commits:\n$(echo "$RECENT" | sed 's/"/\\"/g' | tr '\n' '|' | sed 's/|/\\n/g')\nRead the recent diff before editing: git diff HEAD~5..HEAD -- $REL_FILE"
  }
}
EOF
fi

exit 0

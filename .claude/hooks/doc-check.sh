#!/bin/bash
# Check if code changed without doc update in same commit
CHANGED=$(git diff --name-only HEAD~1 HEAD 2>/dev/null)
[ -z "$CHANGED" ] && exit 0

CODE_CHANGED=false
DOC_CHANGED=false

echo "$CHANGED" | grep -qE '^(services/|pkg/|deploy/)' && CODE_CHANGED=true
echo "$CHANGED" | grep -qE '^(docs/|.*README)' && DOC_CHANGED=true

if [ "$CODE_CHANGED" = true ] && [ "$DOC_CHANGED" = false ]; then
  echo '{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":"WARNING: Code changed without documentation update. Bible rule #10: docs updated in same PR as code."}}'
fi

#!/usr/bin/env bash
# SDA Framework — Session Briefing (Claude Code hook)
# Event: SessionStart (startup|resume)
# Output on stdout becomes context for Claude.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$ROOT"

echo "SDA Framework — Session Briefing ($(date '+%Y-%m-%d %H:%M'))"
echo ""
echo "Branch: $(git branch --show-current)"
echo "Ahead of upstream: $(git rev-list --count experimental/ultra-optimize..HEAD 2>/dev/null || echo 'N/A') commits"
echo ""

echo "Last 10 commits:"
git log --oneline -10 --format="  %h %s (%ar)"
echo ""

echo "Files modified in last 48h:"
recent=$(git log --since="48 hours ago" --name-only --pretty=format: | sort -u | grep -v '^$' | head -20)
if [ -z "$recent" ]; then
    echo "  (none)"
else
    echo "$recent" | sed 's/^/  /'
fi
echo ""

echo "Uncommitted changes:"
status=$(git status --short 2>/dev/null)
if [ -z "$status" ]; then
    echo "  (clean working tree)"
else
    echo "$status" | head -15 | sed 's/^/  /'
    count=$(echo "$status" | wc -l)
    [ "$count" -gt 15 ] && echo "  ... and $((count - 15)) more"
fi
echo ""

echo "Service versions:"
for v in services/*/VERSION; do
    svc=$(basename "$(dirname "$v")")
    ver=$(cat "$v" | tr -d '[:space:]')
    printf "  %-16s v%s\n" "$svc" "$ver"
done

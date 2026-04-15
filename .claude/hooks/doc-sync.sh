#!/usr/bin/env bash
# SDA Framework — Modular docs sync hook
# Event: PreToolUse on Bash, filtered by `if: "Bash(git commit*)"`
# Behavior: scan staged files, compute queue of target docs, dispatch doc-sync agent
# Exit 0 = continue commit. Exit 2 = block (doc overflow).
#
# Skipped when `git commit --no-verify` is used (the hook itself doesn't run).
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$ROOT"

# Collect staged files (added/modified only, no deletes)
STAGED=$(git diff --cached --name-only --diff-filter=AM 2>/dev/null || true)
if [[ -z "$STAGED" ]]; then
    exit 0
fi

# Build queue: source file -> target doc(s)
declare -a QUEUE=()

while IFS= read -r file; do
    case "$file" in
        services/*/internal/*.go|services/*/cmd/*.go)
            svc=$(echo "$file" | cut -d/ -f2)
            if [[ -f "docs/services/$svc.md" ]]; then
                QUEUE+=("$file -> docs/services/$svc.md")
            fi
            # Handlers may affect architecture/flows
            if [[ "$file" == *"handler"* ]]; then
                case "$svc" in
                    auth) QUEUE+=("$file -> docs/flows/login-jwt.md") ;;
                    ws)   QUEUE+=("$file -> docs/architecture/websocket-hub.md") ;;
                    chat|agent) QUEUE+=("$file -> docs/flows/chat-agent-pipeline.md") ;;
                    ingest|extractor) QUEUE+=("$file -> docs/flows/document-ingestion.md") ;;
                    healthwatch) QUEUE+=("$file -> docs/flows/self-healing-triage.md") ;;
                esac
            fi
            ;;
        pkg/*)
            pkg=$(echo "$file" | cut -d/ -f2)
            if [[ -f "docs/packages/$pkg.md" ]]; then
                QUEUE+=("$file -> docs/packages/$pkg.md")
            fi
            ;;
        db/*/migrations/*.sql)
            QUEUE+=("$file -> docs/conventions/migrations.md")
            ;;
        .claude/agents/*.md)
            QUEUE+=("$file -> docs/ai/agents.md")
            ;;
        .claude/hooks/*.sh)
            QUEUE+=("$file -> docs/ai/hooks.md")
            ;;
        .claude/skills/**/*.md)
            QUEUE+=("$file -> docs/ai/skills.md")
            ;;
        deploy/*|.github/workflows/deploy.yml)
            QUEUE+=("$file -> docs/flows/deploy-pipeline.md")
            QUEUE+=("$file -> docs/operations/deploy.md")
            ;;
    esac
done <<< "$STAGED"

if [[ ${#QUEUE[@]} -eq 0 ]]; then
    exit 0  # no code changes mapped to docs
fi

# Deduplicate queue
IFS=$'\n' QUEUE=($(printf '%s\n' "${QUEUE[@]}" | sort -u))

# Emit queue for the agent (hook dispatch not implemented in bash; Claude Code picks this up)
echo "──────────────────────────────────────────────────"
echo "  doc-sync: ${#QUEUE[@]} target doc(s) may be stale"
echo "──────────────────────────────────────────────────"
for entry in "${QUEUE[@]}"; do
    echo "  $entry"
done
echo ""
echo "Tip: run the doc-sync agent after this commit to refresh:"
echo "  claude -p 'run doc-sync on last commit'"
echo ""

# Fail-closed check: verify no target doc exceeds 200 lines currently
OVERFLOW=0
for entry in "${QUEUE[@]}"; do
    doc=$(echo "$entry" | awk -F' -> ' '{print $2}')
    if [[ -f "$doc" ]]; then
        lines=$(wc -l < "$doc")
        if [[ "$lines" -gt 200 ]]; then
            echo "ERROR: $doc is $lines lines (max 200)" >&2
            OVERFLOW=1
        fi
    fi
done

if [[ "$OVERFLOW" -eq 1 ]]; then
    echo "" >&2
    echo "  Commit blocked: at least one target doc exceeds 200 lines." >&2
    echo "  Rewrite the offending doc before committing." >&2
    exit 2
fi

exit 0

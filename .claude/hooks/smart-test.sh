#!/usr/bin/env bash
# SDA Framework — Smart Test Runner
# Reads git diff, maps changed files to test packages, runs only relevant tests.
# Usage: bash .claude/hooks/smart-test.sh [base-ref]
#   base-ref: git ref to diff against (default: HEAD)
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
BASE="${1:-HEAD}"
MAPPING="$ROOT/.claude/hooks/test-file-mapping.json"

if [ ! -f "$MAPPING" ]; then
    echo "ERROR: test-file-mapping.json not found" >&2
    exit 1
fi

# Get changed files
CHANGED=$(git diff --name-only "$BASE" 2>/dev/null || git diff --name-only --cached)
if [ -z "$CHANGED" ]; then
    echo "No changed files detected."
    exit 0
fi

echo "═══════════════════════════════════════════════"
echo "  SDA Framework — Smart Test Runner"
echo "═══════════════════════════════════════════════"
echo ""
echo "▸ Changed files:"
echo "$CHANGED" | sed 's/^/  /'
echo ""

# Collect unique test packages
TESTS=""
RUN_INVARIANTS=false
RUN_SQLC=false

while IFS= read -r file; do
    # Try each mapping pattern
    while IFS= read -r mapping; do
        pattern=$(echo "$mapping" | jq -r '.pattern')
        # Convert glob to regex
        regex=$(echo "$pattern" | sed 's/\./\\./g; s/\*/.*/g')

        if echo "$file" | grep -qP "$regex"; then
            while IFS= read -r test_pkg; do
                if [ "$test_pkg" = "_invariants_only" ]; then
                    RUN_INVARIANTS=true
                elif [ "$test_pkg" = "_sqlc_regen" ]; then
                    RUN_SQLC=true
                else
                    TESTS="$TESTS $test_pkg"
                fi
            done < <(echo "$mapping" | jq -r '.tests[]')
        fi
    done < <(jq -c '.mappings[]' "$MAPPING")
done <<< "$CHANGED"

# Deduplicate
TESTS=$(echo "$TESTS" | tr ' ' '\n' | sort -u | tr '\n' ' ')

# Run invariants if migrations changed
if $RUN_INVARIANTS; then
    echo "▸ Migrations changed — running invariant checks..."
    bash "$ROOT/.claude/hooks/check-invariants.sh" || exit 1
    echo ""
fi

# Warn if sqlc queries changed
if $RUN_SQLC; then
    echo "▸ WARNING: sqlc queries changed — run 'make sqlc' to regenerate"
    echo ""
fi

# Run matched tests
if [ -n "$TESTS" ]; then
    echo "▸ Running tests for changed packages:"
    echo "$TESTS" | tr ' ' '\n' | sed 's/^/  /'
    echo ""

    FAILED=0
    for pkg in $TESTS; do
        echo "  Testing $pkg ..."
        if ! go test "$pkg" -count=1 -timeout 60s 2>&1 | tail -3; then
            FAILED=$((FAILED + 1))
        fi
    done

    echo ""
    if [ $FAILED -gt 0 ]; then
        echo "  ✗ $FAILED package(s) FAILED"
        exit 1
    else
        echo "  ✓ All matched tests passed"
    fi
else
    echo "▸ No test mappings found for changed files."
    echo "  Consider running full suite: make test"
fi

echo ""
echo "═══════════════════════════════════════════════"

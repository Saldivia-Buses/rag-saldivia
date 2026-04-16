#!/usr/bin/env bash
# Run the API + E2E smoke suites against a deployed SDA instance.
#
# Usage:
#   scripts/test-workstation.sh                     # default: http://172.22.100.23
#   TARGET=https://sda.app scripts/test-workstation.sh
#   scripts/test-workstation.sh api                 # API only
#   scripts/test-workstation.sh e2e                 # Playwright only
#
# Requires: bun, the test user (db/tenant/migrations/053_e2e_test_user.up.sql)
# applied on the target.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
TARGET="${TARGET:-http://172.22.100.23}"
SUITE="${1:-all}"

API_DIR="$ROOT_DIR/apps/web/e2e/api"
E2E_DIR="$ROOT_DIR/apps/web/e2e/workstation"
E2E_CFG="$E2E_DIR/playwright.config.ts"

API_PASS=0; API_FAIL=0
E2E_PASS=0; E2E_FAIL=0

print_header() {
    echo ""
    echo "═══════════════════════════════════════════════════════"
    echo "  $1"
    echo "═══════════════════════════════════════════════════════"
}

run_api() {
    print_header "API smoke (TARGET=$TARGET)"
    cd "$ROOT_DIR/apps/web"
    if TARGET="$TARGET" bun test e2e/api/ 2>&1 | tee /tmp/sda-api-test.log; then
        API_PASS=$(grep -cE "^\s*✓" /tmp/sda-api-test.log || echo 0)
    else
        API_PASS=$(grep -cE "^\s*✓" /tmp/sda-api-test.log || echo 0)
        API_FAIL=$(grep -cE "^\s*✗|^\s*\(fail\)" /tmp/sda-api-test.log || echo 0)
        return 1
    fi
}

run_e2e() {
    print_header "E2E smoke (TARGET=$TARGET)"
    cd "$ROOT_DIR/apps/web"
    if TARGET="$TARGET" bunx playwright test --config "$E2E_CFG" 2>&1 | tee /tmp/sda-e2e-test.log; then
        E2E_PASS=$(grep -cE "^\s*✓" /tmp/sda-e2e-test.log || echo 0)
    else
        E2E_PASS=$(grep -cE "^\s*✓" /tmp/sda-e2e-test.log || echo 0)
        E2E_FAIL=$(grep -cE "^\s*✘|FAIL" /tmp/sda-e2e-test.log || echo 0)
        return 1
    fi
}

OVERALL=0

case "$SUITE" in
    api)  run_api  || OVERALL=1 ;;
    e2e)  run_e2e  || OVERALL=1 ;;
    all)  run_api  || OVERALL=1; run_e2e || OVERALL=1 ;;
    *)    echo "unknown suite: $SUITE (use api|e2e|all)" >&2; exit 2 ;;
esac

print_header "Summary"
printf "  API smoke: %d passed, %d failed\n" "$API_PASS" "$API_FAIL"
printf "  E2E smoke: %d passed, %d failed\n" "$E2E_PASS" "$E2E_FAIL"

exit $OVERALL

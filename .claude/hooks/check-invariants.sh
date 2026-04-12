#!/usr/bin/env bash
# SDA Framework — Architectural Invariant Checks
# Runs structural checks that verify the project's integrity.
# Called by pre-commit hook and verification-before-completion skill.
#
# Exit code: 0 = all invariants hold, 1 = violation found

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ERRORS=0
CHECKS=0
PASSED=0

red()   { printf "\033[31m%s\033[0m\n" "$1"; }
green() { printf "\033[32m%s\033[0m\n" "$1"; }
yellow(){ printf "\033[33m%s\033[0m\n" "$1"; }

check() {
    CHECKS=$((CHECKS + 1))
    local name="$1"
    shift
    if "$@" 2>/dev/null; then
        PASSED=$((PASSED + 1))
        green "  ✓ $name"
    else
        ERRORS=$((ERRORS + 1))
        red "  ✗ $name"
    fi
}

echo ""
echo "═══════════════════════════════════════════════"
echo "  SDA Framework — Architectural Invariants"
echo "═══════════════════════════════════════════════"
echo ""

# ─── 1. Go Workspace Sync ──────────────────────────────────────────
echo "▸ Go Workspace"

check "go.work lists all Go services" bash -c '
    cd "'"$ROOT"'"
    for svc in services/*/go.mod; do
        dir=$(dirname "$svc")
        # skip extractor (Python) and scaffold (template)
        [[ "$dir" == *extractor* || "$dir" == *scaffold* ]] && continue
        grep -q "./$dir" go.work || { echo "MISSING in go.work: $dir"; exit 1; }
    done
'

check "go.work lists pkg/" bash -c '
    grep -q "./pkg" "'"$ROOT"'/go.work"
'

check "go.work lists tools/" bash -c '
    grep -q "./tools/cli" "'"$ROOT"'/go.work" && grep -q "./tools/mcp" "'"$ROOT"'/go.work"
'

# ─── 2. Migration Pairs ────────────────────────────────────────────
echo ""
echo "▸ Migration Pairs"

check "tenant migrations: every .up.sql has .down.sql" bash -c '
    cd "'"$ROOT"'"
    for up in db/tenant/migrations/*.up.sql; do
        down="${up%.up.sql}.down.sql"
        [ -f "$down" ] || { echo "MISSING: $down"; exit 1; }
    done
'

check "platform migrations: every .up.sql has .down.sql" bash -c '
    cd "'"$ROOT"'"
    for up in db/platform/migrations/*.up.sql; do
        down="${up%.up.sql}.down.sql"
        [ -f "$down" ] || { echo "MISSING: $down"; exit 1; }
    done
'

check "migration numbers are sequential (tenant)" bash -c '
    cd "'"$ROOT"'"
    prev=0
    for f in db/tenant/migrations/*.up.sql; do
        num=$(basename "$f" | grep -oP "^\d+")
        num=$((10#$num))  # strip leading zeros
        expected=$((prev + 1))
        if [ $prev -ne 0 ] && [ $num -ne $expected ]; then
            echo "GAP: expected $expected, got $num"
            exit 1
        fi
        prev=$num
    done
'

check "migration numbers are sequential (platform)" bash -c '
    cd "'"$ROOT"'"
    prev=0
    for f in db/platform/migrations/*.up.sql; do
        num=$(basename "$f" | grep -oP "^\d+")
        num=$((10#$num))
        expected=$((prev + 1))
        if [ $prev -ne 0 ] && [ $num -ne $expected ]; then
            echo "GAP: expected $expected, got $num"
            exit 1
        fi
        prev=$num
    done
'

# ─── 3. Service Structure ──────────────────────────────────────────
echo ""
echo "▸ Service Structure"

check "every Go service has cmd/main.go" bash -c '
    cd "'"$ROOT"'"
    for svc in services/*/go.mod; do
        dir=$(dirname "$svc")
        [[ "$dir" == *scaffold* ]] && continue
        [ -f "$dir/cmd/main.go" ] || { echo "MISSING: $dir/cmd/main.go"; exit 1; }
    done
'

check "every Go service has VERSION file" bash -c '
    cd "'"$ROOT"'"
    for svc in services/*/go.mod; do
        dir=$(dirname "$svc")
        [[ "$dir" == *scaffold* || "$dir" == *extractor* ]] && continue
        [ -f "$dir/VERSION" ] || { echo "MISSING: $dir/VERSION"; exit 1; }
    done
'

check "VERSION files contain valid semver" bash -c '
    cd "'"$ROOT"'"
    for v in services/*/VERSION; do
        ver=$(cat "$v" | tr -d "[:space:]")
        echo "$ver" | grep -qP "^\d+\.\d+\.\d+$" || { echo "INVALID: $v = $ver"; exit 1; }
    done
'

check "every Go service has Dockerfile" bash -c '
    cd "'"$ROOT"'"
    for svc in services/*/go.mod; do
        dir=$(dirname "$svc")
        [[ "$dir" == *scaffold* ]] && continue
        [ -f "$dir/Dockerfile" ] || { echo "MISSING: $dir/Dockerfile"; exit 1; }
    done
'

# ─── 4. sqlc Sync ──────────────────────────────────────────────────
echo ""
echo "▸ sqlc Configuration"

check "every service with db/queries/*.sql has sqlc.yaml" bash -c '
    cd "'"$ROOT"'"
    for qdir in services/*/db/queries; do
        # skip if only .gitkeep or empty
        sql_count=$(find "$qdir" -name "*.sql" 2>/dev/null | wc -l)
        [ "$sql_count" -eq 0 ] && continue
        svc=$(dirname "$(dirname "$qdir")")
        [ -f "$svc/db/sqlc.yaml" ] || { echo "MISSING: $svc/db/sqlc.yaml"; exit 1; }
    done
'

check "sqlc.yaml points to correct query dir" bash -c '
    cd "'"$ROOT"'"
    for cfg in services/*/db/sqlc.yaml; do
        grep -q "queries" "$cfg" || { echo "BAD CONFIG: $cfg"; exit 1; }
    done
'

# ─── 5. sqlc Generated Code Freshness ─────────────────────────────
echo ""
echo "▸ sqlc Freshness"

check "sqlc queries newer than generated code are flagged" bash -c '
    cd "'"$ROOT"'"
    for cfg in services/*/db/sqlc.yaml; do
        svc_db=$(dirname "$cfg")
        svc=$(dirname "$svc_db")
        repo_dir="$svc/internal/repository"
        [ -d "$repo_dir" ] || continue
        # Find newest .sql file and newest generated .sql.go file
        newest_sql=$(find "$svc_db/queries" -name "*.sql" -printf "%T@\n" 2>/dev/null | sort -rn | head -1)
        newest_gen=$(find "$repo_dir" -name "*.sql.go" -printf "%T@\n" 2>/dev/null | sort -rn | head -1)
        [ -z "$newest_sql" ] || [ -z "$newest_gen" ] && continue
        # If queries are newer than generated code, sqlc needs regen
        newer=$(echo "$newest_sql $newest_gen" | awk "{if (\$1 > \$2) print \"stale\"}")
        [ -z "$newer" ] || { echo "STALE SQLC in $svc: queries newer than generated code — run make sqlc"; exit 1; }
    done
'

# ─── 6. Tenant Isolation Patterns ──────────────────────────────────
echo ""
echo "▸ Tenant Isolation"

check "no hardcoded tenant IDs in Go source" bash -c '
    cd "'"$ROOT"'"
    # Look for UUID-like hardcoded tenant IDs in service code (not test files)
    found=$(grep -rn "tenant.*=.*\"[0-9a-f]\{8\}-[0-9a-f]\{4\}" services/*/internal/ --include="*.go" \
        | grep -v "_test.go" | grep -v "example" | grep -v "// " | head -5) || true
    [ -z "$found" ] || { echo "HARDCODED TENANT: $found"; exit 1; }
'

check "handlers use tenant middleware (spot check)" bash -c '
    cd "'"$ROOT"'"
    # Every service with handlers should import tenant or middleware package
    for hdir in services/*/internal/handler; do
        svc=$(basename "$(dirname "$(dirname "$hdir")")")
        [[ "$svc" == "platform" || "$svc" == "extractor" || "$svc" == "ws" ]] && continue
        grep -rlq "tenant\|middleware" "$hdir/" || { echo "NO TENANT CHECK: $svc"; exit 1; }
    done
'

# ─── 6. Security Patterns ──────────────────────────────────────────
echo ""
echo "▸ Security Patterns"

check "no secrets in Go source files" bash -c '
    cd "'"$ROOT"'"
    found=$(grep -rn "password\s*=\s*\"[^\"]\+\"" services/*/internal/ --include="*.go" \
        | grep -v "_test.go" | grep -v "config\." | grep -v "param" | grep -v "field" | head -3) || true
    [ -z "$found" ] || { echo "POSSIBLE SECRET: $found"; exit 1; }
'

check "no .env files committed" bash -c '
    cd "'"$ROOT"'"
    ! git ls-files --cached | grep -qP "^\.env$|^\.env\.local$|^\.env\.production$"
'

# ─── 7. Docker Compose ─────────────────────────────────────────────
echo ""
echo "▸ Docker Compose"

check "docker-compose.prod.yml exists" bash -c '
    [ -f "'"$ROOT"'/deploy/docker-compose.prod.yml" ]
'

check "every Go service has container in compose" bash -c '
    cd "'"$ROOT"'"
    for svc in services/*/go.mod; do
        name=$(basename "$(dirname "$svc")")
        [[ "$name" == "scaffold" || "$name" == "extractor" ]] && continue
        # platform may not be in compose yet
        # TODO: add platform + bigbrother to compose when deployed
        [[ "$name" == "platform" || "$name" == "bigbrother" ]] && continue
        grep -q "sda-$name\|sda_$name\|$name:" deploy/docker-compose.prod.yml || \
            { echo "MISSING in compose: $name"; exit 1; }
    done
'

# ─── 8. Proto Sync ─────────────────────────────────────────────────
echo ""
echo "▸ Proto & Generated Code"

check "gen/go/ exists if proto/ has .proto files" bash -c '
    cd "'"$ROOT"'"
    if ls proto/**/*.proto >/dev/null 2>&1; then
        [ -d "gen/go" ] || { echo "MISSING: gen/go/"; exit 1; }
    fi
'

# ─── 9. NATS Subject Naming ───────────────────────────────────────
echo ""
echo "▸ NATS Subject Convention"

check "NATS publishes use tenant.{slug} prefix" bash -c '
    cd "'"$ROOT"'"
    # Find NATS Publish calls that dont use tenant prefix
    bad=$(grep -rn "\.Publish(" services/*/internal/ --include="*.go" \
        | grep -v "_test.go" \
        | grep -v "tenant\." \
        | grep -v "// " \
        | grep -v "subject" | head -3) || true
    [ -z "$bad" ] || { echo "NON-TENANT NATS PUBLISH: $bad"; exit 1; }
'

check "NATS consumer subjects defined with tenant.* prefix" bash -c '
    cd "'"$ROOT"'"
    # Verify subject filter constants/vars include tenant.* pattern
    for consumer in $(grep -rl "FilterSubject\|Subjects:" services/*/internal/ --include="*.go" | grep -v _test.go); do
        svc_dir=$(echo "$consumer" | grep -oP "services/[^/]+")
        # Check that somewhere in this service there is a tenant.* subject definition
        grep -rq "tenant\.\*" "$svc_dir/internal/" --include="*.go" || \
            { echo "NO tenant.* SUBJECT in $svc_dir"; exit 1; }
    done
'

# ─── 10. Write→Event Consistency ──────────────────────────────────
echo ""
echo "▸ Write → NATS Event"

check "services with INSERT/UPDATE/DELETE have NATS Publish" bash -c '
    cd "'"$ROOT"'"
    for qdir in services/*/db/queries; do
        sql_count=$(find "$qdir" -name "*.sql" 2>/dev/null | wc -l)
        [ "$sql_count" -eq 0 ] && continue
        svc_dir=$(dirname "$(dirname "$qdir")")
        svc=$(basename "$svc_dir")
        # skip services that are read-only or platform-only
        [[ "$svc" == "platform" || "$svc" == "search" ]] && continue
        # check if any query has a write operation
        has_writes=$(grep -riclP "INSERT|UPDATE|DELETE" "$qdir/" 2>/dev/null) || has_writes=""
        [ -z "$has_writes" ] && continue
        # check if service has Publish calls OR broadcasts
        has_publish=$(grep -rl "Publish\|Broadcast\|publisher" "$svc_dir/internal/" --include="*.go" 2>/dev/null | grep -v _test.go | head -1) || has_publish=""
        [ -n "$has_publish" ] || { echo "NO NATS PUBLISH in $svc (has writes but no events)"; exit 1; }
    done
'

# ─── 11. Service Documentation ────────────────────────────────────
echo ""
echo "▸ Service Documentation"

check "every Go service has README.md" bash -c '
    cd "'"$ROOT"'"
    for svc in services/*/go.mod; do
        dir=$(dirname "$svc")
        [[ "$dir" == *scaffold* ]] && continue
        [ -f "$dir/README.md" ] || { echo "MISSING: $dir/README.md"; exit 1; }
    done
'

# ─── 11. Handler Patterns ─────────────────────────────────────────
echo ""
echo "▸ Handler Patterns"

check "handlers use MaxBytesReader for POST endpoints" bash -c '
    cd "'"$ROOT"'"
    # Every handler file with json.Decode should use MaxBytesReader
    for hfile in services/*/internal/handler/*.go; do
        [[ "$hfile" == *_test.go ]] && continue
        if grep -q "json.NewDecoder" "$hfile" 2>/dev/null; then
            grep -q "MaxBytesReader" "$hfile" || { echo "NO MaxBytesReader: $hfile"; exit 1; }
        fi
    done
'

check "http.Error calls use JSON format (not plain text)" bash -c '
    cd "'"$ROOT"'"
    # Verify http.Error calls use JSON strings, not plain text
    # Acceptable: http.Error(w, `{"error":"msg"}`, code)
    # Bad: http.Error(w, "some plain text", code)
    for hfile in services/*/internal/handler/*.go; do
        [[ "$hfile" == *_test.go || "$hfile" == *ws.go ]] && continue
        svc=$(echo "$hfile" | grep -oP "services/\K[^/]+")
        [[ "$svc" == "ws" || "$svc" == "bigbrother" ]] && continue
        # Find http.Error with plain text (not JSON backtick or double-quote JSON)
        bad=$(grep "http\.Error(" "$hfile" 2>/dev/null \
            | grep -v "json\|JSON\|{.*error\|{.*Error\|writeJSON" \
            | grep -v "^\s*//" | head -3) || true
        [ -z "$bad" ] || { echo "PLAIN TEXT ERROR in $hfile: $bad"; exit 1; }
    done
'

# ─── 12. Dockerfile Security ──────────────────────────────────────
echo ""
echo "▸ Dockerfile Security"

check "Dockerfiles use multi-stage build" bash -c '
    cd "'"$ROOT"'"
    for df in services/*/Dockerfile; do
        svc=$(basename "$(dirname "$df")")
        [[ "$svc" == "extractor" ]] && continue
        stages=$(grep -c "^FROM " "$df" 2>/dev/null) || stages=0
        [ "$stages" -ge 2 ] || { echo "SINGLE-STAGE: $df (need multi-stage)"; exit 1; }
    done
'

check "Dockerfiles run as non-root user" bash -c '
    cd "'"$ROOT"'"
    for df in services/*/Dockerfile; do
        svc=$(basename "$(dirname "$df")")
        [[ "$svc" == "extractor" ]] && continue
        grep -qi "USER\|nonroot\|nobody\|appuser\|65534" "$df" || \
            { echo "RUNS AS ROOT: $df"; exit 1; }
    done
'

# ─── 15. Repowise Index Staleness ─────────────────────────────────
echo ""
echo "▸ Repowise Index"

check "Repowise index is less than 3 days old" bash -c '
    cd "'"$ROOT"'"
    idx_date=$(grep -oP "Last indexed: \K[\d-]+" .claude/CLAUDE.md 2>/dev/null || echo "")
    [ -z "$idx_date" ] && exit 0  # no date found, skip
    idx_epoch=$(date -d "$idx_date" +%s 2>/dev/null || echo 0)
    now_epoch=$(date +%s)
    age_days=$(( (now_epoch - idx_epoch) / 86400 ))
    [ "$age_days" -le 3 ] || { echo "STALE: Repowise indexed $age_days days ago ($idx_date) — run repowise mcp to reindex"; exit 1; }
'

# ─── 16. Frontend Structure ───────────────────────────────────────
echo ""
echo "▸ Frontend"

check "apps/web has package.json" bash -c '
    [ -f "'"$ROOT"'/apps/web/package.json" ]
'

check "no hardcoded API URLs in frontend source" bash -c '
    cd "'"$ROOT"'"
    [ -d "apps/web/src" ] || exit 0
    bad=$(grep -rn "localhost:[0-9]\{4\}\|127\.0\.0\.1:[0-9]" apps/web/src/ --include="*.ts" --include="*.tsx" \
        | grep -v "// " | grep -v "_test" | grep -v ".test." | head -3) || true
    [ -z "$bad" ] || { echo "HARDCODED API URL: $bad"; exit 1; }
'

# ─── Summary ───────────────────────────────────────────────────────
echo ""
echo "═══════════════════════════════════════════════"
if [ $ERRORS -eq 0 ]; then
    green "  ALL $CHECKS INVARIANTS PASSED ✓"
else
    red "  $ERRORS/$CHECKS INVARIANTS FAILED ✗"
    echo "  ($PASSED passed, $ERRORS failed)"
fi
echo "═══════════════════════════════════════════════"
echo ""

exit $ERRORS

---
title: AI: Invariant Checks
audience: ai
last_reviewed: 2026-04-15
related:
  - ./hooks.md
  - ../conventions/migrations.md
---

## Purpose

The 35 architectural invariants enforced by
[.claude/hooks/check-invariants.sh](../../.claude/hooks/check-invariants.sh).
Run by the pre-commit hook (blocking) and the Stop hook (informational).
Each check has exactly one job; failure of any check rejects the commit.

## Run

```bash
bash .claude/hooks/check-invariants.sh
```

## Checks by category

### Go workspace (3)

1. `go.work` lists all Go services
2. `go.work` lists `pkg/`
3. `go.work` lists `tools/cli` and `tools/mcp`

### Migration pairs (4)

4. Tenant migrations: every `.up.sql` has a matching `.down.sql`
5. Platform migrations: every `.up.sql` has a matching `.down.sql`
6. Migration numbers are sequential (tenant) — no gaps
7. Migration numbers are sequential (platform) — no gaps

### Service structure (4)

8. Every Go service has `cmd/main.go`
9. Every Go service has a `VERSION` file
10. `VERSION` files contain valid semver (`X.Y.Z`)
11. Every Go service has a `Dockerfile`

### sqlc configuration (2)

12. Every service with `db/queries/*.sql` has `db/sqlc.yaml`
13. `sqlc.yaml` points to the correct query directory

### sqlc freshness (1)

14. sqlc generated code is not older than its source `.sql` queries — run
    `make sqlc` if stale

### Tenant isolation (2)

15. No hardcoded tenant UUIDs in service code
16. Every handler package imports `tenant` or `middleware`

### Security patterns (2)

17. No `password = "..."` literals in service code
18. No `.env`, `.env.local`, or `.env.production` committed

### Docker compose (2)

19. `deploy/docker-compose.prod.yml` exists
20. Every Go service has a container entry in compose

### Proto (1)

21. `gen/go/` exists if `proto/` has any `.proto` file

### NATS subjects (2)

22. NATS publishes use the `tenant.{slug}` prefix
23. NATS consumers define subjects with the `tenant.*` prefix

### Write → event consistency (1)

24. Services with `INSERT/UPDATE/DELETE` queries also have `Publish`,
    `Broadcast`, or a publisher reference (every write emits an event)

### Service documentation (1)

25. Every Go service has a `README.md`

### Handler patterns (2)

26. Handlers using `json.NewDecoder` also use `http.MaxBytesReader`
27. `http.Error` calls use JSON format, not plain text

### Dockerfile security (2)

28. Dockerfiles use multi-stage build (≥2 `FROM` stages)
29. Dockerfiles declare a non-root `USER`

### Repowise (1)

30. Repowise index timestamp in `.claude/CLAUDE.md` is ≤ 3 days old

### Frontend (2)

31. `apps/web/package.json` exists
32. No hardcoded `localhost:PORT` or `127.0.0.1:PORT` in frontend source

### Docs ↔ code sync (1)

33. (removed — legacy doc was deleted; flow-doc code refs are reviewed by the doc-sync hook)

### Silent failure protection (2)

34. No swallowed errors (`_ = err` pattern) in handler files
35. No service file with > 8 bare `return ..., err` patterns (forces
    error wrapping with context)

## Output

Each check prints `✓` on pass and `✗` on fail. Final line:
`ALL N INVARIANTS PASSED ✓` or `M/N INVARIANTS FAILED ✗`. Exit code is
the failure count (0 = success).

## When a check is wrong

If a check produces a false positive that cannot be fixed by adjusting the
code: edit the check in `check-invariants.sh`, justify in the commit
message, and update this document. Never bypass with `--no-verify`.

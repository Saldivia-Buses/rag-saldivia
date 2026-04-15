# SDA Framework

Multi-tenant SaaS of Go microservices with AI services, per-industry modules.
Backend: inhouse workstation (RTX PRO 6000, 96GB VRAM). Frontend: Next.js on CDN.

**Remote:** https://github.com/Camionerou/rag-saldivia · **Branch:** `2.0.5`

---

## Read first

1. [`docs/README.md`](docs/README.md) — documentation index. Start there.
2. [`docs/architecture/overview.md`](docs/architecture/overview.md) — system map.
3. For the file you are about to edit: read the matching `docs/services/X.md` or `docs/packages/Y.md`.

Documentation is modular (≤200 lines per doc, AI-audience, English). Never edit `docs/plans/*` — they are historical records, not live docs.

---

## Key commands

```bash
make dev          # local dev stack
make test         # Go tests
make lint         # Go lint
make build        # build all services
make sqlc         # regenerate sqlc code after query change
make migrate      # run tenant + platform migrations
make deploy       # production deploy (GitHub Actions path)
make versions     # running vs available service versions
make status       # health of services + GPU
```

Service-specific: `make test-<svc>`, `make build-<svc>`.
New service: `make new-service NAME=<svc>`.

---

## Critical invariants (7 hard rules)

1. **Tenant isolation at every layer.** Every tenant query has `WHERE tenant_id = $1`. Never share a pool across tenants. See `docs/architecture/multi-tenancy.md`.
2. **JWT is the only source of identity.** Verify locally with ed25519. Never trust client headers. See `docs/architecture/auth-jwt.md`.
3. **NATS subjects are tenant-namespaced.** `tenant.{slug}.{service}.{entity}[.{action}]` always. See `docs/architecture/nats-events.md`.
4. **Every write publishes a NATS event.** No polling in frontend. See `docs/conventions/error-handling.md`.
5. **Migration pairs complete.** Every `.up.sql` has a matching `.down.sql`; numbering is sequential. See `docs/conventions/migrations.md`.
6. **Service structure is uniform.** Every Go service: `cmd/main.go` + `VERSION` + `Dockerfile` + `README.md`, registered in `go.work`.
7. **Error responses are JSON.** Never plain text from API handlers. See `docs/conventions/error-handling.md`.

Full list + checks: `bash .claude/hooks/check-invariants.sh` (42 checks).

---

## Agents, skills, hooks, MCP

- [`docs/ai/agents.md`](docs/ai/agents.md) — specialized agents (`gateway-reviewer`, `frontend-reviewer`, `security-auditor`, `test-writer`, `debugger`, `deploy`, `status`, `doc-writer`, `doc-sync`, `plan-writer`, `ingest`).
- [`docs/ai/skills.md`](docs/ai/skills.md) — superpower skills.
- [`docs/ai/hooks.md`](docs/ai/hooks.md) — pre-commit invariants, session briefing, doc-sync.
- [`docs/ai/mcp-servers.md`](docs/ai/mcp-servers.md) — Context7, Repowise, CodeGraphContext, etc.
- [`docs/ai/memory-system.md`](docs/ai/memory-system.md) — durable memory; no version-specific entries.

---

## Conventions

- Go style, errors, ctx: [`docs/conventions/go.md`](docs/conventions/go.md)
- Frontend components, tokens, tests: [`docs/conventions/frontend.md`](docs/conventions/frontend.md)
- Git flow, commits, squash merge: [`docs/conventions/git.md`](docs/conventions/git.md)
- Migrations, sqlc, testing, logging, security: each in `docs/conventions/*.md`.

---

## Operations

- Deploy: [`docs/operations/deploy.md`](docs/operations/deploy.md)
- Runbook + incidents: [`docs/operations/runbook.md`](docs/operations/runbook.md), [`docs/operations/incidents.md`](docs/operations/incidents.md)
- Monitoring: [`docs/operations/monitoring.md`](docs/operations/monitoring.md)
- Backup & restore: [`docs/operations/backup-restore.md`](docs/operations/backup-restore.md)

---

## When in doubt

- Search: `docs/README.md` → relevant subfolder.
- Unfamiliar term: [`docs/glossary.md`](docs/glossary.md).
- Before editing shared code in `pkg/`: grep all importers first (`docs/conventions/go.md` blast-radius rule).

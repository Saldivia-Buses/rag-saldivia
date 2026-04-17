# SDA Framework

Go backend + Next.js frontend, deployed **one container per tenant** (silo model,
all-in-one image — see ADR 022 + ADR 023). Backend on inhouse workstation
(single-box vertical scale — see below). Frontend on CDN. No pool-tenant in the
code: the tenant **is** the container.

**Workstation (`srv-ia-01`, `172.22.100.23`):**
- **CPU:** AMD Threadripper Pro 9975WX (32c/64t)
- **RAM:** 256 GB DDR5 ECC
- **Storage:** 8 TB M.2 NVMe
- **GPU:** NVIDIA RTX PRO 6000 Blackwell Q Edition (96 GB VRAM)

Design implication: this is a **vertical-scale box**, not a cluster. Network latency between
services is artificial (they all run on one machine). Keep that in mind when reaching for
horizontal-scale patterns — most won't buy you anything here.

**Remote:** `github.com/Camionerou/rag-saldivia` · **Main:** `main` · **Working:** `2.0.6`

---

## Layout

```
apps/web/          Next.js app (frontend)
services/          14 Go microservices: agent, auth, bigbrother, chat, erp,
                   extractor, feedback, healthwatch, ingest, notification,
                   platform, search, traces, ws
pkg/               32 shared Go packages (jwt, nats, tenant, httperr, middleware, …)
deploy/            docker-compose, Traefik, scripts
tools/cli/         `sda` CLI (migrations, admin)
docs/plans/        historical plans — read-only, never edit
docs/decisions/    ADRs — mutated via the `decisions` skill
.claude/skills/    specialized skills (see below)
```

### Pinned stack

- **Go 1.25** · chi v5.2 · pgx v5.9 · nats.go v1.50 · golang-jwt/v5 · OTel v1.43 ·
  grpc v1.80 · testcontainers v0.42 · aws-sdk-go-v2 · redis/go-redis/v9
- **Next.js 16.2** · React 19.2 · Tailwind v4 (CSS-first) · Vercel AI SDK v6 ·
  TypeScript 5 · bun · shadcn / Radix / Base UI
- **Industrial:** gopcua (OPC UA), simonvetter/modbus, masterzen/winrm, pkg/sftp
- **Infra:** Postgres, NATS JetStream, Redis, MinIO (S3), Traefik, SGLang, Milvus (legacy)

---

## Commands

```bash
make dev            # infra only (Postgres/NATS/Redis/Traefik in Docker)
make dev-services   # all Go services on host
make dev-frontend   # Next.js dev server (localhost:3000 only)
make dev-all        # infra + services + frontend
make stop           # stop everything
make test           # Go tests
make lint           # Go lint
make build          # build all services
make sqlc           # regenerate sqlc after query change
make migrate        # platform + tenant migrations
make status         # service health + GPU
make versions       # running vs available service versions
make deploy         # production deploy
```

Per-service: `make test-<svc>`, `make build-<svc>`. New: `make new-service NAME=<svc>`.

---

## Non-negotiable rules (live in their skills)

The old "7 hard invariants" block lived here as written aspiration with no
enforcement. It duplicated what the skills already said and violated principle
#0 (reduce). The rules are not gone — they are enforced in the skill that owns
the scope, where Claude will read them before touching the code:

| Rule | Owner skill |
|---|---|
| JWT is the only identity (ed25519, server-side) | `auth-security` |
| Every write publishes a NATS event | `backend-go` |
| Migrations ship in up/down pairs, one sequence | `database` |
| Error responses are JSON | `backend-go` |
| No tenant plumbing in code (ADR 022) | `auth-security` + `database` |
| Reduce before add (ADR 021) | CLAUDE.md principle 0 + `continuous-improvement` |

If enforcement is ever needed back, it goes as a `make check` target (fast, real),
not a wall of text in CLAUDE.md.

---

## This harness is law

The skills, principles, invariants, and ADRs in this repo are not suggestions. They
are the single source of truth that governs every session. When a skill applies,
it is invoked and followed. When a principle contradicts a shortcut, the principle
wins. When an invariant is at risk, the change stops.

Claude's job is to push the project toward the target described here, relentlessly,
session after session, until the code matches the intent. No session is "done"
because time ran out — a session is done when the work meets the bar the harness
sets, or when the user explicitly overrides.

If the harness is wrong, fix the harness first. Don't work around it.

## Principles (apply everywhere)

0. **Reduce before adding.** The project is over-engineered for a one-dev
   vertical-scale box: ~89 k LOC of Go, 15 services, 32 packages. The direction is
   **consolidate, delete, unify** — not add. Every feature request, every refactor,
   every new package starts with the question: *can this be done by removing
   something instead?* When in doubt, delete. A session that ends with net-negative
   LOC and the same behavior is a successful session.
1. **Continuous improvement is the default.** Every session iterates an area — pick one,
   measure it, hunt for wins across correctness / simplicity / perf / clarity / tests,
   ship what clears the bar. If the user hasn't anchored the session to a task, this is
   the job. Karpathy-style auto-research.
2. **Simple beats clever.** Less code for the same behavior is better code. If a
   change adds machinery, justify it — or find a smaller way.
3. **Simple + functional = excellent.** The point is not minimalism for its own
   sake; it is the biggest behavioral win per line of code added.
4. **Scale > laziness, but scale the box you have.** Vertical-scale box (32c/256GB/96GB VRAM).
   Don't pick patterns that only pay off in a cluster. gRPC between services on the same
   kernel is a cost, not a feature.
5. **Reviews are load-bearing.** Findings become fixes, tracked work, or explicit
   waivers. Never "noted".
6. **Evidence before claims.** Tests pass, build green, invariants hold — show
   the output, don't assert it.

## The consolidation target

The north star for this codebase is this shape — not today, but every decision
should bend toward it:

- **From 14 services → 3-5.** Group by domain, not by noun. (See `continuous-improvement`.)
- **From 32 packages → ~10.** Delete unused ones, inline one-importer packages.
- **From ~89 k LOC → whatever it takes** to express the same product with fewer lines.
- **Question multi-tenant** at every opportunity. If it isn't a real product requirement,
  its scaffolding is the biggest single simplification available.

Never add a service, package, or abstraction without first trying to fit it inside
something that already exists.

---

## Skills — how Claude works here

Twelve scope-specialized skills live in `.claude/skills/`. Claude activates one (or more)
automatically based on what you are editing. Each skill is the entire contract for that
scope — read it before acting inside that scope.

| Skill | Activates when |
|---|---|
| `backend-go` | editing `services/**/*.go` or `pkg/**/*.go` |
| `frontend-next` | editing `apps/web/**` |
| `auth-security` | touching JWT, tenant isolation, RBAC, middleware, NATS namespacing |
| `database` | touching migrations, sqlc queries, Postgres schema |
| `rag-pipeline` | touching ingest/search/agent — tree-search, crossdoc, embeddings, Milvus |
| `deploy-ops` | touching deploy/, Dockerfiles, runbook, workstation, VPN |
| `infrastructure-access` | reaching any internal host (`172.22.x.x`) — ERP legacy, NAS, Proxmox, VoIP, PLC |
| `code-review` | reviewing any diff or PR — severity taxonomy + per-scope checklist |
| `decisions` | creating, updating, deprecating ADRs in `docs/decisions/` |
| `parallel-research` | a task has 2+ independent investigation fronts — dispatch subagents |
| `systematic-debugging` | anything unexpected: failing test, runtime error, weird behavior |
| `continuous-improvement` | default mode when no specific task is set — iterate an area |

**No custom subagents live on disk.** When you need parallel research, the
`parallel-research` skill tells Claude how to spin up one-shot `Explore` or
`general-purpose` agents on demand.

---

## Conventions — pointers inside skills

- Go style, errors, ctx propagation → `backend-go`
- Next.js components, server actions, tokens → `frontend-next`
- Git flow, commit style, squash-merge → `backend-go` (§ Commits)
- Migrations, sqlc, testing, logging → `database` + `backend-go`
- Deploy pipeline, runbook, incidents → `deploy-ops`

---

## Before editing any file in `services/` or `pkg/`

1. `git log --oneline -5 -- <file>` — recently touched?
2. If `pkg/*`: list importers before changing any exported symbol. Blast radius first.
3. Read the matching skill (above).
4. Make the change, build, test, check invariants.
5. If the change is architectural, update `docs/decisions/` via the `decisions` skill.

---

## Done checklist

- `make build && make test && make lint` pass.
- Diff is only what the task needs (no drive-by edits).
- No unresolved `TODO`/`FIXME` you introduced.
- Evidence for every claim (test output, build log, diff).
- If behavior changed, the relevant ADR is current.

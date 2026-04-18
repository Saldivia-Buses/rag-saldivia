# SDA Framework

**SDA replaces Histrix** (see ADR 026) — it is not a wrapper, not a RAG
overlay, not an "integration". When SDA covers the parity surface, the
Histrix server gets powered off. Everything else in this repo bends toward
that goal.

Go backend + Next.js frontend, deployed **one container per tenant** (silo
model, all-in-one image with frontend baked in — see ADR 022 + ADR 023 +
ADR 024). Runs on the inhouse workstation (single-box vertical scale — see
below). No pool-tenant in the code: the tenant **is** the container, and
the container carries the entire product (Go app + Next.js + Postgres +
NATS + Redis + MinIO) under `s6-overlay`.

**Workstation (`srv-ia-01`, `172.22.100.23`):**
- **CPU:** AMD Threadripper Pro 9975WX (32c/64t)
- **RAM:** 256 GB DDR5 ECC
- **Storage:** 8 TB M.2 NVMe
- **GPU:** NVIDIA RTX PRO 6000 Blackwell Q Edition (96 GB VRAM)

Design implication: this is a **vertical-scale box**, not a cluster.
Network latency between services is artificial (they all run on one
machine). Don't reach for horizontal-scale patterns — most won't buy you
anything here.

**Remote:** `github.com/Saldivia-Buses/rag-saldivia` · **Main:** `main` ·
**Working:** `2.0.6`

---

## North star (ADR 026)

The employee opens SDA and:

1. Has a modern UI covering **everything** Histrix did (1:1 operational
   parity, better UX).
2. Has a chat where the agent is their representative — "cargame esta
   factura" runs the same write path the UI would, scoped to their
   permissions. Chat ↔ UI capability parity.
3. Builds their own **personal dashboard** (no global dashboard exists).
4. Creates personal routines.
5. Behind the curtain, agents hoard data: mail ingest, internal
   WhatsApp ingest, file ingest via tree-RAG with per-collection ACL.

The parity yardstick is concrete and on disk: **`.intranet-scrape/`**.
- 676 tables (`db-tables.txt`)
- 434 XML-forms (`xml-forms/`) — the complete Histrix screen inventory
- PHP backend + JS frontend captures

Every "new ERP feature" starts by reading the relevant XML-form first.

## The four phases + phase 0 (ADR 026/027)

Sessions pick work **top-down**. Phase 0 failures block every phase.
Phase N+1 work is only valid once Phase N is green (or the item is
waived in an ADR).

### Phase 0 — Transversal (always checked)

- **Data integrity = religion.** Zero ghost rows. Every migrator:
  `rows_read == rows_written + rows_skipped`.
- **Tool security from day one.** Every agent tool declares capabilities;
  user permissions are checked before dispatch.
- **Prod = source of truth.** Workstation SHA == `main` HEAD. Drift
  closes before new work.

### Phase 1 — Histrix parity + shutdown (gating)

Every XML-form has an SDA equivalent (or a waiver). Zero-loss migration
verifiable end-to-end. Seamless-day test passes. Critical Histrix reports
reproducible in SDA. The Histrix server eventually powers off.

### Phase 2 — The SDA layer (what Histrix never was)

Chat as first-class UI with tool-based agent. Hierarchical prompts
(`system.md` → `area.md` → `user.md` → memories). Tree-RAG with
per-collection ACL. Granular tool permission model.

### Phase 3 — Background agents + data hoarding

Mail ingest agent. WhatsApp ingest agent. Memory curator agent.
Analytics + prediction on accumulated data.

### Phase 4 — Differential UX

Personal dashboards (no global). Personal routines. Deep per-user
customization.

**The complete verifiable checklist is ADR 027.** Read it at session
start. The first un-ticked item whose dependencies are green is the
next candidate.

---

## Layout

```
apps/web/                         Next.js app (frontend)
services/
  app/                            Go monolith (5 internal modules per ADR 025)
    internal/{core,rag,realtime,ops,erp,...}
  erp/                            ERP standalone (38k LOC, fusion deferred)
  extractor/                      Python sidecar (PDF OCR, SGLang client)
pkg/                              13 shared Go packages (was 24, down via ADRs 021/025)
deploy/                           docker-compose, s6, Traefik, Dockerfile.all-in-one
tools/cli/                        `sda` CLI (migrations, admin)
tools/eventsgen/                  CUE → Go/TS/Markdown event codegen
.intranet-scrape/                 Histrix scrape — the parity contract
docs/plans/                       historical plans — read-only
docs/decisions/                   ADRs — mutated via `decisions` skill
docs/parity/                      waivers + report equivalence log (Phase 1)
.claude/skills/                   specialized skills (see below)
```

### Pinned stack

- **Go 1.25** · chi v5.2 · pgx v5.9 · nats.go v1.50 · golang-jwt/v5 · OTel v1.43 ·
  testcontainers v0.42 · aws-sdk-go-v2 · redis/go-redis/v9
- **Next.js 16.2** · React 19.2 · Tailwind v4 (CSS-first) · Vercel AI SDK v6 ·
  TypeScript 5 · bun · shadcn / Radix / Base UI
- **Industrial:** gopcua (OPC UA), simonvetter/modbus, masterzen/winrm, pkg/sftp
- **Infra:** Postgres, NATS JetStream, Redis, MinIO (S3), Traefik, SGLang

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

---

## Non-negotiable rules

Live in their skills — read the skill before touching the scope.

| Rule | Owner skill |
|---|---|
| **Phase 0 gates (integrity, tool perms, prod drift)** | `continuous-improvement` + per-scope skill |
| JWT is the only identity (ed25519, server-side) | `auth-security` |
| Every write publishes a NATS event | `backend-go` |
| Migrations ship in up/down pairs, one sequence | `database` |
| Error responses are JSON | `backend-go` |
| No tenant plumbing in code (ADR 022) | `auth-security` + `database` |
| Every agent tool has a declared capability + perm check | `agent-tools` |
| Parity sessions consult `.intranet-scrape/` first | `htx-parity` |
| Migration changes verify `read = written + skipped` | `migration-health` |
| Reduce before add (ADR 021) | `continuous-improvement` |

---

## This harness is law

The skills, principles, ADRs, and the phased checklist (ADR 027) in
this repo are not suggestions. They are the single source of truth that
governs every session. When a skill applies, it is invoked and followed.
When a Phase 0 gate is failing, no Phase 2+ work ships. When an invariant
is at risk, the change stops.

Claude's job is to push the project toward the target in ADR 026,
relentlessly, session after session, until the code matches the intent.
A session is done when the shipped work ticks at least one ADR 027 item
AND Phase 0 is still green — not because time ran out, not because a PR
got long.

If the harness is wrong, **fix the harness first**. Don't work around it.

## Principles (apply everywhere)

0. **Parity before polish.** Phase 0 and Phase 1 of ADR 026 block any
   Phase 2+ work. A pretty dashboard with 0 invoice lines is wrong.
   Fix the missing data first.
1. **Reduce before adding.** Consolidate, delete, unify. A session that
   ends with net-negative LOC + same behavior + a Phase 1 tick is a
   great session.
2. **Continuous improvement is the default.** When the user doesn't
   anchor the session, read ADR 027 top-down, pick the first un-ticked
   un-blocked item, ship. See `continuous-improvement`.
3. **Simple beats clever.** Less code for the same behavior is better
   code. If a change adds machinery, justify it or find a smaller way.
4. **Scale the box you have.** Vertical-scale box (32c/256GB/96GB VRAM).
   Don't pick patterns that only pay off in a cluster.
5. **Reviews are load-bearing.** Findings become fixes, tracked work,
   or explicit waivers. Never "noted".
6. **Evidence before claims.** Tests pass, build green, Phase 0 gates
   hold — show the output, don't assert it.

## The shape target (architecture, not north star)

Architecture that makes the north star feasible (ADR 025):

- **Services:** 14 → 3-5. Today **3** (app, erp, extractor). ERP fusion
  is the remaining lever.
- **Packages:** 24 → ~10. Today **13**. The remaining 13 are all blocked
  by `services/erp` imports — progress gated on the ERP fusion.
- **Binary count inside the container:** 1 Go binary + 1 Next.js +
  infra. `deploy/frontdoor` dissolves with the ERP fusion.

Architecture is load-bearing but not the goal. The goal is the north
star; the shape makes it shippable by one dev.

---

## Skills — how Claude works here

Scope-specialized skills live in `.claude/skills/`. Claude activates one
(or more) automatically based on the task. Each skill is the entire
contract for that scope — read it before acting inside that scope.

| Skill | Activates when |
|---|---|
| `backend-go` | editing `services/**/*.go` or `pkg/**/*.go` |
| `frontend-next` | editing `apps/web/**` |
| `auth-security` | touching JWT, RBAC, middleware, NATS namespacing |
| `database` | touching migrations, sqlc queries, Postgres schema |
| `migration-health` | validating the Histrix → SDA migration (Phase 0/1) |
| `rag-pipeline` | touching the RAG stack (ingest/search/agent under `services/app/internal/rag/`) |
| `prompt-layers` | editing system/area/user prompts, memories, RAG context assembly |
| `agent-tools` | adding/changing agent tools, tool permission declarations |
| `background-agents` | designing mail, WhatsApp, memory curator agents |
| `htx-parity` | planning/implementing Phase 1 parity — consult `.intranet-scrape/` first |
| `deploy-ops` | touching deploy/, Dockerfiles, runbook, workstation, prod drift |
| `infrastructure-access` | reaching any internal host (`172.22.x.x`) |
| `code-review` | reviewing any diff or PR |
| `decisions` | creating, updating, deprecating ADRs |
| `parallel-research` | a task has 2+ independent investigation fronts |
| `systematic-debugging` | anything unexpected |
| `continuous-improvement` | default mode when no task is set — walks ADR 027 |

**No custom subagents live on disk.** When parallel research is needed,
`parallel-research` dispatches one-shot `Explore` or `general-purpose`
agents on demand.

---

## Before editing any file in `services/` or `pkg/`

1. `git log --oneline -5 -- <file>` — recently touched?
2. If `pkg/*`: list importers before changing any exported symbol.
3. If the change touches ERP data flow: read the corresponding XML-form
   in `.intranet-scrape/xml-forms/` first (parity contract).
4. Read the matching skill.
5. Make the change, build, test, verify Phase 0 gates still hold.
6. If architectural, update `docs/decisions/` via the `decisions` skill.
7. If it ticks an ADR 027 item, tick it in the same PR.

---

## Done checklist

- `make build && make test && make lint` pass.
- Diff is only what the task needs (no drive-by edits).
- No `TODO` / `FIXME` introduced.
- Evidence for every claim (test output, build log, diff).
- Phase 0 gates still green.
- ADR 027 tick (if applicable) is in the same PR.
- If architectural: ADR is current.

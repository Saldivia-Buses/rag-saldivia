# SDA Framework

## LEER PRIMERO

1. `docs/plans/2.0.x-plan01-sda-framework.md` — spec completo del sistema (la biblia)
2. `docs/bible.md` — reglas permanentes de trabajo

No empezar a trabajar sin leer estos documentos.

---

## Qué es este proyecto

Plataforma SaaS multi-tenant de microservicios Go con servicios de IA y
módulos de negocio por industria. Backend inhouse (RTX PRO 6000, 96GB VRAM),
frontends en la nube.

- **Repo:** `~/rag-saldivia/` — branch activa: `2.0.x`
- **Remoto:** https://github.com/Camionerou/rag-saldivia

---

## Arquitectura

```
CLOUD                                    INHOUSE (workstation)
┌──────────────┐                         ┌────────────────────────────────┐
│  Next.js     │                         │  Traefik (gateway)             │
│  Frontend    │──── REST/WS ──────────►│    ├─► Auth Service (Go)       │
│  (CDN)       │   Cloudflare Tunnel    │    ├─► WebSocket Hub (Go)     │
└──────────────┘                         │    ├─► Chat Service (Go)      │
                                         │    ├─► RAG Service (Go)       │
CLOUD (APIs)                             │    ├─► Notification (Go)      │
┌──────────────┐                         │    ├─► Platform (Go)          │
│  NVIDIA API  │◄── LLM queries ───────│    ├─► Ingest (Go)            │
│  Nemotron-49B│                         │    └─► [N modular services]   │
└──────────────┘                         │                                │
                                         │  PostgreSQL per-tenant        │
                                         │  Redis per-tenant             │
                                         │  NATS + JetStream             │
                                         │  Milvus + NIMs (RAG)          │
                                         └────────────────────────────────┘
```

---

## Stack

| Componente | Tecnologia |
|---|---|
| Backend | Go (chi + sqlc + slog) |
| Database | PostgreSQL per-tenant |
| Cache | Redis per-tenant + platform |
| Message broker | NATS + JetStream |
| Frontend | Next.js + React + shadcn/ui + Tailwind + TanStack Query |
| Gateway | Traefik |
| RAG | NVIDIA RAG Blueprint v2.5.0 |
| LLM | Nemotron-Super-49B via NVIDIA API |
| Agent framework | NeMo Agent Toolkit |
| Observability | OpenTelemetry + Grafana (Tempo + Prometheus + Loki) |
| CI/CD | GitHub Actions |
| CLI | Go (Cobra) |

---

## Estructura del repo

```
services/                    ← Go microservicios
  auth/                      ← Auth Gateway + RBAC + MFA
  ws/                        ← WebSocket Hub
  chat/                      ← Sesiones + mensajes
  rag/                       ← Proxy NVIDIA Blueprint
  notification/              ← In-app + email
  platform/                  ← Control de tenants (platform admins)
  ingest/                    ← Pipeline de documentos
  .scaffold/                 ← Template para make new-service

pkg/                         ← Go packages compartidos
  jwt/                       ← JWT validation local
  tenant/                    ← Tenant context, DB resolver
  middleware/                ← Auth, logging, tracing
  nats/                      ← NATS helpers
  security/                  ← Rate limiting, brute force
  config/                    ← Config loading

proto/                       ← Protobuf (gRPC contracts)

apps/
  web/                       ← Next.js frontend
  login/                     ← Login page aislada

ai/
  agents/                    ← NeMo Agent Toolkit configs
  guardrails/                ← NeMo Guardrails policies
  models/                    ← Model configs, VRAM profiles

modules/                     ← Modulos verticales por industria

tools/
  cli/                       ← CLI binario (sda)
  mcp/                       ← MCP Server para IA
  pkg/                       ← Logica compartida CLI + MCP

deploy/                      ← Docker Compose, Traefik, scripts
config/                      ← NVIDIA Blueprint configs
vendor/                      ← Blueprint submodule
docs/                        ← Documentacion
```

---

## Comandos clave

```bash
make dev                     # Levantar stack de desarrollo
make stop                    # Bajar servicios
make test                    # Tests Go
make test-auth               # Tests de un servicio especifico
make lint                    # Lint Go
make build                   # Build todos los servicios
make build-auth              # Build un servicio
make new-service NAME=x      # Scaffold servicio nuevo
make proto                   # Generar codigo gRPC
make sqlc                    # Generar codigo sqlc
make migrate                 # Correr migraciones
make deploy                  # Deploy a produccion
make versions                # Ver versiones running vs available
make status                  # Estado de servicios + GPU
```

---

## Convenciones

### Go

| Tipo | Convencion | Ejemplo |
|---|---|---|
| Packages | lowercase, single word | `handler`, `service` |
| Files | snake_case | `user_handler.go` |
| Structs | PascalCase | `UserService` |
| Interfaces | PascalCase, -er suffix | `UserRepository` |
| Functions | PascalCase/camelCase | `CreateUser()`, `hashPassword()` |
| Errors | Wrap con contexto | `fmt.Errorf("get user: %w", err)` |
| Context | Primer parametro siempre | `func (s *Svc) Get(ctx context.Context, id string)` |

### Git

- Branch: `main` protegida, feature branches con PR
- Commits: `tipo(servicio): descripcion` (lowercase)
- Tipos: `feat`, `fix`, `refactor`, `test`, `docs`, `ci`, `chore`
- Squash merge a main
- Docs actualizadas en el mismo PR que el codigo

### Frontend

- Componentes: PascalCase (`VehicleTable.tsx`)
- Hooks: camelCase con `use` (`useEnabledModules.ts`)
- Lib/utils: kebab-case (`module-guard.ts`)

---

## Archivos criticos

| Archivo | Por que |
|---|---|
| `docs/plans/2.0.x-plan01-sda-framework.md` | Spec completo — la biblia del sistema |
| `docs/bible.md` | Reglas permanentes de trabajo |
| `services/{name}/README.md` | Que hace cada servicio |
| `go.work` | Go workspace — modulos registrados |
| `Makefile` | Todos los comandos |
| `deploy/` | Docker Compose configs |
| `config/` | NVIDIA Blueprint configs |

---

## Agents disponibles (`.claude/agents/`)

| Agent | Cuando | Scope |
|---|---|---|
| `gateway-reviewer` | Cambios en `services/*/internal/`, `pkg/` | Handlers chi, middleware, JWT, RBAC, sqlc, NATS events, tenant isolation |
| `frontend-reviewer` | Cambios en `apps/web/`, `apps/login/` | Componentes, hooks, auth, comunicacion con backend Go |
| `security-auditor` | Antes de releases, sospecha de vulnerabilidad | Audit completo: JWT, tenant isolation, SQL injection, NATS, Docker |
| `test-writer` | Tests nuevos o faltantes | Go tests (testify, testcontainers), frontend tests (bun, Playwright) |
| `debugger` | Algo no funciona | Failure modes, logs Docker/Go, config, trazado de codigo |
| `deploy` | Deployar a produccion | Preflight checks, Docker Compose, health verification |
| `status` | Estado de servicios | Health checks Go services + infra Docker + GPU + recursos |
| `doc-writer` | Actualizar docs | CLAUDE.md, bible, README por servicio, ADRs |
| `plan-writer` | Planear feature nueva | Planes con phases, migrations, NATS events, scope control |
| `ingest` | Ingestar documentos | Pipeline ingesta, RAG Blueprint, Milvus |

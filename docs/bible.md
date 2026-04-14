# La Biblia — SDA Framework

> **Documento permanente.** Define como se trabaja en este repo.
> El spec completo del sistema esta en `docs/plans/2.0.x-plan01-sda-framework.md`.
> Este documento define reglas de trabajo, convenciones, y principios que no
> cambian con cada version.

---

## Principios

1. **Cuestiona el requerimiento antes de escribir codigo.** Preguntar "es realmente necesario?" antes de planificar.
2. **Si el plan es dificil de explicar, el plan esta mal.**
3. **Borra lo que no deberia existir. No optimices lo que queda hasta hacerlo.**
4. **Si nadie usaria una v1 rota, el scope esta mal.** Achicar el scope, no la ambicion.
5. **Cada paso debe saber lo que el paso anterior decidio.** (artifact persistence)
6. **Evidencia le gana a conviccion. Conviccion le gana a consenso.**
7. **Arreglalo o pregunta. Nunca lo ignores.**
8. **La seguridad no es un tradeoff. Es una restriccion.**
9. **El output deberia verse mejor de lo que se pidio.**
10. **Si cambias codigo, actualiza la doc en el mismo PR.**

---

## Producto

- **Nombre:** SDA Framework
- **Colores:** tokens de Claude, acento **azure blue**
- **Idioma UI:** espanol
- **Idioma codigo:** ingles (variables, commits, PRs, API docs)
- **Idioma planes:** espanol
- **Modelo de deploy:** multi-tenant SaaS, backend inhouse, frontend cloud
- **Repo:** https://github.com/Camionerou/rag-saldivia

---

## Stack

| Componente | Tecnologia |
|---|---|
| Backend | Go (microservicios) |
| HTTP framework | chi |
| DB access | sqlc |
| Database | PostgreSQL per-tenant |
| Cache | Redis per-tenant + platform |
| Message broker | NATS + JetStream |
| Auth | JWT (golang-jwt) + refresh tokens |
| Frontend | Next.js + React + shadcn/ui + Tailwind + TanStack Query |
| Gateway | Traefik |
| RAG | Tree reasoning (PageIndex-inspired, no vectors) |
| LLM | Model-agnostic via SGLang (slot-per-pipeline-step) |
| OCR | PaddleOCR-VL 1.5 via SGLang |
| Vision | Qwen3.5-9B via SGLang |
| Model server | SGLang (1 instance per model, OpenAI-compatible API) |
| Object storage | MinIO (S3-compatible, migrable a AWS S3) |
| Observability | OpenTelemetry + Grafana stack |
| Logger | slog (Go stdlib) |
| Testing | go test + testify + testcontainers |
| Testing frontend | bun:test + Playwright |
| CI/CD | GitHub Actions (4 gates + 6-phase deploy + AI review) |
| Monitoring | HealthWatch service (port 8014) + Alertmanager → notification |
| Feature flags | Platform service (deterministic rollout, kill switch) |
| CLI | Go (Cobra) |
| MCP | Go |

Para detalle completo del stack, ver el spec.

---

## Git workflow

### Branches

- `main` protegida, siempre deployable
- Feature branches cortas (1-3 dias): `feat/`, `fix/`, `refactor/`, `docs/`
- Todo cambio a main pasa por PR

### Commits

Formato: `tipo(servicio): descripcion` (lowercase)

```
feat(auth): add totp-based mfa flow
fix(chat): handle empty message array on session fork
refactor(rag): extract collection resolver to pkg/tenant
```

Tipos: `feat`, `fix`, `refactor`, `test`, `docs`, `ci`, `chore`

Scopes: `auth`, `chat`, `agent`, `search`, `astro`, `traces`, `ingest`, `extractor`,
`notification`, `ws`, `platform`, `feedback`, `web`, `cli`, `mcp`, `infra`,
`pkg`, `docs`, `deploy`

### Pull Requests

1. Feature branch desde `main`
2. Implementar + tests + docs (en el mismo PR)
3. Push → CI automatico (4 gates) + AI review (3 passes)
4. CI verde + AI review sin critical/high → squash merge
5. Post-merge: version bump + changelog + Docker image + GitHub Release

### CI Pipeline (4 gates secuenciales)

1. **Gate 1 — Verify:** commitlint, go build, go vet (paralelos)
2. **Gate 2 — Test:** go test (necesita Gate 1)
3. **Gate 3 — Security:** gosec + trivy (necesita Gate 1)
4. **Gate 4 — Docker:** build todas las imagenes (necesita Gates 2+3)

Service matrix generada dinamicamente desde `services/*/Dockerfile`.

### AI Review Gates (3 passes paralelos en cada PR)

1. **Quality:** logica, performance, invariantes SDA (Opus)
2. **Security:** auth boundaries, tenant isolation, injection, secrets (Opus)
3. **Dependencies:** vulnerabilidades, base images, migration pairs (Sonnet)

Output en JSON estructurado, scoring via `jq`. Critical/high bloquean merge.
`@claude` disponible en issues/PRs (restringido a org members).

### Quality gates

1. CI 4 gates verdes
2. AI review sin findings critical/high
3. `make test` local (unit + contract)
4. `make lint` (golangci-lint)
5. Docs actualizadas (checklist en PR)

---

## Convenciones Go

| Tipo | Convencion | Ejemplo |
|---|---|---|
| Packages | lowercase, single word | `handler`, `service`, `repository` |
| Files | snake_case | `user_handler.go` |
| Structs | PascalCase | `UserService` |
| Interfaces | PascalCase, -er suffix | `UserRepository` |
| Functions | PascalCase (exported), camelCase (internal) | `CreateUser()`, `hashPassword()` |
| Errors | Siempre devolver, nunca panic en handlers | `return nil, fmt.Errorf("get user: %w", err)` |
| Context | Primer parametro siempre | `func (s *Svc) Get(ctx context.Context, id string)` |
| Tests | Table-driven, archivo `_test.go` junto al codigo | `func TestCreateUser(t *testing.T)` |

### Estructura de cada servicio

```
services/{name}/
  cmd/main.go            ← Entrypoint
  internal/
    handler/             ← HTTP/gRPC handlers
    service/             ← Business logic (interfaces)
    repository/          ← DB access (sqlc generated)
  db/
    migrations/          ← SQL migrations (up + down)
    queries/             ← sqlc .sql files
    sqlc.yaml
  VERSION                ← Semver del servicio
  CHANGELOG.md           ← Autogenerado por git-cliff
  Dockerfile
  go.mod
  README.md              ← Que hace, endpoints, events
```

---

## Convenciones frontend

| Tipo | Convencion | Ejemplo |
|---|---|---|
| Componentes | PascalCase, `.tsx` | `VehicleTable.tsx` |
| Hooks | camelCase con `use` | `useEnabledModules.ts` |
| Lib/utils | kebab-case | `module-guard.ts` |
| Paginas | `page.tsx` (App Router) | `app/(modules)/fleet/page.tsx` |
| Core routes | `app/(core)/` | Siempre disponibles |
| Module routes | `app/(modules)/` | Code-split, lazy loaded |

---

## Documentacion

### Audiencia principal: agentes de IA

La doc esta escrita para que un agente que abra el repo por primera vez
entienda todo sin preguntar. Si un agente se confunde, el doc esta mal.

### Regla: doc en el mismo PR que el codigo

| Cambio | Doc a actualizar |
|---|---|
| Endpoint nuevo | OpenAPI spec del servicio |
| Servicio nuevo | Spec + README del servicio + CLAUDE.md |
| Modulo nuevo | Spec (sistema de modulos) + manifest YAML |
| Decision arquitectonica | ADR en `docs/decisions/` |
| Cambio de convencion | CLAUDE.md + bible.md |

---

## Microversioning

- Cada servicio tiene su propio semver independiente (`VERSION` file)
- Bump automatico desde commit type: `fix` → patch, `feat` → minor, `!` → major
- Tags: `{servicio}/v{version}` (ej: `auth/v1.3.0`)
- Changelog por servicio autogenerado con git-cliff
- Changelog de plataforma (`docs/CHANGELOG.md`) escrito a mano — lo visible para usuarios

---

## Deploy

### Pipeline (6 fases, `.github/workflows/deploy.yml`)

1. **Preflight** (self-hosted) — `deploy/scripts/preflight.sh`
2. **Build + Push** (GitHub-hosted) — SHA-pinned images a GHCR
3. **Deploy Dev** (optional, continue-on-error)
4. **Deploy Prod** (requires environment approval) — captura rollback state antes
5. **Smoke tests** — `deploy/scripts/health-check.sh`
6. **Record + Notify** — `POST /v1/platform/deploys` + notification

**Circuit-breaker:** si smoke tests fallan → rollback automatico via `deploy/scripts/rollback.sh`
(parser seguro, no usa `source`, valida con regex).

### Reglas

- Deploy a prod requiere approval via GitHub Environments
- Images taggeadas por SHA (deterministic), no `latest`
- Rollback captura versiones actuales antes de deployar
- Dependency ordering en compose: postgres → redis → nats → Go services

### Self-Healing Loop

- **HealthWatch** (port 8014): monitorea servicios, Prometheus, Docker
- **Daily triage** (7PM Argentina): AI analiza health summary → crea GitHub Issues
- **Post-deploy verification**: auto-cierra issues si el servicio volvió a healthy
- **Alertmanager → notification**: alertas criticas envian email inmediato

### Comandos clave

```bash
make versions          # Ver versiones running vs expected
make deploy            # Deploy a produccion (via workflow)
make status            # Estado de servicios + GPU
```

---

## Seguridad (restriccion, no feature)

- JWT verificacion local en cada servicio (clave publica compartida)
- Tenant isolation: JWT claim + subdomain deben coincidir
- RBAC granular verificado en cada handler
- Audit log inmutable de toda accion
- Network segmentation: Docker networks por tenant
- Containers distroless/scratch, usuario no-root
- Secrets en Docker secrets o Vault, nunca en env vars planos

---

## Referencias

| Documento | Que contiene | Estado |
|---|---|---|
| `docs/plans/2.0.x-plan01-sda-framework.md` | Spec completo del sistema | Activo |
| `CLAUDE.md` | Mapa para agentes | Activo |
| `docs/bible.md` | Este documento | Activo |
| `docs/decisions/` | ADRs | Activo (12 ADRs de 1.0.x, pendiente actualizar) |
| `README.md` | Overview del proyecto | Activo |

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
| RAG | NVIDIA RAG Blueprint v2.5.0 |
| LLM | Nemotron-Super-49B via NVIDIA API (externo) |
| Observability | OpenTelemetry + Grafana stack |
| Logger | slog (Go stdlib) |
| Testing | go test + testify + testcontainers |
| Testing frontend | bun:test + Playwright |
| CI/CD | GitHub Actions |
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

Scopes: `auth`, `chat`, `rag`, `ingest`, `notification`, `ws`, `platform`,
`web`, `cli`, `mcp`, `agent`, `docai`, `vision`, `feedback`, `infra`

### Pull Requests

1. Feature branch desde `main`
2. Implementar + tests + docs (en el mismo PR)
3. Push → CI automatico
4. Review (Opus o humano)
5. CI verde + review aprobado → squash merge
6. Post-merge: version bump + changelog + Docker image + GitHub Release

### Quality gates

1. `make test` (unit + contract del servicio afectado)
2. `make lint` (golangci-lint)
3. `make build` (compila)
4. Review aprobado
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

- Deploy a prod es siempre manual: `sda deploy auth` o `make deploy-auth`
- Nunca deploy automatico en merge a main
- Servicios CPU: start-first (zero downtime, Railway-style)
- Servicios GPU: drain-then-swap (5-15s buffer via WebSocket Hub)
- Rollback: `sda rollback auth` (pull imagen anterior, re-start)

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

| Documento | Que contiene |
|---|---|
| `docs/plans/2.0.x-plan01-sda-framework.md` | Spec completo del sistema |
| `CLAUDE.md` | Mapa para agentes |
| `docs/decisions/` | ADRs |
| `docs/api/` | OpenAPI specs por servicio |
| `CONTRIBUTING.md` | Workflow para contribuir |
| `SECURITY.md` | Modelo de seguridad |

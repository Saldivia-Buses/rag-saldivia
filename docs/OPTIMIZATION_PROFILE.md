# Perfil de Optimización Claude Code — SDA Framework

> Guía completa para replicar esta configuración en proyectos Go de microservicios.
> Versión: 1.0 | Fecha: 2026-04-12 | Proyecto: SDA Framework (14 services, ~50K lines Go)
> Basado en: astro-v2 OPTIMIZATION_PROFILE v2, adaptado para Go multi-tenant microservices.

---

## 1. Stack de Herramientas

### 1.1 MCP Servers (7 servidores)

| Server | Función | Uso obligatorio |
|--------|---------|-----------------|
| **CodeGraphContext** | Code graph: 270 functions, 108 files, 83 modules. Callers, callees, blast radius, dead code. Watch activo — auto-reindex en cada cambio. | `analyze_code_relationships(find_callers)` antes de editar funciones |
| **Repowise** | Documentation engine: wiki, risk assessment, dead code, architecture diagrams, decision records | `get_overview()` al inicio de cada task, `get_context()` antes de leer archivos |
| **Repomix** | Codebase packing para análisis AI-optimized | Análisis de repos externos, code reviews |
| **GitHub** | Issues, PRs, code search, commits | PRs, issues, code search |
| **Firecrawl** | Web scraping + search | Documentación externa, research |
| **Playwright** | Browser automation | Testing visual, screenshots |
| **Context7** | Library docs (React, Next.js, etc.) | Consultas de APIs y frameworks |

Configuración en `.mcp.json` (proyecto) y `~/.claude/settings.json` (global).

### 1.2 Skills (14 workflows + simplify built-in)

| Skill | Trigger | Disciplina |
|-------|---------|------------|
| `using-superpowers` | Inicio de TODA sesión | Framework de skills |
| `brainstorming` | Antes de crear features/servicios | Diseño antes de implementación |
| `writing-plans` | Specs/requerimientos claros | Plan bite-sized con TDD |
| `test-driven-development` | Antes de escribir código Go | Test primero (table-driven, testify) |
| `systematic-debugging` | Bug o comportamiento inesperado | Root cause (5 whys), fix después |
| `subagent-driven-development` | Tasks independientes por servicio | Subagent por task + 2-stage review |
| `executing-plans` | Plan en sesión separada | Batch execution con checkpoints |
| `dispatching-parallel-agents` | 2+ tasks paralelas | Agentes simultáneos |
| `verification-before-completion` | Antes de declarar "listo" | make test + lint + build + evidencia |
| `requesting-code-review` | Feature completa | gateway-reviewer / frontend-reviewer / security-auditor |
| `receiving-code-review` | Feedback de review | Evaluar antes de implementar |
| `finishing-a-development-branch` | Tests pasan, listo para merge | PR contra 2.0.x |
| `using-git-worktrees` | Trabajo que necesita aislamiento | Worktree aislado |
| `writing-skills` | Crear/editar skills | Verificar antes de deploy |

Viven en `.claude/skills/superpowers/` como Markdown con frontmatter YAML.
También accesibles como slash commands via `.claude/commands/` (symlinks).

### 1.3 Agentes Especializados (10)

| Agente | Scope |
|--------|-------|
| `gateway-reviewer` | Handlers chi, middleware, JWT, RBAC, sqlc, NATS, tenant isolation |
| `frontend-reviewer` | Componentes React, hooks, auth, comunicación con backend |
| `security-auditor` | Auditoría completa: JWT, tenant, SQL injection, NATS, Docker |
| `test-writer` | Tests Go (testify, testcontainers) + frontend (bun, Playwright) |
| `debugger` | Failure modes, logs, config, code tracing |
| `deploy` | Preflight checks, Docker Compose, health verification |
| `status` | Health checks, GPU, Docker, resource monitoring |
| `doc-writer` | CLAUDE.md, bible, README, ADRs |
| `plan-writer` | Planes con phases, migrations, NATS events |
| `ingest` | Pipeline de ingesta, tree generation |

---

## 2. Sistema de Tests (3 niveles)

### Nivel 1: Pre-commit (automático, ~3s)

```
.git/hooks/pre-commit
└── check-invariants.sh (35 checks)
    ├── Go Workspace (3): go.work lists services, pkg, tools
    ├── Migration Pairs (4): up/down pairs, sequential numbering (tenant + platform)
    ├── Service Structure (4): cmd/main.go, VERSION, Dockerfile, semver
    ├── sqlc Config (2): sqlc.yaml exists, points to queries
    ├── sqlc Freshness (1): generated code not older than queries
    ├── Tenant Isolation (2): no hardcoded IDs, handlers use middleware
    ├── Security (2): no secrets in source, no .env committed
    ├── Docker Compose (2): compose exists, services listed
    ├── Proto Sync (1): gen/go exists if proto files exist
    ├── NATS Convention (2): tenant prefix on publishes and consumers
    ├── Write→Event (1): services with writes have NATS publishes
    ├── Documentation (1): every service has README.md
    ├── Handler Patterns (2): MaxBytesReader, JSON error format
    ├── Dockerfile Security (2): multi-stage build, non-root user
    ├── Repowise Index (1): not older than 3 days
    ├── Frontend (2): package.json exists, no hardcoded URLs
    ├── Docs-Code Sync (1): files in CRITICAL_FLOWS.md exist
    └── Silent Failures (2): no swallowed errors, no excessive unwrapped returns
```

### Nivel 2: Smart Test Runner (manual, variable)

```bash
bash .claude/hooks/smart-test.sh   # or /smart-test
```

Lee `git diff`, busca en `.claude/hooks/test-file-mapping.json` qué test packages
son relevantes, y corre solo esos. 52 mappings cubriendo todos los servicios y packages.

Señales especiales:
- `_invariants_only` — si cambiaron migraciones, solo corre invariants
- `_sqlc_regen` — si cambiaron queries, avisa que hay que regenerar

### Nivel 3: Pre-push (automático, ~90s)

```
.git/hooks/pre-push
├── check-invariants.sh (35 checks, ~3s)
├── make build (all services, ~30s)
├── make lint (golangci-lint, ~15s)
└── make test (all tests, ~45s)
```

**Total: 35 invariant checks + full test suite automatizado.**

---

## 3. Documentación Viva

### 3.1 CLAUDE.md (3 niveles, auto-cargado)

| Nivel | Ubicación | Contenido |
|-------|-----------|-----------|
| Global | `~/.claude/CLAUDE.md` (76 líneas) | Skills obligatorias, code intelligence protocol, output style, rationalization detection |
| Proyecto | `.claude/CLAUDE.md` (190+ líneas) | 7 invariantes críticos, file guards, CodeGraphContext + Repowise instructions, agents, self-check |
| Root | `CLAUDE.md` (235 líneas) | Arquitectura, stack, comandos, convenciones, archivos críticos |

**Reglas clave:**
> ANTES DE EDITAR: `analyze_code_relationships(find_callers)` + `git log -5`
> ANTES DE TOCAR `pkg/*`: `find_importers` para ver todo lo que se rompe
> SI CAMBIA UN CRITICAL FLOW: leer `docs/CRITICAL_FLOWS.md` primero
> ANTES DE COMMITEAR: pre-commit hook corre 35 checks automáticamente
> ANTES DE DECLARAR "LISTO": `verification-before-completion` skill MANDATORY

### 3.2 docs/CRITICAL_FLOWS.md (929 líneas, 5 flujos)

Cada flujo tiene: diagrama ASCII, step-by-step con file:function references, invariantes, failure modes.

1. **Auth** — Login → JWT (Ed25519) → MFA → Refresh → Logout + blacklist
2. **Multi-Tenant Routing** — Traefik → headers → tenant.Resolver → per-tenant DB pool
3. **Chat + Agent Pipeline** — Message → Chat gRPC → Agent → LLM → Tools → Traces
4. **Document Ingestion** — Upload → S3 → Extractor (OCR/Vision) → Tree → PostgreSQL
5. **WebSocket Real-Time** — JWT → Upgrade → Hub → NATS Bridge → BroadcastToTenant

### 3.3 docs/bible.md (218 líneas)

10 principios permanentes, stack, git workflow, convenciones Go/frontend, deploy, seguridad.

---

## 4. Hooks Automáticos (5 hooks)

| Hook | Trigger | Acción | Bloquea? |
|------|---------|--------|----------|
| **session-briefing** | `SessionStart` (startup) | Últimos 10 commits, archivos 48h, uncommitted, versions | No |
| **pre-edit** | `PreToolUse` (Edit\|Write) | Alerta si archivo fue modificado recientemente | No (informational) |
| **pre-commit** | `PreToolUse` (Bash git commit*) | 35 invariant checks | Sí (exit 2) |
| **stop-verify** | `Stop` | Corre invariants + cuenta uncommitted files | No (context) |
| **stop-haiku** | `Stop` | Haiku verifica que se mostró evidencia | No (advisory) |

Git hooks adicionales:
- `.git/hooks/pre-commit` — corre check-invariants.sh
- `.git/hooks/pre-push` — corre invariants + build + lint + test

---

## 5. CRITICAL INVARIANTS (7 reglas)

Estas son las reglas de SDA que NUNCA se pueden romper:

1. **Tenant isolation at every layer** — JWT claim == request tenant. sqlc queries incluyen `tenant_id`. `tenant.Resolver` da pools aislados.

2. **JWT is the single source of identity** — UserID, TenantID, Slug, Role vienen del JWT. Verificación local con ed25519 en cada servicio.

3. **NATS subjects are tenant-namespaced** — `tenant.{slug}.{service}.{entity}[.{action}]`. Nunca publicar sin slug.

4. **Every write publishes a NATS event** — Para real-time via WebSocket Hub. Sin polling en frontend.

5. **Migration pairs are complete** — `.up.sql` ↔ `.down.sql`. Números secuenciales sin gaps.

6. **Service structure is uniform** — `cmd/main.go`, `VERSION`, `Dockerfile` (multi-stage, non-root), `README.md`, en `go.work`.

7. **Error responses are JSON** — `http.Error(w, '{"error":"msg"}', code)` o `writeJSON()`. Nunca plain text.

---

## 6. Memoria Persistente (36 archivos)

| Tipo | Cantidad | Propósito |
|------|----------|-----------|
| `feedback_*.md` | 14 | Correcciones de trabajo: TDD, root cause, blast radius, etc. |
| `project_*.md` | 14 | Estado de features, planes, decisiones |
| `user_*.md` | 1 | Perfil de Enzo (rol, expertise) |
| `reference_*.md` | 1 | Links a recursos externos |
| `MEMORY.md` | 1 (índice) | Links a todos los archivos |

**Archivos clave:**
- `feedback_invariants_precommit.md` — SIEMPRE correr 35 checks antes de commit
- `feedback_root_cause.md` — 5 whys antes de fixear
- `feedback_systematic_audit.md` — Bug en 1 servicio → auditar TODOS
- `feedback_blast_radius_pkg.md` — Verificar TODOS los importers antes de tocar pkg/
- `project_optimization_profile.md` — Este perfil existe y cómo funciona

---

## 7. Protecciones contra Errores Silenciosos

| Protección | Qué previene | Check |
|------------|-------------|-------|
| `_ = err` detection | Errores tragados sin logging | Invariant #34 |
| Excessive unwrapped returns | `return err` sin contexto (`fmt.Errorf`) | Invariant #35 |
| MaxBytesReader check | Memory exhaustion por body grande | Invariant #26 |
| JSON error format | Responses que el frontend no puede parsear | Invariant #27 |
| NATS publish check | Writes sin real-time update | Invariant #23 |
| sqlc freshness | Generated code desactualizado vs queries | Invariant #14 |
| Repowise staleness | MCP dando contexto viejo | Invariant #31 |
| Docs-code sync | CRITICAL_FLOWS referenciando archivos borrados | Invariant #33 |

---

## 8. Cómo Replicar en Otro Proyecto Go

### Paso 1: MCP Servers
```bash
pip install repowise && repowise init .
npm install -g codegraphcontext && cgc analyze .
```

### Paso 2: Skills
```bash
cp -r .claude/skills/ <proyecto>/.claude/skills/
```

### Paso 3: CLAUDE.md
Crear `.claude/CLAUDE.md` con:
1. Code intelligence protocol (CodeGraphContext + Repowise)
2. 3-7 invariantes críticos del proyecto
3. File guards → docs de flujos críticos
4. Self-check checklist antes de "done"

### Paso 4: Invariant Tests
Crear `.claude/hooks/check-invariants.sh` verificando:
- Workspace sync (go.work, package.json)
- Migration pairs
- Service structure
- Tenant/auth patterns
- Security patterns
- Docs-code sync

### Paso 5: Smart Test Runner
Crear `.claude/hooks/test-file-mapping.json` mapeando archivos a test packages.
Crear `.claude/hooks/smart-test.sh` que lee git diff y corre tests relevantes.

### Paso 6: Hooks
```json
// .claude/settings.json
{
  "hooks": {
    "SessionStart": [{"matcher": "startup", "hooks": [{"type": "command", "command": "session-briefing.sh"}]}],
    "PreToolUse": [
      {"matcher": "Edit|Write", "hooks": [{"type": "command", "command": "pre-edit-check.sh"}]},
      {"matcher": "Bash", "hooks": [{"type": "command", "if": "Bash(git commit*)", "command": "pre-commit-check.sh"}]}
    ],
    "Stop": [{"hooks": [{"type": "command", "command": "stop-verify.sh"}]}]
  }
}
```

### Paso 7: Git Hooks
```bash
# Pre-commit
echo '#!/bin/bash
bash .claude/hooks/check-invariants.sh || exit 1' > .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

# Pre-push
echo '#!/bin/bash
bash .claude/hooks/check-invariants.sh && make build && make lint && make test' > .git/hooks/pre-push
chmod +x .git/hooks/pre-push
```

### Paso 8: Critical Flows
Documentar los 5-10 flujos más importantes en `docs/CRITICAL_FLOWS.md` con:
- Diagrama ASCII
- Step-by-step con file:function references
- Invariantes por flujo
- Failure modes

### Paso 9: Agentes Especializados
Crear `.claude/agents/` con perfiles para:
- Code review (backend, frontend, security)
- Debugging, testing, deployment
- Documentation, planning

### Paso 10: Memoria
Inicializar `~/.claude/projects/<hash>/memory/MEMORY.md` con feedback y contexto.

---

## 9. Métricas de Impacto

| Métrica | Sin perfil | Con perfil |
|---------|-----------|------------|
| Bugs por desync estructural | ~3/sesión | 0 (35 invariant checks) |
| Blast radius no verificado | Frecuente | 0 (CodeGraphContext find_callers) |
| Tiempo debugging silent failures | 2-5 horas | Inmediato (error surfacing) |
| Tests pre-commit | 0 | 35 checks (~3s) |
| Tests pre-push | 0 | Full suite (~90s) |
| Smart test selection | No | Sí (52 mappings) |
| Flujos documentados | 0 | 5 (929 líneas) |
| Contexto cargado al inicio | ~2K tokens | ~15K tokens (3 CLAUDE.md + memory + briefing) |
| Regresiones en features existentes | Frecuente | Detectado pre-commit |
| Code intelligence antes de editar | No | Obligatorio (CodeGraphContext + Repowise) |
| Review automatizado | No | 10 agentes especializados |
| Skills de workflow | No | 14 skills + 17 slash commands |

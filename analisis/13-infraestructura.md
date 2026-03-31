# 13 — Infraestructura

## Monorepo

### Estructura

```
rag-saldivia-1/           (root)
  apps/
    web/                  → Next.js 16 (la unica app activa)
  packages/
    db/                   → Drizzle ORM + queries + schema
    shared/               → Zod schemas + tipos
    config/               → Config loader
    logger/               → Logger estructurado
```

### Turborepo

**Config:** `turbo.json` en root.

Tareas configuradas:
- `dev` — corre Next.js en dev mode
- `build` — build de produccion
- `test` — unit tests de todos los packages
- `test:coverage` — tests con cobertura
- `lint` — lint de todos los packages
- `type-check` — tsc --noEmit

### Bun Workspaces

**Config:** `package.json` root con `"workspaces": ["apps/*", "packages/*"]`

Los packages se referencian por nombre:
```json
// apps/web/package.json
"dependencies": {
  "@rag-saldivia/db": "workspace:*",
  "@rag-saldivia/shared": "workspace:*",
  "@rag-saldivia/config": "workspace:*",
  "@rag-saldivia/logger": "workspace:*"
}
```

---

## Docker

### `docker-compose.yml`

Servicios definidos:
- Redis (requerido para BullMQ + JWT blacklist)
- Milvus (vector DB para el RAG)
- RAG Server NVIDIA (blueprint)

### Config overrides en `config/`

```
config/
  compose-overrides.yaml            → override general
  compose-overrides-workstation.yaml → override para workstation
  compose-platform-services.yaml    → servicios de plataforma
  compose-openrouter-proxy.yaml     → proxy OpenRouter
  compose-guardrails-cloud.yaml     → guardrails cloud
  admin-overrides.yaml              → overrides admin
  guardrails.yaml, milvus.yaml, models.yaml, observability.yaml, platform.yaml, prompt.yaml
  profiles/
    workstation-1gpu.yaml           → perfil 1x RTX PRO 6000
```

---

## Deploy

### Workstation fisica

- **OS:** Ubuntu 24.04
- **GPU:** 1x RTX PRO 6000 Blackwell (96GB VRAM)
- **LLM:** Nemotron-Super-49B
- **Vector DB:** Milvus
- **Redis:** local
- **SQLite:** local

### Makefile

```bash
make deploy PROFILE=workstation-1gpu
```

El Makefile orquesta:
1. Build de la app
2. Tests
3. Deploy via SSH/rsync
4. Health checks post-deploy

### CI/CD — GitHub Actions

```
.github/workflows/
  ci.yml      → PRs + push a dev: tests, type-check, lint, coverage, components, visual, a11y
  deploy.yml  → Push a main: deploy a workstation
  release.yml → Tags semver: release automation
```

---

## Scripts

| Script | Proposito |
|--------|-----------|
| `scripts/setup.ts` | Onboarding: init DB, seed, install deps |
| `scripts/fix-cli-pkg.py` | Fix del build de la CLI (archivada) |
| `scripts/test-login-final.sh` | Test de integracion de login |
| `scripts/link-libsql.sh` | Linking de libsql |

### Setup (primera vez)

```bash
bun run setup
# Equivale a: bun scripts/setup.ts
# 1. Verifica env vars
# 2. Crea directorio data/
# 3. Inicializa SQLite
# 4. Corre migraciones
# 5. Seedea admin + demo user
# 6. Verifica Redis
```

---

## Configuracion

### Variables de entorno

```env
# Obligatorias
JWT_SECRET=...             # openssl rand -base64 32
SYSTEM_API_KEY=...         # openssl rand -hex 32
DATABASE_PATH=./data/app.db
REDIS_URL=redis://localhost:6379

# Opcionales
RAG_SERVER_URL=http://localhost:8081
RAG_TIMEOUT_MS=120000
MOCK_RAG=false
OPENROUTER_API_KEY=...     # para mock mode
OPENROUTER_MODEL=anthropic/claude-haiku-4-5
JWT_EXPIRY=24h
LOG_LEVEL=INFO
NODE_ENV=development
```

### Archivos de config

```
config/
  platform.yaml       → Config de la plataforma (nombre, version)
  models.yaml          → Modelos AI disponibles
  prompt.yaml          → System prompts
  observability.yaml   → Logging, tracing
  milvus.yaml          → Milvus vector DB config
  guardrails.yaml      → Content moderation
```

---

## Git

### Branch strategy

- **`1.0.x`** — branch activa de desarrollo
- **`main`** — stack Python estable (produccion)
- **Feature branches:** `plan{N}-{slug}` (cuando haya mas gente)

### Hooks (Husky)

```
.husky/
  pre-commit   → lint-staged
  commit-msg   → commitlint
```

### Commitlint

Config en `.commitlintrc.json`:
- Extiende `@commitlint/config-conventional`
- Scopes permitidos: web, db, shared, config, logger, ui, chat, auth, rag, agents, plans, deps, messaging, admin, setup

### Lint-staged

Config en `.lintstagedrc.js`:
- `*.{ts,tsx}` → eslint --fix
- `*.{ts,tsx,json,md}` → prettier --write

---

## Vendor

```
vendor/
  rag-blueprint/    → Git submodule del NVIDIA RAG Blueprint v2.5.0 (commit a67a48c)
```

No se modifica. Solo se referencia para documentacion.

---

## Editor

### .editorconfig
```ini
root = true
[*]
indent_style = space
indent_size = 2
end_of_line = lf
charset = utf-8
trim_trailing_whitespace = true
insert_final_newline = true
```

### TypeScript
```json
// tsconfig.json (apps/web)
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "strict": true,
    "exactOptionalPropertyTypes": true,
    "paths": { "@/*": ["./src/*"] }
  }
}
```

---

## Archivado

### `_archive/`

123 archivos de codigo aspiracional movidos en Plan 13:
- 21 paginas futuras
- 28 API routes legacy
- 70+ componentes avanzados
- Workers, CLI, etc.

**Recuperable con git.** No se elimino del repositorio, solo se movio.

---

## Knip (dead code detection)

Config en `knip.json`. Detecta exports no usados, dependencias no usadas, archivos huerfanos.

---

## CODEOWNERS

```
# Ya existe desde Plan 11
* @Camionerou
```

Se actualizara cuando haya mas colaboradores.

---

## Stack de dependencias completo

### Runtime y build

| Dependencia | Version | Justificacion |
|-----------|---------|--------------|
| Bun | 1.3.x | Runtime TS nativo, rapido |
| TypeScript | 6.0 | `exactOptionalPropertyTypes` |
| Turborepo | 2.x | Build orchestration monorepo |
| Next.js | 16 | App Router, Server Components, proceso unico |

### Base de datos y queue

| Dependencia | Justificacion |
|-----------|--------------|
| Drizzle ORM 0.45 | Type-safe, SQLite+Postgres, migraciones |
| @libsql/client | Driver SQLite, future-proof a Turso |
| ioredis | Redis client (JWT blacklist + BullMQ) |
| BullMQ | Job queue para ingesta |

### Auth y AI

| Dependencia | Justificacion |
|-----------|--------------|
| jose | JWT sign/verify (compatible Edge Runtime) |
| bcrypt | Password hashing |
| ai (Vercel AI SDK) | Protocolo estandar de streaming |
| @ai-sdk/react | useChat hook |
| next-safe-action + zod | Server actions type-safe |

### UI

| Dependencia | Justificacion |
|-----------|--------------|
| Tailwind CSS 4.x | Utility-first, `@theme inline` para dark mode |
| @radix-ui/* | Primitivos accesibles |
| shadcn/ui | Componentes pre-estilizados |
| @tanstack/react-table | Tablas avanzadas |
| lucide-react | Iconos con tree-shaking |
| next-themes | Dark mode class strategy |
| sonner | Toast notifications |
| cmdk | Command palette |

### Testing

| Dependencia | Justificacion |
|-----------|--------------|
| bun:test (built-in) | Test runner nativo |
| happy-dom | DOM virtual (mas rapido que jsdom) |
| @testing-library/react | Queries accesibles |
| playwright | Visual regression + E2E |
| axe-playwright | WCAG AA auditing |

### Dependencias descartadas

| Dependencia | Razon |
|-----------|-------|
| better-sqlite3 | → libsql (ADR-001) |
| jsonwebtoken | → jose (Edge compatible) |
| prisma | → Drizzle (mas ligero) |
| jsdom | → happy-dom (mas rapido) |
| json-render | → markdown rendering directo (Plan 17) |

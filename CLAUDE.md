# RAG Saldivia — Contexto de proyecto

## Qué es este proyecto

Overlay sobre **NVIDIA RAG Blueprint v2.5.0** que agrega autenticación, RBAC, multi-colección, frontend Next.js 15, CLI TypeScript, y perfiles de deployment.

- **No es un fork** — incluye el blueprint como git submodule en `vendor/rag-blueprint/` (commit a67a48c, post-v2.3.0)
- **Repo local:** `~/rag-saldivia/` — branch `experimental/ultra-optimize`
- **Repo remoto:** https://github.com/Camionerou/rag-saldivia
- **Deploy activo:** workstation física Ubuntu 24.04 (1x RTX PRO 6000 Blackwell, 96GB VRAM)

> **Nota:** La branch `main` tiene el stack Python+SvelteKit (estable en producción).
> Esta branch (`experimental/ultra-optimize`) es la reescritura completa en TypeScript.

## Arquitectura de servicios (esta branch)

```
Usuario → Next.js :3000 ——————————————————→ RAG Server :8081
           (UI + auth + proxy)                      ↓
                                           Milvus + NIMs
                                                    ↓
                                       Nemotron-Super-49B
```

**Un único proceso** reemplaza el gateway Python (9000) + el frontend SvelteKit (3000).

## Comandos clave

```bash
# Onboarding (primera vez)
bun run setup

# Desarrollo
bun run dev              # Next.js en :3000
rag status               # Health check de todos los servicios

# Deploy en workstation física (stack Python, branch main)
cd ~/rag-saldivia && make deploy PROFILE=workstation-1gpu

# Tests — opciones disponibles
bun run test                                        # todos los tests via Turborepo
bun test apps/web/src/lib/auth/__tests__/           # solo auth/RBAC (17 tests)
bun test packages/db/src/__tests__/                 # solo DB queries (24 tests)
bun test packages/logger/src/__tests__/             # solo logger/blackbox (24 tests)
bun test packages/config/src/__tests__/             # solo config loader (14 tests)

# CLI (instalar global: cd apps/cli && bun link)
rag users list
rag collections list
rag ingest status
rag audit log
rag status
```

## Estructura (esta branch)

```
apps/
  web/              → Next.js 15 — servidor único (UI + auth + proxy RAG + admin)
  cli/              → CLI TypeScript (rag users/collections/ingest/audit/config/db)
packages/
  shared/           → Zod schemas + tipos compartidos (User, Area, Session, etc.)
  db/               → Drizzle ORM + @libsql/client (12 tablas, reemplaza auth.db + Redis)
  config/           → config loader TypeScript (reemplaza config.py)
  logger/           → logger estructurado + black box replay + rotación de archivos
scripts/
  setup.ts          → onboarding cero-fricción (bun run setup)
docs/
  architecture.md   → arquitectura del nuevo stack
  blackbox.md       → sistema de logging y replay
  cli.md            → referencia completa de la CLI
  onboarding.md     → guía de 5 minutos
  workflows.md      → flujos de trabajo del proyecto (git, tests, features, deploy)
  plans/
    ultra-optimize-plan1-birth.md   → construcción del monorepo TS (completado 2026-03-24)
    ultra-optimize-plan2-testing.md → testing sistemático 7 fases (completado 2026-03-25)
    ultra-optimize-plan3-bugfix.md  → bugfix CodeGraphContext (completado 2026-03-25)
config/             → YAMLs sin cambios
patches/            → patches del blueprint NVIDIA sin cambios
vendor/             → submódulo NVIDIA sin cambios
```

## Stack técnico

| Componente | Tecnología |
|---|---|
| Lenguaje | TypeScript 6.0 |
| Runtime | Bun |
| Framework web | Next.js 15 App Router |
| Base de datos | SQLite vía Drizzle ORM + @libsql/client |
| Auth | JWT (jose) en cookie HttpOnly |
| Validación | Zod (compartido entre todos los paquetes) |
| Build | Turborepo + Bun workspaces |
| CLI | Commander + @clack/prompts + chalk |

## Archivos críticos — leer antes de modificar

- `apps/web/src/middleware.ts` — JWT + RBAC en cada request
- `apps/web/src/lib/auth/jwt.ts` — createJwt, verifyJwt, cookies
- `packages/db/src/schema.ts` — schema completo de 12 tablas SQLite
- `packages/db/src/queries/users.ts` — CRUD de usuarios + permisos
- `packages/logger/src/backend.ts` — logger con rotación de archivos
- `apps/web/src/lib/rag/client.ts` — proxy RAG con modo mock

## Variables de entorno importantes

Ver `.env.example` para la lista completa documentada.

```env
JWT_SECRET=...          # openssl rand -base64 32
SYSTEM_API_KEY=...      # openssl rand -hex 32
RAG_SERVER_URL=http://localhost:8081
DATABASE_PATH=./data/app.db
MOCK_RAG=false          # true para dev sin Docker
LOG_LEVEL=INFO
```

## Patrones importantes (aprendidos en producción)

- **Temporal API** para todos los timestamps → elimina el bug `_ts()` de SQLite
- **Server Components por defecto** → cero JS al browser salvo donde sea necesario
- **Cache con `unstable_cache`** → cachear llamadas al RAG con `tags: ['collections']`
- **SSE**: verificar status HTTP ANTES de streamear (gateway.py tenía bug que siempre retornaba 200)
- **SQLite locking**: `ingestion_queue` usa `locked_at` para locking optimista sin Redis
- **Logger + DB**: `@rag-saldivia/db` importado estáticamente en `packages/logger` — import dinámico fallaba silenciosamente en webpack/Next.js
- **CJS sobre ESM**: paquetes `packages/*` sin `"type": "module"` para compatibilidad con webpack

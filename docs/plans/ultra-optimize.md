# Plan: Ultra-Optimización de rag-saldivia

> Este documento vive en `docs/plans/ultra-optimize.md` dentro de la branch `experimental/ultra-optimize`.
> Se actualiza diariamente junto al `CHANGELOG.md`. Cada tarea completada se marca con fecha.

---

## Contexto

El proyecto es un overlay sobre el NVIDIA RAG Blueprint v2.5.0. El código propio son dos procesos: un gateway Python (FastAPI, puerto 9000) y un frontend SvelteKit (puerto 3000). Este plan los unifica y refactoriza completamente.

Lo que NO cambia: el submódulo NVIDIA (`vendor/rag-blueprint/`), los patches del blueprint (`patches/`), los YAMLs de configuración (`config/`), y los servicios auxiliares pequeños (`services/mode-manager/`, `services/openrouter-proxy/`).

---

## Arquitectura objetivo

Un único servidor Next.js 15 en el puerto 3000 reemplaza tanto el gateway Python como el frontend SvelteKit. Una única base de datos SQLite reemplaza el SQLite de auth más Redis. Un único lenguaje TypeScript 6.0 reemplaza Python más TypeScript.

```
ANTES:  Usuario → SvelteKit :3000 → gateway.py :9000 → RAG :8081
AHORA:  Usuario → Next.js :3000 ——————————————————————→ RAG :8081
```

---

## Stack

- **Monorepo:** Turborepo + Bun workspaces
- **Servidor único:** Next.js 15 App Router — reemplaza gateway Python + frontend SvelteKit
- **Base de datos única:** Drizzle ORM + better-sqlite3 — reemplaza SQLite auth + Redis
- **Validación compartida:** Zod — reemplaza Pydantic (Python) + interfaces TypeScript duplicadas
- **Lenguaje:** TypeScript 6.0 — Temporal API para timestamps, mejor inferencia en JSX
- **CLI:** Bun + Commander + @clack/prompts — reemplaza el CLI Python actual
- **Logging:** paquete propio `packages/logger` — frontend log + backend log + black box replay
- **Git workflow:** Conventional Commits + Commitlint + Husky + Changesets

---

## Estructura del monorepo

```
rag-saldivia/
├── apps/
│   ├── web/          → servidor único (Next.js 15): UI + auth + proxy RAG + admin
│   └── cli/          → CLI TypeScript: gestión completa del sistema desde terminal
├── packages/
│   ├── shared/       → Zod schemas y tipos compartidos entre web y cli
│   ├── db/           → Drizzle schema, migraciones, queries (base de datos única)
│   ├── config/       → config loader TypeScript (reemplaza config.py)
│   └── logger/       → sistema de logging + black box replay
├── docs/
│   └── plans/
│       └── ultra-optimize.md   ← este archivo
├── CHANGELOG.md      → se actualiza con cada tarea completada
├── config/           → YAMLs sin cambios
├── patches/          → sin cambios
├── vendor/           → sin cambios
├── scripts/          → scripts TypeScript (reemplazan *.sh y *.py)
├── turbo.json
└── package.json      → Bun workspaces root
```

---

## Seguimiento diario

Formato de cada tarea: `- [ ] Descripción — estimación`
Al completarla: `- [x] Descripción — completado YYYY-MM-DD`
Cada tarea completada genera una entrada en `CHANGELOG.md` antes de hacer commit.

---

## Paso previo — Crear la branch *(5 min)*

- [x] Crear y publicar la branch `experimental/ultra-optimize` desde `main` — completado 2026-03-24
- [x] Crear `docs/plans/ultra-optimize.md` con este plan y hacer el primer commit en la branch — completado 2026-03-24
- [x] Crear `CHANGELOG.md` con la entrada inicial `[Unreleased]` que registra el inicio de la branch — completado 2026-03-24

A partir de este punto todo el trabajo ocurre en `experimental/ultra-optimize`. El branch `main` no se toca.

---

## Fase 0 — Onboarding cero-fricción *(2-4 hs)*

Objetivo: cualquier persona clona el repo y con un solo comando tiene el sistema corriendo localmente. Sin importar el estado de su máquina.

- [x] Script `scripts/setup.ts`: preflight check de dependencias (Bun, Docker, puertos libres), instalación de paquetes, creación de `.env.local` desde `.env.example`, migraciones de DB, seed de datos de desarrollo, y resumen de estado final — completado 2026-03-24
- [x] `.env.example` completamente documentado: cada variable tiene descripción, valor por defecto para dev, y si es requerida en producción — completado 2026-03-24
- [x] Makefile actualizado con los targets `setup`, `setup-check`, `reset`, `dev` apuntando al nuevo stack — completado 2026-03-24
- [x] Manejo de errores del setup: si falta Bun, Docker o un puerto está ocupado, el mensaje explica exactamente qué hacer y cómo resolverlo. Nunca un stack trace genérico — completado 2026-03-24
- [x] El sistema puede arrancar en modo mock (sin Docker/RAG real) para desarrollo de UI — completado 2026-03-24 (variable MOCK_RAG en .env.example)

---

## Fase 1 — Git workflow + CHANGELOG día cero *(2-4 hs)*

Objetivo: establecer las reglas del juego antes de tocar código. El CHANGELOG arranca con el primer commit de la branch.

- [x] Commitlint + Husky: hook `commit-msg` rechaza commits que no sigan Conventional Commits, hook `pre-push` corre type-check — completado 2026-03-24
- [x] `.github/workflows/ci.yml`: lint + type-check + tests en cada PR usando Bun y Turborepo — completado 2026-03-24
- [x] `.github/workflows/deploy.yml`: reemplaza el actual, deploy solo al hacer tag `v*` — completado 2026-03-24
- [x] `.github/workflows/release.yml`: al publicar un release en GitHub mueve `[Unreleased]` a `[vX.Y.Z]` automáticamente — completado 2026-03-24
- [x] PR template con sección obligatoria de entrada al CHANGELOG. El CI bloquea PRs sin ella — completado 2026-03-24
- [x] Changesets configurado para versionado semántico — completado 2026-03-24

---

## Fase 2 — Monorepo + paquetes compartidos *(4-8 hs)*

Objetivo: crear la infraestructura compartida que usan todos los demás paquetes. Esta fase no tiene UI visible pero es la base de todo.

- [x] `turbo.json` con pipeline build → test → lint. Cache de Turborepo configurado para no repetir trabajo — completado 2026-03-24
- [x] `packages/shared`: schemas Zod de User, Area, Collection, Session, Message, IngestionJob. Exporta tipos TypeScript derivados. Elimina la duplicación entre Pydantic (Python) e interfaces TypeScript actuales — completado 2026-03-24
- [x] `packages/db`: schema Drizzle completo con las tablas users, areas, user_areas, sessions, messages, collections, ingestion_jobs, ingestion_queue, y events (black box). Migraciones generadas. Conexión singleton. Queries nombradas por dominio — completado 2026-03-24
- [x] `packages/db`: tabla `ingestion_queue` reemplaza Redis. Locking por columna `locked_at` evita race conditions sin necesidad de Redis — completado 2026-03-24
- [x] `packages/config`: config loader TypeScript que lee los YAMLs existentes en `config/`, valida con Zod, y exporta config tipada. Reemplaza `saldivia/config.py` — completado 2026-03-24
- [x] `packages/logger`: estructura de archivos y contratos de la API del logger definidos con implementación base (backend, frontend, blackbox, suggestions) — completado 2026-03-24

---

## Fase 3 — Servidor único: apps/web *(30-40 hs total)*

Reemplaza simultáneamente el gateway Python (1238 líneas) y el frontend SvelteKit (~13k líneas).

Principios que aplican a toda la fase:

- Server Components por defecto. Client Components solo donde sea imprescindible (chat SSE, sliders interactivos, modales con estado local)
- Caching activo desde el primer endpoint, no como afterthought
- Todo error va al black box. Ningún catch silencioso

### Fase 3a — Auth core *(4-6 hs)*

- [x] Middleware Next.js: verifica JWT en cada request, aplica RBAC por ruta — completado 2026-03-24
- [x] Endpoints de login, logout y refresh: login emite JWT en cookie HttpOnly, logout invalida sesión en DB — completado 2026-03-24
- [x] Librería interna de auth: createJwt, verifyJwt, extractClaims, hasRole, canAccessRoute, getCurrentUser, requireUser, requireAdmin — completado 2026-03-24
- [x] Tests del flujo de auth completo — completado 2026-03-24

Criterio de done: login funciona, cookie se setea, middleware bloquea rutas protegidas.

### Fase 3b — DB layer *(3-5 hs)*

- [x] Integrar `packages/db` en `apps/web`. Conexión disponible en Server Components y Route Handlers — completado 2026-03-24
- [x] Portar todos los métodos de `saldivia/auth/database.py` (875 líneas) a queries Drizzle tipadas — completado 2026-03-24
- [x] Comandos `db:migrate` y `db:seed` funcionales — completado 2026-03-24

Criterio de done: todas las queries del gateway Python tienen equivalente en Drizzle y los tests pasan.

### Fase 3c — RAG proxy + SSE streaming *(3-5 hs)*

- [x] Route Handler que recibe la query del cliente, verifica permisos de colección, hace proxy SSE hacia RAG :8081, y reenvía el stream al cliente — completado 2026-03-24
- [x] Verificar el status HTTP del RAG antes de empezar a streamear — completado 2026-03-24
- [x] Cliente HTTP interno para el RAG Server con timeout, modo mock, y manejo de errores con sugerencias — completado 2026-03-24
- [x] Cache de 60 segundos para la lista de colecciones — completado 2026-03-24

Criterio de done: chat streaming funciona end-to-end. Los errores del RAG se propagan correctamente al cliente.

### Fase 3d — Collections + ingestion *(4-6 hs)*

- [x] Páginas de colecciones: lista con cache — completado 2026-03-24
- [x] Página de upload: drag & drop de archivos, crea job en `ingestion_queue` — completado 2026-03-24
- [x] Endpoints de ingesta: status de jobs, cancelar (POST /api/upload, GET/DELETE /api/admin/ingestion) — completado 2026-03-24
- [x] Worker de ingesta en TypeScript: reemplaza `ingestion_worker.py` y `watch.py` — completado 2026-03-24

Criterio de done: se puede subir un PDF, el job aparece en la DB y el worker lo procesa.

### Fase 3e — Chat UI *(5-8 hs)*

- [x] Página de lista de sesiones (Server Component) — completado 2026-03-24
- [x] Página de chat específico: historial como Server Component, input y streaming como Client Component — completado 2026-03-24
- [x] Componente de streaming SSE: maneja las fases idle, streaming, done y error — completado 2026-03-24
- [x] Integración crossdoc: portados `useCrossdocDecompose.ts` y `useCrossdocStream.ts` a `apps/web/src/hooks/` adaptados para Next.js — completado 2026-03-24
- [x] Server Actions: crear sesión, renombrar, eliminar, feedback por mensaje — completado 2026-03-24

Criterio de done: chat funciona con RAG estándar y crossdoc. Historial persiste. Feedback funciona.

### Fase 3f — Admin UI *(4-6 hs)*

- [x] Gestión de usuarios: lista, crear con multi-select de áreas, eliminar, activar/desactivar — completado 2026-03-24
- [x] Server Actions para usuarios y áreas (CRUD completo) — completado 2026-03-24
- [x] Gestión de áreas: UI completa con CRUD (crear, editar, eliminar con protección) — completado 2026-03-24
- [x] Permisos: asignación de colecciones a áreas con nivel read/write — completado 2026-03-24
- [x] Config RAG: sliders de parámetros, toggles reranker y guardrails, reset a defaults — completado 2026-03-24
- [x] Estado del sistema: stats cards, jobs activos, refresh — completado 2026-03-24
- [x] Audit log: tabla de eventos filtrable (ya existía en /audit) — completado 2026-03-24

Criterio de done: admin puede crear usuario, asignar a área, cambiar config RAG y ver el audit log.

### Fase 3g — Settings + preferencias *(2-3 hs)*

- [x] Página de settings con secciones: Perfil, Contraseña, Preferencias — completado 2026-03-24
- [x] Server Actions: updateProfile, updatePassword, updatePreferences — completado 2026-03-24

Criterio de done: usuario puede cambiar nombre, contraseña y preferencias RAG. El cambio persiste.

---

## Fase 4 — CLI de clase mundial: apps/cli *(8-12 hs)*

Reemplaza y supera el CLI Python actual (424 líneas). La CLI es la única interfaz de administración del sistema desde la terminal. Habla con el servidor Next.js via API REST.

Stack: Bun + Commander (comandos y flags) + @clack/prompts (wizards interactivos) + chalk (colores) + cli-table3 (tablas).

- [x] Estructura base: entrypoint, Commander, manejo global de errores con sugerencias, cliente HTTP — completado 2026-03-24
- [x] Comandos de usuarios: listar (tabla), crear (wizard interactivo), eliminar (con confirmación) — completado 2026-03-24
- [x] Comandos de colecciones: listar, crear, eliminar — completado 2026-03-24
- [x] Comandos de ingesta: iniciar, ver status (tabla con barra de progreso), cancelar — completado 2026-03-24
- [x] Comandos de sistema: `rag status` (semáforo con latencias), `rag config get/set/reset` — completado 2026-03-24
- [x] Comandos de audit y black box: `rag audit log` con filtros, `rag audit replay`, `rag audit export` — completado 2026-03-24
- [x] Comandos de DB: `rag db migrate/seed/reset`, `rag setup` — completado 2026-03-24
- [x] Modo REPL interactivo: `rag` sin argumentos abre prompt con selector — completado 2026-03-24
- [x] Instalable globalmente via `bun link` (bin en package.json) — completado 2026-03-24

Criterio de done: `rag setup` completa el onboarding desde cero. Todos los comandos tienen output prolijo con colores y tablas.

---

## Fase 5 — Black box: logging + error reporting *(6-10 hs)*

- [x] `packages/logger/backend.ts`: niveles TRACE/DEBUG/INFO/WARN/ERROR/FATAL, escribe a tabla `events` + consola. Formato legible en dev, JSON en producción — completado 2026-03-24
- [x] `packages/logger/frontend.ts`: captura acciones del usuario y errores en el browser con batching hacia `/api/log` — completado 2026-03-24
- [x] `packages/logger/suggestions.ts`: mapeo de errores conocidos a mensajes accionables — completado 2026-03-24
- [x] `packages/logger/blackbox.ts`: `reconstructFromEvents()` + `formatTimeline()` para reconstruir estado del sistema — completado 2026-03-24
- [x] `apps/web`: GET /api/audit (con filtros), GET /api/audit/replay, GET /api/audit/export — completado 2026-03-24
- [x] `apps/web`: GET /api/health para health check de la CLI — completado 2026-03-24
- [x] `apps/web`: página de audit log con tabla filtrable — completado 2026-03-24
- [x] Instrumentación completa de todos los puntos críticos — completado 2026-03-24
- [x] Archivos de log físicos con rotación — completado 2026-03-24 (packages/logger/rotation.ts, rota en 10MB, 5 backups)

Criterio de done: después de simular un crash, `rag audit replay` reconstruye exactamente lo que pasó.

---

## Fase 6 — Docs + limpieza *(4-6 hs)*

- [x] `CHANGELOG.md` completo — se fue llenando durante toda la branch — completado 2026-03-24
- [x] `CLAUDE.md` actualizado: nuevo stack, nuevos comandos, nueva estructura — completado 2026-03-24
- [x] `docs/architecture.md`: diagrama del servidor único, flujos de auth y RAG, DB, caching — completado 2026-03-24
- [x] `docs/blackbox.md`: formato de eventos, cómo usar `rag audit replay`, sugerencias de errores — completado 2026-03-24
- [x] `docs/cli.md`: referencia completa de todos los comandos — completado 2026-03-24
- [x] `docs/onboarding.md`: guía de 5 minutos para nuevos colaboradores — completado 2026-03-24
- [x] `.gitignore` actualizado: `.next/`, `.turbo/`, `logs/`, `data/*.db` — completado 2026-03-24
- [x] `_archive/README.md`: código viejo documentado con motivo del archivado — completado 2026-03-24
- [x] `scripts/health-check.ts`: reemplaza `scripts/health_check.sh` — completado 2026-03-24

---

## Impacto esperado al completar el plan

- Procesos en producción: 2 reducidos a 1
- Lenguajes: 2 (Python + TypeScript) reducidos a 1 (TypeScript)
- Backend Python eliminado completamente (~6.4k líneas)
- Frontend reducido ~35% (~13k líneas Svelte a ~8-9k líneas TypeScript)
- Redis eliminado como dependencia
- Tipos duplicados eliminados (Zod compartido en `packages/shared`)
- Onboarding: de proceso manual complejo a un solo comando
- CLI: de 424 líneas Python limitadas a CLI completa con ~15 comandos y modo interactivo
- GitHub Actions: de 1 workflow básico a 3 workflows (CI, CD, release)

## Tiempo total estimado: 65-100 horas de trabajo

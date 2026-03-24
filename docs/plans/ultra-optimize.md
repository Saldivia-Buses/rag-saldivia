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

- [ ] Middleware Next.js: verifica JWT en cada request, aplica RBAC por ruta — *2 hs*
- [ ] Endpoints de login, logout y refresh: login emite JWT en cookie HttpOnly, logout invalida sesión en DB — *1.5 hs*
- [ ] Librería interna de auth: createJwt, verifyJwt, extractClaims, hasRole, canAccessCollection, getUserAreas — *1 hs*
- [ ] Tests del flujo de auth completo — *30 min*

Criterio de done: login funciona, cookie se setea, middleware bloquea rutas protegidas.

### Fase 3b — DB layer *(3-5 hs)*

- [ ] Integrar `packages/db` en `apps/web`. Conexión disponible en Server Components y Route Handlers — *1 hs*
- [ ] Portar todos los métodos de `saldivia/auth/database.py` (875 líneas) a queries Drizzle tipadas — *3 hs*
- [ ] Comandos `db:migrate` y `db:seed` funcionales — *1 hs*

Criterio de done: todas las queries del gateway Python tienen equivalente en Drizzle y los tests pasan.

### Fase 3c — RAG proxy + SSE streaming *(3-5 hs)*

- [ ] Route Handler que recibe la query del cliente, verifica permisos de colección, hace proxy SSE hacia RAG :8081, y reenvía el stream al cliente — *2 hs*
- [ ] Verificar el status HTTP del RAG antes de empezar a streamear (el gateway Python tenía un bug donde el status siempre era 200 aunque hubiera error) — *1 hs*
- [ ] Cliente HTTP interno para el RAG Server con retry configurable, timeout, y manejo de errores con sugerencias — *1 hs*
- [ ] Cache de 60 segundos para la lista de colecciones — *30 min*

Criterio de done: chat streaming funciona end-to-end. Los errores del RAG se propagan correctamente al cliente.

### Fase 3d — Collections + ingestion *(4-6 hs)*

- [ ] Páginas de colecciones: lista con cache, detalle por nombre — *1 hs*
- [ ] Página de upload: drag & drop de archivos, crea job en `ingestion_queue` — *1 hs*
- [ ] Endpoints de ingesta: status de jobs, cancelar, reintentar — *1 hs*
- [ ] Worker de ingesta en TypeScript: reemplaza `ingestion_worker.py` y `watch.py` — *2 hs*

Criterio de done: se puede subir un PDF, el job aparece en la DB y el worker lo procesa.

### Fase 3e — Chat UI *(5-8 hs)*

- [ ] Página de lista de sesiones (Server Component) — *1 hs*
- [ ] Página de chat específico: historial como Server Component, input y streaming como Client Component — *2 hs*
- [ ] Componente de streaming SSE: maneja las fases idle, streaming, done y error — *1 hs*
- [ ] Integración crossdoc: portar `useCrossdocDecompose.ts` y `useCrossdocStream.ts` de `patches/frontend/new/` directamente (son TypeScript puro, sin dependencias de Svelte) — *1 hs*
- [ ] Server Actions: crear sesión, renombrar, eliminar, feedback por mensaje — *1 hs*

Criterio de done: chat funciona con RAG estándar y crossdoc. Historial persiste. Feedback funciona.

### Fase 3f — Admin UI *(4-6 hs)*

- [ ] Gestión de usuarios: lista, crear con multi-select de áreas, eliminar — *1 hs*
- [ ] Gestión de áreas: CRUD completo, modal de bloqueo si hay usuarios activos — *1 hs*
- [ ] Permisos: asignación de colecciones a áreas con nivel read/write — *1 hs*
- [ ] Config RAG: sliders de parámetros, selector de modelo, toggle guardrails, switch de perfil — *1 hs*
- [ ] Estado del sistema: stats cards, jobs activos con progreso, alertas de ingesta — *1 hs*
- [ ] Audit log: tabla de eventos del black box con filtros por nivel, tipo, usuario y fecha — *1 hs*

Criterio de done: admin puede crear usuario, asignar a área, cambiar config RAG y ver el audit log.

### Fase 3g — Settings + preferencias *(2-3 hs)*

- [ ] Página de settings con 5 secciones: Perfil, Contraseña, Preferencias RAG, Notificaciones, Apariencia — *1 hs*
- [ ] Server Actions para cada sección: updateProfile, updatePassword, updatePreferences, updateNotifications — *1 hs*

Criterio de done: usuario puede cambiar nombre, contraseña y preferencias RAG. El cambio persiste.

---

## Fase 4 — CLI de clase mundial: apps/cli *(8-12 hs)*

Reemplaza y supera el CLI Python actual (424 líneas). La CLI es la única interfaz de administración del sistema desde la terminal. Habla con el servidor Next.js via API REST.

Stack: Bun + Commander (comandos y flags) + @clack/prompts (wizards interactivos) + chalk (colores) + cli-table3 (tablas).

- [ ] Estructura base: entrypoint, configuración de Commander, manejo global de errores con sugerencias, cliente HTTP al servidor — *1 hs*
- [ ] Comandos de usuarios: listar (tabla), crear (wizard interactivo), eliminar (con confirmación), asignar área — *1-2 hs*
- [ ] Comandos de colecciones: listar (tabla con stats), crear, eliminar (con confirmación si tiene docs) — *1 hs*
- [ ] Comandos de ingesta: iniciar, ver status (tabla con barra de progreso), cancelar, modo watch — *2 hs*
- [ ] Comandos de sistema: `rag status` (semáforo de servicios con latencias), `rag config get/set/reset` — *1 hs*
- [ ] Comandos de audit y black box: `rag audit log` con filtros, `rag audit replay` desde fecha, `rag audit export` — *1 hs*
- [ ] Comandos de DB y admin: `rag db migrate/seed/reset`, `rag setup`, `rag sessions list/delete` — *1 hs*
- [ ] Modo REPL interactivo: `rag` sin argumentos abre prompt con autocompletado — *1 hs*
- [ ] Instalación global via `bun link` para tener `rag` como comando del sistema — *15 min*

Criterio de done: `rag setup` completa el onboarding desde cero. Todos los comandos tienen output prolijo con colores y tablas.

---

## Fase 5 — Black box: logging + error reporting *(6-10 hs)*

- [ ] `packages/logger/backend.ts`: niveles TRACE/DEBUG/INFO/WARN/ERROR/FATAL, escribe a tabla `events` en SQLite y a archivo rotado. Formato legible en dev, JSON en producción — *2 hs*
- [ ] `packages/logger/frontend.ts`: captura acciones del usuario y errores en el browser con batching. Envía al endpoint `/api/log` del servidor — *1-2 hs*
- [ ] `packages/logger/suggestions.ts`: mapeo de errores conocidos a mensajes accionables. Al menos: ECONNREFUSED en RAG/Milvus, JWT expired, SQLite BUSY, colección no encontrada, puerto ocupado — *1 hs*
- [ ] `packages/logger/blackbox.ts`: función `reconstruct(fromTs, toTs)` que lee eventos en orden de secuencia y reconstruye el estado del sistema. Output: timeline + estado final + diff con la DB actual — *2-3 hs*
- [ ] Instrumentar todos los puntos críticos de `apps/web` con el logger — *1 hs*
- [ ] Tres archivos de log con rotación: `backend.log` (todo), `errors.log` (solo ERROR y FATAL), `frontend.log` (eventos del browser) — *30 min*

Criterio de done: después de simular un crash, `rag audit replay` reconstruye exactamente lo que pasó.

---

## Fase 6 — Docs + limpieza *(4-6 hs)*

- [ ] `CHANGELOG.md` revisado y completo con todo lo hecho en la branch (fue llenándose durante el proceso) — *1 hs*
- [ ] `CLAUDE.md` actualizado: nuevo stack, nuevos comandos, nueva estructura de carpetas — *1 hs*
- [ ] `docs/architecture.md`: diagrama del servidor único, flujo de auth, flujo de RAG — *1 hs*
- [ ] `docs/blackbox.md`: formato de eventos, cómo usar `rag audit replay`, casos de uso — *30 min*
- [ ] `docs/cli.md`: referencia completa de todos los comandos — *30 min*
- [ ] `docs/onboarding.md`: guía de 5 minutos para nuevos colaboradores — *30 min*
- [ ] `.gitignore` actualizado: `.next/`, `.turbo/`, `logs/*.log`, `data/*.db` — *15 min*
- [ ] Código viejo archivado: `saldivia/`, `services/sda-frontend/`, `cli/` movidos a `_archive/` con un README que explica por qué están ahí — *30 min*
- [ ] Scripts shell reemplazados por equivalentes TypeScript en `scripts/` — *1 hs*

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

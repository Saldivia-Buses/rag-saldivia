# Plan 11: Documentación Perfecta — README, CONTRIBUTING, API Docs, Packages

> Este documento vive en `docs/plans/ultra-optimize-plan11-documentation.md`.
> Se actualiza a medida que se completan las tareas. Cada fase completada genera entrada en CHANGELOG.md.

---

## Contexto

Los Planes 1–10 construyeron el stack técnico, el product roadmap, el design system, la suite de tests y la higiene del código. El resultado es un sistema sólido — pero **el repo no se explica solo**.

Hoy, alguien que clona el repo por primera vez tiene que:
- Leer el README de 89 líneas (incompleto)
- Abrir CLAUDE.md para entender la arquitectura
- Adivinar cómo configurar el entorno
- No tiene dónde leer la referencia de la API (30+ endpoints)
- No sabe qué convenciones de commit usar (CONTRIBUTING no existe)
- No puede reportar una vulnerabilidad (SECURITY no existe)
- No hay LICENSE

**El estándar de industria para un release público** exige que un developer pueda:
1. Entender qué es el proyecto en 30 segundos (README)
2. Tener el entorno funcionando en 5 minutos (Quick Start)
3. Saber exactamente cómo contribuir (CONTRIBUTING)
4. Encontrar cualquier endpoint sin buscar en el código (API Reference)
5. Entender el schema de DB visualmente (ER diagram)
6. Saber que existe una política de seguridad (SECURITY)

Este plan lleva el repo a ese estándar.

**Lo que NO cambia:** el código de producción, los tests, la lógica de negocio. Este plan es documentation-only excepto por JSDoc en funciones críticas.

---

## Prerequisito

**Plan 9 debe estar completado** antes de ejecutar este plan. La documentación refleja el estado limpio del código, no el estado con dead code.

---

## Orden de ejecución

```
F11.1 → F11.2 → F11.3 → F11.4 → F11.5 → F11.6 → F11.7 → F11.8 → F11.9
```

**Por qué este orden:**
- **F11.1 primero:** el README es lo primero que ve cualquier visitante del repo — y los badges del README dependen de que el CI, cobertura y versión estén definidos (Plan 10 y Release)
- **F11.2 y F11.3 juntos:** CONTRIBUTING y SECURITY son documentos cortos e independientes
- **F11.5 antes de F11.6:** el ER diagram del package `db` es prerequisito para el API Reference (las rutas referencian las tablas)
- **F11.8 + F11.9 al final:** actualizar CLAUDE.md y docs/ existentes — se hace al final para reflejar todos los cambios anteriores del plan

---

## Seguimiento

Formato: `- [ ] Descripción — estimación`
Al completar: `- [x] Descripción — completado YYYY-MM-DD`
Cada fase completada → entrada en CHANGELOG.md → commit.

---

## Fase F11.1 — README.md desde cero *(1-2 hs)*

Objetivo: cualquier developer que llega al repo por primera vez entiende qué es, cómo usarlo y cómo contribuir — sin leer ningún otro archivo.

**Archivo:** `README.md` (reemplazar el actual de 89 líneas)

**Estructura exacta — secciones en este orden:**

---

**Sección 1: Header**

Línea 1: `# RAG Saldivia`

Línea 3: badges en esta línea (usar shields.io):
- CI: `[![CI](https://github.com/Camionerou/rag-saldivia/actions/workflows/ci.yml/badge.svg?branch=experimental/ultra-optimize)](https://github.com/Camionerou/rag-saldivia/actions/workflows/ci.yml)`
- Version: `![Version](https://img.shields.io/badge/version-1.0.0-blue)`
- License: `![License](https://img.shields.io/badge/license-MIT-green)`
- Bun: `![Bun](https://img.shields.io/badge/bun-1.3%2B-orange)`

---

**Sección 2: Tagline** (1 línea)

`> Overlay sobre NVIDIA RAG Blueprint v2.5.0 — autenticación JWT, RBAC, multi-colección, admin y CLI TypeScript, en un único proceso Next.js 16.`

---

**Sección 3: Descripción** (3-4 párrafos)

Párrafo 1: qué es (overlay sobre el Blueprint de NVIDIA, no un fork). Qué agrega: auth, RBAC, multi-colección, admin completo, CLI, design system.

Párrafo 2: arquitectura de un solo proceso (Next.js reemplaza el gateway Python + SvelteKit del stack original). Mencionar Redis para JWT revocación, cache, notificaciones y colas de ingesta.

Párrafo 3: estado actual. Las features de UI, autenticación, admin y gestión de usuarios funcionan completamente. Las features de RAG (streaming contra LLMs, ingestión real de documentos en Milvus) requieren la workstation con GPU y estarán disponibles en versiones futuras.

---

**Sección 4: Arquitectura** (diagrama ASCII)

```
Usuario ──→ Next.js :3000 ──────────────────────────→ RAG Server :8081
             (UI + auth + proxy + admin)                      ↓
                      ↓                             Milvus + NIMs (GPU)
                 Redis :6379                                  ↓
           (JWT · cache · pub/sub · BullMQ)        Nemotron-Super-49B
```

Una nota debajo: "El RAG Server y Milvus son componentes del NVIDIA RAG Blueprint. Requieren hardware GPU. La app funciona con `MOCK_RAG=true` sin esos componentes."

---

**Sección 5: Requisitos**

Lista:
- Bun ≥ 1.3 — `curl -fsSL https://bun.sh/install | bash`
- Redis 7+ — `docker run -d -p 6379:6379 redis:alpine`
- Node.js ≥ 22 (opcional, para herramientas de CI)
- RAG Server (opcional para desarrollo — ver `MOCK_RAG`)

---

**Sección 6: Quick Start**

3 bloques de código:

```bash
# 1. Clonar y configurar
git clone https://github.com/Camionerou/rag-saldivia
cd rag-saldivia
git checkout experimental/ultra-optimize
cp .env.example .env.local
# Editar .env.local: agregar JWT_SECRET y verificar REDIS_URL

# 2. Instalar y configurar la DB
bun run setup

# 3. Iniciar el servidor de desarrollo
MOCK_RAG=true bun run dev
```

Después de los comandos:
```
Abrí http://localhost:3000
Credenciales de desarrollo: admin@localhost / changeme
```

---

**Sección 7: Estructura del monorepo** (tabla)

| Path | Descripción |
|---|---|
| `apps/web/` | Next.js 16 — UI + auth + proxy RAG + admin |
| `apps/cli/` | CLI TypeScript — `rag users/collections/ingest/audit/config/db/status` |
| `packages/db/` | Drizzle ORM + libsql — 12 tablas SQLite + Redis |
| `packages/logger/` | Logger estructurado + blackbox replay |
| `packages/shared/` | Schemas Zod compartidos entre web y CLI |
| `packages/config/` | Config loader con validación |

---

**Sección 8: Comandos principales** (tabla)

| Comando | Descripción |
|---|---|
| `bun run dev` | Servidor de desarrollo en :3000 |
| `bun run test` | Tests de lógica (259 tests) |
| `bun run test:components` | Tests de componentes React (154 tests) |
| `bun run test:visual` | Visual regression Playwright (22 tests) |
| `bun run test:a11y` | Auditoría WCAG AA (5 páginas) |
| `bun run test:e2e` | Tests E2E Playwright (flujos críticos) |
| `bun run storybook` | Catálogo de componentes en :6006 |
| `rag status` | Health check del sistema |

---

**Sección 9: Stack técnico** (tabla)

| Componente | Tecnología | Versión |
|---|---|---|
| Runtime | Bun | 1.3+ |
| Framework | Next.js App Router | 16.x |
| Base de datos | SQLite (Drizzle ORM + libsql) | — |
| Cola de tareas | BullMQ + Redis | — |
| Auth | JWT (jose) + Redis blacklist | — |
| Validación | Zod | 4.x |
| Build | Turborepo + Bun workspaces | — |
| CSS | Tailwind v4 | — |
| Componentes | shadcn/ui + Radix | — |
| Testing | bun:test + Playwright | — |

---

**Sección 10: Features en v1.0.0**

Lista de bullets de las features que funcionan ahora:
- Autenticación JWT con refresh y revocación inmediata (Redis)
- RBAC por roles (admin / area_manager / user) y áreas
- 24 páginas de UI con design system "Warm Intelligence"
- Admin completo: usuarios, áreas, permisos, configuración RAG
- CLI con 8 categorías de comandos
- Upload de documentos con cola de ingesta (BullMQ)
- Notificaciones en tiempo real (Redis Pub/Sub + SSE)
- Sesiones de chat con historial, guardados y etiquetas
- Proyectos con contexto persistente
- Dark mode + WCAG AA

---

**Sección 11: Features en versiones futuras**

`> Las siguientes features requieren la workstation con GPU y estarán disponibles en versiones futuras:`

- Streaming de respuestas del LLM (Nemotron-Super-49B)
- Ingestión real de documentos en Milvus
- Consulta multi-colección (crossdoc)
- Grafo de similitud entre documentos
- Vista dividida para comparación de respuestas
- SSO Google / Azure AD

---

**Sección 12: Contributing**

`Ver [CONTRIBUTING.md](CONTRIBUTING.md) para instrucciones de setup y convenciones de código.`

---

**Sección 13: Licencia**

`MIT — ver [LICENSE](LICENSE)`

---

**Criterio de done:**
- El README tiene todas las secciones en el orden de arriba
- Los comandos del Quick Start funcionan exactamente como están escritos
- Los badges tienen URLs válidas
- No hay menciones a `useCrossdocStream`, `SplitView`, `next-auth` ni ningún archivo eliminado en Plan 9

- [ ] Reemplazar `README.md` completo con el contenido de las 13 secciones descritas arriba
- [ ] Verificar que los badges apuntan a URLs correctas (repo Camionerou/rag-saldivia)
- [ ] Ejecutar los comandos del Quick Start para confirmar que funcionan
- [ ] Commit: `docs: reescribir readme desde cero — plan11 f11.1`

**Estado: pendiente**

---

## Fase F11.2 — CONTRIBUTING.md *(1-2 hs)*

Objetivo: cualquier developer puede hacer su primer PR siguiendo solo este documento.

**Archivo:** `CONTRIBUTING.md` (nuevo)

**Estructura requerida:**

1. **Setup del entorno de desarrollo** — paso a paso detallado:
   - Instalar Bun 1.3+
   - Clonar + checkout
   - Instalar Redis (comando Docker)
   - `cp .env.example .env.local` + qué variables configurar
   - `bun run setup` + `bun run dev`
   - Verificar que funciona: abrir `http://localhost:3000`

2. **Cómo correr los tests:**
   - Todos los tests: `bun run test`
   - Solo componentes: `bun run test:components`
   - Visual regression: `bun run test:visual`
   - A11y: `bun run test:a11y`
   - E2E: `bun run test:e2e`
   - Si un visual test falla intencionalmente: `bun run visual:update` + commitear nuevos snapshots

3. **Convenciones de commit (Conventional Commits):**
   - Formato: `type(scope): descripción en minúscula`
   - Tipos: feat, fix, refactor, chore, docs, test, ci, perf
   - El hook de `commitlint` bloquea mensajes inválidos
   - Ejemplos de buenos y malos commits

4. **Flujo de PR:**
   - Crear una branch desde `experimental/ultra-optimize`
   - Nombre de branch: `feat/descripcion` o `fix/descripcion`
   - Abrir PR contra `experimental/ultra-optimize`
   - El CI tiene que pasar completamente
   - El PR template tiene que estar completo

5. **Cómo agregar una página nueva:**
   - Crear el archivo en `apps/web/src/app/(app)/nueva-ruta/page.tsx`
   - Por defecto es Server Component — agregar `"use client"` solo si es necesario
   - Agregar la ruta al NavRail si es necesaria la navegación
   - Agregar la ruta al middleware si requiere autenticación
   - Agregar tests de componente en `src/components/__tests__/`

6. **Cómo agregar una ruta API nueva:**
   - Crear en `apps/web/src/app/api/nueva-ruta/route.ts`
   - Siempre verificar auth con `requireUser()` o `requireAdmin()`
   - Agregar la ruta a `docs/api.md`
   - Agregar test en el E2E si es un endpoint crítico

7. **Cómo agregar una tabla a la DB:**
   - Modificar `packages/db/src/schema.ts`
   - Agregar queries en `packages/db/src/queries/nueva-tabla.ts`
   - Correr `bunx drizzle-kit push` para aplicar el schema
   - Agregar tests en `packages/db/src/__tests__/nueva-tabla.test.ts`
   - Actualizar el ER diagram en `packages/db/README.md`

8. **Architecture Decision Records (ADRs):**
   - Cuándo crear uno: cuando se toma una decisión técnica con trade-offs no obvios
   - Copiar el template de `docs/decisions/000-template.md`
   - Numeración secuencial: el siguiente es `011-*.md`
   - Los ADRs son inmutables una vez merged — si cambia la decisión, crear uno nuevo que la supersede

9. **Debugging y troubleshooting:**
   - Qué hacer si los tests fallan
   - Qué hacer si Redis no conecta
   - Qué hacer si hay errores de tipos

**Criterio de done:**
- Un developer sin contexto puede configurar el entorno y hacer su primer PR siguiendo solo este documento
- Todos los comandos están verificados y funcionan

- [ ] Escribir CONTRIBUTING.md completo
- [ ] Verificar todos los comandos listados
- [ ] Commit: `docs: crear contributing.md — plan11 f11.2`

**Estado: pendiente**

---

## Fase F11.3 — SECURITY.md + LICENSE *(30 min)*

Objetivo: el repo tiene una política de seguridad visible y una licencia declarada.

**SECURITY.md** (nuevo):
- Cómo reportar una vulnerabilidad: email directo (no crear un issue público)
- Qué información incluir en el reporte (versión afectada, pasos para reproducir, impacto)
- Tiempo de respuesta esperado (best effort, sin SLA formal para un proyecto privado)
- Variables de entorno sensibles y cómo manejarlas (`JWT_SECRET`, `REDIS_URL`, `SYSTEM_API_KEY`)
- Qué NO commitear nunca: `.env.local`, archivos `*.key`, credenciales hardcodeadas

**LICENSE** (nuevo):
- MIT License
- Año 2026
- Titular del copyright

**Criterio de done:**
- `SECURITY.md` y `LICENSE` existen en el root del repo
- GitHub detecta la licencia automáticamente (aparece en el sidebar del repo)

- [ ] Crear `SECURITY.md`
- [ ] Crear `LICENSE` (MIT)
- [ ] Commit: `docs: agregar security.md y license mit — plan11 f11.3`

**Estado: pendiente**

---

## Fase F11.4 — CODEOWNERS + issue templates *(30 min)*

Objetivo: el repo tiene una estructura de revisión de código definida y los contribuidores pueden reportar bugs o pedir features de manera estructurada.

**CODEOWNERS** (nuevo en `.github/CODEOWNERS`):
- Define qué paths requieren review de quién
- Si el repo es personal (un solo maintainer), alcanza con `* @Camionerou`
- Paths críticos con revisión específica: `packages/db/`, `apps/web/src/middleware.ts`, `apps/web/src/lib/auth/`

**Issue templates** (nuevos en `.github/ISSUE_TEMPLATE/`):
- `bug_report.md` — plantilla con secciones: Descripción, Pasos para reproducir, Comportamiento esperado, Comportamiento actual, Entorno (OS, Bun version, Redis version), Logs relevantes
- `feature_request.md` — plantilla con secciones: Problema que resuelve, Solución propuesta, Alternativas consideradas, Contexto adicional

**Criterio de done:**
- `git push` y verificar que GitHub usa las templates automáticamente en "New Issue"
- CODEOWNERS activo (GitHub lo detecta automáticamente)

- [ ] Crear `.github/CODEOWNERS`
- [ ] Crear `.github/ISSUE_TEMPLATE/bug_report.md`
- [ ] Crear `.github/ISSUE_TEMPLATE/feature_request.md`
- [ ] Commit: `docs: codeowners + issue templates — plan11 f11.4`

**Estado: pendiente**

---

## Fase F11.5 — READMEs de packages y ER diagram *(1-2 hs)*

Objetivo: cada package del monorepo se entiende sin leer el código fuente.

**`packages/db/README.md`** (más complejo — requiere ER diagram):
- Qué es este package (Drizzle ORM + libsql + Redis)
- Cómo agregar una tabla nueva (3 pasos: schema → queries → test)
- **ER Diagram en Mermaid** con las 12 tablas y sus relaciones:
  - users → areas (N:M via user_areas)
  - chat_sessions → chat_messages (1:N)
  - chat_messages → message_tags (1:N)
  - chat_sessions → saved_responses (1:N)
  - users → chat_sessions (1:N)
  - projects → project_sessions, project_collections (1:N cada una)
  - users → user_memory (1:1)
  - external_sources (standalone)
  - events (log, standalone)
  - prompt_templates (standalone)
  - rate_limits (standalone)
  - shared_sessions (1:1 con chat_sessions)
- Queries disponibles (listado de funciones por archivo)
- Redis patterns usados: blacklist JWT, cache colecciones, sequence counter, log sizes
- Cómo correr los tests: `bun test packages/db/`

**`packages/logger/README.md`:**
- Qué es (logger estructurado + blackbox replay + rotación)
- Niveles disponibles: DEBUG, INFO, WARN, ERROR
- Cómo usar: `import { log } from "@rag-saldivia/logger"`
- Event types disponibles (lista de la tabla en `EventTypeSchema`)
- Black Box: qué es, cómo hacer replay, para qué sirve
- Rotación de archivos: cuándo rota, dónde quedan los logs, variable `LOG_RETENTION_DAYS`
- Cómo correr los tests: `bun test packages/logger/`

**`packages/shared/README.md`:**
- Qué es (schemas Zod compartidos entre web y CLI)
- Cuándo agregar algo aquí (regla: si lo usan dos apps, va en shared)
- Schemas principales: `UserSchema`, `RagParamsSchema`, `EventTypeSchema`, roles, focus modes
- Cómo correr los tests: `bun test packages/shared/`

**`packages/config/README.md`:**
- Qué es (config loader con validación Zod)
- Variables de entorno requeridas vs opcionales (con defaults)
- Cómo agregar una nueva variable de entorno

**`apps/cli/README.md`:**
- Qué es la CLI y para qué sirve
- Instalación global: `cd apps/cli && bun link`
- Referencia de todos los comandos con ejemplos:
  - `rag users list / create / update / delete`
  - `rag collections list`
  - `rag ingest status`
  - `rag audit log`
  - `rag config get / set`
  - `rag db reset / seed`
  - `rag status`

**Criterio de done:**
- 5 READMEs creados/actualizados
- El ER diagram en Mermaid renderiza correctamente en GitHub
- Cualquier developer puede entender qué hace cada package sin leer el código

- [ ] Crear `packages/db/README.md` con el ER diagram Mermaid
- [ ] Crear `packages/logger/README.md`
- [ ] Crear `packages/shared/README.md`
- [ ] Crear `packages/config/README.md`
- [ ] Crear `apps/cli/README.md`
- [ ] Verificar que el ER diagram renderiza en GitHub (Mermaid es soportado nativo)
- [ ] Commit: `docs: readme de packages y cli con er diagram — plan11 f11.5`

**Estado: pendiente**

---

## Fase F11.6 — docs/api.md — Referencia completa de la API *(1-2 hs)*

Objetivo: cualquier integrador (bot de Slack, cliente externo, script) puede construir su integración leyendo solo `docs/api.md`.

**Archivo:** `docs/api.md` (nuevo)

**Estructura:**

1. **Autenticación:** cómo autenticarse (POST /api/auth/login, formato del body, qué devuelve, cómo se usa la cookie)

2. **Endpoints por grupo** — para cada endpoint:
   - Método + ruta
   - Descripción en una línea
   - Autenticación requerida (ninguna / user / admin)
   - Body (si aplica) — campos con tipos
   - Response exitosa — estructura JSON
   - Códigos de error posibles

**Lista exacta de rutas existentes** (verificada con `ls apps/web/src/app/api/` y subdirectorios):

| Grupo | Rutas |
|---|---|
| Auth | `POST /api/auth/login`, `DELETE /api/auth/logout`, `POST /api/auth/refresh` |
| RAG | `POST /api/rag/generate` (SSE), `GET /api/rag/collections`, `POST /api/rag/suggest` |
| Upload | `POST /api/upload` |
| Search | `GET /api/search?q=...` |
| Admin — Users | `GET /api/admin/users`, `POST /api/admin/users`, `PATCH /api/admin/users/[id]`, `DELETE /api/admin/users/[id]` |
| Admin — Areas | `GET /api/admin/areas`, `POST /api/admin/areas`, `PATCH /api/admin/areas/[id]`, `DELETE /api/admin/areas/[id]` |
| Admin — Permissions | `GET /api/admin/permissions`, `PUT /api/admin/permissions` |
| Admin — Config | `GET /api/admin/config`, `PUT /api/admin/config` |
| Admin — Templates | `GET /api/admin/templates`, `POST /api/admin/templates`, `PATCH /api/admin/templates/[id]`, `DELETE /api/admin/templates/[id]` |
| Admin — Ingestion | `GET /api/admin/ingestion`, `DELETE /api/admin/ingestion/[id]`, `GET /api/admin/ingestion/stream` (SSE) |
| Admin — Analytics | `GET /api/admin/analytics` |
| Admin — DB | `POST /api/admin/db/reset` |
| Notifications | `GET /api/notifications`, `GET /api/notifications/stream` (SSE) |
| Memory | `GET /api/memory`, `POST /api/memory`, `DELETE /api/memory` |
| Collections | `GET /api/collections` |
| Extract | `POST /api/extract` |
| Health | `GET /api/health` |
| Integrations | `POST /api/slack`, `POST /api/teams` |
| Log | `POST /api/log` |
| Audit | `GET /api/audit` |
| Share | `GET /api/share/[token]` |

**Verificar antes de documentar:** confirmar que cada ruta existe realmente ejecutando:
```bash
ls apps/web/src/app/api/
ls apps/web/src/app/api/admin/
ls apps/web/src/app/api/rag/
ls apps/web/src/app/api/auth/
```
Si alguna ruta de la tabla no existe, eliminarla del documento.

3. **SSE (Server-Sent Events):** qué rutas devuelven streams SSE, formato de los eventos, cómo consumirlos con `EventSource`

4. **Errores comunes:** tabla de códigos de error y sus causas

5. **Nota sobre features RAG:** los endpoints que llaman al RAG server (`/api/rag/generate`, `/api/rag/collections`, `/api/rag/suggest`) requieren que el RAG server esté disponible. En desarrollo se puede usar `MOCK_RAG=true`. Las features de streaming completo contra Milvus están disponibles en versiones futuras con la workstation.

**Criterio de done:**
- Todos los endpoints listados tienen documentación completa
- Un developer puede construir un cliente sin mirar el código fuente
- Los endpoints de SSE tienen ejemplo de consumo

- [ ] Crear `docs/api.md` con todos los endpoints
- [ ] Verificar que cada endpoint documentado existe realmente en `apps/web/src/app/api/`
- [ ] Verificar que no hay endpoints en el código que no estén en el doc
- [ ] Commit: `docs: api.md — referencia completa de los 30+ endpoints — plan11 f11.6`

**Estado: pendiente**

---

## Fase F11.7 — JSDoc en funciones críticas *(1-2 hs)*

Objetivo: las funciones que cualquier contributor va a tocar tienen documentación inline que explica el *por qué*, no el *qué*.

**Criterio de selección:** solo las funciones que son no-obvias, tienen comportamiento sorpresivo, o tienen constrains críticos de seguridad/performance. No documentar getters simples ni funciones de 3 líneas que se explican solas.

**JSDoc exacto para cada función:**

El formato a usar en cada caso:

```typescript
/**
 * [Una línea: qué hace]
 *
 * [Por qué el comportamiento es no-obvio o cuál es el riesgo si se cambia]
 *
 * @param [param] - [descripción si no es obvio]
 * @returns [descripción si no es obvio]
 */
```

**Función 1 — `getRedisClient()` en `packages/db/src/redis.ts`:**
```
Retorna el cliente Redis singleton. Lanza error en el primer acceso si REDIS_URL no está configurado (fail-fast en startup, no lazy). No usar en Edge runtime — solo en Node.js. Para BullMQ, usar getBullMQConnection() en lugar de este singleton.
```

**Función 2 — `nextSequence()` en `packages/db/src/queries/events.ts`:**
```
Genera el siguiente número de secuencia monotónico usando Redis INCR en la key "events:seq". Si Redis falla, esta función lanza — el evento no se graba. Los números no son necesariamente consecutivos entre reinicios si Redis pierde estado.
```

**Función 3 — `createJwt()` en `apps/web/src/lib/auth/jwt.ts`:**
```
Crea un JWT firmado con jti (JWT ID) único. El jti es requerido para que el logout pueda revocar el token en Redis. Si se elimina el setJti(), el logout dejará de funcionar inmediatamente. El jti se propaga en el header x-user-jti por el middleware.
```

**Función 4 — `extractClaims()` en `apps/web/src/lib/auth/jwt.ts`:**
```
Verifica el JWT y chequea la blacklist Redis. Corre en Node.js runtime (en route handlers). NO corre en middleware.ts (Edge runtime) porque ioredis no es compatible con Edge. Si se necesita verificar revocación en middleware, usar el header x-user-jti que el proxy propaga.
```

**Función 5 — `ragFetch()` en `apps/web/src/lib/rag/client.ts`:**
```
Hace fetch al RAG Server y verifica el status HTTP ANTES de leer el body. Crítico: el servidor RAG puede retornar 200 aunque haya un error en el stream. Sin esta verificación, los errores del RAG server se propagan como texto en el stream sin ser detectados. Con MOCK_RAG=true, no llama al servidor real.
```

**Función 6 — `getCachedRagCollections()` en `apps/web/src/lib/rag/collections-cache.ts`:**
```
Obtiene la lista de colecciones desde Redis (TTL: 60s). IMPORTANTE: llamar invalidateCollectionsCache() después de cualquier POST o DELETE en /api/rag/collections — de lo contrario la UI muestra datos stale por hasta 60 segundos.
```

**Función 7 — `startIngestionWorker()` en `apps/web/src/lib/queue.ts`:**
```
Inicia el worker de BullMQ. Es idempotente en producción — si se llama dos veces, BullMQ lanza advertencia. En el servidor Next.js, este worker se inicia una sola vez en la inicialización del módulo. No llamar desde route handlers.
```

**Función 8 — middleware en `apps/web/src/proxy.ts`:**
```
Entry point del middleware de Next.js. Genera x-request-id UUID para trazabilidad de logs por request. Propaga x-user-jti para que los route handlers puedan verificar revocación de JWT. Corre en Edge runtime — sin acceso a ioredis, base de datos, ni fs.
```

**Función 9 — `writeEvent()` en `packages/logger/src/backend.ts`:**
```
Graba un evento en la tabla events de SQLite. El campo type debe ser un EventType válido del EventTypeSchema — si se pasa un string arbitrario, falla silenciosamente en el validator y el evento se graba con type="unknown". Siempre importar los tipos de @rag-saldivia/shared.
```

**Función 10 — `reconstructFromEvents()` en `packages/logger/src/blackbox.ts`:**
```
Reconstruye el estado del sistema desde los eventos de la DB (blackbox replay). Lee TODOS los eventos en memoria — no usar en producción con más de 100k eventos sin paginar. Solo muestra el estado hasta el evento más reciente procesado, no puede "predecir" estado futuro.
```

**Criterio de done:**
- Cada una de las 10 funciones tiene un JSDoc
- Los comentarios están en español (consistente con el resto del codebase)
- `export PATH="$HOME/.bun/bin:$PATH" && bun run test` → 259 tests pasan (JSDoc no rompe nada)

- [ ] Agregar JSDoc a `getRedisClient()` en `packages/db/src/redis.ts`
- [ ] Agregar JSDoc a `nextSequence()` en `packages/db/src/queries/events.ts`
- [ ] Agregar JSDoc a `createJwt()` en `apps/web/src/lib/auth/jwt.ts`
- [ ] Agregar JSDoc a `extractClaims()` en `apps/web/src/lib/auth/jwt.ts`
- [ ] Agregar JSDoc a `ragFetch()` en `apps/web/src/lib/rag/client.ts`
- [ ] Agregar JSDoc a `getCachedRagCollections()` en `apps/web/src/lib/rag/collections-cache.ts`
- [ ] Agregar JSDoc a `startIngestionWorker()` en `apps/web/src/lib/queue.ts`
- [ ] Agregar JSDoc al middleware en `apps/web/src/proxy.ts`
- [ ] Agregar JSDoc a `writeEvent()` en `packages/logger/src/backend.ts`
- [ ] Agregar JSDoc a `reconstructFromEvents()` en `packages/logger/src/blackbox.ts`
- [ ] `export PATH="$HOME/.bun/bin:$PATH" && bun run test` → todos pasan
- [ ] Commit: `docs: jsdoc en funciones criticas — plan11 f11.7`

**Estado: pendiente**

---

## Fase F11.8 — CLAUDE.md actualizado *(30-45 min)*

Objetivo: CLAUDE.md refleja el estado real del repo post-Plan 9 (sin referencias a archivos eliminados, con los nuevos patrones del Plan 8).

**Secciones a actualizar:**

- **Stack técnico:** Next.js 15.x → 16.x, Zod 3.x → 4.3.x, Drizzle 0.38 → 0.45
- **Comandos clave:** agregar `bun run test:e2e`, `bun run lint`, remover cualquier referencia a comandos obsoletos
- **Patrones importantes — Redis:** ya está documentado (Plan 8) — verificar que sigue siendo correcto
- **Patrones importantes:** remover referencias a `next-safe-action`, `form.ts`, `useCrossdocStream` — eliminados en Plan 9
- **Hooks de React:** remover `useCrossdocStream` y `useCrossdocDecompose` de la tabla
- **Componentes sin tests:** actualizar la lista (algunos pueden haberse cubierto en Plan 10)
- **Plans de implementación:** agregar Planes 9, 10, 11 como completados
- **Planes futuros:** nota sobre features RAG/GPU para versiones futuras

**Criterio de done:**
- Cero referencias a archivos eliminados en el Plan 9
- La tabla de hooks refleja los hooks reales del repo
- Los planes completados están marcados

- [ ] Revisar CLAUDE.md sección por sección contra el estado real del repo
- [ ] Actualizar cada sección desactualizada
- [ ] Commit: `docs: actualizar claude.md post-plan9 — plan11 f11.8`

**Estado: pendiente**

---

## Fase F11.9 — docs/ existentes revisados *(30-45 min)*

Objetivo: todos los documentos en `docs/` son precisos y no tienen información stale.

**Documentos a revisar:**

- **`docs/architecture.md`:** verificar que refleja Next.js 16, Redis obligatorio, BullMQ, los ADRs 008-010. Ya fue actualizado en Plan 8 F5 — hacer un pass de precisión.
- **`docs/onboarding.md`:** agregar Redis como prerequisito. Verificar que los 5 minutos del onboarding siguen siendo correctos post-Plan 8.
- **`docs/workflows.md`:** verificar que el flujo de git refleja el commitlint del Plan 9. Actualizar si hay pasos obsoletos.
- **`docs/testing.md`:** actualizar con los nuevos comandos `test:e2e`, `test:a11y`, cobertura. Reflejar los 5 suites del Plan 10.
- **`docs/blackbox.md`:** verificar precisión post-Plan 8 F7 (requestId, nuevos event types, rotación via Redis).
- **`docs/cli.md`:** verificar que todos los comandos de la CLI existen realmente.
- **`docs/design-system.md`:** verificar que los tokens CSS documentados coinciden con `globals.css` actual.

**Criterio de done:**
- Cada documento en `docs/` fue leído y está actualizado
- No hay referencias a features, archivos o comandos que ya no existen

- [ ] Revisar `docs/architecture.md`
- [ ] Revisar `docs/onboarding.md` — agregar Redis como prerequisito
- [ ] Revisar `docs/workflows.md`
- [ ] Revisar `docs/testing.md` — agregar test:e2e y cobertura
- [ ] Revisar `docs/blackbox.md`
- [ ] Revisar `docs/cli.md`
- [ ] Revisar `docs/design-system.md`
- [ ] Commit: `docs: revisar y actualizar documentos existentes — plan11 f11.9`

**Estado: pendiente**

---

## Criterio de done global del Plan 11

- `README.md` ≥ 300 líneas con todos los badges, Quick Start, arquitectura, stack, features v1.0.0 y features futuras
- `CONTRIBUTING.md` existe y tiene los 9 pasos documentados
- `SECURITY.md` existe
- `LICENSE` existe y GitHub lo detecta automáticamente
- `CODEOWNERS` existe
- 2 issue templates en `.github/ISSUE_TEMPLATE/`
- 5 READMEs de packages/apps creados
- `docs/api.md` cubre los 30+ endpoints
- JSDoc en las 10 funciones críticas
- `CLAUDE.md` sin referencias a código eliminado
- Todos los documentos en `docs/` son precisos

### Checklist de cierre

- [ ] README.md completo con badges y Quick Start verificado
- [ ] CONTRIBUTING.md completo
- [ ] SECURITY.md + LICENSE creados
- [ ] CODEOWNERS + issue templates creados
- [ ] 5 READMEs de packages con ER diagram verificado en GitHub
- [ ] `docs/api.md` completo
- [ ] JSDoc en funciones críticas
- [ ] CLAUDE.md actualizado
- [ ] docs/ revisados
- [ ] CHANGELOG.md actualizado con entrada del Plan 11
- [ ] Commit final: `docs: plan11 completado — documentacion perfecta`
- [ ] `git push`

**Estado: pendiente**

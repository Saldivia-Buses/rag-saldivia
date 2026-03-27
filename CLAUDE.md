# RAG Saldivia — Contexto de proyecto

## Qué es este proyecto

Overlay sobre **NVIDIA RAG Blueprint v2.5.0** que agrega autenticación, RBAC, multi-colección, frontend Next.js 16, CLI TypeScript, design system completo, y suite de testing UI.

- **No es un fork** — incluye el blueprint como git submodule en `vendor/rag-blueprint/` (commit a67a48c)
- **Repo local:** `~/rag-saldivia/` — branch `experimental/ultra-optimize`
- **Repo remoto:** https://github.com/Camionerou/rag-saldivia
- **Deploy activo:** workstation física Ubuntu 24.04 (1x RTX PRO 6000 Blackwell, 96GB VRAM)

> **Nota:** La branch `main` tiene el stack Python+SvelteKit (estable en producción).
> Esta branch (`experimental/ultra-optimize`) es la reescritura completa en TypeScript.

---

## Arquitectura de servicios

```
Usuario → Next.js :3000 ──────────────────→ RAG Server :8081
           (UI + auth + proxy)                      ↓
                                           Milvus + NIMs
                                                    ↓
                                       Nemotron-Super-49B
```

**Un único proceso** reemplaza el gateway Python (9000) + el frontend SvelteKit (3000).

---

## Estructura del monorepo

```
apps/
  web/              → Next.js 16 — servidor único (UI + auth + proxy RAG + admin)
  cli/              → CLI TypeScript (rag users/collections/ingest/audit/config/db)
packages/
  shared/           → Zod schemas + tipos compartidos (User, Area, Session, etc.)
  db/               → Drizzle ORM + @libsql/client (esquema SQLite multi-tabla + Redis)
  config/           → config loader TypeScript (reemplaza config.py)
  logger/           → logger estructurado + black box replay + rotación de archivos
scripts/
  setup.ts          → onboarding cero-fricción (bun run setup)
docs/
  architecture.md   → arquitectura completa del stack
  design-system.md  → guía del design system "Warm Intelligence"
  testing.md        → guía completa de testing (unit + component + visual + E2E)
  blackbox.md       → sistema de logging y replay
  cli.md            → referencia completa de la CLI
  onboarding.md     → guía de 5 minutos
  workflows.md      → flujos de trabajo del proyecto (git, tests, features, deploy)
  plans/            → planes de implementación completados y en curso
```

---

## Stack técnico

| Componente | Tecnología | Versión |
|---|---|---|
| Lenguaje | TypeScript | 6.0 |
| Runtime | Bun | 1.3.x |
| Framework web | Next.js App Router | 16.x |
| Base de datos | SQLite vía Drizzle ORM + @libsql/client | — |
| Auth | JWT (jose) en cookie HttpOnly | — |
| Validación | Zod (compartido entre todos los paquetes) | 4.3.x |
| Build | Turborepo + Bun workspaces | — |
| CLI | Commander + @clack/prompts + chalk | — |
| CSS | Tailwind v4 + @tailwindcss/postcss | 4.x |
| Componentes UI | shadcn/ui + Radix + componentes propios | — |
| Tipografía | Instrument Sans (next/font/google) | — |
| Testing unitario | bun:test | — |
| Testing componentes | @testing-library/react + happy-dom | — |
| Testing visual | Playwright (screenshots de Storybook) | 1.x |
| Catálogo UI | Storybook 8 + react-vite | — |
| Performance UI | react-scan | 0.5.x |

---

## Comandos clave

```bash
# Onboarding (primera vez)
bun run setup

# Desarrollo
bun run dev              # Next.js en :3000
bun run storybook        # Storybook en :6006

# Tests — por capa
bun run test                    # todos los tests (lógica) via Turborepo (~311 tests)
bun test apps/web/src/lib/      # solo lib/ de web
bun test packages/db/           # solo DB queries (~167 tests)
bun run test:components         # component tests con happy-dom (~153 tests)
bun run test:visual             # visual regression Playwright (22 tests)
bun run visual:update           # regenerar baseline de screenshots
bun run test:a11y               # auditoría WCAG AA con axe-playwright
bun run test:e2e                # E2E Playwright (apps/web)
bun run test:ui                 # test:components + test:visual
bun run lint                    # lint vía Turborepo

# Health check
rag status

# Deploy en workstation física (stack Python, branch main)
cd ~/rag-saldivia && make deploy PROFILE=workstation-1gpu

# CLI (instalar global: cd apps/cli && bun link)
rag users list
rag collections list
rag ingest status
rag audit log
rag status
```

---

## Design System "Warm Intelligence"

El proyecto tiene un design system completo aplicado a las 24 páginas de la app.

### Tokens CSS

```css
/* Light mode */
--bg:          #faf8f4   /* crema cálido */
--surface:     #f0ebe0
--surface-2:   #e8e1d4
--border:      #ede9e0
--fg:          #1a1a1a
--fg-muted:    #5a5048
--fg-subtle:   #9a9088
--accent:      #1a5276   /* navy profundo */
--accent-subtle: #d4e8f7
--success:     #2d6a4f
--destructive:  #c0392b
--warning:     #d68910

/* Dark mode (.dark) — warm dark, nunca negro frío */
--bg:          #1a1812
--surface:     #24221a
--accent:      #4a9fd4   /* navy más claro para contraste */
```

### Tipografía

Instrument Sans via `next/font/google`. Variable CSS: `--font-instrument-sans`.
Letter-spacing `-0.01em` en body para mejor apariencia.

### Densidad adaptiva

```html
<!-- Admin/tablas: más información visible -->
<div data-density="compact">...</div>

<!-- Chat/collections: más espacio visual -->
<div data-density="spacious">...</div>
```

### Componentes propios (en `apps/web/src/components/ui/`)

| Componente | Descripción |
|---|---|
| `Button` | 6 variantes: default, destructive, outline, secondary, ghost, link |
| `Badge` | 6 variantes: default, secondary, outline, destructive, success, warning |
| `Input` / `Textarea` | inputs con tokens propios |
| `DataTable` | tabla avanzada con sorting, filtro, paginación (@tanstack/react-table) |
| `StatCard` | tarjeta de estadísticas con delta positivo/negativo |
| `EmptyPlaceholder` | estados vacíos con ícono, título, descripción |
| `Skeleton` / `SkeletonText` / `SkeletonTable` | estados de carga shimmer |

### Storybook

```bash
bun run storybook        # dev en :6006
bun run build:storybook  # build estático en storybook-static/
```

Stories en `apps/web/stories/` organizadas por: `design-system/`, `primitivos/`, `features/`, `layout/`.

Addons activos: **addon-a11y** (WCAG por componente), **addon-themes** (toggle light/dark), **addon-essentials**.

---

## Testing

### Suite actual

| Capa | Comando | Tests | Cobertura |
|---|---|---|---|
| Lógica pura (lib/, db, config, logger) | `bun run test` | ~311 | CI + threshold en paquetes |
| Componentes React | `bun run test:components` | ~153 | Storybook + RTL |
| Visual regression | `bun run test:visual` | 22 | baseline versionado |
| A11y WCAG AA | `bun run test:a11y` | páginas clave | axe-playwright |
| E2E | `bun run test:e2e` | flujos críticos | Playwright (MOCK_RAG) |

### Convención de tests de componentes

```typescript
// Cada archivo de test de componente DEBE tener:
import { afterEach } from "bun:test"
import { cleanup } from "@testing-library/react"

afterEach(cleanup)  // obligatorio — evita contaminación entre tests

// Usar queries escopadas al render, NO screen global:
const { getByRole } = render(<Button>Click</Button>)
// ✅ correcto: getByRole("button")
// ❌ incorrecto: screen.getByRole("button")  — puede encontrar elementos de otros tests
```

### Preloads de test

- **`src/lib/test-setup.ts`** — carga mocks de next/navigation, next/font, next-themes para TODOS los tests
- **`src/lib/component-test-setup.ts`** — agrega GlobalRegistrator (happy-dom) para component tests

```bash
# Tests de componentes (con happy-dom):
bun test --preload ./src/lib/component-test-setup.ts src/components

# Tests de lib (sin happy-dom, más rápidos):
bun test src/lib
```

### Visual regression

El baseline se genera una vez y se commitea:
```bash
bun run visual:update   # genera snapshots de referencia
bun run test:visual     # compara contra baseline
```

Si un cambio de diseño es intencional, regenerar el baseline con `visual:update` y commitear los nuevos PNGs.

---

## Archivos críticos — leer antes de modificar

| Archivo | Por qué es crítico |
|---|---|
| `apps/web/src/middleware.ts` | Re-export: delega en `proxy.ts` (JWT + RBAC en el edge) |
| `apps/web/src/proxy.ts` | Implementación del middleware (auth, roles, `x-request-id`, `x-user-jti`) |
| `apps/web/src/lib/auth/jwt.ts` | createJwt, verifyJwt, cookies |
| `apps/web/src/app/globals.css` | tokens CSS del design system — cambios afectan toda la app |
| `apps/web/src/lib/component-test-setup.ts` | setup de happy-dom — modificar puede romper tests de componentes |
| `packages/db/src/schema.ts` | schema SQLite completo (muchas tablas) |
| `packages/db/src/queries/users.ts` | CRUD de usuarios + permisos |
| `packages/logger/src/backend.ts` | logger con rotación de archivos |
| `apps/web/src/lib/rag/client.ts` | proxy RAG con modo mock |
| `apps/web/.storybook/main.ts` | config de Storybook — viteFinal y addons |
| `apps/web/playwright.config.ts` | config visual regression |

---

## Variables de entorno

Ver `.env.example` para la lista completa documentada.

```env
JWT_SECRET=...             # openssl rand -base64 32
SYSTEM_API_KEY=...         # openssl rand -hex 32
RAG_SERVER_URL=http://localhost:8081
DATABASE_PATH=./data/app.db
MOCK_RAG=false             # true para dev sin Docker
LOG_LEVEL=INFO

# Para tests E2E con Maestro
TEST_ADMIN_EMAIL=admin@localhost
TEST_ADMIN_PASSWORD=changeme
```

### Credenciales de desarrollo (seed)

| Email | Password | Rol |
|---|---|---|
| `admin@localhost` | `changeme` | admin |
| `user@localhost` | `test1234` | user |

---

## Patrones importantes

### CSS y design system
- **Tokens CSS siempre** — nunca hardcodear colores. Usar `var(--accent)`, `text-fg-muted`, `bg-surface`
- **`@theme inline`** en Tailwind v4 — crítico para que el dark mode class-based funcione en runtime
- **`afterEach(cleanup)`** en cada archivo de test de componente — sin esto los tests se contaminan
- **`bg-surface` vs `bg-bg`** — `surface` para cards/paneles elevados, `bg` para el fondo base

### Tests
- **`afterEach(cleanup)` por archivo** — obligatorio en component tests para aislar el DOM
- **Queries escopadas** — usar `const { getByRole } = render(...)` en lugar de `screen.getByRole`
- **`fireEvent` sobre `userEvent`** — en happy-dom, userEvent tiene problemas de compatibilidad
- **`bun run test:components`** — usa `--preload component-test-setup.ts` que activa happy-dom

### Redis (ADR-010)
- **Redis es dependencia requerida** — `getRedisClient()` en `packages/db/src/redis.ts` retorna `Redis` directo, nunca null. Lanza error claro si `REDIS_URL` no configurado
- **NO importar logger en redis.ts** — crearía dependencia circular `db → logger → db` (ADR-005). Usar `console.error` para errores de conexión en redis.ts
- **BullMQ usa conexiones propias** — pasar `{ maxRetriesPerRequest: null }` y crear instancias separadas via `getBullMQConnection()`, no reutilizar el singleton de `getRedisClient()`
- **JWT revocación** — verificada en `extractClaims()` (Node.js runtime), NO en el middleware Edge (`proxy.ts`) sin ioredis. El jti se propaga via header `x-user-jti`
- **Notificaciones** — publisher: `getRedisClient().publish("notifications:{userId}", ...)` en el worker BullMQ. Subscriber: endpoint SSE `/api/notifications/stream` crea conexión dedicada con `.duplicate()`
- **Cache de colecciones** — `getCachedRagCollections()` en `lib/rag/collections-cache.ts`. Llamar `invalidateCollectionsCache()` después de POST/DELETE en `/api/rag/collections`
- **Para tests** — ioredis-mock activo via preload (`packages/db/bunfig.toml` y `apps/web/bunfig.toml`). Tests unitarios no requieren Redis real

### Next.js y React
- **Server Components por defecto** — cero JS al browser salvo donde sea necesario
- **`"use client"`** solo en componentes que necesitan estado, efectos, o APIs de browser
- **`next/font/google`** para Instrument Sans — hace self-hosting automático y optimiza preload
- **`next/dynamic` con `ssr: false`** solo puede usarse en Client Components, no Server Components
- **`lib/rag/stream.ts`** es la utilidad canónica de SSE — nunca duplicar la lógica de lectura de stream

### Base de datos
- **Temporal API** para todos los timestamps — `Date.now()` en lugar del bug `_ts()` de SQLite
- **Ingesta via BullMQ** — `ingestionQueue.add("ingest", { filePath, collection, userId })` en `upload/route.ts`. La tabla `ingestion_queue` SQLite fue eliminada en Plan 8 F8.30
- **CJS sobre ESM** — paquetes `packages/*` sin `"type": "module"` para compatibilidad con webpack

### Logger
- **`@rag-saldivia/db` importado estáticamente** en `packages/logger` — import dinámico fallaba silenciosamente en webpack/Next.js

### SSE / Streaming
- **Verificar status HTTP ANTES de streamear** — el gateway antiguo siempre retornaba 200

### Tailwind v4 específico
- Requiere `postcss.config.js` con `@tailwindcss/postcss` — sin él, las utility classes custom no se generan
- Las nuevas clases custom (como `bg-surface`) solo se generan si aparecen en archivos escaneados por Tailwind

---

## Architecture Decision Records (ADRs)

En `docs/decisions/` hay decisiones de arquitectura documentadas. **Leerlas antes de modificar las áreas que cubren:**

| ADR | Decisión | Área afectada |
|---|---|---|
| 001 | libsql sobre better-sqlite3 | `packages/db/` |
| 002 | CJS sobre ESM en packages | `packages/*/tsconfig.json` |
| 003 | Next.js como proceso único | arquitectura general |
| 004 | Temporal API para timestamps | toda query con fechas |
| 005 | Import estático de db en logger | `packages/logger/src/backend.ts` |
| 006 | Estrategia de testing | toda la suite de tests |
| 007 | Funciones reales sobre helpers locales en tests | `packages/db/src/__tests__/` |
| 008 | Lector SSE compartido | `apps/web/src/lib/rag/stream.ts` |
| 009 | Server Components primero | UI y data fetching |
| 010 | Redis como dependencia requerida | `packages/db/src/redis.ts`, BullMQ, JWT revocación, cache |
| 011 | UI series 1.0.x — strategy Claude-style | `apps/web/src/app/globals.css`, todos los componentes UI |

---

## Workers de background

En `apps/web/src/workers/` — se ejecutan como edge functions o serverless functions:

| Worker | Descripción |
|---|---|
| `ingestion.ts` | Procesa la cola de ingesta de documentos — desbloquea jobs, llama al blueprint NVIDIA |
| `external-sync.ts` | Sincroniza fuentes externas (Google Drive, SharePoint, Confluence) según schedule |

**Importante:** Los workers usan SQLite con locking optimista (`locked_at`). Si un worker muere con un job locked, el job queda bloqueado hasta que expire el TTL.

---

## Hooks de React

En `apps/web/src/hooks/` — todos Client Components:

| Hook | Complejidad | Descripción |
|---|---|---|
| `useRagStream` | Alta (19) | Streaming SSE del RAG — maneja fases: idle → streaming → done. Detecta artifacts automáticamente |
| `useGlobalHotkeys` | Media | Hotkeys globales (Cmd+K para CommandPalette, etc.) |
| `useNotifications` | Media | Notificaciones (`/api/notifications` + SSE `/api/notifications/stream`) |
| `useZenMode` | Baja | Toggle del modo zen (oculta NavRail y SecondaryPanel) |

**Patrón crítico de useRagStream:** verifica el status HTTP **antes** de leer el stream. Si la respuesta no es 200, el stream devuelve un error — no asumir que siempre es exitoso.

---

## Server Actions

En `apps/web/src/app/actions/` — se ejecutan en el servidor, llamadas desde Client Components:

| Archivo | Actions principales |
|---|---|
| `chat.ts` | `actionCreateSession`, `actionDeleteSession`, `actionRenameSession`, `actionAddMessage`, `actionAddFeedback`, `actionToggleSaved`, `actionForkSession`, `actionAddTag`, `actionRemoveTag`, `actionCreateSessionForDoc` |
| `users.ts` | `actionCreateUser`, `actionUpdateUser`, `actionDeleteUser`, `actionListUsers`, `actionAssignArea`, `actionRemoveArea` |
| `areas.ts` | `actionCreateArea`, `actionUpdateArea`, `actionDeleteArea`, `actionListAreas`, `actionSetAreaCollections` |
| `settings.ts` | `actionUpdateProfile`, `actionUpdatePassword`, `actionUpdatePreferences` |
| `config.ts` | `actionGetRagParams`, `actionUpdateRagParams`, `actionResetRagParams` |

**Patrón:** las Server Actions en este proyecto usan `revalidatePath()` implícitamente o actualizan el estado local del componente. Siempre verificar si necesitan `revalidatePath` después de mutaciones.

---

## API Routes

En `apps/web/src/app/api/` — más de 30 routes. Las más críticas:

| Ruta | Método | Descripción |
|---|---|---|
| `/api/auth/login` | POST | Autenticación JWT — devuelve cookie HttpOnly |
| `/api/auth/logout` | DELETE | Invalida la sesión |
| `/api/auth/refresh` | POST | Renueva el JWT |
| `/api/rag/generate` | POST | Proxy al RAG server — maneja SSE streaming (complejidad 17) |
| `/api/rag/collections` | GET/POST | CRUD de colecciones en Milvus |
| `/api/upload` | POST | Sube archivos y los encola para ingesta |
| `/api/admin/ingestion/stream` | GET | SSE en tiempo real del estado de jobs |
| `/api/admin/analytics` | GET | Estadísticas del sistema |
| `/api/admin/users` | GET/POST | CRUD de usuarios |
| Memoria (settings) | Server Actions | Memoria personalizada — no hay `/api/memory` dedicado |
| `/api/notifications` | GET | Lista de notificaciones recientes |
| `/api/notifications/stream` | GET | SSE — notificaciones en tiempo real |
| `/api/health` | GET | Health check del sistema |
| `/api/slack` | POST | Bot endpoint de Slack |
| `/api/teams` | POST | Bot endpoint de Microsoft Teams |

---

## Packages compartidos

### `packages/shared`

Zod schemas y tipos compartidos entre `apps/web` y `apps/cli`.

Archivo crítico: `packages/shared/src/schemas.ts`
- `RagParamsSchema` — parámetros del LLM (temperature, top_p, etc.)
- Tipos de roles: `"admin" | "area_manager" | "user"`
- Focus modes: `"detailed" | "executive" | "technical" | "comparative"`

**Regla:** si necesitás un tipo que existe en web Y en cli, va en `packages/shared`.

### `packages/db`

17 archivos de queries en `packages/db/src/queries/`. Cada uno tiene su test en `__tests__/`.

Archivo más complejo: `users.ts` — incluye verificación de password (bcrypt), manejo de áreas, permisos.

### `packages/logger`

La función `formatPretty` tiene complejidad ciclomática 29 — la función más compleja del proyecto. Formatear logs con colores, timestamps, y niveles.

---

## CI/CD — GitHub Actions

En `.github/workflows/`:

| Workflow | Trigger | Descripción |
|---|---|---|
| `ci.yml` | PRs + push a dev | Tests, type-check, lint, coverage, component tests, visual regression, a11y |
| `deploy.yml` | Push a main | Deploy a workstation física (stack Python) |
| `release.yml` | Tags semver | Release automation |

---

## Componentes sin tests de componente

Los siguientes componentes NO tienen tests de componente aún (pendiente para iteraciones futuras):

**Chat** (complejos — usar E2E de Maestro):
- `ChatInterface` (complejidad 22 — el más complejo de la UI)
- `AnnotationPopover`, `ArtifactsPanel`, `ChatDropZone`
- `CollectionSelector`, `DocPreviewPanel`, `ExportSession`
- `FocusModeSelector`, `PromptTemplates`, `RelatedQuestions`
- `ShareDialog`, `SourcesPanel`, `ThinkingSteps`, `VoiceInput`

**Layout**:
- `AppShell`, `AppShellChrome`, `CommandPalette`, `NavRail`
- `SecondaryPanel`, `WhatsNewPanel`
- `panels/AdminPanel`, `panels/ChatPanel`, `panels/ProjectsPanel`

**Collections**: `DocumentGraph` (y vistas relacionadas en `collections/`)

**Onboarding**: `OnboardingTour`

**Admin restantes**: `AnalyticsDashboard`, `IngestionKanban`, `KnowledgeGapsClient`, `ReportsAdmin`, `WebhooksAdmin`, `IntegrationsAdmin`, `ExternalSourcesAdmin`

---

## Planes de implementación

| Plan | Tema | Estado |
|---|---|---|
| Plan 1 | Monorepo TS — birth | ✅ completado 2026-03-24 |
| Plan 2 | Testing sistemático | ✅ completado 2026-03-25 |
| Plan 3 | Bugfix CodeGraphContext | ✅ completado 2026-03-25 |
| Plan 4 | Product roadmap (features) | ✅ completado 2026-03-25 |
| Plan 5 | Testing foundation (95% cobertura) | ✅ completado 2026-03-26 |
| Plan 6 | UI Testing Suite completa | ✅ completado 2026-03-26 |
| Plan 7 | Design System "Warm Intelligence" | ✅ completado 2026-03-26 |
| Plan 8 | Optimización (Next 16, Redis, BullMQ, ESLint, etc.) | ✅ completado 2026-03-27 |
| Plan 9 | Repo limpio (dead code, husky, commitlint) | ✅ completado 2026-03-27 |
| Plan 10 | Testing (visual, a11y, E2E Playwright, cobertura) | ✅ completado 2026-03-27 |
| Plan 11 | Documentación (README, API, packages, JSDoc) | ✅ completado 2026-03-27 |

Planes futuros pendientes:
- **E2E Maestro** (opcional) — requiere Java 17 + Maestro CLI si se adopta además de Playwright
- **Features RAG/GPU** — ver README (versiones futuras)

### Serie 1.0.x — UI Foundation "Claude-style"

Ver roadmap completo: `docs/roadmap-1.0.x.md`
ADR de estrategia: `docs/decisions/011-ui-series-1.0.x-strategy.md`

| Release | Plan | Scope | Estado |
|---|---|---|---|
| 1.0.1 | Plan 13 | Tokens + Font + NavRail + Login + Chat | En curso |
| 1.0.2 | Plan 14 | /collections, /upload, /extract, /saved, /projects, componentes UI | Pendiente |
| 1.0.3 | Plan 15 | 12 páginas admin + /audit + /settings | Pendiente |
| 1.0.4 | Plan 16 | Dark mode rehecho con tokens oficiales Claude | Pendiente |

---

## Estructura de páginas (24 rutas)

```
/login                  — pública, sin NavRail
/chat                   — lista de sesiones + empty state
/chat/[id]              — conversación con mensajes y input
/collections            — lista de colecciones con tabla
/collections/[name]/graph — grafo de documentos
/upload                 — subida de documentos con drop zone
/extract                — wizard de extracción estructurada
/saved                  — respuestas guardadas
/projects               — proyectos y sus sesiones
/projects/[id]          — detalle de proyecto
/settings               — perfil, contraseña, preferencias
/settings/memory        — memoria del sistema (RAG personalizado)
/audit                  — tabla de eventos con filtros
/admin/users            — CRUD de usuarios con tabla avanzada
/admin/areas            — CRUD de áreas
/admin/permissions      — matriz de permisos colección × área
/admin/rag-config       — sliders de parámetros del LLM
/admin/system           — status del sistema + jobs activos
/admin/ingestion        — kanban de jobs de ingesta (SSE)
/admin/analytics        — dashboard con gráficos recharts
/admin/knowledge-gaps   — brechas de conocimiento detectadas
/admin/reports          — informes programados
/admin/webhooks         — webhooks HTTP para eventos
/admin/integrations     — Slack y Teams bot setup
/admin/external-sources — fuentes externas (GDrive, SharePoint)
/share/[token]          — sesión compartida (pública, solo lectura)
```

---

## Componentes por categoría

```
src/components/
  ui/           — primitivos: button, badge, input, textarea, avatar, table,
                  separator, tooltip, dialog, sheet, command, sonner,
                  theme-toggle, skeleton, stat-card, empty-placeholder, data-table
  layout/       — AppShell, AppShellChrome, NavRail, SecondaryPanel,
                  CommandPalette, WhatsNewPanel, panels/
  chat/         — ChatInterface, SessionList, SourcesPanel, ArtifactsPanel,
                  CollectionSelector, FocusModeSelector, VoiceInput,
                  ExportSession, ShareDialog, ThinkingSteps, RelatedQuestions,
                  AnnotationPopover, PromptTemplates, ChatDropZone
  admin/        — UsersAdmin, AreasAdmin, PermissionsAdmin, RagConfigAdmin,
                  SystemStatus, IngestionKanban, AnalyticsDashboard,
                  KnowledgeGapsClient, ReportsAdmin, WebhooksAdmin,
                  IntegrationsAdmin, ExternalSourcesAdmin
  collections/  — CollectionsList, DocumentGraph
  settings/     — SettingsClient
  upload/       — UploadClient
  extract/      — ExtractionWizard
  audit/        — AuditTable
  projects/     — ProjectsClient
  onboarding/   — OnboardingTour
  dev/          — ReactScan, ReactScanProvider (solo dev)
  providers.tsx — ThemeProvider (next-themes)
```

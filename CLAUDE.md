# RAG Saldivia — Contexto de proyecto

## Qué es este proyecto

Overlay sobre **NVIDIA RAG Blueprint v2.5.0** que agrega autenticación, RBAC, multi-colección, frontend Next.js 15, CLI TypeScript, design system completo, y suite de testing UI.

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
| Framework web | Next.js App Router | 15.x |
| Base de datos | SQLite vía Drizzle ORM + @libsql/client | — |
| Auth | JWT (jose) en cookie HttpOnly | — |
| Validación | Zod (compartido entre todos los paquetes) | 3.x |
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
bun run test                    # todos los tests (lógica) via Turborepo
bun test apps/web/src/lib/      # solo lib/ de web (68 tests)
bun test packages/db/           # solo DB queries (161 tests)
bun run test:components         # component tests con happy-dom (147 tests)
bun run test:visual             # visual regression Playwright (22 tests)
bun run visual:update           # regenerar baseline de screenshots
bun run test:a11y               # auditoría WCAG AA con axe-playwright
bun run test:ui                 # test:components + test:visual

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

### Suite actual (215+ tests en verde)

| Capa | Comando | Tests | Cobertura |
|---|---|---|---|
| Lógica pura (lib/, db, config, logger) | `bun run test` | ~270 | ≥95% |
| Componentes React | `bun run test:components` | 147 | 20 componentes |
| Visual regression | `bun run test:visual` | 22 | 11 stories × 2 temas |
| A11y WCAG AA | `bun run test:a11y` | 5 páginas | login, chat, collections, admin, settings |

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
| `apps/web/src/middleware.ts` | JWT + RBAC en cada request |
| `apps/web/src/lib/auth/jwt.ts` | createJwt, verifyJwt, cookies |
| `apps/web/src/app/globals.css` | tokens CSS del design system — cambios afectan toda la app |
| `apps/web/src/lib/component-test-setup.ts` | setup de happy-dom — modificar puede romper 147 tests |
| `packages/db/src/schema.ts` | schema completo de 12 tablas SQLite |
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

### Next.js y React
- **Server Components por defecto** — cero JS al browser salvo donde sea necesario
- **`"use client"`** solo en componentes que necesitan estado, efectos, o APIs de browser
- **`next/font/google`** para Instrument Sans — hace self-hosting automático y optimiza preload
- **`next/dynamic` con `ssr: false`** solo puede usarse en Client Components, no Server Components
- **Cache con `unstable_cache`** — cachear llamadas al RAG con `tags: ['collections']`

### Base de datos
- **Temporal API** para todos los timestamps — `Date.now()` en lugar del bug `_ts()` de SQLite
- **SQLite locking** — `ingestion_queue` usa `locked_at` para locking optimista sin Redis
- **CJS sobre ESM** — paquetes `packages/*` sin `"type": "module"` para compatibilidad con webpack

### Logger
- **`@rag-saldivia/db` importado estáticamente** en `packages/logger` — import dinámico fallaba silenciosamente en webpack/Next.js

### SSE / Streaming
- **Verificar status HTTP ANTES de streamear** — el gateway antiguo siempre retornaba 200

### Tailwind v4 específico
- Requiere `postcss.config.js` con `@tailwindcss/postcss` — sin él, las utility classes custom no se generan
- Las nuevas clases custom (como `bg-surface`) solo se generan si aparecen en archivos escaneados por Tailwind

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

Planes futuros pendientes:
- **Plan 8** — Dependency Upgrades (Next.js 16, Zod 4, Lucide 1.x, Drizzle 0.45)
- **Plan 9** — E2E Maestro (requiere instalar Java 17 + Maestro CLI)

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
                  AnnotationPopover, PromptTemplates, ChatDropZone, SplitView
  admin/        — UsersAdmin, AreasAdmin, PermissionsAdmin, RagConfigAdmin,
                  SystemStatus, IngestionKanban, AnalyticsDashboard,
                  KnowledgeGapsClient, ReportsAdmin, WebhooksAdmin,
                  IntegrationsAdmin, ExternalSourcesAdmin
  collections/  — CollectionsList, DocumentGraph, CollectionHistory
  settings/     — SettingsClient
  upload/       — UploadClient
  extract/      — ExtractionWizard
  audit/        — AuditTable
  projects/     — ProjectsClient
  auth/         — SSOButton
  onboarding/   — OnboardingTour
  dev/          — ReactScan, ReactScanProvider (solo dev)
  providers.tsx — ThemeProvider (next-themes)
```

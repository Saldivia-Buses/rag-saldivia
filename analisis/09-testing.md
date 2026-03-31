# 09 — Testing

## Suite actual: ~380 tests

| Capa | Comando | Tests | Que testea |
|------|---------|-------|-----------|
| Logica pura + actions + API + proxy | `bun run test` | ~198 | lib/, packages/db, packages/config, packages/logger |
| Componentes + hooks | `bun run test:components` | ~158 | Componentes React con happy-dom |
| Visual regression | `bun run test:visual` | 22 | Screenshots de Storybook (11 stories x 2 temas) |
| A11y WCAG AA | `bun run test:a11y` | paginas clave | Audita con axe-playwright |
| E2E | `bun run test:e2e` | flujos criticos | Playwright browser tests |

---

## Unit tests (`bun run test`)

**Corridos via Turborepo** — ejecuta tests de todos los packages en paralelo.

### packages/db tests (~198 tests en 19 archivos)
Cada modulo de queries tiene su archivo de test:
- `users.test.ts` — CRUD, password, roles, areas
- `areas.test.ts` — CRUD, colecciones, miembros
- `sessions.test.ts` — CRUD, mensajes, fork
- `tags.test.ts` — CRUD tags en sesiones
- `saved.test.ts` — toggle saved, list
- `annotations.test.ts` — CRUD anotaciones
- `templates.test.ts` — CRUD prompt templates
- `projects.test.ts` — CRUD proyectos, sesiones
- `external-sources.test.ts` — CRUD fuentes externas
- `collection-history.test.ts` — Log de cambios
- `reports.test.ts` — CRUD reportes
- `rate-limits.test.ts` — Rate limiting
- `webhooks.test.ts` — CRUD webhooks
- `search.test.ts` — Busqueda full-text, across collections
- `shares.test.ts` — Share tokens
- `memory.test.ts` — User memory
- `messaging.test.ts` — Mensajes, reacciones, mentions
- `channels.test.ts` — Canales, miembros
- `events.test.ts` — Audit log
- `redis.test.ts` — Redis client

**Patron ADR-007:** Funciones reales contra DB en memoria, no mocks. Los tests crean la DB, seedean, y ejecutan queries reales.

### apps/web/src/lib tests
- `utils.test.ts` — Utilidades generales
- `changelog.test.ts` — Generacion de changelog
- `webhook.test.ts` — Verificacion de firmas
- `export.test.ts` — Export de sesiones
- `jwt.test.ts` — Creacion y verificacion JWT

### apps/web/src/__tests__/ (integracion)
Tests de API routes, server actions, proxy middleware.

### packages/config tests
- `config.test.ts` — Carga de configuracion

### packages/logger tests
- `logger.test.ts` — Niveles, formato, rotacion

---

## Component tests (`bun run test:components`)

**Runtime:** happy-dom (via GlobalRegistrator)
**Framework:** @testing-library/react
**Preload:** `--preload ./src/lib/component-test-setup.ts`

### Convenciones OBLIGATORIAS

```typescript
import { afterEach } from "bun:test"
import { cleanup, render, fireEvent } from "@testing-library/react"

// 1. cleanup OBLIGATORIO en cada archivo
afterEach(cleanup)

// 2. Queries escopadas al render, NUNCA screen global
const { getByRole, getByText } = render(<Button>Click</Button>)
// CORRECTO: getByRole("button")
// INCORRECTO: screen.getByRole("button")

// 3. fireEvent sobre userEvent (happy-dom compatibility)
fireEvent.click(getByRole("button"))
// NO: await userEvent.click(...)
```

### Archivos de test (18)

| Archivo | Componentes |
|---------|------------|
| `button.test.tsx` | Button (variantes, sizes, disabled, asChild) |
| `badge.test.tsx` | Badge (variantes) |
| `input.test.tsx` | Input (tipos, placeholder, disabled) |
| `textarea.test.tsx` | Textarea (rows, resize) |
| `avatar.test.tsx` | Avatar (imagen, fallback) |
| `table.test.tsx` | Table (estructura, responsive) |
| `confirm-dialog.test.tsx` | ConfirmDialog (open, confirm, cancel) |
| `data-table.test.tsx` | DataTable (sorting, filter, pagination) |
| `empty-placeholder.test.tsx` | EmptyPlaceholder (icono, titulo) |
| `separator.test.tsx` | Separator (horizontal, vertical) |
| `skeleton.test.tsx` | Skeleton, SkeletonText, SkeletonTable |
| `stat-card.test.tsx` | StatCard (delta, trend) |
| `theme-toggle.test.tsx` | ThemeToggle (switch) |
| `error-boundary.test.tsx` | ErrorBoundary (catch, recovery) |
| `admin-*.test.tsx` | Admin components |
| `chat-*.test.tsx` | Chat components |
| `collections.test.tsx` | CollectionsList |
| `settings.test.tsx` | SettingsClient |

---

## Visual regression (`bun run test:visual`)

**Framework:** Playwright
**Config:** `apps/web/playwright.config.ts`
**Baselines:** Screenshots committeados

### Proceso
1. Storybook debe estar corriendo en `:6006`
2. Playwright abre cada story
3. Screenshot en light mode + dark mode
4. Compara contra baseline (pixel diff)
5. Si un cambio de diseno es intencional → `bun run visual:update`

### Cobertura
11 stories x 2 temas = 22 tests visuales

---

## A11y (`bun run test:a11y`)

**Framework:** axe-playwright
**Standard:** WCAG AA

Testea paginas clave:
- Login
- Chat
- Collections
- Admin
- Settings

---

## E2E (`bun run test:e2e`)

**Framework:** Playwright
**Flujos:** Login, crear sesion, enviar query, ver respuesta, logout

---

## Preloads de test

### `src/lib/test-setup.ts`
Cargado para TODOS los tests. Mockea:
- `next/navigation` (useRouter, usePathname, etc.)
- `next/font/google` (Instrument Sans)
- `next-themes` (useTheme)

### `src/lib/component-test-setup.ts`
Cargado para component tests. Agrega:
- `GlobalRegistrator` de happy-dom
- `@testing-library/react` setup

### `packages/db/bunfig.toml`
Preload de `ioredis-mock` para tests de DB que usan Redis.

---

## Quality gates (despues de CADA fase)

```bash
bunx tsc --noEmit          # tipos — 0 errores
bun run test               # unit tests — todos pasan
bun run lint               # lint — limpio
```

Despues de CADA plan:
```bash
bun run test:components    # component tests
bun run test:visual        # visual regression (si cambios UI)
bun run test:a11y          # accesibilidad (si cambios UI)
```

Antes de release:
```bash
bun run test:e2e           # E2E
# + security audit (agent security-auditor)
```

---

## Componentes SIN tests (pendientes)

**Chat avanzado (usar E2E):**
ChatInterface (complejidad 22), ArtifactPanel, ChatLayout, MarkdownMessage

**Messaging (Plan 25, nuevos):**
Todos los 19 componentes de messaging

**Admin:**
AdminDashboard, AdminPermissions, PermissionMatrix

**Layout:**
AppShell, AppShellChrome, NavRail

**Nota:** 3 de 7 hooks SI tienen tests (useAutoResize, useCopyToClipboard, useLocalStorage). Los 4 restantes no: useMessaging, usePresence, useTyping, useGlobalHotkeys.

---

## Gaps de cobertura — analisis profundo

### Cobertura real por capa

```
                          Con test    Sin test    Cobertura
DB queries (21 modulos)   17          4           81%
Componentes (68 activos)  24          44          35%
Hooks (7)                 3           4           43%
Lib files (~23)           7           ~16         30%
API routes (18)           2           16          11%
Shared schemas (4)        1           3           25%
TOTAL                     54          ~87         ~38%
```

### DB queries SIN tests (4 de 21)

| Query module | Lineas | Prioridad |
|-------------|--------|-----------|
| **`rbac.ts`** | 302 | **CRITICA** — permission resolution, logica compleja |
| **`channels.ts`** | 255 | **ALTA** — CRUD canales (Plan 25) |
| **`messaging.ts`** | 242 | **ALTA** — messages + reactions (Plan 25) |
| `events-cleanup.ts` | 24 | BAJA — modulo minimo |

### Componentes CRITICOS sin test (>500 lineas)

| Componente | Lineas | Imports | Riesgo |
|-----------|--------|---------|--------|
| `ChatInterface.tsx` | 643 | 21 | Mas complejo, 94 condicionales, 0 tests |
| `AdminRoles.tsx` | 626 | 6 | CRUD roles + permisos |
| `AdminUsers.tsx` | 592 | 6 | CRUD usuarios |
| `ArtifactPanel.tsx` | 541 | 5 | Renderiza HTML/SVG con DOMPurify |

### Hooks — 43% cobertura (3 de 7)

**CON tests:** useAutoResize, useCopyToClipboard, useLocalStorage

**SIN tests:**

| Hook | Prioridad de test |
|------|-------------------|
| `useMessaging.ts` | **Alta** — API integration |
| `useGlobalHotkeys.ts` | Media — keyboard events |
| `useTyping.ts` | Media — timer logic |
| `usePresence.ts` | Baja — posiblemente dead code |

### Lib files CRITICOS sin test

| Archivo | Prioridad |
|---------|-----------|
| `rag/ai-stream.ts` | **Alta** — adapter critico NVIDIA → AI SDK |
| `rag/client.ts` | **Alta** — mock mode + error handling |
| `rag/artifact-parser.ts` | **Alta** — parser de artifacts |
| `auth/current-user.ts` | Media — 43 archivos dependen de el |
| `auth/rbac.ts` | Media — logica RBAC |

### Plan de accion priorizado

1. **Sprint 1 (criticos):** tests para `rbac.ts`, `channels.ts`, `messaging.ts`, `ChatInterface.tsx`
2. **Sprint 2 (admin):** tests para `AdminRoles.tsx`, `AdminUsers.tsx`, `ArtifactPanel.tsx`
3. **Sprint 3 (lib):** tests para `ai-stream.ts`, `client.ts`, `artifact-parser.ts`
4. **Sprint 4 (auth routes):** tests para `/api/auth/login`, `/api/auth/logout`, `/api/auth/refresh`
5. **Sprint 5 (messaging):** tests para ChannelView, MessageList, MessageComposer, MessageItem

---

## Accesibilidad — estado actual

### Lo que hay

**Herramientas integradas:**
- `axe-playwright` para auditorias WCAG AA automatizadas (`bun run test:a11y`)
- `@storybook/addon-a11y` — WCAG por componente en Storybook
- Componentes Radix (Dialog, Tooltip, Popover, etc.) son accesibles out-of-the-box

**52 usos de atributos a11y** (aria-*, role=, tabIndex, onKeyDown, sr-only) en 21 archivos de componentes.

**Componentes con mejor a11y:**
- `NavRail.tsx` — 7 atributos (roles, aria-labels, keyboard nav)
- `SessionList.tsx` — 6 atributos
- `ChatInterface.tsx` — 5 atributos
- `ChatInputBar.tsx` — 3 atributos (onKeyDown para Enter/Shift+Enter)
- `AdminRoles.tsx` — 4 atributos

### Lo que falta

| Gap | Impacto | Esfuerzo |
|-----|---------|----------|
| Focus management en navegacion SPA | Tab order se pierde al navegar | Medio |
| Skip-to-content link | No hay forma de saltar el NavRail con teclado | Bajo |
| aria-live regions para streaming | Screen readers no anuncian tokens del RAG en tiempo real | Alto |
| Labels en iconos sin texto | Algunos botones icon-only sin aria-label | Bajo |
| Contraste en dark mode | No verificado sistematicamente (Plan 20 pendiente) | Medio |
| Keyboard nav en messaging | MentionSuggestions, ReactionPicker sin keyboard support | Medio |

### Test:a11y no cubre messaging

Las paginas testeadas con axe-playwright son: login, chat, collections, admin, settings. **Messaging (Plan 25) no tiene tests a11y.**

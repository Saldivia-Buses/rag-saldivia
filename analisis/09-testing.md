# 09 — Testing

## Suite actual: ~1,059 tests (actualizado post Plans 27-29)

| Capa | Comando | Tests | Que testea |
|------|---------|-------|-----------|
| Logica pura + actions + API + proxy | `bun run test` | ~693 | lib/, packages/db, packages/config, packages/logger, auth routes |
| Componentes + hooks | `bun run test:components` | ~314 | Componentes React con happy-dom (47/47 componentes) |
| Visual regression | `bun run test:visual` | 22 | Screenshots de Storybook (11 stories x 2 temas) |
| A11y WCAG AA | `bun run test:a11y` | 8 paginas | Audita con axe-playwright (incl. messaging, admin) |
| E2E | `bun run test:e2e` | ~22 | Playwright: auth, chat, admin access control |

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

## Cobertura post Plans 27-29 (actualizado 2026-04-01)

### Cobertura real por capa

```
                          Con test    Sin test    Cobertura    Cambio
DB queries (21 modulos)   ~20         1           ~95%         81% → 95%
Componentes (75 activos)  47          28          ~63%         35% → 63%
Hooks (7)                 6           1           86%          43% → 86%
Lib files (~25)           ~12         ~13         ~48%         30% → 48%
API routes (19)           ~6          ~13         ~32%         11% → 32%
Shared schemas (4)        1           3           25%          sin cambio
TOTAL                     ~92         ~59         ~61%         38% → 61%
```

### Gaps cerrados por Plans 27-29

| Lo que faltaba | Lo que se hizo | Plan |
|---------------|---------------|------|
| `rbac.ts` sin tests (CRITICO) | 25-30 tests de RBAC | Plan 27 |
| `channels.ts` sin tests | 15-20 tests de channels + messaging | Plan 27 |
| `messaging.ts` sin tests | Incluido en Plan 27 | Plan 27 |
| `ai-stream.ts` sin tests | 8-10 tests de SSE parsing + citations | Plan 27 |
| `rag/client.ts` sin tests | 8-10 tests mock mode + errors | Plan 27 |
| Auth routes sin tests | 12-15 tests login/refresh/logout | Plan 27 |
| ChatInterface sin tests | Tests del componente descompuesto | Plan 29 |
| AdminRoles/AdminUsers sin tests | Tests de componentes + sub-componentes | Plan 29 |
| Messaging 0/19 componentes | 19/19 componentes con tests (79 tests) | Plan 29 |
| Admin 3/11 componentes | 11/11 componentes con tests (36 tests) | Plan 29 |
| Hooks 3/7 | 6/7 (useMessaging, useGlobalHotkeys, useTyping agregados) | Plan 29 |
| E2E limitado | Auth flows, chat, admin access control | Plan 29 |
| A11y 4 paginas | 8 paginas (+ messaging, admin) | Plan 29 |

### Gaps remanentes

| Area | Sin test | Prioridad |
|------|---------|-----------|
| `ArtifactPanel.tsx` | 541 lineas, DOMPurify + parser | Media |
| `MarkdownMessage.tsx` | 308 lineas, Shiki + markdown | Media |
| `usePresence.ts` | Posible dead code | Baja |
| `events-cleanup.ts` | 24 lineas, modulo minimo | Baja |
| ~13 API routes | La mayoria son CRUD simple | Baja |
| 3 shared schemas | core, rag, messaging | Baja |

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

# Testing — Guía completa

> Plans 5 y 6 completados 2026-03-26

---

## Resumen de la suite

| Capa | Herramienta | Comando | Tests | Estado |
|---|---|---|---|---|
| Lógica pura (lib, db, config, logger) | bun:test | `bun run test` | ~270 | ✅ verde |
| Componentes React | @testing-library/react + happy-dom | `bun run test:components` | 147 | ✅ verde |
| Visual regression | Playwright screenshots | `bun run test:visual` | 22 | ✅ verde |
| A11y WCAG AA | axe-playwright | `bun run test:a11y` | 5 páginas | config lista |
| E2E flujos de usuario | Maestro | `maestro test tests/e2e/` | 7 flows | pendiente Java |
| Performance/re-renders | react-scan | `bun run dev` | baseline | ver docs/superpowers/ |

---

## Capa 1 — Tests de lógica pura

### Ubicación

```
packages/db/src/__tests__/          → queries de DB (161 tests, 14 archivos)
packages/logger/src/__tests__/      → logger + blackbox
packages/config/src/__tests__/      → config loader
apps/web/src/lib/__tests__/         → webhook, changelog, export
apps/web/src/lib/auth/__tests__/    → JWT, RBAC
apps/web/src/lib/rag/__tests__/     → detect-artifact, detect-language
```

### Correr

```bash
bun run test                    # todos (Turborepo, en paralelo)
bun test packages/db/           # solo DB
bun test apps/web/src/lib/      # solo lib/ de web
bun run test:coverage           # con reporte de cobertura (threshold 95%)
```

### Patrón de test de DB

```typescript
import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import * as schema from "../schema"

// CRÍTICO: debe ir ANTES de imports que usen getDb()
process.env["DATABASE_PATH"] = ":memory:"

const client = createClient({ url: ":memory:" })
const testDb = drizzle(client, { schema })

beforeAll(async () => {
  await client.executeMultiple(`
    CREATE TABLE IF NOT EXISTS users (...);
    CREATE TABLE IF NOT EXISTS sessions (...);
  `)
})

afterEach(async () => {
  await client.executeMultiple(`DELETE FROM sessions; DELETE FROM users;`)
})

test("createSession guarda la sesión correctamente", async () => {
  // ...
})
```

---

## Capa 2 — Tests de componentes React

### Setup

El setup usa dos archivos de preload:

**`src/lib/test-setup.ts`** — se carga para TODOS los tests (via `bunfig.toml`):
- Extiende expect con matchers de jest-dom
- Mockea: `next/navigation`, `next/font/google`, `next-themes`, `next/dynamic`

**`src/lib/component-test-setup.ts`** — solo para component tests:
- Activa `GlobalRegistrator` (happy-dom)
- Incluye todos los mocks del test-setup.ts
- **Requiere `afterEach(cleanup)` en cada archivo** para aislar el DOM

### Correr

```bash
# Con el preload de happy-dom:
bun run test:components

# Equivale a:
bun test --preload ./src/lib/component-test-setup.ts apps/web/src/components
```

### Reglas críticas

#### 1. `afterEach(cleanup)` en cada archivo

```typescript
import { afterEach } from "bun:test"
import { cleanup } from "@testing-library/react"

afterEach(cleanup)  // ← OBLIGATORIO. Sin esto, los renders se acumulan en el DOM
                    // y getByRole/getByText encuentran elementos de otros tests
```

#### 2. Usar queries escopadas, NO `screen` global

```typescript
// ✅ CORRECTO — busca solo dentro de este render
const { getByRole, getByText } = render(<Button>Click</Button>)
fireEvent.click(getByRole("button"))

// ❌ INCORRECTO — screen busca en TODO el documento (otros tests acumulados)
screen.getByRole("button")  // puede fallar por ambigüedad
```

#### 3. `fireEvent` en lugar de `userEvent`

```typescript
// ✅ happy-dom es compatible con fireEvent
fireEvent.click(getByRole("button"))
fireEvent.change(getByPlaceholderText("email"), { target: { value: "test@test.com" } })

// ❌ userEvent.click() tiene incompatibilidades con happy-dom en multi-file runs
```

#### 4. Mocks de módulos

Para Server Actions, mockear en el mismo archivo de test:

```typescript
mock.module("@/app/actions/users", () => ({
  actionCreateUser: mock(() => Promise.resolve()),
  actionDeleteUser: mock(() => Promise.resolve()),
}))
```

### Estructura de un test de componente completo

```typescript
import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup, fireEvent } from "@testing-library/react"
import { UsersAdmin } from "@/components/admin/UsersAdmin"

afterEach(cleanup)  // siempre primero

// Mock de server actions (si el componente los usa)
mock.module("@/app/actions/users", () => ({
  actionCreateUser: mock(() => Promise.resolve()),
  actionDeleteUser: mock(() => Promise.resolve()),
  actionUpdateUser: mock(() => Promise.resolve()),
}))

const mockUsers = [
  { id: 1, email: "admin@test.com", name: "Admin", role: "admin", active: true, ... },
]

describe("<UsersAdmin />", () => {
  test("renderiza la tabla con usuarios", () => {
    const { getByText } = render(<UsersAdmin users={mockUsers} areas={[]} />)
    expect(getByText("admin@test.com")).toBeInTheDocument()
  })

  test("botón Nuevo usuario presente", () => {
    const { getByRole } = render(<UsersAdmin users={mockUsers} areas={[]} />)
    expect(getByRole("button", { name: /Nuevo usuario/ })).toBeInTheDocument()
  })

  test("sin usuarios muestra EmptyPlaceholder", () => {
    const { getByText } = render(<UsersAdmin users={[]} areas={[]} />)
    expect(getByText("Sin usuarios")).toBeInTheDocument()
  })
})
```

### Cómo agregar tests para un nuevo componente

1. Crear `src/components/<categoria>/__tests__/<Nombre>.test.tsx`
2. Agregar `afterEach(cleanup)` al inicio
3. Mockear todos los server actions y fetches que use el componente
4. Cubrir como mínimo: render básico, variantes/props principales, empty state
5. Correr `bun run test:components` — debe pasar

---

## Capa 3 — Visual regression

Playwright toma screenshots de cada story de Storybook y los compara contra un baseline.

### Storybook debe estar corriendo

```bash
bun run storybook   # en :6006
```

### Comandos

```bash
bun run visual:update    # PRIMERA VEZ o cambio intencional — genera baseline
bun run test:visual      # comparar contra baseline (lo que corre en CI)
bun run visual:show      # abrir reporte HTML con diffs
```

### Flujo de trabajo

1. Implementás un cambio de diseño en un componente
2. Corrés `bun run test:visual` — el test falla (el componente cambió visualmente)
3. Revisás el diff en el reporte (`visual:show`)
4. Si el cambio es correcto: `bun run visual:update` y commitear los nuevos snapshots
5. Si fue un error: revertís el componente

### Dark mode en tests visuales

**Importante:** el `colorScheme: 'dark'` de Playwright NO activa el class-based dark mode de next-themes. Se activa vía JavaScript:

```typescript
// helpers.ts
export async function enableDarkMode(page: Page) {
  await page.evaluate(() => {
    document.documentElement.classList.add("dark")
    localStorage.setItem("theme", "dark")
  })
  await page.waitForTimeout(200) // esperar transición CSS
}
```

### Agregar una story al visual regression

```typescript
// tests/visual/design-system.spec.ts — agregar a STORIES:
{ id: "primitivos-mi-componente--default", name: "mi-componente-default" },

// El ID sigue el patrón: lowercase-con-guiones del title + "--" + nombre-del-export
// "Primitivos/MiComponente" + export Default → "primitivos-mi-componente--default"
```

---

## Capa 4 — A11y (Accesibilidad)

Auditoría WCAG 2.1 AA automática usando axe-playwright.

### Prerequisito

```bash
# App corriendo con MOCK_RAG=true
MOCK_RAG=true bun run dev
```

### Correr

```bash
bun run test:a11y
```

### Páginas auditadas

- `/login`
- `/chat`
- `/collections`
- `/admin/users`
- `/settings`

### Agregar una página al audit

```typescript
// tests/a11y/pages.spec.ts — agregar a PAGES_TO_AUDIT:
{ name: "upload", path: "/upload", requiresAuth: true },
```

### Addon a11y en Storybook

Cada story tiene un panel "Accessibility" que muestra violations en tiempo real. Ver ese panel antes de commitear un componente nuevo.

---

## Capa 5 — E2E con Maestro

Flows YAML que simulan interacciones de usuario real.

### Prerequisito — instalar Maestro

```bash
# Requiere Java 17+
sudo apt install openjdk-17-jre-headless

# Instalar Maestro
curl -fsSL "https://get.maestro.mobile.dev" | bash
```

### Correr flows

```bash
# App corriendo
MOCK_RAG=true bun run dev

# Correr todos los flows
maestro test apps/web/tests/e2e/

# Correr un flow específico
maestro test apps/web/tests/e2e/auth/login-success.yaml
```

### Flows disponibles

```
tests/e2e/
  auth/
    login-success.yaml     — login como admin
    login-invalid.yaml     — credenciales incorrectas muestran error
    logout.yaml            — logout correcto
  chat/
    new-session.yaml       — crear nueva sesión
    send-message.yaml      — enviar mensaje (MOCK_RAG=true)
  admin/
    list-users.yaml        — navegar a usuarios y ver tabla
  collections/
    list.yaml              — navegar a colecciones
```

---

## Capa 6 — Performance (react-scan)

react-scan detecta re-renders innecesarios en tiempo de desarrollo.

### Está activo automáticamente en dev

Cuando corrés `bun run dev`, react-scan se activa y muestra un overlay visual de qué componentes se re-renderizan.

```typescript
// apps/web/src/components/dev/ReactScanProvider.tsx
// Se carga via ReactScanProvider en layout.tsx (solo NODE_ENV=development)
```

### Reporte baseline

El baseline pre-Plan 7 está en `docs/superpowers/react-scan-baseline.md` (fuera de git, en .gitignore).

---

## CI — GitHub Actions

```yaml
# .github/workflows/ci.yml — jobs de UI testing:

component-tests:       # bun test:components (147 tests)
visual-regression:     # bun test:visual sobre Storybook compilado
accessibility:         # axe-playwright en 5 páginas
```

Los jobs de visual regression y a11y usan `bunx wait-on` para esperar que el servidor esté listo antes de correr los tests.

**Visual regression en CI:** Los snapshots NO están en git (`.gitignore`). Para el baseline inicial en CI, hay que correr `bun run visual:update` y commitear los PNGs, o configurar Chromatic.

---

## Troubleshooting común

### "document is not defined" en component tests

Estás corriendo sin el preload de happy-dom. Usar:
```bash
bun run test:components  # no: bun test src/components
```

### Tests de componentes se contaminan entre sí

Falta `afterEach(cleanup)` en el archivo de test. Agregarlo al inicio.

### `getByRole("button")` encuentra múltiples elementos

Usar queries escopadas al render en lugar de `screen` global:
```typescript
const { getByRole } = render(<Component />)  // ✅
screen.getByRole("button")  // ❌
```

### Visual regression falla en CI pero pasa local

Diferencias de antialiasing entre plataformas. El threshold `maxDiffPixels: 10` y `threshold: 0.01` deberían absorberlo. Si no, aumentar `maxDiffPixels` o regenerar el baseline en el mismo sistema que CI (Ubuntu).

### `bg-surface` no genera como clase Tailwind

Falta `postcss.config.js` en `apps/web/`. Debe tener:
```js
module.exports = { plugins: { "@tailwindcss/postcss": {} } }
```

---

## Componentes sin tests (gap conocido)

Los siguientes componentes no tienen tests de componente aún. Son complejos y se cubren mejor con E2E de Maestro:

### Chat (complejidad alta — usar Maestro E2E)
- `ChatInterface` — el más complejo (complejidad ciclomática 22). Maneja SSE streaming, artifacts, feedback, fork, rename, export.
- `AnnotationPopover` — anotaciones sobre texto del asistente
- `ArtifactsPanel` — panel de artifacts detectados (código, tablas, documentos)
- `ChatDropZone` — drag & drop de archivos en el chat
- `CollectionSelector` — selector de colección activa con auto-detección
- `DocPreviewPanel` — preview de documentos referenciados
- `ExportSession` — exportar sesión a Markdown
- `FocusModeSelector` — detallado/ejecutivo/técnico/comparativo
- `PromptTemplates` — templates de prompts predefinidos
- `RelatedQuestions` — preguntas relacionadas post-respuesta
- `ShareDialog` — compartir sesión con URL pública
- `SourcesPanel` — fuentes de la respuesta
- `SplitView` — vista dividida con el documento
- `ThinkingSteps` — animación de "pensando..."
- `VoiceInput` — entrada de voz

### Layout (mejor testeado via visual regression)
- `AppShell`, `AppShellChrome` — wrapper del layout
- `CommandPalette` — paleta de comandos (Cmd+K)
- `NavRail` — barra lateral de navegación
- `SecondaryPanel` — panel secundario (chat/admin/projects)
- `WhatsNewPanel` — panel de novedades
- `panels/AdminPanel`, `panels/ChatPanel`, `panels/ProjectsPanel`

### Collections
- `CollectionHistory` — historial de ingestas por colección
- `DocumentGraph` — grafo de documentos (D3.js)

### Onboarding
- `OnboardingTour` — tour guiado usando driver.js

### Admin (complejos por SSE/fetch)
- `AnalyticsDashboard` — gráficos Recharts (fetch en mount)
- `IngestionKanban` — SSE en tiempo real
- `KnowledgeGapsClient`, `ReportsAdmin`, `WebhooksAdmin`, `IntegrationsAdmin`, `ExternalSourcesAdmin`

### Hooks complejos (sin tests de unit)
- `useRagStream` (complejidad 19) — streaming SSE del RAG
- `useCrossdocStream` (complejidad 22) — crossdoc streaming
- `useCrossdocDecompose`, `useGlobalHotkeys`, `useNotifications`, `useZenMode`

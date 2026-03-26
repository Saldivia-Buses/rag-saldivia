---
name: rag-testing
description: Write and run tests for the RAG Saldivia TypeScript monorepo. Use when writing new tests, adding test coverage, running the test suite, or when the user says "escribir un test", "agregar tests", "correr los tests", "visual regression", "a11y", or mentions a specific package or component to test.
---

# RAG Saldivia — Testing

Reference: `docs/testing.md` (guía completa), `docs/workflows.md` (sección Testing), `docs/decisions/006-testing-strategy.md`.

## Suite de tests actual

| Capa | Comando | Tests | CI |
|------|---------|-------|----|
| Lógica pura (lib, db, config, logger) | `bun run test` | ~270 | ✅ |
| Componentes React | `bun run test:components` | 147 | ✅ |
| Visual regression | `bun run test:visual` | 22 | ✅ |
| A11y WCAG AA | `bun run test:a11y` | 5 páginas | ✅ |
| E2E flows | `maestro test tests/e2e/` | 7 flows | pendiente Java 17 |

## Regla de oro

> **Si el código no tiene test, no está terminado.**

El CI enforcea cobertura ≥ 95% en lógica pura. Para componentes React, ver patrones abajo.

---

## Capa 1 — Lógica pura (packages/*, apps/web/src/lib/)

### Comandos

```bash
bun run test                    # todos via Turborepo
bun run test:coverage           # con threshold 95% (CI en PRs)
bun test packages/db/           # solo DB queries
bun test apps/web/src/lib/      # solo lib/ de web
```

### Dónde agregar tests

| Código nuevo | Test va en |
|---|---|
| Query en `packages/db/src/queries/` | `packages/db/src/__tests__/<dominio>.test.ts` |
| Función pura en `apps/web/src/lib/` | `apps/web/src/lib/__tests__/<nombre>.test.ts` |
| Schema Zod en `packages/shared` | `packages/shared/src/__tests__/` |
| Config loader | `packages/config/src/__tests__/` |
| Server Action | Test de la query subyacente en `packages/db/src/__tests__/` |

### Patrón para packages/db

```typescript
import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import * as schema from "../schema"

// CRÍTICO: ANTES de cualquier import que use getDb()
process.env["DATABASE_PATH"] = ":memory:"

const client = createClient({ url: ":memory:" })
const testDb = drizzle(client, { schema })

beforeAll(async () => {
  await client.executeMultiple(`
    CREATE TABLE IF NOT EXISTS users (...);
    CREATE TABLE IF NOT EXISTS mi_tabla (...);
  `)
})

afterEach(async () => {
  await client.executeMultiple(`DELETE FROM mi_tabla; DELETE FROM users;`)
})

test("describe el comportamiento esperado", async () => {
  // arrange → act → assert
})
```

---

## Capa 2 — Componentes React

### Comando

```bash
# Siempre usar este comando (activa happy-dom via --preload):
bun run test:components

# Para un componente específico:
bun test --preload ./src/lib/component-test-setup.ts src/components/ui/__tests__/button.test.tsx
```

### Reglas OBLIGATORIAS

#### 1. `afterEach(cleanup)` en CADA archivo de test

```typescript
import { afterEach } from "bun:test"
import { cleanup } from "@testing-library/react"

afterEach(cleanup)  // SIN ESTO: los renders se acumulan y getByRole encuentra elementos de otros tests
```

#### 2. Queries escopadas al render (NUNCA screen global)

```typescript
// ✅ CORRECTO
const { getByRole, getByText } = render(<Button>Click</Button>)
expect(getByRole("button")).toBeInTheDocument()

// ❌ INCORRECTO — screen busca en TODO el documento acumulado
screen.getByRole("button")
```

#### 3. `fireEvent` en lugar de `userEvent`

```typescript
// ✅ compatible con happy-dom
fireEvent.click(getByRole("button"))
fireEvent.change(getByPlaceholderText("email"), { target: { value: "test@test.com" } })

// ❌ userEvent tiene incompatibilidades en multi-file runs con happy-dom
```

#### 4. Mockear server actions y módulos de Next.js

```typescript
import { mock } from "bun:test"

// Server actions
mock.module("@/app/actions/users", () => ({
  actionCreateUser: mock(() => Promise.resolve()),
  actionDeleteUser: mock(() => Promise.resolve()),
}))

// next/navigation (ya en component-test-setup.ts, pero si necesitás override):
mock.module("next/navigation", () => ({
  useRouter: () => ({ push: mock(() => {}), refresh: mock(() => {}) }),
  usePathname: () => "/",
  redirect: mock((_url: string) => { throw new Error(`NEXT_REDIRECT: ${_url}`) }),
}))
```

### Template de test de componente completo

```typescript
import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup, fireEvent } from "@testing-library/react"
import { MiComponente } from "@/components/categoria/MiComponente"

afterEach(cleanup)  // ← SIEMPRE al inicio

// Mock de server actions si el componente los usa
mock.module("@/app/actions/mi-dominio", () => ({
  actionMiAction: mock(() => Promise.resolve()),
}))

const mockDatos = [{ id: 1, nombre: "Ejemplo" }]

describe("<MiComponente />", () => {
  test("renderiza sin errores con props mínimas", () => {
    const { getByText } = render(<MiComponente datos={mockDatos} />)
    expect(getByText("Ejemplo")).toBeInTheDocument()
  })

  test("botón de acción presente", () => {
    const { getByRole } = render(<MiComponente datos={mockDatos} />)
    expect(getByRole("button", { name: /Acción/ })).toBeInTheDocument()
  })

  test("sin datos muestra EmptyPlaceholder", () => {
    const { getByText } = render(<MiComponente datos={[]} />)
    expect(getByText("Sin elementos")).toBeInTheDocument()
  })

  test("click en botón llama callback", () => {
    const onClick = mock(() => {})
    const { getByRole } = render(<MiComponente datos={mockDatos} onClick={onClick} />)
    fireEvent.click(getByRole("button"))
    expect(onClick).toHaveBeenCalledTimes(1)
  })
})
```

### Tests de componentes UI primitivos

Para componentes de `src/components/ui/` verificar:
1. Render básico con props mínimas
2. Variantes clave (default, destructive, etc.)
3. Estado disabled
4. Clase CSS de diseño (ej: `bg-primary`, `border-border`)

```typescript
test("variant destructive aplica clases correctas", () => {
  const { getByRole } = render(<Button variant="destructive">Test</Button>)
  expect(getByRole("button").className).toContain("bg-destructive")
})
```

---

## Capa 3 — Visual regression

```bash
# Primera vez / cambio intencional:
bun run visual:update   # genera baseline PNG

# Normal (compara contra baseline):
bun run test:visual

# Ver diffs cuando falla:
bun run visual:show
```

### Flujo cuando un componente cambia visualmente

1. `bun run test:visual` → falla (diff detectado)
2. `bun run visual:show` → revisar el diff HTML
3. Si el cambio es correcto: `bun run visual:update` + commitear los PNGs
4. Si fue un error: revertir el componente

### Dark mode en visual tests

```typescript
// El colorScheme: 'dark' de Playwright NO activa class-based dark mode
// Usar el helper:
import { enableDarkMode } from "./helpers"

test("button — dark mode", async ({ page }) => {
  await page.goto("/?path=/story/primitivos-button--all-variants")
  await page.waitForLoadState("networkidle")
  await enableDarkMode(page)  // agrega .dark al html + localStorage
  await expect(page).toHaveScreenshot("button-dark.png", { threshold: 0.01, maxDiffPixels: 10 })
})
```

---

## Capa 4 — A11y

```bash
# App corriendo con MOCK_RAG=true
bun run test:a11y
```

En Storybook, cada story tiene el panel "Accessibility" — verificar antes de commitear un nuevo componente.

---

## Estado de cobertura actual

### Lógica pura (~270 tests)

| Paquete | Tests | Estado |
|---|---|---|
| `packages/db` (17 archivos de queries) | ~160 | ✅ cubierto |
| `packages/logger` | 24 | ✅ cubierto |
| `packages/config` | 14 | ✅ cubierto |
| `packages/shared` | 6 | ✅ cubierto |
| `apps/web/src/lib/auth` | 17 | ✅ cubierto |
| `apps/web/src/lib/rag` | 28 | ✅ cubierto |
| `apps/web/src/lib/` (export, webhook, changelog) | ~22 | ✅ cubierto |

### Componentes React (147 tests)

| Categoría | Tests | Estado |
|---|---|---|
| `ui/` primitivos (button, badge, input, etc.) | 56 | ✅ |
| `ui/` avanzados (data-table, skeleton, stat-card, etc.) | 24 | ✅ |
| `admin/` (users, areas, permissions, rag-config, system) | 29 | ✅ |
| `chat/SessionList` | 5 | ✅ |
| `collections/`, `upload/`, `settings/`, `audit/`, `projects/`, `extract/` | 33 | ✅ |

### Componentes sin tests (gap conocido)

Ver `docs/testing.md` sección "Componentes sin tests" para la lista completa.
Los componentes de chat complejos (ChatInterface, etc.) se cubren con E2E de Maestro.

---

## Troubleshooting

**"document is not defined"** → Usá `bun run test:components`, no `bun test src/components` directamente.

**Tests se contaminan entre sí** → Falta `afterEach(cleanup)` en el archivo.

**`getByRole` encuentra múltiples elementos** → Usá queries escopadas: `const { getByRole } = render(...)`.

**Visual regression falla en CI pero pasa local** → El threshold `maxDiffPixels: 10` debería absorber diferencias de antialiasing. Si no, regenerar baseline en Ubuntu (mismo SO que CI).

**`bg-surface` no genera como clase Tailwind** → Verificar que existe `apps/web/postcss.config.js` con `@tailwindcss/postcss`.

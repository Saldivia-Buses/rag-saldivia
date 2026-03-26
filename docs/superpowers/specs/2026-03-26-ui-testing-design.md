# Plan 6 — UI Testing Suite

**Fecha:** 2026-03-26  
**Estado:** Aprobado por el usuario  
**Prereqs:** Plan 7 (Design System) — el visual regression baseline requiere el design system terminado  
**Sigue a:** 2026-03-26-design-system-design.md

> **Excepción de orden:** La Fase 1 (react-scan baseline) se ejecuta al comienzo del Plan 7, no después. Captura el estado de performance previo al rediseño. El resto del Plan 6 espera a que Plan 7 esté completo.

---

## Resumen ejecutivo

RAG Saldivia tiene cobertura de tests sólida en lógica de negocio (79 tests, 95%+ coverage en `packages/*` y `apps/web/src/lib`), pero cero tests de interfaz. Este plan construye la capa de testing UI de pies a cabeza: detección de re-renders innecesarios, tests de componentes en aislamiento, visual regression automatizada, flujos E2E con Maestro, auditoría de accesibilidad, e integración completa al CI existente. El objetivo es tener una suite sólida, expandible, y que detecte regresiones visuales y funcionales antes de que lleguen a producción.

---

## Stack de herramientas

| Herramienta | Rol | Capa |
|---|---|---|
| **react-scan** | Detección de re-renders innecesarios en tiempo real | Dev / auditoría |
| **@testing-library/react** | Tests de componentes en aislamiento | Componentes |
| **@testing-library/user-event** | Simulación de interacciones de usuario | Componentes |
| **Storybook** (del Plan 7) | Entorno de aislamiento para tests visuales | Componentes / visual |
| **@storybook/addon-a11y** | Auditoría de accesibilidad por componente | A11y |
| **Playwright** | Visual regression screenshots + E2E de páginas | Visual / E2E |
| **Maestro** | Flujos E2E en YAML legible para flujos críticos | E2E |
| **axe-core** | Auditoría a11y programática en páginas completas | A11y |

### Decisión: Playwright vs Maestro

Ambos se usan, con roles complementarios:

- **Playwright** → visual regression (screenshots de Storybook stories) y auditoría a11y de páginas completas. Integrado al CI de GitHub Actions.
- **Maestro** → flujos E2E de usuario final (login → chat → respuesta). YAML legible, fácil de mantener, no requiere conocer el DOM interno.

No se duplican flujos entre herramientas. Playwright es el runner de CI; Maestro es el lenguaje de E2E.

---

## Fase 1 — react-scan baseline (1 día)

### Objetivo
Documentar el estado actual de performance de renderizado ANTES del rediseño del Plan 7. Esto establece un baseline y detecta los problemas más graves para priorizarlos.

### Implementación

```typescript
// apps/web/src/app/layout.tsx (solo dev)
// react-scan se agrega como script de desarrollo
```

```bash
# Instalar solo en devDependencies
bun add -D react-scan
```

react-scan se activa en `NODE_ENV=development` únicamente. Un componente wrapper en el layout lo inicializa con `scan({ enabled: process.env.NODE_ENV === 'development' })`.

### Output esperado

Documento `docs/superpowers/react-scan-baseline.md` con:
- Lista de componentes con re-renders innecesarios
- Frecuencia de re-render por componente
- Prioridad de corrección (alto/medio/bajo impacto)

Los componentes críticos detectados se corrigen dentro del Plan 7 (memoización, separación de estado).

---

## Fase 2 — Component tests con @testing-library (3-4 días)

### Objetivo
Testear cada componente React en aislamiento: que renderice correctamente, que responda a props, que maneje estados de error y loading.

### Setup

```bash
bun add -D @testing-library/react @testing-library/user-event @testing-library/jest-dom happy-dom
```

Bun test tiene soporte nativo para testing de componentes con `happy-dom`. La configuración de entorno se hace en un `bunfig.toml` **local a `apps/web`** para no afectar los 79 tests existentes de `packages/*` que corren en Node:

```toml
# apps/web/bunfig.toml (nuevo, no el raíz)
[test]
environment = "happy-dom"
preload = ["./src/lib/test-setup.ts"]
```

El `bunfig.toml` raíz no se modifica.

### Estructura de tests de componentes

```
apps/web/src/components/
  ui/
    button.test.tsx
    input.test.tsx
    badge.test.tsx
    ...
  chat/
    ChatInterface.test.tsx
    SessionList.test.tsx
    ...
  admin/
    UsersAdmin.test.tsx
    AnalyticsDashboard.test.tsx
    ...
```

### Convenciones de tests

Cada test de componente cubre:
1. **Render básico**: el componente renderiza sin errores con props mínimas
2. **Props críticas**: variantes principales (variant, size, disabled, etc.)
3. **Interacciones**: clicks, inputs, form submissions con `userEvent`
4. **Estados**: loading, error, vacío
5. **Accesibilidad básica**: roles ARIA presentes, labels correctos

```typescript
// Ejemplo de estructura
describe('Button', () => {
  it('renderiza con texto', () => { ... })
  it('aplica variant destructive correctamente', () => { ... })
  it('está deshabilitado con prop disabled', () => { ... })
  it('llama onClick al hacer click', () => { ... })
  it('muestra spinner en estado loading', () => { ... })
  it('tiene role button y es accesible', () => { ... })
})
```

### Server Components vs Client Components

Next.js 15 App Router usa Server Components por defecto. `@testing-library/react` con `happy-dom` **solo puede testear Client Components** (los que tienen `"use client"`).

**Regla de clasificación:**
- `"use client"` → testeable con @testing-library/react
- Server Component (sin directiva) → testeable solo via Playwright E2E o Maestro

**Estrategia para componentes mixtos:** Si un componente feature es Server Component pero tiene lógica de UI compleja, extraer la parte interactiva a un subcomponente client. Esto ya es el patrón del proyecto (ej: `UsersAdmin` es client, la página `admin/users/page.tsx` es server).

### Prioridad de componentes a testear

**Alta (cubrir primero — todos son Client Components):**
- Todos los componentes `ui/` (12 primitivos shadcn)
- `ChatInterface`, `SessionList`
- `UsersAdmin`, `AreasAdmin`, `PermissionsAdmin`
- `LoginPage` (formulario de auth)

**Media:**
- `IngestionKanban`, `AnalyticsDashboard`
- `CollectionsList`, `DocumentGraph`
- `UploadClient`, `ExtractionWizard`

**Baja:**
- Componentes de features secundarias
- Componentes de integraciones externas

---

## Fase 3 — Visual regression con Playwright (2 días)

### Objetivo
Detectar automáticamente cualquier cambio visual no intencional en el design system. El baseline se establece una vez terminado el Plan 7 y las stories de Storybook.

### Estrategia

Visual regression se corre sobre Storybook — no sobre la app completa. Cada story del Storybook es una URL determinista que representa un componente en un estado específico.

```
http://localhost:6006/?path=/story/primitives-button--primary
http://localhost:6006/?path=/story/primitives-button--destructive
http://localhost:6006/?path=/story/layout-appshell--default
...
```

### Setup Playwright para visual regression

```typescript
// playwright.config.ts
import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './tests/visual',
  snapshotDir: './tests/visual/snapshots',
  use: {
    baseURL: 'http://localhost:6006',
  },
  projects: [
    { name: 'light' },
    // Dark mode es class-based (next-themes usa attribute="class"),
    // NO media query. colorScheme: 'dark' no funciona aquí.
    // El dark mode se activa via storageState o script en cada test.
    { name: 'dark' },
  ],
})
```

```typescript
// tests/visual/design-system.spec.ts
test('Button — light mode', async ({ page }) => {
  await page.goto('/?path=/story/primitives-button--all-variants')
  await expect(page).toHaveScreenshot('button-light.png', {
    threshold: 0.01,
    maxDiffPixels: 10,
  })
})

test('Button — dark mode', async ({ page }) => {
  // Activar dark mode class-based (next-themes usa .dark en <html>)
  await page.goto('/?path=/story/primitives-button--all-variants')
  await page.evaluate(() => {
    document.documentElement.classList.add('dark')
    localStorage.setItem('theme', 'dark')
  })
  await page.waitForTimeout(100) // esperar transición CSS
  await expect(page).toHaveScreenshot('button-dark.png', {
    threshold: 0.01,
    maxDiffPixels: 10,
  })
})
```

### Workflow de visual regression

1. **Baseline**: al terminar Plan 7, correr `playwright test --update-snapshots` para generar screenshots de referencia de todas las stories
2. **CI**: en cada PR, Playwright compara contra el baseline. Si hay diferencia → el test falla
3. **Actualización intencional**: cuando se cambia un componente a propósito, regenerar el snapshot con `--update-snapshots` y commitear

### Coverage de visual regression

| Categoría | Stories a cubrir |
|---|---|
| Primitivos | button, input, textarea, badge, avatar, dialog, table, tooltip, sheet |
| Layout | AppShell light, AppShell dark, NavRail, SecondaryPanel |
| Features | ChatInterface, IngestionKanban, AnalyticsDashboard |
| Login | Formulario + shader de fondo |
| Dark mode | Todas las stories en modo oscuro |

---

## Fase 4 — Maestro E2E flows (2-3 días)

### Objetivo
Testear flujos completos de usuario final. YAML legible y mantenible por cualquier miembro del equipo.

### Setup

```bash
curl -fsSL "https://get.maestro.mobile.dev" | bash
```

> **Nota de compatibilidad web:** Maestro es originalmente mobile-first. El soporte web se activa apuntando al URL del browser con `--host`. Los flows de web usan `tapOn`, `assertVisible` igual que mobile, pero el runner necesita el flag `--browser`. La sintaxis correcta para web **no usa `appId`** — usa un `---` con la URL base configurada en la variable de entorno `MAESTRO_DRIVER_HOST`.

### Estructura de flows

```
tests/
  e2e/
    .maestro.yaml       ← config global: host, port, env
    auth/
      login-success.yaml
      login-invalid-credentials.yaml
      logout.yaml
    chat/
      new-session.yaml
      send-message-mock.yaml
      session-history.yaml
    collections/
      list-collections.yaml
      view-collection.yaml
    admin/
      users-crud.yaml
      areas-crud.yaml
    upload/
      upload-document.yaml
```

### Flows críticos (implementar primero)

**Config global:**
```yaml
# tests/e2e/.maestro.yaml
env:
  BASE_URL: http://localhost:3000
```

**Login:**
```yaml
# tests/e2e/auth/login-success.yaml
---
- openLink: ${BASE_URL}/login
- tapOn: "Email"
- inputText: "admin@saldivia.com"
- tapOn: "Contraseña"
- inputText: "password123"
- tapOn: "Iniciar sesión"
- assertVisible: "Chat"
- assertVisible: "Colecciones"
```

**Chat completo (con MOCK_RAG=true):**
```yaml
# tests/e2e/chat/send-message-mock.yaml
---
- openLink: ${BASE_URL}/chat
- tapOn: "Nueva conversación"
- tapOn: "Escribí tu pregunta"
- inputText: "¿Cuáles son las políticas de la empresa?"
- tapOn: "Enviar"
- assertVisible: "Fuentes"
```

### Integración con MOCK_RAG

Los flows de Maestro corren con `MOCK_RAG=true` en CI para no depender del stack NVIDIA. Los flows de smoke test en staging/producción corren con RAG real.

---

## Fase 5 — Accesibilidad (1-2 días)

### Objetivo
Auditar y corregir problemas de accesibilidad en componentes y páginas. Target: WCAG 2.1 nivel AA.

### Capa 1 — Storybook addon-a11y

Ya incluido en el setup de Storybook (Plan 7). En cada story:
- Detecta violaciones de contraste automáticamente
- Verifica roles ARIA correctos
- Reporta labels faltantes en inputs
- Alerta sobre orden de foco incorrecto

### Capa 2 — Playwright + axe-core en páginas completas

```bash
bun add -D axe-playwright
```

```typescript
// tests/a11y/pages.spec.ts
import { checkA11y } from 'axe-playwright'

const pages = [
  '/login',
  '/chat',
  '/collections',
  '/admin/users',
  '/settings',
]

for (const path of pages) {
  test(`a11y: ${path}`, async ({ page }) => {
    await page.goto(path)
    await checkA11y(page, undefined, {
      detailedReport: true,
      detailedReportOptions: { html: true },
    })
  })
}
```

### Criterios de accesibilidad

- Contraste de texto: ratio ≥ 4.5:1 para texto normal, ≥ 3:1 para texto grande
- Navegación por teclado: tab order lógico en todos los flujos
- Screen reader: todos los elementos interactivos tienen labels
- Focus visible: foco siempre visible (no se elimina el outline)
- Alt text: todas las imágenes tienen alternativas textuales
- Motion: respetar `prefers-reduced-motion` en animaciones

---

## Fase 6 — CI integration (1 día)

### Objetivo
Integrar toda la suite al pipeline de CI existente en `.github/workflows/ci.yml`.

### Jobs a agregar

```yaml
# .github/workflows/ci.yml (adición)

jobs:
  # Jobs existentes: test, coverage

  component-tests:
    name: Component Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
      - run: bun install
      - run: bun test apps/web/src/components

  visual-regression:
    name: Visual Regression
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
      - run: bun install
      - run: bunx playwright install --with-deps chromium
      - run: bunx storybook build
      - name: Serve Storybook y esperar que esté listo
        run: |
          bunx serve storybook-static -p 6006 &
          bunx wait-on http://localhost:6006 --timeout 60000
      - run: bunx playwright test tests/visual/
      - uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: visual-diff
          path: tests/visual/snapshots/

  accessibility:
    name: Accessibility Audit
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
      - run: bun install
      - run: bunx playwright install --with-deps chromium
      - name: Levantar dev server y esperar que esté listo
        run: |
          bun run dev &
          bunx wait-on http://localhost:3000 --timeout 60000
        env:
          MOCK_RAG: "true"
      - run: bunx playwright test tests/a11y/

  e2e-maestro:
    name: E2E Maestro
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
      - run: bun install
      - run: curl -fsSL "https://get.maestro.mobile.dev" | bash
      - name: Levantar dev server y esperar que esté listo
        run: |
          bun run dev &
          bunx wait-on http://localhost:3000 --timeout 60000
        env:
          MOCK_RAG: "true"
      - run: maestro test tests/e2e/auth/
      - run: maestro test tests/e2e/chat/
```

### Scripts npm nuevos

```json
// apps/web/package.json
{
  "scripts": {
    "test:components": "bun test src/components",
    "test:visual": "playwright test tests/visual",
    "test:a11y": "playwright test tests/a11y",
    "test:e2e": "maestro test tests/e2e",
    "test:ui": "bun run test:components && bun run test:visual && bun run test:a11y",
    "visual:update": "playwright test tests/visual --update-snapshots"
  }
}
```

### Turbo pipeline actualizado

```json
// turbo.json (adición)
{
  "tasks": {
    "test:components": { "dependsOn": ["^build"] },
    "test:visual": { "dependsOn": ["build:storybook"] }
  }
}
```

> `test:a11y` y `test:e2e` **no se agregan al pipeline de Turbo** — dependen de un servidor de larga duración (`dev`) que Turbo no puede gestionar como tarea. Se ejecutan directamente en sus jobs de CI (ver arriba) donde el servidor se levanta y se espera con `wait-on` antes de correr los tests.

---

## Criterios de éxito

- [ ] react-scan baseline documentado antes de iniciar Plan 7
- [ ] Tests de componentes para los 62 componentes (al menos render + props críticas)
- [ ] Visual regression baseline establecido al terminar Plan 7
- [ ] 0 regresiones visuales detectadas por Playwright en CI en verde
- [ ] Maestro flows para: login, chat, upload, admin CRUD (mínimo 10 flows)
- [ ] 0 violaciones WCAG AA en componentes primitivos (via addon-a11y)
- [ ] 0 violaciones WCAG AA en páginas críticas (login, chat, admin/users)
- [ ] Todos los jobs nuevos de CI pasando en verde
- [ ] Coverage de componentes ≥ 80% (render + interacciones)
- [ ] `bun run test:ui` corre toda la suite en < 5 minutos localmente

---

## Relación con otros planes

- **Depende de Plan 7**: el baseline de visual regression requiere el design system terminado. Los component tests se escriben junto con las stories de Storybook.
- **Extiende Plan 5**: no reemplaza los tests de lógica existentes — los complementa con la capa visual.
- **Futuros planes**: cualquier nueva feature debe incluir su story de Storybook y su test de componente como parte del Definition of Done.

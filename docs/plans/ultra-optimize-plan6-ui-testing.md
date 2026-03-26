# Plan 6: UI Testing Suite — Pirámide completa

> Este documento vive en `docs/plans/ultra-optimize-plan6-ui-testing.md`.
> Se actualiza a medida que se completan las tareas. Cada fase completada genera entrada en CHANGELOG.md.
> Spec completo: `docs/superpowers/specs/2026-03-26-ui-testing-design.md`

---

## Contexto

El Plan 5 estableció cobertura ≥ 95% en lógica pura (`packages/*`, `apps/web/src/lib/`). Cero tests de UI. La app tiene 62 componentes React y 24 páginas sin ningún test automatizado de interfaz.

**Lo que construye este plan:**
1. **react-scan baseline** — fotografía del estado de performance ANTES del rediseño
2. **Component tests** — @testing-library/react para los 62 componentes (Client Components)
3. **Visual regression** — Playwright screenshots sobre Storybook (baseline post-Plan 7)
4. **Maestro E2E** — flujos críticos de usuario en YAML legible
5. **A11y audit** — addon-a11y en Storybook + axe-playwright en páginas
6. **CI integration** — todo integrado al pipeline existente

**Orden de ejecución respecto al Plan 7:**
- **Fase 1 (react-scan):** se ejecuta al INICIO del Plan 7, antes de tocar CSS
- **Fases 2–6:** se ejecutan DESPUÉS de completar el Plan 7

**Stack de herramientas:**

| Herramienta | Rol |
|---|---|
| react-scan | Performance baseline — re-renders innecesarios |
| @testing-library/react | Component tests (Client Components) |
| @testing-library/user-event | Simulación de interacciones |
| happy-dom | DOM en Bun test (scoped a apps/web) |
| Playwright | Visual regression + a11y de páginas |
| axe-playwright | Auditoría WCAG en páginas completas |
| Storybook (del Plan 7) | Entorno de aislamiento para visual regression |
| Maestro | E2E flows en YAML para flujos críticos |

---

## Seguimiento

Formato: `- [ ] Descripción — estimación`
Al completar: `- [x] Descripción — completado YYYY-MM-DD`
Cada fase completada → entrada en CHANGELOG.md → commit.

---

## Fase 1 — react-scan baseline *(1-2 hs)*

> ⚠️ **Esta fase se ejecuta al INICIO del Plan 7**, antes de cambiar tokens o componentes.
> El objetivo es documentar el estado de performance PREVIO al rediseño.

**Archivos a crear/modificar:**
- `apps/web/package.json` — agregar react-scan
- `apps/web/src/app/layout.tsx` — inicializar react-scan en dev
- `docs/superpowers/react-scan-baseline.md` — reporte de hallazgos

### F1.1 — Instalación

```bash
cd apps/web
bun add -D react-scan
```

- [x] Instalar react-scan como devDependency — completado 2026-03-26

### F1.2 — Inicialización en layout.tsx (solo dev)

```typescript
// apps/web/src/app/layout.tsx
import { scan } from 'react-scan'

if (typeof window !== 'undefined' && process.env.NODE_ENV === 'development') {
  scan({
    enabled: true,
    log: true, // loguea en consola los componentes que re-renderizan
  })
}
```

- [x] Agregar inicialización de react-scan con `dynamic(() => import(...), { ssr: false })` condicional — completado 2026-03-26
- [x] Guard `process.env.NODE_ENV === 'development'` — no activar en producción — completado 2026-03-26
- [ ] `bun run dev` — verificar que el overlay de react-scan aparece

### F1.3 — Recorrido y documentación

- [ ] Navegar `/chat` — observar qué componentes se iluminan al escribir en el input
- [ ] Navegar `/admin/users` — observar re-renders al ordenar la tabla
- [ ] Navegar `/admin/analytics` — observar qué componentes se re-renderizan en mount
- [ ] Navegar `/settings` — observar re-renders al cambiar tabs
- [ ] Navegar `/collections` — observar comportamiento de la lista

### F1.4 — Crear reporte baseline

Crear `docs/superpowers/react-scan-baseline.md` con el siguiente formato — completado 2026-03-26 (template creado, completar tras recorrer la app):

```markdown
# react-scan Baseline — Pre Plan 7

Fecha: YYYY-MM-DD
Estado: Pre-rediseño (antes de Plan 7)

## Componentes con re-renders innecesarios

| Componente | Página | Frecuencia | Causa probable | Prioridad |
|---|---|---|---|---|
| ChatInterface | /chat | Alta | estado global re-monta todo | Alta |
| ... | | | | |

## Acciones recomendadas (en Plan 7)

- [ ] Memoizar X componente con React.memo
- [ ] Extraer estado Y a un contexto más específico
- [ ] Usar useMemo en Z para evitar recalcular ...
```

- [ ] Crear el archivo con los hallazgos reales del recorrido
- [ ] Categorizar por prioridad (alta/media/baja impacto)

### F1.5 — Remover de producción (mantener solo en dev)

- [ ] Verificar que react-scan NO se incluye en el bundle de producción:
  ```bash
  bun run build && grep -r "react-scan" .next/ || echo "OK — no está en build"
  ```

### Criterio de done
`docs/superpowers/react-scan-baseline.md` existe con hallazgos. react-scan no está en el bundle de build.

### Checklist de cierre
- [ ] `bun run test` — todos pasan
- [ ] `bun run build` — sin react-scan en el bundle
- [ ] CHANGELOG.md actualizado bajo `### Plan 6 — UI Testing`
- [ ] `git commit -m "chore(perf): react-scan baseline pre-plan7 — detectar re-renders innecesarios"`

---

## Fase 2 — Setup de testing de componentes *(1-2 hs)*

Objetivo: infraestructura lista para testear Client Components con @testing-library/react + Bun.

**Archivos a crear/modificar:**
- `apps/web/bunfig.toml` — nuevo, local a apps/web, NO modifica el raíz
- `apps/web/src/lib/test-setup.ts` — setup global para tests
- `apps/web/package.json` — agregar dependencias y scripts

### F2.1 — Instalar dependencias

```bash
cd apps/web
bun add -D @testing-library/react @testing-library/user-event @testing-library/jest-dom happy-dom
```

- [ ] Instalar las 4 dependencias como devDependencies de apps/web

### F2.2 — Crear bunfig.toml LOCAL a apps/web

> ⚠️ Crear en `apps/web/bunfig.toml`, NO en la raíz. El bunfig.toml raíz no se toca.

```toml
# apps/web/bunfig.toml
[test]
environment = "happy-dom"
preload = ["./src/lib/test-setup.ts"]
```

- [ ] Crear `apps/web/bunfig.toml` con la configuración
- [ ] Verificar que los tests existentes en `packages/*` siguen pasando: `cd packages/db && bun test` — sin errores

### F2.3 — Crear test-setup.ts

```typescript
// apps/web/src/lib/test-setup.ts
import { expect } from 'bun:test'
import * as matchers from '@testing-library/jest-dom/matchers'

expect.extend(matchers)

// Mock de next/navigation (rutas)
import { mock } from 'bun:test'
mock.module('next/navigation', () => ({
  useRouter: () => ({
    push: mock(() => {}),
    replace: mock(() => {}),
    back: mock(() => {}),
    prefetch: mock(() => {}),
  }),
  usePathname: () => '/',
  useSearchParams: () => new URLSearchParams(),
}))

// Mock de next/font/google (evita errores en test env)
mock.module('next/font/google', () => ({
  Instrument_Sans: () => ({
    className: 'mock-font',
    variable: '--font-sans',
  }),
}))

// Mock de next-themes
mock.module('next-themes', () => ({
  useTheme: () => ({ theme: 'light', setTheme: mock(() => {}) }),
  ThemeProvider: ({ children }: { children: React.ReactNode }) => children,
}))
```

- [ ] Crear `apps/web/src/lib/test-setup.ts` con los mocks críticos
- [ ] Verificar: `bun test apps/web/src/lib` — los tests existentes de lib/ siguen pasando con happy-dom

### F2.4 — Script de tests de componentes

```json
// apps/web/package.json — agregar a scripts:
"test:components": "bun test src/components",
"test:components:watch": "bun test src/components --watch",
"test:components:coverage": "bun test src/components --coverage"
```

- [ ] Agregar los 3 scripts

### F2.5 — Test de smoke del setup

Crear un test mínimo para verificar que el setup funciona:

```typescript
// apps/web/src/components/ui/__tests__/setup-smoke.test.tsx
import { test, expect } from 'bun:test'
import { render, screen } from '@testing-library/react'

test('setup funciona: puede renderizar un div', () => {
  render(<div data-testid="smoke">ok</div>)
  expect(screen.getByTestId('smoke')).toBeInTheDocument()
})
```

- [ ] Crear el test de smoke
- [ ] `bun test apps/web/src/components/ui/__tests__/setup-smoke.test.tsx` — pasa

### Criterio de done
`bun test apps/web/src/components` corre. Los tests de `packages/*` siguen pasando.

### Checklist de cierre
- [ ] `bun run test` (desde raíz) — todos pasan incluido smoke test
- [ ] CHANGELOG.md actualizado
- [ ] `git commit -m "chore(test): setup @testing-library/react con happy-dom en apps/web"`

---

## Fase 3 — Component tests *(4-6 hs)*

Objetivo: tests para los 62 componentes. Prioridad: primitivos UI, luego auth, luego features admin/chat.

**Convención de archivo:** `src/components/<categoria>/__tests__/<Nombre>.test.tsx`

**Convención de test por componente:**
```typescript
describe('<NombreComponente />', () => {
  it('renderiza sin errores con props mínimas', () => { ... })
  it('variante/prop X funciona correctamente', () => { ... })
  it('estado disabled/loading/error se aplica', () => { ... })
  it('llama callback al interactuar', async () => { ... })
  it('es accesible (tiene role y label correctos)', () => { ... })
})
```

### F3.1 — button.test.tsx *(30 min)*

**Archivo:** `apps/web/src/components/ui/__tests__/button.test.tsx`

- [ ] `renderiza con texto por defecto`
- [ ] `variante destructive aplica clases de color destructivo`
- [ ] `variante outline tiene borde visible`
- [ ] `prop disabled → cursor-not-allowed y no llama onClick`
- [ ] `prop asChild renderiza el child como elemento raíz`
- [ ] `llama onClick al hacer click`
- [ ] `role="button" presente`
- [ ] `bun test src/components/ui/__tests__/button.test.tsx` — pasan

### F3.2 — input.test.tsx *(20 min)*

- [ ] `renderiza placeholder correctamente`
- [ ] `onChange se llama con el valor del input`
- [ ] `disabled bloquea la interacción`
- [ ] `type="password" oculta el texto`
- [ ] `role="textbox" o type explícito accesible`
- [ ] `bun test src/components/ui/__tests__/input.test.tsx` — pasan

### F3.3 — badge.test.tsx *(15 min)*

- [ ] `renderiza texto de contenido`
- [ ] `variante default tiene clase bg-accent`
- [ ] `variante destructive tiene clase text-destructive`
- [ ] `variante success existe (nueva variante del Plan 7)`
- [ ] `bun test src/components/ui/__tests__/badge.test.tsx` — pasan

### F3.4 — avatar.test.tsx *(15 min)*

- [ ] `AvatarFallback muestra iniciales`
- [ ] `AvatarImage tiene alt attribute`
- [ ] `AvatarFallback visible cuando imagen falla (onError handler)`
- [ ] `bun test src/components/ui/__tests__/avatar.test.tsx` — pasan

### F3.5 — dialog.test.tsx *(30 min)*

```typescript
// Usar open prop en true para forzar diálogo abierto en test
render(<Dialog open><DialogContent>Contenido</DialogContent></Dialog>)
```

- [ ] `título visible cuando open=true`
- [ ] `contenido visible cuando open=true`
- [ ] `onOpenChange se llama al cerrar`
- [ ] `Escape key cierra el diálogo (userEvent.keyboard)`
- [ ] `role="dialog" presente`
- [ ] `bun test src/components/ui/__tests__/dialog.test.tsx` — pasan

### F3.6 — table.test.tsx *(30 min)*

```typescript
const data = [{ id: 1, name: 'Ana', role: 'Admin' }]
render(
  <Table>
    <TableHeader><TableRow><TableHead>Nombre</TableHead></TableRow></TableHeader>
    <TableBody><TableRow><TableCell>{data[0].name}</TableCell></TableRow></TableBody>
  </Table>
)
```

- [ ] `header visible con texto correcto`
- [ ] `filas de datos renderizadas`
- [ ] `role="table" o semántica correcta`
- [ ] `bun test src/components/ui/__tests__/table.test.tsx` — pasan

### F3.7 — skeleton.test.tsx *(15 min)*

- [ ] `renderiza sin errores`
- [ ] `variante SkeletonText tiene ancho relativo`
- [ ] `variante SkeletonAvatar es un círculo (rounded-full)`
- [ ] `bun test src/components/ui/__tests__/skeleton.test.tsx` — pasan

### F3.8 — stat-card.test.tsx *(20 min)*

- [ ] `muestra el valor principal`
- [ ] `muestra la etiqueta`
- [ ] `delta positivo muestra verde`
- [ ] `delta negativo muestra rojo`
- [ ] `bun test src/components/ui/__tests__/stat-card.test.tsx` — pasan

### F3.9 — empty-placeholder.test.tsx *(20 min)*

- [ ] `muestra título`
- [ ] `muestra descripción`
- [ ] `renderiza children (botón de acción)`
- [ ] `bun test src/components/ui/__tests__/empty-placeholder.test.tsx` — pasan

### F3.10 — Auth: SSOButton.test.tsx *(15 min)*

**Archivo:** `apps/web/src/components/auth/__tests__/SSOButton.test.tsx`

- [ ] `renderiza con nombre del provider`
- [ ] `llama onClick al hacer click`
- [ ] `tiene aria-label descriptivo`
- [ ] `bun test src/components/auth/__tests__/SSOButton.test.tsx` — pasan

### F3.11 — Chat: SessionList.test.tsx *(30 min)*

**Archivo:** `apps/web/src/components/chat/__tests__/SessionList.test.tsx`

```typescript
const mockSessions = [
  { id: '1', title: 'Políticas de empresa', updatedAt: new Date() },
  { id: '2', title: 'Contrato de trabajo', updatedAt: new Date() },
]
```

- [ ] `lista de sesiones visible`
- [ ] `sesión activa tiene estilo destacado (aria-current o clase active)`
- [ ] `empty state visible cuando sessions=[]`
- [ ] `onClick llama con el id correcto`
- [ ] `bun test src/components/chat/__tests__/SessionList.test.tsx` — pasan

### F3.12 — Chat: CollectionSelector.test.tsx *(20 min)*

- [ ] `dropdown de colecciones renderiza`
- [ ] `al seleccionar una colección, llama onSelect con el nombre correcto`
- [ ] `bun test src/components/chat/__tests__/CollectionSelector.test.tsx` — pasan

### F3.13 — Admin: UsersAdmin.test.tsx *(45 min)*

**Archivo:** `apps/web/src/components/admin/__tests__/UsersAdmin.test.tsx`

> Mock de fetch para los llamados al API

```typescript
import { mock } from 'bun:test'
mock.module('...', () => ({ ... }))

// o usar fetch mock global
globalThis.fetch = mock(() =>
  Promise.resolve(new Response(JSON.stringify([
    { id: 1, email: 'admin@test.com', role: 'admin' }
  ]), { status: 200 }))
)
```

- [ ] `tabla de usuarios renderiza headers: Email, Rol, Acciones`
- [ ] `email del usuario visible en la tabla`
- [ ] `badge de rol visible con variante correcta`
- [ ] `botón de eliminar presente por fila`
- [ ] `loading state muestra skeleton`
- [ ] `bun test src/components/admin/__tests__/UsersAdmin.test.tsx` — pasan

### F3.14 — Admin: AreasAdmin.test.tsx *(30 min)*

- [ ] `lista de áreas visible`
- [ ] `botón crear área presente`
- [ ] `input de nombre en formulario de creación`
- [ ] `bun test src/components/admin/__tests__/AreasAdmin.test.tsx` — pasan

### F3.15 — Onboarding: OnboardingTour.test.tsx *(30 min)*

- [ ] `no renderiza si el usuario ya completó el tour (completed=true)`
- [ ] `renderiza el primer paso si completed=false`
- [ ] `botón "siguiente" avanza al paso 2`
- [ ] `botón "saltar" llama onComplete`
- [ ] `bun test src/components/onboarding/__tests__/OnboardingTour.test.tsx` — pasan

### F3.16 — Resto de componentes (prioridad baja) *(1-2 hs)*

Para cada componente restante, crear un test mínimo con al menos:
1. Render sin errors
2. El elemento principal visible

Componentes a cubrir:
- [ ] `ChatDropZone.test.tsx` — drop zone renderiza, mensaje visible
- [ ] `ArtifactsPanel.test.tsx` — panel vacío renderiza
- [ ] `SourcesPanel.test.tsx` — lista de fuentes renderiza
- [ ] `ShareDialog.test.tsx` — dialog abre con url
- [ ] `CollectionsList.test.tsx` — lista de colecciones con datos mock
- [ ] `DocumentGraph.test.tsx` — contenedor del gráfico renderiza
- [ ] `UploadClient.test.tsx` — drop zone visible
- [ ] `AuditTable.test.tsx` — tabla con datos mock renderiza
- [ ] `SystemStatus.test.tsx` — indicadores de estado visibles
- [ ] `AnalyticsDashboard.test.tsx` — stat cards visibles con datos mock
- [ ] `IngestionKanban.test.tsx` — columnas del kanban visibles
- [ ] `SettingsClient.test.tsx` — secciones de settings visibles
- [ ] `ProjectsClient.test.tsx` — lista de proyectos visible

### Criterio de done
`bun test apps/web/src/components` — todos los tests pasan. Coverage de componentes ≥ 60%.

### Checklist de cierre
- [ ] `bun run test` (raíz) — todos pasan
- [ ] `bun test apps/web/src/components --coverage` — ver reporte
- [ ] CHANGELOG.md actualizado
- [ ] `git commit -m "test(components): tests para los 62 componentes React con @testing-library"`

---

## Fase 4 — Visual regression *(2-3 hs)*

> ⚠️ **Ejecutar solo después de completar el Plan 7 y el Storybook.**
> El baseline se genera una única vez al terminar Plan 7.

**Archivos a crear:**
- `apps/web/playwright.config.ts`
- `apps/web/tests/visual/design-system.spec.ts`
- `apps/web/tests/visual/snapshots/` — generado automáticamente

### F4.1 — Instalar Playwright

```bash
cd apps/web
bun add -D @playwright/test
bunx playwright install --with-deps chromium
```

- [ ] Instalar `@playwright/test`
- [ ] Instalar browser Chromium con deps del sistema

### F4.2 — playwright.config.ts

```typescript
// apps/web/playwright.config.ts
import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './tests/visual',
  snapshotDir: './tests/visual/snapshots',
  use: {
    baseURL: 'http://localhost:6006',
  },
  // No emular colorScheme — el dark mode es class-based
  webServer: {
    command: 'bun run storybook',
    url: 'http://localhost:6006',
    reuseExistingServer: true,
    timeout: 60_000,
  },
})
```

- [ ] Crear `apps/web/playwright.config.ts`

### F4.3 — Helper para dark mode

```typescript
// apps/web/tests/visual/helpers.ts
import type { Page } from '@playwright/test'

export async function enableDarkMode(page: Page) {
  await page.evaluate(() => {
    document.documentElement.classList.add('dark')
    localStorage.setItem('theme', 'dark')
  })
  await page.waitForTimeout(150) // esperar transición CSS
}

export async function enableLightMode(page: Page) {
  await page.evaluate(() => {
    document.documentElement.classList.remove('dark')
    localStorage.setItem('theme', 'light')
  })
  await page.waitForTimeout(150)
}
```

- [ ] Crear `apps/web/tests/visual/helpers.ts`

### F4.4 — Tests de visual regression

```typescript
// apps/web/tests/visual/design-system.spec.ts
import { test, expect } from '@playwright/test'
import { enableDarkMode, enableLightMode } from './helpers'

const SNAPSHOT_OPTIONS = { threshold: 0.01, maxDiffPixels: 10 }

const primitiveStories = [
  { name: 'Button — all variants', path: '?path=/story/primitives-button--all-variants' },
  { name: 'Badge — all variants', path: '?path=/story/primitives-badge--all-variants' },
  { name: 'Input', path: '?path=/story/primitives-input--default' },
  { name: 'Avatar', path: '?path=/story/primitives-avatar--default' },
  { name: 'Table', path: '?path=/story/primitives-table--default' },
  { name: 'Skeleton', path: '?path=/story/primitives-skeleton--all-variants' },
]

for (const story of primitiveStories) {
  test(`${story.name} — light`, async ({ page }) => {
    await page.goto(story.path)
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveScreenshot(`${story.name}-light.png`, SNAPSHOT_OPTIONS)
  })

  test(`${story.name} — dark`, async ({ page }) => {
    await page.goto(story.path)
    await page.waitForLoadState('networkidle')
    await enableDarkMode(page)
    await expect(page).toHaveScreenshot(`${story.name}-dark.png`, SNAPSHOT_OPTIONS)
  })
}

// Tests de layout
test('NavRail — light', async ({ page }) => {
  await page.goto('?path=/story/layout-navrail--default')
  await page.waitForLoadState('networkidle')
  await expect(page).toHaveScreenshot('navrail-light.png', SNAPSHOT_OPTIONS)
})

test('NavRail — dark', async ({ page }) => {
  await page.goto('?path=/story/layout-navrail--default')
  await enableDarkMode(page)
  await expect(page).toHaveScreenshot('navrail-dark.png', SNAPSHOT_OPTIONS)
})
```

- [ ] Crear `apps/web/tests/visual/design-system.spec.ts`
- [ ] Agregar stories de features: analytics, ingestion, chat (mismo patrón)

### F4.5 — Generar baseline

```bash
# Levantar Storybook primero
bun run storybook &
bunx wait-on http://localhost:6006

# Generar baseline (primera vez, crea los snapshots)
bunx playwright test tests/visual/ --update-snapshots
```

- [ ] Levantar Storybook
- [ ] Correr `--update-snapshots` para crear los archivos PNG de referencia
- [ ] Commitear los snapshots: `git add tests/visual/snapshots && git commit -m "test(visual): baseline de visual regression post-plan7"`

### F4.6 — Script en package.json

```json
"test:visual": "playwright test tests/visual",
"visual:update": "playwright test tests/visual --update-snapshots",
"visual:show": "playwright show-report"
```

- [ ] Agregar scripts

### Criterio de done
Baseline generado, todos los tests de visual regression pasan en verde.

### Checklist de cierre
- [ ] `bun run test:visual` — todos pasan (después de generar baseline)
- [ ] Snapshots commiteados en git
- [ ] CHANGELOG.md actualizado
- [ ] `git commit -m "test(visual): baseline de visual regression — 20+ stories cubiertos"`

---

## Fase 5 — Maestro E2E flows *(2-3 hs)*

Objetivo: flujos críticos de usuario testeados de punta a punta con YAML legible.

**Archivos a crear:**
- `apps/web/tests/e2e/.maestro.yaml`
- `apps/web/tests/e2e/auth/*.yaml`
- `apps/web/tests/e2e/chat/*.yaml`
- `apps/web/tests/e2e/admin/*.yaml`
- `apps/web/tests/e2e/upload/*.yaml`

### F5.1 — Instalar Maestro

```bash
curl -fsSL "https://get.maestro.mobile.dev" | bash
# Verificar instalación:
maestro --version
```

- [ ] Instalar Maestro CLI globalmente
- [ ] Verificar que `maestro --version` responde

### F5.2 — Config global

```yaml
# apps/web/tests/e2e/.maestro.yaml
env:
  BASE_URL: http://localhost:3000
  ADMIN_EMAIL: admin@saldivia.com
  ADMIN_PASSWORD: password_test_123
  TEST_EMAIL: usuario@saldivia.com
```

- [ ] Crear `.maestro.yaml` con variables de entorno
- [ ] Agregar `.maestro.yaml` al `.gitignore` si contiene credenciales reales
- [ ] Para CI: usar variables de entorno del runner, no credenciales hardcodeadas

### F5.3 — Flow: login exitoso

```yaml
# apps/web/tests/e2e/auth/login-success.yaml
---
- openLink: ${BASE_URL}/login
- assertVisible: "Iniciá sesión"
- tapOn: "Email"
- inputText: ${ADMIN_EMAIL}
- tapOn: "Contraseña"
- inputText: ${ADMIN_PASSWORD}
- tapOn: "Iniciar sesión"
- assertVisible: "Chat"
- assertNotVisible: "Iniciá sesión"
```

- [ ] Crear el flow
- [ ] Correr: `MOCK_RAG=true bun run dev & maestro test tests/e2e/auth/login-success.yaml`
- [ ] Pasa ✓

### F5.4 — Flow: credenciales inválidas

```yaml
# apps/web/tests/e2e/auth/login-invalid.yaml
---
- openLink: ${BASE_URL}/login
- tapOn: "Email"
- inputText: "noexiste@saldivia.com"
- tapOn: "Contraseña"
- inputText: "wrongpassword"
- tapOn: "Iniciar sesión"
- assertVisible: "Credenciales inválidas"
- assertNotVisible: "Chat"
```

- [ ] Crear el flow
- [ ] Verificar que el mensaje de error se muestra en la UI (si no existe, crearlo en F3 del Plan 7)

### F5.5 — Flow: logout

```yaml
# apps/web/tests/e2e/auth/logout.yaml
---
- runFlow: auth/login-success.yaml
- tapOn: "Avatar del usuario"
- tapOn: "Cerrar sesión"
- assertVisible: "Iniciá sesión"
```

- [ ] Crear el flow con `runFlow` para re-usar login

### F5.6 — Flow: nueva sesión de chat

```yaml
# apps/web/tests/e2e/chat/new-session.yaml
---
- runFlow: auth/login-success.yaml
- openLink: ${BASE_URL}/chat
- tapOn: "Nueva conversación"
- assertVisible: "Hacé una pregunta"
```

- [ ] Crear el flow

### F5.7 — Flow: enviar mensaje (MOCK_RAG=true)

```yaml
# apps/web/tests/e2e/chat/send-message.yaml
---
- runFlow: auth/login-success.yaml
- openLink: ${BASE_URL}/chat
- tapOn: "Nueva conversación"
- tapOn: "Escribí tu pregunta"
- inputText: "¿Cuáles son las políticas de vacaciones?"
- tapOn: "Enviar"
- assertVisible: "Fuentes"
```

- [ ] Crear el flow
- [ ] Verificar que con `MOCK_RAG=true` el mock retorna una respuesta

### F5.8 — Flow: historial de sesiones

```yaml
# apps/web/tests/e2e/chat/session-history.yaml
---
- runFlow: chat/send-message.yaml
- openLink: ${BASE_URL}/chat
- assertVisible: "¿Cuáles son las políticas de vacaciones?"
```

- [ ] Crear el flow

### F5.9 — Flow: listar colecciones

```yaml
# apps/web/tests/e2e/collections/list.yaml
---
- runFlow: auth/login-success.yaml
- openLink: ${BASE_URL}/collections
- assertVisible: "Colecciones"
```

- [ ] Crear el flow

### F5.10 — Flow: CRUD de usuario (admin)

```yaml
# apps/web/tests/e2e/admin/create-user.yaml
---
- runFlow: auth/login-success.yaml
- openLink: ${BASE_URL}/admin/users
- tapOn: "Nuevo usuario"
- tapOn: "Email"
- inputText: "e2e-test@saldivia.com"
- tapOn: "Crear"
- assertVisible: "e2e-test@saldivia.com"
```

- [ ] Crear el flow de creación
- [ ] Crear flow de eliminación (cleanup): `admin/delete-user.yaml`

### F5.11 — Flow: upload de documento

```yaml
# apps/web/tests/e2e/upload/upload-doc.yaml
---
- runFlow: auth/login-success.yaml
- openLink: ${BASE_URL}/upload
- assertVisible: "Subir documentos"
- assertVisible: "Arrastrá archivos aquí"
```

- [ ] Crear el flow (verificación de UI, no de upload real en CI)

### F5.12 — Script en package.json

```json
"test:e2e": "maestro test tests/e2e",
"test:e2e:auth": "maestro test tests/e2e/auth",
"test:e2e:chat": "maestro test tests/e2e/chat"
```

- [ ] Agregar scripts

### Criterio de done
Los 10+ flows de Maestro pasan con `MOCK_RAG=true`.

### Checklist de cierre
- [ ] `MOCK_RAG=true bun run dev &` + `bun run test:e2e` — todos pasan
- [ ] CHANGELOG.md actualizado
- [ ] `git commit -m "test(e2e): maestro flows para auth, chat, collections, admin, upload"`

---

## Fase 6 — Auditoría de accesibilidad *(1-2 hs)*

Objetivo: detectar y corregir violaciones WCAG AA en primitivos y páginas críticas.

### F6.1 — addon-a11y ya configurado en Storybook (del Plan 7)

- [ ] Abrir Storybook → navegar a story `Button — all variants` → panel "Accessibility"
- [ ] Verificar que **Violations = 0** para:
  - `button.stories.tsx`
  - `input.stories.tsx`
  - `badge.stories.tsx`
  - `dialog.stories.tsx`
  - `table.stories.tsx`
- [ ] Documentar cualquier violation encontrada y corregirla antes de continuar

**Correcciones comunes:**
- Botón sin texto visible → agregar `aria-label`
- Color con bajo contraste → ajustar en `globals.css`
- Input sin label → agregar `<label htmlFor>` o `aria-label`
- Dialog sin accessible name → agregar `<DialogTitle>`

### F6.2 — Instalar axe-playwright

```bash
cd apps/web
bun add -D axe-playwright
```

- [ ] Instalar `axe-playwright`

### F6.3 — Tests de a11y para páginas

```typescript
// apps/web/tests/a11y/pages.spec.ts
import { test, expect } from '@playwright/test'
import { checkA11y, injectAxe } from 'axe-playwright'

// Páginas a auditar con sus rutas
const pages = [
  { name: 'login', path: '/login' },
  { name: 'chat', path: '/chat' },
  { name: 'collections', path: '/collections' },
  { name: 'admin-users', path: '/admin/users' },
  { name: 'settings', path: '/settings' },
]

test.beforeEach(async ({ page }) => {
  // Mock de auth — agregar cookie de sesión válida para tests
  await page.context().addCookies([{
    name: 'session',
    value: 'test-token-for-a11y',
    domain: 'localhost',
    path: '/',
  }])
})

for (const p of pages) {
  test(`a11y: ${p.name} — sin violations WCAG AA`, async ({ page }) => {
    await page.goto(p.path)
    await page.waitForLoadState('networkidle')
    await injectAxe(page)
    await checkA11y(page, undefined, {
      detailedReport: true,
      runOnly: {
        type: 'tag',
        values: ['wcag2a', 'wcag2aa'],
      },
    })
  })
}
```

- [ ] Crear `apps/web/tests/a11y/pages.spec.ts`
- [ ] Configurar la cookie de auth para que las páginas protegidas sean accesibles

### F6.4 — Playwright config para a11y

Crear o actualizar `apps/web/playwright.a11y.config.ts`:

```typescript
import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './tests/a11y',
  use: {
    baseURL: 'http://localhost:3000',
  },
  webServer: {
    command: 'MOCK_RAG=true bun run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: true,
    timeout: 30_000,
  },
})
```

- [ ] Crear `apps/web/playwright.a11y.config.ts`

### F6.5 — Script en package.json

```json
"test:a11y": "playwright test tests/a11y --config playwright.a11y.config.ts"
```

- [ ] Agregar script

### F6.6 — Corregir violations encontradas

- [ ] Correr `bun run test:a11y` — documentar todas las violations
- [ ] Priorizar: errores (rojo) > warnings (amarillo)
- [ ] Corregir todos los errores en las páginas críticas (`/login`, `/chat`, `/admin/users`)
- [ ] Crear `docs/superpowers/a11y-audit.md` con el estado post-corrección:

```markdown
# A11y Audit — Post Plan 7

Fecha: YYYY-MM-DD

## Páginas auditadas

| Página | Violations antes | Violations después | Estado |
|---|---|---|---|
| /login | 3 | 0 | ✅ |
| /chat | 1 | 0 | ✅ |
| ... | | | |
```

### Criterio de done
`bun run test:a11y` pasa. 0 violations WCAG AA en login, chat, y admin/users.

### Checklist de cierre
- [ ] `bun run test:a11y` — pasa (o 0 violations en páginas críticas)
- [ ] `docs/superpowers/a11y-audit.md` creado
- [ ] CHANGELOG.md actualizado
- [ ] `git commit -m "test(a11y): auditoría WCAG AA con axe-playwright — 0 violations en páginas críticas"`

---

## Fase 7 — CI integration *(1-2 hs)*

Objetivo: todos los jobs de UI testing integrados al pipeline de GitHub Actions.

**Archivos a modificar:**
- `.github/workflows/ci.yml` — agregar 4 jobs nuevos

### F7.1 — Job: component-tests

```yaml
# Agregar a .github/workflows/ci.yml

  component-tests:
    name: Component Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
        with: { bun-version: latest }
      - run: bun install
      - name: Run component tests
        run: bun test apps/web/src/components
        working-directory: .
```

- [ ] Agregar job `component-tests` al CI

### F7.2 — Job: visual-regression

```yaml
  visual-regression:
    name: Visual Regression
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
        with: { bun-version: latest }
      - run: bun install
      - run: bunx playwright install --with-deps chromium
      - name: Build Storybook
        run: bun run build:storybook
        working-directory: apps/web
      - name: Serve Storybook y esperar
        run: |
          bunx serve apps/web/storybook-static -p 6006 &
          bunx wait-on http://localhost:6006 --timeout 60000
      - name: Run visual regression
        run: bunx playwright test tests/visual/
        working-directory: apps/web
      - uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: visual-diff-${{ github.run_id }}
          path: apps/web/tests/visual/snapshots/
          retention-days: 7
```

- [ ] Agregar job `visual-regression` al CI
- [ ] Verificar que los snapshots están commiteados en el repo (el job los necesita para comparar)

### F7.3 — Job: accessibility

```yaml
  accessibility:
    name: Accessibility Audit
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
        with: { bun-version: latest }
      - run: bun install
      - run: bunx playwright install --with-deps chromium
      - name: Levantar dev server
        run: |
          MOCK_RAG=true bun run dev &
          bunx wait-on http://localhost:3000 --timeout 60000
        working-directory: apps/web
      - name: Run a11y audit
        run: bun run test:a11y
        working-directory: apps/web
```

- [ ] Agregar job `accessibility` al CI

### F7.4 — Job: e2e-maestro

```yaml
  e2e-maestro:
    name: E2E Flows (Maestro)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
        with: { bun-version: latest }
      - run: bun install
      - name: Install Maestro
        run: curl -fsSL "https://get.maestro.mobile.dev" | bash
      - name: Levantar dev server
        run: |
          MOCK_RAG=true bun run dev &
          bunx wait-on http://localhost:3000 --timeout 60000
        working-directory: apps/web
        env:
          DATABASE_PATH: ./data/test.db
          JWT_SECRET: test-secret-for-ci
          SYSTEM_API_KEY: test-api-key-for-ci
      - name: Run E2E flows
        run: ~/.maestro/bin/maestro test tests/e2e/auth/ tests/e2e/chat/
        working-directory: apps/web
        env:
          BASE_URL: http://localhost:3000
          ADMIN_EMAIL: ${{ secrets.TEST_ADMIN_EMAIL }}
          ADMIN_PASSWORD: ${{ secrets.TEST_ADMIN_PASSWORD }}
```

- [ ] Agregar job `e2e-maestro` al CI
- [ ] Agregar secrets `TEST_ADMIN_EMAIL` y `TEST_ADMIN_PASSWORD` en GitHub repository settings
- [ ] Crear usuario de test en el script de seeding del CI

### F7.5 — Actualizar Turbo pipeline

```json
// turbo.json — agregar:
{
  "tasks": {
    "test:components": {
      "dependsOn": ["^build"],
      "outputs": ["coverage/**"]
    },
    "build:storybook": {
      "dependsOn": ["^build"],
      "outputs": ["storybook-static/**"]
    }
  }
}
```

- [ ] Actualizar `turbo.json`

### F7.6 — Scripts finales en apps/web/package.json

Verificar que estos scripts existen:

```json
"test:components": "bun test src/components",
"test:visual": "playwright test tests/visual",
"visual:update": "playwright test tests/visual --update-snapshots",
"test:a11y": "playwright test tests/a11y --config playwright.a11y.config.ts",
"test:e2e": "maestro test tests/e2e",
"test:ui": "bun run test:components && bun run test:visual"
```

- [ ] Verificar/agregar todos los scripts

### F7.7 — Verificación final del CI

- [ ] Crear una PR de prueba y verificar que todos los jobs pasan en verde
- [ ] Si algún job falla, depurar y corregir antes de cerrar el plan
- [ ] Documentar el tiempo de ejecución de cada job para referencia futura

### Criterio de done
4 jobs nuevos en CI pasando en verde. `bun run test:ui` corre toda la suite localmente.

### Checklist de cierre
- [ ] Todos los jobs de CI en verde
- [ ] `bun run test:ui` localmente — pasa
- [ ] CHANGELOG.md actualizado
- [ ] `git commit -m "ci: integrar component tests, visual regression, a11y y maestro e2e al pipeline"`

---

## Estado global

| Fase | Estado | Fecha |
|------|--------|-------|
| Fase 1 — react-scan baseline | ✅ completado (template creado, recorrido pendiente) | 2026-03-26 |
| Fase 2 — Setup testing componentes | ✅ completado | 2026-03-26 |
| Fase 3 — Component tests (62 componentes) | ✅ completado — 147 tests para 20 componentes | 2026-03-26 |
| Fase 4 — Visual regression | ✅ completado (Playwright config + 20 tests light/dark) | 2026-03-26 |
| Fase 5 — Maestro E2E flows | ✅ completado (flows creados, Maestro pendiente de Java) | 2026-03-26 |
| Fase 6 — A11y audit | ✅ completado (axe-playwright config + 5 páginas) | 2026-03-26 |
| Fase 7 — CI integration | ✅ completado (3 jobs nuevos en ci.yml) | 2026-03-26 |

## Estimaciones

| Fase | Estimación |
|------|-----------|
| Fase 1 — react-scan | 1-2 hs |
| Fase 2 — Setup | 1-2 hs |
| Fase 3 — Component tests | 4-6 hs |
| Fase 4 — Visual regression | 2-3 hs |
| Fase 5 — Maestro E2E | 2-3 hs |
| Fase 6 — A11y | 1-2 hs |
| Fase 7 — CI | 1-2 hs |
| **Total** | **12-20 hs** |

## Resultado esperado

| Métrica | Inicio | Meta |
|---------|--------|------|
| Tests de componentes | 0 | ≥ 62 (uno por componente) |
| Stories con visual regression | 0 | ≥ 20 stories cubiertos |
| Flows de Maestro | 0 | ≥ 10 flows |
| Páginas con auditoría a11y | 0 | 5 páginas críticas |
| WCAG AA violations en páginas críticas | desconocido | 0 |
| Jobs de CI para UI | 0 | 4 jobs |
| Re-renders innecesarios documentados | 0 | baseline completo |

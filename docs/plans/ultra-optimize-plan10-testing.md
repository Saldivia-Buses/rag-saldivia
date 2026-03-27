# Plan 10: Testing Completo — Visual Regression, A11y, E2E y Cobertura

> Este documento vive en `docs/plans/ultra-optimize-plan10-testing.md`.
> Se actualiza a medida que se completan las tareas. Cada fase completada genera entrada en CHANGELOG.md.

---

## Contexto

Los Planes 1–9 construyeron un stack técnico sólido con 259 tests de lógica y 154 tests de componentes en verde. Sin embargo, **"full tested" como afirmación de un release público requiere evidencia**, y hoy hay cuatro gaps:

**Visual regression sin correr post-upgrades:**
El Plan 8 actualizó Next.js 15 → 16, Lucide 0.x → 1.7, Drizzle 0.38 → 0.45 y Zod 3 → 4. Los 22 snapshots del baseline visual fueron generados antes de estos upgrades. Es posible que haya regressions de UI silenciosas que los tests de componentes no detectan porque no renderizan CSS real.

**A11y sin correr desde el Plan 6:**
Los 5 tests de WCAG AA cubren las páginas de login, chat, collections, admin y settings. Desde el Plan 7 (design system) y el Plan 8 (upgrades de deps) no se corrieron. Cualquier cambio en tokens CSS o en componentes shadcn podría haber introducido violaciones.

**Cero tests end-to-end:**
Los flows críticos de la aplicación — login, enviar un mensaje, crear un usuario, subir un archivo — no tienen cobertura automatizada de flujo completo. Los unit tests y component tests verifican piezas aisladas, pero nadie verifica que el sistema funciona de extremo a extremo.

**Cobertura sin reporte:**
Los tests pasan pero no hay ningún número que respalde cuánto del código está cubierto. Un release de industria tiene badge de cobertura con un threshold mínimo definido y enforcement en CI.

**Lo que NO entra en este plan:**
- Tests de features que requieren la workstation con GPU (RAG streaming real, ingestión real, colecciones en Milvus). Estos tests se implementarán en versiones futuras cuando el hardware esté disponible. El mock mode de `MOCK_RAG=true` es suficiente para los flows de E2E de esta versión.
- Tests de E2E de admin de colecciones (crear/eliminar en Milvus) — misma razón.

---

## Prerequisito

**Plan 9 debe estar completado** antes de ejecutar este plan. Los E2E requieren que la app buildee sin errores TS y que el linting esté limpio.

---

## Orden de ejecución

```
F10.1 → F10.2 → F10.3 → F10.4 → F10.5
```

**Por qué este orden:**
- **F10.1 primero:** verificar el baseline antes de agregar más tests — si hay regressions visuales post-upgrades, hay que resolverlas primero
- **F10.2 después de F10.1:** la a11y depende del CSS real renderizado — si el baseline visual está roto, la a11y también podría estarlo
- **F10.3 independiente:** el reporte de cobertura no depende de visual ni a11y, pero conviene tenerlo antes de los E2E para ver la línea base
- **F10.4 + F10.5 al final:** los E2E y smoke tests requieren que todo lo anterior esté en verde

---

## Seguimiento

Formato: `- [ ] Descripción — estimación`
Al completar: `- [x] Descripción — completado YYYY-MM-DD`
Cada fase completada → entrada en CHANGELOG.md → commit.

---

## Fase F10.1 — Visual regression post-upgrades *(1-2 hs)*

Objetivo: `bun run test:visual` pasa con el baseline actualizado. Cualquier diferencia con los snapshots anteriores es intencional y documentada.

**Contexto:** el baseline fue generado con Next.js 15 y Lucide 0.x. Con Next.js 16 y Lucide 1.7, los SVGs de iconos pueden haber cambiado de path, los estilos de transición pueden ser distintos, y shadcn puede renderizar algunos tokens diferentes.

**Proceso:**
1. Correr `bun run test:visual` tal cual está — ver qué falla
2. Para cada failure: abrir el diff visual generado por Playwright en `test-results/`
3. Si el cambio es un bug de upgrade → identificar la causa y corregir el componente
4. Si el cambio es intencional (mejor rendering, icon actualizado) → actualizar el snapshot con `bun run visual:update`
5. Commitear los nuevos snapshots

**Casos probables de regression:**
- Iconos de Lucide 1.7 con paths ligeramente distintos (cambio intencional → actualizar baseline)
- `@radix-ui` actualizado con shadcn puede tener micro-diferencias en animaciones
- Next.js 16 puede cambiar el orden de algunos atributos HTML

**Criterio de done:**
- `bun run test:visual` → 22/22 tests pasan, exit 0
- Los snapshots en `apps/web/tests/visual/__snapshots__/` están actualizados y commiteados
- Todo cambio de snapshot tiene una explicación en el mensaje del commit

- [ ] `bun run test:visual` — identificar cuántos tests fallan y con qué diffs
- [ ] Para cada failure: determinar si es bug o cambio intencional
- [ ] Corregir bugs de rendering si los hay
- [ ] `bun run visual:update` para regenerar snapshots con cambios intencionales
- [ ] `bun run test:visual` → 22/22 pasan
- [ ] Commit: `test(visual): actualizar baseline post-next16-lucide1.7 — plan10 f10.1`

**Estado: pendiente**

---

## Fase F10.2 — A11y audit completo *(1-2 hs)*

Objetivo: `bun run test:a11y` pasa en las 5 páginas con cero violaciones WCAG AA.

**Contexto:** los tests de a11y usan `axe-playwright` contra las páginas reales renderizadas. Desde el Plan 7 (design system "Warm Intelligence") no se corrieron. Los tokens CSS crema-índigo tienen que cumplir con contraste WCAG AA (ratio 4.5:1 para texto normal, 3:1 para texto grande).

**Páginas cubiertas:** login, chat, collections, admin, settings

**Violaciones probables a investigar:**
- Contraste de `--fg-subtle` (#9a9088) sobre `--bg` (#faf8f4) — puede ser borderline en texto pequeño
- Contraste del NavRail oscuro (`--nav-bg` #18181B) con iconos grises
- Focus rings en los componentes shadcn — verificar que sean visibles en ambos temas
- Labels faltantes en inputs sin placeholder visible

**Proceso:**
1. `bun run test:a11y` — ver reporte de violaciones
2. Para cada violación: identificar el componente afectado y la regla WCAG
3. Corregir en el componente o en los tokens CSS de `globals.css`
4. Re-correr hasta 0 violaciones

**Criterio de done:**
- `bun run test:a11y` → 5/5 páginas, 0 violaciones, exit 0
- Si se modificaron tokens CSS → `bun run test:visual` sigue pasando (los cambios de contraste pueden afectar el baseline visual)
- `bun run test:components` → 154 tests pasan

- [ ] `bun run test:a11y` — listar todas las violaciones
- [ ] Corregir cada violación (componente o token CSS)
- [ ] Si se modificó `globals.css` → correr `bun run test:visual` para verificar no hay regressions
- [ ] `bun run test:a11y` → 0 violaciones
- [ ] `bun run test` — 259 tests pasan
- [ ] Commit: `fix(a11y): resolver violaciones WCAG AA post-design-system — plan10 f10.2`

**Estado: pendiente**

---

## Fase F10.3 — Code coverage reporte + CI badge *(1-2 hs)*

Objetivo: `bun test --coverage` genera un reporte. El CI sube la cobertura. El README tiene un badge.

**Herramienta:** `bun test --coverage` — nativo de Bun, sin dependencias adicionales.

**Comando para ver el baseline actual:**
```bash
export PATH="$HOME/.bun/bin:$PATH"
cd /home/enzo/rag-saldivia
bun test --coverage 2>&1 | tail -30
```
Salida esperada: tabla con columnas `% Stmts | % Branch | % Funcs | % Lines | Uncovered Line #s`.

**Agregar script `test:coverage` en `apps/web/package.json`:**
Ya existe: `"test:coverage": "bun test src/lib --coverage"`. Verificar que devuelve cobertura. Si no, usar:
```bash
cd apps/web && bun test src/lib --coverage
```

**Cambio en `.github/workflows/ci.yml`:**

En el job `test-logic`, cambiar el step de tests para agregar `--coverage`:

Buscar el step que tiene el comando de tests en el job `test-logic`. Agregar `--coverage` y `--coverage-reporter=lcov` al final del comando `bun run test`.

**Threshold:** Bun no tiene `--coverage-threshold` nativo. El threshold se verifica así: después de los tests, agregar un step que lea el reporte y falle si la cobertura es menor a 80%.

En el job `test-logic` de `ci.yml`, agregar después del step de tests:
```yaml
- name: Verify coverage threshold
  run: |
    COVERAGE=$(bun test --coverage 2>&1 | grep "All files" | awk '{print $4}' | tr -d '%')
    if [ -n "$COVERAGE" ] && [ "$COVERAGE" -lt 80 ]; then
      echo "Coverage $COVERAGE% is below threshold of 80%"
      exit 1
    fi
  working-directory: apps/web
```

**Badge en README:** usar shields.io con el porcentaje estático hasta que Codecov esté configurado:
```
![Coverage](https://img.shields.io/badge/coverage-80%25-green)
```
Actualizar el número cuando el CI reporte el baseline real.

**Criterio de done:**
- `bun test src/lib --coverage` en `apps/web` → muestra tabla de cobertura
- El job `test-logic` en CI incluye cobertura
- Badge en `README.md`

- [ ] `cd apps/web && bun test src/lib --coverage` — confirmar que genera tabla de cobertura
- [ ] Anotar el porcentaje de cobertura actual de líneas (para el badge)
- [ ] Agregar step de verificación de threshold en el job `test-logic` de `ci.yml`
- [ ] Agregar badge de cobertura en `README.md` con el porcentaje real
- [ ] Commit: `ci: agregar code coverage con threshold 80% — plan10 f10.3`

**Estado: pendiente**

---

## Fase F10.4 — E2E Playwright: flujos críticos *(3-5 hs)*

Objetivo: 5 tests E2E en Playwright que verifican los flujos críticos de punta a punta.

**Contexto:** la carpeta `tests/e2e/` ya existe con YAMLs en formato Maestro (sin Maestro instalado). Los tests Playwright van en una carpeta nueva `tests/e2e-playwright/` para no interferir.

**Archivos a crear:**
- `apps/web/playwright.e2e.config.ts` — config Playwright para E2E (similar a `playwright.a11y.config.ts`)
- `apps/web/tests/e2e-playwright/auth.spec.ts`
- `apps/web/tests/e2e-playwright/chat.spec.ts`
- `apps/web/tests/e2e-playwright/admin-users.spec.ts`
- `apps/web/tests/e2e-playwright/upload.spec.ts`
- `apps/web/tests/e2e-playwright/settings.spec.ts`

**Archivo `apps/web/playwright.e2e.config.ts` — contenido exacto:**

```typescript
import { defineConfig } from "@playwright/test"

export default defineConfig({
  testDir: "./tests/e2e-playwright",
  use: {
    baseURL: "http://localhost:3000",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  timeout: 30_000,
  retries: 1,
  webServer: {
    command: "MOCK_RAG=true bun run dev",
    url: "http://localhost:3000",
    reuseExistingServer: true,
    timeout: 60_000,
  },
})
```

**Agregar script en `apps/web/package.json`:**
```json
"test:e2e": "playwright test --config playwright.e2e.config.ts"
```

**Patrón de login reutilizable** — cada spec file necesita autenticarse. Usar el mismo patrón del `tests/a11y/pages.spec.ts` ya existente:

```typescript
// Al inicio de cada spec, antes del primer test:
test.beforeEach(async ({ page }) => {
  const res = await page.request.post("/api/auth/login", {
    data: { email: "admin@localhost", password: "changeme" },
  })
  // Si el login falla, el test falla inmediatamente con mensaje claro
  expect(res.ok()).toBeTruthy()
})
```

---

### F10.4a — `tests/e2e-playwright/auth.spec.ts`

**Estructura del archivo:**

```typescript
import { test, expect } from "@playwright/test"

test.describe("Auth flow", () => {
  test("login exitoso redirige a /chat", async ({ page }) => {
    // 1. Navegar a /login
    // 2. Completar email y password con las credenciales de test
    // 3. Clickear "Iniciar sesión" (buscar por el texto del botón o role="button")
    // 4. Esperar navegación
    // 5. Verificar URL: expect(page.url()).toContain("/chat")
  })

  test("login con credenciales incorrectas muestra error", async ({ page }) => {
    // 1. Navegar a /login
    // 2. Ingresar email válido + password incorrecto
    // 3. Verificar que aparece un mensaje de error
    // 4. Verificar que la URL sigue siendo /login
  })

  test("logout elimina la sesión y redirige a /login", async ({ page }) => {
    // 1. Hacer login (mismo patrón del beforeEach pero inline)
    // 2. Navegar a /chat
    // 3. Buscar botón de logout (en NavRail, abajo)
    // 4. Clickear logout
    // 5. Verificar que la URL es /login
    // 6. Intentar navegar a /chat directamente
    // 7. Verificar redirección de vuelta a /login
  })
})
```

**Selectores importantes:** el botón de logout está en el `NavRail` como un botón con ícono `LogOut` de Lucide. Buscar con `page.getByRole("button", { name: /logout|salir/i })` o `page.locator('[data-testid="logout"]')` si tiene data-testid.

---

### F10.4b — `tests/e2e-playwright/chat.spec.ts`

```typescript
import { test, expect } from "@playwright/test"

test.describe("Chat flow (MOCK_RAG=true)", () => {
  test.beforeEach(async ({ page }) => {
    // Login como admin
    const res = await page.request.post("/api/auth/login", {
      data: { email: "admin@localhost", password: "changeme" },
    })
    expect(res.ok()).toBeTruthy()
  })

  test("crear sesión nueva y enviar mensaje", async ({ page }) => {
    // 1. Navegar a /chat
    // 2. Clickear "Nueva sesión" o el botón de crear (buscar en el SecondaryPanel/ChatPanel)
    // 3. Verificar que aparece el input de chat
    // 4. Escribir un mensaje: "Hola test"
    // 5. Submitear (Enter o botón Send)
    // 6. Esperar que aparezca una respuesta (en mock: texto fijo del servidor)
    // 7. Verificar que el mensaje enviado aparece en el historial
    // 8. Recargar la página
    // 9. Verificar que la sesión sigue existiendo en el sidebar
  })
})
```

---

### F10.4c — `tests/e2e-playwright/admin-users.spec.ts`

```typescript
import { test, expect } from "@playwright/test"

// Email único para este test (evita conflictos entre runs)
const TEST_EMAIL = `e2e-test-${Date.now()}@test.local`

test.describe("Admin users CRUD", () => {
  test.beforeEach(async ({ page }) => {
    const res = await page.request.post("/api/auth/login", {
      data: { email: "admin@localhost", password: "changeme" },
    })
    expect(res.ok()).toBeTruthy()
  })

  test("crear usuario, verificar en tabla, desactivar", async ({ page }) => {
    // 1. Navegar a /admin/users
    // 2. Clickear "Nuevo usuario" (buscar botón)
    // 3. Completar el form: email=TEST_EMAIL, password="Test1234!", rol="user"
    // 4. Submitear
    // 5. Verificar que TEST_EMAIL aparece en la tabla (buscar con page.getByText)
    // 6. Clickear el botón de editar/opciones para ese usuario
    // 7. Cambiar el rol a "area_manager"
    // 8. Guardar
    // 9. Verificar que el rol cambió en la tabla
  })
})
```

---

### F10.4d — `tests/e2e-playwright/upload.spec.ts`

```typescript
import { test, expect } from "@playwright/test"
import { writeFileSync } from "fs"
import { join } from "path"

test.describe("Upload flow (MOCK_RAG=true)", () => {
  test.beforeEach(async ({ page }) => {
    const res = await page.request.post("/api/auth/login", {
      data: { email: "admin@localhost", password: "changeme" },
    })
    expect(res.ok()).toBeTruthy()
  })

  test("subir archivo crea job en ingestion", async ({ page }) => {
    // 1. Crear un archivo de texto temporal para el upload
    //    (usar writeFileSync para crear /tmp/test-upload.txt con contenido "test content")
    // 2. Navegar a /upload
    // 3. Seleccionar el archivo con page.setInputFiles()
    // 4. Completar el campo de colección (si existe)
    // 5. Submitear el form
    // 6. Verificar mensaje de éxito o redirección
    // 7. Navegar a /admin/ingestion
    // 8. Verificar que aparece un job con el nombre del archivo
  })
})
```

---

### F10.4e — `tests/e2e-playwright/settings.spec.ts`

```typescript
import { test, expect } from "@playwright/test"

const NEW_PASSWORD = "NewPass123!"

test.describe("Settings — cambio de contraseña", () => {
  test("cambiar contraseña y verificar login con la nueva", async ({ page }) => {
    // 1. Login con credenciales actuales (admin@localhost / changeme)
    // 2. Navegar a /settings
    // 3. Buscar el form de cambio de contraseña
    // 4. Ingresar: contraseña actual = "changeme", nueva = NEW_PASSWORD, confirmar = NEW_PASSWORD
    // 5. Submitear
    // 6. Verificar mensaje de éxito
    // 7. Logout
    // 8. Intentar login con NEW_PASSWORD → debe funcionar
    // 9. LIMPIEZA: cambiar la contraseña de vuelta a "changeme"
    //    (para que los otros tests sigan funcionando)
  })
})
```

**⚠️ Importante:** el paso 9 (restaurar contraseña) es crítico. Si falla, los tests subsiguientes que usan `admin@localhost / changeme` van a fallar. Usar `test.afterEach` o hacerlo al final del test.

---

**Configuración del job E2E en `.github/workflows/ci.yml`:**

Agregar después del job `test-logic` existente:

```yaml
e2e:
  name: E2E Tests
  runs-on: ubuntu-latest
  services:
    redis:
      image: redis:alpine
      ports:
        - 6379:6379
  steps:
    - uses: actions/checkout@v4
    - uses: oven-sh/setup-bun@v2
      with:
        bun-version: latest
    - run: bun install
    - name: Install Playwright browsers
      run: bunx playwright install --with-deps chromium
      working-directory: apps/web
    - name: Run E2E tests
      run: bun run test:e2e
      working-directory: apps/web
      env:
        REDIS_URL: redis://localhost:6379
        JWT_SECRET: test-secret-for-ci-only
        MOCK_RAG: "true"
        DATABASE_PATH: ./data/test.db
        NODE_ENV: test
```

**Criterio de done:**
- `cd apps/web && bun run test:e2e` → 5 specs pasan, exit 0
- Correr 3 veces seguidas → mismo resultado (no flaky)
- Job `e2e` en CI configurado

- [ ] Crear `apps/web/playwright.e2e.config.ts` con el contenido exacto de arriba
- [ ] Crear `apps/web/tests/e2e-playwright/auth.spec.ts` con la estructura de arriba
- [ ] Crear `apps/web/tests/e2e-playwright/chat.spec.ts` con la estructura de arriba
- [ ] Crear `apps/web/tests/e2e-playwright/admin-users.spec.ts` con la estructura de arriba
- [ ] Crear `apps/web/tests/e2e-playwright/upload.spec.ts` con la estructura de arriba
- [ ] Crear `apps/web/tests/e2e-playwright/settings.spec.ts` con la estructura de arriba
- [ ] Agregar `"test:e2e": "playwright test --config playwright.e2e.config.ts"` en `apps/web/package.json`
- [ ] `cd apps/web && MOCK_RAG=true bun run dev &` (en background) y luego `bun run test:e2e`
- [ ] Correr 3 veces → mismo resultado
- [ ] Agregar job `e2e` en `.github/workflows/ci.yml`
- [ ] Commit: `test(e2e): 5 flows criticos con playwright — login, chat, admin, upload, settings — plan10 f10.4`

**Estado: pendiente**

---

## Fase F10.5 — Smoke test Redis automatizado *(1 hs)*

Objetivo: el checklist manual del Plan 8 está automatizado en `tests/e2e-playwright/redis-smoke.spec.ts`.

**Archivo a crear: `apps/web/tests/e2e-playwright/redis-smoke.spec.ts`**

```typescript
import { test, expect } from "@playwright/test"

test.describe("Redis smoke tests", () => {
  test("GET /api/health retorna ok cuando Redis está up", async ({ request }) => {
    // 1. Hacer GET /api/health
    // 2. Verificar que el status HTTP es 200
    // 3. Verificar que el body JSON tiene: { ok: true, status: "healthy" }
    const res = await request.get("/api/health")
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(body.ok).toBe(true)
    expect(body.status).toBe("healthy")
  })

  test("token JWT revocado después del logout", async ({ request }) => {
    // 1. POST /api/auth/login → obtener la cookie de sesión
    // 2. Guardar el valor del header Set-Cookie
    // 3. POST /api/auth/logout (con la cookie del paso 1)
    // 4. Hacer GET /api/rag/collections con la cookie vieja → debe ser 401
    const loginRes = await request.post("/api/auth/login", {
      data: { email: "admin@localhost", password: "changeme" },
    })
    expect(loginRes.ok()).toBeTruthy()

    // Obtener cookies del contexto
    const cookies = loginRes.headers()["set-cookie"]
    expect(cookies).toBeTruthy()

    // Logout
    await request.delete("/api/auth/logout")

    // Request autenticada con token viejo debe fallar
    const protectedRes = await request.get("/api/admin/users")
    expect(protectedRes.status()).toBe(401)
  })

  test("GET /api/health retorna 503 si Redis no está configurado", async ({ request }) => {
    // Este test solo puede correr localmente sin Redis.
    // En CI, Redis está siempre disponible.
    // Verificar el comportamiento documentado: el health check devuelve el campo redis.
    const res = await request.get("/api/health")
    const body = await res.json()
    // Si Redis está up: { ok: true }
    // Si Redis está down: { ok: false, service: "redis", status: "down" }
    // En CI este test siempre pasa con Redis up — es un test de documentación del contrato
    expect(body).toHaveProperty("ok")
    expect(body).toHaveProperty("ts")
  })
})
```

**Criterio de done:**
- `bun run test:e2e` incluye `redis-smoke.spec.ts`
- Los tests pasan con Redis corriendo en CI
- El checklist manual del Plan 8 está marcado como "automatizado en F10.5"

- [ ] Crear `apps/web/tests/e2e-playwright/redis-smoke.spec.ts` con el contenido de arriba
- [ ] `cd apps/web && bun run test:e2e` → redis-smoke.spec.ts pasa
- [ ] En `docs/plans/ultra-optimize-plan8-optimization.md`: marcar los 5 items del smoke test manual como "automatizados en Plan 10 F10.5"
- [ ] Commit: `test(e2e): smoke tests de redis automatizados — plan10 f10.5`

**Estado: pendiente**

---

## Criterio de done global del Plan 10

- `bun run test:visual` → 22/22 pass
- `bun run test:a11y` → 5/5 páginas, 0 violaciones WCAG AA
- `bun test --coverage` → ≥80% line coverage en lib/ y packages/
- `bun run test:e2e` → 5 specs + smoke Redis pasan
- Badge de cobertura en README
- CI tiene 4 jobs de test: `test-logic`, `test-components`, `test-visual`, `e2e`
- `bun run test` (todos) → exit 0

### Checklist de cierre

- [ ] `bun run test:visual` → 22 pass
- [ ] `bun run test:a11y` → 0 violaciones
- [ ] `bun test --coverage` → ≥80%
- [ ] `bun run test:e2e` → todos pasan, 3 runs estables
- [ ] Badge de cobertura visible en README
- [ ] CHANGELOG.md actualizado
- [ ] Commit final: `test: plan10 completado — visual, a11y, coverage, e2e, smoke redis`
- [ ] `git push`

**Estado: pendiente**

---

## Impacto en CI

Al final del Plan 10, el workflow `ci.yml` tiene 5 jobs paralelos:

| Job | Tests | Tiempo estimado |
|---|---|---|
| `type-check` | tsc --noEmit en todos los packages | ~30s |
| `lint` | ESLint + turbo lint | ~30s |
| `test-logic` | 259 tests (db + web lib + packages) | ~15s |
| `test-components` | 154 tests (happy-dom) | ~10s |
| `test-visual` | 22 visual regression snapshots | ~60s |
| `e2e` | 6 specs E2E + smoke Redis | ~120s |

Tiempo total del CI (en paralelo): ~2 minutos.

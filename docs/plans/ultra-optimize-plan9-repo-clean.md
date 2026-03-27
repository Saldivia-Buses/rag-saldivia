# Plan 9: Repo Limpio — Cero Dead Code, Cero Errores, Linting Perfecto

> Este documento vive en `docs/plans/ultra-optimize-plan9-repo-clean.md`.
> Se actualiza a medida que se completan las tareas. Cada fase completada genera entrada en CHANGELOG.md.
>
> **Nota para quien ejecute este plan:** cada checkbox es un paso atómico. No saltear ninguno.
> Si un comando falla, detener y reportar el error antes de continuar.

---

## Contexto

El Plan 8 completó la optimización técnica del stack. Con Redis obligatorio, BullMQ, Next.js 16, Zod 4 y ~2.516 líneas eliminadas, el código es sólido — pero el **repo como artefacto** todavía no está a la altura de un release público.

Tres problemas concretos:

**Archivos que no deberían estar en el repo remoto (46 archivos confirmados):**
Logs de sesiones del MCP (`.playwright-mcp/`), HTMLs y logs de brainstorming (`.superpowers/`), PNGs del browser MCP, un log de runtime, un archivo de env, y specs de diseño interno (`docs/superpowers/`). Están en el remoto ahora mismo.

**Errores TypeScript sin resolver (5 en 2 archivos):**
- `updateTag` no existe en `next/cache` de Next.js 16 (`collections.ts:3`)
- `IngestionEventRecord` incompatible con `exactOptionalPropertyTypes` en `blackbox.ts` (líneas 115, 131, 148, 170)

**Dead code confirmado (8 archivos, 3 deps):**
Features que requieren la workstation con GPU (crossdoc, colección history, split view), un stub de SSO, y wrappers de utilidades nunca usados. Las dependencias `next-safe-action`, `d3` y `@types/d3` no tienen ningún consumidor.

**Lo que NO cambia:** la lógica de negocio, auth, schema de DB, tests existentes.

---

## Orden de ejecución

```
F9.1 → F9.2 → F9.3 → F9.4 → F9.5 → F9.6 → F9.7 → F9.8 → F9.9
```

---

## Seguimiento

Formato: `- [ ] Descripción — estimación`
Al completar: `- [x] Descripción — completado YYYY-MM-DD`
Cada fase completada → entrada en CHANGELOG.md → commit.

---

## Fase F9.1 — Git purge + .gitignore perfecto *(1-2 hs)*

Objetivo: que `git ls-files | grep -E "(\.log|superpowers|playwright-mcp|docs/superpowers|config/\.env)"` devuelva exactamente cero líneas.

**Lista completa de archivos a destrackear (46 archivos — confirmados con `git ls-files`):**

```
.playwright-mcp/console-2026-03-25T00-51-49-379Z.log
.playwright-mcp/console-2026-03-25T00-58-44-542Z.log
.playwright-mcp/console-2026-03-25T01-32-52-713Z.log
.playwright-mcp/console-2026-03-25T17-52-39-824Z.log
.playwright-mcp/console-2026-03-25T17-53-58-305Z.log
.playwright-mcp/console-2026-03-25T17-54-21-544Z.log
.playwright-mcp/console-2026-03-25T17-55-22-769Z.log
.playwright-mcp/console-2026-03-26T04-23-45-664Z.log
.playwright-mcp/console-2026-03-26T04-27-17-494Z.log
.playwright-mcp/console-2026-03-26T04-41-41-575Z.log
.playwright-mcp/console-2026-03-26T15-18-48-920Z.log
.playwright-mcp/page-2026-03-25T17-53-22-812Z.png
.playwright-mcp/page-2026-03-25T17-53-51-738Z.png
.playwright-mcp/page-2026-03-25T17-54-05-350Z.png
.superpowers/brainstorm/22080-1774410234/.server-info
.superpowers/brainstorm/22080-1774410234/.server.log
.superpowers/brainstorm/22080-1774410234/.server.pid
.superpowers/brainstorm/22080-1774410234/design-direction.html
.superpowers/brainstorm/22080-1774410234/final-8.html
.superpowers/brainstorm/22080-1774410234/layout-admin.html
.superpowers/brainstorm/22080-1774410234/layout-app.html
.superpowers/brainstorm/22080-1774410234/palette-light.html
.superpowers/brainstorm/22080-1774410234/roadmap-v2.html
.superpowers/brainstorm/22080-1774410234/roadmap.html
.superpowers/brainstorm/22080-1774410234/stolen-ideas-2.html
.superpowers/brainstorm/22080-1774410234/stolen-ideas-3.html
.superpowers/brainstorm/22080-1774410234/stolen-ideas-4.html
.superpowers/brainstorm/22080-1774410234/stolen-ideas.html
.superpowers/brainstorm/22080-1774410234/waiting-1.html
.superpowers/brainstorm/22080-1774410234/waiting-2.html
.superpowers/brainstorm/22080-1774410234/waiting-final.html
.superpowers/brainstorm/23691-1774492720/.server-stopped
.superpowers/brainstorm/23691-1774492720/.server.log
.superpowers/brainstorm/23691-1774492720/.server.pid
.superpowers/brainstorm/23691-1774492720/acento.html
.superpowers/brainstorm/23691-1774492720/bienvenida.html
.superpowers/brainstorm/23691-1774492720/densidad.html
.superpowers/brainstorm/23691-1774492720/estetica.html
.superpowers/brainstorm/23691-1774492720/recursos.html
.superpowers/brainstorm/23691-1774492720/resumen-diseno.html
.superpowers/brainstorm/23691-1774492720/tipografia.html
apps/web/logs/backend.log
config/.env.saldivia
docs/superpowers/react-scan-baseline.md
docs/superpowers/specs/2026-03-25-product-roadmap-design.md
docs/superpowers/specs/2026-03-26-design-system-design.md
docs/superpowers/specs/2026-03-26-ui-testing-design.md
```

**Comandos exactos para destrackear (ejecutar en el root del repo):**

```bash
git rm --cached -r .playwright-mcp/
git rm --cached -r .superpowers/
git rm --cached apps/web/logs/backend.log
git rm --cached config/.env.saldivia
git rm --cached -r docs/superpowers/
```

**Entradas exactas a agregar al `.gitignore`** (agregar al final del archivo, antes de `# OS`):

```
# MCP browser session artifacts
.playwright-mcp/

# Brainstorming artifacts (internal design sessions)
.superpowers/

# App runtime logs (covered by logs/ but adding specific path)
apps/web/logs/

# Config env files by environment
config/*.env.*

# Internal design specs
docs/superpowers/

# Process ID files
*.pid
*.server-info
*.server-stopped
```

**Criterio de done:**
- `git ls-files | grep -E "(playwright-mcp|superpowers|\.log$|config/\.env)"` → salida vacía
- `.env.example` sigue trackeado: `git ls-files | grep "\.env\.example"` → `.env.example`

- [ ] Ejecutar `git ls-files | grep -E "(playwright-mcp|superpowers|\.log$|config/\.env)"` — confirmar que la lista coincide con los 46 archivos de arriba. Si hay más, agregarlos a la lista antes de continuar.
- [ ] Ejecutar `git rm --cached -r .playwright-mcp/`
- [ ] Ejecutar `git rm --cached -r .superpowers/`
- [ ] Ejecutar `git rm --cached apps/web/logs/backend.log`
- [ ] Ejecutar `git rm --cached config/.env.saldivia`
- [ ] Ejecutar `git rm --cached -r docs/superpowers/`
- [ ] Agregar las 8 entradas nuevas al `.gitignore` (al final, antes del comentario `# OS`)
- [ ] `git ls-files | grep -E "(playwright-mcp|superpowers|\.log$|config/\.env)"` → salida vacía (confirmar)
- [ ] `git ls-files | grep "\.env\.example"` → `.env.example` sigue presente (confirmar)
- [ ] `export PATH="$HOME/.bun/bin:$PATH" && cd apps/web && bun run test` — confirmar 259 tests pasan
- [ ] Commit: `chore(git): purgar 46 archivos incorrectamente trackeados + ampliar gitignore — plan9 f9.1`
- [ ] `git push` — confirmar en GitHub que los archivos desaparecieron

**Estado: pendiente**

---

## Fase F9.2 — TypeScript a cero errores *(30-45 min)*

Objetivo: `cd apps/web && bunx tsc --noEmit` devuelve exit code 0 con exactamente 0 líneas de output.

### Error 1 — `apps/web/src/app/actions/collections.ts:3`

**Problema:** `updateTag` no existe en `next/cache` en Next.js 16. Es una API de la directiva `"use cache"`, no del App Router cache tradicional.

> **Nota sobre la solución real aplicada:** `revalidateTag` tampoco funciona en Next.js 16 sin un segundo argumento. La solución correcta fue eliminar `updateTag` completamente y reemplazarlo por `invalidateCollectionsCache()` — el mecanismo Redis ya existente en el proyecto para invalidar la cache de colecciones. El import quedó solo `import { revalidatePath } from "next/cache"` más el import de `invalidateCollectionsCache`.

La lógica en cada action quedó: `await invalidateCollectionsCache()` + `revalidatePath("/collections")`.

Hay exactamente dos llamadas, en `actionCreateCollection` y `actionDeleteCollection`.

### Errores 2–5 — `packages/logger/src/blackbox.ts` líneas 115, 131, 148, 170

**Problema:** `exactOptionalPropertyTypes: true` no permite asignar `string | undefined` a `string?`. El tipo `IngestionEventRecord` tiene `filename?: string` (sin `undefined` explícito), pero el código pasa expresiones de tipo `string | undefined`.

**Patrón actual (en las 4 funciones `handleIngestion*`):**
```typescript
state.ingestionEvents.push({
  ts: event.ts,
  type: event.type,
  filename: payload["filename"] != null ? String(payload["filename"]) : undefined,
  collection: payload["collection"] != null ? String(payload["collection"]) : undefined,
  sourceId: payload["sourceId"] != null ? String(payload["sourceId"]) : undefined,
})
```

**Patrón correcto (reemplazar en las 4 instancias):**
```typescript
state.ingestionEvents.push({
  ts: event.ts,
  type: event.type,
  ...(payload["filename"] != null ? { filename: String(payload["filename"]) } : {}),
  ...(payload["collection"] != null ? { collection: String(payload["collection"]) } : {}),
  ...(payload["sourceId"] != null ? { sourceId: String(payload["sourceId"]) } : {}),
  ...(payload["error"] != null ? { error: String(payload["error"]) } : {}),
})
```

**Nota:** `handleIngestionFailed` usa `error` en lugar de `sourceId` — aplicar el mismo patrón para ese campo. `handleIngestionStalled` no tiene `sourceId` ni `error` — omitir esas líneas.

### Error adicional — `packages/db/src/test-setup.ts`

**Problema:** `ioredis-mock` no tiene tipos declarados.

**Solución:** crear el archivo `packages/db/src/ioredis-mock.d.ts` con este contenido exacto:

```typescript
declare module "ioredis-mock"
```

**Verificación final:**

```bash
cd apps/web && bunx tsc --noEmit
# Salida esperada: (vacía) — exit code 0

cd ../../packages/db && bunx tsc --noEmit
# Salida esperada: (vacía) — exit code 0
```

- [x] En `apps/web/src/app/actions/collections.ts`: remover `updateTag`, reemplazar con `invalidateCollectionsCache()` + solo `revalidatePath` — completado 2026-03-27
- [x] En `packages/logger/src/blackbox.ts`: aplicar spread condicional en `handleIngestionStarted` (línea ~115) — completado 2026-03-27
- [x] En `packages/logger/src/blackbox.ts`: aplicar spread condicional en `handleIngestionCompleted` (línea ~131) — completado 2026-03-27
- [x] En `packages/logger/src/blackbox.ts`: aplicar spread condicional en `handleIngestionFailed` (línea ~148) — completado 2026-03-27
- [x] En `packages/logger/src/blackbox.ts`: aplicar spread condicional en `handleIngestionStalled` (línea ~170) — completado 2026-03-27
- [ ] Crear `packages/db/src/ioredis-mock.d.ts` con `declare module "ioredis-mock"`
- [x] `cd apps/web && bunx tsc --noEmit` → salida vacía, exit 0 — confirmado 2026-03-27
- [ ] `cd packages/db && bunx tsc --noEmit` → salida vacía, exit 0 (pendiente — ioredis-mock.d.ts falta)
- [x] `export PATH="$HOME/.bun/bin:$PATH" && cd /home/enzo/rag-saldivia && bun run test` → 259 tests pasan — confirmado 2026-03-27
- [x] Commits: `fix(blackbox)` + `fix(collections)` — 2026-03-27

**Estado: parcialmente completado — falta crear `packages/db/src/ioredis-mock.d.ts` para que packages/db pase tsc**

---

## Fase F9.3 — Dead code elimination *(1-2 hs)*

Objetivo: los 8 archivos dead y sus importaciones inexistentes son eliminados. `bunx knip` no reporta ninguno en "Unused files".

### Archivos a eliminar (8 — verificados con grep, cero importaciones externas)

| Archivo | Motivo |
|---|---|
| `apps/web/src/hooks/useCrossdocStream.ts` | Streaming multi-documento — requiere workstation GPU. Versión futura. |
| `apps/web/src/hooks/useCrossdocDecompose.ts` | Descomposición de queries RAG multi-colección. Versión futura. |
| `apps/web/src/components/chat/SplitView.tsx` | Componente para vista crossdoc sin workstation. Versión futura. |
| `apps/web/src/components/collections/CollectionHistory.tsx` | Estadísticas de colección del RAG server. Versión futura. |
| `apps/web/src/lib/auth/next-auth.ts` | Stub de SSO Google/Azure AD. Versión futura. |
| `apps/web/src/lib/safe-action.ts` | Wrapper de `next-safe-action` — cero consumidores. |
| `apps/web/src/lib/form.ts` | Helper de `react-hook-form` — cero consumidores. |
| `scripts/health-check.ts` | Script dead — `/api/health` cumple esa función. |

**Comando para verificar antes de eliminar** (todos deben devolver salida vacía):
```bash
grep -r "useCrossdocStream\|useCrossdocDecompose\|SplitView\|CollectionHistory\|next-auth\|safe-action\|lib/form\|health-check" apps/web/src --include="*.tsx" --include="*.ts" -l | grep -v "Dead code\|plan9"
```
Salida esperada: vacía. Si hay resultados, detener y reportar.

### Server Actions huérfanas — veredictos pre-tomados

Confirmado con grep (`grep -r "actionX" apps/web/src --include="*.tsx" -l | grep -v actions/`):

| Action | Veredicto | Motivo |
|---|---|---|
| `actionListAreas` | **ELIMINAR** | 0 usos fuera de `actions/areas.ts` |
| `actionListSessions` | **ELIMINAR** | 0 usos fuera de `actions/chat.ts` |
| `actionGetSession` | **ELIMINAR** | 0 usos fuera de `actions/chat.ts` |
| `actionGetRagParams` | **ELIMINAR** | 0 usos fuera de `actions/config.ts` |
| `actionResetOnboarding` | **ELIMINAR** | 0 usos fuera de `actions/settings.ts` |
| `actionListUsers` | **ELIMINAR** | 0 usos fuera de `actions/users.ts` |
| `actionAssignArea` | **ELIMINAR** | 0 usos fuera de `actions/users.ts` |
| `actionRemoveArea` | **ELIMINAR** | 0 usos fuera de `actions/users.ts` |
| `actionUpdatePassword` | **MANTENER** | Usado en `src/components/settings/SettingsClient.tsx` |

Para cada action a eliminar: borrar solo la función exportada del archivo. No borrar el archivo completo — los otros exports del mismo archivo pueden estar en uso.

**Exports knip-false-positives — NO tocar:**
- `buttonVariants`, `badgeVariants` — usados via cva() en otros componentes
- Sub-componentes shadcn (`DialogPortal`, `SheetTrigger`, `TableFooter`, etc.) — API pública intencional
- `isAdmin`, `isAreaManager` — usados en `proxy.ts` y `audit/page.tsx`
- `Job` de `queue.ts` — tipo usado en route handlers
- `LOG_LEVEL_PRIORITY` — usado internamente en el logger
- `getCurrentUser` de `current-user.ts` — verificar: si el grep devuelve 0 usos externos, eliminar; si hay usos, mantener

- [ ] Ejecutar el grep de verificación — confirmar salida vacía
- [ ] Eliminar `apps/web/src/hooks/useCrossdocStream.ts`
- [ ] Eliminar `apps/web/src/hooks/useCrossdocDecompose.ts`
- [ ] Eliminar `apps/web/src/components/chat/SplitView.tsx`
- [ ] Eliminar `apps/web/src/components/collections/CollectionHistory.tsx`
- [ ] Eliminar `apps/web/src/lib/auth/next-auth.ts`
- [ ] Eliminar `apps/web/src/lib/safe-action.ts`
- [ ] Eliminar `apps/web/src/lib/form.ts`
- [ ] Eliminar `scripts/health-check.ts`
- [ ] En `apps/web/src/app/actions/areas.ts`: eliminar la función `actionListAreas`
- [ ] En `apps/web/src/app/actions/chat.ts`: eliminar `actionListSessions` y `actionGetSession`
- [ ] En `apps/web/src/app/actions/config.ts`: eliminar `actionGetRagParams`
- [ ] En `apps/web/src/app/actions/settings.ts`: eliminar `actionResetOnboarding`
- [ ] En `apps/web/src/app/actions/users.ts`: eliminar `actionListUsers`, `actionAssignArea`, `actionRemoveArea`
- [ ] Verificar `getCurrentUser`: `grep -r "getCurrentUser" apps/web/src --include="*.tsx" --include="*.ts" -l | grep -v "current-user.ts"` → si vacío, eliminar la función del archivo; si hay resultados, mantener
- [ ] `export PATH="$HOME/.bun/bin:$PATH" && bun run test` → todos los tests pasan
- [ ] `export PATH="$HOME/.bun/bin:$PATH" && cd apps/web && bun run test:components` → 154 tests pasan
- [ ] Commit: `refactor(web): eliminar dead code — crossdoc hooks, sso stub, wrappers, actions huerfanas — plan9 f9.3`

**Estado: pendiente**

---

## Fase F9.4 — Dependencias limpias *(30-45 min)*

Objetivo: `bunx knip` no reporta `next-safe-action`, `d3` ni `@types/d3` como unused dependencies.

**Verificación de `@tanstack/react-table` antes de tocar nada:**
```bash
grep -r "react-table" apps/web/src --include="*.tsx" --include="*.ts" | head -5
```
Si devuelve resultados → **mantener** `@tanstack/react-table`. Si devuelve vacío → eliminar también.

**Dependencias a remover (confirmadas sin uso):**

```bash
cd apps/web
export PATH="$HOME/.bun/bin:$PATH"
bun remove next-safe-action d3
bun remove --dev @types/d3
```

Salida esperada de cada comando: `bun remove` sin errores, lista de paquetes removidos.

**Dependencias que parecen unused pero NO tocar:**
- `@radix-ui/react-checkbox`, `@radix-ui/react-label`, `@radix-ui/react-scroll-area`, `@radix-ui/react-select`, `@radix-ui/react-switch` — dependencias transitivas de shadcn/ui. Eliminarlas puede romper componentes de UI silenciosamente.
- `@happy-dom/global-registrator`, `@testing-library/react`, `@testing-library/user-event`, `happy-dom` — se usan vía `--preload` en CLI, no via import. Knip no puede detectar este patrón.
- `@vitejs/plugin-react` — usado por Storybook (knip ignora `.storybook/`).

- [ ] Ejecutar `grep -r "react-table" apps/web/src --include="*.tsx" --include="*.ts" | head -5` — si hay resultados, anotar que `@tanstack/react-table` se mantiene
- [ ] `cd apps/web && bun remove next-safe-action d3`
- [ ] `cd apps/web && bun remove --dev @types/d3`
- [ ] `export PATH="$HOME/.bun/bin:$PATH" && bun run test` → todos los tests pasan
- [ ] `cd apps/web && bunx tsc --noEmit` → exit 0 (confirmar que remover d3 no rompió types)
- [ ] Commit: `chore(deps): remover next-safe-action, d3 — dependencias sin uso — plan9 f9.4`

**Estado: pendiente**

---

## Fase F9.5 — ESLint config + fix *(1-2 hs)*

Objetivo: `cd apps/web && bunx eslint src --max-warnings 0` devuelve exit 0.

**Situación actual:** `eslint` y `eslint-config-next` están instalados pero no hay `eslint.config.js`. El linter nunca corrió.

**Archivo a crear: `apps/web/eslint.config.js`**

Contenido exacto:

```javascript
import { dirname } from "path"
import { fileURLToPath } from "url"
import { FlatCompat } from "@eslint/eslintrc"

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

const compat = new FlatCompat({ baseDirectory: __dirname })

const eslintConfig = [
  ...compat.extends("next/core-web-vitals", "next/typescript"),
  {
    rules: {
      "no-console": ["warn", { allow: ["warn", "error"] }],
      "@typescript-eslint/no-explicit-any": "warn",
      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
      ],
      "react-hooks/exhaustive-deps": "warn",
    },
  },
]

export default eslintConfig
```

**Instalar `@eslint/eslintrc`** (necesario para `FlatCompat`):
```bash
cd apps/web && bun add --dev @eslint/eslintrc
```

**Proceso de fix:**
1. Crear `eslint.config.js` con el contenido de arriba
2. Correr `bunx eslint src --fix` para auto-corregir lo que se pueda
3. Correr `bunx eslint src` (sin `--fix`) para ver los warnings/errors restantes
4. Resolver manualmente los que no se auto-corrigieron (principalmente `no-console` y `@typescript-eslint/no-explicit-any`)
5. Para los `no-console`: reemplazar `console.log(...)` por el logger estructurado (`log.info(...)`) o eliminar si era debugging
6. Para los `no-explicit-any` que son legítimamente necesarios: agregar `// eslint-disable-next-line @typescript-eslint/no-explicit-any` con un comentario explicando por qué

**Agregar script `lint:eslint` en `apps/web/package.json`** (el script `lint` ya existe y hace `tsc --noEmit` — no pisar):

En `apps/web/package.json`, en la sección `scripts`, agregar:
```json
"lint:eslint": "eslint src --max-warnings 0"
```

**El `turbo.json` ya tiene el task `lint` definido** — no necesita cambios.

- [ ] `cd apps/web && bun add --dev @eslint/eslintrc`
- [ ] Crear `apps/web/eslint.config.js` con el contenido exacto de arriba
- [ ] `cd apps/web && bunx eslint src --fix` — dejar que auto-corrija
- [ ] `cd apps/web && bunx eslint src` — ver los issues restantes
- [ ] Resolver todos los `no-console` (reemplazar con logger o eliminar)
- [ ] Resolver todos los `@typescript-eslint/no-unused-vars` restantes
- [ ] Para los `no-explicit-any` inevitables: agregar `// eslint-disable-next-line` con comentario
- [ ] Agregar `"lint:eslint": "eslint src --max-warnings 0"` en `apps/web/package.json` scripts
- [ ] `cd apps/web && bunx eslint src --max-warnings 0` → exit 0 (confirmar)
- [ ] `export PATH="$HOME/.bun/bin:$PATH" && bun run test` → todos los tests pasan
- [ ] Commit: `feat(lint): crear eslint.config.js y resolver todos los warnings — plan9 f9.5`

**Estado: pendiente**

---

## Fase F9.6 — Husky + commitlint + lint-staged *(1 hs)*

Objetivo: `git commit -m "bad message"` falla. `git commit -m "feat: ok"` con código limpio pasa.

**Instalar dependencias (en el root del monorepo):**
```bash
export PATH="$HOME/.bun/bin:$PATH"
cd /home/enzo/rag-saldivia
bun add --dev husky @commitlint/cli @commitlint/config-conventional lint-staged
```

**Inicializar husky:**
```bash
bunx husky init
```
Esto crea `.husky/pre-commit` con un script de placeholder. Reemplazar su contenido.

**Archivo `.husky/pre-commit` — contenido exacto:**
```sh
#!/bin/sh
export PATH="$HOME/.bun/bin:$PATH"
bunx lint-staged
```

**Archivo `.husky/commit-msg` — crear con este contenido exacto:**
```sh
#!/bin/sh
export PATH="$HOME/.bun/bin:$PATH"
bunx commitlint --edit "$1"
```

```bash
chmod +x .husky/commit-msg
```

**Archivo `commitlint.config.js` — crear en el root con este contenido exacto:**
```javascript
export default {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "type-enum": [
      2,
      "always",
      ["feat", "fix", "refactor", "chore", "docs", "test", "ci", "perf", "style"],
    ],
    "subject-max-length": [2, "always", 100],
  },
}
```

**Archivo `.lintstagedrc.js` — crear en el root con este contenido exacto:**
```javascript
export default {
  "apps/web/src/**/*.{ts,tsx}": ["bunx eslint --fix --max-warnings 0"],
  "packages/**/*.ts": ["bunx eslint --fix --max-warnings 0"],
}
```

**Agregar `"prepare": "husky"` en el `package.json` del root** (si no existe ya):
En la sección `scripts` del `package.json` raíz, agregar o verificar que existe:
```json
"prepare": "husky"
```

**Test de los hooks:**
```bash
# Debe FALLAR:
git commit -m "bad message"
# Salida esperada: "⧗   input: bad message" + error de commitlint

# Debe PASAR (con código limpio):
git commit --allow-empty -m "test: verificar hooks de commitlint"
# Luego hacer: git reset HEAD~1  (para deshacer el commit vacío de prueba)
```

- [ ] `bun add --dev husky @commitlint/cli @commitlint/config-conventional lint-staged`
- [ ] `bunx husky init`
- [ ] Reemplazar contenido de `.husky/pre-commit` con el script de arriba
- [ ] Crear `.husky/commit-msg` con el script de arriba
- [ ] `chmod +x .husky/commit-msg`
- [ ] Crear `commitlint.config.js` en el root con el contenido de arriba
- [ ] Crear `.lintstagedrc.js` en el root con el contenido de arriba
- [ ] Verificar que `"prepare": "husky"` existe en el `package.json` raíz — si no, agregarlo
- [ ] Verificar que `.husky/` está en `.gitignore` de manera que NO se ignore (los hooks deben commitearse)
- [ ] Test: `git commit -m "bad message"` → debe FALLAR con error de commitlint
- [ ] `git reset HEAD~1` si el test anterior accidentalmente pasó
- [ ] `export PATH="$HOME/.bun/bin:$PATH" && bun run test` → todos los tests pasan
- [ ] Commit: `chore(git): husky + commitlint + lint-staged — enforcement en punto de commit — plan9 f9.6`

**Estado: pendiente**

---

## Fase F9.7 — Auditoría console.log y TODOs *(30-45 min)*

Objetivo: cero `console.log` en código de producción. Cero `TODO`/`FIXME`/`HACK` en código de producción.

**Buscar console.log en producción:**
```bash
grep -rn "console\.log" apps/web/src packages/*/src --include="*.ts" --include="*.tsx" | grep -v "__tests__\|test\."
```
Para cada resultado: reemplazar con el logger estructurado (`log.info(...)`, `log.debug(...)`, etc.) o eliminar si era debugging temporal.

**`console.warn` y `console.error` son aceptables** (eslint los permite). No tocarlos.

**Buscar TODOs/FIXMEs:**
```bash
grep -rn "TODO\|FIXME\|HACK\|XXX" apps/web/src packages/*/src --include="*.ts" --include="*.tsx" | grep -v "__tests__\|test\."
```
Para cada resultado: resolverlo inline, crear un issue de GitHub, o eliminar si ya no es relevante. No es aceptable dejarlo como está.

**Buscar código comentado** (bloques de código entre `//` consecutivos):
```bash
grep -rn "^[[:space:]]*//" apps/web/src --include="*.tsx" --include="*.ts" | grep -v "eslint-disable\| @\|import\|http\|https\|Copyright\|License" | head -30
```
Para cada bloque de código comentado: eliminar (si está comentado, no hace falta; si hace falta, tiene que estar activo).

- [ ] Ejecutar grep de `console.log` — registrar todos los resultados
- [ ] Reemplazar cada `console.log` de producción con el logger o eliminar
- [ ] Ejecutar grep de TODOs/FIXMEs — registrar todos los resultados
- [ ] Resolver o crear issue para cada TODO/FIXME
- [ ] Ejecutar grep de código comentado — evaluar y eliminar bloques innecesarios
- [ ] `export PATH="$HOME/.bun/bin:$PATH" && bun run test` → todos los tests pasan
- [ ] Commit: `chore(code): eliminar console.log, resolver todos los todos — plan9 f9.7`

**Estado: pendiente**

---

## Fase F9.8 — Knip config limpia *(30 min)*

Objetivo: `bunx knip` devuelve exit 0.

**Cambios exactos al `knip.json`:**

El `knip.json` actual tiene entry patterns redundantes en los packages (knip los infiere automáticamente). Reemplazar el archivo completo con esta versión limpia:

```json
{
  "workspaces": {
    "apps/web": {
      "entry": [
        "src/app/**/{page,layout,route}.ts{x,}",
        "src/workers/*.ts"
      ]
    },
    "apps/cli": {
      "entry": ["src/commands/*.ts"]
    }
  },
  "ignore": [
    "**/__tests__/**",
    "**/*.test.*",
    "apps/web/src/lib/test-setup.ts",
    "apps/web/src/lib/component-test-setup.ts",
    "apps/web/stories/**",
    "apps/web/.storybook/**",
    "apps/web/tests/**"
  ],
  "ignoreDependencies": [
    "@radix-ui/react-checkbox",
    "@radix-ui/react-label",
    "@radix-ui/react-scroll-area",
    "@radix-ui/react-select",
    "@radix-ui/react-switch",
    "@happy-dom/global-registrator",
    "@testing-library/react",
    "@testing-library/user-event",
    "happy-dom",
    "@vitejs/plugin-react"
  ]
}
```

**Explicación de los cambios:**
- Removidos `src/index.ts` de packages/* (knip los infiere del `main` en `package.json`)
- Removido `src/middleware.ts` de apps/web (knip lo infiere como entry de Next.js)
- Removida entrada `**/*.spec.*` del ignore (redundante)
- Agregado `ignoreDependencies` para los falsos positivos de Radix y testing tools
- Agregado `apps/web/tests/**` al ignore para que knip no procese los helpers de test

**Verificación:**
```bash
export PATH="$HOME/.bun/bin:$PATH"
bunx knip
```
Salida esperada: sin output, exit 0.

Si hay items restantes en la salida: investigar uno a uno. Si son falsos positivos legítimos, agregar al `ignoreDependencies`. Si son dead code real, eliminarlo.

- [ ] Reemplazar `knip.json` con el contenido exacto de arriba
- [ ] `bunx knip` → salida vacía, exit 0 (confirmar)
- [ ] Si hay items residuales, investigar y resolver cada uno antes de continuar
- [ ] `export PATH="$HOME/.bun/bin:$PATH" && bun run test` → todos los tests pasan
- [ ] Commit: `chore(knip): limpiar configuracion — ignorar falsos positivos documentados — plan9 f9.8`

**Estado: pendiente**

---

## Criterio de done global del Plan 9

Ejecutar todos estos comandos en secuencia. Todos deben pasar:

```bash
export PATH="$HOME/.bun/bin:$PATH"
cd /home/enzo/rag-saldivia

# 1. Git limpio
git ls-files | grep -E "(playwright-mcp|superpowers|\.log$|config/\.env)"
# Esperado: vacío

# 2. TypeScript limpio
cd apps/web && bunx tsc --noEmit
# Esperado: sin output, exit 0
cd /home/enzo/rag-saldivia/packages/db && bunx tsc --noEmit
# Esperado: sin output, exit 0

# 3. Knip limpio
cd /home/enzo/rag-saldivia && bunx knip
# Esperado: sin output, exit 0

# 4. ESLint limpio
cd apps/web && bunx eslint src --max-warnings 0
# Esperado: sin output, exit 0

# 5. Tests pasan
cd /home/enzo/rag-saldivia && bun run test
# Esperado: todos pasan
bun run test:components
# Esperado: 154 pass
```

### Checklist de cierre

- [ ] `git ls-files | grep -E "(playwright-mcp|superpowers)"` → vacío
- [ ] `bunx tsc --noEmit` (web + db) → exit 0
- [ ] `bunx knip` → exit 0
- [ ] `bunx eslint src --max-warnings 0` → exit 0
- [ ] `bun run test` → todos pasan
- [ ] `bun run test:components` → 154 pasan
- [ ] `git commit -m "bad"` → rechazado por commitlint
- [ ] CHANGELOG.md actualizado con entrada del Plan 9
- [ ] Commit final: `chore: plan9 completado — repo limpio, ts perfecto, linting, hooks`
- [ ] `git push` — confirmar repo remoto limpio

**Estado: pendiente**

---

## Mapa de cambios

| Cambio | Fase | Archivos afectados |
|---|---|---|
| 46 archivos purgados del repo | F9.1 | `.gitignore` + `git rm --cached` |
| `updateTag` → `revalidateTag` | F9.2 | `actions/collections.ts` |
| `IngestionEventRecord` spread condicional | F9.2 | `packages/logger/src/blackbox.ts` |
| `ioredis-mock` types declarados | F9.2 | `packages/db/src/ioredis-mock.d.ts` (nuevo) |
| 8 archivos dead eliminados | F9.3 | Los 8 listados |
| 8 server actions huérfanas eliminadas | F9.3 | `apps/web/src/app/actions/*.ts` |
| `next-safe-action`, `d3`, `@types/d3` removidos | F9.4 | `apps/web/package.json` |
| `eslint.config.js` creado + `@eslint/eslintrc` | F9.5 | `apps/web/eslint.config.js` (nuevo) |
| `console.log` eliminados de producción | F9.5 + F9.7 | Varios |
| Husky + commitlint + lint-staged | F9.6 | `.husky/`, `commitlint.config.js`, `.lintstagedrc.js` (nuevos) |
| TODOs/FIXMEs resueltos | F9.7 | Varios |
| `knip.json` simplificado con `ignoreDependencies` | F9.8 | `knip.json` |

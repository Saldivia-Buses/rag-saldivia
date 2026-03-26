# Workflows — RAG Saldivia

> Branch: `experimental/ultra-optimize`
> Última actualización: 2026-03-25

Este documento describe los flujos de trabajo que usamos en el proyecto. Es la referencia para cómo desarrollar, testear, hacer commits, planificar features y deployar.

---

## 1. Workflow de desarrollo local

### Arrancar el entorno

```bash
# Asegurarse de tener el .env correcto
cat apps/web/.env.local   # verificar JWT_SECRET, SYSTEM_API_KEY, MOCK_RAG

# Arrancar el servidor
bun run dev
# → http://localhost:3000
# → admin@localhost / changeme

# CLI (si no está instalada globalmente)
cd apps/cli && bun link
```

### Ciclo de desarrollo de una feature

1. **Leer contexto**: antes de modificar cualquier archivo, leer los archivos críticos relevantes
2. **Modo mock**: usar `MOCK_RAG=true` para desarrollar sin Docker
3. **Type-check continuo**: el pre-push hook corre type-check automáticamente
4. **Lint**: `bun run lint` (solo en `apps/web` por ahora)

### Archivos críticos — leer antes de modificar

| Archivo | Por qué leerlo primero |
|---------|------------------------|
| `apps/web/src/middleware.ts` | Controla toda la auth y RBAC; afecta cada request |
| `apps/web/src/lib/auth/jwt.ts` | Cómo se leen los claims; tiene lógica de headers vs cookies |
| `packages/db/src/schema.ts` | Schema de las 12 tablas; cambios requieren migración |
| `packages/shared/src/schemas.ts` | Zod schemas compartidos; cambiar aquí afecta web y cli |
| `packages/logger/src/backend.ts` | Logger usa import estático de `@rag-saldivia/db` (no dinámico) |

---

## 2. Workflow de testing

### Cuándo correr tests

- **Antes de hacer commit**: tests del área que modificaste
- **Antes de hacer push**: `bun run test` completo (también lo corre el hook pre-push vía type-check)
- **Al agregar funcionalidad nueva**: crear el test primero (o inmediatamente después)

### Comandos

```bash
# Suite completa — todos los workspaces con Turborepo
bun run test

# Por paquete — más rápido durante desarrollo
bun test apps/web/src/lib/auth/__tests__/      # auth + RBAC (17 tests)
bun test packages/db/src/__tests__/            # DB queries (24 tests)
bun test packages/logger/src/__tests__/        # logger + blackbox (24 tests)
bun test packages/config/src/__tests__/        # config loader (14 tests)

# Total: 79 tests unitarios
```

### Dónde agregar tests nuevos

| Qué testear | Dónde poner el archivo |
|-------------|------------------------|
| Queries de DB nuevas | `packages/db/src/__tests__/` |
| Lógica de auth / RBAC | `apps/web/src/lib/auth/__tests__/` |
| Config loader | `packages/config/src/__tests__/` |
| Logger / blackbox | `packages/logger/src/__tests__/` |
| Hooks React | `apps/web/src/hooks/__tests__/` (crear si no existe) |

### Cómo escribir tests (patrón del proyecto)

```typescript
import { describe, test, expect, beforeEach } from "bun:test";

// Para tests de DB: usar instancia en memoria
import { createClient } from "@libsql/client";
import { drizzle } from "drizzle-orm/libsql";
import * as schema from "../schema";

let db: ReturnType<typeof drizzle>;

beforeEach(async () => {
  const client = createClient({ url: ":memory:" });
  db = drizzle(client, { schema });
  // inicializar schema con SQL directo...
});

describe("nombre del módulo", () => {
  test("describe qué debería pasar", async () => {
    // arrange → act → assert
  });
});
```

**Reglas:**
- Siempre usar DB en memoria (`:memory:`), nunca el archivo real
- Todos los `await import(...)` al nivel del módulo, no dentro de callbacks
- Los tests de logger: verificar que el output contiene el tipo de evento, no asumir formato JSON/pretty

### Cobertura actual y huecos

| Área | Tests | Huecos conocidos |
|------|-------|------------------|
| Auth + RBAC | ✅ 17 tests | — |
| DB queries (users, areas) | ✅ 24 tests | sessions, events (sin tests) |
| Logger + blackbox | ✅ 24 tests | — |
| Config loader | ✅ 14 tests | — |
| API routes | ❌ sin tests unitarios | verificadas manualmente en Plan 2 |
| Middleware | ❌ sin tests unitarios | edge runtime, difícil testear |
| Hooks React | ❌ sin tests | `useRagStream` candidato prioritario |
| Worker de ingesta | ❌ sin tests | lógica de locking SQLite |

---

## 3. Workflow de Git

### Conventional Commits

Todos los commits deben seguir el formato:
```
<tipo>(<scope>): <descripción en minúsculas>
```

**Tipos permitidos:** `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `perf`, `ci`

**Scopes del proyecto:**
```
web        → apps/web
cli        → apps/cli
db         → packages/db
shared     → packages/shared
config     → packages/config
logger     → packages/logger
docs       → cualquier archivo en docs/
scripts    → scripts/
ci         → .github/workflows/
deps       → dependencias (package.json)
```

**Ejemplos correctos:**
```bash
git commit -m "feat(web): agregar paginación en /admin/users"
git commit -m "fix(db): corregir removeAreaCollection para filtrar por collectionName"
git commit -m "test(logger): agregar tests de reconstructFromEvents con eventos de ingesta"
git commit -m "docs: actualizar architecture.md con diagrama de auth service-to-service"
git commit -m "refactor(web): extraer lógica SSE de ChatInterface a useRagStream"
```

El hook `commit-msg` de Husky rechaza automáticamente commits que no sigan este formato.

### Actualizar CHANGELOG.md

**Regla:** cada tarea completada genera una entrada en `CHANGELOG.md` **antes** de hacer commit.

Las entradas se organizan en secciones por plan dentro de `[Unreleased]`. Esto hace navegable el historial durante el desarrollo sin cambiar la estructura de release (cuando se publica, todo `[Unreleased]` se mueve a `[vX.Y.Z]`).

```markdown
## [Unreleased]

### Plan 4 — Product Roadmap (2026-03-25)

#### Added
- `apps/web/src/components/...`: descripción concisa de qué se agregó — YYYY-MM-DD *(Plan 4 F1.7)*

#### Fixed
- `paquete/archivo.ts`: descripción del bug y cómo se corrigió — YYYY-MM-DD

#### Changed
- `paquete/archivo.ts`: qué cambió y por qué — YYYY-MM-DD

### Plan 3 — Bugfix (2026-03-25)

#### Fixed
- ...
```

**Reglas:**
- Cada plan nuevo abre su propia sección `### Plan N — Nombre (YYYY-MM-DD)` al tope de `[Unreleased]`
- Las categorías dentro de cada sección son `#### Added` | `#### Changed` | `#### Fixed` | `#### Security` | `#### Deprecated` | `#### Removed`
- El ID de tarea del plan va al final de la línea: `*(Plan 4 F1.7)*`
- Las entradas del mismo plan se agrupan, nunca se intercalan con las de otro plan

### Crear un PR

1. Asegurarse de que el CHANGELOG tiene la entrada correspondiente
2. Hacer push de la branch
3. Crear PR con el template (sección CHANGELOG obligatoria)
4. El CI valida: commitlint + changelog check + type-check + tests + lint

### Pre-push hook

El hook `pre-push` corre `type-check` antes de cada push. Si falla, el push se cancela.
Bun está en `~/.bun/bin/bun` — el hook usa `$(which bun || echo /home/enzo/.bun/bin/bun)`.

---

## 4. Workflow de planificación

### Cuándo crear un plan

Crear un plan nuevo en `docs/plans/` cuando:
- La feature requiere más de 2-3 horas de trabajo
- Hay múltiples fases interdependientes
- Es importante rastrear el progreso con fechas

### Formato de un plan

```markdown
# Plan: [Nombre descriptivo]

> Documento en docs/plans/[nombre].md
> Se actualiza a medida que se completan tareas.

## Contexto
Qué existe hoy y por qué se hace este trabajo.

## Seguimiento
Formato: `- [ ] Descripción — estimación`
Al completar: `- [x] Descripción — completado YYYY-MM-DD`

## Fase 1 — Nombre *(estimación)*

Objetivo: una oración sobre qué debe quedar funcionando al terminar esta fase.

- [ ] Tarea A — 30 min
- [ ] Tarea B — 1 hs

Criterio de done: condición objetiva y verificable.

### Checklist de cierre
- [ ] `bun run test` — todos pasan
- [ ] CHANGELOG.md actualizado bajo `[Unreleased] > ### Plan N — Nombre`
- [ ] `git commit -m "feat(scope): ..."` hecho

**Estado: pendiente**

---

## Estado general

| Fase | Estado | Fecha |
|------|--------|-------|
| Fase 1 — Nombre | ⏳ pendiente | — |
```

### Convención de nombres

```
docs/plans/[tema]-plan[N]-[descripcion].md
```

Ejemplos:
- `ultra-optimize-plan1-birth.md`
- `ultra-optimize-plan2-testing.md`
- `ultra-optimize-plan4-e2e-playwright.md`

### Planes actuales (todos completados)

| Plan | Estado | Descripción |
|------|--------|-------------|
| [plan1-birth](./plans/ultra-optimize-plan1-birth.md) | ✅ completado 2026-03-24 | Construcción del monorepo TypeScript desde cero |
| [plan2-testing](./plans/ultra-optimize-plan2-testing.md) | ✅ completado 2026-03-25 | Testing sistemático en 7 fases, 15 bugs encontrados |
| [plan3-bugfix](./plans/ultra-optimize-plan3-bugfix.md) | ✅ completado 2026-03-25 | Bugfix + refactor de complejidad (CodeGraphContext) |

---

## 5. Workflow de features nuevas

El ciclo completo para agregar una feature:

```
1. Leer contexto → archivos críticos relevantes
2. Crear plan (si es complejo) → docs/plans/
3. Implementar → Server Component / Client Component / Server Action / API route
4. Agregar tests unitarios → bun test <ruta>
5. Actualizar CHANGELOG.md → [Unreleased] / Added o Changed
6. Commit → feat(scope): descripción
7. Type-check → bun run type-check (automático en pre-push)
```

### Dónde poner cada cosa

| Qué | Dónde |
|-----|-------|
| Página nueva | `apps/web/src/app/(app)/[ruta]/page.tsx` (Server Component) |
| Mutación desde el servidor | `apps/web/src/app/actions/[dominio].ts` (Server Action) |
| Endpoint REST | `apps/web/src/app/api/[ruta]/route.ts` (Route Handler) |
| Lógica de UI con estado | `apps/web/src/components/[nombre].tsx` (Client Component) |
| Hook con lógica de fetch | `apps/web/src/hooks/use[Nombre].ts` |
| Query de DB nueva | `packages/db/src/queries/[dominio].ts` |
| Tipo compartido (web + cli) | `packages/shared/src/schemas.ts` |
| Comando CLI nuevo | `apps/cli/src/commands/[nombre].ts` + registrar en `index.ts` |

### Reglas de arquitectura

- **Server Components por defecto** — solo usar `"use client"` donde sea imprescindible (chat SSE, sliders, modales con estado)
- **Validar con Zod** — usar schemas de `@rag-saldivia/shared` para inputs de API
- **Loggear con el black box** — todo punto crítico emite un evento al logger backend
- **Timestamps siempre en epoch ms** — `Temporal.Now.instant().epochMilliseconds`
- **Nunca `Bun.*` en código que corre en Next.js** — usar `fs/promises`, `crypto` de Node

---

## 6. Workflow de deploy

### Deploy de producción (branch main — stack Python/SvelteKit)

El stack en producción activo en la workstation es el de `main`:

```bash
git checkout main
make deploy PROFILE=workstation-1gpu
```

### Deploy del nuevo stack (cuando esté listo)

El proceso aún no está definido en detalle. Cuando `experimental/ultra-optimize` esté listo para producción:

1. Merge a `main` vía PR
2. Tag `v1.0.0` en GitHub
3. El workflow `.github/workflows/release.yml` mueve `[Unreleased]` → `[v1.0.0]`
4. El workflow `.github/workflows/deploy.yml` se dispara con el tag

### Health check

```bash
rag status                          # semáforo de todos los servicios
curl http://localhost:3000/api/health  # verificación directa
```

---

## 7. Decisiones de arquitectura (ADRs)

Las decisiones técnicas importantes se documentan en `docs/decisions/` como Architecture Decision Records.

### Cuándo crear un ADR

Crear un ADR cuando:
- Se elige una tecnología sobre otra con trade-offs no obvios (ej. `@libsql` vs `better-sqlite3`)
- Se establece una convención que se desvía del default del ecosistema (ej. CJS sobre ESM)
- Se toma una decisión de arquitectura que afecta múltiples partes del sistema
- Se resuelve un bug que revela un patrón a seguir o evitar en el futuro

**No** crear un ADR para: decisiones triviales, preferencias de estilo, cosas que ya documenta el CHANGELOG.

### Formato

Usar el template en `docs/decisions/000-template.md`:

```
# ADR-NNN: [Título]
Fecha, Estado (Propuesto | Aceptado | Deprecado | Reemplazado por ADR-XXX)
## Contexto — el problema y las restricciones
## Opciones consideradas — con pros/cons
## Decisión — cuál y por qué
## Consecuencias — positivas y trade-offs
## Referencias — archivos o entradas de CHANGELOG relevantes
```

### Convención de nombres

```
docs/decisions/NNN-kebab-case-titulo.md
```

El número es secuencial. Nunca reusar un número aunque se deprece el ADR.

### ADRs existentes

| ADR | Título | Estado |
|-----|--------|--------|
| [001](./decisions/001-libsql-over-better-sqlite3.md) | @libsql/client en lugar de better-sqlite3 | Aceptado |
| [002](./decisions/002-cjs-over-esm.md) | CJS sobre ESM en paquetes del monorepo | Aceptado |
| [003](./decisions/003-nextjs-single-process.md) | Next.js como proceso único (reemplaza Python gateway + SvelteKit) | Aceptado |
| [004](./decisions/004-temporal-api-timestamps.md) | Temporal API para timestamps | Aceptado |
| [005](./decisions/005-static-import-logger-db.md) | Import estático de @rag-saldivia/db en el logger | Aceptado |

---

## 8. Workflow de debugging con Black Box

Cuando algo falla en producción:

```bash
# Ver últimos eventos del sistema
rag audit log --limit 50

# Reconstruir timeline de una fecha específica
rag audit replay --from 2026-03-25

# Filtrar por tipo de evento
rag audit log --type rag.error
rag audit log --type auth.login

# Exportar todos los eventos para análisis
rag audit export > events-$(date +%Y%m%d).json
```

Ver [docs/blackbox.md](./blackbox.md) para el formato completo de eventos y guía de replay.


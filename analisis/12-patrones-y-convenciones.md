# 12 — Patrones y Convenciones

## Convenciones de naming

### Archivos

| Tipo | Convencion | Ejemplo |
|------|-----------|---------|
| Componentes React | PascalCase.tsx | `ChatInterface.tsx` |
| Pages/routes | page.tsx, layout.tsx | `app/(app)/chat/page.tsx` |
| Hooks | camelCase.ts con `use` | `usePresence.ts` |
| Lib/utils | kebab-case.ts | `lib/rag/client.ts` |
| Tests | mismo nombre + `.test.ts(x)` | `ChatInterface.test.tsx` |
| CSS | kebab-case.css | `globals.css` |
| Plans | `1.0.x-plan{N}-{slug}.md` | `1.0.x-plan14-ai-sdk.md` |
| ADRs | `{NNN}-{slug}.md` | `012-stack-definitivo.md` |

### Codigo

| Tipo | Convencion | Ejemplo |
|------|-----------|---------|
| Componentes React | PascalCase | `function ChatInterface()` |
| Funciones | camelCase | `getUserById()` |
| Variables | camelCase | `const sessionList` |
| Constantes | UPPER_SNAKE_CASE | `MAX_RETRIES` |
| Tipos/Interfaces | PascalCase | `type UserRole` |
| Zod schemas | PascalCase + Schema | `RagParamsSchema` |
| Env vars | UPPER_SNAKE_CASE | `JWT_SECRET` |

### Git

| Tipo | Convencion | Ejemplo |
|------|-----------|---------|
| Branches | `plan{N}-{slug}` o directo en `1.0.x` | `plan14-ai-sdk` |
| Commits | `tipo(scope): desc — planN fN` | `feat(chat): integrate ai sdk — plan14 f1` |
| Tags | `v{MAJOR}.{MINOR}.{PATCH}` | `v1.1.0` |

### Tipos de commit

| Tipo | Cuando |
|------|--------|
| `feat` | Nueva funcionalidad |
| `fix` | Correccion de bug |
| `refactor` | Cambio sin cambio de comportamiento |
| `style` | CSS/UI sin cambio de logica |
| `test` | Agregar o modificar tests |
| `docs` | Documentacion |
| `chore` | Mantenimiento (deps, config, CI) |
| `ci` | Cambios de CI/CD |

**Scopes:** web, db, shared, config, logger, ui, chat, auth, rag, agents, plans, deps, messaging, admin, setup

**Reglas:** ingles, minuscula despues del tipo, sin punto final, max 72 chars, siempre referenciar plan y fase.

---

## Patrones de codigo

### Next.js y React

1. **Server Components por defecto** — cero JS al browser salvo donde sea necesario (ADR-009)
2. **`"use client"` solo donde se necesita** — estado, efectos, APIs de browser
3. **`next/font/google`** para Instrument Sans — self-hosting automatico
4. **`next/dynamic` con `ssr: false`** solo en Client Components
5. **Cache con `unstable_cache`** para llamadas al RAG con tags

### Streaming

1. **Verificar status HTTP ANTES de streamear** — patron critico heredado del bug del gateway Python
2. **`ai-stream.ts` transforma NVIDIA SSE al protocolo AI SDK** — nunca consumir SSE raw en el cliente
3. **Citations como `data-sources` custom parts** — no como JSON en el texto
4. **`useChat` de `@ai-sdk/react`** en ChatInterface — no hooks custom

### Base de datos

1. **`Date.now()` para timestamps** — nunca `_ts()` de SQLite (ADR-004)
2. **CJS en packages/** — sin `"type": "module"` (ADR-002)
3. **Funciones reales en tests** — no mocks de queries (ADR-007)
4. **`getRedisClient()` nunca retorna null** — lanza error (ADR-010)
5. **No importar logger en redis.ts** — dependencia circular (ADR-005)
6. **Import estatico de db en logger** — dinamico falla en webpack (ADR-005)

### Server Actions

1. **Todas usan `authAction` o `adminAction`** de `lib/safe-action.ts`
2. **Input validado con Zod** via `.schema(z.object({...}))`
3. **Retorno wrapped:** callers acceden a `result?.data`
4. **`clean()` helper** para bridge Zod optional → `exactOptionalPropertyTypes`
5. **`revalidatePath()`** despues de mutaciones si corresponde

### CSS y Design System

1. **Tokens CSS siempre** — nunca hardcodear colores
2. **`var(--accent)`, `text-fg-muted`, `bg-surface`** — utility classes con tokens
3. **`@theme inline`** en Tailwind v4 — critico para dark mode class-based
4. **`bg-surface` vs `bg-bg`** — surface para cards/paneles, bg para fondo base
5. **`postcss.config.js` con `@tailwindcss/postcss`** — sin el, utilities custom no se generan

### Testing

1. **`afterEach(cleanup)` por archivo** — obligatorio en component tests
2. **Queries escopadas** — `const { getByRole } = render(...)`, no `screen.getByRole`
3. **`fireEvent` sobre `userEvent`** — happy-dom compatibility
4. **Preloads separados** — `component-test-setup.ts` (happy-dom) vs `test-setup.ts` (mocks)
5. **DB real en tests** — no mocks, funciones reales contra SQLite en memoria

### Seguridad

1. **JWT con jti unico** — requerido para revocacion
2. **Revocacion en extractClaims(), NO en proxy.ts** — Edge no tiene ioredis
3. **Cookie HttpOnly + SameSite=Lax** — no accesible desde JS, proteccion CSRF basica
4. **SYSTEM_API_KEY para S2S** — no requiere JWT
5. **Nunca `--no-verify`, `chmod 777`, `--force` push** — guard rules

---

## Patron de archivos criticos

Antes de modificar estos archivos, leer el codigo Y entender el impacto:

| Archivo | Impacto de un cambio |
|---------|---------------------|
| `proxy.ts` | Rompe auth para TODA la app |
| `lib/auth/jwt.ts` | Rompe login/logout/refresh |
| `lib/safe-action.ts` | Rompe TODAS las server actions |
| `globals.css` | Cambia el look de TODA la app |
| `schema/core.ts` | Requiere migracion de DB |
| `component-test-setup.ts` | Puede romper 158 tests |

---

## Workflow de desarrollo

### Inicio de sesion
```
1. Opus lee MEMORY.md + CLAUDE.md (automatico)
2. Enzo dice que quiere hacer
3. Opus confirma y arranca
```

### Ejecucion de un plan
```
Para cada fase:
  1. Opus implementa los cambios
  2. Quality gates: tsc → test → lint
  3. Si falla → fix forward
  4. Si scope drift → parar y reportar
  5. Reporte: archivos, tests, estado
  6. Enzo: "dale" / "mostra diff" / "para" / "volve atras"
```

### Escalacion a Enzo
**Opus escala cuando:**
- Scope drift (archivos fuera del plan)
- > 3 fixes no planeados
- Decision de producto
- Algo podria romper funcionalidad
- Cambio de dependencia
- Guard rules

**Opus NO escala para:**
- Cambios dentro del scope
- Fixes de tests del mismo cambio
- Ajustes de imports, formatting

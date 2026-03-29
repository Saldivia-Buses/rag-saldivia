# Saldivia RAG

## LEER PRIMERO — antes de cualquier trabajo

1. `docs/bible.md` — reglas permanentes (workflow, stack, protocolos, naming)
2. `docs/plans/1.0.x-plan-maestro.md` — roadmap y planes de la versión actual

No empezar a trabajar sin leer estos documentos.

---

## Qué es este proyecto

Overlay sobre **NVIDIA RAG Blueprint v2.5.0** que agrega autenticación, RBAC,
multi-colección, y frontend Next.js 16. Deploy en workstation física con 1x RTX
PRO 6000 Blackwell (96GB VRAM).

- **Repo:** `~/rag-saldivia/` — branch activa: `1.0.x`
- **Remoto:** https://github.com/Camionerou/rag-saldivia

---

## Arquitectura

```
Usuario --> Next.js :3000 (UI + auth + proxy) --> RAG Server :8081
                                                       |
                                                  Milvus + NIMs
                                                       |
                                                  Nemotron-Super-49B
```

Un único proceso Next.js reemplaza el gateway Python + frontend SvelteKit del stack viejo.

---

## Stack técnico

| Componente | Tecnología |
|---|---|
| Framework | Next.js 16 App Router |
| Lenguaje | TypeScript 6 |
| Runtime | Bun 1.3.x |
| Base de datos | SQLite vía Drizzle ORM + @libsql/client |
| Auth | JWT (jose) en cookie HttpOnly + Redis blacklist |
| Queue | BullMQ + Redis |
| AI/Streaming | Vercel AI SDK (`ai` + `@ai-sdk/react`) |
| Validación | Zod (compartido entre paquetes) |
| CSS | Tailwind v4 + shadcn/ui + Radix |
| Monorepo | Turborepo + Bun workspaces |
| Testing | bun:test + happy-dom + @testing-library/react + Playwright |

---

## Estructura del repo

```
apps/
  web/                  --> Next.js (UI + API + auth)
    src/
      app/              --> rutas y API routes
      components/       --> componentes React por dominio
      hooks/            --> (useChat from @ai-sdk/react)
      lib/              --> lógica (auth, rag, utils)

packages/
  db/                   --> Drizzle ORM + queries + schema
  shared/               --> Zod schemas + tipos
  config/               --> config loader
  logger/               --> logger estructurado

docs/
  bible.md              --> reglas permanentes del proyecto
  plans/                --> planes de implementación
  decisions/            --> ADRs (11 activas)
  templates/            --> templates (plan, commit, PR, version, ADR, artifact)
  artifacts/            --> resultados de reviews/audits
  toolbox.md            --> herramientas externas

_archive/               --> código aspiracional (recuperable con git)
```

---

## Comandos clave

```bash
bun run dev                     # Next.js en :3000
bun run test                    # unit tests via Turborepo
bun run test:components         # component tests (happy-dom) — correr desde apps/web/
bun run test:visual             # visual regression Playwright
bun run test:a11y               # WCAG AA con axe-playwright
bun run test:e2e                # E2E Playwright
bun run lint                    # lint via Turborepo
bun run storybook               # Storybook en :6006
bunx tsc --noEmit               # type check (correr desde apps/web/)
```

---

## Páginas activas (5 rutas)

```
/login                  --> pública, sin NavRail
/chat                   --> lista de sesiones + empty state
/chat/[id]              --> conversación con mensajes y input
/collections            --> lista de colecciones
/settings               --> perfil, contraseña, preferencias
```

---

## API routes activas

| Ruta | Método | Descripción |
|---|---|---|
| `/api/auth/login` | POST | JWT + cookie HttpOnly |
| `/api/auth/logout` | DELETE | Invalida sesión |
| `/api/auth/refresh` | POST | Renueva JWT |
| `/api/rag/generate` | POST | Proxy SSE al RAG server |
| `/api/rag/collections` | GET/POST | CRUD colecciones Milvus |
| `/api/rag/collections/[name]` | DELETE | Eliminar colección |
| `/api/rag/document/[name]` | GET | Documento por nombre |
| `/api/rag/suggest` | POST | Preguntas relacionadas |
| `/api/health` | GET | Health check |

---

## Server Actions activas

| Archivo | Actions |
|---|---|
| `actions/auth.ts` | `actionLogout` |
| `actions/chat.ts` | `actionCreateSession`, `actionDeleteSession`, `actionRenameSession`, `actionAddMessage`, `actionAddFeedback`, `actionToggleSaved`, `actionForkSession` |
| `actions/collections.ts` | `actionSetAreaCollections` |
| `actions/config.ts` | `actionUpdateRagParams`, `actionResetRagParams` |
| `actions/settings.ts` | `actionUpdateProfile`, `actionUpdatePassword`, `actionUpdatePreferences` |

---

## Componentes activos

```
components/
  ui/                 --> 18 primitivos: button, badge, input, textarea, avatar,
                          table, separator, tooltip, dialog, sheet, command,
                          sonner, theme-toggle, skeleton, stat-card,
                          empty-placeholder, data-table, popover
  chat/               --> ChatInterface, SessionList, SourcesPanel
  layout/             --> AppShell, AppShellChrome, NavRail
  collections/        --> CollectionsList
  settings/           --> SettingsClient, MemoryClient
  dev/                --> ReactScan, ReactScanProvider (solo dev)
  error-boundary.tsx
  providers.tsx       --> ThemeProvider (next-themes)
```

---

## Design system

### Tokens CSS

```css
/* Light mode */
--bg: #faf8f4;  --surface: #f0ebe0;  --surface-2: #e8e1d4;
--border: #ede9e0;  --fg: #1a1a1a;  --fg-muted: #5a5048;
--fg-subtle: #9a9088;  --accent: #1a5276;  --accent-subtle: #d4e8f7;
--success: #2d6a4f;  --destructive: #c0392b;  --warning: #d68910;
```

- **Font:** Instrument Sans via `next/font/google`
- **Tokens CSS siempre** — nunca hardcodear colores
- **`bg-surface`** para cards/paneles, **`bg-bg`** para fondo base
- **`@theme inline`** en Tailwind v4 para dark mode class-based

---

## Testing

| Capa | Comando | Tests |
|---|---|---|
| Lógica pura | `bun run test` | ~77 |
| Componentes | `bun run test:components` (desde apps/web/) | ~99 |
| Visual | `bun run test:visual` | 22 baselines |
| A11y | `bun run test:a11y` | páginas clave |
| E2E | `bun run test:e2e` | flujos críticos |

### Convenciones de component tests

```typescript
import { afterEach } from "bun:test"
import { cleanup, render, fireEvent } from "@testing-library/react"

afterEach(cleanup)  // OBLIGATORIO

// Queries escopadas, NO screen global:
const { getByRole } = render(<Button>Click</Button>)

// fireEvent sobre userEvent (happy-dom compatibility)
```

### Preloads
- **Component tests:** `--preload ./src/lib/component-test-setup.ts`
- **Lib tests:** `--preload ./src/lib/test-setup.ts`
- **DB tests:** ioredis-mock via `packages/db/bunfig.toml`

---

## Archivos críticos

| Archivo | Por qué |
|---|---|
| `apps/web/src/proxy.ts` | Middleware real: JWT + RBAC en edge |
| `apps/web/src/lib/auth/jwt.ts` | createJwt, verifyJwt, cookies |
| `apps/web/src/app/globals.css` | Tokens CSS — cambios afectan toda la app |
| `apps/web/src/lib/rag/ai-stream.ts` | Adapter: NVIDIA SSE → AI SDK Data Stream |
| `apps/web/src/components/chat/ChatInterface.tsx` | Componente más complejo de la UI |
| `apps/web/src/lib/component-test-setup.ts` | Setup happy-dom — modificar puede romper tests |
| `packages/db/src/schema.ts` | Schema SQLite completo |
| `packages/db/src/queries/users.ts` | CRUD usuarios + permisos |

---

## Variables de entorno

```env
JWT_SECRET=...             # openssl rand -base64 32
SYSTEM_API_KEY=...         # openssl rand -hex 32
RAG_SERVER_URL=http://localhost:8081
DATABASE_PATH=./data/app.db
REDIS_URL=redis://localhost:6379
MOCK_RAG=false             # true para dev sin Docker
LOG_LEVEL=INFO
```

### Credenciales de desarrollo (seed)

| Email | Password | Rol |
|---|---|---|
| `admin@localhost` | `changeme` | admin |
| `user@localhost` | `test1234` | user |

---

## Patrones importantes

### Streaming (AI SDK)
- Verificar status HTTP **ANTES** de streamear (patrón crítico en `lib/rag/client.ts`)
- `lib/rag/ai-stream.ts` transforma SSE de NVIDIA al protocolo AI SDK Data Stream
- `useChat` de `@ai-sdk/react` en ChatInterface (reemplaza `useRagStream`)
- Citations pasan como `data-sources` custom parts en el stream

### Redis (ADR-010)
- `getRedisClient()` nunca retorna null — lanza error
- NO importar logger en `redis.ts` (dependencia circular, ADR-005)
- BullMQ usa `getBullMQConnection()` con `{ maxRetriesPerRequest: null }`
- JWT revocación verificada en `extractClaims()`, NO en middleware edge

### Next.js
- Server Components por defecto, `"use client"` solo donde necesario
- `next/font/google` para Instrument Sans
- `next/dynamic` con `ssr: false` solo en Client Components

### Base de datos
- `Date.now()` para timestamps (ADR-004), no `_ts()` de SQLite
- CJS sobre ESM en packages (ADR-002)

---

## ADRs vigentes

| ADR | Decisión |
|---|---|
| 001 | libsql sobre better-sqlite3 |
| 002 | CJS sobre ESM en packages |
| 003 | Next.js como proceso único |
| 004 | Temporal API para timestamps |
| 005 | Import estático de db en logger |
| 006 | Estrategia de testing |
| 007 | Funciones reales sobre helpers en tests |
| 008 | Lector SSE compartido (superada — AI SDK adoptado en Plan 14) |
| 009 | Server Components primero |
| 010 | Redis como dependencia requerida |
| 011 | UI strategy (superada por Plan Maestro 1.0.x) |

---

## Agents disponibles (`.claude/agents/`)

| Agent | Cuándo |
|---|---|
| `frontend-reviewer` | Cambios en componentes/UI |
| `gateway-reviewer` | Cambios en API routes/auth |
| `security-auditor` | Antes de releases |
| `test-writer` | Tests nuevos |
| `debugger` | Algo no funciona |
| `doc-writer` | Actualizar docs |
| `deploy` | Deploy |
| `status` | Estado de servicios |
| `plan-writer` | Planes nuevos |
| `ingest` | Ingestar documentos |

Todos corren como Opus. Ver `.claude/agents/` para checklists detalladas.

---

## Referencias

- `docs/bible.md` — reglas permanentes
- `docs/plans/1.0.x-plan-maestro.md` — roadmap actual
- `docs/toolbox.md` — herramientas externas evaluadas
- `docs/templates/` — templates de plan, commit, PR, version, ADR, artifact
- `docs/decisions/` — ADRs
- `_archive/` — código aspiracional (admin, upload, extract, projects, CLI, etc.)

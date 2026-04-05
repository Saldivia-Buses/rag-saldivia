# Plan 8 — Optimización: Performance, Code Quality y Dependency Upgrades

> **Estado:** COMPLETADO — 2026-03-27
> **Branch:** `experimental/ultra-optimize`
> **Líneas originales:** ~2067 → comprimido a resumen post-ejecución

---

## Qué se hizo

Eliminación de deuda técnica acumulada en Plans 1-7: duplicación de código, bugs de performance, antipatrones React/Next.js, dependencias desactualizadas, y calidad estructural.

### Problemas resueltos

**Duplicación eliminada:**
- Lógica SSE duplicada en 6 archivos → lector compartido en `lib/rag/`
- `getCachedRagCollections` definida 2 veces → dead code eliminado
- `unknown[]` para sources → `CitationSchema` de `packages/shared`

**Performance fixes:**
- N+1 query en `getRateLimit` → `WHERE targetId IN (...)`
- `listWebhooksByEvent` filtraba en JS → filtro en SQL
- `canAccessCollection` sin caché → caché por request
- Sync bcrypt → async bcrypt
- Shiki recreado cada render → singleton + LRU cache
- `d3` (~450KB) y `react-pdf` (~600KB) → lazy loaded

**React/Next.js:**
- `settings/memory/page.tsx` con `"use client"` + raw `fetch()` → Server Component + Server Actions
- Cero memoización → `React.memo` + `useCallback` en ChatInterface
- Error Boundaries en rutas críticas (`/chat`, `/admin`)

**Dependency upgrades:**
- Next.js 15 → 16
- Drizzle 0.38 → 0.45
- Zod 3 → 4
- Lucide 0.x → 1.7

**Infraestructura:**
- Redis como dependencia requerida (ADR-010) — eliminados 11 workarounds
- BullMQ para cola de ingesta (reemplaza SQLite queue)
- CI paralelizado + cache de Bun

### Fases ejecutadas

| Fase | Qué |
|------|-----|
| 0 | Bundle analysis con `@next/bundle-analyzer` |
| 1 | Deduplicación: SSE reader, CitationSchema, dead code |
| 2 | React refactor: memo, callbacks, Server Components |
| 3 | Drizzle sync entre packages |
| 4 | Dependency upgrades (Next.js 16, Zod 4, etc.) |
| 5 | Docs: `architecture.md` actualizado, ADR-008, ADR-009 |
| 6 | Error Boundaries + CI paralelo |
| 7 | Logger improvements, event types |
| 8 | Redis obligatorio + BullMQ + JTI en JWT |

### Resultado

- ~2,516 líneas de dead code eliminadas
- 4 major dependency upgrades
- N+1 query eliminado del hot path
- Error Boundaries en rutas críticas
- ADR-008 (SSE compartido), ADR-009 (Server Components primero), ADR-010 (Redis requerido)

### Commits

| Fase | Commit | Descripción |
|------|--------|-------------|
| Pre | `fc3da02` | fix(web): excluir test-setup files del tsconfig — bun:test no es tipo de next.js |
| Plan | `bddfe68` | docs(plans): plan 8 — optimización, performance, Redis + BullMQ |
| F0 | `023276f` | docs(perf): baseline de medicion pre-plan8 |
| F1 | `200b919` | refactor(web): extraer sse reader, citation type, eliminar duplicados, cache canaccess |
| F2 | `46574e6` | perf(web): server pattern, memoizacion, lazy loading, next-safe-action, rhf, nuqs |
| F3.7 | `2b2d6ac` | chore(deps): sincronizar drizzle-orm en packages/db y apps/web |
| F3.8 | `5f81494` | ci(lint): extender type-check a todos los packages via turbo |
| F3.9 | `3693757` | chore(deps): drizzle-kit push reemplaza init.ts manual |
| F3.10 | `d9670e7` | chore(deps): remover deps sin uso + alinear exactoptionalpropertytypes |
| F4.9 | `03ac3ed` | chore(deps): upgrade next.js a 16.2.1 |
| F4.10 | `9e8d596` | chore(deps): upgrade drizzle-orm a 0.45.1, drizzle-kit a 0.31.10 |
| F4.11 | `312a614` | chore(deps): upgrade lucide-react a 1.7.0 |
| F4.12 | `23d2a8b` | chore(deps): upgrade zod a 4.3.6 |
| F4.13 | `f453c55` | chore(deps): upgrade @libsql/client a 0.17.2 |
| F5 | `e2e2f52` | docs: actualizar architecture.md post-plan8 |
| F6.15 | `2da2e42` | feat(web): error boundaries en chat y admin |
| F6.16 | `2dd5b36` | ci: paralelizar jobs, cache de bun, turbo --affected |
| F7 | `d596bba` | feat(logger): mejoras logging, retention, csv export, refactor |
| F3 docs | `c5bcaac` | docs(changelog): fase 3 completada — plan8 f3 |
| F4 docs | `b28d9e5` | docs(changelog): fase 4 completada — plan8 f4 |
| F4 docs | `e4a79a9` | docs(plan): marcar fase 4 como completada en plan8 |
| F6 docs | `c939340` | docs(plans): marcar fase 6 como completada — plan8 f6 |
| Infra | `4cb7b57` | chore: trackear archivos previamente ignorados — data, logs, superpowers, playwright, turbo |
| Infra | `442e735` | chore(git): agregar .turbo/ y logs/ a .gitignore |
| Infra | `e30052a` | chore(git): ignorar db, uploads, playwright test-results |
| F8 | `a2e5404` | feat: redis como dependencia requerida — eliminar 11 workarounds de single-instance |

# Plan 9 — Repo Limpio: Cero Dead Code, Cero Errores, Linting Perfecto

> **Estado:** COMPLETADO — 2026-03-27
> **Branch:** `experimental/ultra-optimize`
> **Líneas originales:** ~679 → comprimido a resumen post-ejecución

---

## Qué se hizo

Limpieza completa del repo como artefacto: archivos trackeados incorrectamente, errores TypeScript, dead code, dependencias sin uso, linting.

### Problemas resueltos

**47 archivos removidos del tracking:**
- `.playwright-mcp/` — logs y screenshots de sesiones MCP
- `.superpowers/brainstorm/` — HTMLs de brainstorming
- `apps/web/logs/backend.log` — log de runtime
- `config/.env.saldivia` — archivo de env
- `docs/superpowers/` — specs de diseño interno

**5 errores TypeScript corregidos:**
- `updateTag` no existe en `next/cache` de Next.js 16
- `IngestionEventRecord` incompatible con `exactOptionalPropertyTypes`

**Dead code eliminado (8 archivos, 3 deps):**
- Features que requieren GPU (crossdoc, collection history, split view)
- Stub de SSO
- Dependencias sin consumidores: `d3`, `@types/d3`

### Fases ejecutadas

| Fase | Qué |
|------|-----|
| F9.1 | Git purge + `.gitignore` perfecto |
| F9.2 | Fix errores TypeScript |
| F9.3 | Eliminar dead code confirmado |
| F9.4 | Eliminar dependencias sin uso |
| F9.5 | ESLint limpio en todo el monorepo |
| F9.6 | Knip scan de exports sin usar |
| F9.7 | Husky + commitlint verificado |
| F9.8 | Console.log audit |
| F9.9 | Verificación final |

### Commits

| Fase | Commit | Descripción |
|------|--------|-------------|
| F9.1 | `47af532` | chore(ci): purgar artefactos trackeados y ampliar gitignore |
| F9.2 | `69de39d` | fix(blackbox): spread condicional en handleingestion* para exactoptionalpropertytypes |
| F9.2 | `c9b3933` | fix(db): completar f9.2 — ioredis-mock types, excluir test-setup de tsc, purgar turbo cache |
| F9.3-9 | `f5e30a8` | chore(web): plan9 completado — dead code, deps, eslint, knip, husky y consola |
| Docs | `09bab25` | docs(plans): agregar planes 9-12 para preparacion release 1.0.0 |
| Docs | `7f2fc0f` | docs(changelog): registrar f9.2 parcialmente completado — blackbox y collections fixes |
| Docs | `08138e3` | docs(plans): alinear checklist plan9 con commits y verificación commitlint |
| Docs | `fb55fbe` | docs(plans): marcar git push verificado en checklist plan9 |
| Docs | `bb764d2` | docs(claude): remover referencia a collectionhistory eliminado en plan9 |
| Docs | `09bab25` | docs(plans): agregar planes 9-12 para preparacion release 1.0.0 |
| Docs | `7f2fc0f` | docs(changelog): registrar f9.2 parcialmente completado |
| Docs | `08138e3` | docs(plans): alinear checklist plan9 con commits y verificación commitlint |
| Docs | `fb55fbe` | docs(plans): marcar git push verificado en checklist plan9 |
| Docs | `bb764d2` | docs(claude): remover referencia a collectionhistory eliminado en plan9 |

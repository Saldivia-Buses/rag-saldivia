# Plan 3 — Bugfix & Code Quality (CodeGraph Analysis)

> **Estado:** COMPLETADO — 2026-03-25
> **Branch:** `experimental/ultra-optimize`
> **Líneas originales:** ~165 → comprimido a resumen post-ejecución

---

## Qué se hizo

Análisis del codebase con **CodeGraphContext MCP** (106 archivos, 289 funciones). Se encontraron bugs latentes, código muerto y funciones de alta complejidad. Se corrigieron los bugs reales y se refactorizaron las funciones más complejas.

### Hallazgos y acciones

| ID | Tipo | Acción |
|----|------|--------|
| BUG-1 | `removeAreaCollection` ignoraba `collectionName` en WHERE | Corregido — agregado `and()` con ambas condiciones |
| BUG-2 | `actionSetAreaCollections` sin callers | **Falso positivo** — ya estaba conectada en callback |
| BUG-3 | `actionUpdateArea` logueaba `"user.updated"` | Corregido — 3 event types cambiados a prefijo `area.*` |
| DEAD-1 | `progressBar` sin callers | **Falso positivo** — CodeGraph no rastreó import con `.js` |
| CMPLX-1 | `ChatInterface` complejidad 48 | Refactorizado → 22 (extraído `useRagStream` hook) |
| CMPLX-2 | `reconstructFromEvents` complejidad 34 | Refactorizado → ~5 (event handlers map) |

### Resultado

- 1 bug de datos crítico corregido (WHERE incompleto)
- 3 event types de auditoría corregidos
- 2 falsos positivos del grafo documentados
- `ChatInterface` 48 → 22, `reconstructFromEvents` 34 → ~5

### Commits

| Fase | Commit | Descripción |
|------|--------|-------------|
| F1+F2 | `57bd443` | fix(db): where en remove-area-collection ignoraba el nombre de coleccion |
| F2 | `5ddb6c4` | fix(shared): agregar area.created/updated/deleted al union de event types |
| F4 | `502a625` | refactor(chat): extraer logica sse al hook use-rag-stream |
| Infra | `1132afe` | chore: add .cgcignore for codegraphcontext |

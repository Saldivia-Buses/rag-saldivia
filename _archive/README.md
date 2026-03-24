# _archive/

Este directorio contiene el código del stack original de RAG Saldivia (Python + SvelteKit).
Fue reemplazado por la reescritura completa en TypeScript en la branch `experimental/ultra-optimize`.

## Contenido

- `saldivia/` → Gateway Python (FastAPI, puerto 9000) + SDK + tests
- `services/sda-frontend/` → Frontend SvelteKit 5 (BFF, puerto 3000)  
- `cli/` → CLI Python original (424 líneas)

## Por qué está archivado y no eliminado

- Referencia durante la migración
- Tests de regresión: verificar que el nuevo stack tiene feature parity
- Rollback de emergencia si se necesita volver al stack Python

## Stack reemplazante

Ver `apps/web/` (Next.js 15) y `apps/cli/` (TypeScript + Commander).

La nueva arquitectura está documentada en `docs/architecture.md`.

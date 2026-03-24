# legacy/

Este directorio contiene el código del stack original de RAG Saldivia (Python + SvelteKit).
Fue reemplazado por la reescritura completa en TypeScript en la branch `experimental/ultra-optimize`.

## Contenido

- `saldivia/` → Gateway Python (FastAPI, puerto 9000) + SDK Python + tests
- `sda-frontend/` → Frontend SvelteKit 5 (BFF, puerto 3000)
- `cli/` → CLI Python original (424 líneas)
- `auth-gateway/` → Dockerfile del gateway Python
- `ingestion-worker/` → Dockerfile del worker Python

## Por qué está en legacy y no eliminado

- Referencia durante la migración
- Tests de regresión: verificar que el nuevo stack tiene feature parity
- Rollback de emergencia si se necesita volver al stack Python

## Stack reemplazante

| Legacy | Nuevo |
|---|---|
| `saldivia/gateway.py` (Python) | `apps/web/` (Next.js 15) |
| `sda-frontend/` (SvelteKit) | `apps/web/` (Next.js 15) |
| `cli/` (Python) | `apps/cli/` (TypeScript) |

La nueva arquitectura está documentada en `docs/architecture.md`.

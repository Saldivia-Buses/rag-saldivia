# `@rag-saldivia/shared`

**Schemas Zod** y tipos compartidos entre `apps/web`, scripts y cualquier consumidor que deba hablar el mismo contrato que el servidor.

## Regla

Si **dos aplicaciones** (p. ej. web y CLI) o **web + DB** necesitan el mismo tipo o validación, **va aquí** — no duplicar schemas en cada app.

## Contenido principal

| Área | Ejemplos |
|------|----------|
| Usuario / sesión | Tipos derivados de tablas o DTOs de API |
| RAG | `RagParamsSchema` — parámetros del LLM (temperature, top_p, …) |
| Eventos | `EventTypeSchema`, `LogLevelSchema`, `LogEventSchema` |
| Roles | `"admin" \| "area_manager" \| "user"` |
| Focus modes | `FocusModeId` / modos de prompt en UI |
| Login | `LoginRequestSchema` |

Archivo principal: `src/schemas.ts`.

## Tests

```bash
bun test packages/shared/
```

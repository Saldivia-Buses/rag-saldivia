# ADR-008: Extracción de SSE reader compartido a `lib/rag/stream.ts`

**Estado:** Aceptado  
**Fecha:** 2026-03-27  
**Contexto:** Plan 8 — Fase 1 (Eliminación de duplicación)

---

## Contexto

La lógica de leer un stream SSE (`getReader() + TextDecoder + parseo de líneas data:`) estaba copiada en 5 lugares del código:

| Archivo | Forma |
|---|---|
| `hooks/useRagStream.ts` | inline en el hook |
| `hooks/useCrossdocStream.ts` | función local `collectStream` |
| `hooks/useCrossdocDecompose.ts` | función local `collectSseText` |
| `app/api/slack/route.ts` | inline en el handler |
| `app/api/teams/route.ts` | inline en el handler |

Cada copia tenía variantes sutiles (distintos guards, con/sin detección de repetición, con/sin filtro de longitud de token), lo que dificultaba razonar sobre el comportamiento del sistema.

## Decisión

Crear `apps/web/src/lib/rag/stream.ts` con tres funciones públicas:

- **`parseSseLine(line: string): string | null`** — extrae el token de contenido de una línea `data: {...}`. Retorna null para `[DONE]`, líneas sin prefijo, JSON malformado o delta sin content.
- **`readSseTokens(body: ReadableStream<Uint8Array>): AsyncGenerator<string>`** — generator que yields tokens individuales. Incluye buffering de líneas parciales para manejar chunks que cortan en mitad de una línea SSE.
- **`collectSseText(response, options?): Promise<string>`** — acumula todo el texto del stream. Soporta `maxChars` y `detectRepetition`. Maneja tanto SSE como respuestas JSON estándar.

La detección de repetición (portada de `useCrossdocStream`) vive en `stream.ts` como función interna — no se expone porque es un detalle de implementación.

## Por qué en `lib/rag/stream.ts` y no en `packages/shared`

`packages/shared` exporta tipos y schemas Zod usados tanto en `apps/web` como en `apps/cli`. Las funciones de `stream.ts` dependen de APIs de browser/Node (`ReadableStream`, `Response`, `TextDecoder`) que no tienen sentido en la CLI ni en código de servidor puro. Vivir en `lib/rag/` las mantiene cerca de su contexto de uso sin contaminar el package compartido.

## Consecuencias

**Positivo:**
- La lógica SSE existe en un solo lugar — cualquier fix o mejora se aplica automáticamente a todos los usuarios
- `useCrossdocStream` y `useCrossdocDecompose` eliminaron ~50 líneas cada uno
- `slack/route.ts` y `teams/route.ts` eliminaron ~20 líneas de boilerplate cada uno
- El buffering de líneas parciales (ausente en las implementaciones originales) beneficia a todas las callers automáticamente

**Negativo / Trade-offs:**
- `useRagStream` no puede usar `readSseTokens` directamente porque necesita parsear `delta.sources` (del mismo evento SSE) además del contenido. Usa `parseSseLine` para el contenido y parsea sources inline.
- `collectSseText` no tiene el filtro `token.length < 500` que `useCrossdocStream` tenía — se consideró exceso defensivo dado que el RAG server no produce tokens malformados de longitud excesiva.

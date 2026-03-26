# ADR-003: Next.js como proceso único (reemplaza Python gateway + SvelteKit)

**Fecha:** 2026-03-24
**Estado:** Aceptado

---

## Contexto

El stack original (`main`) tiene dos procesos separados:
- `saldivia/gateway.py` (FastAPI, puerto 9000) — autenticación, RBAC, proxy al RAG server, audit log
- `services/sda-frontend/` (SvelteKit, puerto 3000) — UI, BFF pattern

Ambos procesos se coordinan vía HTTP interno. El deploy requiere Docker Compose con múltiples servicios Python. La branch `experimental/ultra-optimize` es una reescritura completa del stack.

## Opciones consideradas

- **Mantener la separación gateway + frontend, reescribir en TypeScript:** misma arquitectura, diferente lenguaje. Pros: separación de responsabilidades clara. Contras: sigue siendo dos procesos coordinados; duplica el overhead de deploy; el BFF pattern de SvelteKit ya hace lo mismo que un gateway en Next.js App Router.
- **Next.js App Router como proceso único:** UI + auth middleware + proxy RAG + admin en un solo proceso. Los Server Components y Route Handlers reemplazan el gateway; el middleware de Next.js reemplaza el auth layer. Pros: un solo proceso, un solo deploy, TypeScript end-to-end, cero overhead de comunicación interna. Contras: mezcla responsabilidades UI y backend en el mismo proceso.
- **Nest.js como gateway + Next.js como frontend:** separación limpia pero con TypeScript. Contras: sigue siendo dos procesos; Nest.js agrega complejidad de DI/decorators innecesaria para este caso.

## Decisión

Elegimos **Next.js como proceso único** porque el App Router de Next.js 15 es suficientemente potente como gateway HTTP gracias a:
- `middleware.ts` en el edge para JWT + RBAC en cada request (reemplaza `auth_middleware` de FastAPI)
- Route Handlers (`route.ts`) para endpoints REST (reemplaza endpoints FastAPI)
- Server Actions para mutaciones (elimina la necesidad de endpoints POST manuales)
- `SYSTEM_API_KEY` header para auth service-to-service (CLI → Next.js)

El proceso escucha en puerto 3000 y se comunica con el RAG server (`localhost:8081`) vía HTTP interno desde Server Components y Route Handlers.

## Consecuencias

**Positivas:**
- Un solo proceso para deployar, monitorear y reiniciar.
- TypeScript end-to-end: tipos compartidos entre UI, server y CLI via `packages/shared`.
- Sin overhead de red interno entre gateway y frontend.
- `bun run dev` levanta todo el stack de la aplicación en un comando.
- El middleware de Next.js corre en el edge runtime — verificación JWT es O(1) y no bloquea el thread principal.

**Negativas / trade-offs:**
- El proceso Next.js mezcla responsabilidades de presentación y proxy HTTP. Si en el futuro se necesita escalar el gateway independientemente de la UI, habría que separar.
- SSE requiere cuidado especial: verificar el status HTTP de la respuesta del RAG server **antes** de iniciar el stream. El gateway Python original tenía un bug donde siempre retornaba 200 aunque el RAG fallara — documentado en `CLAUDE.md` como patrón a evitar.
- El worker de ingesta (`apps/web/src/workers/ingestion.ts`) corre dentro del mismo proceso Next.js en desarrollo. En producción debería ser un proceso separado o un cron job.

## Referencias

- `apps/web/src/middleware.ts` — auth + RBAC
- `apps/web/src/lib/rag/client.ts` — proxy RAG con modo mock
- `apps/web/src/workers/ingestion.ts` — worker de ingesta
- `CLAUDE.md` — arquitectura de servicios
- `docs/architecture.md`

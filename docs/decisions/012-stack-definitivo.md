# ADR-012: Stack definitivo para la serie 1.0.x

**Fecha:** 2026-03-29
**Estado:** Aceptada
**Contexto:** Plan 13 — decisiones de stack tomadas por Enzo

---

## Contexto

Las decisiones de stack se tomaron originalmente para un prototipo rápido (Plans 1-12).
La visión del producto creció (SaaS multi-tenant, backend-heavy, WhatsApp/email). Se
necesitaba evaluar formalmente si el stack actual sigue siendo correcto.

Prioridad definida por Enzo: primero que funcione para testers (Enzo, su papá, su tío),
después evaluar para SaaS.

## Opciones consideradas

### Opción A — Mantener stack actual (Next.js + SQLite + Bun)

- **Pros:** ya funciona, TypeScript end-to-end, rápido para single-tenant, equipo de 1
- **Contras:** SQLite no escala a multi-tenant, Next.js mezcla UI y backend

### Opción B — Migrar a stack enterprise (Nest.js + Postgres + Docker)

- **Pros:** separación backend/frontend, Postgres escala, standard enterprise
- **Contras:** reescritura masiva, más complejidad, 1 persona no lo mantiene

### Opción C — Stack actual + migración incremental preparada

- **Pros:** no reescribir, Drizzle facilita migrar DB, packages/ permite extraer API
- **Contras:** depende de que Drizzle cumpla la promesa de portabilidad

## Decisión

**Opción C.** El stack actual se mantiene con preparación para escalar:

| Componente | Tecnología | Migración futura |
|---|---|---|
| Frontend + Backend | Next.js 16 App Router | Extraíble via packages/ |
| Base de datos | SQLite (libsql) | Postgres via Drizzle (1 día) |
| ORM | Drizzle | Soporta SQLite y Postgres |
| Runtime | Bun | — |
| Auth | JWT (jose) + Redis blacklist | — |
| Queue | BullMQ + Redis | — |
| AI/Streaming | Vercel AI SDK (`ai`) | Reemplaza SSE manual (Plan 14) |
| Generative UI | json-render | Plan 17 |
| Validación | Zod | — |
| CSS | Tailwind v4 + shadcn/ui + Radix | Tokens claude-azure |
| Monorepo | Turborepo + Bun workspaces | — |
| Testing | bun:test + happy-dom + Playwright | — |

**Modelo de deploy:** single-tenant by deployment. Cada empresa = su propio servidor.
No es multi-tenant con DB compartida.

**Branding:** "Saldivia RAG", acento azure blue (no naranja Claude), UI en español,
código en inglés.

## Consecuencias

**Positivas:**
- No hay reescritura. Se sigue construyendo sobre lo que funciona.
- Vercel AI SDK estandariza el streaming (industria standard).
- json-render agrega respuestas ricas sin trabajo custom.
- Drizzle como ORM permite migrar DB sin tocar lógica de negocio.
- packages/ permite extraer API backend si se necesita separar.

**Negativas / trade-offs:**
- SQLite tiene límite de concurrencia (~50 usuarios simultáneos).
- Next.js mezcla UI y backend (mitigado por packages/).
- Bun es menos maduro que Node.js en producción.
- La migración a Postgres no está testeada aún.

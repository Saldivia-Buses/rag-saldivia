# 01 — Resumen Ejecutivo

## Que es Saldivia RAG

Un **overlay empresarial** sobre el NVIDIA RAG Blueprint v2.5.0 que agrega todo lo que el blueprint no tiene:

- Autenticacion JWT con revocacion via Redis
- RBAC granular (roles + permisos por coleccion)
- Frontend Next.js 16 con design system propio
- Sistema de mensajeria interna (Slack-like)
- Panel de administracion completo
- Multi-coleccion con permisos por area

**Modelo de negocio:** Single-tenant by deployment. Cada empresa = su propio servidor con el sistema completo. No es multi-tenant con DB compartida.

**Hardware target:** Workstation fisica Ubuntu 24.04 con 1x RTX PRO 6000 Blackwell (96GB VRAM) corriendo Nemotron-Super-49B.

---

## Numeros clave (medidos con repomix Tree-sitter compression)

| Metrica | Valor |
|---------|-------|
| Archivos TS/TSX activos (sin tests) | ~215 |
| Caracteres de codigo activo | ~200,000 |
| Tokens de codigo activo | ~50,000 |
| Lineas de codigo activo | ~6,800 |
| Tablas SQLite | 31 en 4 modulos |
| Paginas (rutas) | 16 activas |
| API routes | 21 (+/api/feedback, /api/connectors/callback, /api/auth/sso/*) |
| Server actions | 50+ en 10 archivos |
| Componentes React | ~80 activos (+ ErrorRecovery, AdminSso, AdminConnectors, SsoProviderForm) |
| Hooks custom | 7 (6 con test — solo usePresence sin test) |
| Modulos de queries | 21 |
| Tests totales | ~1,059 (693 unit + 314 component + 22 visual + 8 a11y + ~22 E2E) |
| ADRs documentadas | 12 (001-012) |
| Planes completados | 21 (Plans 13-30, 32-34) |
| Planes pendientes | 1 (Plan 19 — deploy testers) |
| Archivos archivados | 123 en `_archive/` |
| Agents especializados | 10 (todos Opus) |
| Documentos en docs/ | 32+ |

---

## Estado actual (2026-03-31)

### Funcional y activo
- Login con JWT (cookie HttpOnly)
- Chat RAG con streaming (AI SDK)
- Colecciones con permisos por area
- Panel admin: usuarios, roles, areas, permisos, config RAG
- Mensajeria interna: canales, DMs, threads, reactions, presencia
- Settings: perfil, password, preferencias
- Design system completo con Storybook
- Suite de testing con 380+ tests en verde

### Completado post-analisis (Plans 26-30)
- Config unification + constants centralizados (Plan 26)
- JWT access token 15min + refresh token 7d con rotation (Plan 26)
- SQLite PRAGMAs (WAL, foreign_keys, busy_timeout) (Plan 26)
- Next.js standalone output + compression + security headers (Plan 26)
- Redis hardening (memory limits, persistence) (Plan 26)
- Loading skeletons para TODAS las rutas (Plan 26)
- ~125 tests nuevos: RBAC, channels, ai-stream, RAG client, auth routes (Plan 27)
- FTS5 MATCH input sanitization (Plan 28)
- AES-256-GCM encryption para external_sources.credentials (Plan 28)
- ChatInterface descompuesto: 643 → 360 lineas (Plan 28)
- AdminRoles: 626 → 238 lineas, AdminUsers: 592 → 260 lineas (Plan 28)
- Error feedback system con dialog + toast + API (Plan 28)
- Component tests: 158 → 314, cobertura 19% → 100% (Plan 29)
- E2E + a11y expansion (8 paginas auditadas) (Plan 29)
- Dark mode token fixes + semantic tokens + Mermaid theme (Plan 30)

### Pendiente
- **Plan 19:** Deploy para testers (Enzo, su papa, su tio)
- **Plan 30 F3-F4:** Visual contrast audit + baseline update (pendiente verificacion visual)

### Modo de desarrollo
- `MOCK_RAG=true` permite desarrollar sin GPU/Docker
- OpenRouter como fallback para LLM real sin GPU local
- SQLite local para desarrollo, Redis requerido para BullMQ y JWT revocation

---

## Stack en una tabla

| Componente | Tecnologia | Version |
|-----------|-----------|---------|
| Runtime | Bun | 1.3.x |
| Framework | Next.js App Router | 16 |
| Lenguaje | TypeScript | 6.0 |
| Base de datos | SQLite (libsql) + Drizzle ORM | 0.45 |
| Auth | JWT (jose) + Redis blacklist | — |
| Queue | BullMQ + Redis | — |
| AI/Streaming | Vercel AI SDK (`ai` + `@ai-sdk/react`) | — |
| Validacion | Zod + next-safe-action | 3.x |
| CSS | Tailwind v4 + shadcn/ui + Radix | 4.x |
| Monorepo | Turborepo + Bun workspaces | 2.x |
| Testing | bun:test + happy-dom + Playwright | — |
| Catalogo UI | Storybook 8 | — |

---

## Arquitectura en 1 diagrama

```
Usuario (Browser :3000)
        |
   Next.js 16 App Router (proceso unico)
        |
   +---------+-----------+-----------+
   |         |           |           |
 proxy.ts  Pages     API Routes  Server Actions
 (JWT+RBAC) (14)      (18)       (50+)
   |         |           |           |
   +---------+-----------+-----------+
        |                    |
   +----+----+          +---+---+
   |         |          |       |
 SQLite    Redis     RAG :8081  WebSocket
(Drizzle) (blacklist  (NVIDIA)  (presencia)
           + BullMQ)    |
                    +---+---+
                    |       |
                  Milvus  Nemotron-49B
                 (vectores) (LLM, 96GB)
```

---

## Quien trabaja en esto

**Enzo Saldivia** — unico desarrollador. Claude Opus (Claude Code CLI) hace todo: planifica, ejecuta, revisa, testea, documenta. El workflow es OODA-SQ con sprint sequence formal y quality gates en cada fase.

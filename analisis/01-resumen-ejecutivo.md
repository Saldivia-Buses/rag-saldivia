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
| Archivos TS/TSX activos (sin tests) | 206 |
| Caracteres de codigo activo | 191,487 |
| Tokens de codigo activo | 48,231 |
| Lineas de codigo activo | 6,488 |
| Tablas SQLite | 31 en 4 modulos |
| Paginas (rutas) | 16 activas |
| API routes | 18 |
| Server actions | 50+ en 10 archivos |
| Componentes React | 68 activos (34 sin test) |
| Hooks custom | 7 (3 con test: useAutoResize, useCopyToClipboard, useLocalStorage) |
| Tablas SQLite | 20+ en 4 modulos |
| Modulos de queries | 21 |
| Tests totales | ~380 (198 unit + 158 component + 22 visual) |
| ADRs documentadas | 12 (001-012) |
| Planes completados | 13 (Plan 13-25) |
| Planes pendientes | 2 (Plan 19, 20) |
| Archivos archivados | 123 en `_archive/` |
| Agents especializados | 10 (todos Opus) |
| Documentos en docs/ | 32 |

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

### Pendiente
- **Plan 19:** Deploy para testers (Enzo, su papa, su tio) — requiere Plans 13-18
- **Plan 20:** Dark mode — requiere Plan 16
- E2E testing con Playwright (flujos criticos)

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

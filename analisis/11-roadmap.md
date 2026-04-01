# 11 — Roadmap

## Serie 1.0.x — Estado completo

La serie 1.0.x NO es una secuencia de releases granulares. Es **una version grande** construida por multiples planes. Cada plan aporta algo concreto.

---

## Tabla de planes

| Plan | Foco | Prerequisito | Estado | Fecha |
|------|------|-------------|--------|-------|
| **13** | Reorganizacion agentica | — | Completado | 2026-03-29 |
| **14** | Vercel AI SDK (reemplazar SSE manual) | 13 | Completado | 2026-03-29 |
| **15** | UI tokens + branding (azure, "Saldivia RAG") | 13 | Completado | 2026-03-29 |
| **16** | Core UI polish (login, chat, collections, settings, espanol) | 15 | Completado | 2026-03-29 |
| **17** | Markdown rendering + OpenRouter mock | 14 | Completado | 2026-03-29 |
| **18** | Testing + visual regression con nueva UI | 16 | Completado | 2026-03-29 |
| **19** | Deploy para testers (Enzo, papa, tio) | 13-18 | **Pendiente** | — |
| **20** | Dark mode → renumerado Plan 30 | 16 | — | — |
| **21** | RBAC admin (roles, permisos granulares) | — | Completado | 2026-03-30 |
| **22** | Collection polish + confirm dialogs + tests | 21 | Completado | 2026-03-30 |
| **23** | Code Hardening (security, perf, a11y, tests) | 18 | Completado | 2026-03-30 |
| **24** | Admin system (collections, areas, permissions, config) | 23 | Completado | 2026-03-30 |
| **25** | Internal Messaging System (canales, DMs, threads) | 23, 24 | Completado | 2026-03-31 |
| **26** | Production Hardening & Config Unification | 25 | Completado | 2026-04-01 |
| **27** | Testing Coverage Sprint (DB, lib, auth routes) | 26 | Completado | 2026-04-01 |
| **28** | Security Fixes + Decomposition + Error UX | 27 | Completado | 2026-04-01 |
| **29** | Testing Sprint II (components, E2E, a11y) | 28 | Completado | 2026-04-01 |
| **30** | Dark Mode Refinement | 28 | Completado (F3-F4 visual pending) | 2026-04-01 |

---

## Detalle de cada plan completado

### Plan 13 — Reorganizacion agentica
**Problema:** Agents escritos para stack Python viejo. CLAUDE.md describia 24 paginas como activas. Codigo aspiracional mezclado con core.
**Solucion:**
- Reescritura de 10 agents para el stack TS
- CLAUDE.md refleja solo lo activo
- 123 archivos movidos a `_archive/`
- Template de plan estandar
- Sprint sequence formal
- `docs/toolbox.md` creado

### Plan 14 — Vercel AI SDK
**Problema:** SSE manual custom, fragil, sin estandar.
**Solucion:**
- `ai-stream.ts`: adapter NVIDIA SSE → AI SDK Data Stream
- `useChat` reemplaza `useRagStream` custom
- Citations como `data-sources` custom parts
- Mock con OpenRouter para dev sin GPU

### Plan 15 — UI tokens + branding
**Problema:** Tokens genericos, sin identidad visual.
**Solucion:**
- Design system "Warm Intelligence" con tokens claude + azure
- Instrument Sans como font principal
- Tokens CSS en `globals.css`
- `@theme inline` para dark mode class-based

### Plan 16 — Core UI polish
**Problema:** UI en ingles, sin pulir.
**Solucion:**
- Toda la UI en espanol
- Login, chat, collections, settings pulidos
- NavRail de 64px con iconos
- Empty states con EmptyPlaceholder

### Plan 17 — Markdown rendering + OpenRouter mock
**Problema:** Respuestas del RAG sin formato rico.
**Solucion:**
- MarkdownMessage con syntax highlighting
- Artifact parser (code, html, svg, mermaid)
- OpenRouter como mock LLM real sin GPU

### Plan 18 — Testing + visual regression
**Problema:** Sin tests de la UI nueva.
**Solucion:**
- 158 component tests con happy-dom
- 22 visual regression tests con Playwright
- A11y tests con axe-playwright
- Preloads configurados

### Plan 21 — RBAC admin
**Problema:** Permisos demasiado simples (solo 3 roles).
**Solucion:**
- Schema: roles, permissions, role_permissions, user_role_assignments
- Queries RBAC completas
- Seed con roles y permisos por defecto
- Presencia (lastSeen)

### Plan 22 — Collection polish
**Problema:** Colecciones sin funcionalidad completa.
**Solucion:**
- CollectionSelector con multi-select
- ConfirmDialog para acciones destructivas
- Tests adicionales

### Plan 23 — Code Hardening
**Problema:** Seguridad, performance, y a11y sin revision sistematica.
**Solucion:**
- Security headers, rate limiting, input sanitization
- Performance: caching, lazy loading, bundle optimization
- A11y: WCAG AA en paginas clave
- Tests adicionales

### Plan 24 — Admin system
**Problema:** Sin panel de administracion.
**Solucion:**
- 7 paginas admin: dashboard, users, roles, areas, permissions, collections, config
- 11 componentes admin
- Server actions con adminAction middleware
- DataTable avanzada para todas las listas

### Plan 25 — Internal Messaging
**Problema:** Sin comunicacion entre usuarios del sistema.
**Solucion:**
- 19 componentes de messaging
- 9 API routes
- Schema: channels, members, messages, reactions, mentions, pins
- WebSocket sidecar para presencia y typing
- Hooks: useMessaging, usePresence, useTyping

---

## Detalle de planes 26-30

### Plan 26 — Production Hardening & Config Unification
**Objetivo:** Zero features nuevas. 9 fases de hardening para produccion.
**Entregables:** Config centralizada (17 constantes), JWT access+refresh rotation, SQLite PRAGMAs (WAL), Next.js standalone + compression + security headers, Redis hardening, loading skeletons para todas las rutas, Cache-Control headers, React.memo optimizations, Redis caching para admin stats, backup script, 9 archivos dead code eliminados.

### Plan 27 — Testing Coverage Sprint
**Objetivo:** Subir cobertura de 38% a ~55%. ~125 tests nuevos.
**Entregables:** Tests RBAC (25-30), channels+messaging (15-20), ai-stream adapter (8-10), RAG client (8-10), auth routes (12-15). DB queries 81% → 95%.

### Plan 28 — Security Fixes + Decomposition + Error UX
**Objetivo:** 2 fixes de seguridad, 3 decomposiciones de componentes, sistema de error feedback.
**Entregables:** FTS5 sanitization, AES-256-GCM para credentials, ChatInterface 643→360 lineas, AdminRoles 626→238, AdminUsers 592→260 (7 sub-componentes nuevos), error feedback system (lib + UI + API).

### Plan 29 — Testing Sprint II (Components + E2E + A11y)
**Objetivo:** Cobertura de componentes 19% → 100%.
**Entregables:** 314 component tests (vs 158 anterior), hooks 6/7, messaging 19/19, admin 11/11, E2E auth+chat+admin, a11y 8 paginas.

### Plan 30 — Dark Mode Refinement
**Objetivo:** Refinar tokens dark mode, migrar a tokens semanticos.
**Entregables:** accent-fg fix, 5x text-white → text-accent-fg, Mermaid theme-aware, bg-white → bg-bg. Pendiente: visual contrast audit (F3-F4).

---

## Planes pendientes

### Plan 19 — Deploy para testers
**Objetivo:** Primera version deployable para 3 testers (Enzo, su papa, su tio).
**Prerequisitos:** Plans 13-18 (completados). Plans 26-30 lo hacen mucho mas production-ready.
**Que falta:** Configurar deploy en workstation fisica, seedear datos, documentar setup.

---

## Planes futuros (en definicion)

| Plan | Foco | Estado |
|------|------|--------|
| **32** | Self-Healing Error UX | Plan en escritura |
| **33** | Conectores Externos (Google Drive, SharePoint, Confluence) | Plan en escritura |
| **34** | SSO (SAML/OIDC) | Plan en escritura |

### Otros futuros (sin plan formal)

| Area | Ideas |
|------|-------|
| RAG engine | Knowledge gaps, RAG Fusion, Graph RAG, CRAG |
| RAG quality | Hallucination detection, RAGAS evaluation |
| Colaboracion | RAG multiplayer (queries compartidas en canales) |
| Integraciones | Slack bot, Teams bot, MCP server |
| Scale | Postgres migration, multi-tenant |
| Mobile | PWA con voice input |

---

## Sprint sequence (como se ejecuta cada plan)

```
THINK   → Opus cuestiona scope. Es necesario? Scope minimo? Ya existe solucion al 80%?
PLAN    → Plan detallado en docs/plans/. Enzo aprueba.
EXECUTE → Opus ejecuta fase por fase. Contexto completo.
REVIEW  → 2 pasadas: correcto/compila? + scope drift/security?
TEST    → tsc → test → lint (por fase). + components/visual/a11y (por plan).
SHIP    → Docs + CHANGELOG. Commit. Tag si es release.
```

**Commits:** `tipo(scope): descripcion — planN fX`
**Branch:** directo en `1.0.x` (no hay feature branches mientras sea Enzo solo)

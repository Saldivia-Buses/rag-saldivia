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
| **20** | Dark mode refinamiento | 16 | **Pendiente** | — |
| **21** | RBAC admin (roles, permisos granulares) | — | Completado | 2026-03-30 |
| **22** | Collection polish + confirm dialogs + tests | 21 | Completado | 2026-03-30 |
| **23** | Code Hardening (security, perf, a11y, tests) | 18 | Completado | 2026-03-30 |
| **24** | Admin system (collections, areas, permissions, config) | 23 | Completado | 2026-03-30 |
| **25** | Internal Messaging System (canales, DMs, threads) | 23, 24 | Completado | 2026-03-31 |

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

## Planes pendientes

### Plan 19 — Deploy para testers
**Objetivo:** Primera version deployable para 3 testers (Enzo, su papa, su tio).
**Prerequisitos:** Plans 13-18 (completados)
**Que falta:** Configurar deploy en workstation fisica, seedear datos, documentar setup.

### Plan 20 — Dark mode
**Objetivo:** Refinar el dark mode que ya tiene tokens definidos.
**Prerequisito:** Plan 16 (completado)
**Que falta:** Ajustar contraste en todos los componentes, visual regression en dark.

---

## Planes futuros (sin definir)

| Area | Ideas |
|------|-------|
| RAG features | Knowledge gaps, sugerencias automaticas, multi-modal |
| Integraciones | Slack bot, Teams bot, webhooks |
| Admin avanzado | Analytics dashboard, reportes, auditoria avanzada |
| Scale | Postgres, multi-tenant, horizontal scaling |
| Mobile | PWA o app nativa |

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

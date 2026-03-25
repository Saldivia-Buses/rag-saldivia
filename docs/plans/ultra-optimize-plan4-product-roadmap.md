# Plan: Ultra-Optimize Plan 4 — Product Roadmap: UI/UX & Features

> Este documento vive en `docs/plans/ultra-optimize-plan4-product-roadmap.md` dentro de la branch `experimental/ultra-optimize`.
> Se actualiza a medida que se completan las features. Cada tarea completada se marca con fecha.

---

## Contexto

Los Planes 1, 2 y 3 construyeron, verificaron y limpiaron el stack técnico completo. Este plan construye la experiencia de producto encima del stack ya validado.

**Spec fuente:** `docs/superpowers/specs/2026-03-25-product-roadmap-design.md` — 50 features en 4 fases, de menor a mayor dificultad.

**Estado de partida:**
- CSS: tokens genéricos (blanco/negro sin personalidad) en `apps/web/src/app/globals.css`
- Layout: `AppShell.tsx` — sidebar único de 224px. Sin panel contextual secundario.
- Componentes: solo Lucide icons + HTML semántico + CSS variables inline. Sin librería de componentes.
- Dark mode: `@media prefers-color-scheme`. Sin toggle, sin persistencia, sin `.dark` class.
- Tailwind: v4 con `@import "tailwindcss"` puro. Sin `@theme` ni tokens custom.

**Lo que NO cambia:** el stack de auth, DB, RAG proxy, CLI, logger — todo sigue funcionando.

---

## Stack técnico additions

| Fase | Librería | Uso |
|---|---|---|
| 0 | `next-themes` | Dark mode toggle, persistido en localStorage con script anti-FOUC |
| 0 | `shadcn/ui` + `radix-ui/*` | Sistema de componentes (Button, Input, Dialog, Table, etc.) |
| 0 | `clsx`, `tailwind-merge` | Utilidad `cn()` para composición de clases |
| 1 | `react-hotkeys-hook` | Atajos de teclado globales |
| 1 | `sonner` | Toasts (incluido en shadcn) |
| 2 | `cmdk` | Command palette (incluido en shadcn) |
| 2 | `recharts` | Gráficos en analytics dashboard |
| 2 | `driver.js` | Onboarding tour interactivo |
| 3 | `react-pdf` | Preview de PDF inline |
| 3 | `@visx/network` o `d3` | Grafo de similitud entre documentos |
| 3 | `next-auth` v5 | SSO Google / Azure AD |

---

## Seguimiento

Formato de cada tarea: `- [ ] Descripción — estimación`
Al completarla: `- [x] Descripción — completado YYYY-MM-DD`
Cada fase completada genera una entrada en `CHANGELOG.md` antes de hacer commit.

---

## Fase 0 — Fundación *(8-12 hs)*

Objetivo: reemplazar los tokens genéricos por la paleta crema-índigo, instalar shadcn/ui, agregar dark mode toggle y reescribir AppShell con layout tri-columna. Sin esta fase no existe el diseño. Prerequisito de todo.

Principio clave: **AppShell** sigue siendo Server Component. Toda la lógica de estado de UI (zen mode, notificaciones, etc.) vive en `AppShellChrome` — el Client Component wrapper creado en esta fase.

### Fase 0a — Design tokens light/dark *(1-2 hs)*

- [x] Reescribir `apps/web/src/app/globals.css`: paleta crema-índigo, tokens canónicos (`--bg`, `--sidebar-bg`, `--nav-bg`, `--accent`, `--fg`) + aliases de compatibilidad (`--background: var(--bg)`, `--primary: var(--accent)`, etc.) para que los componentes existentes no requieran cambios. Dark mode via clase `.dark`. Directiva `@theme` para Tailwind. `@media print` agregado para export de sesión (Fase 1) — completado 2026-03-25
- [x] Tokens viejos mantenidos como aliases — cero cambios en componentes existentes — completado 2026-03-25
- [x] `bun run test` — 79/79 tests pasan — completado 2026-03-25

Criterio de done: fondo es `#FAFAF9` (crema), no blanco puro. 79 tests pasan.
**Estado: completado 2026-03-25**

### Fase 0b — shadcn/ui setup *(2-3 hs)*

- [x] `bunx shadcn@latest init` falló en modo no interactivo — creado `apps/web/components.json` manualmente (style default, stone, Tailwind v4) — completado 2026-03-25
- [x] `bun add clsx tailwind-merge` + crear `src/lib/utils.ts` con función `cn()` — completado 2026-03-25
- [x] 13 componentes instalados: button, input, textarea, dialog, popover, table, badge, avatar, separator, tooltip, sheet, sonner, command — completado 2026-03-25
- [x] `<Toaster />` de sonner agregado a `layout.tsx` — completado 2026-03-25
- [x] Tokens del spec intactos — shadcn no pisó globals.css — completado 2026-03-25
- [x] `bun run test` — 79/79 tests pasan — completado 2026-03-25

Criterio de done: `import { Button } from "@/components/ui/button"` funciona sin error. Toaster visible en layout.
**Estado: completado 2026-03-25**

### Fase 0c — Dark mode toggle *(1-2 hs)*

- [ ] `cd apps/web && bun add next-themes`
- [ ] Crear `apps/web/src/components/providers.tsx`: Client Component con `<ThemeProvider attribute="class" defaultTheme="light" enableSystem={false} storageKey="rag-theme">`. `attribute="class"` hace que next-themes agregue/remueva la clase `.dark` en `<html>`.
- [ ] Actualizar `apps/web/src/app/layout.tsx`: envolver `{children}` con `<Providers>`, agregar `suppressHydrationWarning` en `<html>` (necesario para que Next.js no se queje del mismatch server/client en la clase dark).
- [ ] Crear `apps/web/src/components/ui/theme-toggle.tsx`: botón con íconos Sun/Moon de lucide-react. Usa `useTheme()` de next-themes. Al clic alterna `light ↔ dark`. Estilos via shadcn `Button variant="ghost" size="icon"`.
- [ ] Verificar en DevTools: el `<head>` incluye el script anti-FOUC inyectado por next-themes. Recargar en dark mode no produce flash blanco.

> **Nota:** next-themes usa localStorage con script bloqueante para evitar FOUC — no cookie. La spec menciona "persistido en cookie para SSR"; la implementación logra el mismo resultado (sin flash) usando el mecanismo nativo de next-themes. Si en el futuro se requiere SSR estricto con cookie, agregar lectura de cookie en middleware.

- [ ] `bun run test` — 79 tests pasan. Commit: `feat(web): dark mode toggle con next-themes + script anti-FOUC — Fase 0c`

Criterio de done: toggle funciona, dark persiste al recargar, sin flash.

### Fase 0d — Dual sidebar layout *(4-5 hs)*

Estructura objetivo: `NavRail (44px) | SecondaryPanel (~168px, contextual) | main (flex-1)`

**Archivos a crear:**
- `apps/web/src/components/layout/NavRail.tsx` — Client Component. Barra de íconos 44px. Fondo `var(--nav-bg)` (#18181B siempre oscuro). Items: Chat, Colecciones, Upload, Audit, Admin, Settings — con `Tooltip` de shadcn. ThemeToggle + logout al fondo.
- `apps/web/src/components/layout/AppShellChrome.tsx` — **Client Component** que envuelve NavRail + SecondaryPanel. Concentra todo el estado de UI (zen mode en Fase 1, notificaciones, hotkeys). Separar de AppShell.tsx permite que AppShell siga siendo Server Component.
- `apps/web/src/components/layout/SecondaryPanel.tsx` — Client Component. Usa `usePathname()` para determinar qué panel mostrar: `/chat` → ChatPanel, `/admin` → AdminPanel, resto → `null` (sin panel).
- `apps/web/src/components/layout/panels/ChatPanel.tsx` — Contenedor del panel de sesiones para rutas `/chat`. Por ahora muestra la estructura (header "Sesiones" + botón nueva sesión + slot). La lista real de sesiones sigue en `SessionList.tsx` y se integra aquí en Fase 1.
- `apps/web/src/components/layout/panels/AdminPanel.tsx` — Nav admin con dos secciones: "Gestión" (Usuarios, Áreas, Permisos, Config RAG) y "Observabilidad" (Sistema — Fase 2 agrega Analytics, Monitoring, Brechas, Historial). Usa `usePathname()` para active state.

**Archivos a modificar:**
- `apps/web/src/components/layout/AppShell.tsx` — reescribir para que sea Server Component puro que solo renderiza `<AppShellChrome user={user}>{children}</AppShellChrome>`.

- [x] Crear `NavRail.tsx` — íconos con tooltips shadcn, active state con `usePathname()`, fondo `var(--nav-bg)`, ancho 44px, ThemeToggle + logout al fondo — completado 2026-03-25
- [x] Crear `panels/ChatPanel.tsx` — header "Sesiones" + botón nueva sesión + slot para SessionList (Fase 1) — completado 2026-03-25
- [x] Crear `panels/AdminPanel.tsx` — secciones "Gestión" y "Observabilidad", extensible para Fase 2 sin rediseño — completado 2026-03-25
- [x] Crear `SecondaryPanel.tsx` — usa `usePathname()`, retorna ChatPanel/AdminPanel/null. Ancho 168px, fondo `var(--sidebar-bg)` — completado 2026-03-25
- [x] Crear `AppShellChrome.tsx` — Client Component con `isZen=false` placeholder (F1.11 lo implementa) — completado 2026-03-25
- [x] Reescribir `AppShell.tsx` — Server Component puro que delega a AppShellChrome — completado 2026-03-25
- [x] `bun run test` — 79/79 tests pasan, sin linter errors — completado 2026-03-25

Criterio de done: `bun run test` pasa (79 tests), layout tri-columna funciona en light y dark, nav contextual funciona por ruta, sin regressions visibles.
**Estado: completado 2026-03-25**

---

## Fase 1 — Quick wins *(14-20 hs)*

Objetivo: 14 features de alto impacto y bajo esfuerzo. Pueden desarrollarse en paralelo una vez que Fase 0 está completa. Al terminar esta fase, la branch está lista para el primer deploy sobre `main`.

Criterio global: las 14 features accesibles y testeadas. `bun run test` pasa.

### F1.5 — Thinking steps visibles *(completado 2026-03-25)*

- [x] Contrato documentado en `useRagStream.ts`: simulación UI-side con timing; cuando el backend exponga eventos SSE `thinking`, se conectan allí — completado 2026-03-25
- [x] Crear `apps/web/src/components/chat/ThinkingSteps.tsx`: steps colapsables, auto-colapsa 1.8s después de terminar — completado 2026-03-25
- [x] Integrar `ThinkingSteps` en `ChatInterface.tsx` — completado 2026-03-25

### F1.6 — Feedback 👍/👎 *(completado 2026-03-25)*

- [x] Verificar `submitFeedback` persiste en `message_feedback` — confirmado 2026-03-25
- [x] Botones migrados a `Button variant="ghost" size="icon"` de shadcn, color índigo/rojo según estado — completado 2026-03-25

### F1.7 — Modos de foco *(completado 2026-03-25)*

- [x] `FOCUS_MODES` definido en `packages/shared/src/schemas.ts` (4 modos: detallado, ejecutivo, técnico, comparativo) — completado 2026-03-25
- [x] `FocusModeSelector.tsx` con pills, persistido en localStorage — completado 2026-03-25
- [x] Prepend del system message en `/api/rag/generate` — completado 2026-03-25
- [x] Tests en `packages/shared/src/__tests__/focus-modes.test.ts` (6 tests) — completado 2026-03-25

### F1.8 — Voz en input *(completado 2026-03-25)*

- [x] `VoiceInput.tsx` con Web Speech API, fallback graceful si no soportado — completado 2026-03-25
- [x] Transcripción en tiempo real al textarea del chat — completado 2026-03-25
- [x] Integrado en `ChatInterface.tsx` junto al botón de envío — completado 2026-03-25

### F1.9 — Export de sesión *(completado 2026-03-25)*

- [x] `apps/web/src/lib/export.ts`: `exportToMarkdown()` con fuentes, `exportToPDF()`, `downloadFile()` — completado 2026-03-25
- [x] `@media print` en `globals.css` — completado 2026-03-25 (Fase 0a)
- [x] `ExportSession.tsx` con Popover PDF / Markdown en el header del chat — completado 2026-03-25
- [x] Tests en `apps/web/src/lib/__tests__/export.test.ts` (8 tests) — completado 2026-03-25

### F1.10 — Respuestas guardadas *(completado 2026-03-25)*

- [x] Tabla `saved_responses` en schema + `init.ts` + migración aplicada — completado 2026-03-25
- [x] `packages/db/src/queries/saved.ts`: `saveResponse`, `unsaveResponse`, `unsaveByMessageId`, `listSavedResponses`, `isSaved` — completado 2026-03-25
- [x] Server Action `actionToggleSaved` en `chat.ts` — completado 2026-03-25
- [x] Ícono bookmark en mensajes asistente, toggle saved/unsaved — completado 2026-03-25
- [x] Página `/saved` con empty state — completado 2026-03-25
- [x] `/saved` en NavRail (ícono `Bookmark`) — completado 2026-03-25
- [x] Tests en `packages/db/src/__tests__/saved.test.ts` (13 tests) — completado 2026-03-25

### F1.11 — Modo Zen *(completado 2026-03-25)*

- [x] `useZenMode.ts`: `Cmd+Shift+Z` toggle, `Esc` cierra — completado 2026-03-25
- [x] `AppShellChrome.tsx` oculta NavRail/SecondaryPanel en modo zen — completado 2026-03-25
- [x] Badge "ESC para salir" `fixed bottom-4 right-4` — completado 2026-03-25

### F1.12 — Notificaciones in-app *(completado 2026-03-25)*

- [x] `GET /api/notifications` — eventos `ingestion.completed`, `ingestion.error`, `user.created` — completado 2026-03-25
- [x] `useNotifications.ts` con polling 30s, sonner toasts, localStorage de IDs vistos — completado 2026-03-25
- [x] Badge rojo unificado (notificaciones + versión nueva) en NavRail — completado 2026-03-25

### F1.13 — Multilenguaje automático *(completado 2026-03-25)*

- [x] `detectLanguageHint()` en `lib/rag/client.ts` — inglés por keywords, no-latinos por regex — completado 2026-03-25
- [x] Inyección en `/api/rag/generate` como system message — completado 2026-03-25
- [x] Tests en `apps/web/src/lib/rag/__tests__/detect-language.test.ts` (13 tests) — completado 2026-03-25

### F1.14 — Atajos de teclado globales *(completado 2026-03-25)*

- [x] `react-hotkeys-hook` instalado — completado 2026-03-25
- [x] `useGlobalHotkeys.ts`: `Cmd+N` → `/chat`. j/k diferidos a Fase 2. — completado 2026-03-25
- [x] Aplicado en `AppShellChrome.tsx` — completado 2026-03-25

### F1.15 — Regenerar respuesta *(completado 2026-03-25)*

- [x] Botón `↻` en mensajes asistente, visible en hover, carga el último query del usuario en el input — completado 2026-03-25

### F1.16 — Copy con formato *(completado 2026-03-25)*

- [x] Botón copy inline con ícono Check al confirmar (2s), copia contenido al portapapeles — completado 2026-03-25

> **Nota:** el plan pedía un CopyButton.tsx separado con Popover (MD/texto/HTML). Implementado inline en ChatInterface.tsx como botón simple que copia markdown raw. El Popover con 3 formatos puede hacerse en Fase 2 si se prioriza.

### F1.17 — Stats de query visibles *(completado 2026-03-25)*

- [x] `responseTimeMs` calculado en `handleSend`, `sourcesCount` de `result.sources.length` — completado 2026-03-25
- [x] Stats `{ms}ms · {N} docs` inline en `ChatInterface.tsx`, visible al hover — completado 2026-03-25

> **Nota:** `tokensUsed` omitido porque el RAG mock no emite evento `usage`. Se conecta cuando el blueprint lo exponga.

### F1.18 — "¿Qué hay de nuevo?" in-app *(completado 2026-03-25)*

- [x] `apps/web/src/lib/changelog.ts`: `parseChangelog()` extraída y testeada — completado 2026-03-25
- [x] `GET /api/changelog` usa `parseChangelog()` + version de package.json — completado 2026-03-25
- [x] `WhatsNewPanel.tsx`: Sheet con entradas del CHANGELOG renderizadas con `marked` — completado 2026-03-25
- [x] Logo "R" en NavRail abre el panel; badge unificado con notificaciones — completado 2026-03-25
- [x] Tests en `apps/web/src/lib/__tests__/changelog.test.ts` (6 tests) — completado 2026-03-25

Criterio global Fase 1: las 14 features accesibles. `bun run test` pasa (126 tests). Ready para primer deploy.
**Estado: completado 2026-03-25**

---

## Fase 2 — Esfuerzo medio *(60-80 hs total)*

Objetivo: 20 features con diseño de componente no trivial o cambios en el backend. Cada una tiene su propio sub-plan en `docs/superpowers/plans/` antes de empezar a codear.

Criterio global: las 20 features completas. Analytics muestra datos reales.

### Índice de features

| # | Feature | Archivos clave | Prerequisito |
|---|---|---|---|
| 19 | **Panel de fuentes / citas** | `SourcesPanel.tsx`, extender `useRagStream.ts` | Fase 1 |
| 20 | **Preguntas relacionadas** | `RelatedQuestions.tsx`, `POST /api/rag/suggest` | #19 |
| 21 | **Multi-colección en query** | `CollectionSelector.tsx`, extender `/api/rag/generate` | — |
| 22 | **Anotar fragmentos** | `AnnotationPopover.tsx`, tabla `annotations` | #19 |
| 23 | **Command palette Cmd+K** | `CommandPalette.tsx`, `cmdk` (ya en shadcn) | F1.14 |
| 24 | **Etiquetas + bulk** | `SessionTags.tsx`, tabla `session_tags`, bulk actions | — |
| 25 | **Compartir sesión** | `ShareDialog.tsx`, tabla `session_shares`, ruta `/share/[token]` | — |
| 26 | **Colecciones desde UI** | reescribir `/collections/page.tsx`, extender `/api/rag/collections` | — |
| 27 | **Chat con doc específico** | botón en UI de colección, sesión con doc pre-seleccionado | #26 |
| 28 | **Templates de query** | `PromptTemplates.tsx`, tabla `prompt_templates`, admin CRUD | — |
| 29 | **Ingestion monitoring mejorado** | `IngestionKanban.tsx`, SSE endpoint para jobs en tiempo real | — |
| 30 | **Analytics dashboard** | `/admin/analytics/page.tsx`, `recharts`, queries de agregación DB | F1.6 |
| 31 | **Brechas de conocimiento** | `/admin/knowledge-gaps/page.tsx`, heurística de baja confianza, TF-IDF simple | #30 |
| 32 | **Historial de colecciones** | `CollectionHistory.tsx`, tabla `collection_history` | #26 |
| 33 | **Informes programados** | tabla `scheduled_reports`, extender worker. Email requiere SMTP externo vía env. | — |
| 34 | **Vista dividida (split view)** | `SplitView.tsx`, estado de dos sesiones paralelas | — |
| 35 | **Drag & drop al chat** | `ChatDropZone.tsx`, colección temporal `@temp-[sessionId]`. **Validar viabilidad primero:** ¿el blueprint soporta colecciones efímeras? ¿worker < 30s para < 100MB? Si no, redefinir como upload normal pre-seleccionado. | — |
| 36 | **Rate limiting por área** | tabla `rate_limits`, check en `/api/rag/generate` | — |
| 37 | **Onboarding interactivo** | `OnboardingTour.tsx`, `driver.js`, flag `onboarding_completed` en tabla `users` | — |
| 38 | **Webhooks salientes** | tabla `webhooks`, extender worker, admin UI | — |

**Estado: pendiente**

---

## Fase 3 — Alta complejidad *(80-120 hs total)*

Objetivo: 12 features que requieren arquitectura nueva, tablas nuevas, o integración con sistemas externos. Cada una tiene su propio sub-plan detallado.

**Hito mínimo de deploy** (criterio para mergear Fase 3): features 39, 40, 41, 47 completas. SSO funciona en staging. El resto (42–46, 48–50) son incrementales post-hito.

### Índice de features

| # | Feature | Descripción técnica | Dependencias |
|---|---|---|---|
| 39 | **Búsqueda universal** | FTS5 de SQLite sobre sesiones + fragmentos + templates. Resultados en Command Palette en tiempo real. | F2.23 (Cmd+K) |
| 40 | **Preview de doc inline** | `react-pdf` (PDF.js). Panel lateral con el PDF y el fragmento exacto resaltado. Requiere que el RAG server exponga el path o bytes del documento. | #19 (panel fuentes) |
| 41 | **Proyectos con contexto** | Entidad `Project`: agrupa sesiones, asigna colecciones, tiene instrucciones custom. Todas las sesiones del proyecto heredan el contexto. Panel en sidebar. | — |
| 42 | **Artifacts panel** | Detectar `:::artifact` en stream o heurística (tabla > 5 cols, bloque markdown > 40 líneas). Panel lateral activable. Guardable, exportable, versionable. | F3.41 |
| 43 | **Bifurcación de conversaciones** | Botón "Bifurcar desde aquí" en mensaje. Nueva sesión con historial hasta ese punto. Indicador de vinculación entre sesiones. | — |
| 44 | **Memoria de usuario** | Tabla `user_memory`. Preferencias inferidas y explícitas. UI para ver/editar. Inyectado en cada query como contexto adicional. | — |
| 45 | **Superficie proactiva** | Job periódico: cruza docs nuevos en colecciones con historial de queries del usuario. Notificaciones "X docs nuevos podrían interesarte". | F1.12, F2.30 |
| 46 | **Grafo de documentos** | Página `/collections/[name]/graph`. D3 o `@visx/network`. Similitud semántica entre docs via embeddings de Milvus. Nodos clicables que abren el doc. | — |
| 47 | **SSO (Google / Azure AD)** | `next-auth` v5 + OIDC/SAML 2.0. Modo mixto: usuarios SSO con `sso_provider`/`sso_subject` nuevos campos en tabla `users`, `password_hash` null. Usuarios con password existentes siguen funcionando. Middleware RBAC y cookies HttpOnly no cambian. | — |
| 48 | **Auto-ingesta externa** | Google Drive, SharePoint, Confluence. OAuth + colección destino + schedule. Worker gestiona sync. | F2.38 (webhooks) |
| 49 | **Bot Slack / Teams** | App Slack/Teams. Llama al API interno con `SYSTEM_API_KEY` + `userId`. Respeta RBAC. | — |
| 50 | **Extracción estructurada** | Usuario define campos (Nombre, Fecha, Monto). Sistema procesa todos los docs de la colección extrayendo esos campos. Resultado exportable como CSV/Excel. | — |

**Estado: pendiente**

---

## Estado global

| Fase | Estado | Fecha |
|------|--------|-------|
| Fase 0 — Fundación (4 features) | ✅ completado | 2026-03-25 |
| Fase 1 — Quick wins (14 features) | ✅ completado | 2026-03-25 |
| Fase 2 — Esfuerzo medio (20 features) | 🔲 pendiente | — |
| Fase 3 — Alta complejidad (12 features) | 🔲 pendiente | — |

## Tiempo total estimado

| Fase | Estimación |
|---|---|
| Fase 0 | 8-12 hs |
| Fase 1 | 14-20 hs |
| Fase 2 | 60-80 hs (con sub-planes por feature) |
| Fase 3 | 80-120 hs (con sub-planes por feature) |
| **Total** | **162-232 hs** |

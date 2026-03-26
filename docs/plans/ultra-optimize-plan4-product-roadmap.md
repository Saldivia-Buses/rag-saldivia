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

Objetivo: 20 features con diseño de componente no trivial o cambios en el backend.

Criterio global: las 20 features completas. Analytics muestra datos reales. `bun run test` pasa.

**Tablas DB nuevas en esta fase:** `annotations`, `session_tags`, `session_shares`, `prompt_templates`, `collection_history`, `scheduled_reports`, `rate_limits`, `webhooks`. Campo nuevo: `onboarding_completed` en `users`.

---

### F2.19 — Panel de fuentes / citas *(2-3 hs)*

**Archivos:**
- Create: `apps/web/src/components/chat/SourcesPanel.tsx`
- Modify: `apps/web/src/hooks/useRagStream.ts` (ya captura sources, exponer al componente)
- Modify: `apps/web/src/components/chat/ChatInterface.tsx`

El stream del RAG Blueprint incluye fuentes en el campo `choices[0].delta.sources` — ya se parsea en `useRagStream.ts`. Esta feature las muestra en un panel colapsable debajo de cada respuesta asistente.

- [x] Crear `SourcesPanel.tsx`: acepta `sources: Array<{ document_name?: string; content?: string; score?: number }>`. Colapsado por default si hay 0 fuentes. Expandido si hay ≥ 1. Muestra: nombre del doc, fragmento relevante (truncado a 150 chars), score como badge si existe.
- [x] En `ChatInterface.tsx`: pasar `result.sources` al mensaje asistente después del stream y renderizar `<SourcesPanel sources={msg.sources} />` debajo del contenido.
- [x] Test: `SourcesPanel` con 0 fuentes no renderiza nada (o renderiza estado vacío). Con 2 fuentes muestra 2 items.
- [x] Commit: `feat(chat): panel de fuentes y citas debajo de respuestas — f2.19`

---

### F2.20 — Preguntas relacionadas *(2-3 hs)*

**Archivos:**
- Create: `apps/web/src/app/api/rag/suggest/route.ts`
- Create: `apps/web/src/components/chat/RelatedQuestions.tsx`
- Modify: `apps/web/src/components/chat/ChatInterface.tsx`

- [x] Crear `POST /api/rag/suggest`: recibe `{ query, collection, lastResponse }`. En modo MOCK_RAG retorna 3 sugerencias hardcodeadas. En modo real: hace query al RAG con prompt `"Given this Q&A, suggest 3 follow-up questions in the same language:"`. Retorna `{ questions: string[] }`.
- [x] Crear `RelatedQuestions.tsx`: muestra 3-4 chips con las sugerencias. Al hacer clic en uno llama al callback `onSelect(question)` que lo pone en el input del chat.
- [x] En `ChatInterface.tsx`: después de cada respuesta completada (phase === "done"), llamar al endpoint `/api/rag/suggest` y renderizar `<RelatedQuestions>` debajo del SourcesPanel.
- [x] Test unitario: el endpoint retorna array de strings, no vacío en modo mock.
- [x] Commit: `feat(chat): preguntas relacionadas despues de cada respuesta — f2.20`

---

### F2.21 — Multi-colección en query *(2-3 hs)*

**Archivos:**
- Create: `apps/web/src/components/chat/CollectionSelector.tsx`
- Modify: `apps/web/src/hooks/useRagStream.ts`
- Modify: `apps/web/src/app/api/rag/generate/route.ts`

- [x] Crear `CollectionSelector.tsx`: multi-checkbox dropdown (Popover + shadcn) que lista las colecciones disponibles del usuario. Estado persiste en `localStorage["selected_collections"]`. Mostrar debajo del input, junto a FocusModeSelector.
- [x] Extender `useRagStream.ts`: aceptar `collections: string[]` en options y enviarlo en el body del fetch como `collection_names`.
- [x] En `/api/rag/generate`: si `body.collection_names` es array, verificar acceso a cada colección y pasarlas al RAG server. El Blueprint acepta múltiples colecciones via `collection_name` con coma o array según versión.
- [x] Test: con `collection_names: ["col-a", "col-b"]` el route verifica acceso a ambas.
- [x] Commit: `feat(chat): selector multi-coleccion en query — f2.21`

---

### F2.22 — Anotar fragmentos de respuesta *(2-3 hs)*

**Archivos:**
- Create: `apps/web/src/components/chat/AnnotationPopover.tsx`
- Create: `packages/db/src/schema.ts` — tabla `annotations`
- Create: `packages/db/src/queries/annotations.ts`
- Create: `apps/web/src/app/actions/chat.ts` — action `actionSaveAnnotation`

- [x] Agregar tabla al schema: `annotations` (id, userId, sessionId, messageId, selectedText, note, createdAt).
- [x] Agregar a `init.ts` el `CREATE TABLE IF NOT EXISTS annotations`.
- [x] Crear `packages/db/src/queries/annotations.ts`: `saveAnnotation`, `listAnnotationsBySession`.
- [x] Crear `AnnotationPopover.tsx`: al seleccionar texto en un mensaje asistente (`onMouseUp`/`onSelectionChange`), mostrar un popover flotante con opciones: "💾 Guardar fragmento", "❓ Preguntar sobre esto" (pone el texto en el input), "💬 Comentar" (abre input para nota libre).
- [x] Test unitario: `saveAnnotation` persiste correctamente en memoria.
- [x] Commit: `feat(chat): anotar fragmentos de respuesta — f2.22`

---

### F2.23 — Command palette Cmd+K *(3-4 hs)*

**Archivos:**
- Create: `apps/web/src/components/layout/CommandPalette.tsx`
- Modify: `apps/web/src/components/layout/AppShellChrome.tsx`
- Modify: `apps/web/src/hooks/useGlobalHotkeys.ts`

`cmdk` ya instalado vía shadcn (componente `Command`).

- [x] Crear `CommandPalette.tsx`: Dialog con `Command` de shadcn. Grupos de comandos:
  - **Navegar:** Nueva sesión, /chat, /collections, /upload, /saved, /admin
  - **Sesiones:** buscar en historial de sesiones del usuario (llamada a `listSessionsByUser` filtrada por texto)
  - **Modo de foco:** cambiar entre Detallado / Ejecutivo / Técnico / Comparativo
  - **Sistema:** Tema claro/oscuro, Modo Zen
- [x] En `useGlobalHotkeys.ts`: agregar `Cmd+K` para abrir/cerrar la paleta.
- [x] En `AppShellChrome.tsx`: controlar estado `paletteOpen` + renderizar `<CommandPalette>`.
- [x] Test: el componente renderiza sin errores. Los grupos de comandos tienen al menos un item.
- [x] Commit: `feat(web): command palette cmd+k con grupos de acciones — f2.23`

---

### F2.24 — Etiquetas en sesiones + bulk actions *(3-4 hs)*

**Archivos:**
- Create: `packages/db/src/schema.ts` — tabla `session_tags`
- Create: `packages/db/src/queries/tags.ts`
- Modify: `apps/web/src/components/chat/SessionList.tsx`
- Create: `apps/web/src/app/actions/chat.ts` — actions de tags y bulk

- [x] Tabla `session_tags` en schema + `init.ts`: (sessionId FK chat_sessions, tag TEXT, PRIMARY KEY (sessionId, tag)).
- [x] `packages/db/src/queries/tags.ts`: `addTag(sessionId, tag)`, `removeTag(sessionId, tag)`, `listTagsByUser(userId)`.
- [x] En `SessionList.tsx`: input inline para agregar tags a cada sesión (click en "+" junto a la sesión). Filtro por tag en el header de la lista.
- [x] Checkbox en cada sesión para selección múltiple. Toolbar de bulk actions cuando hay ≥ 1 seleccionada: "Exportar todo", "Eliminar todo".
- [x] Test unitario: `addTag` y `removeTag` en memoria.
- [x] Commit: `feat(chat): etiquetas en sesiones + bulk actions — f2.24`

---

### F2.25 — Compartir sesión *(2-3 hs)*

**Archivos:**
- Create: `packages/db/src/schema.ts` — tabla `session_shares`
- Create: `packages/db/src/queries/shares.ts`
- Create: `apps/web/src/components/chat/ShareDialog.tsx`
- Create: `apps/web/src/app/(public)/share/[token]/page.tsx`
- Create: `apps/web/src/app/api/share/route.ts`

- [x] Tabla `session_shares`: (id UUID PK, sessionId FK, userId FK, token TEXT UNIQUE, expiresAt INTEGER, createdAt INTEGER).
- [x] `packages/db/src/queries/shares.ts`: `createShare(sessionId, userId, ttlDays?)`, `getShareByToken(token)`, `revokeShare(id, userId)`.
- [x] `GET /api/share?token=X`: valida token, verifica expiración, retorna session + messages. Ruta pública (en middleware `PUBLIC_ROUTES`).
- [x] `POST /api/share`: crear token (32 bytes hex con `crypto.randomBytes`). Auth requerida.
- [x] `ShareDialog.tsx`: muestra el link generado, botón copiar, fecha de expiración, aviso de privacidad ("No compartir sesiones con información sensible"). Admin puede configurar si la feature está habilitada.
- [x] Página `/share/[token]` (route group `(public)` sin auth): muestra la sesión en modo read-only. Sin AppShell. Si token expirado/inválido → 404.
- [x] Test unitario: `createShare` genera token de 64 chars hex. `getShareByToken` retorna null para token inválido.
- [x] Commit: `feat(web): compartir sesion con token publico — f2.25`

---

### F2.26 — Colecciones desde UI *(2-3 hs)*

**Archivos:**
- Rewrite: `apps/web/src/app/(app)/collections/page.tsx` (actualmente es page básica)
- Create: `apps/web/src/components/collections/CollectionsList.tsx`
- Create: `apps/web/src/app/api/rag/collections/[name]/route.ts`
- Modify: `apps/web/src/app/api/rag/collections/route.ts`

- [x] Extender `GET /api/rag/collections`: incluir lista de documentos y estado de ingesta por colección (llamar al RAG server para obtener metadata).
- [x] Agregar `POST /api/rag/collections` (solo admin): crear colección via RAG server API. `DELETE /api/rag/collections/[name]`: eliminar con confirmación.
- [x] Reescribir `/collections/page.tsx` como Server Component que fetcha la lista de colecciones con metadata.
- [x] Crear `CollectionsList.tsx`: tabla con nombre, descripción, cantidad de docs, estado de ingesta, acciones (ver docs, chat, eliminar).
- [x] Test: `POST /api/rag/collections` sin auth retorna 401. Con auth de user retorna 403.
- [x] Commit: `feat(collections): pagina de colecciones con CRUD desde UI — f2.26`

---

### F2.27 — Chat con documento específico *(1-2 hs)*

**Archivos:**
- Modify: `apps/web/src/components/collections/CollectionsList.tsx`
- Modify: `apps/web/src/app/actions/chat.ts`

Prerequisito: F2.26 completada.

- [x] En la lista de documentos de cada colección, agregar botón "Preguntar sobre este doc".
- [x] Al hacer clic: `actionCreateSession({ collection, title: "Chat: ${docName}", initialContext: docName })` y navegar a la nueva sesión.
- [x] El system prompt de la sesión incluye `"Only use information from document: ${docName}"`.
- [x] Commit: `feat(collections): chat con documento especifico — f2.27`

---

### F2.28 — Templates de query *(2-3 hs)*

**Archivos:**
- Create: `packages/db/src/schema.ts` — tabla `prompt_templates`
- Create: `packages/db/src/queries/templates.ts`
- Create: `apps/web/src/components/chat/PromptTemplates.tsx`
- Create: `apps/web/src/app/(app)/admin/templates/page.tsx`

- [x] Tabla `prompt_templates`: (id INTEGER PK, title TEXT, prompt TEXT, focusMode TEXT, createdBy FK users, active BOOLEAN, createdAt).
- [x] `packages/db/src/queries/templates.ts`: `listActiveTemplates()`, `createTemplate(data)`, `deleteTemplate(id)`.
- [x] `PromptTemplates.tsx`: dropdown tipo "Seleccionar template" sobre el input del chat. Al elegir uno, pone el prompt en el input y setea el focusMode recomendado.
- [x] Página `/admin/templates`: tabla de templates activos con acciones crear/eliminar. Solo admins.
- [x] Test unitario: `listActiveTemplates` retorna solo templates con `active = true`.
- [x] Commit: `feat(chat): templates de query con admin CRUD — f2.28`

---

### F2.29 — Ingestion monitoring mejorado *(3-4 hs)*

**Archivos:**
- Create: `apps/web/src/components/admin/IngestionKanban.tsx`
- Create: `apps/web/src/app/api/admin/ingestion/stream/route.ts`
- Modify: `apps/web/src/app/(app)/admin/ingestion/page.tsx` (crear si no existe)

- [x] Crear `GET /api/admin/ingestion/stream`: SSE endpoint que hace polling cada 3s a la DB y emite los jobs actualizados. Solo admin. Usar `ReadableStream` + `TransformStream` como en `/api/rag/generate`.
- [x] Crear `IngestionKanban.tsx`: columnas Pendiente / En progreso / Completado / Error. Cards con filename, colección, progreso (barra), tier, tiempo transcurrido. Botón "Retry" en cards de Error (llama a `PATCH /api/admin/ingestion/[id]` con `{ action: "retry" }`). Detalle de error expandible en un `Dialog`.
- [x] Conectar el SSE al kanban para actualización en tiempo real (reemplaza la tabla estática actual).
- [x] Agregar `PATCH /api/admin/ingestion/[id]` para action retry: cambia status a `pending`, reset `retryCount++`.
- [x] Commit: `feat(admin): ingestion kanban con progreso en tiempo real — f2.29`

---

### F2.30 — Analytics dashboard *(4-5 hs)*

**Archivos:**
- Create: `apps/web/src/app/(app)/admin/analytics/page.tsx`
- Create: `apps/web/src/components/admin/AnalyticsDashboard.tsx`
- Create: `apps/web/src/app/api/admin/analytics/route.ts`

Prerequisito: F1.6 (tabla `message_feedback` con datos). `bun add recharts`.

- [x] Crear `GET /api/admin/analytics`: queries de agregación sobre `events`, `chat_sessions`, `message_feedback`. Retorna:
  - `queriesByDay`: `SELECT DATE(ts/1000,'unixepoch') as day, COUNT(*) FROM events WHERE type='rag.stream_started' GROUP BY day ORDER BY day DESC LIMIT 30`
  - `topCollections`: `SELECT collection_name, COUNT(*) FROM events WHERE type='rag.stream_started' GROUP BY collection_name ORDER BY COUNT(*) DESC LIMIT 10`
  - `feedbackDistribution`: `SELECT rating, COUNT(*) FROM message_feedback GROUP BY rating`
  - `topUsers`: `SELECT user_id, COUNT(*) as queries FROM events WHERE type='rag.stream_started' GROUP BY user_id ORDER BY queries DESC LIMIT 10`
- [x] `bun add recharts` en `apps/web`.
- [x] Crear `AnalyticsDashboard.tsx` Client Component: LineChart (queries/día), BarChart (top colecciones), PieChart (feedback +/-), tabla (top usuarios).
- [x] Página `/admin/analytics`: Server Component que fetcha datos y pasa al dashboard.
- [x] Agregar link "Analytics" en `AdminPanel.tsx` sección Observabilidad.
- [x] Test: el endpoint retorna las 4 claves esperadas con arrays.
- [x] Commit: `feat(admin): analytics dashboard con recharts — f2.30`

---

### F2.31 — Brechas de conocimiento *(2-3 hs)*

**Archivos:**
- Create: `apps/web/src/app/(app)/admin/knowledge-gaps/page.tsx`
- Create: `apps/web/src/app/api/admin/knowledge-gaps/route.ts`

Prerequisito: F2.30.

Definición operativa (del spec): un query es "brecha" si la respuesta del asistente:
- No tiene fuentes (`sources` vacío o null), O
- Es < 80 tokens Y contiene patterns de incertidumbre ("no encuentro", "no tengo información", "no sé", "no encontré", "I don't know", "I couldn't find").

- [x] Crear `GET /api/admin/knowledge-gaps`: busca en `chat_messages` con `role='assistant'` mensajes cortos (< 80 words) con keywords de incertidumbre, LEFT JOIN con `events` para obtener el query original. Retorna los últimos 100, agrupados por colección.
- [x] Página `/admin/knowledge-gaps`: tabla con columnas: Query original, Respuesta (truncada), Colección, Fecha, Acción ("Ingestar más docs sobre esto" → lleva a /upload).
- [x] Botón "Exportar CSV" que descarga los gaps como CSV.
- [x] Agregar link "Brechas" en `AdminPanel.tsx` sección Observabilidad.
- [x] Commit: `feat(admin): brechas de conocimiento detectadas por heuristica — f2.31`

---

### F2.32 — Historial de colecciones *(2-3 hs)*

**Archivos:**
- Create: `packages/db/src/schema.ts` — tabla `collection_history`
- Create: `packages/db/src/queries/collection-history.ts`
- Create: `apps/web/src/components/collections/CollectionHistory.tsx`
- Modify: `apps/web/src/workers/ingestion.ts`

- [x] Tabla `collection_history`: (id TEXT PK, collection TEXT, userId FK, action TEXT enum('added','removed'), docCount INTEGER, sizeBytes INTEGER, createdAt INTEGER).
- [x] En `ingestion.ts` worker: al completar un job exitosamente, insertar un registro en `collection_history`.
- [x] `collection-history.ts`: `listHistoryByCollection(collection)`, `recordIngestionEvent(data)`.
- [x] `CollectionHistory.tsx`: timeline vertical de commits de ingesta por colección. Cada entrada muestra fecha, usuario, docs agregados, tamaño.
- [x] Integrar en la página de detalle de colección (`/collections/[name]`).
- [x] Test unitario: `recordIngestionEvent` persiste con los campos correctos.
- [x] Commit: `feat(collections): historial de ingestas como commits de coleccion — f2.32`

---

### F2.33 — Informes programados *(3-4 hs)*

**Archivos:**
- Create: `packages/db/src/schema.ts` — tabla `scheduled_reports`
- Create: `packages/db/src/queries/reports.ts`
- Create: `apps/web/src/app/(app)/admin/reports/page.tsx`
- Modify: `apps/web/src/workers/ingestion.ts`

- [x] Tabla `scheduled_reports`: (id TEXT PK, userId FK, query TEXT, collection TEXT, schedule TEXT, destination enum('saved','email'), email TEXT nullable, active BOOLEAN, lastRun INTEGER, nextRun INTEGER, createdAt INTEGER).
- [x] `packages/db/src/queries/reports.ts`: `listActiveReports()`, `createReport(data)`, `updateLastRun(id, nextRun)`.
- [x] Extender el worker `ingestion.ts`: cada 5 minutos, buscar `scheduled_reports` cuyo `nextRun <= now`. Para cada uno: ejecutar el query via el cliente RAG interno, guardar resultado en `saved_responses` (destino=saved) o enviar email (destino=email, requiere `SMTP_HOST` configurado). Si SMTP no configurado: log de warning, no falla.
- [x] Página `/admin/reports`: crear/listar/eliminar informes. Formulario: query, colección, schedule (select: diario/semanal/mensual), destino.
- [x] Agregar link "Informes" en `AdminPanel.tsx`.
- [x] Test unitario: `listActiveReports` retorna solo reports con `active = true` y `nextRun <= Date.now()`.
- [x] Commit: `feat(admin): informes programados con destino saved o email — f2.33`

---

### F2.34 — Vista dividida (split view) *(2-3 hs)*

**Archivos:**
- Create: `apps/web/src/components/chat/SplitView.tsx`
- Modify: `apps/web/src/app/(app)/chat/[id]/page.tsx`
- Modify: `apps/web/src/components/chat/ChatInterface.tsx`

- [x] Crear `SplitView.tsx`: wrapper que acepta dos `<ChatInterface>` side by side. Botón "Split" en el header del chat que activa el modo dividido. En modo dividido: los dos paneles tienen `width: 50%`, cada uno con su propia sesión independiente.
- [x] En la página `/chat/[id]`: si `searchParams.split=true`, cargar dos sesiones (la actual + la última del usuario) y renderizar `<SplitView>`.
- [x] Botón "Cerrar split" vuelve al modo single.
- [x] Commit: `feat(chat): vista dividida con dos sesiones paralelas — f2.34`

---

### F2.35 — Drag & drop al chat *(2-3 hs)*

**Pre-validación de viabilidad requerida antes de implementar.**

> Antes de codear: verificar que el NVIDIA RAG Blueprint v2.5.0 soporta crear colecciones efímeras (sin TTL fijo en Milvus). Si no es viable, esta feature se redefine como "upload normal con pre-selección de sesión".

**Archivos:**
- Create: `apps/web/src/components/chat/ChatDropZone.tsx`
- Modify: `apps/web/src/components/chat/ChatInterface.tsx`
- Modify: `apps/web/src/app/api/upload/route.ts`

- [x] **Paso previo:** revisar docs del Blueprint v2.5.0 para confirmar viabilidad de colecciones efímeras en Milvus. Si no: implementar como upload normal.
- [x] Crear `ChatDropZone.tsx`: overlay de drag & drop sobre el área del chat. `onDrop`: sube el archivo a `POST /api/upload?collection=@temp-${sessionId}&temp=true`. Muestra progreso inline.
- [x] Extender `POST /api/upload`: si `?temp=true`, usar `@temp-${sessionId}` como nombre de colección. Crear la colección si no existe.
- [x] Al completar la ingesta, auto-seleccionar la colección temporal en el selector del chat.
- [x] Commit: `feat(chat): drag and drop de pdf al chat con coleccion temporal — f2.35`

---

### F2.36 — Rate limiting por área/usuario *(2-3 hs)*

**Archivos:**
- Create: `packages/db/src/schema.ts` — tabla `rate_limits`
- Create: `packages/db/src/queries/rate-limits.ts`
- Modify: `apps/web/src/app/api/rag/generate/route.ts`
- Create: `apps/web/src/app/(app)/admin/settings/rate-limits/page.tsx`

- [x] Tabla `rate_limits`: (id INTEGER PK, targetType enum('user','area'), targetId INTEGER, maxQueriesPerHour INTEGER, active BOOLEAN, createdAt INTEGER).
- [x] `rate-limits.ts`: `getRateLimit(userId, areaIds)` retorna el límite aplicable (user-level tiene precedencia sobre area-level). `countQueriesLastHour(userId)` cuenta eventos `rag.stream_started` de las últimas 3600s.
- [x] En `/api/rag/generate`: antes de procesar, verificar si el usuario tiene rate limit activo. Si `countQueriesLastHour >= maxQueriesPerHour` → retornar 429 con mensaje descriptivo.
- [x] Página `/admin/settings/rate-limits`: tabla de límites configurados. Formulario para agregar (tipo, ID, max/hora). Solo admins.
- [x] Test unitario: `getRateLimit` retorna null cuando no hay límite configurado.
- [x] Commit: `feat(admin): rate limiting por area/usuario — f2.36`

---

### F2.37 — Onboarding interactivo *(2-3 hs)*

**Archivos:**
- Create: `apps/web/src/components/onboarding/OnboardingTour.tsx`
- Modify: `packages/db/src/schema.ts` — campo `onboardingCompleted` en `users`
- Modify: `apps/web/src/app/(app)/layout.tsx`

- [x] Agregar campo `onboarding_completed INTEGER NOT NULL DEFAULT 0` a `users` en schema + `init.ts`.
- [x] `bun add driver.js` en `apps/web`.
- [x] Crear `OnboardingTour.tsx`: al montar, si `user.onboardingCompleted === false`, inicializar driver.js con 5 pasos:
  1. NavRail: "Esta es la barra de navegación"
  2. Chat: "Aquí hacés tus preguntas"
  3. FocusModeSelector: "Elegí el modo de respuesta"
  4. Colecciones: "Tus documentos organizados en colecciones"
  5. Settings: "Configurá tu perfil"
  Al terminar o saltar: llamar a Server Action `actionCompleteOnboarding()`.
- [x] Server Action `actionCompleteOnboarding`: `UPDATE users SET onboarding_completed = 1 WHERE id = userId`.
- [x] En `/settings`: botón "Reiniciar tour de onboarding" que llama a `actionResetOnboarding()`.
- [x] En `(app)/layout.tsx`: pasar `user.onboardingCompleted` al componente. Renderizar `<OnboardingTour>` si false.
- [x] Commit: `feat(web): onboarding interactivo con driver.js — f2.37`

---

### F2.38 — Webhooks salientes *(3-4 hs)*

**Archivos:**
- Create: `packages/db/src/schema.ts` — tabla `webhooks`
- Create: `packages/db/src/queries/webhooks.ts`
- Create: `apps/web/src/app/(app)/admin/webhooks/page.tsx`
- Modify: `apps/web/src/workers/ingestion.ts`
- Create: `apps/web/src/lib/webhook.ts`

- [x] Tabla `webhooks`: (id TEXT PK, userId FK, url TEXT, events TEXT (JSON array de event types), secret TEXT (para HMAC), active BOOLEAN, createdAt INTEGER).
- [x] `packages/db/src/queries/webhooks.ts`: `listActiveWebhooks()`, `listWebhooksByEvent(eventType)`, `createWebhook(data)`, `deleteWebhook(id, userId)`.
- [x] `apps/web/src/lib/webhook.ts`: función `dispatchWebhook(webhook, payload)` que hace `POST` al URL con header `X-Signature: HMAC-SHA256(secret, JSON.stringify(payload))` y timeout de 5s. No relanza errores — logea si falla.
- [x] En `ingestion.ts` worker: al cambiar estado a `done` o `error`, llamar `listWebhooksByEvent("ingestion.completed"/"ingestion.error")` y dispatch para cada uno.
- [x] En `/api/rag/generate` route: si la respuesta tiene baja confianza (sin fuentes), dispatch a webhooks de tipo `query.low_confidence`.
- [x] Página `/admin/webhooks`: tabla + formulario crear (URL, eventos a escuchar, secret). Solo admins.
- [x] Test unitario: `dispatchWebhook` no lanza excepciones si el URL no responde.
- [x] Commit: `feat(admin): webhooks salientes con HMAC — f2.38`

---

Criterio global Fase 2: las 20 features completas. Analytics muestra datos reales. `bun run test` pasa.
**Estado: completado 2026-03-25**

---

## Fase 3 — Alta complejidad *(80-120 hs total)*

Objetivo: 12 features que requieren arquitectura nueva, tablas nuevas, o integración con sistemas externos.

Criterio global: hito mínimo (features 39, 40, 41, 47) listas para deploy. El resto (42–46, 48–50) son incrementales post-hito.

**Tablas DB nuevas en esta fase:** `projects`, `project_sessions`, `project_collections`, `user_memory`, `external_sources`. Campos nuevos: `sso_provider`/`sso_subject` en `users`, `forked_from` en `chat_sessions`. Tablas FTS5 virtuales: `sessions_fts`, `messages_fts`.

---

### F3.39 — Búsqueda universal *(3-4 hs)* 🔑 hito mínimo

**Archivos:**
- Modify: `packages/db/src/init.ts` — tablas FTS5 virtuales
- Create: `packages/db/src/queries/search.ts`
- Create: `apps/web/src/app/api/search/route.ts`
- Modify: `apps/web/src/components/layout/CommandPalette.tsx`

SQLite FTS5 soporta búsqueda full-text nativa. Se indexan `chat_sessions.title`, `chat_messages.content`, `prompt_templates.title + prompt`, `saved_responses.content`.

- [ ] Agregar a `init.ts` las tablas FTS5 virtuales:
  ```sql
  CREATE VIRTUAL TABLE IF NOT EXISTS sessions_fts USING fts5(id, title, content=chat_sessions, content_rowid=rowid);
  CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(id, content, content=chat_messages, content_rowid=rowid);
  ```
- [ ] Agregar triggers en `init.ts` para mantener FTS sincronizado con inserts/updates/deletes en `chat_sessions` y `chat_messages`.
- [ ] Crear `packages/db/src/queries/search.ts`: función `universalSearch(query, userId, limit)` que consulta las tablas FTS5 y retorna resultados unificados con tipo (`session` | `message` | `template` | `saved`) y snippet resaltado.
- [ ] Crear `GET /api/search?q=...`: llama a `universalSearch` con el userId del token. Retorna máximo 20 resultados.
- [ ] En `CommandPalette.tsx`: agregar grupo "Resultados" que llama al endpoint en tiempo real mientras el usuario escribe (debounce 300ms). Cada resultado muestra tipo, título y snippet.
- [ ] Test unitario: `universalSearch("")` retorna array vacío. `universalSearch("test")` con datos en DB retorna resultados.
- [ ] Commit: `feat(web): busqueda universal fts5 en command palette — f3.39`

---

### F3.40 — Preview de doc inline *(3-4 hs)* 🔑 hito mínimo

**Archivos:**
- Create: `apps/web/src/components/chat/DocPreviewPanel.tsx`
- Modify: `apps/web/src/components/chat/SourcesPanel.tsx`
- Create: `apps/web/src/app/api/rag/document/[name]/route.ts`

Prerequisito: F2.19 (SourcesPanel). El RAG Blueprint expone documentos en `/v1/documents/{name}` — si no, usar fallback de descarga del blob.

- [ ] `bun add react-pdf` en `apps/web`. Agregar `react-pdf` a `transpilePackages` en `next.config.ts`.
- [ ] Crear `GET /api/rag/document/[name]`: hace proxy al RAG server para obtener el PDF como stream. Solo usuarios con acceso a la colección. Retorna el blob con `Content-Type: application/pdf`.
- [ ] Crear `DocPreviewPanel.tsx`: Sheet lateral (shadcn) que renderiza el PDF con `react-pdf`. Acepta props `documentName` y `highlightText`. El fragmento destacado se resalta con CSS sobre el canvas de PDF.js.
- [ ] En `SourcesPanel.tsx`: al hacer clic en el nombre de un documento, abrir `DocPreviewPanel` con el nombre y fragmento correspondiente.
- [ ] Test: `DocPreviewPanel` con `documentName=""` no crashea. Renderiza un estado de loading.

> **Nota de viabilidad:** Si el Blueprint no expone `/v1/documents/{name}`, el panel muestra el fragmento de texto del source sin renderizado PDF. Documentar en comentario.

- [ ] Commit: `feat(chat): preview de doc inline con react-pdf — f3.40`

---

### F3.41 — Proyectos con contexto *(5-7 hs)* 🔑 hito mínimo

**Archivos:**
- Modify: `packages/db/src/schema.ts` — tablas `projects`, `project_sessions`, `project_collections`
- Create: `packages/db/src/queries/projects.ts`
- Create: `apps/web/src/app/(app)/projects/` — páginas
- Create: `apps/web/src/components/layout/panels/ProjectsPanel.tsx`
- Modify: `apps/web/src/components/layout/SecondaryPanel.tsx`
- Modify: `apps/web/src/app/api/rag/generate/route.ts`

- [ ] Agregar al schema:
  - `projects`: (id UUID PK, userId FK, name TEXT, description TEXT, instructions TEXT — system prompt adicional, createdAt INTEGER)
  - `project_sessions`: (projectId FK, sessionId FK, PRIMARY KEY compuesto)
  - `project_collections`: (projectId FK, collectionName TEXT, PRIMARY KEY compuesto)
- [ ] Migrar tabla en la DB real con `bun -e` inline.
- [ ] Crear `packages/db/src/queries/projects.ts`: `createProject`, `listProjects(userId)`, `getProject(id)`, `addSessionToProject`, `addCollectionToProject`, `deleteProject`.
- [ ] Crear páginas:
  - `/projects` — lista de proyectos del usuario
  - `/projects/[id]` — detalle: sesiones asignadas, colecciones, instrucciones, botón nueva sesión en contexto
- [ ] Crear `ProjectsPanel.tsx`: panel secundario para rutas `/projects`. Lista de proyectos con íconos, botón crear.
- [ ] En `SecondaryPanel.tsx`: agregar `/projects` → `ProjectsPanel`.
- [ ] Agregar `/projects` al NavRail (ícono `FolderKanban`).
- [ ] En `/api/rag/generate`: si la sesión pertenece a un proyecto, prepend las `instructions` del proyecto como system message adicional.
- [ ] Test unitario: `createProject` y `listProjects` contra SQLite en memoria.
- [ ] Commit: `feat(web): proyectos con contexto — tablas, paginas, panel sidebar — f3.41`

---

### F3.42 — Artifacts panel *(3-4 hs)*

**Archivos:**
- Create: `apps/web/src/components/chat/ArtifactsPanel.tsx`
- Modify: `apps/web/src/hooks/useRagStream.ts`
- Modify: `apps/web/src/components/chat/ChatInterface.tsx`

Prerequisito: F3.41.

El stream puede incluir bloques `:::artifact{type="document|table|code"}` o heurística: bloque markdown de >40 líneas, o tabla con >5 columnas.

- [ ] En `useRagStream.ts`: detectar `:::artifact` en el delta acumulado. Si se detecta, extraer el bloque y emitirlo via callback `onArtifact`. Si no hay marcador, aplicar heurística: bloque de código >40 líneas o tabla >5 columnas.
- [ ] Agregar `onArtifact?: (artifact: { type: string; content: string }) => void` a `UseRagStreamOptions`.
- [ ] Crear `ArtifactsPanel.tsx`: Sheet lateral que muestra el artifact. Para `code`: resaltado de sintaxis con `<pre>`. Para `table`: renderiza el markdown como tabla. Para `document`: texto enriquecido. Botones: Guardar (a `saved_responses`), Exportar (Markdown/CSV según tipo).
- [ ] En `ChatInterface.tsx`: estado `currentArtifact` alimentado por `onArtifact`. Badge "Artifact" en el header que abre el panel.
- [ ] Commit: `feat(chat): artifacts panel — deteccion en stream y panel lateral — f3.42`

---

### F3.43 — Bifurcación de conversaciones *(2-3 hs)*

**Archivos:**
- Modify: `packages/db/src/schema.ts` — campo `forked_from` en `chat_sessions`
- Modify: `packages/db/src/init.ts`
- Modify: `apps/web/src/app/actions/chat.ts`
- Modify: `apps/web/src/components/chat/ChatInterface.tsx`
- Modify: `apps/web/src/components/chat/SessionList.tsx`

- [ ] Agregar campo `forked_from TEXT REFERENCES chat_sessions(id) ON DELETE SET NULL` a `chat_sessions` en schema + `init.ts`. Migrar con `ALTER TABLE`.
- [ ] Server Action `actionForkSession(sessionId, upToMessageId)`: copia la sesión y los mensajes hasta `upToMessageId`. Setea `forked_from = sessionId` en la nueva sesión. Retorna el id de la nueva sesión.
- [ ] En `ChatInterface.tsx`: botón "Bifurcar desde aquí" (ícono `GitBranch`) en cada mensaje del asistente, visible en hover. Al clic llama a `actionForkSession` y navega a la nueva sesión.
- [ ] En `SessionList.tsx`: mostrar badge visual (ícono `GitBranch` pequeño) en sesiones con `forked_from` no null. Al hover, tooltip "Bifurcada de: [título original]".
- [ ] Test unitario: `actionForkSession` crea sesión con `forked_from` y copia solo los mensajes hasta el punto indicado.
- [ ] Commit: `feat(chat): bifurcacion de conversaciones — f3.43`

---

### F3.44 — Memoria de usuario *(3-4 hs)*

**Archivos:**
- Modify: `packages/db/src/schema.ts` — tabla `user_memory`
- Create: `packages/db/src/queries/memory.ts`
- Create: `apps/web/src/app/(app)/settings/memory/page.tsx`
- Modify: `apps/web/src/app/api/rag/generate/route.ts`
- Modify: `apps/web/src/components/settings/SettingsClient.tsx`

- [ ] Tabla `user_memory`: (id INTEGER PK, userId FK, key TEXT, value TEXT, source enum('explicit','inferred'), createdAt, updatedAt. UNIQUE(userId, key)).
- [ ] `packages/db/src/queries/memory.ts`: `setMemory(userId, key, value, source)`, `getMemory(userId)`, `deleteMemory(userId, key)`.
- [ ] Migrar tabla.
- [ ] En `/api/rag/generate`: llamar `getMemory(userId)`, si hay entradas construir un system message con el contexto: `"User preferences: [key1: value1, key2: value2]"`. Prepend antes del array de mensajes.
- [ ] Crear `/settings/memory`: tabla de preferencias con columnas clave/valor/fuente. Botón eliminar por fila. Formulario para agregar preferencia explícita.
- [ ] Agregar tab "Memoria" en `SettingsClient.tsx` que navega a `/settings/memory`.
- [ ] Test unitario: `setMemory` + `getMemory` con datos correctos en memoria.
- [ ] Commit: `feat(web): memoria de usuario inyectada en queries — f3.44`

---

### F3.45 — Superficie proactiva *(2-3 hs)*

**Archivos:**
- Modify: `apps/web/src/workers/ingestion.ts`
- Modify: `apps/web/src/app/api/notifications/route.ts`

Prerequisito: F1.12 (notificaciones), F2.30 (analytics).

Cuando se completa una ingesta de docs nuevos, el worker cruza los términos del documento con los queries recientes del usuario (tabla `events` tipo `rag.stream_started`) y, si hay coincidencia semántica simple (keywords en común), genera una notificación del tipo `proactive.docs_available`.

- [ ] En `ingestion.ts` worker al completar job: llamar `checkProactiveSurface(collection, userId)`.
- [ ] Implementar `checkProactiveSurface`: obtener queries recientes del usuario en esa colección (últimos 30 días, eventos `rag.stream_started`). Extraer keywords del filename/collection. Si hay solapamiento → insertar evento `proactive.docs_available` en tabla `events` con payload `{ collection, filename, matchedQueries: N }`.
- [ ] En `/api/notifications/route.ts`: agregar `"proactive.docs_available"` a los tipos de notificación retornados.
- [ ] Commit: `feat(web): superficie proactiva — notificaciones de docs nuevos relevantes — f3.45`

---

### F3.46 — Grafo de documentos *(4-5 hs)*

**Archivos:**
- Create: `apps/web/src/app/(app)/collections/[name]/graph/page.tsx`
- Create: `apps/web/src/components/collections/DocumentGraph.tsx`
- Create: `apps/web/src/app/api/collections/[name]/embeddings/route.ts`

- [ ] `bun add d3` en `apps/web`. Agregar a `transpilePackages`.
- [ ] Crear `GET /api/collections/[name]/embeddings`: llama al RAG server para obtener embeddings de los documentos de la colección. Si el Blueprint no los expone directamente, retorna una estructura simulada con similitud aleatoria para MVP. Retorna `{ nodes: [{id, name}], edges: [{source, target, weight}] }`.
- [ ] Crear `DocumentGraph.tsx` (Client Component): visualización D3 force-directed. Nodos = documentos, aristas = similitud semántica (peso). Nodos clicables que navegan a la URL del documento o abren el `DocPreviewPanel` (F3.40). Zoom + drag. Colores por cluster (calculado con simple thresholding del weight).
- [ ] Crear página `/collections/[name]/graph`: Server Component que fetcha la colección y renderiza `<DocumentGraph />`.
- [ ] Agregar botón "Ver grafo" en `CollectionsList.tsx` (F2.26).
- [ ] Commit: `feat(collections): grafo de documentos con d3 — f3.46`

---

### F3.47 — SSO (Google / Azure AD) *(6-8 hs)* 🔑 hito mínimo

**Archivos:**
- Modify: `packages/db/src/schema.ts` — campos `sso_provider`, `sso_subject` en `users`
- Modify: `packages/db/src/init.ts`
- Create: `apps/web/src/app/api/auth/[...nextauth]/route.ts`
- Create: `apps/web/src/components/auth/SSOButton.tsx`
- Modify: `apps/web/src/app/(auth)/login/page.tsx`
- Modify: `apps/web/src/middleware.ts`

Coexistencia: usuarios SSO usan NextAuth session; usuarios con password usan el JWT propio. El middleware maneja ambos.

- [ ] `bun add next-auth@beta @auth/core` en `apps/web`.
- [ ] Agregar campos al schema: `sso_provider TEXT` y `sso_subject TEXT` en tabla `users`. UNIQUE(sso_provider, sso_subject). `password_hash` queda null para usuarios SSO. Migrar con `ALTER TABLE`.
- [ ] Crear `apps/web/src/app/api/auth/[...nextauth]/route.ts` con NextAuth v5:
  - Provider Google (`GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`)
  - Provider Azure AD (`AZURE_AD_CLIENT_ID`, `AZURE_AD_CLIENT_SECRET`, `AZURE_AD_TENANT_ID`)
  - En `authorize` callback: buscar usuario por `sso_provider + sso_subject`. Si no existe, crear con `role: "user"`, `active: true`. Si existe pero `active: false` → denegar.
  - Emitir JWT propio (jose) con los mismos claims que el flujo de password, para compatibilidad con el middleware RBAC existente.
- [ ] Crear `SSOButton.tsx`: botón "Continuar con Google / Microsoft" usando `signIn()` de NextAuth.
- [ ] En `login/page.tsx`: agregar `<SSOButton>` debajo del formulario de email/password, separado por un "o".
- [ ] En `middleware.ts`: verificar también la cookie de sesión de NextAuth (`next-auth.session-token`) como alternativa al JWT propio. Si el token NextAuth es válido, extraer claims del mismo y continuar el flujo RBAC.
- [ ] Agregar al `.env.example`: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `AZURE_AD_CLIENT_ID`, `AZURE_AD_CLIENT_SECRET`, `AZURE_AD_TENANT_ID`, `NEXTAUTH_SECRET`, `NEXTAUTH_URL`.
- [ ] Test: login con credenciales de password sigue funcionando. Usuario SSO creado en DB al primer login.
- [ ] Commit: `feat(auth): sso google y azure ad con nextauth v5 — f3.47`

---

### F3.48 — Auto-ingesta externa *(6-8 hs)*

**Archivos:**
- Modify: `packages/db/src/schema.ts` — tabla `external_sources`
- Create: `packages/db/src/queries/external-sources.ts`
- Create: `apps/web/src/app/(app)/admin/external-sources/page.tsx`
- Create: `apps/web/src/workers/external-sync.ts`
- Modify: `apps/web/src/workers/ingestion.ts`

Prerequisito: F2.38 (webhooks).

- [ ] Tabla `external_sources`: (id UUID PK, userId FK, provider enum('google_drive','sharepoint','confluence'), name TEXT, credentials TEXT (JSON cifrado), collectionDest TEXT, schedule TEXT, active BOOLEAN, lastSync INTEGER, createdAt INTEGER).
- [ ] `external-sources.ts`: `createSource`, `listSources(userId)`, `listActiveSources()`, `updateLastSync(id)`.
- [ ] Crear `apps/web/src/workers/external-sync.ts`: worker separado. Cada 5 minutos fetcha `listActiveSources()`. Para cada source:
  - Google Drive: usa `googleapis` SDK — listar archivos modificados desde `lastSync`, descargar y subir a `/api/upload`.
  - SharePoint: usa `@microsoft/microsoft-graph-client` — mismo patrón.
  - Confluence: usa Confluence REST API — exportar como PDF y subir.
  - `bun add googleapis @microsoft/microsoft-graph-client`
- [ ] Página `/admin/external-sources`: formulario para configurar fuentes (provider, credenciales OAuth, colección destino, schedule). Lista de fuentes activas con estado y último sync.
- [ ] Agregar link "Fuentes externas" en `AdminPanel.tsx`.
- [ ] Commit: `feat(admin): auto-ingesta desde google drive sharepoint confluence — f3.48`

---

### F3.49 — Bot Slack / Teams *(4-6 hs)*

**Archivos:**
- Create: `apps/web/src/app/api/slack/route.ts`
- Create: `apps/web/src/app/api/teams/route.ts`
- Create: `apps/web/src/app/(app)/admin/integrations/page.tsx`

- [ ] Crear `POST /api/slack`: handler de eventos Slack (slash command o bot mention). Extrae el `userId` de un mapping Slack→sistema almacenado en DB (tabla `bot_user_mappings`) o permite autenticación via token de API. Llama al endpoint interno `/api/rag/generate` con `SYSTEM_API_KEY` y el userId resuelto. Retorna la respuesta al canal de Slack via `response_url` o `chat.postMessage`. Respeta RBAC (si el usuario no tiene acceso, retorna mensaje de error al canal).
- [ ] Crear `POST /api/teams`: mismo patrón para Microsoft Teams via Adaptive Cards.
- [ ] Agregar tabla `bot_user_mappings` (slack_user_id/teams_user_id → system userId) al schema + `init.ts`.
- [ ] Página `/admin/integrations`: configuración de tokens de Slack/Teams, mapeo de usuarios. Solo admins.
- [ ] Agregar al `.env.example`: `SLACK_BOT_TOKEN`, `SLACK_SIGNING_SECRET`, `TEAMS_BOT_ID`, `TEAMS_BOT_PASSWORD`.
- [ ] Commit: `feat(web): bot slack y teams respetando rbac — f3.49`

---

### F3.50 — Extracción estructurada *(4-5 hs)*

**Archivos:**
- Create: `apps/web/src/app/(app)/extract/page.tsx`
- Create: `apps/web/src/app/api/extract/route.ts`
- Create: `apps/web/src/components/extract/ExtractionWizard.tsx`

- [ ] Crear `POST /api/extract`: recibe `{ collection, fields: [{name, description}] }`. Para cada documento de la colección, hace un query al RAG server con el prompt: `"Extract the following fields from this document: [fields]. Return JSON only."`. Agrega los resultados a una tabla en memoria.
- [ ] Crear `ExtractionWizard.tsx` (Client Component): wizard de 3 pasos:
  1. Seleccionar colección
  2. Definir campos (nombre + descripción, dinámico con `+`)
  3. Ejecutar y ver resultados en tabla
  Botón "Exportar CSV" al final.
- [ ] Crear página `/extract` accesible para todos los usuarios autenticados.
- [ ] Agregar `/extract` al NavRail (ícono `Table2`).
- [ ] Commit: `feat(web): extraccion estructurada a tabla exportable como csv — f3.50`

---

Criterio global Fase 3: features 39, 40, 41, 47 completas y testeadas. SSO funciona en staging. `bun run test` pasa. Las demás (42–46, 48–50) se completan en sub-sprints post-hito.
**Estado: pendiente**

---

## Estado global

| Fase | Estado | Fecha |
|------|--------|-------|
| Fase 0 — Fundación (4 features) | ✅ completado | 2026-03-25 |
| Fase 1 — Quick wins (14 features) | ✅ completado | 2026-03-25 |
| Fase 2 — Esfuerzo medio (20 features) | ✅ completado | 2026-03-25 |
| Fase 3 — Alta complejidad (12 features) | 🔲 pendiente | — |

## Tiempo total estimado

| Fase | Estimación |
|---|---|
| Fase 0 | 8-12 hs |
| Fase 1 | 14-20 hs |
| Fase 2 | 60-80 hs (con sub-planes por feature) |
| Fase 3 | 80-120 hs (con sub-planes por feature) |
| **Total** | **162-232 hs** |

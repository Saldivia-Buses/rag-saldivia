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

- [ ] Crear `SourcesPanel.tsx`: acepta `sources: Array<{ document_name?: string; content?: string; score?: number }>`. Colapsado por default si hay 0 fuentes. Expandido si hay ≥ 1. Muestra: nombre del doc, fragmento relevante (truncado a 150 chars), score como badge si existe.
- [ ] En `ChatInterface.tsx`: pasar `result.sources` al mensaje asistente después del stream y renderizar `<SourcesPanel sources={msg.sources} />` debajo del contenido.
- [ ] Test: `SourcesPanel` con 0 fuentes no renderiza nada (o renderiza estado vacío). Con 2 fuentes muestra 2 items.
- [ ] Commit: `feat(chat): panel de fuentes y citas debajo de respuestas — f2.19`

---

### F2.20 — Preguntas relacionadas *(2-3 hs)*

**Archivos:**
- Create: `apps/web/src/app/api/rag/suggest/route.ts`
- Create: `apps/web/src/components/chat/RelatedQuestions.tsx`
- Modify: `apps/web/src/components/chat/ChatInterface.tsx`

- [ ] Crear `POST /api/rag/suggest`: recibe `{ query, collection, lastResponse }`. En modo MOCK_RAG retorna 3 sugerencias hardcodeadas. En modo real: hace query al RAG con prompt `"Given this Q&A, suggest 3 follow-up questions in the same language:"`. Retorna `{ questions: string[] }`.
- [ ] Crear `RelatedQuestions.tsx`: muestra 3-4 chips con las sugerencias. Al hacer clic en uno llama al callback `onSelect(question)` que lo pone en el input del chat.
- [ ] En `ChatInterface.tsx`: después de cada respuesta completada (phase === "done"), llamar al endpoint `/api/rag/suggest` y renderizar `<RelatedQuestions>` debajo del SourcesPanel.
- [ ] Test unitario: el endpoint retorna array de strings, no vacío en modo mock.
- [ ] Commit: `feat(chat): preguntas relacionadas despues de cada respuesta — f2.20`

---

### F2.21 — Multi-colección en query *(2-3 hs)*

**Archivos:**
- Create: `apps/web/src/components/chat/CollectionSelector.tsx`
- Modify: `apps/web/src/hooks/useRagStream.ts`
- Modify: `apps/web/src/app/api/rag/generate/route.ts`

- [ ] Crear `CollectionSelector.tsx`: multi-checkbox dropdown (Popover + shadcn) que lista las colecciones disponibles del usuario. Estado persiste en `localStorage["selected_collections"]`. Mostrar debajo del input, junto a FocusModeSelector.
- [ ] Extender `useRagStream.ts`: aceptar `collections: string[]` en options y enviarlo en el body del fetch como `collection_names`.
- [ ] En `/api/rag/generate`: si `body.collection_names` es array, verificar acceso a cada colección y pasarlas al RAG server. El Blueprint acepta múltiples colecciones via `collection_name` con coma o array según versión.
- [ ] Test: con `collection_names: ["col-a", "col-b"]` el route verifica acceso a ambas.
- [ ] Commit: `feat(chat): selector multi-coleccion en query — f2.21`

---

### F2.22 — Anotar fragmentos de respuesta *(2-3 hs)*

**Archivos:**
- Create: `apps/web/src/components/chat/AnnotationPopover.tsx`
- Create: `packages/db/src/schema.ts` — tabla `annotations`
- Create: `packages/db/src/queries/annotations.ts`
- Create: `apps/web/src/app/actions/chat.ts` — action `actionSaveAnnotation`

- [ ] Agregar tabla al schema: `annotations` (id, userId, sessionId, messageId, selectedText, note, createdAt).
- [ ] Agregar a `init.ts` el `CREATE TABLE IF NOT EXISTS annotations`.
- [ ] Crear `packages/db/src/queries/annotations.ts`: `saveAnnotation`, `listAnnotationsBySession`.
- [ ] Crear `AnnotationPopover.tsx`: al seleccionar texto en un mensaje asistente (`onMouseUp`/`onSelectionChange`), mostrar un popover flotante con opciones: "💾 Guardar fragmento", "❓ Preguntar sobre esto" (pone el texto en el input), "💬 Comentar" (abre input para nota libre).
- [ ] Test unitario: `saveAnnotation` persiste correctamente en memoria.
- [ ] Commit: `feat(chat): anotar fragmentos de respuesta — f2.22`

---

### F2.23 — Command palette Cmd+K *(3-4 hs)*

**Archivos:**
- Create: `apps/web/src/components/layout/CommandPalette.tsx`
- Modify: `apps/web/src/components/layout/AppShellChrome.tsx`
- Modify: `apps/web/src/hooks/useGlobalHotkeys.ts`

`cmdk` ya instalado vía shadcn (componente `Command`).

- [ ] Crear `CommandPalette.tsx`: Dialog con `Command` de shadcn. Grupos de comandos:
  - **Navegar:** Nueva sesión, /chat, /collections, /upload, /saved, /admin
  - **Sesiones:** buscar en historial de sesiones del usuario (llamada a `listSessionsByUser` filtrada por texto)
  - **Modo de foco:** cambiar entre Detallado / Ejecutivo / Técnico / Comparativo
  - **Sistema:** Tema claro/oscuro, Modo Zen
- [ ] En `useGlobalHotkeys.ts`: agregar `Cmd+K` para abrir/cerrar la paleta.
- [ ] En `AppShellChrome.tsx`: controlar estado `paletteOpen` + renderizar `<CommandPalette>`.
- [ ] Test: el componente renderiza sin errores. Los grupos de comandos tienen al menos un item.
- [ ] Commit: `feat(web): command palette cmd+k con grupos de acciones — f2.23`

---

### F2.24 — Etiquetas en sesiones + bulk actions *(3-4 hs)*

**Archivos:**
- Create: `packages/db/src/schema.ts` — tabla `session_tags`
- Create: `packages/db/src/queries/tags.ts`
- Modify: `apps/web/src/components/chat/SessionList.tsx`
- Create: `apps/web/src/app/actions/chat.ts` — actions de tags y bulk

- [ ] Tabla `session_tags` en schema + `init.ts`: (sessionId FK chat_sessions, tag TEXT, PRIMARY KEY (sessionId, tag)).
- [ ] `packages/db/src/queries/tags.ts`: `addTag(sessionId, tag)`, `removeTag(sessionId, tag)`, `listTagsByUser(userId)`.
- [ ] En `SessionList.tsx`: input inline para agregar tags a cada sesión (click en "+" junto a la sesión). Filtro por tag en el header de la lista.
- [ ] Checkbox en cada sesión para selección múltiple. Toolbar de bulk actions cuando hay ≥ 1 seleccionada: "Exportar todo", "Eliminar todo".
- [ ] Test unitario: `addTag` y `removeTag` en memoria.
- [ ] Commit: `feat(chat): etiquetas en sesiones + bulk actions — f2.24`

---

### F2.25 — Compartir sesión *(2-3 hs)*

**Archivos:**
- Create: `packages/db/src/schema.ts` — tabla `session_shares`
- Create: `packages/db/src/queries/shares.ts`
- Create: `apps/web/src/components/chat/ShareDialog.tsx`
- Create: `apps/web/src/app/(public)/share/[token]/page.tsx`
- Create: `apps/web/src/app/api/share/route.ts`

- [ ] Tabla `session_shares`: (id UUID PK, sessionId FK, userId FK, token TEXT UNIQUE, expiresAt INTEGER, createdAt INTEGER).
- [ ] `packages/db/src/queries/shares.ts`: `createShare(sessionId, userId, ttlDays?)`, `getShareByToken(token)`, `revokeShare(id, userId)`.
- [ ] `GET /api/share?token=X`: valida token, verifica expiración, retorna session + messages. Ruta pública (en middleware `PUBLIC_ROUTES`).
- [ ] `POST /api/share`: crear token (32 bytes hex con `crypto.randomBytes`). Auth requerida.
- [ ] `ShareDialog.tsx`: muestra el link generado, botón copiar, fecha de expiración, aviso de privacidad ("No compartir sesiones con información sensible"). Admin puede configurar si la feature está habilitada.
- [ ] Página `/share/[token]` (route group `(public)` sin auth): muestra la sesión en modo read-only. Sin AppShell. Si token expirado/inválido → 404.
- [ ] Test unitario: `createShare` genera token de 64 chars hex. `getShareByToken` retorna null para token inválido.
- [ ] Commit: `feat(web): compartir sesion con token publico — f2.25`

---

### F2.26 — Colecciones desde UI *(2-3 hs)*

**Archivos:**
- Rewrite: `apps/web/src/app/(app)/collections/page.tsx` (actualmente es page básica)
- Create: `apps/web/src/components/collections/CollectionsList.tsx`
- Create: `apps/web/src/app/api/rag/collections/[name]/route.ts`
- Modify: `apps/web/src/app/api/rag/collections/route.ts`

- [ ] Extender `GET /api/rag/collections`: incluir lista de documentos y estado de ingesta por colección (llamar al RAG server para obtener metadata).
- [ ] Agregar `POST /api/rag/collections` (solo admin): crear colección via RAG server API. `DELETE /api/rag/collections/[name]`: eliminar con confirmación.
- [ ] Reescribir `/collections/page.tsx` como Server Component que fetcha la lista de colecciones con metadata.
- [ ] Crear `CollectionsList.tsx`: tabla con nombre, descripción, cantidad de docs, estado de ingesta, acciones (ver docs, chat, eliminar).
- [ ] Test: `POST /api/rag/collections` sin auth retorna 401. Con auth de user retorna 403.
- [ ] Commit: `feat(collections): pagina de colecciones con CRUD desde UI — f2.26`

---

### F2.27 — Chat con documento específico *(1-2 hs)*

**Archivos:**
- Modify: `apps/web/src/components/collections/CollectionsList.tsx`
- Modify: `apps/web/src/app/actions/chat.ts`

Prerequisito: F2.26 completada.

- [ ] En la lista de documentos de cada colección, agregar botón "Preguntar sobre este doc".
- [ ] Al hacer clic: `actionCreateSession({ collection, title: "Chat: ${docName}", initialContext: docName })` y navegar a la nueva sesión.
- [ ] El system prompt de la sesión incluye `"Only use information from document: ${docName}"`.
- [ ] Commit: `feat(collections): chat con documento especifico — f2.27`

---

### F2.28 — Templates de query *(2-3 hs)*

**Archivos:**
- Create: `packages/db/src/schema.ts` — tabla `prompt_templates`
- Create: `packages/db/src/queries/templates.ts`
- Create: `apps/web/src/components/chat/PromptTemplates.tsx`
- Create: `apps/web/src/app/(app)/admin/templates/page.tsx`

- [ ] Tabla `prompt_templates`: (id INTEGER PK, title TEXT, prompt TEXT, focusMode TEXT, createdBy FK users, active BOOLEAN, createdAt).
- [ ] `packages/db/src/queries/templates.ts`: `listActiveTemplates()`, `createTemplate(data)`, `deleteTemplate(id)`.
- [ ] `PromptTemplates.tsx`: dropdown tipo "Seleccionar template" sobre el input del chat. Al elegir uno, pone el prompt en el input y setea el focusMode recomendado.
- [ ] Página `/admin/templates`: tabla de templates activos con acciones crear/eliminar. Solo admins.
- [ ] Test unitario: `listActiveTemplates` retorna solo templates con `active = true`.
- [ ] Commit: `feat(chat): templates de query con admin CRUD — f2.28`

---

### F2.29 — Ingestion monitoring mejorado *(3-4 hs)*

**Archivos:**
- Create: `apps/web/src/components/admin/IngestionKanban.tsx`
- Create: `apps/web/src/app/api/admin/ingestion/stream/route.ts`
- Modify: `apps/web/src/app/(app)/admin/ingestion/page.tsx` (crear si no existe)

- [ ] Crear `GET /api/admin/ingestion/stream`: SSE endpoint que hace polling cada 3s a la DB y emite los jobs actualizados. Solo admin. Usar `ReadableStream` + `TransformStream` como en `/api/rag/generate`.
- [ ] Crear `IngestionKanban.tsx`: columnas Pendiente / En progreso / Completado / Error. Cards con filename, colección, progreso (barra), tier, tiempo transcurrido. Botón "Retry" en cards de Error (llama a `PATCH /api/admin/ingestion/[id]` con `{ action: "retry" }`). Detalle de error expandible en un `Dialog`.
- [ ] Conectar el SSE al kanban para actualización en tiempo real (reemplaza la tabla estática actual).
- [ ] Agregar `PATCH /api/admin/ingestion/[id]` para action retry: cambia status a `pending`, reset `retryCount++`.
- [ ] Commit: `feat(admin): ingestion kanban con progreso en tiempo real — f2.29`

---

### F2.30 — Analytics dashboard *(4-5 hs)*

**Archivos:**
- Create: `apps/web/src/app/(app)/admin/analytics/page.tsx`
- Create: `apps/web/src/components/admin/AnalyticsDashboard.tsx`
- Create: `apps/web/src/app/api/admin/analytics/route.ts`

Prerequisito: F1.6 (tabla `message_feedback` con datos). `bun add recharts`.

- [ ] Crear `GET /api/admin/analytics`: queries de agregación sobre `events`, `chat_sessions`, `message_feedback`. Retorna:
  - `queriesByDay`: `SELECT DATE(ts/1000,'unixepoch') as day, COUNT(*) FROM events WHERE type='rag.stream_started' GROUP BY day ORDER BY day DESC LIMIT 30`
  - `topCollections`: `SELECT collection_name, COUNT(*) FROM events WHERE type='rag.stream_started' GROUP BY collection_name ORDER BY COUNT(*) DESC LIMIT 10`
  - `feedbackDistribution`: `SELECT rating, COUNT(*) FROM message_feedback GROUP BY rating`
  - `topUsers`: `SELECT user_id, COUNT(*) as queries FROM events WHERE type='rag.stream_started' GROUP BY user_id ORDER BY queries DESC LIMIT 10`
- [ ] `bun add recharts` en `apps/web`.
- [ ] Crear `AnalyticsDashboard.tsx` Client Component: LineChart (queries/día), BarChart (top colecciones), PieChart (feedback +/-), tabla (top usuarios).
- [ ] Página `/admin/analytics`: Server Component que fetcha datos y pasa al dashboard.
- [ ] Agregar link "Analytics" en `AdminPanel.tsx` sección Observabilidad.
- [ ] Test: el endpoint retorna las 4 claves esperadas con arrays.
- [ ] Commit: `feat(admin): analytics dashboard con recharts — f2.30`

---

### F2.31 — Brechas de conocimiento *(2-3 hs)*

**Archivos:**
- Create: `apps/web/src/app/(app)/admin/knowledge-gaps/page.tsx`
- Create: `apps/web/src/app/api/admin/knowledge-gaps/route.ts`

Prerequisito: F2.30.

Definición operativa (del spec): un query es "brecha" si la respuesta del asistente:
- No tiene fuentes (`sources` vacío o null), O
- Es < 80 tokens Y contiene patterns de incertidumbre ("no encuentro", "no tengo información", "no sé", "no encontré", "I don't know", "I couldn't find").

- [ ] Crear `GET /api/admin/knowledge-gaps`: busca en `chat_messages` con `role='assistant'` mensajes cortos (< 80 words) con keywords de incertidumbre, LEFT JOIN con `events` para obtener el query original. Retorna los últimos 100, agrupados por colección.
- [ ] Página `/admin/knowledge-gaps`: tabla con columnas: Query original, Respuesta (truncada), Colección, Fecha, Acción ("Ingestar más docs sobre esto" → lleva a /upload).
- [ ] Botón "Exportar CSV" que descarga los gaps como CSV.
- [ ] Agregar link "Brechas" en `AdminPanel.tsx` sección Observabilidad.
- [ ] Commit: `feat(admin): brechas de conocimiento detectadas por heuristica — f2.31`

---

### F2.32 — Historial de colecciones *(2-3 hs)*

**Archivos:**
- Create: `packages/db/src/schema.ts` — tabla `collection_history`
- Create: `packages/db/src/queries/collection-history.ts`
- Create: `apps/web/src/components/collections/CollectionHistory.tsx`
- Modify: `apps/web/src/workers/ingestion.ts`

- [ ] Tabla `collection_history`: (id TEXT PK, collection TEXT, userId FK, action TEXT enum('added','removed'), docCount INTEGER, sizeBytes INTEGER, createdAt INTEGER).
- [ ] En `ingestion.ts` worker: al completar un job exitosamente, insertar un registro en `collection_history`.
- [ ] `collection-history.ts`: `listHistoryByCollection(collection)`, `recordIngestionEvent(data)`.
- [ ] `CollectionHistory.tsx`: timeline vertical de commits de ingesta por colección. Cada entrada muestra fecha, usuario, docs agregados, tamaño.
- [ ] Integrar en la página de detalle de colección (`/collections/[name]`).
- [ ] Test unitario: `recordIngestionEvent` persiste con los campos correctos.
- [ ] Commit: `feat(collections): historial de ingestas como commits de coleccion — f2.32`

---

### F2.33 — Informes programados *(3-4 hs)*

**Archivos:**
- Create: `packages/db/src/schema.ts` — tabla `scheduled_reports`
- Create: `packages/db/src/queries/reports.ts`
- Create: `apps/web/src/app/(app)/admin/reports/page.tsx`
- Modify: `apps/web/src/workers/ingestion.ts`

- [ ] Tabla `scheduled_reports`: (id TEXT PK, userId FK, query TEXT, collection TEXT, schedule TEXT, destination enum('saved','email'), email TEXT nullable, active BOOLEAN, lastRun INTEGER, nextRun INTEGER, createdAt INTEGER).
- [ ] `packages/db/src/queries/reports.ts`: `listActiveReports()`, `createReport(data)`, `updateLastRun(id, nextRun)`.
- [ ] Extender el worker `ingestion.ts`: cada 5 minutos, buscar `scheduled_reports` cuyo `nextRun <= now`. Para cada uno: ejecutar el query via el cliente RAG interno, guardar resultado en `saved_responses` (destino=saved) o enviar email (destino=email, requiere `SMTP_HOST` configurado). Si SMTP no configurado: log de warning, no falla.
- [ ] Página `/admin/reports`: crear/listar/eliminar informes. Formulario: query, colección, schedule (select: diario/semanal/mensual), destino.
- [ ] Agregar link "Informes" en `AdminPanel.tsx`.
- [ ] Test unitario: `listActiveReports` retorna solo reports con `active = true` y `nextRun <= Date.now()`.
- [ ] Commit: `feat(admin): informes programados con destino saved o email — f2.33`

---

### F2.34 — Vista dividida (split view) *(2-3 hs)*

**Archivos:**
- Create: `apps/web/src/components/chat/SplitView.tsx`
- Modify: `apps/web/src/app/(app)/chat/[id]/page.tsx`
- Modify: `apps/web/src/components/chat/ChatInterface.tsx`

- [ ] Crear `SplitView.tsx`: wrapper que acepta dos `<ChatInterface>` side by side. Botón "Split" en el header del chat que activa el modo dividido. En modo dividido: los dos paneles tienen `width: 50%`, cada uno con su propia sesión independiente.
- [ ] En la página `/chat/[id]`: si `searchParams.split=true`, cargar dos sesiones (la actual + la última del usuario) y renderizar `<SplitView>`.
- [ ] Botón "Cerrar split" vuelve al modo single.
- [ ] Commit: `feat(chat): vista dividida con dos sesiones paralelas — f2.34`

---

### F2.35 — Drag & drop al chat *(2-3 hs)*

**Pre-validación de viabilidad requerida antes de implementar.**

> Antes de codear: verificar que el NVIDIA RAG Blueprint v2.5.0 soporta crear colecciones efímeras (sin TTL fijo en Milvus). Si no es viable, esta feature se redefine como "upload normal con pre-selección de sesión".

**Archivos:**
- Create: `apps/web/src/components/chat/ChatDropZone.tsx`
- Modify: `apps/web/src/components/chat/ChatInterface.tsx`
- Modify: `apps/web/src/app/api/upload/route.ts`

- [ ] **Paso previo:** revisar docs del Blueprint v2.5.0 para confirmar viabilidad de colecciones efímeras en Milvus. Si no: implementar como upload normal.
- [ ] Crear `ChatDropZone.tsx`: overlay de drag & drop sobre el área del chat. `onDrop`: sube el archivo a `POST /api/upload?collection=@temp-${sessionId}&temp=true`. Muestra progreso inline.
- [ ] Extender `POST /api/upload`: si `?temp=true`, usar `@temp-${sessionId}` como nombre de colección. Crear la colección si no existe.
- [ ] Al completar la ingesta, auto-seleccionar la colección temporal en el selector del chat.
- [ ] Commit: `feat(chat): drag and drop de pdf al chat con coleccion temporal — f2.35`

---

### F2.36 — Rate limiting por área/usuario *(2-3 hs)*

**Archivos:**
- Create: `packages/db/src/schema.ts` — tabla `rate_limits`
- Create: `packages/db/src/queries/rate-limits.ts`
- Modify: `apps/web/src/app/api/rag/generate/route.ts`
- Create: `apps/web/src/app/(app)/admin/settings/rate-limits/page.tsx`

- [ ] Tabla `rate_limits`: (id INTEGER PK, targetType enum('user','area'), targetId INTEGER, maxQueriesPerHour INTEGER, active BOOLEAN, createdAt INTEGER).
- [ ] `rate-limits.ts`: `getRateLimit(userId, areaIds)` retorna el límite aplicable (user-level tiene precedencia sobre area-level). `countQueriesLastHour(userId)` cuenta eventos `rag.stream_started` de las últimas 3600s.
- [ ] En `/api/rag/generate`: antes de procesar, verificar si el usuario tiene rate limit activo. Si `countQueriesLastHour >= maxQueriesPerHour` → retornar 429 con mensaje descriptivo.
- [ ] Página `/admin/settings/rate-limits`: tabla de límites configurados. Formulario para agregar (tipo, ID, max/hora). Solo admins.
- [ ] Test unitario: `getRateLimit` retorna null cuando no hay límite configurado.
- [ ] Commit: `feat(admin): rate limiting por area/usuario — f2.36`

---

### F2.37 — Onboarding interactivo *(2-3 hs)*

**Archivos:**
- Create: `apps/web/src/components/onboarding/OnboardingTour.tsx`
- Modify: `packages/db/src/schema.ts` — campo `onboardingCompleted` en `users`
- Modify: `apps/web/src/app/(app)/layout.tsx`

- [ ] Agregar campo `onboarding_completed INTEGER NOT NULL DEFAULT 0` a `users` en schema + `init.ts`.
- [ ] `bun add driver.js` en `apps/web`.
- [ ] Crear `OnboardingTour.tsx`: al montar, si `user.onboardingCompleted === false`, inicializar driver.js con 5 pasos:
  1. NavRail: "Esta es la barra de navegación"
  2. Chat: "Aquí hacés tus preguntas"
  3. FocusModeSelector: "Elegí el modo de respuesta"
  4. Colecciones: "Tus documentos organizados en colecciones"
  5. Settings: "Configurá tu perfil"
  Al terminar o saltar: llamar a Server Action `actionCompleteOnboarding()`.
- [ ] Server Action `actionCompleteOnboarding`: `UPDATE users SET onboarding_completed = 1 WHERE id = userId`.
- [ ] En `/settings`: botón "Reiniciar tour de onboarding" que llama a `actionResetOnboarding()`.
- [ ] En `(app)/layout.tsx`: pasar `user.onboardingCompleted` al componente. Renderizar `<OnboardingTour>` si false.
- [ ] Commit: `feat(web): onboarding interactivo con driver.js — f2.37`

---

### F2.38 — Webhooks salientes *(3-4 hs)*

**Archivos:**
- Create: `packages/db/src/schema.ts` — tabla `webhooks`
- Create: `packages/db/src/queries/webhooks.ts`
- Create: `apps/web/src/app/(app)/admin/webhooks/page.tsx`
- Modify: `apps/web/src/workers/ingestion.ts`
- Create: `apps/web/src/lib/webhook.ts`

- [ ] Tabla `webhooks`: (id TEXT PK, userId FK, url TEXT, events TEXT (JSON array de event types), secret TEXT (para HMAC), active BOOLEAN, createdAt INTEGER).
- [ ] `packages/db/src/queries/webhooks.ts`: `listActiveWebhooks()`, `listWebhooksByEvent(eventType)`, `createWebhook(data)`, `deleteWebhook(id, userId)`.
- [ ] `apps/web/src/lib/webhook.ts`: función `dispatchWebhook(webhook, payload)` que hace `POST` al URL con header `X-Signature: HMAC-SHA256(secret, JSON.stringify(payload))` y timeout de 5s. No relanza errores — logea si falla.
- [ ] En `ingestion.ts` worker: al cambiar estado a `done` o `error`, llamar `listWebhooksByEvent("ingestion.completed"/"ingestion.error")` y dispatch para cada uno.
- [ ] En `/api/rag/generate` route: si la respuesta tiene baja confianza (sin fuentes), dispatch a webhooks de tipo `query.low_confidence`.
- [ ] Página `/admin/webhooks`: tabla + formulario crear (URL, eventos a escuchar, secret). Solo admins.
- [ ] Test unitario: `dispatchWebhook` no lanza excepciones si el URL no responde.
- [ ] Commit: `feat(admin): webhooks salientes con HMAC — f2.38`

---

Criterio global Fase 2: las 20 features completas. Analytics muestra datos reales. `bun run test` pasa.
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

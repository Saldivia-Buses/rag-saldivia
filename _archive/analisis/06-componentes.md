# 06 — Componentes React (68 activos)

## UI Primitivos (`components/ui/`) — 19 componentes

Basados en shadcn/ui + Radix + tokens custom del design system.

| Componente | Variantes / Notas | Dependencias |
|-----------|-------------------|-------------|
| `avatar.tsx` | Imagen + fallback con iniciales | Radix Avatar |
| `badge.tsx` | 6 variantes: default, secondary, outline, destructive, success, warning | — |
| `button.tsx` | 6 variantes: default, destructive, outline, secondary, ghost, link. Sizes: default, sm, lg, icon | Radix Slot |
| `command.tsx` | Command palette (busqueda con keybindings) | cmdk |
| `confirm-dialog.tsx` | Modal de confirmacion con variante destructive | Radix Dialog |
| `data-table.tsx` | Tabla avanzada con sorting, filtro, paginacion | @tanstack/react-table |
| `dialog.tsx` | Modal generico | Radix Dialog |
| `empty-placeholder.tsx` | Estado vacio con icono, titulo, descripcion | Lucide icons |
| `input.tsx` | Text input con tokens del design system | — |
| `popover.tsx` | Popover posicional | Radix Popover |
| `separator.tsx` | Linea divisora visual | Radix Separator |
| `sheet.tsx` | Panel slide-out (drawer) | Radix Dialog |
| `skeleton.tsx` | Loading shimmer (skeleton, skeleton-text, skeleton-table) | — |
| `sonner.tsx` | Toast notifications | sonner |
| `stat-card.tsx` | Tarjeta de estadistica con delta positivo/negativo | — |
| `table.tsx` | Tabla primitiva HTML con estilos | — |
| `textarea.tsx` | Multi-line input | — |
| `theme-toggle.tsx` | Switch light/dark mode | next-themes |
| `tooltip.tsx` | Tooltip en hover | Radix Tooltip |

---

## Chat (`components/chat/`) — 8 componentes

| Componente | Complejidad | Descripcion |
|-----------|------------|-------------|
| `ChatInterface.tsx` | **Alta (22)** | Componente mas complejo de la UI. Integra useChat (AI SDK), streaming, citations, artifacts, input, sources panel. Es el nucleo de la experiencia RAG. |
| `ChatLayout.tsx` | Baja | Layout wrapper que organiza sidebar + chat area |
| `ChatInputBar.tsx` | Media | Input de mensaje con auto-resize, adjuntos, keyboard shortcuts |
| `SessionList.tsx` | Media | Sidebar con lista de sesiones, search, create, delete |
| `SourcesPanel.tsx` | Media | Panel de citations/fuentes del RAG — core feature |
| `ArtifactPanel.tsx` | Media | Panel para code artifacts, HTML, SVG, mermaid |
| `CollectionSelector.tsx` | Baja | Multi-select de colecciones para la query |
| `MarkdownMessage.tsx` | Media | Renderiza markdown con syntax highlighting, latex, links |

---

## Admin (`components/admin/`) — 11 componentes

| Componente | Descripcion |
|-----------|-------------|
| `AdminDashboard.tsx` | Overview con stat cards (usuarios, sesiones, queries) |
| `AdminLayout.tsx` | Layout con sidebar de navegacion admin |
| `AdminUsers.tsx` | CRUD usuarios con DataTable, crear/editar/eliminar |
| `AdminRoles.tsx` | CRUD roles con permisos, color picker, icon selector |
| `AdminAreas.tsx` | CRUD areas con miembros y colecciones asignadas |
| `AdminCollections.tsx` | Gestion de colecciones Milvus (crear, eliminar, historial) |
| `AdminPermissions.tsx` | Matriz de permisos area x coleccion (read/write/admin) |
| `AdminRagConfig.tsx` | Sliders para parametros del LLM (temperature, top_k, etc.) |
| `PermissionMatrix.tsx` | Tabla interactiva de permisos con toggles |
| `RoleBadge.tsx` | Badge coloreado con icono para mostrar rol |
| `UserRoleSelector.tsx` | Dropdown para asignar roles a usuarios |

---

## Messaging (`components/messaging/`) — 19 componentes (Plan 25)

Sistema de mensajeria interna completo tipo Slack.

| Componente | Descripcion |
|-----------|-------------|
| `ChannelView.tsx` | Vista principal de un canal con header + messages + composer |
| `ChannelList.tsx` | Sidebar con canales, DMs, unread counts |
| `ChannelHeader.tsx` | Titulo del canal + info + acciones |
| `ChannelCreateDialog.tsx` | Modal para crear canal (public/private) |
| `DirectMessageDialog.tsx` | Modal para iniciar DM |
| `MessageList.tsx` | Feed de mensajes con scroll infinito |
| `MessageItem.tsx` | Mensaje individual con avatar, timestamp, acciones |
| `MessageComposer.tsx` | Input de mensaje con mentions, adjuntos, emoji |
| `MessageActions.tsx` | Menu contextual: pin, react, reply, edit, delete |
| `CommandPalette.tsx` | Paleta de comandos global para messaging |
| `MentionSuggestions.tsx` | Autocomplete de @mentions |
| `ReactionPicker.tsx` | Selector de emojis para reacciones |
| `FileAttachment.tsx` | Display de archivo adjunto con preview |
| `PinnedMessages.tsx` | Lista de mensajes pinneados del canal |
| `ThreadPanel.tsx` | Panel lateral para threads/respuestas |
| `TypingIndicator.tsx` | "Fulano esta escribiendo..." |
| `PresenceIndicator.tsx` | Circulo verde/gris de online/offline |
| `UnreadBadge.tsx` | Badge numerico de mensajes no leidos |
| `VoiceInput.tsx` | Grabacion de mensajes de voz |

---

## Collections (`components/collections/`) — 2 componentes

| Componente | Descripcion |
|-----------|-------------|
| `CollectionsList.tsx` | Lista de colecciones con permisos del usuario |
| `CollectionDetail.tsx` | Detalle de coleccion + historial de ingesta |

---

## Settings (`components/settings/`) — 2 componentes

| Componente | Descripcion |
|-----------|-------------|
| `SettingsClient.tsx` | Pagina de settings: perfil, password, preferencias |
| `MemoryClient.tsx` | Gestion de memoria personalizada del RAG |

---

## Layout (`components/layout/`) — 3 componentes

| Componente | Descripcion |
|-----------|-------------|
| `AppShell.tsx` | Shell principal: NavRail + content area |
| `AppShellChrome.tsx` | Chrome browser-like wrapper |
| `NavRail.tsx` | Sidebar izquierda de 64px con iconos de navegacion |

---

## Otros

| Componente | Descripcion |
|-----------|-------------|
| `error-boundary.tsx` | Error boundary de React con fallback UI |
| `providers.tsx` | ThemeProvider de next-themes |
| `dev/ReactScan.tsx` | React performance scanner (solo dev) |
| `dev/ReactScanProvider.tsx` | Provider del scanner (solo dev) |

---

## Tests de componentes (24 archivos)

Archivos de test en `components/__tests__/`:

| Test | Componentes testeados |
|------|----------------------|
| `button.test.tsx` | Button (6 variantes, sizes, disabled, asChild) |
| `badge.test.tsx` | Badge (6 variantes) |
| `input.test.tsx` | Input (tipos, placeholder, disabled) |
| `textarea.test.tsx` | Textarea (rows, resize, disabled) |
| `avatar.test.tsx` | Avatar (imagen, fallback) |
| `table.test.tsx` | Table (rows, headers, responsive) |
| `confirm-dialog.test.tsx` | ConfirmDialog (open, confirm, cancel, destructive) |
| `data-table.test.tsx` | DataTable (sorting, filtering, pagination) |
| `empty-placeholder.test.tsx` | EmptyPlaceholder (icon, title, description) |
| `separator.test.tsx` | Separator (horizontal, vertical) |
| `skeleton.test.tsx` | Skeleton, SkeletonText, SkeletonTable |
| `stat-card.test.tsx` | StatCard (delta positivo/negativo, trend) |
| `theme-toggle.test.tsx` | ThemeToggle (light/dark switch) |
| `error-boundary.test.tsx` | ErrorBoundary (catch, fallback, recovery) |
| `admin-*.test.tsx` | AdminDashboard, AdminUsers, AdminRoles, etc. |
| `chat-*.test.tsx` | ChatInterface, SessionList, CollectionSelector |
| `collections.test.tsx` | CollectionsList |
| `settings.test.tsx` | SettingsClient |

---

## Componentes archivados (`_archive/components/`)

70+ componentes aspiracionales movidos a `_archive/` en Plan 13:

**Chat avanzado:** VoiceInput, ArtifactsPanel, DocPreviewPanel, ExportSession, ShareDialog, RelatedQuestions, PromptTemplates, ThinkingSteps, AnnotationPopover, ChatDropZone, FocusModeSelector

**Layout avanzado:** CommandPalette (viejo), WhatsNewPanel, SecondaryPanel, panels/AdminPanel, panels/ChatPanel, panels/ProjectsPanel

**Features futuras:** ExtractionWizard, UploadClient, ProjectsClient, OnboardingTour, DocumentGraph, CollectionHistory (viejo)

**Admin viejo:** 12 componentes admin del stack anterior

---

## Hotspots de complejidad (medidos)

### Top 10 archivos mas complejos (actualizado post Plan 28)

| # | Archivo | Lineas | Estado | Riesgo |
|---|---------|--------|--------|--------|
| 1 | `ArtifactPanel.tsx` | 541 | Sin cambios | ALTO |
| 2 | `ChatInterface.tsx` | ~360 | Descompuesto (Plan 28): -44% | MEDIO (era CRITICO) |
| 3 | `MarkdownMessage.tsx` | 308 | Sin cambios | MEDIO |
| 4 | `AdminDashboard.tsx` | 254 | Sin cambios | MEDIO |
| 5 | `AdminRoles.tsx` | ~238 | Descompuesto (Plan 28): -62% | BAJO (era ALTO) |
| 6 | `AdminUsers.tsx` | ~260 | Descompuesto (Plan 28): -56% | BAJO (era ALTO) |
| 7 | `NavRail.tsx` | 229 | Sin cambios | MEDIO |
| 8 | `AdminPermissions.tsx` | 225 | Sin cambios | MEDIO |
| 9 | `PermissionMatrix.tsx` | 205 | Sin cambios | MEDIO |
| 10 | `MessageItem.tsx` | 194 | Sin cambios | BAJO |

### Decomposicion (Plan 28) — reduccion total: 1,861 → 858 lineas (-54%)

**ChatInterface.tsx: 643 → ~360 lineas.** Extraidos:
- `ChatEmptyState` — estado vacio cuando no hay sesion
- `ChatMessages` — rendering de mensajes con streaming

**AdminRoles.tsx: 626 → ~238 lineas.** Extraidos:
- `RoleCard` — card individual de rol con badge y permisos
- `RoleForm` — formulario de crear/editar rol

**AdminUsers.tsx: 592 → ~260 lineas.** Extraidos:
- `CreateUserForm` — formulario de creacion de usuario
- `PasswordResetCell` — celda de reset password en la tabla

**7 sub-componentes nuevos** con responsabilidad unica. Todos con tests (Plan 29).

### Acoplamiento por imports (>10 imports)

| Archivo | Imports | Tipo |
|---------|---------|------|
| `ChatInterface.tsx` | 21 | AI SDK + hooks + actions + types + 6 componentes |
| `NavRail.tsx` | 11 | lucide icons + next/link + next-themes + hooks |
| `CollectionsList.tsx` | 11 | actions + db types + UI components |
| `api/rag/generate/route.ts` | 11 | auth + rag client + ai-stream + db + logger |

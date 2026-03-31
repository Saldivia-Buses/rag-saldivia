# 14 — Mapa de Archivos Completo

## Conteo general

| Directorio | Archivos | Tipo |
|-----------|----------|------|
| `apps/web/src/app/` | 54 | Routes, pages, layouts, API, actions |
| `apps/web/src/components/` | 68 | Componentes React |
| `apps/web/src/hooks/` | 7 | Custom hooks |
| `apps/web/src/lib/` | 30 | Logica core |
| `apps/web/src/types/` | 1 | Type definitions (globals.d.ts) |
| `apps/web/src/workers/` | 1 | Background workers (messaging-notifications.ts) |
| `apps/web/src/__tests__/` | 8 | Integration tests (incl. hooks/) |
| `packages/db/src/` | 55 | Schema + queries + tests |
| `packages/shared/src/` | 6 | Zod schemas |
| `packages/config/src/` | 3 | Config loader |
| `packages/logger/src/` | 8 | Logger |
| `docs/` | 32 | Documentacion |
| `scripts/` | 5 | Scripts de setup |
| `config/` | 12 | Config de deploy |
| `.claude/agents/` | 10 | Agent definitions |
| `_archive/` | 123 | Codigo archivado |
| **Total estimado** | **~450** | — |

---

## apps/web/src/app/ (54 archivos)

### Layouts
```
app/layout.tsx                          → Root layout (font, theme, providers)
app/(app)/layout.tsx                    → Protected layout (AppShell)
app/(app)/admin/layout.tsx              → Admin layout (sidebar nav)
app/(app)/chat/layout.tsx               → Chat layout
app/(app)/messaging/layout.tsx          → Messaging layout
```

### Pages (14 rutas)
```
app/page.tsx                            → Redirect a /chat
app/(auth)/login/page.tsx               → Login (publica)
app/(app)/chat/page.tsx                 → Lista de sesiones
app/(app)/chat/[id]/page.tsx            → Conversacion
app/(app)/collections/page.tsx          → Lista de colecciones
app/(app)/collections/[name]/page.tsx   → Detalle de coleccion
app/(app)/settings/page.tsx             → Perfil, password, preferencias
app/(app)/admin/page.tsx                → Admin dashboard
app/(app)/admin/users/page.tsx          → CRUD usuarios
app/(app)/admin/roles/page.tsx          → CRUD roles
app/(app)/admin/areas/page.tsx          → CRUD areas
app/(app)/admin/permissions/page.tsx    → Matriz de permisos
app/(app)/admin/collections/page.tsx    → Gestion de colecciones
app/(app)/admin/config/page.tsx         → Parametros RAG
app/(app)/messaging/page.tsx            → Lista de canales
app/(app)/messaging/[channelId]/page.tsx → Canal individual
```

### API Routes (18 endpoints)
```
app/api/auth/login/route.ts             → POST login
app/api/auth/logout/route.ts            → DELETE logout
app/api/auth/refresh/route.ts           → POST refresh

app/api/rag/generate/route.ts           → POST streaming SSE
app/api/rag/collections/route.ts        → GET/POST colecciones
app/api/rag/collections/[name]/route.ts → DELETE coleccion
app/api/rag/collections/[name]/history/route.ts → GET historial
app/api/rag/document/[name]/route.ts    → GET documento
app/api/rag/suggest/route.ts            → POST sugerencias

app/api/messaging/channels/route.ts     → GET/POST canales
app/api/messaging/channels/[id]/route.ts → GET/PUT/DELETE canal
app/api/messaging/channels/[id]/members/route.ts → GET/POST miembros
app/api/messaging/messages/route.ts     → POST mensaje
app/api/messaging/messages/[id]/pin/route.ts → POST pin/unpin
app/api/messaging/messages/[id]/reactions/route.ts → POST reaccion
app/api/messaging/search/route.ts       → GET busqueda
app/api/messaging/upload/route.ts       → POST upload

app/api/health/route.ts                 → GET health check
```

### Server Actions (10 archivos)
```
app/actions/auth.ts                     → actionLogout
app/actions/chat.ts                     → 11 actions de chat
app/actions/collections.ts              → actionCreateCollection, actionDeleteCollection
app/actions/config.ts                   → actionUpdateRagParams, actionResetRagParams
app/actions/settings.ts                 → 6 actions de settings
app/actions/admin.ts                    → 5 actions de admin
app/actions/roles.ts                    → 8 actions de RBAC
app/actions/areas.ts                    → 4 actions de areas
app/actions/templates.ts                → 3 actions de templates
app/actions/messaging.ts                → Actions de messaging
```

---

## apps/web/src/components/ (68 archivos activos)

```
components/
  ui/                           (19 archivos)
    avatar.tsx
    badge.tsx
    button.tsx
    command.tsx
    confirm-dialog.tsx
    data-table.tsx
    dialog.tsx
    empty-placeholder.tsx
    input.tsx
    popover.tsx
    separator.tsx
    sheet.tsx
    skeleton.tsx
    sonner.tsx
    stat-card.tsx
    table.tsx
    textarea.tsx
    theme-toggle.tsx
    tooltip.tsx

  chat/                         (8 archivos)
    ChatInterface.tsx
    ChatLayout.tsx
    ChatInputBar.tsx
    SessionList.tsx
    SourcesPanel.tsx
    ArtifactPanel.tsx
    CollectionSelector.tsx
    MarkdownMessage.tsx

  admin/                        (11 archivos)
    AdminDashboard.tsx
    AdminLayout.tsx
    AdminUsers.tsx
    AdminRoles.tsx
    AdminAreas.tsx
    AdminCollections.tsx
    AdminPermissions.tsx
    AdminRagConfig.tsx
    PermissionMatrix.tsx
    RoleBadge.tsx
    UserRoleSelector.tsx

  messaging/                    (19 archivos)
    ChannelView.tsx
    ChannelList.tsx
    ChannelHeader.tsx
    ChannelCreateDialog.tsx
    DirectMessageDialog.tsx
    MessageList.tsx
    MessageItem.tsx
    MessageComposer.tsx
    MessageActions.tsx
    CommandPalette.tsx
    MentionSuggestions.tsx
    ReactionPicker.tsx
    FileAttachment.tsx
    PinnedMessages.tsx
    ThreadPanel.tsx
    TypingIndicator.tsx
    PresenceIndicator.tsx
    UnreadBadge.tsx
    VoiceInput.tsx

  collections/                  (2 archivos)
    CollectionsList.tsx
    CollectionDetail.tsx

  settings/                     (2 archivos)
    SettingsClient.tsx
    MemoryClient.tsx

  layout/                       (3 archivos)
    AppShell.tsx
    AppShellChrome.tsx
    NavRail.tsx

  dev/                          (2 archivos)
    ReactScan.tsx
    ReactScanProvider.tsx

  (root)                        (2 archivos)
    error-boundary.tsx
    providers.tsx

  __tests__/                    (18 archivos de test)
    button.test.tsx
    badge.test.tsx
    input.test.tsx
    textarea.test.tsx
    avatar.test.tsx
    table.test.tsx
    confirm-dialog.test.tsx
    data-table.test.tsx
    empty-placeholder.test.tsx
    separator.test.tsx
    skeleton.test.tsx
    stat-card.test.tsx
    theme-toggle.test.tsx
    error-boundary.test.tsx
    admin-*.test.tsx
    chat-*.test.tsx
    collections.test.tsx
    settings.test.tsx
```

---

## apps/web/src/hooks/ (7 archivos)

```
hooks/
  useAutoResize.ts              → Auto-expand textarea
  useCopyToClipboard.ts         → Copy con feedback
  useGlobalHotkeys.ts           → Cmd+K, etc.
  useLocalStorage.ts            → Persistent state
  useMessaging.ts               → Messaging API
  usePresence.ts                → WebSocket presencia
  useTyping.ts                  → "Escribiendo..." indicator
```

---

## apps/web/src/lib/ (30 archivos)

```
lib/
  auth/
    jwt.ts                      → createJwt, verifyJwt, extractClaims
    current-user.ts             → requireUser, requireAdmin (Server Components)
    rbac.ts                     → hasRole, canAccessRoute
    permissions.ts              → Permission checker
    __tests__/jwt.test.ts

  rag/
    client.ts                   → ragGenerateStream, ragFetch, mock mode
    ai-stream.ts                → Adapter NVIDIA SSE → AI SDK
    stream.ts                   → Low-level SSE parser
    artifact-parser.ts          → Parse code artifacts
    collections-cache.ts        → In-memory cache
    __tests__/                  → Stream tests

  ws/
    protocol.ts                 → WS message types
    client.ts                   → Client-side WS handler
    presence.ts                 → Presence protocol
    publish.ts                  → Pub-sub helpers
    sidecar.ts                  → Sidecar process

  safe-action.ts                → authAction, adminAction middleware
  utils.ts                      → cn, formatDate, etc.
  api-utils.ts                  → API response helpers
  webhook.ts                    → Webhook signature verification
  changelog.ts                  → Changelog generator
  export.ts                     → Session export
  queue.ts                      → BullMQ setup
  defaults.ts                   → DEFAULT_COLLECTION
  test-setup.ts                 → Test preload (mocks)
  component-test-setup.ts       → Happy-dom setup

  __tests__/
    utils.test.ts
    changelog.test.ts
    webhook.test.ts
    export.test.ts
```

---

## packages/db/src/ (55 archivos)

```
db/src/
  schema/
    core.ts                     → 13 tablas (users, areas, RBAC, etc.)
    chat.ts                     → 11 tablas (sessions, messages, etc.)
    messaging.ts                → 6 tablas (channels, messages, etc.)
    events.ts                   → 1 tabla (events)
    relations.ts                → Drizzle relations
    index.ts                    → Re-exports

  queries/                      (21 archivos)
    users.ts, areas.ts, sessions.ts, messages.ts, events.ts,
    saved.ts, annotations.ts, tags.ts, shares.ts, templates.ts,
    collection-history.ts, reports.ts, rate-limits.ts, webhooks.ts,
    search.ts, projects.ts, memory.ts, external-sources.ts,
    rbac.ts, channels.ts, messaging.ts

  __tests__/                    (19 archivos)
    users.test.ts, areas.test.ts, sessions.test.ts, tags.test.ts,
    saved.test.ts, annotations.test.ts, templates.test.ts,
    projects.test.ts, external-sources.test.ts, collection-history.test.ts,
    reports.test.ts, rate-limits.test.ts, webhooks.test.ts,
    search.test.ts, shares.test.ts, memory.test.ts,
    messaging.test.ts, channels.test.ts, events.test.ts, redis.test.ts

  connection.ts, redis.ts, init.ts, migrate.ts, seed.ts,
  test-setup.ts, fts5.ts, index.ts
```

---

## docs/ (32 archivos)

```
docs/
  bible.md                      → Reglas permanentes
  architecture.md               → Arquitectura del sistema
  design-system.md              → Tokens, tipografia, componentes
  api.md                        → Referencia de API routes
  testing.md                    → Estrategia de testing
  blackbox.md                   → Event logging y replay
  onboarding.md                 → Guia de 5 minutos
  workflows.md                  → Git, tests, features, deploy
  metodologia-de-trabajo.md     → Metodologia (espanol)
  roadmap-1.0.x.md              → Roadmap
  toolbox.md                    → Herramientas externas
  cli.md                        → Referencia CLI (archivada)

  decisions/                    (13 ADRs)
    000-adr-template.md
    001-libsql.md → ... → 012-stack-definitivo.md

  plans/                        (23 planes)
    1.0.x-plan-maestro.md       → Plan maestro
    1.0.x-plan13-*.md → ... → 1.0.x-plan25-*.md
    1.0.x-plans-23-24-22-unified.md

  templates/                    (6 templates)
    plan-template.md, pr-template.md, adr-template.md,
    commit-conventions.md, version-template.md, artifact-template.md

  artifacts/                    → Resultados de reviews/audits
```

---

## .claude/agents/ (10 archivos)

```
agents/
  frontend-reviewer.md          → Review de componentes/UI
  gateway-reviewer.md           → Review de API/auth
  security-auditor.md           → Auditoria de seguridad
  test-writer.md                → Escribir tests
  debugger.md                   → Debugging sistematico
  doc-writer.md                 → Actualizar docs
  deploy.md                     → Deploy a workstation
  status.md                     → Estado de servicios
  plan-writer.md                → Escribir planes
  ingest.md                     → Ingestar documentos
```

---

## Otros archivos root

```
CHANGELOG.md                    → 121KB, formato Keep a Changelog
CLAUDE.md                       → 12KB, contexto para Claude Code
CONTRIBUTING.md                 → 5KB, guia para agentes AI
LICENSE                         → MIT
Makefile                        → Deploy y operaciones
README.md                       → Intro al proyecto
SECURITY.md                     → Politica de seguridad
package.json                    → Monorepo config
turbo.json                      → Turborepo config
bun.lock                        → Lockfile
bunfig.toml                     → Bun config
docker-compose.yml              → Docker services
knip.json                       → Dead code detection
.commitlintrc.json              → Commit linting
.editorconfig                   → Editor config
.env.example                    → Variables de entorno
.gitignore                      → Git ignore
.gitmodules                     → Submodule (vendor/rag-blueprint)
.lintstagedrc.js                → Lint-staged config
.mcp.json                       → MCP servers config
```

---

## Metricas medidas (repomix Tree-sitter compression)

### Totales

| Scope | Archivos | Chars | Tokens | Lineas |
|-------|----------|-------|--------|--------|
| `apps/web/src/**/*.{ts,tsx}` (sin tests) | 155 | 128,740 | 32,632 | 4,527 |
| `packages/**/*.ts` (sin tests) | 51 | 62,747 | 15,599 | 1,961 |
| **Total codigo activo** | **206** | **191,487** | **48,231** | **6,488** |

### Distribucion por directorio (apps/web/src)

| Directorio | Archivos | Proporcion |
|-----------|----------|-----------|
| `app/` (routes, pages, actions, API) | 54 | 35% |
| `components/` (sin tests) | 68 | 44% |
| `lib/` (sin tests) | 18 | 12% |
| `hooks/` | 7 | 4% |
| `workers/` + `types/` + root | 2 | 1% |

### Distribucion por directorio (packages)

| Directorio | Archivos | Proporcion |
|-----------|----------|-----------|
| `db/src/queries/` | 21 | 41% |
| `db/src/schema/` | 6 | 12% |
| `db/src/` (infra) | 9 | 18% |
| `logger/src/` | 7 | 14% |
| `shared/src/` | 5 | 10% |
| `config/src/` | 2 | 4% |

### Top 5 archivos mas grandes (apps/web)

| Archivo | Tokens |
|---------|--------|
| `ChatInterface.tsx` | 1,065 |
| `lib/ws/sidecar.ts` | 758 |
| `AdminUsers.tsx` | 754 |
| `lib/rag/artifact-parser.ts` | 726 |
| `AdminRoles.tsx` | 719 |

### Top 5 archivos mas grandes (packages)

| Archivo | Tokens |
|---------|--------|
| `logger/src/blackbox.ts` | 958 |
| `db/src/queries/rbac.ts` | 884 |
| `db/src/schema/core.ts` | 857 |
| `logger/src/backend.ts` | 766 |
| `db/src/queries/messaging.ts` | 714 |

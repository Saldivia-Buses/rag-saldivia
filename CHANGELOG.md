# Changelog

Todos los cambios notables de este proyecto se documentan en este archivo.

Formato basado en [Keep a Changelog](https://keepachangelog.com/es/1.1.0/).
Versionado basado en [Semantic Versioning](https://semver.org/lang/es/).

---

## [Unreleased]

### Added

- `packages/db/src/schema.ts`: tabla `bot_user_mappings` (platform, externalUserId, systemUserId) — 2026-03-25 *(Plan 4 F3.49)*
- `apps/web/src/app/api/slack/route.ts`: handler de slash commands Slack — verifica firma HMAC, resuelve userId via mapping, consulta RAG y responde via response_url — 2026-03-25 *(Plan 4 F3.49)*
- `apps/web/src/app/api/teams/route.ts`: handler de mensajes Teams — respeta RBAC via mapping de usuarios — 2026-03-25 *(Plan 4 F3.49)*
- `apps/web/src/app/(app)/admin/integrations/page.tsx` y `IntegrationsAdmin.tsx`: UI de configuración con URLs y guía de setup — 2026-03-25 *(Plan 4 F3.49)*
- `packages/db/src/schema.ts`: tabla `external_sources` (provider, credentials, collectionDest, schedule, lastSync) — 2026-03-25 *(Plan 4 F3.48)*
- `packages/db/src/queries/external-sources.ts`: `createExternalSource`, `listExternalSources`, `listActiveSourcesToSync`, `updateSourceLastSync`, `deleteExternalSource` — 2026-03-25 *(Plan 4 F3.48)*
- `apps/web/src/workers/external-sync.ts`: worker MVP que detecta fuentes listas para sync y registra logs; implementación OAuth completa pendiente de credenciales reales — 2026-03-25 *(Plan 4 F3.48)*
- `apps/web/src/app/(app)/admin/external-sources/page.tsx` y `ExternalSourcesAdmin.tsx`: UI para configurar fuentes externas — 2026-03-25 *(Plan 4 F3.48)*
- `packages/db/src/schema.ts`: campos `sso_provider` y `sso_subject` en tabla `users` — 2026-03-25 *(Plan 4 F3.47)*
- `apps/web/src/lib/auth/next-auth.ts`: configuración NextAuth v5 con providers Google y Microsoft Entra ID; modo mixto (SSO + password); al primer login SSO crea usuario o vincula cuenta existente; emite JWT propio para compatibilidad RBAC — 2026-03-25 *(Plan 4 F3.47)*
- `apps/web/src/app/api/auth/[...nextauth]/route.ts`: handler de NextAuth — 2026-03-25 *(Plan 4 F3.47)*
- `apps/web/src/components/auth/SSOButton.tsx`: botones Google y Microsoft en página de login (solo visibles si los env vars están configurados) — 2026-03-25 *(Plan 4 F3.47)*
- `.env.example`: variables SSO y NextAuth documentadas — 2026-03-25 *(Plan 4 F3.47)*
- `apps/web/src/app/api/collections/[name]/embeddings/route.ts`: retorna grafo de similitud — intenta obtener docs del RAG server, simula similitud para MVP — 2026-03-25 *(Plan 4 F3.46)*
- `apps/web/src/components/collections/DocumentGraph.tsx`: visualización SVG force-directed sin dependencia de d3-force (simulación propia ligera); zoom, colores por cluster, click en nodo — 2026-03-25 *(Plan 4 F3.46)*
- `apps/web/src/app/(app)/collections/[name]/graph/page.tsx`: página del grafo por colección — 2026-03-25 *(Plan 4 F3.46)*
- `apps/web/src/components/collections/CollectionsList.tsx`: botón "Grafo" por colección — 2026-03-25 *(Plan 4 F3.46)*
- `apps/web/src/workers/ingestion.ts`: `checkProactiveSurface` — cruza keywords del doc nuevo con queries recientes del usuario; si hay match genera evento `proactive.docs_available` — 2026-03-25 *(Plan 4 F3.45)*
- `apps/web/src/app/api/notifications/route.ts`: `proactive.docs_available` agregado a los tipos de notificación — 2026-03-25 *(Plan 4 F3.45)*
- `packages/db/src/schema.ts`: tabla `user_memory` (key, value, source explicit/inferred, UNIQUE per user+key) — 2026-03-25 *(Plan 4 F3.44)*
- `packages/db/src/queries/memory.ts`: `setMemory` (upsert), `getMemory`, `deleteMemory`, `getMemoryAsContext` — 2026-03-25 *(Plan 4 F3.44)*
- `apps/web/src/app/api/rag/generate/route.ts`: inyección de memoria del usuario como system message — 2026-03-25 *(Plan 4 F3.44)*
- `apps/web/src/app/(app)/settings/memory/page.tsx`: UI para ver/agregar/eliminar preferencias de memoria — 2026-03-25 *(Plan 4 F3.44)*
- `packages/db/src/schema.ts`: campo `forked_from` en `chat_sessions` (TEXT nullable, sin FK circular en Drizzle) — 2026-03-25 *(Plan 4 F3.43)*
- `apps/web/src/app/actions/chat.ts`: Server Action `actionForkSession` — copia sesión y mensajes hasta el punto indicado, setea `forked_from` — 2026-03-25 *(Plan 4 F3.43)*
- `apps/web/src/components/chat/ChatInterface.tsx`: botón `GitBranch` en mensajes del asistente para bifurcar desde ese punto — 2026-03-25 *(Plan 4 F3.43)*
- `apps/web/src/components/chat/SessionList.tsx`: badge `GitBranch` en sesiones con `forked_from` no null — 2026-03-25 *(Plan 4 F3.43)*
- `apps/web/src/hooks/useRagStream.ts`: detección de artifacts al finalizar stream — marcador `:::artifact` explícito o heurística (código ≥ 40 líneas, tabla ≥ 5 cols); callback `onArtifact` — 2026-03-25 *(Plan 4 F3.42)*
- `apps/web/src/components/chat/ArtifactsPanel.tsx`: Sheet lateral para código/tabla/documento — botones guardar y exportar; resaltado de código en `<pre>` — 2026-03-25 *(Plan 4 F3.42)*
- `packages/db/src/schema.ts`: tablas `projects`, `project_sessions`, `project_collections` — 2026-03-25 *(Plan 4 F3.41)*
- `packages/db/src/queries/projects.ts`: `createProject`, `listProjects`, `getProject`, `updateProject`, `deleteProject`, `addSessionToProject`, `addCollectionToProject`, `getProjectBySession` — 2026-03-25 *(Plan 4 F3.41)*
- `apps/web/src/app/api/projects/route.ts`: GET/POST/DELETE para proyectos — 2026-03-25 *(Plan 4 F3.41)*
- `apps/web/src/app/(app)/projects/page.tsx` y `[id]/page.tsx`: páginas de proyectos — 2026-03-25 *(Plan 4 F3.41)*
- `apps/web/src/components/projects/ProjectsClient.tsx`: grid de proyectos con crear/eliminar — 2026-03-25 *(Plan 4 F3.41)*
- `apps/web/src/components/layout/panels/ProjectsPanel.tsx`: panel secundario para rutas `/projects` — 2026-03-25 *(Plan 4 F3.41)*
- `apps/web/src/components/layout/NavRail.tsx`: ícono `FolderKanban` para `/projects` — 2026-03-25 *(Plan 4 F3.41)*
- `apps/web/src/app/api/rag/generate/route.ts`: inyección del contexto del proyecto como system message — 2026-03-25 *(Plan 4 F3.41)*
- `apps/web/src/components/chat/DocPreviewPanel.tsx`: panel Sheet lateral para preview de PDF con react-pdf (carga dinámica SSR-safe), paginación, fallback a texto cuando el Blueprint no expone el documento — 2026-03-25 *(Plan 4 F3.40)*
- `apps/web/src/app/api/rag/document/[name]/route.ts`: proxy al RAG server para obtener PDF; retorna 404 con nota si el endpoint no está disponible — 2026-03-25 *(Plan 4 F3.40)*
- `apps/web/src/components/chat/SourcesPanel.tsx`: nombre de cada fuente ahora es botón clic que abre `DocPreviewPanel` con el fragmento resaltado — 2026-03-25 *(Plan 4 F3.40)*
- `packages/db/src/queries/search.ts`: `universalSearch(query, userId, limit)` — busca con FTS5 (sesiones + mensajes) con fallback a LIKE; también busca en templates y saved_responses — 2026-03-25 *(Plan 4 F3.39)*
- `packages/db/src/init.ts`: tablas FTS5 virtuales `sessions_fts` y `messages_fts` con triggers de sincronización automática — 2026-03-25 *(Plan 4 F3.39)*
- `apps/web/src/app/api/search/route.ts`: endpoint `GET /api/search?q=...` con debounce 300ms — 2026-03-25 *(Plan 4 F3.39)*
- `apps/web/src/components/layout/CommandPalette.tsx`: integración de búsqueda universal — grupo "Resultados para X" con tipo (session/message/saved/template) y snippet; aparece cuando query ≥ 2 chars — 2026-03-25 *(Plan 4 F3.39)*

### Added

- `packages/db/src/schema.ts`: tabla `webhooks` (url, events JSON, secret HMAC, active) — 2026-03-25 *(Plan 4 F2.38)*
- `packages/db/src/queries/webhooks.ts`: `createWebhook` (genera secret aleatorio), `listWebhooksByEvent`, `listAllWebhooks`, `deleteWebhook` — 2026-03-25 *(Plan 4 F2.38)*
- `apps/web/src/lib/webhook.ts`: `dispatchWebhook` con firma HMAC-SHA256 en header `X-Signature`; timeout 5s; no interrumpe el flujo principal si falla; `dispatchEvent` busca webhooks activos para el tipo de evento — 2026-03-25 *(Plan 4 F2.38)*
- `apps/web/src/workers/ingestion.ts`: dispatch de `ingestion.completed` al terminar un job — 2026-03-25 *(Plan 4 F2.38)*
- `apps/web/src/app/api/rag/generate/route.ts`: dispatch de `query.completed` al finalizar un stream — 2026-03-25 *(Plan 4 F2.38)*
- `apps/web/src/app/api/admin/webhooks/route.ts`: GET/POST/DELETE para gestión de webhooks — 2026-03-25 *(Plan 4 F2.38)*
- `apps/web/src/components/admin/WebhooksAdmin.tsx`: UI para crear/listar/eliminar webhooks con selector de eventos y copia del secret — 2026-03-25 *(Plan 4 F2.38)*
- `apps/web/src/app/(app)/admin/webhooks/page.tsx`: página `/admin/webhooks` — 2026-03-25 *(Plan 4 F2.38)*
- `packages/db/src/schema.ts`: campo `onboarding_completed` en tabla `users` (default false) — 2026-03-25 *(Plan 4 F2.37)*
- `apps/web/src/components/onboarding/OnboardingTour.tsx`: tour driver.js de 5 pasos (nav, chat, modos de foco, colecciones, settings); se activa al primer login; saltable; llama a `actionCompleteOnboarding` al terminar — 2026-03-25 *(Plan 4 F2.37)*
- `apps/web/src/app/actions/settings.ts`: Server Actions `actionCompleteOnboarding` y `actionResetOnboarding` — 2026-03-25 *(Plan 4 F2.37)*
- `apps/web/src/app/(app)/layout.tsx`: renderiza `OnboardingTour` condicionalmente si `onboardingCompleted === false` — 2026-03-25 *(Plan 4 F2.37)*
- `packages/db/src/__tests__/users.test.ts` y `saved.test.ts`: columna `onboarding_completed` agregada al SQL de test — 2026-03-25 *(bugfix)*
- `packages/db/src/schema.ts`: tabla `rate_limits` (targetType user/area, targetId, maxQueriesPerHour) — 2026-03-25 *(Plan 4 F2.36)*
- `packages/db/src/queries/rate-limits.ts`: `getRateLimit` (user-level primero, luego área), `countQueriesLastHour`, `createRateLimit`, `listRateLimits`, `deleteRateLimit` — 2026-03-25 *(Plan 4 F2.36)*
- `apps/web/src/app/api/rag/generate/route.ts`: check de rate limit antes de procesar — retorna 429 si el usuario superó su límite/hora — 2026-03-25 *(Plan 4 F2.36)*
- `apps/web/src/components/chat/ChatDropZone.tsx`: drop zone sobre el área del chat — overlay al arrastrar, sube el archivo via `/api/upload` y confirma con toast; colecciones efímeras descartadas por viabilidad (Blueprint v2.5.0 no las soporta en Milvus) — 2026-03-25 *(Plan 4 F2.35)*
- `apps/web/src/components/chat/SplitView.tsx`: wrapper de vista dividida — botón Columns2 abre segundo panel con sesión independiente, botón X cierra; cada panel tiene su propio ChatInterface — 2026-03-25 *(Plan 4 F2.34)*
- `packages/db/src/schema.ts`: tabla `scheduled_reports` (query, collection, schedule, destination, email, nextRun) — 2026-03-25 *(Plan 4 F2.33)*
- `packages/db/src/queries/reports.ts`: `createReport`, `listActiveReports`, `listReportsByUser`, `updateLastRun`, `deleteReport` — 2026-03-25 *(Plan 4 F2.33)*
- `apps/web/src/workers/ingestion.ts`: procesador de informes programados — corre cada 5 min, ejecuta query via RAG, guarda en Guardados o env\u00eda por email (si SMTP configurado) — 2026-03-25 *(Plan 4 F2.33)*
- `apps/web/src/app/api/admin/reports/route.ts`: GET/POST/DELETE para informes programados — 2026-03-25 *(Plan 4 F2.33)*
- `apps/web/src/components/admin/ReportsAdmin.tsx`: formulario + lista de informes — 2026-03-25 *(Plan 4 F2.33)*
- `apps/web/src/app/(app)/admin/reports/page.tsx`: p\u00e1gina `/admin/reports` — 2026-03-25 *(Plan 4 F2.33)*
- `packages/db/src/schema.ts`: tabla `collection_history` (collection, userId, action, filename, docCount) — 2026-03-25 *(Plan 4 F2.32)*
- `packages/db/src/queries/collection-history.ts`: `recordIngestionEvent`, `listHistoryByCollection` — 2026-03-25 *(Plan 4 F2.32)*
- `apps/web/src/workers/ingestion.ts`: registrar en `collection_history` al completar job exitosamente — 2026-03-25 *(Plan 4 F2.32)*
- `apps/web/src/components/collections/CollectionHistory.tsx`: timeline de ingestas por colección — 2026-03-25 *(Plan 4 F2.32)*
- `apps/web/src/app/api/collections/[name]/history/route.ts`: endpoint GET para historial de una colección — 2026-03-25 *(Plan 4 F2.32)*
- `apps/web/src/app/api/admin/knowledge-gaps/route.ts`: detecta respuestas del asistente con baja confianza (< 80 palabras + keywords de incertidumbre), retorna hasta 100 gaps — 2026-03-25 *(Plan 4 F2.31)*
- `apps/web/src/components/admin/KnowledgeGapsClient.tsx`: tabla de brechas con link a sesión, exportar CSV, botón "Ingestar documentos" — 2026-03-25 *(Plan 4 F2.31)*
- `apps/web/src/app/(app)/admin/knowledge-gaps/page.tsx`: página `/admin/knowledge-gaps` — 2026-03-25 *(Plan 4 F2.31)*
- `apps/web/src/app/api/admin/analytics/route.ts`: queries de agregación — queries/día, top colecciones, distribución feedback, top usuarios — 2026-03-25 *(Plan 4 F2.30)*
- `apps/web/src/components/admin/AnalyticsDashboard.tsx`: dashboard con recharts — LineChart queries/día, BarChart colecciones, PieChart feedback, tabla top usuarios; stats cards con totales — 2026-03-25 *(Plan 4 F2.30)*
- `apps/web/src/app/(app)/admin/analytics/page.tsx`: página `/admin/analytics` — 2026-03-25 *(Plan 4 F2.30)*
- `apps/web/src/app/api/admin/ingestion/stream/route.ts`: SSE endpoint que emite estado de jobs cada 3s — 2026-03-25 *(Plan 4 F2.29)*
- `apps/web/src/app/api/admin/ingestion/[id]/route.ts`: PATCH con `action: "retry"` para reintentar jobs fallidos — 2026-03-25 *(Plan 4 F2.29)*
- `apps/web/src/components/admin/IngestionKanban.tsx`: kanban de 4 columnas (Pendiente/En progreso/Completado/Error) con barra de progreso, detalle de error expandible, botón retry, indicador SSE en tiempo real — 2026-03-25 *(Plan 4 F2.29)*
- `apps/web/src/app/(app)/admin/ingestion/page.tsx`: página de monitoring de ingesta — 2026-03-25 *(Plan 4 F2.29)*
- `packages/db/src/schema.ts`: tabla `prompt_templates` (title, prompt, focusMode, createdBy, active) — 2026-03-25 *(Plan 4 F2.28)*
- `packages/db/src/queries/templates.ts`: `listActiveTemplates`, `createTemplate`, `deleteTemplate` — 2026-03-25 *(Plan 4 F2.28)*
- `apps/web/src/app/api/admin/templates/route.ts`: GET lista templates activos, POST crea (admin), DELETE elimina (admin) — 2026-03-25 *(Plan 4 F2.28)*
- `apps/web/src/components/chat/PromptTemplates.tsx`: selector de templates como Popover con título y preview del prompt; al elegir setea input + focusMode — 2026-03-25 *(Plan 4 F2.28)*
- `apps/web/src/app/actions/chat.ts`: Server Action `actionCreateSessionForDoc` — crea sesión con system message que restringe el contexto al documento específico — 2026-03-25 *(Plan 4 F2.27)*
- `apps/web/src/components/collections/CollectionsList.tsx`: botón "Chat" por colección + helper `handleChatWithDoc` para crear sesión anclada a un doc — 2026-03-25 *(Plan 4 F2.27)*
- `apps/web/src/app/(app)/collections/page.tsx`: página de colecciones Server Component con lista + metadata — 2026-03-25 *(Plan 4 F2.26)*
- `apps/web/src/components/collections/CollectionsList.tsx`: tabla de colecciones con acciones Chat y Eliminar (solo admin); formulario inline para crear nueva colección — 2026-03-25 *(Plan 4 F2.26)*
- `apps/web/src/app/api/rag/collections/route.ts`: POST para crear colección (solo admin) — 2026-03-25 *(Plan 4 F2.26)*
- `apps/web/src/app/api/rag/collections/[name]/route.ts`: DELETE para eliminar colección (solo admin) — 2026-03-25 *(Plan 4 F2.26)*
- `packages/db/src/schema.ts`: tabla `session_shares` (token UUID 64-char hex, expiresAt) — 2026-03-25 *(Plan 4 F2.25)*
- `packages/db/src/queries/shares.ts`: `createShare`, `getShareByToken`, `getShareWithSession`, `revokeShare` — 2026-03-25 *(Plan 4 F2.25)*
- `apps/web/src/app/api/share/route.ts`: POST crea token, DELETE revoca — 2026-03-25 *(Plan 4 F2.25)*
- `apps/web/src/app/(public)/share/[token]/page.tsx`: página pública read-only sin auth; muestra sesión + aviso de privacidad; 404 si token inválido/expirado — 2026-03-25 *(Plan 4 F2.25)*
- `apps/web/src/middleware.ts`: `/share/` agregado a PUBLIC_ROUTES — 2026-03-25 *(Plan 4 F2.25)*
- `apps/web/src/components/chat/ShareDialog.tsx`: Dialog para generar/copiar/revocar el link de compartir, con aviso de privacidad — 2026-03-25 *(Plan 4 F2.25)*
- `packages/db/src/schema.ts`: tabla `session_tags` (sessionId, tag, PK compuesta) — 2026-03-25 *(Plan 4 F2.24)*
- `packages/db/src/queries/tags.ts`: `addTag`, `removeTag`, `listTagsBySession`, `listTagsByUser` — 2026-03-25 *(Plan 4 F2.24)*
- `apps/web/src/components/chat/SessionList.tsx`: badges de etiquetas por sesión, input inline para agregar tags, filtro por tag en el header, bulk selection con toolbar (exportar/eliminar seleccionadas) — 2026-03-25 *(Plan 4 F2.24)*
- `apps/web/src/app/actions/chat.ts`: Server Actions `actionAddTag`, `actionRemoveTag` — 2026-03-25 *(Plan 4 F2.24)*
- `apps/web/src/components/layout/CommandPalette.tsx`: command palette con `cmdk` — grupos Navegar (chat, colecciones, upload, saved, audit, settings, admin), Apariencia (tema, zen), Sesiones recientes filtradas por texto — 2026-03-25 *(Plan 4 F2.23)*
- `apps/web/src/app/api/chat/sessions/route.ts`: endpoint GET que lista sesiones del usuario para la palette — 2026-03-25 *(Plan 4 F2.23)*
- `apps/web/src/hooks/useGlobalHotkeys.ts`: agregado `Cmd+K` para abrir command palette — 2026-03-25 *(Plan 4 F2.23)*
- `packages/db/src/schema.ts`: tabla `annotations` (selectedText, note, FK a session y message) — 2026-03-25 *(Plan 4 F2.22)*
- `packages/db/src/queries/annotations.ts`: `saveAnnotation`, `listAnnotationsBySession`, `deleteAnnotation` — 2026-03-25 *(Plan 4 F2.22)*
- `apps/web/src/components/chat/AnnotationPopover.tsx`: popover flotante al seleccionar texto en respuestas asistente — opciones Guardar, Preguntar sobre esto, Comentar con nota libre — 2026-03-25 *(Plan 4 F2.22)*
- `apps/web/src/app/actions/chat.ts`: Server Action `actionSaveAnnotation` — 2026-03-25 *(Plan 4 F2.22)*
- `apps/web/src/components/chat/CollectionSelector.tsx`: selector multi-checkbox de colecciones disponibles del usuario, persistido en localStorage; muestra las colecciones activas como Popover junto al input — 2026-03-25 *(Plan 4 F2.21)*
- `apps/web/src/hooks/useRagStream.ts`: acepta `collections?: string[]` para multi-colección; lo incluye como `collection_names` en el body del fetch — 2026-03-25 *(Plan 4 F2.21)*
- `apps/web/src/app/api/rag/generate/route.ts`: verificación de acceso a cada colección en `collection_names`; si alguna está denegada retorna 403 — 2026-03-25 *(Plan 4 F2.21)*
- `apps/web/src/app/api/rag/suggest/route.ts`: endpoint POST que genera 3-4 preguntas de follow-up; modo mock retorna sugerencias hardcodeadas, modo real usa el RAG server con prompt de generación + fallback al mock si falla — 2026-03-25 *(Plan 4 F2.20)*
- `apps/web/src/components/chat/RelatedQuestions.tsx`: chips de preguntas sugeridas debajo de la última respuesta; al clicar pone la pregunta en el input — 2026-03-25 *(Plan 4 F2.20)*
- `apps/web/src/components/chat/SourcesPanel.tsx`: panel colapsable de fuentes bajo cada respuesta asistente — muestra nombre del doc, fragmento (truncado a 2 líneas), relevance score como badge; visible solo cuando `sources.length > 0` — 2026-03-25 *(Plan 4 F2.19)*
- `apps/web/src/components/chat/ChatInterface.tsx`: integración de `SourcesPanel` bajo el contenido de cada mensaje asistente — 2026-03-25 *(Plan 4 F2.19)*

### Changed

- `apps/web/src/components/layout/AppShell.tsx`: reescrito como Server Component puro — delega toda la UI a `AppShellChrome` — 2026-03-25 *(Plan 4 Fase 0d)*

### Added

- `apps/web/src/components/chat/ThinkingSteps.tsx`: steps colapsables del proceso de razonamiento visibles durante streaming — simulación UI-side con timing (paso 1 inmediato, paso 2 a 700ms, paso 3 a 1500ms); se auto-colapsa 1.8s después de que el stream termina; cuando el RAG server exponga eventos SSE de tipo `thinking`, se conectan en `useRagStream` — 2026-03-25 *(Plan 4 F1.5)*
- `apps/web/src/lib/changelog.ts`: `parseChangelog(raw, limit)` extraída a utilidad testeable — 2026-03-25 *(tests Fase 1)*
- `apps/web/src/app/api/changelog/route.ts`: endpoint GET que parsea CHANGELOG.md y retorna las últimas 5 entradas + versión actual del package.json — 2026-03-25 *(Plan 4 F1.18)*
- `apps/web/src/components/layout/WhatsNewPanel.tsx`: Sheet lateral con entradas del CHANGELOG renderizadas con `marked`; `useHasNewVersion()` hook que compara versión actual con `localStorage["last_seen_version"]` — 2026-03-25 *(Plan 4 F1.18)*
- `apps/web/src/components/layout/NavRail.tsx`: logo "R" abre el panel "¿Qué hay de nuevo?" al clic; badge rojo unificado para `unreadCount > 0` o versión nueva no vista — 2026-03-25 *(Plan 4 F1.18)*
- `apps/web/src/components/chat/ChatInterface.tsx`: regenerar respuesta con botón `↻` (pone el último query del usuario en el input) F1.15; copy al portapapeles con ícono Check al confirmar F1.16; stats `{ms}ms · {N} docs` visibles al hover debajo del último mensaje asistente F1.17 — 2026-03-25
- `apps/web/src/hooks/useGlobalHotkeys.ts`: `Cmd+N` → navegar a `/chat`; `j/k` y Esc de sesiones diferidos a Fase 2 (requieren estado centralizado del panel) — 2026-03-25 *(Plan 4 F1.14)*
- `apps/web/src/lib/rag/client.ts`: `detectLanguageHint(text)` — detecta inglés (por palabras clave) y caracteres no-latinos; retorna instrucción "Respond in the same language as the user's message." si aplica — 2026-03-25 *(Plan 4 F1.13)*
- `apps/web/src/app/api/rag/generate/route.ts`: inyección de `detectLanguageHint` como system message cuando el último mensaje del usuario no está en español — 2026-03-25 *(Plan 4 F1.13)*
- `apps/web/src/app/api/notifications/route.ts`: endpoint GET que retorna eventos recientes de tipos `ingestion.completed`, `ingestion.error`, `user.created` (este último solo para admins) — 2026-03-25 *(Plan 4 F1.12)*
- `apps/web/src/hooks/useNotifications.ts`: polling cada 30s, emite toasts con sonner para notificaciones no vistas (gestionado en localStorage), retorna `unreadCount` — 2026-03-25 *(Plan 4 F1.12)*
- `apps/web/src/components/layout/NavRail.tsx`: badge rojo sobre el ícono "R" cuando `unreadCount > 0` — 2026-03-25 *(Plan 4 F1.12)*
- `packages/db/src/__tests__/saved.test.ts`: 13 tests de queries `saved_responses` (saveResponse, listSavedResponses, unsaveResponse, unsaveByMessageId, isSaved) contra SQLite en memoria — 2026-03-25 *(tests Fase 1)*
- `packages/shared/src/__tests__/focus-modes.test.ts`: 6 tests de estructura FOCUS_MODES (cantidad, IDs únicos, labels, systemPrompts, modo ejecutivo) — 2026-03-25 *(tests Fase 1)*
- `packages/shared/package.json`: agregado script `"test": "bun test src/__tests__"` para Turborepo — 2026-03-25 *(tests Fase 1)*
- `apps/web/src/lib/__tests__/export.test.ts`: 8 tests de `exportToMarkdown()` (título, colección, mensajes, fuentes, orden, vacío) — 2026-03-25 *(tests Fase 1)*
- `apps/web/src/lib/__tests__/changelog.test.ts`: 6 tests de `parseChangelog()` (Unreleased, versiones, contenido, límite, vacío, orden) — 2026-03-25 *(tests Fase 1)*
- `apps/web/src/lib/rag/__tests__/detect-language.test.ts`: 13 tests de `detectLanguageHint()` (español no inyecta, inglés inyecta, CJK/cirílico/árabe inyectan) — 2026-03-25 *(tests Fase 1)*
- `apps/web/src/hooks/useZenMode.ts`: hook `useZenMode()` — toggle con `Cmd+Shift+Z`, cierre con `Esc` — 2026-03-25 *(Plan 4 F1.11)*
- `apps/web/src/components/layout/AppShellChrome.tsx`: modo Zen oculta NavRail y SecondaryPanel; badge "ESC para salir" en `fixed bottom-4 right-4` — 2026-03-25 *(Plan 4 F1.11)*
- `packages/db/src/schema.ts`: tabla `saved_responses` (id, userId, messageId nullable, content, sessionTitle, createdAt) — 2026-03-25 *(Plan 4 F1.10)*
- `packages/db/src/queries/saved.ts`: `saveResponse`, `unsaveResponse`, `unsaveByMessageId`, `listSavedResponses`, `isSaved` — 2026-03-25 *(Plan 4 F1.10)*
- `packages/db/src/init.ts`: SQL de creación de tabla `saved_responses` con índice — 2026-03-25 *(Plan 4 F1.10)*
- `apps/web/src/app/actions/chat.ts`: Server Action `actionToggleSaved` (guarda/desuarda por messageId) — 2026-03-25 *(Plan 4 F1.10)*
- `apps/web/src/app/(app)/saved/page.tsx`: página `/saved` — lista de respuestas guardadas con empty state — 2026-03-25 *(Plan 4 F1.10)*
- `apps/web/src/lib/export.ts`: `exportToMarkdown()` (serializa sesión a MD con fuentes), `exportToPDF()` (window.print()), `downloadFile()` — 2026-03-25 *(Plan 4 F1.9)*
- `apps/web/src/components/chat/ExportSession.tsx`: Popover con opciones "Markdown" y "PDF (imprimir)" en el header del chat — 2026-03-25 *(Plan 4 F1.9)*
- `apps/web/src/components/chat/VoiceInput.tsx`: botón micrófono con Web Speech API — transcripción en tiempo real a `lang="es-AR"`, botón MicOff en rojo al grabar, fallback graceful si el browser no soporta SpeechRecognition (no renderiza nada) — 2026-03-25 *(Plan 4 F1.8)*
- `packages/shared/src/schemas.ts`: `FOCUS_MODES` + `FocusModeId` — 4 modos (detallado, ejecutivo, técnico, comparativo) con system prompt para cada uno — 2026-03-25 *(Plan 4 F1.7)*
- `apps/web/src/components/chat/FocusModeSelector.tsx`: selector de modos como pills, persistido en localStorage, `useFocusMode()` hook — 2026-03-25 *(Plan 4 F1.7)*
- `apps/web/src/app/api/rag/generate/route.ts`: prepend de system message según `focus_mode` recibido en el body — 2026-03-25 *(Plan 4 F1.7)*
- `apps/web/src/hooks/useRagStream.ts`: acepta `focusMode` en options y lo envía en el body del fetch — 2026-03-25 *(Plan 4 F1.7)*
- `apps/web/src/components/chat/ChatInterface.tsx`: integración de `ThinkingSteps` (F1.5), feedback shadcn (F1.6), modos de foco (F1.7), voice input (F1.8), ExportSession en header (F1.9), bookmark Guardar respuesta (F1.10) — 2026-03-25

### Fixed

- `apps/web/src/components/ui/theme-toggle.tsx`: hydration mismatch — el server renderizaba el `title` del botón con el tema default mientras el cliente ya conocía el tema guardado en localStorage; fix: `mounted` state con `useEffect` para evitar renderizar contenido dependiente del tema hasta después de la hidratación — 2026-03-25

### Changed

- `apps/web/src/app/globals.css`: design tokens reemplazados con paleta crema-índigo — tokens canónicos `--bg #FAFAF9`, `--sidebar-bg #F2F0F0`, `--nav-bg #18181B`, `--accent #7C6AF5`/`#9D8FF8` (dark), `--fg #18181B`/`#FAFAF9` (dark); aliases de compatibilidad apuntan a los canónicos vía `var()` para que los componentes existentes no requieran cambios; dark mode migrado de `@media prefers-color-scheme` a clase `.dark` en `<html>` (prerequisito de next-themes); directiva `@theme` agrega utilidades Tailwind para los nuevos tokens; agregado `@media print` para export de sesión (Fase 1) — 2026-03-25 *(Plan 4 Fase 0a)*

### Added

- `apps/web/src/components/layout/NavRail.tsx`: barra de íconos 44px, fondo `var(--nav-bg)` siempre oscuro, items con `Tooltip` de shadcn, ThemeToggle + logout al fondo, active state via `usePathname()` — 2026-03-25 *(Plan 4 Fase 0d)*
- `apps/web/src/components/layout/AppShellChrome.tsx`: Client Component wrapper de la shell — concentra estado de UI (zen mode, notificaciones, hotkeys en fases siguientes) — 2026-03-25 *(Plan 4 Fase 0d)*
- `apps/web/src/components/layout/SecondaryPanel.tsx`: panel contextual 168px — renderiza ChatPanel en `/chat`, AdminPanel en `/admin`, nada en otras rutas — 2026-03-25 *(Plan 4 Fase 0d)*
- `apps/web/src/components/layout/panels/ChatPanel.tsx`: panel de sesiones para rutas `/chat` con slot para SessionList (Fase 1) — 2026-03-25 *(Plan 4 Fase 0d)*
- `apps/web/src/components/layout/panels/AdminPanel.tsx`: nav admin con secciones "Gestión" y "Observabilidad" (extensible para Fase 2 sin rediseño) — 2026-03-25 *(Plan 4 Fase 0d)*
- `apps/web/src/components/providers.tsx`: ThemeProvider de next-themes (`attribute="class"`, `defaultTheme="light"`, `storageKey="rag-theme"`) — dark mode via clase `.dark` en `<html>` con script anti-FOUC automático — 2026-03-25 *(Plan 4 Fase 0c)*
- `apps/web/src/components/ui/theme-toggle.tsx`: botón Sun/Moon que alterna light/dark usando `useTheme()` de next-themes — 2026-03-25 *(Plan 4 Fase 0c)*
- `apps/web/components.json`: configuración shadcn/ui (style default, base color stone, Tailwind v4, paths `@/components/ui` y `@/lib/utils`) — 2026-03-25 *(Plan 4 Fase 0b)*
- `apps/web/src/lib/utils.ts`: función `cn()` de `clsx + tailwind-merge` — 2026-03-25 *(Plan 4 Fase 0b)*
- `apps/web/src/components/ui/`: 13 componentes shadcn instalados — button, input, textarea, dialog, popover, table, badge, avatar, separator, tooltip, sheet, sonner, command (cmdk) — 2026-03-25 *(Plan 4 Fase 0b)*
- `apps/web/src/app/layout.tsx`: `<Toaster />` de sonner + `<Providers>` de next-themes + `suppressHydrationWarning` en `<html>` — 2026-03-25 *(Plan 4 Fase 0b/0c)*

- `docs/workflows.md`: nuevo documento que centraliza los 7 flujos de trabajo del proyecto (desarrollo local, testing, git/commits, planificación, features nuevas, deploy, debugging con black box) — 2026-03-25

### Changed

- `CLAUDE.md`: corregido `better-sqlite3` → `@libsql/client`, "14 tablas" → "12 tablas", sección de tests expandida con todos los comandos, planes renombrados correctamente, nota sobre import estático del logger — 2026-03-25
- `docs/architecture.md`: corregido `better-sqlite3` → `@libsql/client`, referencia `ultra-optimize.md` → `ultra-optimize-plan1-birth.md`, documentada auth service-to-service, tabla de tablas actualizada a 12 — 2026-03-25
- `docs/onboarding.md`: comandos de test completos con conteo de tests por suite, nota sobre ubicación de `apps/web/.env.local`, referencia a `docs/workflows.md` — 2026-03-25
- `packages/db/package.json`: agregado script `"test": "bun test src/__tests__"` — Turborepo ahora corre esta suite en `bun run test` — 2026-03-25
- `packages/logger/package.json`: agregado script `"test": "bun test src/__tests__"` — 2026-03-25
- `packages/config/package.json`: agregado script `"test": "bun test src/__tests__"` — 2026-03-25
- `apps/web/package.json`: agregado script `"test": "bun test src/lib"` — 2026-03-25

### Fixed

- Pipeline de tests: `bun run test` desde la raíz ahora ejecuta los 79 tests unitarios via Turborepo — antes los workspaces no tenían script `"test"` y el pipeline completaba silenciosamente sin correr nada — 2026-03-25

### Changed

- `apps/web/src/components/chat/ChatInterface.tsx`: refactor — complejidad reducida de 48 a 22; lógica de fetch + SSE + abort extraída al hook `useRagStream`; `updateLastAssistantMessage` extraída como helper puro
- `apps/web/src/hooks/useRagStream.ts`: nuevo hook que encapsula fetch SSE, lectura del stream, abort controller y callbacks `onDelta`/`onSources`/`onError` — complejidad 19 (autónomo y testeable)
- `packages/logger/src/blackbox.ts`: refactor `reconstructFromEvents` — complejidad reducida de 34 a ~5; cada tipo de evento tiene handler nombrado (`handleAuthLogin`, `handleRagQuery`, `handleError`, `handleUserCreatedOrUpdated`, `handleUserDeleted`, `handleDefault`); despacho via `EVENT_HANDLERS` map

### Fixed

- `packages/db/src/queries/areas.ts`: `removeAreaCollection` ignoraba el parámetro `collectionName` en el WHERE — borraba todas las colecciones del área en lugar de solo la especificada; agregado `and(eq(areaId), eq(collectionName))` y actualizado import de `drizzle-orm` — 2026-03-25 *(encontrado con CodeGraphContext MCP, Plan 3 Fase 1a)*
- `apps/web/src/app/actions/areas.ts`: event types incorrectos en audit log — `actionCreateArea` emitía `"collection.created"`, `actionUpdateArea` emitía `"user.updated"`, `actionDeleteArea` emitía `"collection.deleted"`; corregidos a `"area.created"`, `"area.updated"`, `"area.deleted"` respectivamente — 2026-03-25 *(Plan 3 Fase 2a)*

### Added

- `packages/db/src/__tests__/areas.test.ts`: 8 tests de queries de áreas contra SQLite en memoria — `removeAreaCollection` (selectiva, cross-área, inexistente, última), `setAreaCollections` (reemplaza, vacía), `addAreaCollection` (default read, upsert) — 2026-03-25 *(Plan 3 Fase 1a)*

### Fixed

- `apps/web/src/app/api/auth/login/route.ts`: login con cuenta desactivada retornaba 401 en lugar de 403 — `verifyPassword` devuelve null para inactivos sin distinguir de contraseña incorrecta; agregado `getUserByEmail` check previo para detectar cuenta inactiva — 2026-03-25 *(encontrado en Fase 6e)*
- `apps/web/src/app/api/admin/db/reset/route.ts` y `seed/route.ts`: corregir errores de type-check (initDb inexistente, bcrypt-ts no disponible, null check en insert) — 2026-03-25
- `apps/web/src/lib/auth/jwt.ts`: agregar `iat` y `exp` al objeto retornado desde headers del middleware — 2026-03-25

- `packages/logger/src/backend.ts`: reemplazar lazy-load dinámico `import("@rag-saldivia/db" as any)` por import estático — en webpack/Next.js el dynamic import fallaba silenciosamente y ningún evento backend se persistía — 2026-03-25 *(encontrado en Fase 5)*
- `packages/logger/src/backend.ts`: `persistEvent` pasaba `userId=0` (SYSTEM_API_KEY) a la tabla events que tiene FK constraint a users.id — fix: escribir null cuando userId ≤ 0 — 2026-03-25 *(encontrado en Fase 5)*
- `packages/logger/package.json`: agregar `@rag-saldivia/db` como dependencia explícita del paquete logger — 2026-03-25

### Added

- `apps/web/src/components/chat/SessionList.tsx`: rename de sesión — botón lápiz en hover activa input inline; Enter/botón Confirmar guarda, Escape cancela; el nombre actualiza en la lista inmediatamente — 2026-03-25

- `apps/web/src/app/api/admin/permissions/route.ts`: endpoint POST/DELETE para asignar/quitar colecciones a áreas (necesario para flujos E2E) — 2026-03-25
- `apps/web/src/app/api/admin/users/[id]/areas/route.ts`: endpoint POST/DELETE para asignar/quitar áreas a usuarios (necesario para flujos E2E) — 2026-03-25
- `apps/web/src/app/api/admin/users/route.ts` y `[id]/route.ts`: endpoints GET/POST/DELETE/PATCH para gestión de usuarios desde la CLI — 2026-03-25
- `apps/web/src/app/api/admin/areas/route.ts` y `[id]/route.ts`: endpoints GET/POST/DELETE para gestión de áreas desde la CLI — 2026-03-25
- `apps/web/src/app/api/admin/config/route.ts` y `reset/route.ts`: endpoints GET/PATCH/POST para config RAG desde la CLI — 2026-03-25
- `apps/web/src/app/api/admin/db/migrate/route.ts`, `seed/route.ts`, `reset/route.ts`: endpoints de administración de DB desde la CLI — 2026-03-25

### Fixed

- `apps/web/src/middleware.ts`: agregar soporte para `SYSTEM_API_KEY` como auth de servicio — el CLI recibía 401 en todos los endpoints admin porque el middleware solo verificaba JWTs — 2026-03-25 *(encontrado en Fase 4b)*
- `apps/web/src/lib/auth/jwt.ts`: `extractClaims` leía Authorization header e intentaba verificarlo como JWT incluso cuando el middleware ya había autenticado via SYSTEM_API_KEY; ahora lee `x-user-*` headers del middleware si están presentes — 2026-03-25 *(encontrado en Fase 4b)*
- `apps/cli/src/client.ts`: corregir rutas de ingestion (`/api/ingestion/status` → `/api/admin/ingestion`) — 2026-03-25 *(encontrado en Fase 4d)*
- `apps/cli/src/commands/ingest.ts`: adaptador para respuesta `{ queue, jobs }` del API en lugar de array plano — 2026-03-25 *(encontrado en Fase 4d)*
- `apps/cli/src/commands/config.ts` + `apps/cli/src/index.ts`: agregar parámetro opcional `[key]` a `config get` para mostrar un parámetro específico — 2026-03-25 *(encontrado en Fase 4e)*

- `packages/config/src/__tests__/config.test.ts`: Fase 1d — 14 tests: loadConfig (env mínima, defaults, precedencia de env vars, MOCK_RAG como boolean, perfil YAML, perfil inexistente, error en producción), loadRagParams (defaults correctos, sin undefined), AppConfigSchema (validación: objeto mínimo, jwtSecret corto, logLevel inválido, URL inválida) — 2026-03-25

### Fixed

- `apps/web/src/app/actions/settings.ts`: agregar `revalidatePath("/", "layout")` para invalidar el layout al cambiar el nombre de perfil — 2026-03-25 *(encontrado en Fase 3h)*
- `apps/web/src/app/api/rag/generate/route.ts`: validación de `messages` faltante — body vacío `{}` retornaba 200 en lugar de 400; agregado guard que verifica que `messages` sea array no vacío antes de procesar — 2026-03-25 *(encontrado en Fase 2b)*
- `apps/web/src/app/api/admin/ingestion/[id]/route.ts`: DELETE con ID inexistente retornaba 200 en lugar de 404; agregado SELECT previo para verificar existencia antes del UPDATE — 2026-03-25 *(encontrado en Fase 2c)*

- Branch `experimental/ultra-optimize` iniciada — 2026-03-24
- Plan de trabajo `docs/plans/ultra-optimize.md` con seguimiento de tareas por fase — 2026-03-24
- `scripts/setup.ts`: script de onboarding cero-fricción con preflight check, instalación, migraciones, seed y resumen visual — 2026-03-24
- `.env.example` completamente documentado con todas las variables del nuevo stack — 2026-03-24
- `package.json` raíz mínimo para Bun workspaces con script `bun run setup` — 2026-03-24
- `Makefile`: nuevos targets `setup`, `setup-check`, `reset`, `dev` para el nuevo stack — 2026-03-24
- `.commitlintrc.json`: Conventional Commits enforced con scopes definidos para el proyecto — 2026-03-24
- `.husky/commit-msg` y `.husky/pre-push`: hooks de Git para validar commits y type-check — 2026-03-24
- `.github/workflows/ci.yml`: CI completo (commitlint, changelog check, type-check, tests, lint) en cada PR — 2026-03-24
- `.github/workflows/deploy.yml`: deploy solo en tag `v*` o workflow_dispatch — 2026-03-24
- `.github/workflows/release.yml`: mueve `[Unreleased]` a `[vX.Y.Z]` al publicar release — 2026-03-24
- `.github/pull_request_template.md`: PR template con sección obligatoria de CHANGELOG — 2026-03-24
- `.changeset/config.json`: Changesets para versionado semántico — 2026-03-24
- `turbo.json`: pipeline Turborepo (build → test → lint) con cache — 2026-03-24
- `package.json`: Bun workspaces root con scripts `dev`, `build`, `test`, `db:migrate`, `db:seed` — 2026-03-24
- `packages/shared`: schemas Zod completos (User, Area, Collection, Session, Message, IngestionJob, LogEvent, RagParams, UserPreferences, ApiResponse) — elimina duplicación entre Pydantic + interfaces TS — 2026-03-24
- `packages/db`: schema Drizzle completo (14 tablas), conexión singleton, queries por dominio (users, areas, sessions, events), seed, migración — 2026-03-24
- `packages/db`: tabla `ingestion_queue` reemplaza Redis — locking por columna `locked_at` — 2026-03-24
- `packages/config`: config loader TypeScript con Zod, deep-merge de YAMLs, overrides de env vars, admin-overrides persistentes — reemplaza `saldivia/config.py` — 2026-03-24
- `packages/logger`: logger estructurado (backend + frontend + blackbox + suggestions) con niveles TRACE/DEBUG/INFO/WARN/ERROR/FATAL — 2026-03-24
- `apps/web`: app Next.js 15 iniciada (package.json, tsconfig, next.config.ts) — 2026-03-24
- `apps/web/src/middleware.ts`: middleware de auth + RBAC en el edge — verifica JWT, redirecciona a login, bloquea por rol — 2026-03-24
- `apps/web/src/lib/auth/jwt.ts`: createJwt, verifyJwt, extractClaims, makeAuthCookie (cookie HttpOnly) — 2026-03-24
- `apps/web/src/lib/auth/rbac.ts`: hasRole, canAccessRoute, isAdmin, isAreaManager — 2026-03-24
- `apps/web/src/lib/auth/current-user.ts`: getCurrentUser, requireUser, requireAdmin para Server Components — 2026-03-24
- `apps/web`: endpoints auth (POST /api/auth/login, DELETE /api/auth/logout, POST /api/auth/refresh) — 2026-03-24
- `apps/web`: endpoint POST /api/log para recibir eventos del browser — 2026-03-24
- `apps/web`: página de login con form de email/password — 2026-03-24
- `apps/web`: Server Actions para usuarios (crear, eliminar, activar, asignar área) — 2026-03-24
- `apps/web`: Server Actions para áreas (crear, editar, eliminar con protección si hay usuarios) — 2026-03-24
- `apps/web`: Server Actions para chat (sesiones y mensajes) — 2026-03-24
- `apps/web`: Server Actions para settings (perfil, contraseña, preferencias) — 2026-03-24
- `apps/web/src/lib/rag/client.ts`: cliente RAG con modo mock, timeout, manejo de errores accionables — 2026-03-24
- `apps/web`: POST /api/rag/generate — proxy SSE al RAG Server con verificación de permisos — 2026-03-24
- `apps/web`: GET /api/rag/collections — lista colecciones con cache 60s filtrada por permisos — 2026-03-24
- `apps/web`: AppShell (layout con sidebar de navegación) — 2026-03-24
- `apps/web`: páginas de chat (lista de sesiones + interfaz de chat con streaming SSE + feedback) — 2026-03-24
- `apps/web`: página de admin/users con tabla y formulario de creación — 2026-03-24
- `apps/web`: página de settings con Perfil, Contraseña y Preferencias — 2026-03-24
- `apps/cli`: CLI completa con Commander + @clack/prompts + chalk + cli-table3 — 2026-03-24
- `apps/cli`: `rag status` — semáforo de servicios con latencias — 2026-03-24
- `apps/cli`: `rag users list/create/delete` — gestión de usuarios con wizard interactivo — 2026-03-24
- `apps/cli`: `rag collections list/create/delete` — gestión de colecciones — 2026-03-24
- `apps/cli`: `rag ingest start/status/cancel` — ingesta con barra de progreso — 2026-03-24
- `apps/cli`: `rag config get/set/reset` — configuración RAG — 2026-03-24
- `apps/cli`: `rag audit log/replay/export` — audit log y black box replay — 2026-03-24
- `apps/cli`: `rag db migrate/seed/reset`, `rag setup` — administración de DB — 2026-03-24
- `apps/cli`: modo REPL interactivo (sin argumentos) con selector de comandos — 2026-03-24
- `apps/web`: GET /api/audit — events con filtros (level, type, source, userId, fecha) — 2026-03-24
- `apps/web`: GET /api/audit/replay — black box reconstruction desde fecha — 2026-03-24
- `apps/web`: GET /api/audit/export — exportar todos los eventos como JSON — 2026-03-24
- `apps/web`: GET /api/health — health check público para la CLI y monitoring — 2026-03-24
- `apps/web`: página /audit con tabla de eventos filtrable por nivel y tipo — 2026-03-24
- `docs/architecture.md`: arquitectura completa del nuevo stack (servidor único, DB, auth, caching) — 2026-03-24
- `docs/blackbox.md`: guía del sistema de black box logging y replay — 2026-03-24
- `docs/cli.md`: referencia completa de todos los comandos de la CLI — 2026-03-24
- `docs/onboarding.md`: guía de 5 minutos para nuevos colaboradores — 2026-03-24
- `.gitignore`: agregado `.next/`, `.turbo/`, `logs/`, `data/*.db`, `bun.lockb` — 2026-03-24
- `apps/web/src/lib/auth/__tests__/jwt.test.ts`: tests completos del flujo de auth (JWT, RBAC) — 2026-03-24
- `apps/web/src/app/api/upload/route.ts`: endpoint de upload de archivos con validación de permisos y tamaño — 2026-03-24
- `apps/web/src/app/api/admin/ingestion/route.ts`: listado y cancelación de jobs de ingesta — 2026-03-24
- `apps/web/src/workers/ingestion.ts`: worker de ingesta en TypeScript con retry, locking SQLite, graceful shutdown — 2026-03-24
- `apps/web/src/app/(app)/upload/page.tsx`: página de upload con drag & drop — 2026-03-24
- `apps/web/src/hooks/useCrossdocDecompose.ts`: hook crossdoc portado de patches/ adaptado a Next.js — 2026-03-24
- `apps/web/src/hooks/useCrossdocStream.ts`: orquestación crossdoc (decompose → parallel queries → follow-ups → synthesis) — 2026-03-24
- `apps/web/src/app/(app)/admin/areas/page.tsx`: gestión de áreas con CRUD completo — 2026-03-24
- `apps/web/src/app/(app)/admin/permissions/page.tsx`: asignación colecciones → áreas con nivel read/write — 2026-03-24
- `apps/web/src/app/(app)/admin/rag-config/page.tsx`: config RAG con sliders y toggles — 2026-03-24
- `apps/web/src/app/(app)/admin/system/page.tsx`: estado del sistema con stats cards y jobs activos — 2026-03-24
- `packages/logger/src/rotation.ts`: rotación de archivos de log (10MB, 5 backups) — 2026-03-24
- `CLAUDE.md`: actualizado con el nuevo stack TypeScript — 2026-03-24
- `legacy/`: código del stack original (Python + SvelteKit) movido a carpeta `legacy/` — 2026-03-24
- `legacy/scripts/`: scripts bash y Python del stack original movidos a `legacy/` — 2026-03-24
- `legacy/pyproject.toml` + `legacy/uv.lock`: archivos Python movidos a `legacy/` — 2026-03-24
- `legacy/docs/`: docs del stack viejo movidos a `legacy/` (analysis, contributing, deployment, development-workflow, field-testing, plans-fase8, problems-and-solutions, roadmap, sessions, testing) — 2026-03-24
- `scripts/health-check.ts`: reemplaza health_check.sh — verifica servicios con latencias — 2026-03-24
- `README.md` y `scripts/README.md`: reescritos para el nuevo stack TypeScript — 2026-03-24
- `bun.lock`: lockfile de Bun commiteado para reproducibilidad de dependencias — 2026-03-24
- `scripts/link-libsql.sh`: script que crea symlinks de @libsql en apps/web/node_modules para WSL2 — 2026-03-24
- `scripts/test-login-final.sh`: script de test de los endpoints de auth — 2026-03-24
- `docs/plans/ultra-optimize-plan2-testing.md`: plan de testing granular en 7 fases creado — 2026-03-24
- `apps/web/src/types/globals.d.ts`: declaración de módulo `*.css` para permitir `import "./globals.css"` como side-effect sin error TS2882 — 2026-03-24
- `apps/web/src/lib/auth/__tests__/jwt.test.ts`: Fase 1a/1b — 17 tests: createJwt, verifyJwt (token inválido/firmado mal/expirado), extractClaims (cookie/header/sin token), makeAuthCookie (HttpOnly/Secure en prod), RBAC (getRequiredRole, canAccessRoute) — 2026-03-24
- `packages/db/src/__tests__/users.test.ts`: Fase 1c — 16 tests contra SQLite en memoria: createUser (email normalizado/rol/dup lanza error), verifyPassword (correcta/incorrecta/inexistente/inactivo), listUsers (vacío/múltiples/campos), updateUser (nombre/rol/desactivar), deleteUser (elimina usuario + CASCADE en user_areas) — 2026-03-24
- `packages/logger/src/__tests__/logger.test.ts`: Fase 1e — 24 tests: shouldLog por nivel (5), log.info/warn/error/debug/fatal/request no lanzan (7), output contiene tipo de evento (3), reconstructFromEvents vacío/orden/stats/usuarios/queries/errores (6), formatTimeline (3) — 2026-03-24

### Changed

- `apps/web/tsconfig.json`: excluir `**/__tests__/**` y `**/*.test.ts` del type-check — `bun:test` y asignación a `NODE_ENV` no son válidos en el contexto de `tsc` — 2026-03-24
- `package.json`: agregado `overrides: { "drizzle-orm": "^0.38.0" }` para forzar una sola instancia en la resolución de tipos — 2026-03-24
- `apps/web/package.json`: agregado `drizzle-orm` como dependencia directa para que TypeScript resuelva los tipos desde la misma instancia que `packages/db` — 2026-03-24
- `.gitignore`: agregado `*.tsbuildinfo` — 2026-03-24
- `package.json`: agregado campo `packageManager: bun@1.3.11` requerido por Turborepo 2.x — 2026-03-24
- `packages/db/package.json`: eliminado `type: module` para compatibilidad con webpack CJS — 2026-03-24
- `packages/shared/package.json`: eliminado `type: module` para compatibilidad con webpack CJS — 2026-03-24
- `packages/config/package.json`: eliminado `type: module` para compatibilidad con webpack CJS — 2026-03-24
- `packages/logger/package.json`: eliminado `type: module` para compatibilidad con webpack CJS — 2026-03-24
- `packages/*/src/*.ts`: eliminadas extensiones `.js` de todos los imports relativos (incompatibles con webpack) — 2026-03-24
- `packages/db/src/schema.ts`: agregadas relaciones Drizzle (`usersRelations`, `areasRelations`, `userAreasRelations`, etc.) necesarias para queries con `with` — 2026-03-24
- `packages/shared/src/schemas.ts`: campo `email` del `LoginRequestSchema` acepta `admin@localhost` (sin TLD) — 2026-03-24
- `apps/web/next.config.ts`: configuración completa para compatibilidad con WSL2 y monorepo Bun:
  - `outputFileTracingRoot: __dirname` para evitar detección incorrecta del workspace root
  - `transpilePackages` para paquetes workspace TypeScript
  - `serverExternalPackages` para excluir `@libsql/client` y la cadena nativa del bundling webpack
  - `webpack.externals` con función que excluye `libsql`, `@libsql/*` y archivos `.node` — 2026-03-24

### Fixed

- `apps/cli/package.json`: agregadas dependencias workspace faltantes `@rag-saldivia/logger` y `@rag-saldivia/db` — `audit.ts` importaba `formatTimeline`/`reconstructFromEvents` y `DbEvent` de esos paquetes pero Bun no los encontraba — 2026-03-24
- `packages/logger/package.json`: agregado export `./suggestions` faltante — `apps/cli/src/output.ts` importaba `getSuggestion` de `@rag-saldivia/logger/suggestions` sin que estuviera declarado en `exports` — 2026-03-24
- `apps/web/src/middleware.ts`: agregado `/api/health` a `PUBLIC_ROUTES` — el endpoint retornaba 401 al CLI y a cualquier sistema de monitoreo externo — 2026-03-24 *(encontrado en Fase 0)*
- `apps/web/src/lib/auth/__tests__/jwt.test.ts`: `await import("../rbac.js")` dentro del callback de `describe` lanzaba `"await" can only be used inside an "async" function` — movido al nivel del módulo junto con los demás imports — 2026-03-24 *(encontrado en Fase 1a)*
- `apps/web/src/lib/auth/__tests__/jwt.test.ts`: test `makeAuthCookie incluye Secure en producción` referenciaba `validClaims` definido en otro bloque `describe` — reemplazado por claims inline en el test — 2026-03-24 *(encontrado en Fase 1b)*
- `packages/logger/src/__tests__/logger.test.ts`: mismo patrón `await import` dentro de callbacks `describe` (×3 bloques) — todos los imports movidos al nivel del módulo — 2026-03-24 *(encontrado en Fase 1e)*
- `packages/logger/src/__tests__/logger.test.ts`: tests de formato JSON en producción asumían que cambiar `NODE_ENV` post-import afectaría el logger, pero `isDev` se captura en `createLogger()` al momento del import — tests rediseñados para verificar el output directamente y testear `formatJson` con datos conocidos — 2026-03-24 *(encontrado en Fase 1e)*
- `packages/db/src/queries/users.ts`: reemplazado `Bun.hash()` con `crypto.createHash('sha256')` — `Bun` global no disponible en el contexto `tsc` de `apps/web`; `crypto` nativo es compatible con Node.js y Bun — 2026-03-24
- `apps/web/src/workers/ingestion.ts`: reemplazado `Bun.file()` / `file.exists()` / `file.arrayBuffer()` con `fs/promises` `access` + `readFile` — mismo motivo que `Bun.hash` — 2026-03-24
- `apps/web/src/components/audit/AuditTable.tsx`: eliminado `import chalk from "chalk"` — importado pero nunca usado; chalk es un paquete CLI y no pertenece a un componente React — 2026-03-24
- `apps/web/src/lib/auth/current-user.ts`: `redirect` de `next/navigation` importado estáticamente en lugar de con `await import()` dinámico — TypeScript infiere correctamente que `redirect()` retorna `never`, resolviendo el error TS2322 de `CurrentUser | null` — 2026-03-24
- `packages/logger/src/backend.ts`: corregidos tres errores de tipos: (1) tipo de `_writeToFile` ajustado a `LogFilename` literal union; (2) TS2721 "cannot invoke possibly null" resuelto capturando en variable local antes del `await`; (3) import dinámico de `@rag-saldivia/db` casteado para evitar TS2307 — 2026-03-24
- `packages/logger/src/blackbox.ts`: eliminado `import type { DbEvent } from "@rag-saldivia/db"` — reemplazado por definición inline para cortar la dependencia `logger → db` que causaba TS2307 en el contexto de `apps/web` — 2026-03-24
- `.husky/pre-push`: reemplazado `bun` por ruta dinámica `$(which bun || echo /home/enzo/.bun/bin/bun)` — el PATH de husky en WSL2 no incluye `~/.bun/bin/` y el hook bloqueaba el push — 2026-03-24

- DB: migrado de `better-sqlite3` (requería compilación nativa con node-gyp, falla en Bun) a `@libsql/client` (JS puro, sin compilación, compatible con Bun y Node.js) — 2026-03-24
- DB: creado `packages/db/src/init.ts` con SQL directo (sin drizzle-kit) para inicialización en entornos sin build tools — 2026-03-24
- DB: `packages/db/src/migrate.ts` actualizado para usar `init.ts` en lugar del migrador de drizzle-kit — 2026-03-24
- DB: agregado `bcrypt-ts` como dependencia explícita de `packages/db` — 2026-03-24
- DB: agregado `@libsql/client` como dependencia de `packages/db` reemplazando `better-sqlite3` — 2026-03-24
- DB: conexión singleton migrada a `drizzle-orm/libsql` con `createClient({ url: "file:..." })` — 2026-03-24
- Next.js: Next.js 15.5 detectaba `/mnt/c/Users/enzo/package-lock.json` (filesystem Windows montado en WSL2) como workspace root, ignorando `src/app/`. Resuelto renombrando ese `package-lock.json` abandonado a `.bak` — 2026-03-24
- Next.js: resuelta incompatibilidad entre `transpilePackages` y `serverExternalPackages` al usar los mismos paquetes en ambas listas — 2026-03-24
- Next.js: webpack intentaba bundear `@libsql/client` → `libsql` (addon nativo) → cargaba `README.md` como módulo JS. Resuelto con `webpack.externals` personalizado — 2026-03-24
- Next.js: `@libsql/client` no era accesible en runtime de Node.js (los paquetes de Bun se guardan en `.bun/`, no en `node_modules/` estándar). Resuelto creando symlinks en `apps/web/node_modules/@libsql/` — 2026-03-24
- Next.js: conflicto de instancias de `drizzle-orm` (TypeError `referencedTable` undefined) al excluirlo del bundling. Resuelto manteniéndolo en el bundle de webpack — 2026-03-24
- Next.js: `.env.local` debe vivir en `apps/web/` (el directorio del proyecto), no solo en la raíz del monorepo — 2026-03-24
- Bun workspaces en WSL2: `bun install` en filesystem Windows (`/mnt/c/`) no crea symlinks en `node_modules/.bin/`. Resuelto clonando el repo en el filesystem nativo de Linux (`~/rag-saldivia/`). **En Ubuntu nativo (deployment target) este problema no existe** — 2026-03-24
- `scripts/link-libsql.sh`: workaround específico de WSL2 para crear symlinks de `@libsql` manualmente. **No necesario en Ubuntu nativo ni en producción (workstation Ubuntu 24.04)** — 2026-03-24

---

<!-- Instrucciones:
  - Cada tarea completada genera una entrada en [Unreleased] antes de hacer commit
  - Al publicar una release, [Unreleased] se mueve a [vX.Y.Z] con la fecha
  - Categorías: Added | Changed | Deprecated | Removed | Fixed | Security
-->

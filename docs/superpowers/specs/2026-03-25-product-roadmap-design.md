# RAG Saldivia — Product Roadmap & Design Spec

**Fecha:** 2026-03-25  
**Branch:** experimental/ultra-optimize  
**Sesión de brainstorming:** 50 features en 4 fases, menor a mayor dificultad

---

## Decisiones de diseño base

### Identidad visual

| Token | Light | Dark |
|---|---|---|
| Background | `#FAFAF9` (crema cálida) | `#1C1917` (stone oscuro) |
| Sidebar | `#F2F0F0` (gris perlado) | `#292524` (stone medio) |
| Acento principal | `#7C6AF5` (índigo suave) | `#9D8FF8` (índigo claro) |
| Texto | `#18181B` | `#FAFAF9` |
| Borde | `#E3E1E0` | `#3F3B38` |

**Principio:** light mode como primario, dark mode como alternativa siempre disponible. Warm neutral — no frío ni corporativo. Referencia: Claude.ai.

### Librería de componentes

**shadcn/ui** sobre Radix UI + Tailwind CSS.

- Componentes viven dentro del repo (`components/ui/`) — sin black-box
- Tailwind config centraliza los tokens de color
- Compatible con Next.js 15 App Router y Server Components
- Ecosystem: Radix primitives, cmdk (command palette), vaul (drawer), sonner (toasts)

### Estructura de layout

```
┌──────────────────────────────────────────────────────┐
│  Nav (44px) │  Panel secundario (~160px)  │  Contenido  │
│             │                             │             │
│  íconos     │  /chat  → lista sesiones    │  chat /     │
│  oscuros    │  /admin → nav admin         │  tablas /   │
│  fixed      │  contextual por ruta        │  formularios│
└──────────────────────────────────────────────────────┘
```

- **Nav izquierda:** 44px, fondo oscuro (`#18181B`), íconos con tooltips. Items: Chat, Colecciones, Upload, Audit, Admin (según rol).
- **Panel secundario:** ~160px, fondo sidebar. En `/chat`: lista de sesiones con grupos por fecha, botón nueva sesión, búsqueda. En `/admin`: nav con sub-ítems — secciones actuales (Usuarios, Áreas, Permisos, Config RAG, Sistema) más las nuevas de Fase 2 (Analytics, Monitoring de ingesta, Brechas de conocimiento, Historial de colecciones) agrupadas bajo encabezados "Gestión" y "Observabilidad". El plan de layout de Fase 0 debe incluir el nav admin extensible sin rediseño.
- **Contenido principal:** `flex-1`, ocupa el resto. Scroll independiente.

---

## Roadmap — 50 features en 4 fases

### Fase 0 — Fundación (4 features)
> Prerequisito de todo. Sin esto el diseño no existe.

| # | Feature | Descripción | Esfuerzo |
|---|---|---|---|
| 1 | **Design tokens light/dark** | Paleta crema-índigo en variables CSS + Tailwind config. Clase `dark` en `<html>` via next-themes. | ● |
| 2 | **shadcn/ui setup** | Instalar y configurar: Button, Input, Textarea, Dialog, Popover, Table, Badge, Avatar, Separator, Tooltip, Toast (sonner), Command (cmdk), Sheet (drawer). | ● |
| 3 | **Dark mode toggle** | next-themes, persistido en cookie para SSR sin flash. Toggle en el nav inferior. | ● |
| 4 | **Dual sidebar layout** | Reescribir AppShell con estructura nav + panel contextual + contenido. Panel cambia según ruta activa. | ●● |

---

### Fase 1 — Quick wins (14 features, todas ●)
> Impacto alto, esfuerzo bajo. Pueden ir en paralelo una vez que Fase 0 está.

| # | Feature | Descripción |
|---|---|---|
| 5 | **Thinking steps visibles** | Mientras el modelo genera, mostrar steps colapsables: "Buscando en colección...", "Encontré 4 fragmentos relevantes...", "Sintetizando...". Componente de streaming ya existe en `useRagStream`. |
| 6 | **Feedback 👍/👎** | Botones en cada mensaje del asistente. Persiste en tabla `feedback` en DB. Visible en el analytics dashboard (Fase 2). |
| 7 | **Modos de foco** | Selector antes del input: Detallado / Ejecutivo / Técnico / Comparativo. Cada modo inyecta instrucciones al system prompt. |
| 8 | **Voz en input** | Botón micrófono en el input. Web Speech API nativa (Chrome/Edge). Transcripción en tiempo real. Fallback graceful si el browser no lo soporta. |
| 9 | **Export de sesión** | Botón "Exportar" en la sesión activa. Opciones: PDF (print CSS) y Markdown (serializar mensajes). Incluye fuentes citadas si las hay. |
| 10 | **Respuestas guardadas / pinned** | Ícono bookmark en cada respuesta. Vista `/saved` muestra todas las guardadas. Persiste en tabla `saved_responses` en DB. |
| 11 | **Modo Zen** | Atajo `Cmd+Shift+Z`. Oculta nav lateral y panel secundario. Solo el chat. Badge de salida con `Esc`. Ideal para trabajo profundo. |
| 12 | **Notificaciones** | Toast + badge en nav cuando: ingestion job completa, ingestion falla, nuevo usuario registrado (admin). In-app, no push. Sistema de eventos ya existe en `packages/logger`. |
| 13 | **Multilenguaje automático** | Detectar idioma del query con `Intl` / simple heurística. Inyectar en system prompt: "Respond in the same language as the user's message". Zero config. |
| 14 | **Atajos de teclado globales** | `Cmd+N` nueva sesión, `Cmd+K` command palette (Fase 2), `j/k` navegar sesiones, `Esc` cerrar paneles. Implementar con `useHotkeys`. |
| 15 | **Regenerar respuesta** | Botón `↻` en cada respuesta del asistente. Re-envía el mismo query. Opcional: permite cambiar el modo de foco antes de regenerar. |
| 16 | **Copy con formato** | Botón copy en cada respuesta. Opciones: Markdown (raw), Texto plano (strip markdown), HTML. Clipboard API. |
| 17 | **Stats de query visibles** | Debajo de cada respuesta, en pequeño y colapsado: tiempo de respuesta (ms), cantidad de documentos consultados, tokens usados. Solo visible al hover. |
| 18 | **"¿Qué hay de nuevo?" in-app** | Fuente: `CHANGELOG.md` del repo, parseado en build time como static data (o API route que lo lee en runtime). Punto rojo en el avatar cuando la versión en `package.json` es mayor a la última vista, persistida en `localStorage` como `last_seen_version`. Panel lateral con las últimas 5 entradas del CHANGELOG. |

---

### Fase 2 — Esfuerzo medio (20 features, todas ●●)
> Requieren diseño de componente no trivial o cambios en el backend.

#### Chat

| # | Feature | Descripción |
|---|---|---|
| 19 | **Panel de fuentes / citas** | Panel lateral derecho (o debajo del mensaje) con los documentos fuente usados. Muestra: nombre del doc, fragmento relevante, relevance score. Datos vienen del RAG server response. |
| 20 | **Preguntas relacionadas** | Después de cada respuesta, 3-4 sugerencias de follow-up generadas por el modelo. Click en una la envía como nuevo mensaje. |
| 21 | **Multi-colección en query** | Selector multi-checkbox de colecciones antes del input. El query se envía con la lista de colecciones activas. |
| 22 | **Anotar fragmentos de respuesta** | Seleccionás texto de una respuesta → popover con opciones: "Guardar fragmento", "Hacer pregunta sobre esto", "Comentar". Anclado al fragmento exacto. |

#### Navegación

| # | Feature | Descripción |
|---|---|---|
| 23 | **Command palette Cmd+K** | Usando `cmdk`. Acciones: nueva sesión, cambiar colección, navegar a ruta, buscar en historial, cambiar modo. Extensible con grupos de comandos. |
| 24 | **Etiquetas en sesiones + bulk** | Tags `#legal`, `#urgente` en sesiones. Filtro por tag en la lista. Selección múltiple con checkbox → acciones bulk: exportar, archivar, eliminar, compartir. |
| 25 | **Compartir sesión** | Genera token único (32 bytes hex, opaco). Ruta pública `/share/[token]` muestra la sesión en modo read-only sin autenticación. Expiración configurable (default: 7 días). **Política de datos:** el usuario es responsable de no compartir sesiones con información sensible. La UI muestra un aviso al generar el link. Los admins pueden desactivar la feature por área o globalmente. |

#### Conocimiento

| # | Feature | Descripción |
|---|---|---|
| 26 | **Colecciones desde UI** | Página `/collections` con lista completa. Acciones: crear (nombre + descripción), eliminar (con confirmación), ver documentos individuales, estado de ingesta. |
| 27 | **Chat con documento específico** | Desde la lista de documentos de una colección, botón "Preguntar sobre este doc". Abre nueva sesión con el doc pre-seleccionado como único contexto. |
| 28 | **Templates de query** | Admin crea templates con título, prompt base y modo de foco recomendado. Usuarios los ven como punto de partida en el input. Persiste en tabla `prompt_templates`. |

#### Admin

| # | Feature | Descripción |
|---|---|---|
| 29 | **Ingestion monitoring mejorado** | Vista kanban de jobs: Pendiente / En progreso / Completado / Error. Progreso en tiempo real via SSE. Botón retry en jobs fallidos. Detalle de error expandible. |
| 30 | **Analytics dashboard** | Página `/admin/analytics`: queries por día (gráfico), colecciones más consultadas (barras), distribución feedback +/- (donut), usuarios más activos (tabla). |
| 31 | **Brechas de conocimiento** | Tabla de queries que el RAG respondió con baja confianza o sin fuentes. Definición operativa: un query se marca como "brecha" si (a) el response no incluye campo `sources` o el array está vacío, **o** (b) el campo `confidence` del response del blueprint es < 0.4 (si existe). Si el blueprint no expone confidence, se usa heurística: respuesta < 80 tokens con cualquier pattern de "no sé / no encontré". El umbral exacto se ajusta en el plan de implementación. Agrupados por tema via clustering simple (TF-IDF). Exportable. Guía directa de qué ingestar. |
| 32 | **Historial de colecciones** | Cada ingesta registrada como "commit": timestamp, usuario, docs agregados/eliminados, tamaño resultante. Vista timeline por colección. |
| 33 | **Informes programados** | Admin configura: query + colección + schedule (cron) + destino (email o Guardados). Ejecutado por el worker de ingestion. **Email:** requiere SMTP externo configurado vía variables de entorno (`SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`) — es alcance nuevo, no existe en el stack actual. Si SMTP no está configurado, solo funciona el destino Guardados. |

#### Power user / UX

| # | Feature | Descripción |
|---|---|---|
| 34 | **Vista dividida (split view)** | Botón en la toolbar del chat: divide la vista horizontalmente en dos paneles de chat independientes. Cada uno mantiene su propia sesión y selección de colección. |
| 35 | **Drag & drop al chat** | Drop zone sobre el área de chat. PDF soltado → ingesta on-the-fly en colección temporal `@temp-[sessionId]`. La sesión usa esa colección automáticamente. **Prerequisito de viabilidad:** el plan de esta feature debe validar que el NVIDIA RAG Blueprint soporta crear colecciones efímeras sin TTL fijo en Milvus y que el worker puede procesar documentos < 100MB en tiempo real (< 30s). Si no es viable sin rediseño mayor, la feature se redefine como "subir al flujo de Upload normal con pre-selección de sesión". |
| 36 | **Rate limiting por área/usuario** | Admin configura en Settings: máximo queries/hora por usuario y/o por área. Implementado con tabla `rate_limits` en DB + check en el endpoint `/api/rag/generate`. |
| 37 | **Onboarding interactivo** | Al primer login (flag `onboarding_completed` en tabla users): tour tooltip-driven de 5 pasos. Saltable. Reiniciable desde Settings. Usa `driver.js` o implementación propia. |
| 38 | **Webhooks salientes** | Admin configura URLs + eventos a escuchar (ingestion.completed, query.low_confidence, user.created). El worker hace POST con payload JSON cuando el evento ocurre. |

---

### Fase 3 — Alta complejidad (12 features, todas ●●●)
> Cada una requiere arquitectura nueva, nuevas tablas, o integración con sistemas externos.

| # | Feature | Descripción | Dependencias |
|---|---|---|---|
| 39 | **Búsqueda universal** | Índice full-text sobre sesiones + fragmentos de docs + templates + guardados. Resultados en tiempo real dentro del command palette. | Fase 2: Cmd+K |
| 40 | **Preview de doc inline** | Click en fuente citada → panel lateral con el PDF renderizado (PDF.js) y el fragmento exacto resaltado en amarillo. | Fase 2: panel de fuentes |
| 41 | **Proyectos con contexto** | Entidad nueva: `Project`. Agrupa sesiones, tiene colecciones asignadas e instrucciones custom. Todas las sesiones del proyecto heredan el contexto. Panel en sidebar. | — |
| 42 | **Artifacts panel** | Cuando el stream de respuesta incluye un bloque delimitado con `:::artifact` (tipo: `document`, `table`, o `code`) el panel lateral se activa con ese contenido. El blueprint/server debe emitir ese marcador; si no lo hace, se detecta heurísticamente (bloque markdown de > 40 líneas o tabla de > 5 columnas). Guardable, exportable, versionable. El contrato exacto del stream se define en el plan de implementación de esta feature. | Fase 3: Proyectos |
| 43 | **Bifurcación de conversaciones** | En cualquier mensaje: botón "Bifurcar desde aquí". Crea una nueva sesión con el historial hasta ese punto. Ambas sesiones quedan vinculadas con indicador visual. | — |
| 44 | **Memoria de usuario** | Tabla `user_memory` con preferencias inferidas y explícitas. UI para ver y editar qué recuerda el sistema. Inyectado en cada query como contexto adicional. | — |
| 45 | **Superficie proactiva** | Job periódico que cruza documentos nuevos en colecciones con el historial de queries del usuario. Genera notificaciones "X docs nuevos podrían interesarte". | Fase 1: notificaciones, Fase 2: analytics |
| 46 | **Grafo de documentos** | Página `/collections/[name]/graph`. Visualización D3/Visx de similitud semántica entre docs (embeddings from Milvus). Nodos clicables que abren el doc. | — |
| 47 | **SSO (Google / Azure AD)** | NextAuth.js v5 con providers OIDC/SAML 2.0. **Coexistencia con el stack actual:** el JWT propio (`jose`) se reemplaza por el session token de NextAuth para usuarios SSO; los usuarios con password existentes siguen funcionando (modo mixto). Al primer login SSO se crea un registro en la tabla `users` con `sso_provider` y `sso_subject` nuevos campos; `password_hash` queda null. El middleware RBAC y las cookies HttpOnly no cambian. No se soportan cuentas vinculadas (un email = una cuenta). Admin configura el provider en Config RAG. | — |
| 48 | **Auto-ingesta externa** | Conectores para Google Drive, SharePoint, Confluence. Admin configura credenciales OAuth + colección destino + schedule. Worker de ingestion gestiona el sync. | Fase 2: webhooks |
| 49 | **Bot Slack / Teams** | App de Slack/Teams que recibe mensajes, llama al API interno con SYSTEM_API_KEY + userId del solicitante, responde con la respuesta + fuentes. Respeta RBAC. | — |
| 50 | **Extracción estructurada a tabla** | El usuario define campos ("Nombre, Fecha, Monto"). El sistema procesa todos los docs de la colección extrayendo esos campos. Resultado: tabla exportable como CSV/Excel. | — |

**Hito mínimo de Fase 3** (criterio de deploy): features 39, 40, 41, 47. El resto de Fase 3 (42–46, 48–50) son incrementales post-hito; pueden completarse en sub-sprints dentro de la misma fase sin bloquear el deploy. Cada feature de Fase 3 tiene su propio plan antes de arrancar.

---

## Stack técnico additions

| Necesidad | Librería recomendada |
|---|---|
| Command palette | `cmdk` (ya usada por shadcn) |
| Hotkeys | `@github/hotkeys` o `useHotkeys` |
| Toasts / notificaciones | `sonner` |
| Onboarding tour | `driver.js` |
| PDF viewer | `react-pdf` (PDF.js wrapper) |
| Charts (analytics) | `recharts` |
| Graph visualization | `@visx/network` o `d3` |
| SSO | `next-auth` v5 |
| Scheduled jobs | Extender worker de ingestion existente |

---

## Principios de implementación

1. **Menor a mayor dificultad siempre** — Fase 0 → 1 → 2 → 3, sin saltear.
2. **Quick wins primero dentro de cada fase** — dentro de Fase 2, empezar por lo ●● más simple.
3. **No romper lo existente** — cada feature nueva tiene su propio plan antes de tocar código.
4. **Tests antes del merge** — cada feature de Fase 0-2 lleva tests unitarios. Fase 3 lleva E2E Playwright.
5. **Design system primero** — Fase 0 se completa antes de construir cualquier componente nuevo.
6. **Deploy incremental** — Fases 0+1 son suficientes para el primer deploy de esta branch sobre main.

---

## Criterio de done por fase

| Fase | Criterio |
|---|---|
| **Fase 0** | `bun run test` pasa, layout nuevo funciona en light y dark, sin regresiones |
| **Fase 1** | Las 14 features están accesibles y testeadas. Ready para primer deploy. |
| **Fase 2** | Las 20 features están completas. Analytics muestra datos reales. |
| **Fase 3** | Features críticas (búsqueda, preview, proyectos) completas. SSO funciona en staging. |

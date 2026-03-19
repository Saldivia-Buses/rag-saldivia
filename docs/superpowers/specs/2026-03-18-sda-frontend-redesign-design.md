# SDA Frontend — Rediseño Profesional Completo

**Fecha:** 2026-03-18
**Proyecto:** rag-saldivia / services/sda-frontend
**Stack:** SvelteKit 5, Tailwind CSS 4, TypeScript
**Estado:** Aprobado por usuario

---

## 1. Contexto y problemas actuales

El frontend actual tiene los siguientes problemas críticos identificados en code review:

### Estabilidad / Crashes
- Sin `<svelte:boundary>` — cualquier error en render crashea toda la app
- Sin try/catch en page server loaders
- Sin retry logic en SSE streaming del chat
- `+error.svelte` pages vacías sin UI útil

### UX / Legibilidad
- Fuentes de 7–9px ilegibles (chat usa `text-[7px]`, `text-[9px]`)
- Sidebar de 40px — demasiado angosta para ser usable
- Sin Markdown rendering en respuestas del LLM (código, listas, tablas se muestran como texto plano)
- Sin auto-scroll al recibir mensajes nuevos
- Sin estados vacíos, sin skeleton loading, sin toasts

### Arquitectura
- 800+ colores hardcodeados en lugar de CSS variables (que ya existen pero no se usan)
- Sin design system — 0 componentes reutilizables en `$lib/components/ui/`
- Sidebar tiene su propio mini-historial duplicando lógica del layout

### Features faltantes
- Admin: solo crear/desactivar usuarios — sin editar rol, área, estado
- Colecciones: sin crear ni borrar colecciones desde UI
- Upload de documentos: no existe en el frontend
- Auditoría: sin filtros, sin paginación, sin export
- Sin RAG configuration panel
- Sin gestión de áreas

---

## 2. Decisiones de diseño aprobadas

### Paleta — Saldivia Warm Adaptive
Color de marca extraído de `saldiviabuses.com.ar` (preloader CSS):
`rgb(32, 147, 211)` → **`#2093D3`** (azul azure Saldivia)

```css
/* ===== DARK MODE (default) ===== */
--bg-base:      #181510;   /* fondo principal */
--bg-surface:   #1e1a14;   /* cards, panels */
--bg-hover:     #2a2418;   /* hover states */
--bg-accent-dim:#143248;   /* user message bubbles */
--border:       #2e2820;   /* bordes */
--border-focus: #3a3025;   /* inputs con foco */
--text:         #ede8e0;   /* texto principal */
--text-muted:   #8c7b6a;   /* texto secundario */
--text-faint:   #4a4035;   /* labels, placeholders */
--accent:       #2093d3;   /* azul Saldivia — único acento */
--accent-light: #4db8f0;   /* accent sobre dark surfaces */
--success:      #4ade80;
--warning:      #fbbf24;
--danger:       #f87171;
--info:         #60a5fa;

/* ===== LIGHT MODE ===== */
--bg-base:      #faf8f4;
--bg-surface:   #f2ede3;
--bg-hover:     #e6dfd4;
--bg-accent-dim:#ddedf7;
--border:       #ddd5c7;
--border-focus: #c8bfb3;
--text:         #1a2030;
--text-muted:   #8c7b6a;   /* mismo en ambos modos */
--text-faint:   #b0a090;
--accent:       #2093d3;   /* mismo en ambos modos */
--accent-light: #1a7ab5;
```

### Tipografía
- Font: **Roboto** (ya usado en saldiviabuses.com.ar)
- Base: 14px / line-height 24px
- Mínimo absoluto: 11px para labels y metadata
- Headings: 13px (sm), 15px (md), 18px (lg), 22px (xl)

### Layout general
- Sidebar colapsable: **220px** expandida ↔ **56px** colapsada (solo íconos)
- Topbar: 44px de altura fija
- Chat: 3 paneles — historial (200px) | mensajes (flex) | fuentes (220px)

---

## 3. Fases de implementación

### FASE 1 — Fundación (prerequisito para todo)

**Objetivo:** Estabilidad, design system y layout base. Sin esto las fases siguientes son inestables.

**Archivos a crear — `src/lib/components/ui/`:**
- `Button.svelte` — variantes: primary, secondary, danger, ghost; states: loading, disabled
- `Input.svelte` — con label, error message, icon slot
- `Select.svelte` — styled, con opciones
- `Badge.svelte` — variantes de color semánticas
- `Modal.svelte` — backdrop, slot de contenido, close on Esc/backdrop
- `Card.svelte` — surface container
- `Skeleton.svelte` — shimmer animation configurable
- `Toast.svelte` + `toast.svelte.ts` store — éxito/error/info/warning, auto-dismiss 4s
- `Tooltip.svelte` — on hover, positioning automático
- `Toggle.svelte` — dark/light mode (usa mode-watcher ya instalado)

**CSS:**
- `src/app.css` — reescribir con design tokens completos (dark + light vars)
- Tipografía base, reset, scrollbar customizado

**Layout:**
- `Sidebar.svelte` — rediseño completo, 220px expandida / 56px colapsada, secciones (Principal / Admin / Cuenta), labels visibles, footer con avatar + theme toggle
- `src/routes/(app)/+layout.svelte` — integrar toggle y sidebar colapsable

**Estabilidad:**
- `<svelte:boundary>` en `+layout.svelte` y en el componente de chat principal
- `src/routes/+error.svelte` y `src/routes/(app)/+error.svelte` — UI completa con botón retry y home
- Try/catch en todos los `+page.server.ts` loaders
- `hooks.server.ts` — handleError mejorado con logging

**Features UX:**
- Dark/Light toggle persistido en localStorage via mode-watcher
- Empty states con sugerencias en `/chat`
- ⌘K command palette básico (navegación entre páginas)

---

### FASE 2 — Chat Pro

**Objetivo:** Experiencia de chat a nivel ChatGPT/Perplexity.

**Dependencias:** Fase 1 completa.

**Nuevas dependencias npm:**
- `marked` — Markdown → HTML
- `highlight.js` — syntax highlighting en bloques de código

**Cambios en `src/routes/(app)/chat/[id]/+page.svelte`:**
- Eliminar mini-sidebar de historial inline (pasa al layout o componente separado)
- Markdown rendering en respuestas del assistant via `marked`
- Syntax highlighting en bloques `code` con `highlight.js`
- Auto-scroll inteligente: sigue al fondo automáticamente, pero frena si el usuario scrolleó arriba
- Botones por mensaje: Copiar respuesta, Regenerar, 👍, 👎
- Stop streaming: botón visible durante generación que aborta el fetch

**Historial mejorado (`ChatHistory.svelte`):**
- Agrupar por fecha: "Hoy", "Ayer", "Esta semana", fechas anteriores
- Búsqueda en historial
- Borrar conversación individual

**Panel de fuentes:**
- Expandible: click en fuente abre excerpt completo en modal
- Mostrar: documento, página, score de relevancia si está disponible

**Selector de colección mejorado:**
- Multi-colección en una misma query (si el backend lo soporta)
- Crossdoc toggle más visible

**Input:**
- Textarea auto-grow (1–6 líneas)
- Contador de caracteres si es largo
- Shortcut: Enter envía, Shift+Enter nueva línea

---

### FASE 3 — Upload Pro

**Objetivo:** Ingestión de documentos enterprise desde la UI.

**Nueva ruta:** `/collections/[name]/upload` (o panel dentro de `/collections/[name]`)

**Componentes:**
- `DropZone.svelte` — drag & drop, múltiples archivos, validación de tipo y tamaño
- `UploadQueue.svelte` — lista de archivos con estado individual
- `UploadQueueItem.svelte` — progress bar, estado, retry/cancel

**Estados por archivo:**
- `pending` — en cola esperando
- `uploading` — subiendo al servidor
- `processing` — servidor está ingestando
- `done` — listo en vectorstore
- `error` — falló con mensaje y opción retry

**API routes nuevas en BFF:**
- `POST /api/collections/[name]/upload` — recibe archivo multipart, lo envía al RAG server
- `GET /api/collections/[name]/documents` — lista documentos en colección
- `DELETE /api/collections/[name]/documents/[docId]` — borrar documento

**Página de colección `/collections/[name]`:**
- Stats: chunks, documentos, último update
- Listado de documentos con: nombre, fecha, tamaño, estado
- Borrar documento individual (confirm modal)
- Botón "Subir documentos" → abre DropZone

**Colecciones index `/collections`:**
- Botón "Nueva colección" → modal con nombre y descripción
- Stats en cada card (chunks, docs, estado)
- Borrar colección (confirm: "Esta acción eliminará X documentos del vectorstore")

---

### FASE 4 — Admin Pro Max

**Objetivo:** Panel de administración completo estilo Discord.

**Nueva estructura de rutas admin:**
```
/admin
  /users          — gestión de usuarios (mejorado)
  /areas          — gestión de áreas (nuevo)
  /permissions    — matrix de permisos (nuevo)
  /rag-config     — configuración del RAG (nuevo)
  /system         — dashboard de sistema (nuevo)
```

**Layout admin (`/admin/+layout.svelte`):**
- Sidebar secundario Discord-style (200px) dentro del main
- Secciones: Usuarios & Áreas / Sistema RAG / Monitoreo

**Usuarios (`/admin/users`):**
- Tabla con búsqueda y filtros: todos / activos / inactivos / por rol / por área
- Editar usuario: modal con campos email, nombre, rol, área, estado activo/inactivo
- Crear usuario: mismo modal en modo creación
- Desactivar/Reactivar usuario
- Regenerar API key desde admin
- Ver historial de accesos del usuario

**Áreas (`/admin/areas`):**
- CRUD completo de áreas
- Asignar usuarios a área
- Ver stats: usuarios en área, consultas del mes

**Permisos (`/admin/permissions`):**
- Matrix visual: filas = áreas/roles, columnas = colecciones + acciones
- Acciones: leer, upload, borrar docs, crossdoc
- Click en celda para toggle permiso
- Guardar todos los cambios juntos (no on-the-fly para evitar estados intermedios)

**RAG Config (`/admin/rag-config`):**
- Sliders con preview en tiempo real de valores
- Parámetros: temperatura, top-k, max tokens, similarity threshold, chunk size, chunk overlap
- Config de NIMs: URLs de embedding model y reranker
- Config de Milvus: collection settings
- Botón "Probar configuración" → query de prueba con los valores actuales
- Guardar / Resetear a defaults

**System Dashboard (`/admin/system`):**
- Cards con métricas: consultas hoy / esta semana, usuarios activos, docs procesados
- Estado de servicios: RAG Server, Milvus, NIM Embed, NIM Reranker (ping con latencia)
- Gráfico simple de consultas por día (últimos 7 días)

---

### FASE 5 — Polish & Extras

**Objetivo:** Detalles que marcan la diferencia entre bueno y profesional.

**Command Palette (⌘K):**
- Navegar entre páginas: Chat, Colecciones, Admin
- Acciones rápidas: Nueva consulta, Subir doc, Nueva colección
- Buscar en historial de chat
- Componente: `CommandPalette.svelte` con fuzzy search

**Animaciones:**
- Transiciones de página: fade-slide suave (150ms)
- Sidebar collapse: smooth width transition
- Toast slide-in desde esquina
- Modal fade + scale in/out

**Responsive:**
- Mobile (< 768px): sidebar se convierte en drawer overlay
- Tablet (768–1024px): sidebar colapsada por defecto

**Dashboard Home (`/`):**
- Redirect a `/chat` si ya tiene sesiones activas
- Si es la primera vez: onboarding con 3 pasos (qué es SDA, cómo hacer una consulta, cómo subir docs)

**Accesibilidad:**
- ARIA labels en todos los controles interactivos
- Focus management en modales (focus trap)
- Navegación por teclado en sidebar y command palette

---

## 4. Arquitectura de componentes (árbol final)

```
src/
├── app.css                          ← design tokens completos
├── lib/
│   ├── components/
│   │   ├── ui/                      ← design system
│   │   │   ├── Button.svelte
│   │   │   ├── Input.svelte
│   │   │   ├── Select.svelte
│   │   │   ├── Badge.svelte
│   │   │   ├── Modal.svelte
│   │   │   ├── Card.svelte
│   │   │   ├── Skeleton.svelte
│   │   │   ├── Toast.svelte
│   │   │   ├── Tooltip.svelte
│   │   │   └── Toggle.svelte
│   │   ├── layout/
│   │   │   ├── Sidebar.svelte       ← rediseño completo
│   │   │   ├── Topbar.svelte
│   │   │   └── CommandPalette.svelte
│   │   ├── chat/
│   │   │   ├── ChatHistory.svelte
│   │   │   ├── ChatMessages.svelte
│   │   │   ├── MessageBubble.svelte
│   │   │   ├── MarkdownRenderer.svelte
│   │   │   ├── SourcesPanel.svelte
│   │   │   └── ChatInput.svelte
│   │   ├── upload/
│   │   │   ├── DropZone.svelte
│   │   │   ├── UploadQueue.svelte
│   │   │   └── UploadQueueItem.svelte
│   │   └── admin/
│   │       ├── AdminSidebar.svelte
│   │       ├── PermissionsMatrix.svelte
│   │       └── RagConfigPanel.svelte
│   ├── stores/
│   │   ├── chat.svelte.ts           ← existente, mejorado
│   │   ├── toast.svelte.ts          ← nuevo
│   │   └── theme.svelte.ts          ← nuevo (dark/light)
│   └── server/
│       ├── gateway.ts               ← existente
│       └── auth.ts                  ← existente
└── routes/
    ├── (auth)/login/
    ├── (app)/
    │   ├── +layout.svelte           ← sidebar + toast outlet
    │   ├── +layout.server.ts
    │   ├── +error.svelte            ← UI completa
    │   ├── chat/
    │   │   ├── +page.svelte         ← empty state
    │   │   └── [id]/+page.svelte    ← chat rediseñado
    │   ├── collections/
    │   │   ├── +page.svelte         ← grid mejorado
    │   │   └── [name]/
    │   │       ├── +page.svelte     ← detalle + docs list
    │   │       └── upload/+page.svelte
    │   ├── admin/
    │   │   ├── +layout.svelte       ← admin sidebar
    │   │   ├── users/+page.svelte
    │   │   ├── areas/+page.svelte
    │   │   ├── permissions/+page.svelte
    │   │   ├── rag-config/+page.svelte
    │   │   └── system/+page.svelte
    │   ├── audit/+page.svelte       ← con filtros + paginación
    │   └── settings/+page.svelte
    └── api/
        └── collections/
            └── [name]/
                ├── documents/+server.ts
                └── upload/+server.ts
```

---

## 5. Patrones técnicos clave

### Error handling
```svelte
<!-- En +layout.svelte -->
<svelte:boundary onerror={(e, reset) => { toastStore.error(e.message); }}>
  {@render children()}
  {#snippet failed(error, reset)}
    <ErrorFallback {error} {reset} />
  {/snippet}
</svelte:boundary>
```

### Toast store (Svelte 5 runes)
```typescript
// src/lib/stores/toast.svelte.ts
class ToastStore {
  toasts = $state<Toast[]>([]);
  success(msg: string, duration = 4000) { ... }
  error(msg: string, duration = 6000) { ... }
  info(msg: string, duration = 4000) { ... }
}
export const toastStore = new ToastStore();
```

### Theme
```typescript
// Usa mode-watcher (ya instalado)
import { ModeWatcher } from 'mode-watcher';
// En +layout.svelte: <ModeWatcher />
// Para cambiar: setMode('dark') | setMode('light') | toggleMode()
```

### Markdown rendering
```typescript
import { marked } from 'marked';
import hljs from 'highlight.js';
marked.setOptions({ highlight: (code, lang) => hljs.highlight(code, { language: lang }).value });
```

### Auto-scroll inteligente
```svelte
// Scroll al fondo solo si el usuario estaba en el fondo
let isAtBottom = true;
// onscroll: actualizar isAtBottom
// cuando llega token: if (isAtBottom) scrollToBottom()
```

---

## 5b. Endpoints de gateway ya existentes (no hay que crear)

El gateway (`saldivia/gateway.py`) ya tiene implementado:
- `PUT /admin/users/{id}` — editar usuario (rol, área, estado)
- `GET/POST/PUT/DELETE /admin/areas` — CRUD completo de áreas
- `GET/POST/DELETE /admin/areas/{id}/collections` — permisos de colección por área
- `POST /admin/users/{id}/reset-key` — regenerar API key
- `GET /admin/audit` — audit log (con filtros en query params)
- `GET /v1/collections/{name}/stats` — stats de colección

El frontend simplemente no los usa aún — es trabajo de UI puro.

## 5c. Endpoints de gateway FALTANTES (hay que crear antes de la feature)

### Para Fase 3 — Upload Pro

Agregar a `saldivia/gateway.py` antes de implementar el frontend:

```python
# Listar documentos de una colección (proxy a Milvus)
GET /v1/collections/{collection_name}/documents
# Borrar documento de vectorstore
DELETE /v1/collections/{collection_name}/documents/{document_id}
# Crear colección vacía en Milvus
POST /v1/collections
# Borrar colección completa
DELETE /v1/collections/{collection_name}
```

Si Milvus/NVIDIA RAG Blueprint no expone estas operaciones directamente, la alternativa es scope reducido: upload funciona pero sin list/delete desde UI (dejar para cuando el backend lo soporte).

### Para Fase 4 — RAG Config

Agregar a `saldivia/gateway.py`:

```python
# Obtener config RAG actual
GET /admin/rag-config
# Actualizar parámetros RAG (temperatura, top-k, etc.)
PUT /admin/rag-config
# Estado de servicios (ping a cada NIM + Milvus)
GET /admin/system/health
# Métricas básicas (consultas hoy, usuarios activos)
GET /admin/system/stats
```

La config se puede almacenar en SQLite (AuthDB) o en variables de entorno writeables — a definir en el plan de implementación.

## 5d. Correcciones al spec original (basadas en code review)

1. **BFF routes**: `api/collections/[name]/documents/+server.ts` maneja `GET` (list), `POST` (implícito via upload), `DELETE` con query param `?doc_id=` — un único archivo por recurso siguiendo el patrón existente del proyecto.

2. **npm deps nuevas (Fase 2)**: Agregar explícitamente al `package.json`:
   ```bash
   npm install marked highlight.js
   npm install -D @types/marked
   ```

3. **svelte:boundary firma correcta** (Svelte 5):
   ```svelte
   <svelte:boundary onerror={(error: Error, reset: () => void) => {
     toastStore.error(error.message);
     console.error('[boundary]', error);
   }}>
     {@render children()}
     {#snippet failed(error, reset)}
       <ErrorFallback {error} {reset} />
     {/snippet}
   </svelte:boundary>
   ```

4. **Admin Discord-style**: Referencia visual aprobada en mockup interactivo (`full-preview.html` tab "Admin Pro Max") — admin sidebar de 200px dentro del main content, separado del sidebar de app.

## 6. Orden de implementación

```
FASE 1 (Fundación)       ← empezar aquí, bloquea todo lo demás
  └─ Design tokens
  └─ Componentes UI base
  └─ Sidebar rediseñada
  └─ Error boundaries
  └─ Dark/light toggle
  └─ Toast system

FASE 2 (Chat Pro)        ← después de Fase 1
  └─ Markdown + hljs
  └─ Auto-scroll
  └─ Historial mejorado
  └─ Acciones por mensaje

FASE 3 (Upload Pro)      ← después de Fase 1
  └─ DropZone
  └─ Queue + progress
  └─ API routes BFF
  └─ Colecciones CRUD

FASE 4 (Admin Pro Max)   ← después de Fases 1 y 3
  └─ Admin layout
  └─ Users edit
  └─ Areas CRUD
  └─ Permissions matrix
  └─ RAG config

FASE 5 (Polish)          ← última, después de todo
  └─ Command palette
  └─ Animaciones
  └─ Responsive
  └─ A11y
```

---

## 7. Testing

- Cada componente UI base con Vitest unit test
- `src/lib/stores/toast.svelte.ts` — tests de estados
- `src/lib/stores/chat.svelte.ts` — tests de streaming states (existentes mejorados)
- Integration tests en `+page.server.ts` críticos (login, chat, upload)

---

*Diseño aprobado por Enzo Saldivia — 2026-03-18*

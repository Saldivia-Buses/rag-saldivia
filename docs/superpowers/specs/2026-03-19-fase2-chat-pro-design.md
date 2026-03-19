# Fase 2 — Chat Pro: Diseño

**Fecha:** 2026-03-19
**Proyecto:** RAG Saldivia — SDA Frontend
**Fase:** 2 de 5

---

## Goal

Mejorar la experiencia de chat al nivel ChatGPT/Perplexity: Markdown rendering completo, historial con búsqueda, panel de fuentes toggleable, stop streaming, y smart auto-scroll.

## Decisiones de diseño

| Aspecto | Decisión | Razonamiento |
|---|---|---|
| Historial de sesiones | Panel inline izquierdo 200px con búsqueda client-side | Siempre visible, sin friction, búsqueda filtra localmente |
| Panel de fuentes | Toggle button en header; panel derecho on/off | Chat a full ancho por defecto; badge muestra count de fuentes |
| Markdown rendering | Completo: headers, listas, código con syntax highlight, tablas | Sistema RAG de documentación técnica — tablas y código son frecuentes |
| Stop streaming | Botón ■ visible durante generación que aborta el fetch | Evita esperar respuestas largas fuera de scope |
| Auto-scroll | Smart: auto-scrollea al fondo salvo que el usuario haya scrolleado; botón "↓ Ir al fondo" | Permite leer mensajes anteriores sin perder contexto del streaming |

## Arquitectura

### Componentes nuevos

```
src/lib/components/chat/
├── HistoryPanel.svelte       — panel izquierdo 200px, sesiones + búsqueda
├── MessageList.svelte        — área scrollable + smart auto-scroll
├── MarkdownRenderer.svelte   — Markdown → HTML seguro
├── SourcesPanel.svelte       — panel derecho toggle, excerpts de fuentes
└── ChatInput.svelte          — textarea auto-grow + enviar/stop
```

### Archivos modificados

| Archivo | Cambio |
|---|---|
| `src/routes/(app)/chat/[id]/+page.svelte` | Reemplazado por coordinador fino (~80 líneas) |
| `src/lib/stores/chat.svelte.ts` | Agrega `abortController` para stop streaming |

### Dependencias nuevas

```json
"marked": "^x.x.x",
"highlight.js": "^x.x.x",
"dompurify": "^x.x.x",
"@types/dompurify": "^x.x.x"
```

## Componentes en detalle

### `HistoryPanel.svelte`

**Props:** `sessions: ChatSession[]`, `currentId: string`, `userId: number`

- Input de búsqueda que filtra `sessions` por `title` via `$derived` (client-side, sin fetch)
- "Nueva consulta" como `<a href="/chat">` al tope
- Cada sesión: `<a href="/chat/{id}" data-sveltekit-preload-data="false">` con título truncado y fecha
- Sesión activa marcada con `border-l-2 border-[var(--accent)]`
- Ancho: 200px fijo

### `MessageList.svelte`

**Props:** `messages: Message[]`, `streaming: boolean`, `streamingContent: string`

- `bind:this={scrollEl}` en el contenedor scrollable
- `$effect` observa `messages.length` y `streamingContent`:
  - Si `scrollEl.scrollHeight - scrollEl.scrollTop - scrollEl.clientHeight < 100` → auto-scroll suave al fondo
  - Si el usuario está más arriba → `atBottom = false` → aparece botón flotante "↓ Ir al fondo"
- Botón "↓ Ir al fondo" en esquina inferior derecha, clickeable, fuerza scroll y restaura `atBottom = true`
- Mensajes `assistant` pasan `content` a `<MarkdownRenderer>`
- Mensajes `user` muestran texto plano (sin Markdown)
- Durante streaming: muestra `streamingContent` en `<MarkdownRenderer>` + cursor parpadeante `▋`

### `MarkdownRenderer.svelte`

**Props:** `content: string`

- `{@html DOMPurify.sanitize(marked.parse(content))}`
- `marked` configurado con `highlight.js` para syntax highlighting en code blocks
- Subset de lenguajes importados: `bash`, `python`, `javascript`, `typescript`, `json`, `yaml`, `sql`
- CSS scoped para:
  - `pre code` — fondo oscuro, monospace, border-radius, padding
  - `table` — bordes con `var(--border)`, header con `var(--bg-surface)`
  - `h1/h2/h3` — `var(--text)`, font-weight semibold, margin vertical
  - `ul/ol` — padding-left, spacing entre items
  - `strong` — `var(--text)` (ya es blanco en dark, no hace falta override)
  - `a` — `var(--accent)`, hover underline
- DOMPurify con `ALLOWED_TAGS` permisivo (permite `table`, `code`, `pre`, `strong`, etc.) pero bloquea `script`, `iframe`, `on*` handlers

### `SourcesPanel.svelte`

**Props:** `sources: Source[]`, `open: boolean`, `ontoggle: () => void`

- Cuando `open && sources.length > 0`: ancho `w-64` (256px) con `transition-[width]`
- Cuando cerrado: `w-0 overflow-hidden`
- Cada source: nombre del documento (bold, truncado), página (`p. N`), excerpt (`line-clamp-4`)
- Primera fuente: `border-l-2 border-[var(--accent)]`, segunda: `border-[var(--accent-hover)]`, resto: `border-[var(--border)]`
- El botón toggle vive en el header de `+page.svelte` (no en este componente)

### `ChatInput.svelte`

**Props:** `streaming: boolean`, `onsubmit: (query: string) => void`, `onstop: () => void`

- `bind:value={input}` en textarea
- Textarea auto-crece: `rows={1}`, CSS `max-height: 120px; overflow-y: auto; resize: none`
- Cuando `streaming=false`: botón Send (icono `Send`), disabled si `input.trim() === ''`
- Cuando `streaming=true`: botón Stop (icono `Square` o `StopCircle`), siempre habilitado
- `onkeydown`: Enter sin Shift → `onsubmit(input.trim())`, Shift+Enter → nueva línea
- Al submit: limpia `input` después de llamar `onsubmit`

### `ChatStore` — cambios

```typescript
abortController = $state<AbortController | null>(null);

startStream() {
    this.abortController = new AbortController();
    this.streaming = true;
    this.streamingContent = '';
    this.sources = [];
}

stopStream() {
    this.abortController?.abort();
    this.finalizeStream();
}
```

El fetch en `+page.svelte` pasa `signal: chat.abortController!.signal`. En el `catch`:
- `err.name === 'AbortError'` → no agrega mensaje de error (stop manual)
- Cualquier otro error → `chat.appendToken('\n[Error: conexión interrumpida]')`

## Layout del `[id]/+page.svelte` resultante

```
<div class="flex h-screen overflow-hidden">
  <HistoryPanel {sessions} {currentId} {userId} />

  <div class="flex-1 flex flex-col border-r border-[var(--border)] min-w-0">
    <!-- Header: collection select + crossdoc + sources toggle badge -->
    <header>
      <select bind:value={selectedCollection}>...</select>
      <label>Crossdoc <input type="checkbox" bind:checked={chat.crossdoc} /></label>
      <button onclick={() => sourcesOpen = !sourcesOpen} disabled={chat.sources.length === 0}>
        Fuentes ({chat.sources.length})
      </button>
    </header>

    <MessageList messages={chat.messages} streaming={chat.streaming}
                 streamingContent={chat.streamingContent} />

    <ChatInput streaming={chat.streaming}
               onsubmit={sendMessage}
               onstop={() => chat.stopStream()} />
  </div>

  <SourcesPanel sources={chat.sources} open={sourcesOpen}
                ontoggle={() => sourcesOpen = !sourcesOpen} />
</div>
```

## Data flow

```
ChatInput.onsubmit(query)
  → chat.addUserMessage(query)
  → chat.startStream()           // nuevo AbortController
  → fetch /api/chat/stream/{id}  // con signal
    → SSE tokens → chat.appendToken()
    → MessageList auto-scrollea
    → SSE citations → chat.setSources()
    → badge Fuentes se actualiza
  → stream done → chat.finalizeStream()
    → mensaje guardado en chat.messages
    → MarkdownRenderer renderiza contenido final

ChatInput.onstop()
  → chat.stopStream()            // abort() + finalizeStream()
  → fetch cancela (AbortError)
  → catch ignora AbortError
  → lo que llegó hasta ese punto queda en messages
```

## Testing

| Test | Qué verifica |
|---|---|
| `MarkdownRenderer` unit | Input con `**bold**`, `# H1`, código Python → HTML correcto; `<script>` sanitizado |
| `ChatStore` unit | `stopStream()` llama `abort()`; estado vuelve a `streaming=false` |
| `HistoryPanel` unit | Array de 5 sesiones, búsqueda "aries" filtra a las que tienen ese título |
| `MessageList` unit | `scrollHeight - scrollTop > 100` → botón "↓ Ir al fondo" visible; click → auto-scroll |
| Integration `+page.server.ts` | Mock gateway → `session`, `history`, `collections` se cargan correctamente |

## Out of scope (Fases siguientes)

- Edición/renombrado de sesiones (Fase 4)
- Export de conversación a PDF/Markdown
- Citas numeradas inline `[1]` en el texto (requiere cambios en el gateway)
- Modo mobile / responsive (Fase 5)

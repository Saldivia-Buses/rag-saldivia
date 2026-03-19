# Fase 5 — Crossdoc Pro

**Fecha:** 2026-03-19
**Proyecto:** RAG Saldivia — SDA Frontend
**Estado:** Diseño aprobado, pendiente de implementación

---

## Objetivo

Port del pipeline crossdoc de React (`patches/frontend/new/`) a SvelteKit 5 / Svelte 5 runes. Pipeline de 4 fases: Decompose → Parallel RAG → Follow-up retries → Synthesis.

Permite al usuario activar un modo "Crossdoc" donde su pregunta se descompone en N sub-queries independientes, se ejecutan en paralelo contra la knowledge base, y los resultados se sintetizan en una respuesta final unificada.

---

## Decisiones de diseño

| Decisión | Elección | Motivo |
|----------|---------|--------|
| Settings UI | Chip `⚡ Crossdoc` en ChatInput → popover compacto | Discoverable, no agrega panel, patrón moderno (Claude.ai, Perplexity) |
| Progress display | Inline en el assistant bubble (4 fases coloreadas) | Transición seamless al texto de síntesis, no contamina historial |
| Decomposition view | Accordion debajo de la respuesta | La respuesta es el protagonista; sub-queries son detalle opcional |
| BFF architecture | `gateway.ts` +2 funciones + 3 endpoints thin | Una sola fuente de verdad para comunicación con gateway |
| Store | `CrossdocStore` singleton con Svelte 5 runes | Consistente con `ChatStore`, `CollectionsStore` |
| Chat integration | Branch en `sendMessage()`, path normal intacto | Fase 2 bloqueada; mínima fricción con código existente |

---

## Arquitectura

### Flujo general

```
ChatInput (chip ⚡ Crossdoc toggle)
  → chat.crossdoc = true/false

ChatPage.sendMessage(question):
  if chat.crossdoc → crossdocStore.run(question, chat)
  else             → streamNormal(question, chat)   // path existente sin tocar

crossdocStore.run(question, chat):
  chat.startStream()                               // bubble vacío aparece
  Phase 1 → POST /api/crossdoc/decompose           → string[]
  Phase 2 → N × POST /api/crossdoc/subquery        → Promise.allSettled()
  Phase 3 → retry fallidos (si followUpRetries)    → más POSTs a /api/crossdoc/subquery
  Phase 4 → POST /api/crossdoc/synthesize          → SSE → chat.appendToken()
  chat.finalizeStream(crossdocResults)             // mensaje guardado con results
```

### Archivos — nuevos y modificados

```
# BFF
src/lib/server/gateway.ts                         MODIFICAR: +gatewayGenerateText, +gatewayGenerateStream
src/routes/api/crossdoc/decompose/+server.ts      NUEVO
src/routes/api/crossdoc/subquery/+server.ts       NUEVO
src/routes/api/crossdoc/synthesize/+server.ts     NUEVO

# Core
src/lib/crossdoc/types.ts                         NUEVO
src/lib/crossdoc/pipeline.ts                      NUEVO
src/lib/stores/crossdoc.svelte.ts                 NUEVO

# Componentes UI
src/lib/components/chat/CrossdocProgress.svelte   NUEVO
src/lib/components/chat/DecompositionView.svelte  NUEVO
src/lib/components/chat/CrossdocSettingsPopover.svelte  NUEVO
src/lib/components/chat/ChatInput.svelte          MODIFICAR: +chip +popover
src/routes/(app)/chat/[id]/+page.svelte           MODIFICAR: +crossdoc branch en sendMessage
src/lib/stores/chat.svelte.ts                     MODIFICAR: finalizeStream acepta crossdocResults?
src/lib/components/chat/MessageList.svelte        MODIFICAR: +DecompositionView por mensaje
```

---

## BFF Layer

### `gateway.ts` — funciones nuevas

```typescript
/**
 * Llamada LLM (con o sin KB) → devuelve texto completo.
 * Para decompose y sub-queries donde no se necesita SSE.
 */
export async function gatewayGenerateText(
  opts: {
    messages: { role: string; content: string }[];
    use_knowledge_base?: boolean;
    collection_names?: string[];
    vdb_top_k?: number;
    reranker_top_k?: number;
    max_tokens?: number;
  },
  signal?: AbortSignal
): Promise<string>

/**
 * Llamada LLM → devuelve Response SSE para proxy directo al browser.
 * Para synthesis donde se necesita streaming en tiempo real.
 */
export async function gatewayGenerateStream(
  opts: {
    messages: { role: string; content: string }[];
    use_knowledge_base?: boolean;
    max_tokens?: number;
  },
  signal?: AbortSignal
): Promise<Response>
```

Ambas usan `GATEWAY_URL` y `SYSTEM_API_KEY` ya definidos en el módulo. Sin duplicar configuración.

### `POST /api/crossdoc/decompose`

```
Auth: locals.user requerido (401 si no)
Body: { question: string, maxSubQueries?: number, collection_names?: string[] }
→ gatewayGenerateText(decompose prompt, use_knowledge_base: false)
→ parsea líneas, aplica Jaccard dedup (threshold 0.65), cap a maxSubQueries
→ { subQueries: string[] }
```

Prompt de decompose (portado desde `useCrossdocDecompose.ts`):
> "You are a search query decomposer... Generate multiple retrieval-focused sub-queries, one per line, no numbering."

### `POST /api/crossdoc/subquery`

```
Auth: locals.user requerido (401 si no)
Body: { query: string, collection_names?: string[], vdbTopK?: number, rerankerTopK?: number }
→ gatewayGenerateText(query, use_knowledge_base: true, enable_reranker: true)
→ { content: string, success: boolean }
  success = content.length > 3 && no empty-result patterns
```

No persiste en historial. No está acoplado a session ID.

### `POST /api/crossdoc/synthesize`

```
Auth: locals.user requerido (401 si no)
Body: { question: string, results: SubResult[] }
→ construye synthesis prompt con los results exitosos
→ gatewayGenerateStream(synthesis prompt, use_knowledge_base: false)
→ SSE proxy directo al browser (igual que /api/chat/stream/[id])
```

Prompt de synthesis (portado desde `useCrossdocStream.ts`):
> "You are a senior engineer writing a comprehensive technical answer. Based on the following retrieval results from multiple sub-queries, write a single unified answer..."

---

## Tipos

```typescript
// src/lib/crossdoc/types.ts

export interface CrossdocOptions {
  maxSubQueries: number;       // default 4 (0 = ilimitado)
  synthesisModel: string;      // '' = usar el LLM por defecto
  followUpRetries: boolean;    // default true
  showDecomposition: boolean;  // default false
  vdbTopK: number;             // default 10
  rerankerTopK: number;        // default 5
}

export interface CrossdocProgress {
  phase: 'decomposing' | 'querying' | 'retrying' | 'synthesizing' | 'done' | 'error';
  subQueries: string[];
  completed: number;
  total: number;
  results: SubResult[];
  error?: string;
}

export interface SubResult {
  query: string;
  content: string;
  success: boolean;
}
```

---

## CrossdocStore (Svelte 5)

```typescript
// src/lib/stores/crossdoc.svelte.ts

const MAX_PARALLEL = 6;
const JACCARD_THRESHOLD = 0.65;

class CrossdocStore {
  progress = $state<CrossdocProgress | null>(null);
  options  = $state<CrossdocOptions>({
    maxSubQueries: 4,
    synthesisModel: '',
    followUpRetries: true,
    showDecomposition: false,
    vdbTopK: 10,
    rerankerTopK: 5,
  });

  private abortCtrl: AbortController | null = null;

  async run(question: string, chat: ChatStore): Promise<void>
  stop(): void   // llama abortCtrl.abort()
  reset(): void  // progress = null
}

export const crossdoc = new CrossdocStore();
```

**`run()` — pseudocódigo:**

```
1. abortCtrl = new AbortController()
   chat.startStream()

2. progress = { phase: 'decomposing', ... }
   subQueries = await POST /api/crossdoc/decompose

3. progress = { phase: 'querying', total: subQueries.length, completed: 0 }
   results = []
   for cada batch de MAX_PARALLEL sub-queries:
     batchResults = await Promise.allSettled(batch.map(q => POST /api/crossdoc/subquery))
     results.push(...batchResults)
     progress.completed += batch.length

4. if followUpRetries && results.some(r => !r.success):
     progress.phase = 'retrying'
     failed = results.filter(r => !r.success).map(r => r.query)
     alternatives = await POST /api/crossdoc/decompose (con prompt follow-up)
     for cada alternative: POST /api/crossdoc/subquery y agregar a results si success

5. progress.phase = 'synthesizing'
   SSE stream desde POST /api/crossdoc/synthesize
   → cada token: chat.appendToken(token)

6. progress = { phase: 'done', results }
   chat.finalizeStream({ crossdocResults: results })
```

**Manejo de abort:** en cada await, si `abortCtrl.signal.aborted`, salir limpio. `chat.finalizeStream()` siempre se llama (con el contenido parcial si fue abortado).

---

## Componentes UI

### `CrossdocProgress.svelte`

Visible en el assistant bubble cuando `chat.streaming && chat.crossdoc`.

- 4 pills: `decomposing` / `querying` / `retrying` / `synthesizing`
- Colores: verde (done) / azul pulsante (active) / gris (pending)
- Progress bar numérica solo en fases `querying` y `retrying`
- Desaparece cuando `phase === 'done'` (el texto de síntesis toma el espacio)

### `DecompositionView.svelte`

Visible debajo de la respuesta si `crossdoc.options.showDecomposition && message.crossdocResults`.

- Accordion colapsado por defecto: "⚡ Sub-queries usadas (N) ▾"
- Al expandir: lista de sub-queries con `✓` (verde) o `✗` (rojo) por resultado
- Solo se muestra en mensajes que tienen `crossdocResults` en metadata

### `CrossdocSettingsPopover.svelte`

Popover compacto que emerge arriba del chip `⚡ Crossdoc` en el ChatInput.

**Contenido:**
- `Max sub-queries`: number input, min=0, max=20 (0 = ilimitado)
- `Synthesis model`: text input, placeholder "(default LLM)"
- `Follow-up retries`: toggle
- `Show decomposition`: toggle

**Comportamiento:**
- Click en chip: abre/cierra popover
- Click fuera (use:clickOutside): cierra
- Todos los valores hacen `bind` directo a `crossdoc.options.*`

### `ChatInput.svelte` — cambios

Agregar junto a los controles existentes:

```svelte
<CrossdocSettingsPopover />
<!-- El chip interno togglea chat.crossdoc -->
<!-- Color azul cuando chat.crossdoc = true, gris cuando false -->
```

---

## Cambios en componentes existentes

### `chat.svelte.ts`

```typescript
// Extender finalizeStream para aceptar crossdocResults opcionales
finalizeStream(opts?: { crossdocResults?: SubResult[] }) {
  if (this.streamingContent) {
    this.messages.push({
      ...
      crossdocResults: opts?.crossdocResults,
    });
  }
  ...
}
```

Agregar `crossdocResults?: SubResult[]` a la interfaz `Message`.

### `MessageList.svelte`

```svelte
{#each chat.messages as msg}
  <!-- render mensaje existente -->
  {#if msg.role === 'assistant' && msg.crossdocResults && crossdoc.options.showDecomposition}
    <DecompositionView results={msg.crossdocResults} />
  {/if}
{/each}
```

---

## Tests

| Archivo | Qué verifica |
|---------|-------------|
| `crossdoc.svelte.test.ts` | `run()` llama endpoints en orden correcto; `stop()` aborta limpiamente; `progress` transiciona por todas las fases; `options` defaults correctos |
| `decompose/decompose.test.ts` | Parsea sub-queries correctamente; Jaccard dedup elimina duplicados; cap a maxSubQueries; 401 sin auth |
| `subquery/subquery.test.ts` | `success: false` con contenido vacío o patrones vacíos; `success: true` con contenido válido; 401 sin auth |
| `synthesize/synthesize.test.ts` | SSE proxy correcto al browser; construye synthesis prompt con results; 401 sin auth |
| `CrossdocProgress.svelte` | Pills con colores correctos por fase; progress bar visible solo en querying/retrying; se oculta en done |
| `CrossdocSettingsPopover.svelte` | Click chip abre popover; click outside cierra; bind correcto a crossdoc.options |
| `DecompositionView.svelte` | Colapsado por defecto; expand muestra lista con ✓/✗; no renderiza sin crossdocResults |

---

## Dependencias con otras fases

- **Fase 2** (bloqueada): `ChatStore` ya tiene `crossdoc`, `abortController`, `appendToken`, `startStream`, `finalizeStream`. Solo se extiende `finalizeStream` para aceptar `crossdocResults?`.
- **Fase 17** (CommandPalette): no afecta esta fase.
- **`patches/frontend/new/`**: fuente de referencia para prompts y lógica de dedup/repetición. No se copia — se reimplementa en SvelteKit/Svelte 5.

---

## Notas de implementación

- **Jaccard dedup** se implementa en `src/lib/crossdoc/pipeline.ts` como función pura (fácil de testear)
- **Repetition detection** del `useCrossdocStream.ts` original (`detectRepetition`) se porta también a `pipeline.ts`
- **MAX_PARALLEL = 6** — no sobrepasar para no saturar el gateway
- **MAX_RESPONSE_CHARS = 15000** — cap por sub-query para no explotar el synthesis prompt
- El `crossdoc` store es un singleton exportado, no se instancia por componente

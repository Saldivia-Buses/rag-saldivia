# 08 — Streaming y RAG

## Vision general

El sistema conecta al usuario con un RAG server NVIDIA (Blueprint v2.5.0) que corre en hardware con GPU. El flujo de streaming es el core de la experiencia — cada token se muestra en tiempo real mientras el LLM genera la respuesta.

---

## Componentes del pipeline

```
ChatInterface (browser)
     |
     | useChat (AI SDK @ai-sdk/react)
     |
     v
POST /api/rag/generate (route handler)
     |
     | extractClaims() — auth + revocacion Redis
     |
     v
ragGenerateStream() (lib/rag/client.ts)
     |
     | fetch("http://localhost:8081/v1/chat/completions", { stream: true })
     | CRITICO: verifica status HTTP ANTES de streamear
     |
     v
RAG Server NVIDIA :8081
     |
     | SSE en formato OpenAI:
     | data: {"choices":[{"delta":{"content":"token"}}]}
     | data: {"choices":[{"delta":{"sources":[...]}}]}
     | data: [DONE]
     |
     v
createRagStreamResponse() (lib/rag/ai-stream.ts)
     |
     | Adapter: NVIDIA SSE → AI SDK Data Stream protocol
     | text-start → text-delta (cada token) → data-sources (citations) → text-end
     |
     v
Response (ReadableStream)
     |
     | useChat consume el stream en el browser
     |
     v
ChatInterface renderiza tokens incrementalmente
```

---

## Archivos del pipeline

### `lib/rag/client.ts` (262 lineas)

**Funcion principal: `ragGenerateStream(body, signal)`**

```typescript
type RagGenerateRequest = {
  messages: Array<{ role: string; content: string }>
  use_knowledge_base?: boolean
  collection_name?: string
  collection_names?: string[]
  temperature?: number
  top_p?: number
  max_tokens?: number
  vdb_top_k?: number
  reranker_top_k?: number
  use_reranker?: boolean
}
```

**Flujo:**
1. Si `MOCK_RAG=true` → redirige a mock (OpenRouter o hardcoded)
2. Crea AbortController con timeout configurable (`RAG_TIMEOUT_MS`, default 120s)
3. Hace fetch al RAG Server con `Accept: text/event-stream`
4. **CRITICO:** Verifica `response.ok` ANTES de retornar el stream
5. Si error → retorna `RagError` con codigo, mensaje, y sugerencia
6. Si OK → retorna `{ stream: ReadableStream, contentType: string }`

**Tipos de error:**
| Codigo | Causa |
|--------|-------|
| `UNAVAILABLE` | ECONNREFUSED — RAG server no esta corriendo |
| `TIMEOUT` | AbortError — excedio el timeout |
| `FORBIDDEN` | HTTP 4xx del upstream |
| `UPSTREAM_ERROR` | HTTP 5xx o sin body |

**`ragFetch(path, options)`:** Para requests no-streaming (colecciones, documentos). Timeout de 10s.

---

### `lib/rag/ai-stream.ts` (117 lineas)

**Funcion principal: `createRagStreamResponse(ragStream)`**

Transforma SSE de NVIDIA al protocolo AI SDK Data Stream.

**Input:** ReadableStream de bytes SSE
**Output:** Response con AI SDK Data Stream

**Logica:**
1. Crea `createUIMessageStream` del AI SDK
2. Lee chunks del stream con `reader.read()`
3. Decodifica bytes a texto, splitea por `\n`
4. Para cada linea:
   - Extrae token de texto con `parseSseLine()` → escribe `text-delta`
   - Busca `delta.sources` → parsea con `CitationSchema.array()` → escribe `data-sources`
5. Al final: escribe `text-end`

**Tipos custom:**
```typescript
type RagDataTypes = {
  sources: { citations: Citation[] }
}
type RagUIMessage = UIMessage<unknown, RagDataTypes>
```

---

### `lib/rag/stream.ts`

Parser de bajo nivel para SSE. `parseSseLine(line)` extrae el token de texto de una linea `data: {...}`.

---

### `lib/rag/artifact-parser.ts`

Parsea artifacts del contenido del LLM:
```html
<artifact type="code" language="python" title="Mi script">
codigo aqui
</artifact>
```

Tipos soportados: code, html, svg, mermaid, table, text.

---

### `lib/rag/collections-cache.ts`

Cache en memoria de colecciones del RAG server. Evita llamadas repetitivas a la API de Milvus.

---

## Mock Mode (desarrollo sin GPU)

### Con OpenRouter (`OPENROUTER_API_KEY` definida)
- Usa OpenRouter como proxy a LLMs en la nube
- Modelo configurable: `OPENROUTER_MODEL` (default: `anthropic/claude-haiku-4-5`)
- Inyecta system prompt para artifacts
- Mismo flujo de streaming SSE (formato OpenAI compatible)

### Sin OpenRouter (fallback hardcoded)
- Respuesta simulada con tokens enviados cada 80ms
- Muestra mensaje explicando que esta en modo mock
- Util para development de UI sin ninguna API key

---

## Deteccion de idioma

`detectLanguageHint(text)` en `client.ts`:
- Detecta caracteres no-latinos (CJK, arabe, cirilico) → "Respond in same language"
- Detecta palabras clave en ingles → "Respond in same language"
- Spanish por defecto (la UI esta en espanol)

---

## Parametros RAG configurables

Desde el admin panel (`/admin/config`), se pueden ajustar:

| Parametro | Default | Descripcion |
|-----------|---------|-------------|
| `temperature` | 0.7 | Creatividad del LLM |
| `top_p` | 0.9 | Nucleus sampling |
| `max_tokens` | 1024 | Longitud maxima de respuesta |
| `vdb_top_k` | 10 | Documentos recuperados de Milvus |
| `reranker_top_k` | 5 | Documentos despues de reranking |
| `use_reranker` | true | Activar/desactivar reranker |

Estos parametros se guardan en la DB y se leen con `loadRagParams()` de `packages/config`.

---

## WebSocket (Messaging)

El sistema de messaging usa WebSocket para funciones en tiempo real:

### `lib/ws/protocol.ts`
Define tipos de mensajes WS: new_message, typing, presence_update, etc.

### `lib/ws/client.ts`
Cliente WS que reconecta automaticamente.

### `lib/ws/presence.ts`
Protocolo de presencia: heartbeat cada N segundos, actualiza `users.lastSeen`.

### `lib/ws/publish.ts`
Publish-subscribe para broadcast a miembros de un canal.

### `lib/ws/sidecar.ts`
Proceso sidecar para el WebSocket server (Next.js no soporta WS nativo en el App Router).

### Hooks relacionados
- `usePresence` — trackea online/offline
- `useTyping` — indicador "escribiendo..."
- `useMessaging` — operaciones CRUD de mensajes

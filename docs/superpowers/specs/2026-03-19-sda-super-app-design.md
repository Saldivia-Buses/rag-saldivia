# SDA Super App — Spec Maestro Fases 3-22

**Fecha:** 2026-03-19
**Proyecto:** RAG Saldivia — SDA Frontend
**Alcance:** Fases 3 a 22 (Fases 1 y 2 están BLOQUEADAS — no tocar)

---

## Contexto

### Stack técnico

| Capa | Tecnología |
|------|-----------|
| Frontend | SvelteKit 5, Svelte 5 runes (`$state`, `$derived`, `$props`, `$effect`) |
| Estilos | Tailwind CSS 4, design tokens CSS vars (`--accent`, `--bg-*`, `--text-*`, etc.) |
| UI components | Fase 1 entregó: Button, Badge, Card, Skeleton, Input, Modal, Toast, Sidebar |
| BFF | SvelteKit server routes (`+server.ts`, `+page.server.ts`) |
| Gateway | FastAPI Python en `saldivia/gateway.py` — 29 endpoints |
| Auth | JWT cookie `sda_session` (8h), roles: admin/area_manager/user |
| Vector DB | Milvus — colecciones multitenant |
| Streaming | SSE via httpx streaming proxy, AbortController en frontend |

### Fases anteriores (intocables)

- **Fase 1** ✅ Fundación del diseño: design tokens, componentes UI, Sidebar, dark/light mode
- **Fase 2** 📐 Chat Pro: HistoryPanel, MessageList, MarkdownRenderer, SourcesPanel, ChatInput, stop streaming, auto-scroll

### Convenciones de código

```svelte
<!-- Svelte 5 runes — siempre así -->
let { data } = $props();
let loading = $state(false);
let filtered = $derived(items.filter(i => i.active));
$effect(() => { /* side effect */ });
```

```typescript
// BFF server functions — usar el cliente tipado
import { gatewayFoo } from '$lib/server/gateway';
```

---

## Fase 3 — CI/CD Auto-deploy

**Goal:** Pipeline de deploy automático via GitHub Actions que detecta pushes a `main`, despliega en Brev, hace health check, notifica el resultado, y puede hacer rollback automático si falla.

### Arquitectura

```
Push a main (GitHub)
  → GitHub Actions workflow (.github/workflows/deploy.yml)
    → SSH a nvidia-enterprise-rag-deb106 con deploy key
      → git pull origin main
      → make deploy PROFILE=brev-2gpu
      → Health check: GET /health × 5 intentos (30s timeout)
        → OK   → notifica éxito (webhook al SSE de Fase 13 cuando esté disponible)
        → FAIL → git revert HEAD + make deploy (rollback) + notifica fallo
```

### Archivos

```
.github/
└── workflows/
    ├── deploy.yml           — workflow principal: deploy a Brev en push a main
    └── test.yml             — (futuro) correr tests en PR antes de merge

scripts/
└── health_check.sh          — verifica que todos los servicios respondan (gateway, RAG, Milvus)

Makefile                     — agregar targets: rollback, health (si no existen)
```

### `deploy.yml`

```yaml
name: Deploy to Brev

on:
  push:
    branches: [main]
  workflow_dispatch:           # permite deploy manual desde GitHub UI

jobs:
  deploy:
    runs-on: ubuntu-latest
    timeout-minutes: 15

    steps:
      - name: Setup SSH
        uses: webfactory/ssh-agent@v0.9.0
        with:
          ssh-private-key: ${{ secrets.BREV_DEPLOY_KEY }}

      - name: Add Brev to known_hosts
        run: |
          ssh-keyscan -H ${{ secrets.BREV_HOST }} >> ~/.ssh/known_hosts

      - name: Deploy
        env:
          BREV_HOST: ${{ secrets.BREV_HOST }}
          BREV_USER: ${{ secrets.BREV_USER }}
        run: |
          ssh $BREV_USER@$BREV_HOST "
            set -e
            cd ~/rag-saldivia
            git fetch origin main
            git checkout main
            git pull origin main
            # Inyectar git SHA para Fase 20 (FeedbackWidget)
            # \$ = evaluar en Brev, no en el runner de GitHub Actions
            export VITE_GIT_SHA=\$(git rev-parse --short HEAD)
            make deploy PROFILE=brev-2gpu
          "

      - name: Health check
        env:
          BREV_HOST: ${{ secrets.BREV_HOST }}
          BREV_USER: ${{ secrets.BREV_USER }}
        run: |
          ssh $BREV_USER@$BREV_HOST "bash ~/rag-saldivia/scripts/health_check.sh"

      - name: Rollback on failure
        if: failure()
        env:
          BREV_HOST: ${{ secrets.BREV_HOST }}
          BREV_USER: ${{ secrets.BREV_USER }}
        run: |
          ssh $BREV_USER@$BREV_HOST "
            cd ~/rag-saldivia
            git revert HEAD --no-edit
            make deploy PROFILE=brev-2gpu
          "
          # NOTA: NO hacemos git push del revert — el revert queda local en Brev.
          # --no-edit genera mensaje automático "Revert '<commit>'"
          # No necesitamos [skip ci] porque este commit nunca llega a GitHub.

      - name: Notify result
        if: always()
        run: |
          STATUS=${{ job.status }}
          SHA=$(echo ${{ github.sha }} | cut -c1-8)
          echo "Deploy $STATUS — commit $SHA"
          # En Fase 13 se puede conectar aquí con webhook al notification_manager
```

### `scripts/health_check.sh`

```bash
#!/bin/bash
set -e

MAX_RETRIES=5
SLEEP=10

echo "Checking gateway..."
for i in $(seq 1 $MAX_RETRIES); do
  if curl -sf http://localhost:9000/health > /dev/null 2>&1; then
    echo "✓ Gateway OK"
    break
  fi
  echo "Attempt $i/$MAX_RETRIES — waiting ${SLEEP}s..."
  sleep $SLEEP
  if [ $i -eq $MAX_RETRIES ]; then
    echo "✗ Gateway health check failed"
    exit 1
  fi
done

echo "Checking RAG server..."
curl -sf http://localhost:8081/health > /dev/null || { echo "✗ RAG server down"; exit 1; }
echo "✓ RAG server OK"

echo "Checking frontend..."
curl -sf http://localhost:3000/ > /dev/null || { echo "✗ Frontend down"; exit 1; }
echo "✓ Frontend OK"

echo "All services healthy ✓"
```

### Secrets de GitHub requeridos

| Secret | Valor |
|--------|-------|
| `BREV_DEPLOY_KEY` | SSH private key (sin passphrase) con acceso a la instancia |
| `BREV_HOST` | IP o hostname de nvidia-enterprise-rag-deb106 |
| `BREV_USER` | Usuario SSH (ubuntu o el configurado en Brev) |

**Setup:** `ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/brev_deploy` → private key a GitHub Secrets → public key a `~/.ssh/authorized_keys` en Brev.

### Notificación en Slack/Discord (opcional, zero-config)

Si el repo tiene un webhook configurado, se puede agregar un step final que postea el resultado en un canal. Out-of-scope por ahora — la Fase 13 (notificaciones SSE) cubre la notificación in-app.

### Tests (Fase 3)

| Test | Qué verifica |
|------|-------------|
| `health_check.sh` local | Con servicios corriendo → exit 0; con gateway down → exit 1 |
| Workflow dispatch manual | Trigger manual desde GitHub UI → deploy exitoso |
| Push a branch non-main | No dispara el workflow (solo `main`) |

---

## Fase 4 — Colecciones Pro + Upload Básico

**Goal:** UI completa para gestión de colecciones (CRUD) y primera versión de upload de documentos.

### Funcionalidades

1. **Página `/collections`** — grid mejorado con stats, estado de salud, acción de crear
2. **Página `/collections/[name]`** — detalle: stats, documentos listados, delete collection
3. **Crear colección** — modal con Input nombre + schema selector
4. **Upload básico** — drag & drop o file picker, sube a `/api/upload`, feedback de resultado
5. **BFF** — proxy de `POST /v1/documents` (ingesta) + `DELETE /v1/collections/{name}` + `POST /v1/collections`

### Archivos

```
src/routes/(app)/collections/
├── +page.svelte             — grid de colecciones (ya existe, expandir)
├── +page.server.ts          — load: list + stats (ya existe, expandir)
├── [name]/
│   ├── +page.svelte         — detalle de colección
│   └── +page.server.ts      — load: stats, actions: delete
└── _components/
    ├── CollectionCard.svelte — card con stats + health badge
    ├── CreateModal.svelte    — modal crear colección
    └── DeleteModal.svelte    — confirm delete

src/routes/(app)/upload/
├── +page.svelte             — reemplazar stub actual
└── +page.server.ts          — action: upload

src/routes/api/collections/
├── +server.ts               — POST (crear)
└── [name]/
    └── +server.ts           — DELETE

src/routes/api/upload/
└── +server.ts               — POST multipart → gateway /v1/documents

src/lib/server/gateway.ts    — agregar: gatewayCreateCollection, gatewayDeleteCollection, gatewayIngestDocument

src/lib/stores/collections.svelte.ts  — CollectionsStore (singleton `collectionsStore`):
                                          list: Collection[], load(), create(), delete()
                                          Requerido por Fase 17 (CommandPalette) para filtrar colecciones
```

### Componente `CollectionCard.svelte`

```svelte
<script lang="ts">
  let { name, stats, href } = $props<{
    name: string;
    stats: { entity_count: number; index_type: string; has_sparse: boolean } | null;
    href: string;
  }>();
</script>

<a {href} class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] p-4
                  hover:border-[var(--accent)] transition-colors block group">
  <div class="flex items-start justify-between mb-3">
    <span class="text-sm font-semibold text-[var(--text)] truncate">{name}</span>
    {#if stats?.has_sparse}
      <Badge variant="blue">Sparse</Badge>
    {/if}
  </div>
  {#if stats}
    <div class="text-2xl font-bold text-[var(--text)] mb-1">
      {stats.entity_count.toLocaleString()}
    </div>
    <div class="text-xs text-[var(--text-faint)]">entidades · {stats.index_type}</div>
  {:else}
    <Skeleton class="h-8 w-24 mb-1" />
  {/if}
</a>
```

### Upload page

- Zona drag & drop con `dragenter/dragleave/drop` handlers nativos
- Acepta: `.pdf`, `.txt`, `.md`, `.docx` — validado client-side
- Selector de colección destino (`<select>` con colecciones disponibles)
- Progress bar falso (simulado) durante upload — la ingesta real es async
- Toast de éxito/error

### BFF upload endpoint

```typescript
// src/routes/api/upload/+server.ts
export async function POST({ request, cookies }) {
  const session = await getSession(cookies);
  const formData = await request.formData();
  const file = formData.get('file') as File;
  const collection = formData.get('collection') as string;

  // Forward multipart al gateway
  const gw = new FormData();
  gw.append('file', file);
  gw.append('collection_name', collection);

  const resp = await fetch(`${GATEWAY_URL}/v1/documents`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${SYSTEM_API_KEY}`, 'X-User-Id': String(session.userId) },
    body: gw,
  });
  return json(await resp.json(), { status: resp.status });
}
```

### Tests (Fase 4)

| Test | Qué verifica |
|------|-------------|
| `CollectionCard` unit | Stats undefined → Skeleton visible; href correcto |
| `CreateModal` unit | Submit vacío → error; nombre válido → emit evento |
| BFF `POST /api/collections` | Mock gateway → 201 con nombre |
| BFF `DELETE /api/collections/[name]` | Mock gateway → 204 |
| BFF `POST /api/upload` | Archivo .pdf + collection → forward correcto al gateway |

---

## Fase 5 — Crossdoc Pro

**Goal:** Port del pipeline crossdoc de React (`patches/frontend/new/`) a SvelteKit 5 / Svelte 5 runes. 4 fases: Decompose → Parallel RAG → Follow-up retries → Synthesis.

**Dependencia con Fase 2:** Esta fase utiliza del `ChatStore` (Fase 2):
- `chat.finalizeStream(content)` — para finalizar el stream con la síntesis
- `chat.crossdoc: boolean` — flag que activa el pipeline
- `chat.abortController` — para stop streaming durante crossdoc
- `chat.appendToken(token)` — para streaming de la síntesis en tiempo real

Antes de implementar Fase 5, verificar que Fase 2 fue entregada y que el `ChatStore` expone estas APIs.

### Contexto técnico

En `patches/frontend/new/` existen hooks React que implementan el pipeline completo:
- `useCrossdocDecompose.ts` (cc=32): LLM → N sub-queries
- `useCrossdocStream.ts` (cc=27): 4-phase pipeline con AbortController, onProgress
- `SaldiviaSection.tsx`: Settings UI para crossdoc params

Estos hooks deben reimplementarse como stores/funciones Svelte 5. No copiar React — reescribir.

### Archivos

```
src/lib/crossdoc/
├── decompose.ts             — función decompose(question, opts) → string[]
├── pipeline.ts              — orchestrator: 4 fases con progress callbacks
├── synthesis.ts             — función synthesize(results[], opts) → SSE stream
└── types.ts                 — interfaces: CrossdocOptions, CrossdocProgress, SubResult

src/lib/stores/crossdoc.svelte.ts
                             — store Svelte 5: state + actions para el pipeline

src/routes/api/crossdoc/
├── decompose/+server.ts     — POST: query → sub-queries via LLM
└── synthesize/+server.ts    — POST: results[] → síntesis SSE

src/lib/components/chat/
├── CrossdocProgress.svelte  — barra de progreso 4 fases durante streaming
└── DecompositionView.svelte — muestra sub-queries expandibles (toggle)
```

### `CrossdocOptions` y tipos

```typescript
// src/lib/crossdoc/types.ts
export interface CrossdocOptions {
  maxSubQueries: number;      // default 4
  synthesisModel: string;     // model ID para síntesis
  followUpRetries: number;    // default 1
  showDecomposition: boolean; // mostrar sub-queries en UI
  vdbTopK: number;            // default 10
  rerankerTopK: number;       // default 5
}

export interface CrossdocProgress {
  phase: 'decomposing' | 'querying' | 'retrying' | 'synthesizing' | 'done';
  subQueries: string[];
  completed: number;
  total: number;
  results: SubResult[];
}

export interface SubResult {
  query: string;
  success: boolean;
  sources: Source[];
  content: string;
}
```

### `CrossdocStore` (Svelte 5)

```typescript
// src/lib/stores/crossdoc.svelte.ts
class CrossdocStore {
  progress = $state<CrossdocProgress | null>(null);
  options = $state<CrossdocOptions>({
    maxSubQueries: 4,
    synthesisModel: 'default',
    followUpRetries: 1,
    showDecomposition: false,
    vdbTopK: 10,
    rerankerTopK: 5,
  });
  abortController = $state<AbortController | null>(null);

  async run(question: string, sessionId: string): Promise<void> { ... }
  abort() { this.abortController?.abort(); }
  reset() { this.progress = null; }
}
export const crossdoc = new CrossdocStore();
```

### Pipeline (4 fases)

```
1. DECOMPOSE (BFF /api/crossdoc/decompose)
   → LLM prompt: "Descomponé la pregunta en N sub-preguntas independientes"
   → retorna string[]

2. PARALLEL RAG (BFF /api/chat/stream/{id} — N requests paralelos)
   → Promise.allSettled(subQueries.map(q => streamQuery(q)))
   → SubResult[] con success/failure por cada uno

3. FOLLOW-UP RETRIES (solo para failed sub-queries)
   → LLM genera query alternativa para los que fallaron
   → retry una vez

4. SYNTHESIS (BFF /api/crossdoc/synthesize)
   → Envía todos los results[] al LLM con SYNTH prompt
   → SSE stream de la respuesta final
   → Actualiza chat.streamingContent mientras viene
```

### `CrossdocProgress.svelte`

```svelte
<!-- Muestra barra de progreso durante el pipeline -->
{#if crossdoc.progress && crossdoc.progress.phase !== 'done'}
  <div class="px-4 py-2 border-t border-[var(--border)] bg-[var(--bg-surface)]">
    <div class="text-xs text-[var(--text-muted)] mb-1">
      {phaseLabel[crossdoc.progress.phase]} —
      {crossdoc.progress.completed}/{crossdoc.progress.total}
    </div>
    <div class="h-1 bg-[var(--bg-hover)] rounded-full overflow-hidden">
      <div class="h-full bg-[var(--accent)] transition-all"
           style="width: {(crossdoc.progress.completed/crossdoc.progress.total)*100}%"></div>
    </div>
  </div>
{/if}
```

### Integración con Chat

- En `+page.svelte` del chat: si `chat.crossdoc === true` → usar `crossdoc.run()` en lugar del fetch directo
- El toggle crossdoc en el header activa/desactiva `chat.crossdoc`
- `crossdoc.run()` al finalizar llama `chat.finalizeStream()` con la síntesis

### Tests (Fase 5)

| Test | Qué verifica |
|------|-------------|
| `decompose.ts` unit | 1 pregunta → array de N strings; LLM error → throws |
| `pipeline.ts` unit | Mock 3 sub-queries, 1 falla → retry, síntesis con 3 results |
| `CrossdocStore` unit | `run()` actualiza `progress` fase a fase; `abort()` para |
| `CrossdocProgress` unit | phase=querying, completed=2, total=4 → barra 50% |
| BFF `/api/crossdoc/decompose` | Mock gateway → array de sub-queries |

---

## Fase 6 — Upload Inteligente

**Goal:** Sistema de ingesta con clasificación por tiers, progress real-time, polling adaptativo, PDFs grandes auto-splitteados. Basado en `scripts/smart_ingest.py`.

### Tier system

| Tier | Páginas | Poll interval | Restart threshold |
|------|---------|---------------|-----------------|
| tiny | ≤20 | 2s | 30s |
| small | ≤80 | 3s | 45s |
| medium | ≤250 | 5s | 90s |
| large | >250 | 10s | 120s |

**Deadlock threshold:** 45s sin progreso → marcar como stalled
**PDF split:** >200 páginas → split automático en chunks de 200, ingesta secuencial — **hecho server-side por el gateway** (Python: `pypdf` o equivalente). El frontend no necesita parsear el PDF; el gateway devuelve el tier y page count en la respuesta del upload.

**Flujo de datos:**
1. Frontend sube el archivo completo vía `POST /api/upload`
2. Gateway extrae page count, clasifica tier, splitea si >200 páginas
3. Gateway retorna `{ jobId, tier, pageCount, chunks: N }` → frontend crea `IngestionJob` con esos datos
4. Frontend inicia `IngestPoller` con el `jobId` y poll interval del tier

### Archivos

```
src/lib/ingestion/
├── tier.ts                  — classifyTier(pageCount) → Tier (client-side, para mostrar badge estimado pre-upload)
├── poller.ts                — IngestPoller class: poll loop con adaptive interval
└── types.ts                 — IngestionJob, IngestionStatus, Tier

src/lib/stores/ingestion.svelte.ts
                             — store: jobs[], active job tracking

src/routes/(app)/upload/
├── +page.svelte             — reemplazar versión Fase 3 con UI pro
└── +page.server.ts          — action: initiate ingestion

src/routes/api/ingestion/
├── [jobId]/status/+server.ts — GET: proxy al gateway job status
└── resume/+server.ts         — POST: resume stopped job

src/lib/components/upload/
├── DropZone.svelte           — drag & drop zone mejorada (Fase 3 base)
├── IngestionQueue.svelte     — lista de jobs activos/completados
├── JobCard.svelte            — card individual: tier badge, progress, ETA
└── TierBadge.svelte          — badge colored: tiny/small/medium/large
```

### `IngestPoller`

```typescript
// src/lib/ingestion/poller.ts
export class IngestPoller {
  private jobId: string;
  private tier: Tier;
  private lastProgress: number = 0;
  private lastProgressAt: number = Date.now();

  async poll(onUpdate: (status: IngestionStatus) => void): Promise<void> {
    const interval = TIER_CONFIG[this.tier].pollInterval;
    while (true) {
      // fetchStatus llama a GET /api/ingest/{jobId}/status (BFF → gateway GET /v1/ingest/{jobId}/status)
      const status = await fetchStatus(this.jobId);
      onUpdate(status);

      if (status.state === 'completed' || status.state === 'failed') break;

      // Deadlock detection
      if (status.progress === this.lastProgress) {
        if (Date.now() - this.lastProgressAt > DEADLOCK_THRESHOLD) {
          onUpdate({ ...status, state: 'stalled' });
          break;
        }
      } else {
        this.lastProgress = status.progress;
        this.lastProgressAt = Date.now();
      }

      await sleep(interval * 1000);
    }
  }
}
```

### `IngestionQueue.svelte`

- Muestra jobs activos con: nombre de archivo, tier badge, progress bar, estado (queued/running/done/failed/stalled)
- Jobs completados colapsan con checkmark verde
- Jobs fallados/stalled muestran botón "Reintentar"
- Orden: activos primero, luego completados (últimos 10), luego fallados

### `JobCard.svelte`

```svelte
<!-- ETA calculado basado en progress rate -->
<div class="flex items-center gap-3 py-3 border-b border-[var(--border)] last:border-0">
  <TierBadge tier={job.tier} />
  <div class="flex-1 min-w-0">
    <div class="text-sm font-medium text-[var(--text)] truncate">{job.filename}</div>
    <div class="text-xs text-[var(--text-muted)]">{job.pageCount} páginas</div>
    <div class="mt-1.5 h-1.5 bg-[var(--bg-hover)] rounded-full overflow-hidden">
      <div class="h-full bg-[var(--accent)] transition-all"
           style="width: {job.progress}%"></div>
    </div>
  </div>
  <div class="text-xs text-[var(--text-faint)] text-right">
    <div>{job.progress}%</div>
    {#if job.eta}<div>~{job.eta}s</div>{/if}
  </div>
</div>
```

### Tests (Fase 6)

| Test | Qué verifica |
|------|-------------|
| `tier.ts` unit | 15 págs → tiny; 100 págs → medium; 300 págs → large |
| `poller.ts` unit | progreso estancado >45s → state='stalled' |
| `IngestStore` unit | addJob → jobs.length++; updateJob → reemplaza por jobId |
| `JobCard` unit | progress=75 → barra al 75%; state=stalled → botón reintentar |
| BFF `/api/ingestion/[id]/status` | Mock gateway → IngestionStatus |

---

## Fase 7 — Chat Sesiones Pro

**Goal:** Gestión completa del historial de sesiones: renombrar, eliminar, exportar, pin, feedback por mensaje, y follow-ups sugeridos.

### Funcionalidades

1. **Renombrar sesión** — doble-click en título del HistoryPanel → inline edit
2. **Eliminar sesión** — botón en HistoryPanel con confirm modal
3. **Exportar conversación** — botones: Markdown, JSON (descarga local sin backend)
4. **Pinear sesión** — sesiones pinadas aparecen al tope del panel (localStorage)
5. **Feedback por mensaje** — thumbs up/down por respuesta del assistant; persiste en gateway
6. **Follow-ups sugeridos** — al final de cada respuesta: 2-3 preguntas relacionadas como chips clickeables
7. **BFF** — `DELETE /api/sessions/[id]`, `PATCH /api/sessions/[id]` (rename), `POST /api/messages/[id]/feedback`

### Archivos

```
src/lib/components/chat/
├── HistoryPanel.svelte      — agregar: renombrar (inline edit), delete, pin (Fase 2 base)
├── MessageList.svelte       — agregar: feedback buttons, follow-up chips (Fase 2 base)
├── FollowUpChips.svelte     — chips de preguntas sugeridas
└── FeedbackButtons.svelte   — thumbs up/down con estado local + persist

src/routes/api/sessions/
├── [id]/
│   ├── +server.ts           — PATCH (rename), DELETE
│   └── feedback/+server.ts  — POST: messageId + rating

src/lib/stores/chat.svelte.ts
                             — agregar: pinnedSessions (localStorage), renameSession, deleteSession
```

### Renombrar inline

```svelte
<!-- En HistoryPanel.svelte, sesión activa -->
{#if editingId === session.id}
  <input
    bind:value={editTitle}
    onblur={() => commitRename(session.id)}
    onkeydown={(e) => { if (e.key === 'Enter') commitRename(session.id); if (e.key === 'Escape') editingId = null; }}
    class="flex-1 bg-transparent border-b border-[var(--accent)] text-sm outline-none"
    autofocus
  />
{:else}
  <button
    ondblclick={() => { editingId = session.id; editTitle = session.title; }}
    class="flex-1 text-left text-sm truncate"
  >{session.title}</button>
{/if}
```

### Export functions (client-side)

```typescript
// exportar como Markdown
function exportMarkdown(session: ChatSession, messages: Message[]) {
  const lines = [`# ${session.title}`, `_${new Date(session.created_at).toLocaleDateString()}_`, ''];
  for (const m of messages) {
    lines.push(`**${m.role === 'user' ? 'Vos' : 'SDA'}:** ${m.content}`, '');
  }
  downloadText(lines.join('\n'), `${session.title}.md`);
}
```

### `FollowUpChips.svelte`

- Aparece al final de cada mensaje del assistant (después de finalizar stream)
- **Mecanismo: client-side puro, sin backend.** Las sugerencias se generan a partir del último mensaje del assistant usando plantillas heurísticas:
  - Extraer sustantivos/frases clave del primer párrafo de la respuesta (split por `. `, tomar primeras 3-4 frases)
  - Aplicar 2-3 plantillas: `"¿Podés ampliar sobre {tema}?"`, `"¿Cuáles son los riesgos de {tema}?"`, `"¿Hay más documentos sobre {tema}?"`
  - Dedupe con la pregunta original (no repetir lo ya preguntado)
- Props: `suggestions: string[]`, `onselect: (q: string) => void`
- El `MessageList.svelte` genera `suggestions` llamando a `generateFollowUps(lastMessage.content)` después de `finalizeStream()` — función pura en `src/lib/chat/followups.ts`
- **No requiere cambios en gateway ni backend**

### Feedback por mensaje

- 2 botones pequeños (thumbs up / thumbs down) en cada mensaje assistant
- Estado: `rating = $state<'up' | 'down' | null>(null)`
- Al hacer click → `POST /api/messages/[msgId]/feedback` con `{ rating: 'up' | 'down' }`
- Persiste en gateway (endpoint nuevo: `POST /v1/messages/{id}/feedback`)
- Visual: botón activo → color `var(--accent)` para up, `var(--danger)` para down

**Schema de DB requerido** — agregar tabla en `saldivia/auth/database.py`:
```sql
CREATE TABLE IF NOT EXISTS message_feedback (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id  TEXT    NOT NULL,  -- ID del mensaje en chat_messages
    user_id     INTEGER NOT NULL,
    rating      TEXT    NOT NULL CHECK(rating IN ('up', 'down')),
    created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
    UNIQUE(message_id, user_id)    -- un voto por usuario por mensaje
);
```
El `message_id` corresponde al campo `id` de `ChatMessage` en `auth/models.py`. El gateway hace `INSERT OR REPLACE` para manejar cambio de voto.

### Tests (Fase 7)

| Test | Qué verifica |
|------|-------------|
| HistoryPanel renombrar | dblclick → input visible; blur → emite rename event |
| HistoryPanel delete | click delete → modal; confirm → emite delete event |
| exportMarkdown | 3 mensajes → string MD correcto |
| FeedbackButtons | click up → rating='up' → botón styled; click again → toggle null |
| FollowUpChips | 3 suggestions → 3 chips; click → onselect llamado con texto |

---

## Fase 8 — Admin Usuarios Pro

**Goal:** Panel de administración completo para usuarios: editar nombre/rol/áreas, resetear API key, desactivar/reactivar, paginar, y vista detalle.

### Funcionalidades actuales (stub existente)

La ruta `/(app)/admin/users/` existe con table básica de usuarios. Agregar:

1. **Modal editar usuario** — nombre, rol, área asignada
2. **Reset API key** — botón por usuario → gateway `POST /admin/users/{id}/reset-key`
3. **Desactivar/reactivar** — toggle activo/inactivo
4. **Paginación** — 20 usuarios por página con cursor
5. **Búsqueda** — filter client-side por nombre/email
6. **Vista detalle** — `/admin/users/[id]` con: info, API key generada, historial de sesiones, áreas

### Archivos

```
src/routes/(app)/admin/users/
├── +page.svelte             — tabla pro + search + paginación
├── +page.server.ts          — load users paginados; actions: create, deactivate
└── [id]/
    ├── +page.svelte         — detalle de usuario
    └── +page.server.ts      — load user detail; actions: update, resetKey, deactivate

src/lib/components/admin/
├── UserTable.svelte         — table con search, sort, pagination
├── UserRow.svelte           — fila: avatar, nombre, rol badge, acciones
├── UserEditModal.svelte     — modal editar: nombre, rol, área
├── UserDetailCard.svelte    — card info completa en detalle page
└── ResetKeyModal.svelte     — confirm + muestra nueva key (one-time)

src/lib/server/gateway.ts    — agregar: gatewayUpdateUser, gatewayResetUserKey,
                               gatewayDeactivateUser, gatewayGetUserDetail
```

### `UserTable.svelte`

```svelte
<script lang="ts">
  let { users } = $props<{ users: User[] }>();
  let search = $state('');
  let filtered = $derived(
    users.filter(u =>
      u.name.toLowerCase().includes(search.toLowerCase()) ||
      u.email.toLowerCase().includes(search.toLowerCase())
    )
  );
</script>

<div class="mb-3">
  <Input bind:value={search} placeholder="Buscar usuarios..." />
</div>
<table class="w-full text-sm">
  <thead>...</thead>
  <tbody>
    {#each filtered as user}
      <UserRow {user} oneditclick={...} onresetkey={...} ondeactivate={...} />
    {/each}
  </tbody>
</table>
```

### Modal reset key

- Warning: "Esto invalidará la key actual del usuario"
- Confirm → `POST /admin/users/{id}/reset-key`
- Respuesta incluye nueva key en plain text
- Mostrar en modal con Copy button y warning "Solo visible ahora"
- Patrón idéntico a settings/+page.svelte existente

### Roles RBAC

```
admin       → puede editar todo, desactivar, resetear keys
area_manager → puede ver su área, no puede editar otros usuarios
user        → sin acceso al panel admin
```

Gateway valida el rol — el frontend solo oculta/deshabilita botones según `data.user.role`.

### Tests (Fase 8)

| Test | Qué verifica |
|------|-------------|
| UserTable search | "juan" filtra correctamente; "" muestra todos |
| UserEditModal | submit sin nombre → error; rol válido → emite update |
| ResetKeyModal | confirm → muestra nueva key; copy button → clipboard |
| BFF `PATCH /admin/users/[id]` | Mock gateway → 200 con user actualizado |
| BFF `POST /admin/users/[id]/reset-key` | Mock → nueva key en response |

---

## Fase 9 — Admin Áreas + Permisos

**Goal:** CRUD completo de áreas organizacionales y matriz visual de permisos por colección.

### Modelo de datos

```
Area {
  id: number
  name: string
  description: string
  users: User[]
  collections: AreaCollection[]  -- permisos por colección
}

AreaCollection {
  area_id: number
  collection_name: string
  permission: 'READ' | 'WRITE' | 'ADMIN'
}
```

### Funcionalidades

1. **`/admin/areas`** — lista de áreas con conteo de usuarios y colecciones
2. **Crear área** — modal: nombre + descripción
3. **`/admin/areas/[id]`** — detalle: lista de usuarios del área, matriz de permisos
4. **Matriz de permisos** — grid visual: filas = colecciones, columnas = READ/WRITE/ADMIN; toggles
5. **Agregar/remover usuarios** — select users para asignar al área
6. **BFF** — `POST/PATCH/DELETE /api/areas/*`, `POST/DELETE /api/areas/[id]/collections/[name]`

### Archivos

```
src/routes/(app)/admin/areas/
├── +page.svelte             — lista áreas
├── +page.server.ts          — load; actions: create, delete
└── [id]/
    ├── +page.svelte         — detalle: users + permisos matrix
    └── +page.server.ts      — load detail; actions: update, grantCollection, revokeCollection,
                               addUser, removeUser

src/lib/components/admin/
├── AreaCard.svelte           — card: nombre, # usuarios, # colecciones, acciones
├── AreaCreateModal.svelte    — modal crear área
├── PermissionsMatrix.svelte  — tabla visual permisos (núcleo de esta fase)
└── UserSelector.svelte       — multiselect para asignar usuarios a área

src/lib/server/gateway.ts    — agregar: gatewayCreateArea, gatewayUpdateArea,
                               gatewayDeleteArea, gatewayAddUserToArea,
                               gatewayRemoveUserFromArea
```

### `PermissionsMatrix.svelte`

```svelte
<!-- Columnas: colección | READ | WRITE | ADMIN -->
<table class="w-full text-sm">
  <thead>
    <tr>
      <th class="text-left py-2 text-xs text-[var(--text-faint)] font-medium">Colección</th>
      {#each ['READ', 'WRITE', 'ADMIN'] as perm}
        <th class="text-center py-2 text-xs text-[var(--text-faint)] font-medium">{perm}</th>
      {/each}
    </tr>
  </thead>
  <tbody>
    {#each collections as col}
      <tr class="border-t border-[var(--border)]">
        <td class="py-2 font-medium text-[var(--text)]">{col}</td>
        {#each ['READ', 'WRITE', 'ADMIN'] as perm}
          <td class="text-center py-2">
            <input type="checkbox"
                   checked={hasPermission(areaId, col, perm)}
                   onchange={() => togglePermission(areaId, col, perm)}
                   class="accent-[var(--accent)] w-4 h-4 cursor-pointer" />
          </td>
        {/each}
      </tr>
    {/each}
  </tbody>
</table>
```

- `togglePermission` → si estaba activo: `DELETE /api/areas/{id}/collections/{name}`; si no: `POST /api/areas/{id}/collections/{name}` con `{ permission }`
- Cambios optimistas con rollback si falla

### Tests (Fase 9)

| Test | Qué verifica |
|------|-------------|
| PermissionsMatrix | colección con READ=true → checkbox checked; toggle → emite grantCollection |
| AreaCard | muestra nombre, # usuarios, # colecciones |
| BFF `POST /api/areas` | Mock → área creada con id |
| BFF `POST /api/areas/[id]/collections/[name]` | Mock → 200 con permiso |
| BFF `DELETE /api/areas/[id]/collections/[name]` | Mock → 204 |

---

## Fase 10 — Admin RAG Config

**Goal:** UI para configurar parámetros del RAG engine: temperatura, top-k, modelos, guardrails, y cambio de perfil de deployment.

### Parámetros configurables (de `saldivia/config.py` ENV_MAPPING)

| Parámetro | Tipo | Rango |
|-----------|------|-------|
| `temperature` | float | 0.0 - 2.0 |
| `max_tokens` | int | 256 - 8192 |
| `top_p` | float | 0.0 - 1.0 |
| `top_k` | int | 1 - 100 |
| `vdb_top_k` | int | 1 - 50 |
| `reranker_top_k` | int | 1 - 20 |
| `llm_model` | string | lista de modelos |
| `embedding_model` | string | lista |
| `reranker_model` | string | lista |
| `guardrails_enabled` | bool | — |

### Archivos

```
src/routes/(app)/admin/rag-config/
├── +page.svelte             — CREAR ESTE ARCHIVO (ruta existe en Sidebar pero falta)
└── +page.server.ts          — load: config actual; actions: updateConfig, resetToDefaults

src/lib/components/admin/
├── ConfigSlider.svelte      — slider + input numérico sincronizados
├── ModelSelector.svelte     — select con modelos disponibles (fetched del gateway)
├── GuardrailsToggle.svelte  — toggle con descripción
└── ProfileSwitcher.svelte   — selector de perfil (brev-2gpu, workstation-1gpu, full-cloud)

src/routes/api/admin/config/
└── +server.ts               — GET/PATCH: proxy a gateway /admin/config
```

### Gateway endpoint nuevo requerido

```python
# En saldivia/gateway.py agregar:
@router.get("/admin/config")
async def get_rag_config(user: User = Depends(require_admin)):
    return config.get_all_params()

@router.patch("/admin/config")
async def update_rag_config(params: dict, user: User = Depends(require_admin)):
    config.update(params)
    return {"ok": True}
```

### `ConfigSlider.svelte`

```svelte
<script lang="ts">
  let { label, value = $bindable(), min, max, step = 0.1, description } = $props<{
    label: string; value: number; min: number; max: number;
    step?: number; description?: string;
  }>();
</script>

<div class="mb-4">
  <div class="flex justify-between mb-1">
    <label class="text-sm font-medium text-[var(--text)]">{label}</label>
    <input type="number" bind:value {min} {max} {step}
           class="w-20 text-right text-sm bg-transparent border-b border-[var(--border)]
                  text-[var(--text)] outline-none focus:border-[var(--accent)]" />
  </div>
  <input type="range" bind:value {min} {max} {step} class="w-full accent-[var(--accent)]" />
  {#if description}
    <p class="text-xs text-[var(--text-faint)] mt-1">{description}</p>
  {/if}
</div>
```

### Página de config

- Secciones: Generación | Vector DB | Modelos | Guardrails | Perfil
- "Guardar cambios" → `PATCH /api/admin/config` con los parámetros → Toast "Configuración actualizada"
- "Restaurar defaults" → reset a valores del perfil activo + Toast
- Warning banner si `profile !== 'brev-2gpu'` en producción

### `ProfileSwitcher.svelte` — comportamiento especial

Cambiar de perfil es una operación **distinta** a PATCH /admin/config:
- Endpoint separado: `POST /api/admin/profile` con `{ profile: 'brev-2gpu' | 'workstation-1gpu' | 'full-cloud' }`
- **Alcance de esta fase: cambio de perfil en runtime solamente** — el gateway actualiza los parámetros en memoria usando `config.switch_profile(name)` (carga el YAML del nuevo perfil), pero **NO reinicia los servicios Docker**. El restart es manual.
- La UI muestra qué perfil está activo actualmente y permite seleccionar otro; al confirmar muestra un Toast "Perfil cambiado a brev-2gpu — algunos cambios requieren reinicio manual"
- **Por qué no se puede hacer restart automático:** el gateway corre dentro de Docker y no tiene acceso al host socket por defecto. Montar `/var/run/docker.sock` es out-of-scope (riesgo de seguridad). El restart se hace vía `ssh` + `make restart` por el operador.
- **Gateway endpoint nuevo**: `POST /admin/profile` que solo llama `config.switch_profile(name)` y retorna `{ ok: true, profile: name }`

### Tests (Fase 10)

| Test | Qué verifica |
|------|-------------|
| ConfigSlider | slider=0.5 sync con number input; fuera de rango → clamped |
| ModelSelector | lista de modelos renderizada; onChange emite modelo |
| BFF `GET /api/admin/config` | Mock → objeto config completo |
| BFF `PATCH /api/admin/config` | Payload { temperature: 0.8 } → gateway recibe |
| Página carga | sin /admin/rag-config existente → no 404 después de esta fase |

---

## Fase 11 — Admin Sistema + GPU

**Goal:** Dashboard de salud del sistema, monitoreo GPU real-time, y control del ModeManager (query/ingest mode switching).

### Componentes del sistema a monitorear

```
Gateway (puerto 9000)     — latencia, uptime, error rate
RAG Server (puerto 8081)  — status
Milvus                    — connected / entity count global
GPU #0 (embed/rerank)     — VRAM %, temp, modo actual
GPU #1 (LLM)              — VRAM %, temp, model loaded
```

### Funcionalidades

1. **`/admin/system`** — dashboard general con health cards
2. **GPU metrics** — polling cada 5s con SSE o fetch (según lo que soporte el gateway)
3. **Mode Manager control** — botones "Modo Query" / "Modo Ingest" con confirm
4. **Service health** — checks de conectividad a cada servicio
5. **Logs recientes** — últimas 20 líneas de logs del gateway (si el endpoint existe)

### Archivos

```
src/routes/(app)/admin/system/
├── +page.svelte             — dashboard sistema
└── +page.server.ts          — load: health inicial

src/routes/api/admin/
├── health/+server.ts        — GET: proxy a gateway /health
├── gpu/+server.ts           — GET: métricas GPU actuales
└── mode/+server.ts          — POST: {mode: 'query'|'ingest'} → ModeManager.switch_to_*

src/lib/components/admin/
├── HealthCard.svelte        — card: service name, status badge, latencia
├── GpuGauge.svelte          — gauge circular: VRAM % (SVG)
├── ModeSwitcher.svelte      — botones query/ingest con confirm dialog
└── SystemTimeline.svelte    — timeline de eventos del sistema (uptime, switches)
```

### `GpuGauge.svelte`

```svelte
<!-- SVG circular gauge -->
<script lang="ts">
  let { label, percent, temp } = $props<{ label: string; percent: number; temp: number }>();
  const r = 36;
  const circumference = 2 * Math.PI * r;
  let dashoffset = $derived(circumference * (1 - percent / 100));
</script>

<div class="flex flex-col items-center gap-1">
  <svg width="88" height="88" viewBox="0 0 88 88">
    <circle cx="44" cy="44" r={r} fill="none" stroke="var(--bg-hover)" stroke-width="8"/>
    <circle cx="44" cy="44" r={r} fill="none" stroke="var(--accent)" stroke-width="8"
            stroke-dasharray={circumference} stroke-dashoffset={dashoffset}
            stroke-linecap="round" transform="rotate(-90 44 44)" class="transition-all"/>
    <text x="44" y="48" text-anchor="middle" class="text-sm font-bold fill-[var(--text)]">
      {percent}%
    </text>
  </svg>
  <div class="text-xs text-[var(--text-muted)]">{label}</div>
  <div class="text-xs text-[var(--text-faint)]">{temp}°C</div>
</div>
```

### Polling con `$effect`

```svelte
<!-- En +page.svelte del sistema -->
$effect(() => {
  let interval: number;
  if (autoRefresh) {
    interval = setInterval(async () => {
      gpuData = await fetchGpuMetrics();
    }, 5000);
  }
  return () => clearInterval(interval);
});
```

### Gateway endpoints nuevos requeridos

```python
@router.get("/admin/gpu")
async def get_gpu_metrics(user: User = Depends(require_admin)):
    return mode_manager.get_status()

@router.post("/admin/mode")
async def switch_mode(mode: str, user: User = Depends(require_admin)):
    if mode == 'ingest':
        await mode_manager.switch_to_ingest_mode()
    else:
        await mode_manager.switch_to_query_mode()
    return {"mode": mode}
```

### Tests (Fase 11)

| Test | Qué verifica |
|------|-------------|
| GpuGauge | percent=75 → dashoffset correcto; percent=0 → gauge vacío |
| HealthCard | status='healthy' → badge verde; 'degraded' → amarillo; 'down' → rojo |
| ModeSwitcher | click ingest → confirm modal; confirm → emit mode change |
| BFF `GET /api/admin/health` | Mock gateway → health payload |
| BFF `POST /api/admin/mode` | {mode:'ingest'} → gateway recibe |

---

## Fase 12 — Auditoría Pro

**Goal:** Página de auditoría completa con filtros, paginación, export CSV, estadísticas, y auto-refresh.

### Funcionalidades actuales

`/(app)/audit/+page.svelte` existe — muestra tabla básica de audit log. Agregar:

1. **Filtros** — usuario, acción, colección, rango de fechas
2. **Paginación** — cursor-based, 50 entradas por página
3. **Export CSV** — descarga client-side del resultado filtrado
4. **Stats summary** — total de eventos, usuarios únicos, acción más frecuente
5. **Auto-refresh** — toggle para refrescar cada 30s
6. **Entry detail** — expandir fila para ver payload completo

### Archivos

```
src/routes/(app)/audit/
├── +page.svelte             — página pro (reemplazar existente)
└── +page.server.ts          — load con filtros desde searchParams

src/lib/components/audit/
├── AuditFilters.svelte      — panel de filtros: user, action, dates
├── AuditTable.svelte        — tabla con filas expandibles
├── AuditRow.svelte          — fila individual con expand
├── AuditStats.svelte        — cards de estadísticas
└── ExportButton.svelte      — genera y descarga CSV

src/routes/api/audit/
└── +server.ts               — GET con filtros → proxy gateway /admin/audit
```

### Filtros

```svelte
<script lang="ts">
  let filters = $state({
    userId: '',
    action: '',
    collection: '',
    dateFrom: '',
    dateTo: '',
  });

  // Sync filtros con URL params para deep-linkability
  $effect(() => {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([k, v]) => { if (v) params.set(k, v); });
    goto(`?${params.toString()}`, { replaceState: true, keepFocus: true });
  });
</script>
```

### Export CSV

```typescript
function exportCSV(entries: AuditEntry[]) {
  const headers = ['Fecha', 'Usuario', 'Acción', 'Colección', 'IP', 'Resultado'];
  const rows = entries.map(e => [
    e.created_at, e.user_email, e.action, e.collection ?? '', e.ip, e.success ? 'OK' : 'Error'
  ]);
  const csv = [headers, ...rows].map(r => r.join(',')).join('\n');
  downloadText(csv, `audit-${Date.now()}.csv`);
}
```

### `AuditStats.svelte`

- Total de eventos en el período seleccionado
- Usuarios únicos activos
- Acción más frecuente (con badge)
- Tasa de éxito (% de acciones sin error)

### Tests (Fase 12)

| Test | Qué verifica |
|------|-------------|
| AuditFilters | cambio de userId → URL param actualizado |
| AuditTable | 10 entries → 10 filas; click fila → expande payload |
| exportCSV | 3 entries → CSV con 4 líneas (header + 3) |
| AuditStats | 10 eventos, 3 únicos → renderiza correctamente |
| BFF `GET /api/audit?action=chat` | Mock → filtro forwarded al gateway |

---

## Fase 13 — Dashboard Analytics

**Goal:** Página principal de analytics con charts de uso, heatmap de actividad, feed de actividad reciente, y quick actions.

### Métricas a mostrar

| Métrica | Cómo obtener |
|---------|-------------|
| Mensajes por día (últimos 30d) | Agregar audit log por acción `chat` |
| Queries por colección | Agregar sessions por collection |
| Usuarios activos | distinct users en audit últimos 7d |
| Ingesta por día | audit action `ingest` |
| Latencia P50/P95 | gateway expone métricas (o calculado de audit timestamps) |

### Archivos

```
src/routes/(app)/dashboard/
├── +page.svelte             — dashboard principal (NUEVO — agregar a nav)
└── +page.server.ts          — load: analytics data

src/lib/components/dashboard/
├── MetricCard.svelte        — número grande + tendencia (↑↓) + sparkline
├── BarChart.svelte          — gráfico de barras SVG nativo (sin deps externas)
├── ActivityHeatmap.svelte   — grid 7×52 GitHub-style (SVG)
├── ActivityFeed.svelte      — timeline de eventos recientes
├── QuickActions.svelte      — botones: Nueva consulta, Subir doc, Nueva área
└── CollectionUsageChart.svelte — barras horizontales por colección

src/routes/api/analytics/
└── +server.ts               — GET: agrega datos para dashboard
```

### `BarChart.svelte` (SVG nativo, sin Chart.js)

```svelte
<script lang="ts">
  let { data, label } = $props<{ data: { x: string; y: number }[]; label: string }>();
  let maxVal = $derived(Math.max(...data.map(d => d.y)));
  const HEIGHT = 80;
  const BAR_W = 8;
  const GAP = 3;
</script>

<div>
  <p class="text-xs text-[var(--text-faint)] mb-1">{label}</p>
  <svg width={data.length * (BAR_W + GAP)} height={HEIGHT + 20}>
    {#each data as item, i}
      {@const barH = maxVal > 0 ? (item.y / maxVal) * HEIGHT : 0}
      <rect
        x={i * (BAR_W + GAP)} y={HEIGHT - barH}
        width={BAR_W} height={barH}
        rx="2" fill="var(--accent)" opacity="0.8"
      />
      <!-- tooltip en hover via title -->
      <title>{item.x}: {item.y}</title>
    {/each}
  </svg>
</div>
```

**Nota:** Sin dependencias externas (Chart.js, D3, etc.) — todo SVG nativo. Esto mantiene el bundle pequeño y evita deps problemáticas.

### `ActivityHeatmap.svelte`

- 52 semanas × 7 días (GitHub-style)
- Color intensity basado en número de queries ese día
- 4 niveles: 0 queries → `var(--bg-hover)`, 1-5 → low, 6-20 → mid, >20 → `var(--accent)`
- Tooltip con fecha + count al hover

### `MetricCard.svelte`

```svelte
<!-- Número grande, delta, sparkline -->
<div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] p-4">
  <div class="text-xs text-[var(--text-faint)] mb-1 uppercase tracking-wider">{title}</div>
  <div class="text-3xl font-bold text-[var(--text)]">{value.toLocaleString()}</div>
  {#if delta !== undefined}
    <div class="text-xs mt-1 {delta >= 0 ? 'text-[var(--success)]' : 'text-[var(--danger)]'}">
      {delta >= 0 ? '↑' : '↓'} {Math.abs(delta)}% vs semana pasada
    </div>
  {/if}
</div>
```

### Tests (Fase 13)

| Test | Qué verifica |
|------|-------------|
| BarChart | data=[{x:'Lun',y:5},...] → rect height proporcional |
| ActivityHeatmap | 365 celdas; 0 queries → color bg-hover |
| MetricCard | delta=-10 → texto rojo "↓ 10%" |
| BFF `GET /api/analytics` | Mock audit → agrega mensajes por día |

---

## Fase 14 — Notificaciones Real-Time

**Goal:** Sistema de notificaciones con SSE permanente, notification center, y updates en tiempo real de jobs de ingesta.

### Tipos de notificaciones

| Tipo | Trigger | Acción |
|------|---------|--------|
| `ingest_complete` | Job de ingesta termina | Link a colección |
| `ingest_failed` | Job falla | Link a upload con retry |
| `ingest_stalled` | Deadlock detectado | Acción retry |
| `user_created` | Admin crea usuario | — |
| `system_alert` | GPU memory crítica, etc. | Link a /admin/system |
| `new_message` | (futuro) mención en chat | — |

### Arquitectura

```
Browser → GET /api/notifications/stream (SSE permanente, lifetime de la sesión)
       ← event: notification {type, title, body, href, id}

Gateway → POST /internal/notify (webhook interno de otros servicios)
        → broadcast a todos los SSE streams activos del mismo user
```

### Archivos

```
src/routes/api/notifications/
├── stream/+server.ts        — SSE: mantiene conexión viva, envía eventos
└── dismiss/+server.ts       — POST: marcar notificación como leída

src/lib/stores/notifications.svelte.ts
                             — store: notifications[], unreadCount, conectar SSE

src/lib/components/
├── NotificationBell.svelte  — icono en header con badge de unread count
└── NotificationCenter.svelte — dropdown/panel con lista de notificaciones

src/lib/components/layout/
└── AppShell.svelte          — wrapper que inicia SSE de notificaciones al montar
```

### `NotificationStore`

```typescript
class NotificationStore {
  notifications = $state<Notification[]>([]);
  unreadCount = $derived(this.notifications.filter(n => !n.read).length);
  private es: EventSource | null = null;

  connect() {
    this.es = new EventSource('/api/notifications/stream');
    this.es.onmessage = (e) => {
      const notif = JSON.parse(e.data) as Notification;
      this.notifications = [notif, ...this.notifications].slice(0, 50);
    };
    this.es.onerror = () => {
      // reconnect después de 5s
      setTimeout(() => this.connect(), 5000);
    };
  }

  dismiss(id: string) {
    this.notifications = this.notifications.map(n =>
      n.id === id ? { ...n, read: true } : n
    );
    fetch('/api/notifications/dismiss', { method: 'POST', body: JSON.stringify({ id }) });
  }

  disconnect() { this.es?.close(); }
}
export const notifs = new NotificationStore();
```

### `NotificationBell.svelte`

- Campana en el header (junto a otras acciones)
- Badge rojo con `notifs.unreadCount` cuando > 0
- Click → abre/cierra `NotificationCenter`
- Animación shake cuando llega notificación nueva

### `NotificationCenter.svelte`

- Dropdown o panel deslizante (200px ancho en desktop)
- Lista de últimas 20 notificaciones
- Cada item: ícono por tipo, título, body truncado, tiempo relativo ("hace 3 min")
- Click → navega a `href` + marca como leída
- "Marcar todas como leídas" en el header del panel

### Gateway endpoint SSE

**Asunción de deployment:** Brev corre uvicorn con `--workers 1` (worker único). Las colas asyncio en memoria son válidas porque hay un solo proceso. Si en el futuro se escala a múltiples workers, habría que migrar a Redis pub/sub — out of scope por ahora.

```python
# En saldivia/gateway.py — agregar módulo notification_manager.py

# notification_manager.py
from asyncio import Queue
from collections import defaultdict
from typing import Dict, List

class NotificationManager:
    def __init__(self):
        self._queues: Dict[int, List[Queue]] = defaultdict(list)

    def subscribe(self, user_id: int) -> Queue:
        q: Queue = Queue()
        self._queues[user_id].append(q)
        return q

    def unsubscribe(self, user_id: int, q: Queue):
        try:
            self._queues[user_id].remove(q)
        except ValueError:
            pass  # ya fue removida

    async def notify(self, user_id: int, event: dict):
        for q in self._queues.get(user_id, []):
            await q.put(event)

notification_manager = NotificationManager()  # singleton, válido con --workers 1

@router.get("/v1/notifications/stream")
async def notification_stream(user: User = Depends(require_auth)):
    async def event_gen():
        queue = notification_manager.subscribe(user.id)
        try:
            while True:
                notif = await queue.get()
                yield f"data: {json.dumps(notif)}\n\n"
        finally:
            notification_manager.unsubscribe(user.id, queue)
    return StreamingResponse(event_gen(), media_type="text/event-stream")
```

### Tests (Fase 14)

| Test | Qué verifica |
|------|-------------|
| NotificationStore | addNotification → unreadCount++ |
| NotificationStore | dismiss → notif.read=true; unreadCount-- |
| NotificationBell | unreadCount=3 → badge visible con "3" |
| BFF `/api/notifications/stream` | Mock event → llega a store |

---

## Fase 15 — Settings Pro

**Goal:** Configuración personal avanzada: parámetros crossdoc, preferencias de UI, zona horaria, y gestión de múltiples API keys.

### Secciones (extending settings/+page.svelte existente)

1. **Crossdoc avanzado** — maxSubQueries, synthesisModel, followUpRetries, showDecomposition, vdbTopK, rerankerTopK (de SaldiviaSection.tsx en patches)
2. **Preferencias UI** — idioma de respuestas (español/inglés), densidad de información (compact/normal/spacious), auto-scroll (on/off)
3. **Zona horaria** — select con Intl.supportedValuesOf('timeZone'), persiste en user profile
4. **API Keys** — lista de keys named del usuario (Phase 15 base), crear named key, revocar

### Archivos

```
src/routes/(app)/settings/
├── +page.svelte             — expandir existente con nuevas secciones
└── +page.server.ts          — expandir con nuevas actions

src/lib/components/settings/
├── CrossdocSettings.svelte  — parámetros crossdoc (reutiliza ConfigSlider de Fase 9)
├── UIPreferences.svelte     — preferencias visuales
├── TimezoneSelector.svelte  — select con búsqueda de TZ
└── ApiKeyList.svelte        — lista de named keys con revoke

src/lib/stores/preferences.svelte.ts
                             — store: preferencias del usuario (localStorage + gateway sync)
```

### `PreferencesStore`

```typescript
class PreferencesStore {
  // Persiste en localStorage + sync al gateway en /api/me/preferences
  crossdoc = $state<CrossdocOptions>(DEFAULT_CROSSDOC_OPTIONS);
  uiDensity = $state<'compact' | 'normal' | 'spacious'>('normal');
  responseLanguage = $state<'es' | 'en'>('es');
  timezone = $state(Intl.DateTimeFormat().resolvedOptions().timeZone);

  save() {
    localStorage.setItem('sda_prefs', JSON.stringify($state.snapshot(this)));
    fetch('/api/me/preferences', { method: 'PUT', body: JSON.stringify(...) });
  }

  load() {
    const saved = localStorage.getItem('sda_prefs');
    if (saved) Object.assign(this, JSON.parse(saved));
  }
}
```

### Crossdoc settings (port de SaldiviaSection.tsx)

```svelte
<!-- CrossdocSettings.svelte -->
<ConfigSlider bind:value={prefs.crossdoc.maxSubQueries} min={1} max={8} step={1}
  label="Máximo de sub-consultas" description="Cuántas sub-preguntas genera el decomposer" />

<ConfigSlider bind:value={prefs.crossdoc.followUpRetries} min={0} max={3} step={1}
  label="Reintentos de follow-up" description="Intentos extra para sub-consultas fallidas" />

<ConfigSlider bind:value={prefs.crossdoc.vdbTopK} min={1} max={30} step={1}
  label="VDB Top-K" description="Documentos recuperados por sub-consulta" />

<label class="flex items-center gap-2 text-sm">
  <input type="checkbox" bind:checked={prefs.crossdoc.showDecomposition}
         class="accent-[var(--accent)]" />
  Mostrar descomposición de consulta en el chat
</label>
```

### Tests (Fase 15)

| Test | Qué verifica |
|------|-------------|
| PreferencesStore | load de localStorage → crossdoc.maxSubQueries correcto |
| PreferencesStore | save → localStorage actualizado |
| CrossdocSettings | slider maxSubQueries → store actualizado |
| TimezoneSelector | búsqueda "America" → filtra correctamente |

---

## Fase 16 — Developer + API

**Goal:** Herramientas para developers: named API keys con scopes, Swagger UI embebido, webhook configuration, y API playground.

### Funcionalidades

1. **Named API keys** — crear keys con nombre + scope + expiry; listar; revocar
2. **Swagger UI** — embed del spec OpenAPI del gateway en un iframe + URL copiable
3. **Webhook configuration** — URL + secret + eventos a subscripcionar
4. **API Playground** — editor de queries con collection selector y response viewer

### Archivos

```
src/routes/(app)/developer/
├── +page.svelte             — hub: keys, swagger, webhooks, playground
└── +page.server.ts          — load: user's named keys

src/lib/components/developer/
├── ApiKeyManager.svelte     — CRUD de named keys
├── ApiKeyCreateModal.svelte — modal: nombre, scope, expiry
├── SwaggerEmbed.svelte      — iframe con Swagger UI del gateway
├── WebhookConfig.svelte     — form: URL, secret, eventos
└── ApiPlayground.svelte     — editor + send + response viewer

src/routes/api/developer/
├── keys/+server.ts          — GET/POST named keys
├── keys/[id]/+server.ts     — DELETE
└── webhooks/+server.ts      — GET/POST/DELETE webhooks
```

### Named API Keys

```typescript
interface NamedApiKey {
  id: string;
  name: string;
  prefix: string;  // primeros 8 chars: "rsk_live..."
  scope: 'read' | 'write' | 'admin';
  expiresAt: string | null;
  lastUsedAt: string | null;
  createdAt: string;
}
```

- Lista muestra: nombre, prefijo (truncado), scope badge, última vez usada, fecha expiración
- Al crear → modal muestra key completa UNA SOLA VEZ con copy button
- Revocar → confirm modal + optimistic delete

### API Playground

```svelte
<!-- Editor básico: pregunta, colección, parámetros, send, response -->
<div class="grid grid-cols-2 gap-4 h-full">
  <div class="flex flex-col gap-3">
    <label class="text-xs font-medium text-[var(--text-faint)]">Pregunta</label>
    <textarea bind:value={query} rows={4} class="..." />
    <select bind:value={collection}>...</select>
    <Button onclick={sendQuery} disabled={loading}>Enviar consulta</Button>
  </div>
  <div>
    <label class="text-xs font-medium text-[var(--text-faint)]">Respuesta</label>
    <pre class="bg-[var(--bg-hover)] rounded p-3 text-xs overflow-auto h-[400px]">
      {JSON.stringify(response, null, 2)}
    </pre>
  </div>
</div>
```

### Tests (Fase 16)

| Test | Qué verifica |
|------|-------------|
| ApiKeyManager | lista vacía → empty state; key en lista → prefijo visible |
| ApiKeyCreateModal | submit → emite create event con {name, scope, expiresAt} |
| BFF `POST /api/developer/keys` | Mock → devuelve key completa |
| BFF `DELETE /api/developer/keys/[id]` | Mock → 204 |

---

## Fase 17 — Búsqueda Global + Command Palette

**Goal:** Command palette (Cmd+K) con búsqueda global over sesiones, colecciones, usuarios (admin), y acciones de navegación.

### Funcionalidades

1. **Atajo Cmd+K** — abre command palette modal desde cualquier página
2. **Búsqueda de sesiones** — filtra historial de chat por título
3. **Navegación rápida** — acciones: "Ir a colecciones", "Nueva consulta", "Configuración"
4. **Búsqueda de colecciones** — navega directo a `/collections/{name}`
5. **Búsqueda de usuarios** (admin only) — filtra por nombre/email
6. **Acciones contextuales** — según la página actual: "Exportar conversación", "Crear colección"
7. **Historial de comandos** — últimos 5 comandos usados (localStorage)

### Archivos

```
src/lib/components/ui/CommandPalette.svelte
                             — modal de búsqueda global (componente central)
src/lib/stores/commandPalette.svelte.ts
                             — store: open, query, results, selectedIndex
src/lib/actions/shortcut.ts  — Svelte action: use:shortcut={{ key:'k', meta:true }}
src/routes/(app)/+layout.svelte
                             — agregar CommandPalette al layout; registrar shortcut
```

### Command Palette UX

```
[Cmd+K] → modal con overlay
┌─────────────────────────────────────────┐
│ 🔍  Buscar o escribir un comando...      │
├─────────────────────────────────────────┤
│ ▶ ACCIONES                              │
│   → Nueva consulta              Ctrl+N  │
│   → Subir documento                     │
│   → Ir a Colecciones                    │
├─────────────────────────────────────────┤
│ ▶ SESIONES RECIENTES                    │
│   → "Análisis de balances Q4 2024"      │
│   → "Consulta sobre proveedores..."     │
├─────────────────────────────────────────┤
│ ▶ COLECCIONES                           │
│   → saldivia  (12.4k entidades)         │
│   → contratos (3.2k entidades)          │
└─────────────────────────────────────────┘
```

### `CommandPaletteStore`

```typescript
class CommandPaletteStore {
  open = $state(false);
  query = $state('');
  selectedIndex = $state(0);

  // sessions y collections se importan de otros stores:
  // import { chat } from '$lib/stores/chat.svelte';        → chat.sessions
  // import { collectionsStore } from '$lib/stores/collections.svelte';  → collectionsStore.list
  results = $derived.by(() => {
    if (!this.query) return DEFAULT_ACTIONS;
    return [
      ...filterSessions(this.query, chat.sessions),
      ...filterCollections(this.query, collectionsStore.list),
      ...filterActions(this.query),
    ];
  });

  toggle() { this.open = !this.open; if (this.open) this.query = ''; }
  close() { this.open = false; }
  select(item: CommandItem) { item.action(); this.close(); }
  moveDown() { this.selectedIndex = Math.min(this.selectedIndex + 1, this.results.length - 1); }
  moveUp() { this.selectedIndex = Math.max(this.selectedIndex - 1, 0); }
}
```

### Keyboard navigation

- `↑` / `↓` → navegar items
- `Enter` → ejecutar selected
- `Escape` → cerrar
- Typing → filtra resultados en tiempo real (sin fetch — todo client-side)

### Svelte action `shortcut`

```typescript
// src/lib/actions/shortcut.ts
export function shortcut(node: HTMLElement, { key, meta, ctrl, handler }: ShortcutOptions) {
  function handleKeydown(e: KeyboardEvent) {
    if (meta && !e.metaKey) return;
    if (ctrl && !e.ctrlKey) return;
    if (e.key === key) { e.preventDefault(); handler(); }
  }
  window.addEventListener('keydown', handleKeydown);
  return { destroy: () => window.removeEventListener('keydown', handleKeydown) };
}
```

### Tests (Fase 17)

| Test | Qué verifica |
|------|-------------|
| CommandPaletteStore | toggle → open=true; toggle again → open=false |
| CommandPaletteStore | query="saldivia" → filtra colecciones/sesiones |
| CommandPaletteStore | moveDown → selectedIndex++ |
| shortcut action | Cmd+K → handler llamado; Esc cierra |
| CommandPalette UI | 3 results → 3 items; selectedIndex=1 → segundo item highlighted |

---

## Fase 18 — Polish, Responsive, a11y, Onboarding

**Goal:** Refinamiento final: responsiveness móvil/tablet, accesibilidad WCAG AA, onboarding para usuarios nuevos, y performance optimization.

### Responsive (mobile-first)

| Breakpoint | Cambios |
|-----------|---------|
| `sm` (<640px) | Sidebar collapsa a bottom nav; HistoryPanel oculto; CommandPalette pantalla completa |
| `md` (640-1024px) | Sidebar 56px (iconos); panels laterales se superponen como drawers |
| `lg` (≥1024px) | Layout actual (sin cambios) |

### Archivos responsive

```
src/lib/components/layout/
├── BottomNav.svelte          — nav móvil: Chat, Colecciones, Settings
├── Drawer.svelte             — drawer lateral para móvil (HistoryPanel, SourcesPanel)
└── AppShell.svelte           — wrapper que elige layout según screen size

src/lib/stores/viewport.svelte.ts
                              — store: isMobile, isTablet, breakpoint
```

### `ViewportStore`

```typescript
class ViewportStore {
  width = $state(typeof window !== 'undefined' ? window.innerWidth : 1024);
  isMobile = $derived(this.width < 640);
  isTablet = $derived(this.width >= 640 && this.width < 1024);
  isDesktop = $derived(this.width >= 1024);

  init() {
    if (typeof window === 'undefined') return;
    const observer = new ResizeObserver(() => { this.width = window.innerWidth; });
    observer.observe(document.body);
    return () => observer.disconnect();
  }
}
```

### Accesibilidad (a11y) WCAG AA

Checklist de implementación:

- [ ] Todos los botones icon-only tienen `aria-label`
- [ ] Modales usan `role="dialog"`, `aria-modal="true"`, `aria-labelledby`
- [ ] Foco atrapado dentro de modales (focus trap)
- [ ] Navegación por Tab completa en todos los formularios
- [ ] Contrast ratio ≥ 4.5:1 en texto normal, ≥ 3:1 en texto grande
- [ ] `aria-live="polite"` en zonas de streaming y notificaciones
- [ ] `prefers-reduced-motion`: desactivar animaciones si está activo
- [ ] Skip link "Saltar al contenido principal" al inicio del HTML

### `FocusTrap` action

```typescript
// src/lib/actions/focusTrap.ts
export function focusTrap(node: HTMLElement) {
  const focusable = node.querySelectorAll<HTMLElement>(
    'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
  );
  const first = focusable[0];
  const last = focusable[focusable.length - 1];

  function handleKeydown(e: KeyboardEvent) {
    if (e.key !== 'Tab') return;
    if (e.shiftKey && document.activeElement === first) {
      e.preventDefault(); last.focus();
    } else if (!e.shiftKey && document.activeElement === last) {
      e.preventDefault(); first.focus();
    }
  }

  node.addEventListener('keydown', handleKeydown);
  first?.focus();
  return { destroy: () => node.removeEventListener('keydown', handleKeydown) };
}
```

### Onboarding

- Modal de bienvenida para nuevos usuarios (primer login, basado en `user.created_at` < 24h)
- 3 pasos: "Bienvenido a SDA" → "Elegí una colección" → "Hacé tu primera consulta"
- Persiste en `localStorage` si el usuario lo cierra sin completar
- Quick tour con `position: fixed` tooltips sobre elementos clave (arrow apuntando)

### Performance

- [ ] Lazy loading de componentes pesados (`import()` dinámico para Swagger embed, chart components)
- [ ] `sveltekit:prefetch` en links de navegación frecuente
- [ ] Imágenes/íconos: usar `lucide-svelte` (ya tree-shakeable)
- [ ] `$derived` en lugar de `$effect` + variable para cómputos derivados
- [ ] Skeleton loading en todas las páginas con server load
- [ ] Cache de `+page.server.ts` con `setHeaders({ 'cache-control': 'private, max-age=60' })`

### Tests (Fase 18)

| Test | Qué verifica |
|------|-------------|
| ViewportStore | resize a 400px → isMobile=true |
| focusTrap action | Tab en último focusable → vuelve al primero |
| BottomNav | renderiza en mobile; no renderiza en desktop |
| Modal a11y | aria-modal=true; aria-labelledby apunta al h2 del modal |

---

## Dependencias nuevas por fase

| Fase | Package | Razón |
|------|---------|-------|
| 2 (ya spec) | `marked`, `highlight.js`, `dompurify` | Markdown rendering |
| 6 | Ninguna | Page count y split lo hace el gateway (Python) |
| 13 | Ninguna (SVG nativo) | Charts sin deps |
| 14 | Ninguna (EventSource nativo) | SSE sin polyfill |
| 18 | Ninguna | Todo nativo |
| 19 | Ninguna | Error tracking con logging nativo Python + fetch nativo |
| 20 | Ninguna | Canvas API nativa para screenshot |
| 21 | `ragas>=0.2`, LLM judge vía OpenRouter | Evaluación automática faithfulness/relevancy/recall/precision en CI |
| 21 | `apscheduler>=3.10` | BackgroundScheduler para cron diario del flywheel en gateway startup |

**Principio:** Minimizar dependencias externas. SVG nativo > Chart.js. EventSource nativo > socket.io.

**Nota Fase 21:** `ragas` requiere un LLM judge para calcular métricas. Usar el mismo `OPENROUTER_API_KEY` ya disponible. Configurar como variable de entorno en GitHub Actions Secrets.

---

## Endpoints de gateway nuevos requeridos

| Fase | Método | Endpoint | Función |
|------|--------|----------|---------|
| 4 | `POST` | `/v1/collections` | Crear colección |
| 4 | `DELETE` | `/v1/collections/{name}` | Eliminar colección |
| 6 | `GET` | `/v1/ingest/{jobId}/status` | Estado de job de ingesta (poller) |
| 7 | `POST` | `/v1/messages/{id}/feedback` | Feedback thumbs |
| 10 | `GET/PATCH` | `/admin/config` | RAG config CRUD |
| 10 | `POST` | `/admin/profile` | Cambiar perfil (runtime) |
| 11 | `GET` | `/admin/gpu` | Métricas GPU |
| 11 | `POST` | `/admin/mode` | Switch mode |
| 14 | `GET` | `/v1/notifications/stream` | SSE notificaciones |
| 15 | `GET` | `/api/me/preferences` | Leer preferencias del usuario |
| 15 | `PUT` | `/api/me/preferences` | Guardar preferencias (crossdoc opts, UI density, etc.) |
| 16 | `GET/POST` | `/api/keys` | Named API keys |
| 16 | `DELETE` | `/api/keys/{id}` | Revocar key |
| 16 | `GET/POST` | `/api/webhooks` | Webhooks CRUD |
| 16 | `DELETE` | `/api/webhooks/{id}` | Eliminar webhook |
| 19 | `POST` | `/internal/errors` | Recibir error reports del BFF |
| 19 | `GET` | `/admin/errors` | Listar errores con filtros |
| 19 | `GET` | `/admin/errors/stats` | Estadísticas de errores |
| 20 | `POST` | `/v1/feedback` | Guardar feedback UX de usuario |
| 20 | `GET` | `/admin/feedback` | Listar feedback UX (admin) |
| 21 | `GET` | `/admin/flywheel/report` | Último DailyReport del flywheel |
| 21 | `GET` | `/admin/flywheel/history` | Historia de evaluaciones RAGAS |
| 21 | `POST` | `/admin/flywheel/evaluate` | Trigger manual de evaluación |
| 22 | `POST` | `/v1/human-feedback` | Submit corrección / sugerencia |
| 22 | `GET` | `/admin/human-feedback` | Cola de revisión (admin) |
| 22 | `PATCH` | `/admin/human-feedback/{id}` | Aprobar / rechazar |

---

## Orden de implementación sugerido

Las fases son mayormente independientes, pero hay dependencias lógicas:

```
3 (CI/CD) ← PRIMERO — beneficia a todas las fases siguientes
4 (Colecciones) ← base para 6 (Upload), 5 usa /chat
5 (Crossdoc) ← puede ir en paralelo con 6
6 (Upload) ← depende de Fase 4 (collections selector)
7 (Chat Sesiones Pro) ← puede ir en paralelo con 4-6
8 (Admin Users) ← puede ir en paralelo con 4-7
9 (Admin Áreas) ← depende de 8 (UserSelector)
10 (RAG Config) ← independiente
11 (Sistema + GPU) ← después de 10 (reutiliza ConfigSlider)
12 (Auditoría) ← independiente
13 (Dashboard) ← después de 12 (usa mismos datos de audit)
14 (Notifs RT) ← independiente (puede hacerse en cualquier momento)
15 (Settings Pro) ← después de 5 (Crossdoc) para reutilizar CrossdocSettings
16 (Developer) ← independiente
17 (Command Palette) ← después de varias fases (necesita sesiones, colecciones)
19 (Observability) ← puede ir en paralelo con cualquier fase (es transversal)
20 (Feedback Widget) ← después de 19 (comparte storage y admin UI)
21 (Data Flywheel) ← después de 7 (thumbs), 20 (widget) y 22 (correcciones humanas)
22 (Feedback Humano) ← después de 7 (chat) y 21 (golden dataset)
18 (Polish) ← última, siempre
```

**Agrupamiento paralelo ideal:**
- Sprint 0: **3 (CI/CD)** — configura el pipeline antes de todo
- Sprint A: 4 + 7 + 8 (base de datos y gestión)
- Sprint B: 5 + 6 (documentos e IA)
- Sprint C: 9 + 10 + 11 (admin avanzado)
- Sprint D: 12 + 13 + 14 (observabilidad)
- Sprint E: 15 + 16 + 17 (developer tools + search)
- Sprint F: **19 + 20** (error tracking + feedback widget)
- Sprint G: **21 + 22** (flywheel + feedback humano) — cierran el loop de calidad
- Sprint H: **18** (polish)

---

## Fase 19 — Observability: Error Tracking Completo

**Goal:** Captura y registro centralizado de todos los errores posibles en el sistema: frontend, BFF, gateway, y capa NVIDIA (NIMs, GPU, RAG Server, Milvus). Admin UI para visualizar, filtrar y exportar.

### Capas de captura

| Capa | Qué se captura | Cómo |
|------|---------------|------|
| Frontend Svelte | `window.onerror`, `unhandledrejection`, fetch errors, SSE drops, errores en Svelte error boundaries | Middleware global en `+layout.svelte` |
| BFF SvelteKit | Excepciones en server routes, timeouts al gateway, errores de sesión expirada | `handleError` hook en `hooks.server.ts` |
| Gateway FastAPI | Todos los HTTP 4xx/5xx, excepciones no capturadas, timeouts a RAG server | FastAPI exception handler global + middleware |
| NIM/RAG-específico | Timeout de inferencia LLM (>120s), NIM HTTP 5xx, Triton connection errors, embedding failures | Captura en `saldivia/providers.py` |
| GPU/CUDA | GPU OOM (CUDA error 2), temperatura crítica (>85°C), VRAM >95% | Polling en `mode_manager.py` + alertas |
| Milvus | Connection loss, query timeout, insert failures | Captura en `saldivia/collections.py` |
| Streaming SSE | Chunks malformados, stream cortado prematuramente, encode errors | En `gateway.py` stream handler |

### Schema SQLite

```sql
CREATE TABLE IF NOT EXISTS error_logs (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp    TEXT    NOT NULL DEFAULT (datetime('now')),
    level        TEXT    NOT NULL CHECK(level IN ('error','warning','critical')),
    source       TEXT    NOT NULL CHECK(source IN ('frontend','bff','gateway','nim','gpu','milvus','rag')),
    error_type   TEXT    NOT NULL,   -- e.g. 'NIMTimeout', 'GPUOutOfMemory', 'FetchError'
    message      TEXT    NOT NULL,
    stack_trace  TEXT,
    user_id      INTEGER,
    session_id   TEXT,
    request_id   TEXT,              -- correlación entre frontend ↔ gateway
    metadata     TEXT,             -- JSON: browser, URL, CUDA error code, etc.
    resolved     INTEGER NOT NULL DEFAULT 0,
    created_at   TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_error_logs_timestamp ON error_logs(timestamp DESC);
CREATE INDEX idx_error_logs_type ON error_logs(error_type);
CREATE INDEX idx_error_logs_source ON error_logs(source);
```

### Archivos

```
saldivia/
├── error_tracker.py         — ErrorTracker singleton: log(), get_stats(), get_errors()
├── middleware/
│   └── error_middleware.py  — FastAPI middleware: captura toda excepción → error_tracker.log()
└── providers.py             — MODIFICAR: wrap de llamadas NIM/RAG con captura de errores específicos

services/sda-frontend/src/
├── hooks.server.ts          — MODIFICAR: agregar handleError → POST /api/errors (BFF)
├── hooks.client.ts          — NUEVO: window.onerror + unhandledrejection → POST /api/errors
└── routes/
    ├── api/errors/+server.ts — BFF: recibe errores del cliente → forward a gateway
    └── (app)/admin/errors/
        ├── +page.svelte     — tabla de errores con filtros y stats
        └── +page.server.ts  — load: errores paginados + stats
```

### `ErrorTracker` (Python)

```python
# saldivia/error_tracker.py
import sqlite3, json, traceback
from datetime import datetime

class ErrorTracker:
    def log(self, level: str, source: str, error_type: str, message: str,
            stack_trace: str = None, user_id: int = None, session_id: str = None,
            request_id: str = None, metadata: dict = None):
        with self._db() as conn:
            conn.execute("""
                INSERT INTO error_logs
                  (level, source, error_type, message, stack_trace,
                   user_id, session_id, request_id, metadata)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
            """, (level, source, error_type, message, stack_trace,
                  user_id, session_id, request_id,
                  json.dumps(metadata) if metadata else None))

error_tracker = ErrorTracker()  # singleton
```

### `hooks.server.ts` (SvelteKit)

```typescript
// Captura todos los errores del servidor SvelteKit
export const handleError: HandleServerError = async ({ error, event }) => {
  const err = error as Error;
  await fetch(`${GATEWAY_URL}/internal/errors`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${SYSTEM_API_KEY}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({
      level: 'error', source: 'bff',
      error_type: err.name || 'UnknownError',
      message: err.message,
      stack_trace: err.stack,
      metadata: { url: event.url.pathname, method: event.request.method },
    }),
  }).catch(() => {}); // silencioso: no queremos error-en-el-error-handler
  return { message: 'Error interno del servidor' };
};
```

### `hooks.client.ts` (SvelteKit)

```typescript
// Captura errores de JavaScript no manejados en el browser
window.addEventListener('unhandledrejection', (event) => {
  reportError('UnhandledPromiseRejection', event.reason?.message, event.reason?.stack);
});
window.addEventListener('error', (event) => {
  reportError('UncaughtError', event.message, event.error?.stack, {
    filename: event.filename, lineno: event.lineno,
  });
});

function reportError(type: string, message: string, stack?: string, meta?: object) {
  fetch('/api/errors', {
    method: 'POST', keepalive: true, // keepalive para que se envíe aunque la página se cierre
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ error_type: type, message, stack_trace: stack,
                           metadata: { url: location.href, ...meta } }),
  }).catch(() => {});
}
```

### Errores NVIDIA-específicos

```python
# En saldivia/providers.py — wrap de llamadas NIM
async def embed(self, texts: list[str]) -> list[list[float]]:
    try:
        resp = await self.client.post('/v1/embeddings', json={'input': texts}, timeout=30)
        resp.raise_for_status()
        return resp.json()['data']
    except httpx.TimeoutException:
        error_tracker.log('error', 'nim', 'NIMEmbedTimeout',
                          f'Embedding timeout after 30s for {len(texts)} texts')
        raise
    except httpx.HTTPStatusError as e:
        error_tracker.log('critical', 'nim', f'NIMEmbedHTTP{e.response.status_code}',
                          str(e), metadata={'status': e.response.status_code})
        raise

# En saldivia/mode_manager.py — monitoreo GPU
def _check_gpu_alerts(self):
    status = self.get_status()
    for gpu in status.gpus:
        if gpu.vram_percent > 95:
            error_tracker.log('critical', 'gpu', 'GPUVRAMCritical',
                              f'GPU {gpu.id} VRAM at {gpu.vram_percent}%',
                              metadata={'gpu_id': gpu.id, 'vram': gpu.vram_percent})
        if gpu.temperature > 85:
            error_tracker.log('warning', 'gpu', 'GPUTemperatureHigh',
                              f'GPU {gpu.id} at {gpu.temperature}°C')
```

### Admin UI `/admin/errors`

- Tabla con: timestamp, level badge (error/warning/critical), source badge (frontend/gateway/gpu/nim…), error_type, message truncado, usuario
- **Agrupación por error_type**: "NIMTimeout × 47 en las últimas 24h" — detecta patrones
- Filtros: level, source, error_type, dateFrom/dateTo, userId, resolved
- Click en fila → modal con stack_trace completo + metadata JSON formateado
- Botón "Marcar como resuelto" → `PATCH /admin/errors/{id}` `{ resolved: true }`
- Export JSON de los resultados filtrados
- Stats cards: total errores hoy, críticos sin resolver, fuente con más errores, error más frecuente

### Tests (Fase 19)

| Test | Qué verifica |
|------|-------------|
| `ErrorTracker.log()` | Inserta en SQLite; campos obligatorios presentes |
| `error_middleware` | 500 en endpoint → error_tracker.log llamado con source='gateway' |
| `providers.py` embed timeout | httpx.TimeoutException → log con type='NIMEmbedTimeout' |
| `hooks.client.ts` | unhandledrejection → fetch a /api/errors con payload correcto |
| BFF `POST /api/errors` | Forward al gateway con campos correctos |
| Admin `/admin/errors` | Filtro source='gpu' → solo errores GPU en tabla |

---

## Fase 20 — Feedback Widget + Telemetría Máxima

**Goal:** Widget flotante global de feedback con captura de máxima información contextual. Admin UI para revisar, filtrar y exportar feedback. Integrado con los error logs de Fase 19.

### Schema SQLite

```sql
CREATE TABLE IF NOT EXISTS feedback (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    type            TEXT    NOT NULL CHECK(type IN ('bug','suggestion','other')),
    rating          INTEGER CHECK(rating BETWEEN 1 AND 5),
    message         TEXT    NOT NULL,
    -- Contexto capturado automáticamente
    user_id         INTEGER NOT NULL,
    user_role       TEXT,
    session_id      TEXT,              -- chat session activa si aplica
    url             TEXT    NOT NULL,
    route           TEXT,
    collection      TEXT,             -- colección activa si aplica
    -- Browser/sistema
    user_agent      TEXT,
    viewport_w      INTEGER,
    viewport_h      INTEGER,
    -- Telemetría
    recent_actions  TEXT,             -- JSON: últimas 20 acciones del usuario
    recent_errors   TEXT,             -- JSON: últimos 5 errores de error_logs del session
    deploy_sha      TEXT,             -- git SHA del deploy activo
    screenshot_b64  TEXT,             -- base64 PNG (opcional, capturado con canvas API)
    -- Timestamps
    time_on_page    INTEGER,          -- segundos en la página actual
    created_at      TEXT    NOT NULL DEFAULT (datetime('now'))
);
```

### Archivos

```
services/sda-frontend/src/
├── lib/
│   ├── stores/actionLog.svelte.ts    — buffer circular de las últimas 20 acciones
│   ├── components/
│   │   ├── FeedbackButton.svelte     — botón flotante (esquina inferior derecha)
│   │   └── FeedbackModal.svelte      — modal con form + preview de datos capturados
│   └── utils/screenshot.ts           — captura canvas → base64 PNG
└── routes/
    ├── api/feedback/+server.ts       — BFF: recibe feedback → gateway /v1/feedback
    └── (app)/admin/feedback/
        ├── +page.svelte              — tabla de feedback con filtros
        └── +page.server.ts           — load: feedback paginado
```

### `ActionLogStore`

```typescript
// Buffer circular de acciones del usuario — útil para contexto de bugs
class ActionLogStore {
  private actions = $state<UserAction[]>([]);

  log(type: string, detail?: object) {
    const action = { type, detail, timestamp: Date.now(), url: location.href };
    this.actions = [...this.actions.slice(-19), action]; // últimas 20
  }

  getRecent(): UserAction[] { return [...this.actions]; }
}
export const actionLog = new ActionLogStore();

// Uso en toda la app:
// actionLog.log('chat:submit', { collection: 'saldivia', crossdoc: true });
// actionLog.log('navigation', { to: '/collections' });
// actionLog.log('upload:start', { filename: 'doc.pdf', size: 1024000 });
```

### `FeedbackModal.svelte`

```svelte
<script lang="ts">
  let type = $state<'bug' | 'suggestion' | 'other'>('bug');
  let rating = $state(0);
  let message = $state('');
  let includeScreenshot = $state(false);
  let submitting = $state(false);

  // Datos capturados automáticamente — el usuario los puede ver en "Detalles técnicos"
  let autoData = $derived({
    url: location.href,
    userAgent: navigator.userAgent,
    viewport: `${window.innerWidth}×${window.innerHeight}`,
    recentActions: actionLog.getRecent(),
    deploySha: import.meta.env.VITE_GIT_SHA,
    timeOnPage: Math.round((Date.now() - pageLoadTime) / 1000),
  });

  async function submit() {
    submitting = true;
    const screenshot = includeScreenshot ? await captureScreenshot() : null;
    await fetch('/api/feedback', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ type, rating, message, screenshot_b64: screenshot, ...autoData }),
    });
    submitting = false;
    // Toast "Gracias por tu feedback" + cierra modal
  }
</script>

<!-- UI del modal -->
<div class="flex gap-2 mb-4">
  {#each ['bug', 'suggestion', 'other'] as t}
    <button onclick={() => type = t}
            class="px-3 py-1.5 rounded-full text-xs border transition-colors
                   {type === t ? 'bg-[var(--accent)] text-white border-[var(--accent)]'
                               : 'border-[var(--border)] text-[var(--text-muted)]'}">
      {t === 'bug' ? '🐛 Bug' : t === 'suggestion' ? '💡 Sugerencia' : '💬 Otro'}
    </button>
  {/each}
</div>

<!-- Rating: 5 estrellas -->
<div class="flex gap-1 mb-4">
  {#each [1,2,3,4,5] as n}
    <button onclick={() => rating = n} class="text-xl {n <= rating ? 'text-yellow-400' : 'text-[var(--border)]'}">★</button>
  {/each}
</div>

<textarea bind:value={message} placeholder="Describí el problema o sugerencia..."
          rows={4} class="w-full ..." />

<label class="flex items-center gap-2 text-xs mt-3">
  <input type="checkbox" bind:checked={includeScreenshot} />
  Incluir captura de pantalla
</label>

<!-- Detalles técnicos colapsables -->
<details class="mt-3">
  <summary class="text-xs text-[var(--text-faint)] cursor-pointer">Ver datos técnicos enviados</summary>
  <pre class="text-xs bg-[var(--bg-hover)] p-2 rounded mt-1 overflow-auto max-h-32">
    {JSON.stringify(autoData, null, 2)}
  </pre>
</details>
```

### `screenshot.ts`

```typescript
// Captura del viewport usando canvas API — sin librerías externas
export async function captureScreenshot(): Promise<string | null> {
  try {
    const stream = await navigator.mediaDevices.getDisplayMedia({ video: true });
    const video = document.createElement('video');
    video.srcObject = stream;
    await video.play();

    const canvas = document.createElement('canvas');
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    canvas.getContext('2d')!.drawImage(video, 0, 0);
    stream.getTracks().forEach(t => t.stop());

    return canvas.toDataURL('image/png');
  } catch {
    return null; // usuario rechazó el permiso → silencioso
  }
}
```

**Nota:** `getDisplayMedia` requiere que el usuario apruebe explícitamente la captura de pantalla. El checkbox "Incluir captura" solo está disponible si el usuario lo activa — nunca se captura sin consentimiento.

### `FeedbackButton.svelte`

```svelte
<!-- Botón flotante fijo, siempre visible -->
<button
  onclick={() => modalOpen = true}
  class="fixed bottom-6 right-6 z-40
         bg-[var(--accent)] text-white
         rounded-full px-4 py-2 text-sm font-medium shadow-lg
         hover:bg-[var(--accent-hover)] transition-colors
         flex items-center gap-2"
>
  <MessageSquare size={15} />
  Feedback
</button>

{#if modalOpen}
  <Modal onclose={() => modalOpen = false} title="Envianos tu feedback">
    <FeedbackModal onsubmit={() => modalOpen = false} />
  </Modal>
{/if}
```

- El botón vive en `/(app)/+layout.svelte` — visible en todas las páginas de la app
- `z-40` para estar sobre el contenido pero debajo de modales (`z-50`)

### Admin UI `/admin/feedback`

- Tabla con: fecha, tipo badge, rating (estrellas), usuario, URL, mensaje truncado
- Click en fila → detalle completo: mensaje, datos técnicos (viewport, UA, acciones recientes), screenshot si fue capturado
- Filtros: tipo, rating mínimo, dateFrom/dateTo
- Export CSV con todo el payload
- Stats: total feedback esta semana, distribución por tipo, rating promedio

### Integración con Fase 19

- Al cargar `autoData` en el feedback modal, incluir automáticamente `recentErrors`: últimos 5 errores de `error_logs` del mismo `user_id` en las últimas 2h
- Esto permite correlacionar "usuario reportó bug" con "había estos errores en el sistema en ese momento"
- Campo `recent_errors` en `feedback` tabla referencia a IDs de `error_logs`

### Tests (Fase 20)

| Test | Qué verifica |
|------|-------------|
| `ActionLogStore` | log 25 acciones → solo últimas 20 en buffer |
| `FeedbackModal` | submit sin mensaje → botón disabled; con mensaje → habilitado |
| `FeedbackModal` | autoData incluye url, userAgent, recentActions |
| BFF `POST /api/feedback` | Payload completo → gateway recibe con todos los campos |
| Admin `/admin/feedback` | Filtro type='bug' → solo bugs en tabla |
| `captureScreenshot` | Permiso denegado → retorna null sin throw |

---

## Fase 21 — Data Flywheel: Mejora Continua Automática

**Goal:** Cerrar el loop entre el feedback de usuarios y la calidad del sistema RAG. Cada interacción genera señales; el flywheel las procesa y surfacea acciones concretas para mejorar retrieval, chunks y parámetros. Se nutre de Fases 7, 20 y 22.

### El loop completo

```
Usuarios interactúan con el RAG
  ↓
Señales capturadas automáticamente (thumbs, reformulaciones, abandonos)
  ↓
Agregación diaria → DailyReport en SQLite
  ↓
Admin ve insights accionables en /admin/flywheel
  ↓
Admin aplica mejoras (re-ingestar, ajustar params, re-chunkar)
  ↓
Deploy via CI/CD (Fase 3) con evaluación de calidad automática
  ↓
Mejores respuestas → mejores señales → loop
```

### Señales de entrada

| Señal | Fuente | Significado |
|-------|--------|------------|
| Thumbs up/down | `message_feedback` (Fase 7) | Calidad directa de la respuesta |
| Follow-up inmediato | Chat sessions consecutivas | Respuesta incompleta |
| Reformulación de query | Sessions con queries similares sin satisfaction | Sistema no entendió |
| Abandono de sesión | Session sin reply del usuario post-respuesta | Respuesta inaceptable |
| Corrección humana | `human_feedback` aprobado (Fase 22) | Ground truth explícito |
| Feedback widget tipo 'bug' | `feedback` (Fase 20) | Problema sistémico |

### Métricas calculadas

```python
# saldivia/flywheel/metrics.py
class FlywheelMetrics:
    def compute_daily(self) -> DailyReport:
        return DailyReport(
            # Satisfacción por colección (thumbs up / total)
            collection_satisfaction=self._satisfaction_by_collection(),
            # Queries reformuladas 2+ veces → el sistema no pudo responder
            unanswered_queries=self._find_reformulation_chains(),
            # Chunks nunca usados por el reranker → candidatos a re-chunking
            unused_chunks=self._find_low_usage_chunks(),
            # Tendencia rolling 7 días
            satisfaction_trend=self._compute_trend(days=7),
            # Clusters de queries sin buena respuesta = brechas en el KB
            knowledge_gaps=self._cluster_unanswered_queries(),
        )
```

### Golden Dataset & Evaluación automática en CI/CD

**Golden Dataset** (`data/golden_qa.jsonl`): pares pregunta-respuesta validados por humanos.
- Se construye a partir de: respuestas con thumbs up + correcciones aprobadas de Fase 22
- Formato: `{"question": "...", "expected_answer": "...", "collection": "...", "sources": [...]}`
- Mínimo 20 pares para evaluar; objetivo a mediano plazo: 200+

**En el workflow de Fase 3**, se agrega un step antes del deploy:
```yaml
- name: RAG Quality Evaluation
  run: |
    ssh $BREV_USER@$BREV_HOST "
      cd ~/rag-saldivia
      uv run python -m saldivia.flywheel.evaluate \
        --golden-dataset data/golden_qa.jsonl \
        --min-faithfulness 0.75 \
        --min-relevancy 0.70 \
        --fail-on-regression 0.05
    "
```
Si las métricas caen más de 5% vs el deploy anterior → el CI falla → no se deploya.

### Tabla SQLite

```sql
CREATE TABLE IF NOT EXISTS rag_evaluations (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    run_at            TEXT    NOT NULL DEFAULT (datetime('now')),
    trigger           TEXT    NOT NULL CHECK(trigger IN ('deploy','manual','scheduled')),
    deploy_sha        TEXT,
    faithfulness      REAL,
    answer_relevancy  REAL,
    context_recall    REAL,
    context_precision REAL,
    avg_satisfaction  REAL,
    total_queries     INTEGER,
    unanswered_rate   REAL,
    passed            INTEGER NOT NULL DEFAULT 1,
    notes             TEXT
);
```

### Archivos

```
saldivia/flywheel/
├── metrics.py           — FlywheelMetrics: señales agregadas diarias
├── evaluate.py          — evaluación RAGAS contra golden dataset
├── gaps.py              — clustering de queries sin respuesta
└── scheduler.py         — APScheduler (BackgroundScheduler) iniciado en gateway startup_event;
                           job: compute_daily() a las 02:00 → persist en SQLite → notify admins vía SSE
                           Válido con --workers 1 (no hay race conditions entre workers)

data/
└── golden_qa.jsonl      — dataset de evaluación (versionado en git)

src/routes/(app)/admin/flywheel/
├── +page.svelte         — dashboard: score, tendencias, acciones sugeridas
└── +page.server.ts      — load: último reporte + historia
```

### Admin UI `/admin/flywheel`

- **Score general**: número 0-100% + badge verde/amarillo/rojo + trend 7d
- **Por colección**: satisfaction%, total queries, trend, botón "Tomar acción"
- **Brechas detectadas**: clusters de queries sin respuesta → botón "Sugerir ingesta" pre-rellena el upload
- **Chunks problemáticos**: documentos con chunks sin uso → botón "Re-ingestar"
- **Historia de evaluaciones RAGAS**: línea de tiempo por deploy, indicador de regresiones
- **Acciones priorizadas**: "Aumentar vdb_top_k en colección 'contratos' (satisfaction 45%)"

### Tests (Fase 21)

| Test | Qué verifica |
|------|-------------|
| `FlywheelMetrics` | 3 thumbs up, 7 down → satisfaction 30% |
| `gaps.py` | 5 queries similares sin satisfacción → 1 cluster |
| `evaluate.py` | 10 pares golden → métricas RAGAS entre 0-1 |
| CI step | Métricas < threshold → exit 1 (bloquea deploy) |
| Admin flywheel UI | Satisfaction < 50% → badge rojo |

---

## Fase 22 — Portal de Feedback Humano & Sugerencias

**Goal:** Canal estructurado para que usuarios y admins aporten feedback sobre el *contenido y calidad del RAG* — distinto del widget de Fase 20 que es sobre la app en sí. Alimenta directamente el flywheel (Fase 21) y el golden dataset.

### Tipos de feedback

| Tipo | Quién | Qué genera |
|------|-------|-----------|
| Corrección de respuesta | Cualquier usuario | Par (query, corrección) → golden dataset si admin aprueba |
| Documento desactualizado | Cualquier usuario | Flag en cola de revisión → re-ingesta si se aprueba |
| Sugerencia de documento | Cualquier usuario | Task de ingesta pre-poblada para admin |
| Valoración de colección | Area Manager | Señal de satisfacción para flywheel |
| Feature request | Cualquier usuario | Cola separada para el equipo |

### Flujo — corrección de respuesta

```
Usuario lee respuesta → click "Reportar error" (botón inline en mensaje)
  → Modal: "¿Qué está mal?" + "¿Cuál sería la respuesta correcta?" (opcional)
  → Guardado en human_feedback con type='correction', status='pending'
  → Admin ve en /admin/feedback-queue
  → Aprueba → par (query, corrección) → golden_qa.jsonl
  → Próximo CI/CD usa ese par como test case
```

### Flujo — sugerencia de documento

```
Usuario → botón "Sugerir documento" (visible en /feedback y en sidebar)
  → Form: tema + descripción + URL o archivo (opcional)
  → Guardado con type='doc_suggestion'
  → Admin en inbox → "Aceptar" → crea IngestionJob con trigger='human_suggestion'
  → Notifica al usuario cuando el doc fue ingestado (Fase 14 SSE)
```

### Schema SQLite

```sql
CREATE TABLE IF NOT EXISTS human_feedback (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    type            TEXT NOT NULL CHECK(type IN (
                      'correction','outdated_doc','doc_suggestion',
                      'collection_rating','feature_request'
                    )),
    status          TEXT NOT NULL DEFAULT 'pending'
                         CHECK(status IN ('pending','approved','rejected','in_progress')),
    user_id         INTEGER NOT NULL,
    session_id      TEXT,
    message_id      TEXT,         -- para correcciones de mensajes específicos
    collection      TEXT,
    source_doc      TEXT,         -- documento reportado como desactualizado
    original_text   TEXT,         -- respuesta original que se corrige
    corrected_text  TEXT,         -- corrección propuesta
    description     TEXT NOT NULL,
    url_suggestion  TEXT,
    rating          INTEGER CHECK(rating BETWEEN 1 AND 5),
    reviewed_by     INTEGER,
    reviewed_at     TEXT,
    review_notes    TEXT,
    created_at      TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_human_feedback_status ON human_feedback(status, type);
```

### Archivos

```
src/routes/(app)/
├── feedback/
│   ├── +page.svelte          — portal usuario: mis sugerencias + nueva
│   └── +page.server.ts       — load + actions: submit
└── admin/feedback-queue/
    ├── +page.svelte          — inbox admin: pendientes filtrados por tipo
    ├── +page.server.ts       — actions: approve, reject, escalate
    └── [id]/
        ├── +page.svelte      — detalle completo con contexto
        └── +page.server.ts   — approve con anotación → golden dataset

src/lib/components/feedback/
├── CorrectionButton.svelte   — botón inline en MessageList al hover
├── CorrectionModal.svelte    — form con contexto del mensaje
├── DocSuggestionModal.svelte — form de sugerencia de documento
├── FeedbackPortal.svelte     — "Mis sugerencias" para usuarios
├── FeedbackInbox.svelte      — cola admin con filtros y acciones bulk
└── FeedbackReviewCard.svelte — card expandible para revisar un item
```

### `CorrectionButton.svelte`

```svelte
<!-- inline en MessageList, visible al hover del mensaje -->
<button
  onclick={() => correctionOpen = true}
  class="opacity-0 group-hover:opacity-100 transition-opacity
         text-xs text-[var(--text-faint)] hover:text-[var(--warning)]
         flex items-center gap-1"
  aria-label="Reportar error en esta respuesta"
>
  <AlertTriangle size={12} />
  Reportar error
</button>
```

### Admin Feedback Inbox

```
Filtros: [Todos] [Correcciones ●8] [Docs sugeridos ●3] [Desactualizados] [Features]

┌──────────────────────────────────────────────────────────────────┐
│ ● Pendiente  Corrección  |  María G.  |  hace 2h                │
│ "La respuesta sobre el artículo 45 está incompleta..."           │
│ Colección: contratos  →  [Ver detalle]  [Aprobar]  [Rechazar]   │
├──────────────────────────────────────────────────────────────────┤
│ ● Pendiente  Doc sugerido  |  Carlos M.  |  hace 5h             │
│ "Falta info sobre licitaciones 2025"  |  URL adjunta            │
│          [Ver detalle]  [Crear ingesta]  [Rechazar]             │
└──────────────────────────────────────────────────────────────────┘
```

### Portal de usuario `/feedback`

- "Mis sugerencias": lista con estado (pendiente → en revisión → aprobado/rechazado)
- Notificación SSE (Fase 14) cuando su sugerencia fue procesada
- Contador motivacional: "Tu feedback mejoró 3 respuestas este mes"
- Botón "Nueva sugerencia" siempre visible, shortcut en Command Palette (Fase 17)

### Integración con Flywheel

- Correcciones aprobadas → `update_golden_dataset(feedback_id)` → `golden_qa.jsonl`
- Docs sugeridos aprobados → `IngestionJob(trigger='human_suggestion')`
- Colecciones con alta tasa de correcciones → flywheel las prioriza para revisión
- Métrica: "tasa de corrección por colección" = proxy de calidad del knowledge base

### Tests (Fase 22)

| Test | Qué verifica |
|------|-------------|
| CorrectionButton | Visible en hover de mensaje assistant; invisible en user messages |
| CorrectionModal | Submit vacío → disabled; con texto → habilitado; submit → toast |
| FeedbackInbox | Filtro 'correction' → solo correcciones; count badge correcto |
| BFF `POST /api/human-feedback` | Persiste con status='pending'; retorna id |
| Admin approve correction | Aprobado → `golden_qa.jsonl` tiene el nuevo par |
| Integración flywheel | 5 correcciones en colección X → flywheel.compute_daily() marca X como high priority |

---

## Out of scope (todas las fases)

- SSO / OAuth / SAML (requiere cambios en gateway auth core)
- Multi-tenant con isolation total de datos (requiere Milvus namespace por tenant)
- Chat en tiempo real entre usuarios (WebSockets bidireccional)
- Mobile apps nativas (iOS/Android)
- Export de conversación a PDF con formato (requiere servidor de renderizado)
- Citas inline numeradas `[1]` en texto (requiere cambios en el RAG pipeline del blueprint)

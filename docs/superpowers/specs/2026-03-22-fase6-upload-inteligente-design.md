# Fase 6 — Upload Inteligente: Design Doc

**Fecha:** 2026-03-22
**Estado:** Aprobado
**Contexto:** RAG Saldivia — overlay sobre NVIDIA RAG Blueprint v2.5.0

---

## Problema

El upload actual (Fase 3) es un HTTP request blocking con timeout de 600s. La barra de progreso es simulada (avanza hasta 85% y se congela). Si el usuario cierra la pestaña, el upload se cancela. No hay visibilidad real del estado de ingesta.

---

## Solución

Sistema de ingesta asíncrono con:
- Upload non-blocking: devuelve `job_id` inmediatamente
- Progreso real basado en datos del ingestor (extraction + indexing)
- Persistencia server-side de jobs en SQLite (recoverable al recargar)
- Tier classification por page count con polling adaptativo

---

## Decisiones de diseño

| Decisión | Elección | Razón |
|---|---|---|
| Modo upload | Non-blocking siempre | 600s blocking es inadmisible para el browser |
| Page count | `pypdf` en gateway | Tier conocido antes del primer poll, sin nueva infra |
| Progress | `extraction_completed×60% + documents_completed×40%` | Datos reales del ingestor, no simulación |
| Persistencia | SQLite `ingestion_jobs` en AuthDB | Sin infra nueva, recoverable entre recargas |
| Deadlock | Por tier (30/60/90/120s) | Tolerancia proporcional al tamaño del doc |

---

## 1. Capa de datos — `ingestion_jobs` en SQLite

Nueva tabla en `saldivia/auth/database.py`:

```sql
CREATE TABLE IF NOT EXISTS ingestion_jobs (
    id           TEXT PRIMARY KEY,
    user_id      INTEGER NOT NULL,
    task_id      TEXT NOT NULL,        -- task_id del ingestor NV-Ingest
    filename     TEXT NOT NULL,
    collection   TEXT NOT NULL,
    tier         TEXT NOT NULL,        -- tiny | small | medium | large
    page_count   INTEGER,              -- NULL para no-PDFs
    state        TEXT DEFAULT 'pending', -- pending | running | completed | failed | stalled
    progress     INTEGER DEFAULT 0,    -- 0-100, calculado por gateway
    created_at   TEXT NOT NULL,
    completed_at TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

Métodos nuevos en `AuthDB`:
- `create_ingestion_job(user_id, task_id, filename, collection, tier, page_count) → job_id`
- `get_ingestion_job(job_id) → dict | None`
- `get_active_ingestion_jobs(user_id) → list[dict]` — state IN ('pending','running','stalled')
- `update_ingestion_job(job_id, state, progress, completed_at=None)`

---

## 2. Gateway — cambios en `gateway.py`

### Nueva dependencia

```toml
# pyproject.toml
"pypdf>=4.0.0"
```

### Helper de clasificación (nuevo en gateway.py)

```python
def extract_page_count(file_bytes: bytes, filename: str) -> int | None:
    """Extrae page count de un PDF. Devuelve None para no-PDFs."""
    if not filename.lower().endswith('.pdf'):
        return None
    try:
        from pypdf import PdfReader
        import io
        reader = PdfReader(io.BytesIO(file_bytes))
        return len(reader.pages)
    except Exception:
        return None

def classify_tier(page_count: int | None, file_size: int) -> str:
    """Clasifica el tier por páginas (PDF) o tamaño (otros formatos)."""
    if page_count is not None:
        if page_count <= 20:  return "tiny"
        if page_count <= 80:  return "small"
        if page_count <= 250: return "medium"
        return "large"
    # Fallback por tamaño para no-PDFs
    if file_size < 100_000:   return "tiny"
    if file_size < 500_000:   return "small"
    if file_size < 5_000_000: return "medium"
    return "large"
```

### Modificar `POST /v1/documents`

```python
@app.post("/v1/documents")
async def ingest(request: Request, user: User = Depends(get_user_from_token)):
    # ... RBAC checks existentes ...

    form = await request.form()
    file = form.get("file")  # UploadFile
    data_str = form.get("data", "{}")
    data = json.loads(data_str)
    collection_name = data.get("collection_name", "")

    file_bytes = await file.read()
    page_count = extract_page_count(file_bytes, file.filename)
    tier = classify_tier(page_count, len(file_bytes))

    # Llamar al ingestor con blocking=False
    gw_form = FormData()
    gw_form.append("file", (file.filename, file_bytes, file.content_type))
    gw_form.append("data", json.dumps({**data, "blocking": False}))

    async with httpx.AsyncClient(timeout=30) as client:
        resp = await client.post(f"{INGESTOR_URL}/v1/documents", ...)

    ingestor_response = resp.json()
    task_id = ingestor_response.get("task_id")

    # Registrar job en SQLite
    job_id = db.create_ingestion_job(
        user_id=user.id,
        task_id=task_id,
        filename=file.filename,
        collection=collection_name,
        tier=tier,
        page_count=page_count,
    )

    db.log_action(user_id=user.id, action="ingest", ip_address=...)

    return {
        "job_id": job_id,
        "tier": tier,
        "page_count": page_count,
        "filename": file.filename,
    }
```

### Nuevo `GET /v1/jobs`

```python
@app.get("/v1/jobs")
async def list_jobs(user: User = Depends(get_user_from_token)):
    """Lista jobs activos del usuario (para recovery al recargar)."""
    if not user:
        raise HTTPException(status_code=401)
    jobs = db.get_active_ingestion_jobs(user.id)
    return {"jobs": jobs}
```

### Nuevo `GET /v1/jobs/{job_id}/status`

```python
@app.get("/v1/jobs/{job_id}/status")
async def job_status(job_id: str, user: User = Depends(get_user_from_token)):
    job = db.get_ingestion_job(job_id)
    if not job:
        raise HTTPException(status_code=404)
    if job["user_id"] != user.id:
        raise HTTPException(status_code=403)

    # Proxy al ingestor
    async with httpx.AsyncClient(timeout=10) as client:
        resp = await client.get(f"{INGESTOR_URL}/v1/status?task_id={job['task_id']}")
    result = resp.json()

    # Calcular progreso real
    state = result.get("state", "UNKNOWN")
    nv = result.get("nv_ingest_status", {})
    res = result.get("result", {})
    total = max(res.get("total_documents", 1), 1)
    extracted = nv.get("extraction_completed", 0)
    completed = res.get("documents_completed", 0)

    if state == "FINISHED":
        progress = 100
        new_state = "completed"
    elif state == "FAILED":
        progress = job["progress"]
        new_state = "failed"
    else:
        progress = int((extracted / total * 60) + (completed / total * 40))
        new_state = "running" if progress > 0 else "pending"

    # Actualizar SQLite
    completed_at = None
    if new_state in ("completed", "failed"):
        completed_at = datetime.now().isoformat()
    db.update_ingestion_job(job_id, new_state, progress, completed_at)

    return {
        "job_id": job_id,
        "state": new_state,
        "progress": progress,
        "tier": job["tier"],
        "page_count": job["page_count"],
        "filename": job["filename"],
        "collection": job["collection"],
        "created_at": job["created_at"],
    }
```

---

## 3. BFF — rutas SvelteKit

### Modificar `src/routes/api/upload/+server.ts`

Parsear la nueva respuesta del gateway: `{ job_id, tier, page_count, filename }` y devolverla al cliente.

### Nuevo `src/routes/api/ingestion/[jobId]/status/+server.ts`

```typescript
export const GET: RequestHandler = async ({ params, locals }) => {
    if (!locals.user) throw error(401);
    const resp = await fetch(`${GATEWAY_URL}/v1/jobs/${params.jobId}/status`, {
        headers: { Authorization: `Bearer ${locals.user.token}` }
    });
    const body = await resp.json();
    return json(body, { status: resp.status });
};
```

### Modificar `src/routes/(app)/upload/+page.server.ts`

```typescript
export const load: PageServerLoad = async ({ locals }) => {
    const [collectionsRes, jobsRes] = await Promise.all([
        gatewayListCollections(),
        gatewayListActiveJobs(locals.user),  // nuevo
    ]);
    return {
        collections: collectionsRes.collections,
        activeJobs: jobsRes.jobs,  // hidratados en el store al montar
    };
};
```

---

## 4. Frontend — tipos, store y poller

### `src/lib/ingestion/types.ts`

```typescript
interface TierConfig {
    label: string;
    color: 'green' | 'blue' | 'amber' | 'red';
    maxPages: number;
    pollInterval: number;
    deadlockThreshold: number;
    expectedMaxDuration: number;
    secondsPerPage: number;
}

export const TIER_CONFIG = {
    tiny: {
        label: 'Pequeño',   color: 'green',
        maxPages: 20,       pollInterval: 2,
        deadlockThreshold: 30, expectedMaxDuration: 45, secondsPerPage: 1.5,
    },
    small: {
        label: 'Mediano',   color: 'blue',
        maxPages: 80,       pollInterval: 3,
        deadlockThreshold: 60, expectedMaxDuration: 120, secondsPerPage: 2.0,
    },
    medium: {
        label: 'Grande',    color: 'amber',
        maxPages: 250,      pollInterval: 5,
        deadlockThreshold: 90, expectedMaxDuration: 480, secondsPerPage: 2.5,
    },
    large: {
        label: 'Muy grande', color: 'red',
        maxPages: Infinity, pollInterval: 10,
        deadlockThreshold: 120, expectedMaxDuration: 1800, secondsPerPage: 3.5,
    },
} as const satisfies Record<string, TierConfig>;

export type Tier = keyof typeof TIER_CONFIG;

export type JobState = 'pending' | 'running' | 'completed' | 'failed' | 'stalled';

export interface IngestionJob {
    jobId: string;
    filename: string;
    collection: string;
    tier: Tier;
    pageCount: number | null;
    state: JobState;
    progress: number;          // 0-100
    eta: number | null;        // segundos restantes estimados
    startedAt: number;         // Date.now() al crear
    lastProgressAt: number;    // para deadlock detection
}

export function classifyTier(pageCount: number): Tier {
    if (pageCount <= TIER_CONFIG.tiny.maxPages)   return 'tiny';
    if (pageCount <= TIER_CONFIG.small.maxPages)  return 'small';
    if (pageCount <= TIER_CONFIG.medium.maxPages) return 'medium';
    return 'large';
}

export function classifyTierBySize(bytes: number): Tier {
    if (bytes <   100_000) return 'tiny';
    if (bytes <   500_000) return 'small';
    if (bytes < 5_000_000) return 'medium';
    return 'large';
}

export function estimateETA(tier: Tier, progress: number, elapsedSeconds: number): number | null {
    if (progress >= 100) return 0;
    if (progress <= 0) return TIER_CONFIG[tier].expectedMaxDuration;
    const totalEstimated = (elapsedSeconds / progress) * 100;
    return Math.max(0, Math.round(totalEstimated - elapsedSeconds));
}
```

### `src/lib/ingestion/poller.ts`

```typescript
export class IngestPoller {
    private jobId: string;
    private tier: Tier;
    private lastProgress: number = -1;
    private lastProgressAt: number = Date.now();
    private stopped = false;

    constructor(jobId: string, tier: Tier) {
        this.jobId = jobId;
        this.tier = tier;
    }

    stop() { this.stopped = true; }

    async poll(onUpdate: (status: Partial<IngestionJob>) => void): Promise<void> {
        const config = TIER_CONFIG[this.tier];

        while (!this.stopped) {
            const resp = await fetch(`/api/ingestion/${this.jobId}/status`);
            if (!resp.ok) {
                onUpdate({ state: 'failed' });
                break;
            }

            const data = await resp.json();
            const now = Date.now();
            const elapsed = (now - this.lastProgressAt) / 1000;
            const eta = estimateETA(this.tier, data.progress, (now - startedAt) / 1000);

            // Deadlock detection
            if (data.progress === this.lastProgress) {
                if (elapsed > config.deadlockThreshold) {
                    onUpdate({ state: 'stalled', eta: null });
                    break;
                }
            } else {
                this.lastProgress = data.progress;
                this.lastProgressAt = now;
            }

            onUpdate({ state: data.state, progress: data.progress, eta });

            if (data.state === 'completed' || data.state === 'failed') break;

            await sleep(config.pollInterval * 1000);
        }
    }
}
```

### `src/lib/stores/ingestion.svelte.ts`

```typescript
// Store reactivo con SvelteKit 5 runes
let jobs = $state<IngestionJob[]>([]);

export const ingestionStore = {
    get jobs() { return jobs; },

    addJob(job: IngestionJob) {
        jobs = [...jobs, job];
    },

    updateJob(jobId: string, updates: Partial<IngestionJob>) {
        jobs = jobs.map(j => j.jobId === jobId ? { ...j, ...updates } : j);
    },

    removeJob(jobId: string) {
        jobs = jobs.filter(j => j.jobId !== jobId);
    },

    hydrateFromServer(serverJobs: ServerJob[]) {
        // Llamado en +page.svelte onMount con data.activeJobs
        // Solo hidrata jobs que el store no conoce todavía
        for (const sj of serverJobs) {
            if (!jobs.find(j => j.jobId === sj.job_id)) {
                this.addJob(serverJobToIngestionJob(sj));
            }
        }
    }
};
```

---

## 5. Componentes UI

### `TierBadge.svelte`
Badge pequeño con color y label del tier. Props: `tier: Tier`.

### `JobCard.svelte`
Card individual con:
- Nombre de archivo (truncado) + `TierBadge`
- Page count si disponible
- Progress bar real (transición CSS suave)
- Estado textual: "En cola" / "Extrayendo" / "Indexando" / "Completado" / "Error" / "Sin progreso"
- ETA en segundos cuando está disponible
- Botón "Reintentar" cuando `state = stalled | failed`

### `IngestionQueue.svelte`
Lista de jobs. Orden: activos primero (por progreso desc), luego completados (últimos 5), luego fallados.

### `DropZone.svelte`
Mejora del drag & drop de Fase 3. Agrega preview del tier estimado al seleccionar el archivo (client-side, por tamaño para no-PDF o indicando que se calculará al subir para PDF).

---

## 6. Flujo completo

```
1. Usuario elige archivo → DropZone muestra tier estimado (por tamaño)

2. Click "Subir":
   POST /api/upload (FormData: file + collection)
   → BFF → gateway
   → gateway: extrae page_count (pypdf), clasifica tier
   → gateway: POST ingestor blocking=false → { task_id }
   → gateway: INSERT ingestion_jobs → { job_id }
   ← { job_id, tier, page_count, filename }
   → ingestionStore.addJob(...)
   → new IngestPoller(job_id, tier).poll(...)

3. Cada pollInterval segundos:
   GET /api/ingestion/{job_id}/status
   → BFF → gateway
   → gateway: GET ingestor /v1/status?task_id=xxx
   → gateway: calcula progress = extraction×60% + indexing×40%
   → gateway: UPDATE ingestion_jobs
   ← { state, progress, tier, filename, ... }
   → ingestionStore.updateJob(...)
   → IngestPoller detecta deadlock si sin progreso > deadlockThreshold

4. state = 'completed':
   → progress = 100, toast "listo"
   → IngestPoller para
   → job queda en store como completado (visible 5 min, luego removeJob)

5. Al recargar la página:
   → +page.server.ts carga activeJobs del gateway
   → +page.svelte onMount: ingestionStore.hydrateFromServer(data.activeJobs)
   → Por cada job hidratado: new IngestPoller(...).poll(...)
   → Usuario ve jobs en progreso sin perder nada
```

---

## 7. Tests

| Test | Archivo | Qué verifica |
|---|---|---|
| tier classification | `types.test.ts` | 15p→tiny, 85p→small, 200p→medium, 300p→large |
| tier by size | `types.test.ts` | 50KB→tiny, 600KB→small, 10MB→large |
| estimateETA | `types.test.ts` | progress=50, elapsed=60 → eta≈60 |
| deadlock detection | `poller.test.ts` | Sin progreso >threshold → state=stalled |
| poller termina | `poller.test.ts` | state=completed → loop termina |
| store addJob | `ingestion.test.ts` | jobs.length++ |
| store updateJob | `ingestion.test.ts` | actualiza por jobId, no toca otros |
| store hydrateFromServer | `ingestion.test.ts` | no duplica jobs ya existentes |
| JobCard progress | `JobCard.test.ts` | progress=75 → barra al 75% |
| JobCard stalled | `JobCard.test.ts` | state=stalled → botón "Reintentar" visible |
| BFF status proxy | `upload.test.ts` | mock gateway → IngestionStatus correcto |
| gateway job ownership | `test_gateway.py` | otro user no puede ver job ajeno (403) |

---

## 8. Archivos modificados/creados

### Backend (`saldivia/`)
| Archivo | Cambio |
|---|---|
| `auth/database.py` | Nueva tabla `ingestion_jobs` + 4 métodos |
| `gateway.py` | Modificar POST /v1/documents + 2 endpoints nuevos |
| `pyproject.toml` | Agregar `pypdf>=4.0.0` |

### BFF (`services/sda-frontend/src/`)
| Archivo | Cambio |
|---|---|
| `routes/api/upload/+server.ts` | Parsear `{ job_id, tier, page_count }` |
| `routes/api/ingestion/[jobId]/status/+server.ts` | Nuevo |
| `routes/(app)/upload/+page.server.ts` | Cargar activeJobs en load() |
| `routes/(app)/upload/+page.svelte` | Reescribir completamente |
| `lib/server/gateway.ts` | Agregar `gatewayListActiveJobs()` |

### Frontend (`services/sda-frontend/src/lib/`)
| Archivo | Cambio |
|---|---|
| `ingestion/types.ts` | Nuevo |
| `ingestion/poller.ts` | Nuevo |
| `stores/ingestion.svelte.ts` | Nuevo |
| `components/upload/DropZone.svelte` | Nuevo |
| `components/upload/TierBadge.svelte` | Nuevo |
| `components/upload/JobCard.svelte` | Nuevo |
| `components/upload/IngestionQueue.svelte` | Nuevo |

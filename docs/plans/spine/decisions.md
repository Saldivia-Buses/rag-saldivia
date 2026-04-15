# Plan 26 — Fase 0: Decisiones Técnicas Cementadas

> **Estado:** approved — cierra Fase 0 de Plan 26.
> **Fecha:** 2026-04-15
> **Autoriza:** Enzo (decisión conjunta durante plan review)
> Este doc es el contrato técnico del Spine. Cualquier desviación posterior requiere ADR nuevo en `docs/architecture/decisions/`.

---

## D1. Spec format: CUE

**Decisión:** CUE v0.9.2+ como fuente de verdad para shapes de evento.

**Alternativas evaluadas:**
- **Protobuf** — rechazado. Requiere plugin per-lenguaje, wire format binario inutilizable en logs/NATS CLI. Overkill para JSON payload.
- **JSON Schema** — rechazado. Validación débil, unions torpes, sin tipos primitivos elegantes.
- **YAML + tipos Go manuales** — rechazado. Sin validación de shape, propenso a drift entre productor y consumer.

**Por qué CUE gana:** validación nativa, comentarios preservables, tipos primitivos + enums por disjunción, deriva a JSON Schema trivial, golang binary distributable.

**Tutorial:** `docs/conventions/cue.md`.

---

## D2. UUID v7 para event IDs

**Decisión:** `github.com/google/uuid` v1.6+ con `uuid.NewV7()`.

**Razón:** UUIDv7 es time-ordered (los primeros 48 bits son unix millis), lo que permite índices de B-tree eficientes en `event_outbox` y `processed_events` sin necesidad de columna `created_at` separada para ordering. Compatible con UUID v4 existente (mismo formato, mismo espacio).

**Descartado:** ULID (no standard library Go oficial, menos soporte), snowflake (requiere coord central).

---

## D3. Wire format: JSON

**Decisión:** JSON sin Protobuf como wire format del envelope.

**Razón:** el envelope **tiene que ser legible** en logs estructurados y en `nats sub tenant.>` sin schema. Binario no se puede debuggear a ojo. El overhead de JSON en NATS es aceptable (los payloads típicos son <2KB).

**Consecuencia:** cambio a Protobuf queda fuera de scope permanentemente. Si alguna vez se necesita binario por performance, sería un wrapper paralelo, no reemplazo.

---

## D4. Outbox y processed_events en tenant DB

**Decisión:**
- `event_outbox` vive en cada tenant DB (migration `055_event_outbox`).
- `processed_events` vive en cada tenant DB (migration `056_processed_events`).
- `dead_events` vive en platform DB (migration `010_dead_events`) — es operacional cross-tenant.

**Razón preserva invariante #1 (tenant isolation):** el `INSERT INTO event_outbox` está en la misma `tx` que el INSERT del mensaje/usuario/job. Esos viven en tenant DB. Outbox en platform requeriría 2PC (rechazado) o riesgo de write atómico solo a platform (rechazado).

**Idempotencia en tenant DB** por la misma razón: el handler ejecuta en `tx` de tenant DB, inserta `processed_events` en el mismo `tx`, commit atómico. Sin 2PC cross-DB = garantía exactly-once real.

**Consumers cross-tenant** (ninguno hoy) usarían `db/platform/migrations/010_platform_processed_events` (a crear cuando aparezca el caso).

---

## D5. DLQ supervisor en healthwatch, HA de 2 réplicas

**Decisión:** Spine-DLQ consumer vive en `services/healthwatch`, con 2 réplicas mínimo usando JetStream queue subscription (`DeliverGroup="healthwatch-dlq"`).

**Razón:**
- No crear microservicio nuevo para esto — encaja con el mandato self-healing de healthwatch.
- Queue subscription garantiza que si una réplica cae, la otra sigue procesando sin pérdida ni duplicación.
- JetStream `DLQ` stream retiene 30 días → zero-loss incluso si ambas réplicas caen.

**Consecuencia:** `docker-compose.prod.yml` y `k8s/healthwatch.yaml` se actualizan a 2 réplicas como parte de Fase 4.

**Si healthwatch crece demasiado:** escindir a `services/spine-dlq` separado. Umbral: si healthwatch añade >3 responsabilidades nuevas no relacionadas a observabilidad, reconsiderar.

---

## D6. Sanity check — traces se expresa limpio en envelope

**Objetivo:** verificar que los 3 eventos actuales de `services/traces` se pueden migrar a `Envelope[T]` sin pérdida de información. Si no, ajustar el envelope antes de cementarlo.

### Mapeo de `TraceStartEvent`

```go
// Hoy (services/traces/internal/service/traces.go:25)
type TraceStartEvent struct {
    TraceID   string  // → Envelope.ID (convertir a UUIDv7)
    TenantID  string  // → Envelope.TenantID (redundante)
    SessionID string  // → Payload.SessionID
    UserID    string  // → Payload.UserID
    Query     string  // → Payload.Query
}

// Envelope[TraceStartV1]
{
    id:             "<uuidv7>",
    tenant_id:      "saldivia",
    type:           "traces.start",
    schema_version: 1,
    occurred_at:    "<when query started>",
    recorded_at:    "<when published>",
    trace_id:       "<same as id for roots>",  // OK
    payload: {
        session_id: "...",
        user_id:    "...",
        query:      "..."
    }
}
```

**Observación:** `TraceID` actual es `string` (podría ser UUID libre), pero `Envelope.ID` es UUIDv7. Para migración, el trace genera `id = NewV7()` al inicio, y el `trace_id` del envelope apunta al mismo. Compatible.

### Mapeo de `TraceEndEvent`

```go
type TraceEndEvent struct {
    TraceID           string     // → Envelope.CorrelationID (link al start)
    TenantID          string     // → Envelope.TenantID
    Status            string     // → Payload.Status ("completed"|"failed")
    ModelsUsed        []string   // → Payload.ModelsUsed
    TotalDurationMS   int        // → Payload.TotalDurationMS
    TotalInputTokens  int        // → Payload.TotalInputTokens
    TotalOutputTokens int        // → Payload.TotalOutputTokens
    TotalCostUSD      float64    // → Payload.TotalCostUSD
    ToolCallCount     int        // → Payload.ToolCallCount
    Error             string     // → Payload.Error (omitempty)
}
```

**Observación:** fits 1:1. `TraceID` del end se mapea a `Envelope.CorrelationID` para linkear start ↔ end. El `Envelope.ID` del end es nuevo UUIDv7.

### Mapeo de `TraceEvent` (con `Seq`)

```go
type TraceEvent struct {
    TraceID    string          // → Envelope.CorrelationID
    TenantID   string          // → Envelope.TenantID
    Seq        int             // → Payload.Seq (intra-trace ordering)
    EventType  string          // → Payload.EventType ("llm_call"|"tool_call"|"error")
    Data       json.RawMessage // → Payload.Data (ver nota)
    DurationMS int             // → Payload.DurationMS
}
```

**Observación clave — `Data json.RawMessage`:** varía por `EventType`. En CUE se modela como disjunción tipada:

```cue
events: "traces.event": {
    version: 1
    payload: {
        trace_id:     string
        seq:          int
        duration_ms?: int
        event_type: "llm_call" | "tool_call" | "error"
        data: #LLMCallData | #ToolCallData | #ErrorData  // discriminated union
    }
}

#LLMCallData: { model: string, input_tokens: int, output_tokens: int, cost_usd: number }
#ToolCallData: { tool: string, args: {...}, result?: {...} }
#ErrorData: { message: string, stack?: string }
```

**Consecuencia:** el envelope soporta este shape sin modificar. La disjunción con `#` es la única feature CUE que habilita el caso Data-polimórfico — se documenta en `docs/conventions/cue.md` cuando aparezca el primer caso real (plan 28).

### Resultado del sanity check

**OK — envelope actual soporta los 3 eventos de traces sin modificaciones estructurales.** Los campos `TraceID` externo se mapean vía `CorrelationID` (start ↔ end ↔ events). `Seq` queda en payload (intra-trace ordering, no es responsabilidad del envelope). `parent_span_id`, si aparece alguna vez, va a `CausationID` (ya `*uuid.UUID` en el shape).

**Sin cambios al `Envelope[T]` propuesto.**

---

## D7. Traces freeze a bug fixes — enforced, no honor system

**Decisión:** `services/traces/internal/` queda congelado a bug fixes hasta Plan 28 (migración a spine).

**Mecanismos activos (commit `9892913a`):**
1. `.github/CODEOWNERS` → `/services/traces/internal/ @Camionerou` con comentario explícito del freeze.
2. `.github/pull_request_template.md` → checkbox requiere `traces-exception` label para funcionalidad nueva.
3. Plan 28 pendiente de escritura — cubrirá migración a `spine.Consume[TraceStartV1]` y reducción de boilerplate.

**Qué cuenta como bug fix:** corregir un crash, fix de dato incorrecto, mejora de performance sin nueva funcionalidad. Aditivo a features = exception.

---

## D8. Tutorial CUE como entregable Fase 0

**Decisión:** `docs/conventions/cue.md` escrito antes de Fase 1 (commit `37677413`).

**Contenido:** install + checksum, minimal syntax, 4 Type examples (chat, ingest, auth, platform), regen workflow, breaking change checklist. 199 líneas, bajo el límite modular de 200.

**Si CUE friccional post-Fase 1:** reevaluar migración a JSON Schema. Decisión reevaluable — no atada al diseño del envelope.

---

## Pendientes (no bloquean Fase 1)

- Escribir `scripts/setup-dev.sh` con `cue` download + checksum (Fase 1).
- Agregar `actions/cache` para `~/.cache/cue` en `.github/workflows/ci.yml` (Fase 1).
- Escribir Plan 28 (migración traces a spine). Placeholder mencionado, no redactado.

---

## Firma del Fase 0

- [x] D1. CUE decidido, tutorial escrito
- [x] D2. UUID v7 fijado
- [x] D3. JSON wire format fijado
- [x] D4. Outbox + processed_events en tenant DB, dead_events en platform
- [x] D5. DLQ en healthwatch HA (2 réplicas queue sub)
- [x] D6. Sanity check traces — envelope sin modificaciones
- [x] D7. Traces freeze enforced via CODEOWNERS + PR template
- [x] D8. Tutorial CUE entregado

**Fase 0 cerrada. Proceder a Fase 1.**

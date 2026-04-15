---
title: Convention: CUE Specs for Events
audience: ai
last_reviewed: 2026-04-15
related:
  - ../plans/2.0.x-plan26-spine.md
---

# CUE para specs de eventos

> Audiencia: cualquiera que agregue o modifique un tipo de evento del spine. Plan 26 usa CUE como fuente de verdad para envelopes. `make events-gen` produce Go, TypeScript y docs a partir de `pkg/events/spec/*.cue`.

## Por qué CUE

Necesitamos tipos compartidos entre backend Go y frontend TS, con validación de shape y comentarios semánticos. Alternativas evaluadas:

| Opción | Pros | Contras |
|---|---|---|
| **CUE** ✅ | validación nativa, tipos primitivos + unions, comentarios, deriva JSON Schema | curva inicial, binary extra |
| Protobuf | maduro, tooling | wire format binario, overkill para JSON, requiere `.proto` + plugin por lenguaje |
| JSON Schema | standard, tooling web | validación débil, sin tipos union elegantes |
| YAML + code | simple | sin validación de tipos, errores al regenerar |

CUE gana porque el envelope es JSON legible y queremos garantías de shape.

## Instalación

```bash
# macOS
brew install cue-lang/tap/cue

# Linux (scripts/setup-dev.sh lo hace con checksum)
curl -L https://github.com/cue-lang/cue/releases/download/v0.9.2/cue_v0.9.2_linux_amd64.tar.gz -o /tmp/cue.tgz
echo "<SHA256>  /tmp/cue.tgz" | sha256sum -c -
tar -xzf /tmp/cue.tgz -C /usr/local/bin cue
cue version  # v0.9.2
```

CI usa `actions/cache` sobre `~/.cache/cue`. Agregar nueva versión implica actualizar `scripts/setup-dev.sh` + el paso de cache en `.github/workflows/ci.yml`.

---

## Estructura de una spec

Cada familia de eventos vive en `pkg/events/spec/<family>.cue`. Un archivo tiene un struct `events` con una entry por `Type`:

```cue
// pkg/events/spec/notify.cue
package events

events: "chat.new_message": {
    // Versión del schema. Monotónica uint8. Bumpear ante breaking change.
    version: 1

    // Shape del subject NATS. "{slug}" se reemplaza con el tenant slug en runtime.
    subject: "tenant.{slug}.notify.chat.new_message"

    // Shape del payload (lo que va dentro de Envelope.payload).
    payload: {
        user_id:    string  // quien recibe el badge
        session_id: string
        message_id: string
        title:      string
        body:       string
        channel:    "in_app" | "email" | "both"  // enum por union de strings literales
    }

    // Productores y consumers esperados. El codegen lo usa para generar docs,
    // no para validar en runtime (el lint SÍ valida en CI).
    publishers: ["chat"]
    consumers:  ["notification", "ws"]
}
```

### Sintaxis mínima que usamos

| Feature | Ejemplo |
|---|---|
| Primitivos | `string`, `int`, `bool`, `number` |
| Enum | `"in_app" \| "email" \| "both"` (union de literales) |
| Opcional / Default | `correlation_id?: string` / `channel: *"in_app" \| "email"` |
| Lista homogénea | `[...string]` |
| Comentario | `// ...` (preservado como docstring Go/JSDoc) |

Specs planas, un archivo por familia. No usamos comprehensions ni múltiples módulos CUE. Templates (`#Def`) quedan **reservados para Plan 28** cuando aparezcan payloads polimórficos (ej. `traces.event.Data` con disjunción `#LLMCallData | #ToolCallData | #ErrorData`). En Plan 26 no se usan.

---

## Ejemplos reales (los 4 initial Types de Plan 26)

### 1. `chat.new_message`

```cue
events: "chat.new_message": {
    version: 1
    subject: "tenant.{slug}.notify.chat.new_message"
    payload: {
        user_id:    string
        session_id: string
        message_id: string
        title:      string
        body:       string
        channel:    "in_app" | "email" | "both"
    }
    publishers: ["chat"]
    consumers:  ["notification", "ws"]
}
```

### 2. `ingest.completed`

```cue
events: "ingest.completed": {
    version: 1
    subject: "tenant.{slug}.ingest.completed"
    payload: {
        job_id:          string
        collection_name: string
        doc_count:       int
        chunk_count:     int
        duration_ms:     int
    }
    publishers: ["ingest"]
    consumers:  ["notification", "ws"]
}
```

### 3. `auth.login_success`

```cue
events: "auth.login_success": {
    version: 1
    subject: "tenant.{slug}.auth.login_success"
    payload: { user_id: string, email: string, ip_address: string, user_agent: string }
    publishers: ["auth"]
    consumers:  ["notification"]
}
```

### 4. `platform.lifecycle` (subject NO-tenant, platform-wide)

```cue
events: "platform.lifecycle": {
    version: 1
    subject: "platform.lifecycle.{action}"
    payload: {
        action:      "tenant_created" | "tenant_deleted" | "tenant_suspended"
        tenant_id:   string
        tenant_slug: string
        by_user_id:  string
    }
    publishers: ["platform"]
    consumers:  ["auth", "chat", "ingest", "healthwatch", "ws"]  // DrainerRegistry + ws subs
}
```

**Nota:** `action` usa underscores (`tenant_created`), no dots. La razón es que `{action}` se interpola en el subject y `pkg/spine.BuildSubject` valida cada segmento con `^[a-zA-Z0-9_-]+$`. Si necesitás subjects jerárquicos con dots, el template debe tenerlos literales, no vía substitución (ej. `platform.lifecycle.tenant.{verb}` con `verb: "created"`).

---

## Regenerar después de editar una spec

```bash
make events-gen    # produce pkg/events/gen/*.go, apps/web/src/lib/events/gen/*.ts, docs/events/*.md
make events-validate  # verifica que lo commiteado == lo regenerado (CI lo corre)
```

Si `events-validate` falla en CI, corré `make events-gen` local, commiteá los archivos generados.

---

## Checklist: "cambié una spec, ¿qué hago?"

### Es aditivo (campo opcional nuevo con default sano)
1. Agregar el campo a `pkg/events/spec/<family>.cue`.
2. Dejar `version` igual.
3. `make events-gen`.
4. Commitear spec + generados.
5. Productor y consumers se actualizan cuando quieran — consumers viejos ignoran el campo sin error.

### Es breaking (rename, remove, type change, optional→required, semantics change)
1. **Copiar** la spec actual a `<type>_v1.cue` (ej. `chat.new_message_v1.cue`).
2. **Bump** `version: 2` en la spec principal.
3. **Modificar** la spec con los cambios breaking.
4. `make events-gen` → ahora hay `Envelope[ChatNewMessageV1]` y `Envelope[ChatNewMessageV2]`.
5. Productor migra a V2 inmediatamente. Opcional: dual-publish V1+V2 hasta que consumers migren.
6. Consumers migran uno a uno.
7. Cuando `spine_consume_total{schema_version="1", type="chat.new_message"} == 0` por ≥14 días → borrar `chat.new_message_v1.cue`.

### Regla dura (enforceada por lint)
**Nunca cambiar el shape del payload manteniendo el mismo `version`.** El lint de CUE en CI compara `pkg/events/spec/*.cue` contra `origin/2.0.5` y falla si un Type tiene diff estructural con mismo version.

---

## Referencias

- Spec oficial: https://cuelang.org/docs/
- Tutorial interactivo: https://cuelang.org/play/
- Plan 26: `docs/plans/2.0.x-plan26-spine.md`
- Editor: VSCode `cue.cue`, Vim `cue-lang/cue`, JetBrains plugin CUE

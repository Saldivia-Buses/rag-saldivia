---
name: subagent-driven-development
description: "Dispatch fresh subagent per task with TDD + two-stage review. Use when executing plan tasks."
user_invocable: true
---

# Subagent-Driven Development — SDA Framework

Cuando ejecutás tareas de plan, despachá un subagente fresco por cada tarea independiente.

## Protocol

### Step 1: Task Briefing

Cada subagente recibe un prompt auto-contenido con:
- **Goal**: Qué implementar (archivos específicos, funciones, endpoints)
- **Context**: Qué existe, qué patterns seguir, qué `pkg/` usar
- **Constraints**: Tenant isolation, auth middleware, NATS naming, convenciones sqlc
- **Scope**: Máximo 3 archivos por agente
- **TDD requirement**: Test primero, implementación después
- **Commit policy**: Commitear después de cada archivo que compila y pasa tests
- **Verification**: Qué comandos correr (`make test`, `make lint`, `go vet`)

### Step 2: Dispatch

```
Agent({
  description: "Implement [task name]",
  prompt: "[briefing auto-contenido con file paths, patterns, constraints, instrucción de commit]",
  subagent_type: "general-purpose"  // o tipo especializado
})
```

### Step 3: Política de commits incrementales (OBLIGATORIA)

**Todo prompt de subagente que escribe código debe incluir:**

```
REGLA: después de cada archivo que compila (`go build ./...` sin errores),
commitear inmediatamente:
  git add [archivo_modificado]
  git commit -m "tipo(scope): descripción"

NO acumules cambios en múltiples archivos antes de commitear.
Si el contexto se corta sin commit, se pierde TODO el trabajo.
```

### Step 4: Two-Stage Review

Cuando el subagente retorna:

**Stage 1 — Spec Review:**
- ¿La implementación matchea el spec/plan?
- ¿Todos los requisitos están atendidos?
- ¿Se respeta tenant isolation?

**Stage 2 — Code Quality Review:**
- ¿Sigue convenciones Go (chi, sqlc, slog)?
- ¿Error paths manejados con wrapped errors?
- ¿Context propagado correctamente?
- ¿Tests con table-driven patterns?

### Step 5: Persistencia de conocimiento

**Antes de marcar una tarea como completa**, preguntarse:
- ¿El agente encontró algo no obvio? (comportamiento de librería, pattern nuevo, bug)
- ¿Ese conocimiento está en el código (como comentario) o en docs/decisions/?
- ¿Debería ir a MEMORY.md para futuras sesiones?

Si sí → guardar antes de seguir:
```bash
# Opción: actualizar decision records
update_decision_records(action="create", title="...", decision="...", rationale="...")

# Opción: guardar en MEMORY.md de la sesión principal
# (la conversación principal hace esto, no el subagente)
```

**Ejemplos de cosas que DEBEN persistirse:**
- `pgtype.Timestamptz.Scan()` no acepta RFC3339 en Go 1.25 — necesita formato PostgreSQL
- `go test github.com/...` vía module path usa caché viejo en workspace mode
- `make test ./services/...` rompe con go.work — usar `cd services/X && go test ./...`
- La interfaz IngestService cambió y el mock en ingest_test.go estaba desactualizado

### Step 6: Integration

Después de todos los subagentes:
1. `make test` — suite completa
2. `make lint`
3. `make build`
4. Verificar conflictos entre outputs
5. Verificar flujo NATS entre servicios si aplica

## Agentes Especializados Disponibles

| Agente | Usar para |
|-------|---------|
| `gateway-reviewer` | Revisar Go handlers, middleware, auth, NATS |
| `frontend-reviewer` | Revisar React components, hooks, auth |
| `security-auditor` | Security audit antes de release |
| `test-writer` | Escribir Go tests + frontend tests |
| `debugger` | Diagnosticar failures |

## Anti-patterns
- Dispatching sin contexto auto-contenido
- Saltarse el two-stage review
- Subagentes que modifican los mismos archivos (conflictos)
- NO incluir instrucción de commit incremental en el prompt
- No correr integration tests después de que todos los subagentes completan
- No persistir knowledge no-obvio a docs/decisions/ o MEMORY.md

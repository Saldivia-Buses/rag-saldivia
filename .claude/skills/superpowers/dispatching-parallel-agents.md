---
name: dispatching-parallel-agents
description: "Run 2+ independent tasks simultaneously via parallel agents. Use when plan has tasks with no dependencies between them."
user_invocable: true
---

# Dispatching Parallel Agents — SDA Framework

When a plan has 2+ tasks that are independent (different services, different files), run them in parallel.

## Protocol

### Step 1: Identify Parallelizable Tasks
Tasks are parallelizable when:
- They modify different services
- They modify different files within the same service
- They have no data dependency (output of one is not input of another)
- They don't share database migrations

Tasks are NOT parallelizable when:
- One task creates a `pkg/` package that another task imports
- One task writes a migration that another task's sqlc queries depend on
- Both tasks modify the same handler or service file

### Step 2: Size Each Task Correctly

**REGLA: máximo 3 archivos por agente para tareas de escritura.**

Si el scope es mayor (ej: 10 handlers que testear):
- Dividilo en lotes de 3
- Cada lote es un agente separado
- Los lotes pueden correr en paralelo si no tienen conflictos de archivo

Por qué: los agentes se cortan por contexto. Si tienen 10 archivos y se cortan después del 7mo sin commitear, se pierden los 7 que hicieron.

### Step 3: Decidir si usar worktree

**Usar `isolation: "worktree"` cuando:**
- El cambio es un refactor arriesgado (cambia API pública, firma de funciones)
- La tarea puede romper el build si queda a medias
- Querés probar algo experimental sin tocar main

**NO usar `isolation: "worktree"` cuando:**
- La tarea es puramente aditiva (agregar test files, nueva feature sin tocar existente)
- Los archivos de distintos agentes no se solapan
- La tarea es pequeña y commitea inmediatamente

Para tests: **no usar worktrees**. Los test files son aditivos. El merge overhead no vale la pena.

### Step 4: Incluir instrucciones de commit en cada prompt

Todo agente que escribe código DEBE tener en su prompt:
```
Regla: commitear después de cada archivo que pase `go build ./...`.
No acumules trabajo sin commitear — si el contexto se corta, se pierde.
```

### Step 5: Dispatch en mensaje único
Todos los agentes paralelos en un solo bloque:

```
Agent({ description: "Task A", prompt: "..." })
Agent({ description: "Task B", prompt: "..." })
Agent({ description: "Task C", prompt: "..." })
```

### Step 6: Recolección e integración

Cuando los agentes retornan:
1. Verificar que commitaron (git log --oneline -5)
2. Si un agente NO commitó pero hizo cambios → están en el repo principal como unstaged
3. `git diff --stat HEAD` para ver qué cambió
4. `make test && make lint && make build` — verificar integración
5. Resolver conflictos si los hay

## Recuperar trabajo de un agente fallido

Si un agente termina sin commitear:
```bash
git status --short          # ver archivos modificados/nuevos
git diff --stat HEAD        # ver cuánto cambió
go build ./...              # ¿compila?
go test [paquete] -count=1  # ¿los tests pasan?
```

Si compila y los tests pasan → commitear manualmente.
Si hay errores → el agente dejó trabajo a medias, decidir si arreglarlo o descartar.

## SDA Parallel Patterns

| Pattern | Agent A | Agent B |
|---------|---------|---------|
| Service + Frontend | Go handler/service | React component/hook |
| Multi-service | auth changes | chat changes |
| Backend + Tests | Implementation | Test writer |
| Code + Docs | Implementation | doc-writer agent |
| Tests por lotes | handlers A-C | handlers D-F |

## Anti-patterns
- Dispatching agents con scope de 6+ archivos sin dividir
- Usar worktrees para trabajo aditivo (overhead innecesario)
- NO incluir instrucción de commit incremental en el prompt
- Dispatching agents que editarán el mismo archivo
- No revisar `git status` después de que un agente termina
- No correr tests de integración después de que todos los agentes completan

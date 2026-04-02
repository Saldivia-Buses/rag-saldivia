# 15 — Agentes y Workflow

## Modelo de trabajo

```
Enzo ←→ Claude Opus (Claude Code CLI)
              ↓
         Opus planifica → Enzo aprueba
              ↓
         Opus ejecuta → quality gates → commit
              ↓
         Repeat por fase
```

**Opus hace TODO:** planifica, ejecuta codigo, revisa, testea, documenta.
Un solo modelo con contexto completo. Sin delegacion, sin telefono descompuesto.

---

## 10 Agents especializados

Todos son **perfiles especializados de Opus** — mismo modelo, distinta checklist.
Definidos en `.claude/agents/`.

| Agent | Cuando se usa | Que verifica |
|-------|--------------|-------------|
| `frontend-reviewer` | Despues de cambios en componentes/UI | Accesibilidad, design system, testing, performance |
| `gateway-reviewer` | Despues de cambios en API routes/auth | Seguridad, error handling, validacion, RBAC |
| `security-auditor` | Antes de releases | Auth, RBAC, API keys, SQL injection, XSS, CSRF |
| `test-writer` | Cuando se necesitan tests nuevos | Cobertura, convenciones, happy-dom, patron ADR-007 |
| `debugger` | Cuando algo no funciona | Repro steps, logs, error analysis, diff |
| `doc-writer` | Cuando hay que actualizar docs | CHANGELOG, ADRs, planes, CLAUDE.md |
| `deploy` | Cuando se deploya | Build, tests, health checks, rollback plan |
| `status` | Para ver estado de servicios | Database, Redis, RAG, WebSocket, health endpoints |
| `plan-writer` | Para escribir planes nuevos | Template, prerequisitos, fases, entregables |
| `ingest` | Para ingestar documentos | Queue BullMQ, Milvus, metadata, formato |

---

## Workflow OODA-SQ

```
OBSERVE → ORIENT → DECIDE → ACT
                              └── Implement → Simplify → Review → Docs
```

### Fases y gates

| Fase | Gate | Que se hace | Skills/Tools |
|------|------|------------|-------------|
| OBSERVE | 0 | Explorar codebase, entender estado actual | firecrawl, CodeGraphContext, repomix, Explore agent |
| ORIENT | 1 | Brainstorming, explorar opciones | brainstorming skill |
| DECIDE | 2 | Escribir plan de implementacion | writing-plans skill |
| IMPLEMENT | 3 | TDD, ejecutar plan, debugging | subagent-driven-development, TDD, parallel agents |
| SIMPLIFY | 4 | Limpiar, eliminar dead code | simplify skill, CodeGraphContext find_dead_code |
| REVIEW | 5 | Code review automatico | gateway-reviewer, frontend-reviewer agents |
| DOCS | 6 | Documentar cambios | doc-writer agent, changelog-generator, CLAUDE.md |

### Cambios triviales (≤3 lineas)
Solo GATE 3 + 4 + 6. No invocar brainstorming ni writing-plans.

---

## Sprint sequence

```
THINK   → Cuestionar scope. Es necesario? Scope minimo? Ya existe solucion al 80%?
PLAN    → Plan detallado en docs/plans/. Enzo aprueba.
EXECUTE → Fase por fase. Contexto completo.
REVIEW  → 2 pasadas: correcto/compila + scope drift/security
TEST    → tsc → test → lint (por fase)
SHIP    → Docs + CHANGELOG. Commit. Tag si es release.
```

---

## Quality gates

### Despues de CADA fase (obligatorio)
```bash
bunx tsc --noEmit          # tipos — 0 errores
bun run test               # unit tests — todos pasan
bun run lint               # lint — limpio
```

### Despues de CADA plan completado
```bash
bun run test:components    # component tests con happy-dom
bun run test:visual        # visual regression (si cambios UI)
bun run test:a11y          # accesibilidad (si cambios UI)
```

### Antes de un release
```bash
bun run test:e2e           # E2E Playwright
# Security audit (agent security-auditor)
# Review completo de CLAUDE.md
```

**Si un gate falla:** la fase NO se commitea. Se arregla primero.

---

## Intensity modes

| Modo | Cuando | Que se hace |
|------|--------|------------|
| **Quick** | 1-3 archivos, docs, config | Opus revisa directo, tsc + tests |
| **Standard** | Fases normales de un plan | Review 2 pasadas, tests completos, scope drift |
| **Thorough** | Releases, seguridad, refactors grandes | Todo standard + security audit + review exhaustivo |

---

## Protocolo de escalacion

### Opus escala a Enzo cuando:
- Scope drift (archivos fuera del plan)
- > 3 fixes no planeados (WTF likelihood)
- Decision de producto (naming, UX, features)
- Algo podria romper funcionalidad existente
- Cambio de dependencia (agregar/quitar paquete)
- Cualquier operacion de guard rules

### Opus NO escala para:
- Cambios rutinarios dentro del scope
- Fixes de tests del mismo cambio
- Ajustes de imports, formatting

---

## Recovery de errores

| Situacion | Protocolo |
|-----------|-----------|
| Tests fallan post-commit | Fix forward (nuevo commit). Nunca revert. |
| Build no compila | Diagnosticar, arreglar. Si no → escalar. |
| Scope drift | SIEMPRE parar. Tolerancia cero. |
| Resultado inesperado | Diagnosticar, reintentar (max 2). Si no → escalar. |
| Conflicto de merge | Conservar cambios recientes. Ambiguedad → escalar. |

---

## Supervision

Despues de cada fase, Opus reporta:

```
✅ Plan N, Fase N completada

Cambios:
- archivo1.ts (nuevo)
- archivo2.ts (modificado)

Tests: X passed, 0 failed
Types: 0 errors
Lint: clean

Sigo con la Fase N+1 o queres revisar algo?
```

**Respuestas de Enzo:**
- "dale" → continuar
- "mostra diff" → mostrar cambios
- "para" → detener
- "volve atras" → revertir fase

---

## Loops de mejora continua

Enzo puede pedir loops autonomos entre planes:

| Loop | Metrica | Comando |
|------|---------|---------|
| Code quality | tsc errors, lint warnings → 0 | "mejora la calidad" |
| Test coverage | % coverage ↑ | "subi la cobertura" |
| Dead code | Lineas eliminadas | "limpia dead code" |
| Performance | Segundos de build ↓ | "optimiza performance" |
| Dependencies | Deps outdated → 0 | "actualiza dependencias" |
| Security | Findings → 0 | "revisa seguridad" |
| Complexity | Cyclomatic complexity ↓ | "reduci complejidad" |

Cada iteracion es atomica: un cambio, un test, un commit o revert. Max 10 iteraciones.

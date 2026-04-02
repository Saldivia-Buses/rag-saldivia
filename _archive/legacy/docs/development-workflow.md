# Development Workflow — OODA-SQ + Mission Brief

## Flujo completo de una sesión

```
[INICIO DE SESIÓN]
      ↓
Mission Brief ──► INTEL + SITUATION + MISSION + EXECUTION
      ↓
OODA-SQ por cada feature/task
      ↓
[CUANDO ENZO LO PIDE]
      ↓
Actualizar Roadmap (docs/roadmap.md)
```

Ver también:
- `docs/roadmap.md` — visión macro del proyecto
- `docs/sessions/` — briefs históricos por sesión

---

## Mission Brief (inicio de sesión)

Ejecutar **antes de tocar código**, al arrancar cada sesión.

### Estructura

```
🔭 INTEL     → firecrawl search sobre tecnologías relevantes a lo que haremos
📊 SITUATION → tests + bugs + deuda técnica + servicios + últimos commits
🎯 MISSION   → qué hacer hoy y por qué (priorizado, basado en roadmap)
⚡ EXECUTION → tasks que entran directo al OODA-SQ
```

### Tools

| Sección | Tool |
|---------|------|
| INTEL | `firecrawl search "[tech] best practices 2026"` |
| SITUATION — tests | `uv run pytest saldivia/tests/ -v` |
| SITUATION — bugs | `mcp__github__list_issues` |
| SITUATION — deuda | `mcp__CodeGraphContext__find_dead_code` + `find_most_complex_functions` |
| SITUATION — servicios | subagente `status` |
| SITUATION — historial | `git log --oneline -10` |
| MISSION | análisis de SITUATION + `docs/roadmap.md` + memoria del proyecto |

### Output

`docs/sessions/YYYY-MM-DD-brief.md`

```markdown
# Session Brief — YYYY-MM-DD

## 🔭 INTEL
[Estado del arte relevante para lo que vamos a hacer hoy]

## 📊 SITUATION
- **Tests:** X/Y pasan
- **Bugs abiertos:** N issues
- **Deuda técnica:** [funciones complejas, dead code]
- **Servicios:** [estado]
- **Últimos cambios:** [resumen git log]

## 🎯 MISSION
1. [Prioridad 1] — porque [razón]
2. [Prioridad 2] — porque [razón]

## ⚡ EXECUTION
- [ ] Task 1 — success criteria
- [ ] Task 2 — success criteria
```

**Nota:** El INTEL del Brief **es** el OBSERVE del primer OODA-SQ. No repetir el firecrawl.

---

## La Regla Fundamental

Todo cambio no trivial sigue el **OODA-SQ loop**:

```
OBSERVE ──► ORIENT ──► DECIDE ──► ACT
   ▲                              │
   └──────── siguiente iteración ─┘
```

**ACT** tiene 4 sub-fases en orden obligatorio:
```
Implement ──► Simplify ──► Review ──► Docs Sync
```

Este proceso no es opcional. Saltear fases produce features mal diseñadas, bugs en producción y deuda técnica.

---

## Trivial vs No-trivial

### Trivial (implementar directo)

- Typo en código o documentación
- Cambiar un valor de config
- Agregar un log statement
- Renombrar una variable (1-3 ocurrencias)
- Formateo sin cambio de lógica

**Regla:** Si el cambio toca ≤3 líneas y el fix es obvio → GATE 3 + 4 + 6 solamente.

### No-trivial (OODA-SQ completo)

- Feature nueva (endpoint, componente UI, comando CLI)
- Bug que toca >3 archivos o cambia lógica en >1 función
- Refactor (extraer función, partir archivo, cambiar estructuras de datos)
- Nueva dependencia (paquete npm, librería Python, servicio Docker)
- Optimización de rendimiento (cache, batch, paralelización)
- Cambio de comportamiento en features existentes (aunque sea 1 línea)

**Regla:** Si necesitás pensar cómo implementarlo → OODA-SQ completo.

---

## FASE 0 — OBSERVE [GATE 0]

**Cuándo:** Al arrancar la sesión + antes de cada feature nueva.

**Objetivo:** Estado del arte externo + contexto interno del codebase.

### Tools

| Tool | Cuándo usarlo |
|------|--------------|
| `firecrawl search "tema 2026"` | Estado del arte, mejores prácticas actuales |
| `firecrawl scrape "url"` | Documentación específica de una librería |
| `mcp__CodeGraphContext__find_code` | Buscar símbolos, funciones, clases por nombre |
| `mcp__CodeGraphContext__analyze_code_relationships` | Callers, callees, imports, exports |
| `mcp__repomix__pack_codebase` | Empaquetar módulo completo para análisis AI |
| Subagente `Explore` | Exploración profunda cuando Glob/Grep no alcanzan |

### Pregunta clave
> *¿Qué está haciendo el mundo con esto? ¿Estamos alineados?*

### Gate 0
- [ ] `firecrawl search` ejecutado sobre el tema
- [ ] Resultado documentado: "¿qué hace el mundo?"
- [ ] Codebase relevante mapeado (CGC / Repomix)

---

## FASE 1 — ORIENT [GATE 1]

**Objetivo:** Spec con diseño comparativo. Sección "Nosotros vs Estado del Arte" obligatoria.

### Tools

| Tool | Cuándo usarlo |
|------|--------------|
| Skill `superpowers:brainstorming` | Toda feature no trivial — genera spec con opciones |
| Subagente `Plan` | Decisiones de arquitectura complejas |

### Sección obligatoria en el spec

```markdown
## Estado del Arte
- **Práctica actual:** [qué hace el mundo]
- **Nuestra solución:** [qué vamos a hacer]
- **Veredicto:** ✅ Alineados / ⚠️ Divergimos porque [razón válida]
```

### Gate 1
- [ ] Spec aprobada por Enzo
- [ ] Sección "Estado del Arte" completa con veredicto
- [ ] Decisión técnica justificada
- [ ] Guardada en `docs/superpowers/specs/YYYY-MM-DD-feature-design.md`

---

## FASE 2 — DECIDE [GATE 2]

**Objetivo:** Plan con tasks verificables, uno por uno.

### Tools

| Tool | Cuándo usarlo |
|------|--------------|
| Skill `superpowers:writing-plans` | Generar plan desde la spec aprobada |
| Subagente `plan-writer` | Features complejas multi-módulo |

### Estructura de cada task

```markdown
### Task N — [Nombre imperativo]
- **Archivos:** path/to/file.py
- **Qué hacer:** descripción concreta
- **Success criteria:** cómo verifico que está bien
```

### Gate 2
- [ ] Plan aprobado por Enzo
- [ ] Cada task tiene success criteria verificable
- [ ] Dependencias identificadas (librerías, config, migraciones)
- [ ] Guardado en `docs/superpowers/plans/YYYY-MM-DD-feature.md`

---

## FASE 3 — IMPLEMENT [GATE 3]

**Objetivo:** Ejecutar el plan. TDD cuando aplica.

### Tools

| Tool | Cuándo usarlo |
|------|--------------|
| Skill `superpowers:subagent-driven-development` | Ejecución del plan task por task |
| Skill `superpowers:test-driven-development` | Features con cobertura de tests |
| Skill `superpowers:dispatching-parallel-agents` | Tasks independientes en paralelo |
| Skill `superpowers:using-git-worktrees` | Aislamiento para features grandes |
| Subagente `debugger` | Cuando algo falla y hay que diagnosticar |

### Reglas
- No desviar del plan sin actualizarlo primero
- No saltear tests para "hacerlo funcionar"
- No commitear hasta pasar GATE 5

### Gate 3
- [ ] Todos los tests nuevos pasan
- [ ] Sin errores de tipo ni warnings
- [ ] Happy path verificado manualmente

---

## FASE 4 — SIMPLIFY [GATE 4]

**Objetivo:** Misma funcionalidad, menos código. DRY + comentado.

### Tools

| Tool | Cuándo usarlo |
|------|--------------|
| Skill `simplify` | Pass automático de simplificación post-implementación |
| `mcp__CodeGraphContext__find_dead_code` | Detectar código sin uso |
| `mcp__CodeGraphContext__calculate_cyclomatic_complexity` | Detectar funciones muy complejas |
| `mcp__CodeGraphContext__find_most_complex_functions` | Top N funciones a simplificar |

### Checklist de simplificación

1. **Duplicación** — bloques iguales en >1 lugar → extraer helper/función/constante
2. **Longitud** — funciones >30 líneas → dividir con nombre descriptivo
3. **Comentarios** — lógica no obvia → comentar el *por qué*, no el *qué*
   - ❌ `# incrementa i`
   - ✅ `# offset +1 porque el índice es 1-based en la API de Milvus`
4. **Naming** — nombres que no se explican solos → renombrar
5. **Dead code** — imports/funciones/vars sin uso → eliminar

### Gate 4
- [ ] Sin duplicación (DRY)
- [ ] Lógica no obvia tiene comentario de "por qué"
- [ ] Sin dead code
- [ ] Skill `simplify` ejecutado

---

## FASE 5 — REVIEW [GATE 5]

**Objetivo:** Cero errores, cero regresiones.

### Tools

| Tool | Cuándo usarlo |
|------|--------------|
| Skill `superpowers:requesting-code-review` | Review completo del trabajo |
| Skill `superpowers:verification-before-completion` | Verificar antes de declarar "listo" |
| Subagente `superpowers:code-reviewer` | Cuando se completa un step major |
| Subagente `gateway-reviewer` | Si se tocó `saldivia/gateway.py` o auth/RBAC |
| Subagente `frontend-reviewer` | Si se tocó `services/sda-frontend/` |
| Subagente `security-auditor` | Antes de releases o si se tocó auth/permisos |

### Comando de tests
```bash
cd ~/rag-saldivia && uv run pytest saldivia/tests/ -v
```

### Gate 5
- [ ] Todos los tests pasan (unit + integration)
- [ ] 0 regresiones en tests existentes
- [ ] Review especializado ejecutado (gateway / frontend según corresponda)
- [ ] Sin vulnerabilidades obvias

---

## FASE 6 — DOCS SYNC [GATE 6]

**Objetivo:** Documentación 100% en sync con el código. Siempre.

### Tools

| Tool | Cuándo usarlo |
|------|--------------|
| Subagente `doc-writer` | Sync completo de docs al cierre de la iteración |
| Skill `claude-md-management:revise-claude-md` | Actualizar CLAUDE.md si cambió arquitectura |
| Skill `changelog-generator` | Generar entrada de changelog desde commits |

### Qué actualizar

1. **READMEs** de zonas afectadas:
   - `saldivia/` → `saldivia/README.md`
   - `services/sda-frontend/` → `services/sda-frontend/README.md`
   - `cli/` → `cli/README.md`
   - `scripts/` → `scripts/README.md`
   - `config/` → `config/README.md`
   - Root → `README.md`

2. **Docstrings** en funciones nuevas o modificadas:
   - Python: `"""Qué hace. Args: ... Returns: ... Raises: ..."""`
   - TypeScript/Svelte: JSDoc `/** ... */`

3. **CHANGELOG.md** — entrada bajo `[Unreleased]`:
   ```markdown
   ## [Unreleased]
   ### Added
   - Rate limiting por usuario en /auth/session
   ### Fixed
   - Upload limit no se aplicaba en rutas alternativas
   ```

4. **CLAUDE.md del proyecto** — solo si cambió arquitectura, patrones clave, puertos o servicios.

### Gate 6
- [ ] READMEs de zonas afectadas actualizados
- [ ] Docstrings en funciones nuevas/modificadas
- [ ] CHANGELOG.md con nueva entrada
- [ ] CLAUDE.md actualizado si aplica

---

## Resumen de Tools por Fase

```
OBSERVE   → firecrawl + mcp__CodeGraphContext + mcp__repomix + Explore (subagente)
ORIENT    → superpowers:brainstorming + Plan (subagente)
DECIDE    → superpowers:writing-plans + plan-writer (subagente)
IMPLEMENT → subagent-driven-development + TDD + dispatching-parallel-agents + worktrees + debugger
SIMPLIFY  → simplify + CGC (dead_code, complexity)
REVIEW    → requesting-code-review + code-reviewer + gateway/frontend-reviewer + security-auditor
DOCS      → doc-writer + revise-claude-md + changelog-generator
```

---

## Commits

**Regla:** Commits SOLO cuando Enzo los pide explícitamente. Nunca proactivos.

**NO hacer:**
- Commitear después de completar una tarea sin que Enzo lo pida
- Asumir que "listo" significa "commitear"

---

## Storage de Specs y Plans

- **Specs** (ORIENT) → `docs/superpowers/specs/YYYY-MM-DD-feature-design.md`
- **Plans** (DECIDE) → `docs/superpowers/plans/YYYY-MM-DD-feature.md`
- **Naming:** fecha ISO + descripción en kebab-case

Cuando arrancás una feature nueva:
1. Leer specs relevantes para entender decisiones pasadas
2. Leer plans similares para ver cómo se implementó algo parecido
3. Seguir la misma estructura

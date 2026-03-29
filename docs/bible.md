# La Biblia — Saldivia RAG

> **Documento permanente.** Define cómo se trabaja en este repo para siempre.
> Los planes de cada versión (1.0.x, 2.0.x, etc.) son documentos separados
> en `docs/plans/`. Este documento es la referencia que no cambia de versión
> a versión — solo se actualiza cuando cambia algo fundamental.
>
> **Plan maestro actual:** `docs/plans/1.0.x-plan-maestro.md`

---

## Principios

1. **Cuestioná el requerimiento antes de escribir código.** Opus tiene el deber
   de preguntar "¿es realmente necesario?" antes de planificar.
2. **Si el plan es difícil de explicar, el plan está mal.**
3. **Borrá lo que no debería existir. No optimices lo que queda hasta hacerlo.**
4. **Si nadie usaría una v1 rota, el scope está mal.** Achicar el scope, no la ambición.
5. **Cada paso debe saber lo que el paso anterior decidió.** (artifact persistence)
6. **Evidencia le gana a convicción. Convicción le gana a consenso.**
7. **Arreglalo o preguntá. Nunca lo ignores.**
8. **La seguridad no es un tradeoff. Es una restricción.**
9. **El output debería verse mejor de lo que se pidió.**

---

## Producto

- **Nombre:** Saldivia RAG
- **Logo:** el nombre es el logo (texto, sin imagen)
- **Colores:** tokens de Claude, acento **azure blue** (no naranja)
- **Idioma UI:** español
- **Idioma código:** inglés (variables, commits, docs técnicos)
- **Modelo de deploy:** single-tenant by deployment (cada empresa = su servidor)
- **Repo:** https://github.com/Camionerou/rag-saldivia

---

## Stack definitivo

| Componente | Tecnología | Nota |
|---|---|---|
| Frontend + Backend | Next.js 16 App Router (unificado) | Extraíble a futuro via packages/ |
| Base de datos | SQLite (libsql) | Migrar a Postgres con Drizzle cuando escale |
| ORM | Drizzle | Soporta SQLite y Postgres |
| Runtime | Bun | TypeScript nativo, rápido |
| Auth | JWT (jose) + Redis blacklist | — |
| Queue | BullMQ + Redis | — |
| AI/Streaming | Vercel AI SDK (`ai`) | Reemplaza SSE manual |
| Generative UI | json-render | Respuestas ricas del RAG |
| Validación | Zod | Compartido entre todos los paquetes |
| CSS | Tailwind v4 + shadcn/ui + Radix | Tokens claude-azure |
| Monorepo | Turborepo + Bun workspaces | — |
| Testing | bun:test + happy-dom + Playwright | — |

SQLite es suficiente para single-tenant. Drizzle facilita migrar a Postgres en 1 día.
Next.js unificado porque la lógica está en `packages/` — extraíble si se necesita API separada.

---

## Organización del código

```
apps/
  web/                    → Next.js (frontend + API + auth)
    src/
      app/                → rutas (pages + API routes)
      components/         → componentes React por dominio
      hooks/              → hooks custom
      lib/                → lógica (auth, rag, utils)
    public/               → assets estáticos

packages/                 → lógica de negocio reutilizable
  db/                     → Drizzle ORM + queries + schema
  shared/                 → Zod schemas + tipos
  config/                 → config loader
  logger/                 → logger estructurado

docs/
  bible.md                → este documento
  plans/                  → planes de implementación (por versión)
  decisions/              → ADRs
  templates/              → templates (plan, commit, PR, version, ADR, artifact)
  artifacts/              → resultados de reviews/audits
  toolbox.md              → herramientas externas

_archive/                 → código archivado (recuperable)

.claude/agents/           → agents especializados (todos Opus)
.cursor/rules/ + skills/  → config Cursor (backup)
```

---

## Workflow — Cómo se trabaja

**Centro de operaciones: Claude Code CLI.**
**Opus hace todo** — planifica, ejecuta, revisa, documenta, testea.

### Modelo de ejecución

```
Enzo ←→ Opus (Claude Code CLI)
              ↓
         Opus planifica → Enzo aprueba
              ↓
         Opus ejecuta → quality gates → commit
              ↓
         Repeat por fase
```

### Agents especializados (todos Opus)

Los agents en `.claude/agents/` son **perfiles especializados de Opus**, no
modelos distintos. Mismo modelo, distinta checklist.

| Agent | Cuándo |
|---|---|
| `frontend-reviewer` | Después de cambios en componentes/UI |
| `gateway-reviewer` | Después de cambios en API routes/auth |
| `security-auditor` | Antes de releases |
| `test-writer` | Tests nuevos |
| `debugger` | Algo no funciona |
| `doc-writer` | Actualizar docs |
| `deploy` | Deploy |
| `status` | Estado de servicios |
| `plan-writer` | Planes nuevos |
| `ingest` | Ingestar documentos |

### Sprint sequence

Cada plan sigue esta secuencia. No se salta ningún paso.

```
THINK   → ¿Es necesario? ¿Scope mínimo? ¿Ya existe solución al 80%?
PLAN    → Plan detallado en docs/plans/. Enzo aprueba.
EXECUTE → Opus ejecuta fase por fase. Contexto completo, sin delegación.
REVIEW  → Opus auto-revisa: ¿correcto? ¿compila? ¿scope drift?
TEST    → Quality gates: tsc → test → lint
SHIP    → Docs + CHANGELOG. Commit. Tag si es release.
```

### Intensity modes

| Modo | Cuándo | Qué se hace |
|---|---|---|
| **Quick** | 1-3 archivos, docs, config | Opus revisa directo, tsc + tests |
| **Standard** | Fases normales de un plan | Review 2 pasadas, tests completos, scope drift |
| **Thorough** | Releases, seguridad, refactors grandes | Todo + security audit + review exhaustivo |

---

## Herramientas — cuándo usar cada una

### MCPs

| MCP | Para qué | Cuándo |
|---|---|---|
| **CodeGraphContext** | Complejidad, dead code, dependencias, callers | Antes de refactors |
| **Playwright** | Browser, screenshots, testing visual | Verificación de UI |
| **GitHub** | Repos remotos, issues, archivos | Evaluar repos/herramientas |
| **context7** | Docs de librerías (Next.js, Drizzle, Zod, etc.) | **Antes de usar una API de librería** |

### Tools nativas

| Tool | Para qué | Cuándo |
|---|---|---|
| **Read** | Leer archivos | Siempre — antes de modificar |
| **Grep** | Buscar patrones | Encontrar usos, imports, strings |
| **Glob** | Buscar archivos | Por nombre/patrón |
| **Edit** | Modificar archivos | Cambios quirúrgicos |
| **Write** | Crear archivos | Archivos nuevos |
| **Bash** | Comandos terminal | git, bun, tsc |
| **WebFetch** | Leer URLs | Investigar repos, docs |
| **WebSearch** | Buscar en la web | Soluciones, comparaciones |

### Reglas de uso

1. **context7 ANTES que WebSearch** para docs de librerías
2. **CodeGraphContext ANTES de refactors** — verificar callers antes de mover algo
3. **Grep/Glob > repomix** para búsquedas locales
4. **WebFetch > firecrawl** para la mayoría de URLs
5. **Playwright para verificación visual** post cambios de UI
6. **Nunca leer 10 archivos uno por uno** — usar Grep primero

---

## Protocolos de trabajo

### 1. Convenciones de naming

**Archivos:**

| Tipo | Convención | Ejemplo |
|---|---|---|
| Componentes React | `PascalCase.tsx` | `ChatInterface.tsx` |
| Pages/routes | `page.tsx`, `layout.tsx` | `app/(app)/chat/page.tsx` |
| Hooks | `camelCase.ts` con `use` | `useRagStream.ts` |
| Lib/utils | `kebab-case.ts` | `lib/rag/client.ts` |
| Tests | mismo nombre + `.test.ts(x)` | `ChatInterface.test.tsx` |
| Plans | `1.0.x-plan{N}-{slug}.md` | `1.0.x-plan14-ai-sdk.md` |
| ADRs | `{NNN}-{slug}.md` | `012-stack-definitivo.md` |

**Código:**

| Tipo | Convención | Ejemplo |
|---|---|---|
| Componentes | `PascalCase` | `function ChatInterface()` |
| Funciones | `camelCase` | `getUserById()` |
| Variables | `camelCase` | `const sessionList` |
| Constantes | `UPPER_SNAKE_CASE` | `MAX_RETRIES` |
| Tipos | `PascalCase` | `type UserRole` |
| Zod schemas | `PascalCase` + `Schema` | `RagParamsSchema` |
| Env vars | `UPPER_SNAKE_CASE` | `JWT_SECRET` |

**Git:**

| Tipo | Convención | Ejemplo |
|---|---|---|
| Branches | `plan{N}-{slug}` | `plan14-ai-sdk` |
| Commits | `tipo(scope): desc — planN fN` | `feat(chat): integrate ai sdk — plan14 f1` |
| Tags | `v{MAJOR}.{MINOR}.{PATCH}` | `v1.1.0` |

**Tipos de commit:**

| Tipo | Cuándo |
|---|---|
| `feat` | Nueva funcionalidad |
| `fix` | Corrección de bug |
| `refactor` | Cambio sin cambio de comportamiento |
| `style` | CSS/UI sin cambio de lógica |
| `test` | Agregar o modificar tests |
| `docs` | Documentación |
| `chore` | Mantenimiento (deps, config, CI) |
| `ci` | Cambios de CI/CD |

**Scopes:** `web`, `db`, `shared`, `config`, `logger`, `ui`, `chat`, `auth`,
`rag`, `agents`, `plans`, `deps`

**Reglas:** inglés, minúscula después del tipo, sin punto final, max 72 chars,
siempre referenciar plan y fase al final.

**Versionado:** semver. MAJOR = breaking. MINOR = features. PATCH = bugfixes.
Release: CHANGELOG → version bump → commit `chore(release): vX.Y.Z` → tag → push.

### 2. Git workflow

- Se trabaja **directo en la branch activa** (hoy `1.0.x`)
- Un commit por fase completada
- Cuando haya más gente → feature branches + PRs

### 3. Inicio de sesión

```
1. Opus lee MEMORY.md + CLAUDE.md (automático)
2. Enzo dice qué quiere hacer
3. Opus confirma y arranca
```

Si es trabajo significativo → Opus lee la biblia y el plan maestro actual.

### 4. Quality gates

**Después de CADA fase:**
```bash
bunx tsc --noEmit    # tipos
bun run test         # unit tests
bun run lint         # lint
```

**Después de CADA plan:** + `test:components`, `test:visual`, `test:a11y`
**Antes de release:** + `test:e2e`, security audit

Si un gate falla → se arregla antes de commitear.

### 5. Escalación

**Opus escala a Enzo cuando:**
- Scope drift (archivos fuera del plan)
- > 3 fixes no planeados (WTF likelihood)
- Decisión de producto
- Algo podría romper funcionalidad existente
- Cambio de dependencia
- Cualquier guard rule

**Opus NO escala para:**
- Cambios dentro del scope del plan
- Fixes de tests del mismo cambio
- Ajustes de imports, formatting

### 6. Recovery

| Situación | Protocolo |
|---|---|
| Tests fallan | Fix forward (nuevo commit). Nunca revert. |
| Build no compila | Diagnosticar, arreglar. Si no → escalar. |
| Scope drift | SIEMPRE parar. Tolerancia cero. |

### 7. Supervisión

Después de cada fase, Opus reporta:

```
✅ Plan N, Fase N completada
Cambios: [archivos]
Tests: X passed, 0 failed
¿Sigo o revisás?
```

Enzo responde: "dale" / "mostrame el diff" / "pará" / "volvé atrás"

---

## Cómo crece el proyecto — expansión

```
Enzo dice algo → Opus THINK → plan o descarte → toolbox actualizado
```

**Nueva herramienta:** Opus investiga → evalúa → agrega a toolbox → plan si se integra
**Nueva feature:** Opus THINK → plan con template → agrega al roadmap
**Bug:** chico → fix directo | grande → plan nuevo
**Mejora de workflow:** discutir → actualizar biblia

### Reglas

1. Todo plan nuevo se registra en el plan maestro de la versión activa
2. El toolbox se actualiza automáticamente con cada herramienta mencionada
3. **El roadmap es flexible.** Los planes se pueden reordenar, insertar, o
   eliminar en cualquier momento. Si entre Plan 14 y Plan 15 surge algo
   urgente, se crea Plan 14.5 (o se renumeran). El roadmap es una guía,
   no un contrato.
4. Planes paralelos son posibles si no comparten archivos
5. La biblia solo se modifica por conversación con Enzo
6. **Numeración de planes intercalados:** si se inserta un plan entre 14 y 15,
   se numera como 14b. Si hay muchos intercalados, se renumera todo el roadmap.

---

## Archivos operativos del repo

### CHANGELOG.md

Formato: [Keep a Changelog](https://keepachangelog.com/). Secciones: Added, Changed, Fixed, Removed.

**Cuándo se actualiza:**
- Al completar cada plan, Opus agrega los cambios a `## [Unreleased]`
- Al hacer release, se mueve de `[Unreleased]` a `[X.Y.Z] — YYYY-MM-DD`
- Idioma del CHANGELOG: inglés

**Formato de entrada:**
```markdown
## [Unreleased]

### Added
- Vercel AI SDK for chat streaming (Plan 14)

### Changed
- NavRail reduced to 3 links: Chat, Collections, Settings (Plan 13)

### Removed
- Archived 64 aspirational components to _archive/ (Plan 13)
```

### LICENSE

MIT. Ya existe. No se cambia.

### CONTRIBUTING.md

**Audiencia principal: agentes de IA**, no humanos.

Debe explicar:
1. Leer `docs/bible.md` primero (reglas del proyecto)
2. Leer el plan maestro de la versión activa
3. Seguir las convenciones de naming y commits
4. Correr quality gates antes de commitear
5. No tocar archivos fuera del scope del plan

Cuando haya más gente (humanos), se agrega sección para contribuidores humanos.

### SECURITY.md

Ya existe (Plan 11). Se mantiene actualizado si cambia algo de auth/seguridad.

### README.md

**Audiencia dual:** humanos que descubren el repo + agentes que necesitan orientarse.

Secciones:
1. Qué es Saldivia RAG (1 párrafo)
2. Setup rápido (`bun install && bun run dev`)
3. Stack técnico (tabla)
4. Estructura del repo (tree simplificado)
5. Para agentes: "Leer `docs/bible.md`"
6. Para humanos: link a docs/

Se reescribe cada vez que hay cambios estructurales grandes.

### .editorconfig

Ya existe (Plan 12). Garantiza formato consistente (2 spaces, LF, UTF-8, trim whitespace).

### .gitignore

Ya existe. Se actualiza si se agregan nuevas herramientas o directorios.

### CODEOWNERS

Ya existe (Plan 11). Se actualiza cuando haya más gente.

---

## Mejora continua (inspirado en karpathy/autoresearch)

El repo debe estar en **mejora continua permanente**. No solo features nuevas —
también calidad, performance, tests, seguridad, dependencias.

### El concepto

Autoresearch de Karpathy usa un loop autónomo para mejorar modelos de ML:
`proponer cambio → testear → mejoró? → keep / discard → repeat`

Aplicamos el mismo concepto al repo:

### Loops disponibles

Enzo puede pedir "corré el loop de X" y Opus ejecuta iteraciones autónomas:

| Loop | Qué hace | Métrica | Comando de Enzo |
|---|---|---|---|
| **Code quality** | Analiza código → encuentra mejoras → implementa → tests pasan? | tsc errors, lint warnings → 0 | "mejorá la calidad del código" |
| **Test coverage** | Busca código sin tests → escribe test → pasa? | % coverage ↑ | "subí la cobertura" |
| **Dead code** | Busca código no usado → elimina → tests pasan? | Líneas eliminadas | "limpiá dead code" |
| **Performance** | Mide build/test time → optimiza → más rápido? | Segundos ↓ | "optimizá performance" |
| **Dependencies** | Busca updates → actualiza → tests pasan? | Deps outdated → 0 | "actualizá dependencias" |
| **Security** | Escanea vulnerabilidades → arregla → audit limpio? | Findings → 0 | "revisá seguridad" |
| **Complexity** | Busca funciones complejas → simplifica → complexity baja? | Cyclomatic complexity ↓ | "reducí complejidad" |

### Cómo funciona un loop

```
1. Opus mide el estado actual (baseline)
2. Opus identifica una mejora concreta
3. Opus implementa la mejora
4. Opus corre quality gates
5. Si la métrica mejoró Y tests pasan → commit
6. Si no → revert el cambio
7. Repeat (máx 10 iteraciones por sesión, o hasta que Enzo diga "pará")
```

### Reglas de mejora continua

1. **Cada iteración es atómica.** Un cambio, un test, un commit o revert.
2. **La métrica manda.** Si no se puede medir, no se puede mejorar.
3. **No romper nada.** Si los tests fallan, el cambio se descarta.
4. **Enzo puede parar en cualquier momento.** "pará" detiene el loop.
5. **Se reporta al final.** Resumen: N iteraciones, N mejoras, N descartadas.
6. **No es un plan.** Los loops no necesitan plan formal. Son mejoras
   incrementales que no cambian arquitectura ni agregan features.

### Cuándo correr loops

- **Entre planes:** cuando un plan termina y el siguiente no arrancó todavía
- **Días tranquilos:** cuando Enzo quiere mejorar sin agregar features
- **Después de un plan grande:** para limpiar y pulir lo que se hizo
- **Cuando Enzo lo pida:** "dale, mejorame el repo un rato"

---

## Guard rules (operaciones prohibidas sin OK de Enzo)

| Categoría | Bloqueado |
|---|---|
| Destrucción masiva | `rm -rf /`, wildcard removes |
| Historia Git | `--force` push, `reset --hard`, borrar branches remotas |
| Base de datos | `DROP TABLE`, `DROP DATABASE`, `TRUNCATE` |
| Producción | Deploy sin verificación |
| Seguridad | `chmod 777`, `--no-verify` |
| Código remoto | `curl ... \| sh` |

---

## Conflict resolution

- **Plan vs Review:** agrupar si hay acoplamiento real, separar si no
- **Scope creep durante ejecución:** es scope creep por defecto → parar
- **Security vs todo:** security siempre gana

---

## Mantenimiento de este documento

### Trigger events

| Evento | Qué se actualiza | Quién |
|---|---|---|
| Stack cambia | Sección Stack | Opus + OK de Enzo |
| Protocolo cambia | Sección Protocolos | Opus + OK de Enzo |
| Convención nueva | Sección Naming | Opus + OK de Enzo |
| Herramienta nueva | `docs/toolbox.md` (no este doc) | Opus |
| MCP cambia | Sección Herramientas | Opus |

### Regla de oro

**Si Opus lee algo en la biblia, tiene que ser verdad HOY.**
Si no es verdad → se actualiza inmediatamente.

### Verificación periódica

Cada 5 planes completados, Opus revisa en 5 minutos:
¿todo actualizado? ¿algo creció demasiado? ¿contradicciones?

### CLAUDE.md como puerta de entrada

CLAUDE.md se carga automáticamente al inicio de cada conversación de Claude Code.
Es el ÚNICO archivo que se lee siempre. Por eso, **las primeras líneas** de
CLAUDE.md deben ser:

```markdown
# Saldivia RAG

## LEER PRIMERO — antes de cualquier trabajo
1. `docs/bible.md` — reglas permanentes (workflow, stack, protocolos, naming)
2. `docs/plans/1.0.x-plan-maestro.md` — roadmap y planes de la versión actual

No empezar a trabajar sin leer estos documentos.
```

Esto garantiza que cualquier agente (Opus u otro) que abra una sesión en
este repo sepa inmediatamente dónde ir para orientarse.

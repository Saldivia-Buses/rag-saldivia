# Custom Agents para RAG Saldivia — Spec de diseño

**Fecha:** 2026-03-22
**Estado:** Aprobado, pendiente de implementación
**Autor:** Enzo Saldivia

> **Nota de infraestructura (2026-03-23):** Este documento fue escrito durante la era cloud (Brev + RunPod). Las referencias a `nvidia-enterprise-rag-deb106`, `brev-2gpu`, `runpod-rag`, y SSH remoto aplican ahora a la **workstation física Ubuntu 24.04** con perfil `workstation-1gpu`. El deploy local es `make deploy PROFILE=workstation-1gpu` sin SSH.

---

## Objetivo

Crear 10 agentes personalizados de Claude Code para el proyecto RAG Saldivia que cubran todas las áreas críticas: deployment, status, ingesta, code review (backend + frontend), seguridad, testing, debugging, documentación y planning. Los agentes viven en `.claude/agents/` dentro del repo para ser versionados y disponibles en cualquier sesión.

---

## Estructura de archivos

```
/Users/enzo/rag-saldivia/
└── .claude/
    ├── settings.local.json     (ya existe)
    └── agents/
        ├── deploy.md
        ├── status.md
        ├── ingest.md
        ├── gateway-reviewer.md
        ├── frontend-reviewer.md
        ├── security-auditor.md
        ├── test-writer.md
        ├── debugger.md
        ├── doc-writer.md
        └── plan-writer.md
```

---

## Convenciones globales

- Todos los agentes usan `memory: project` — el knowledge acumulado se versiona en `.claude/agent-memory/<name>/` y es compartido con el equipo
- Los agentes de solo lectura sin ejecución de comandos usan `permissionMode: plan`
- Los agentes que ejecutan Bash (status, debugger) usan `permissionMode: default` o `acceptEdits`
- Los agentes que escriben código usan `permissionMode: acceptEdits`
- Los agentes que necesitan confirmación humana (deploy, ingest) usan `permissionMode: default`
- Los agentes de análisis profundo usan `effort: high` o `effort: max`
- El agente `test-writer` usa `isolation: worktree` para generar tests en branch aislado
- `security-auditor` usa `model: opus` por ser el rol de mayor criticidad
- `status` usa `model: haiku` por ser el rol más simple y frecuente
- Todos los agentes que analizan código tienen acceso a `CodeGraphContext` y/o `repomix`
- Todos los agentes tienen acceso implícito a firecrawl para consultar docs externas

---

## Catálogo de agentes

### 1. `deploy`

**Propósito:** Deployar a Brev con preflight checks automáticos
**Triggers:** "deployar", "subir a brev", "push a producción", `make deploy`

**Frontmatter:**
```yaml
model: sonnet
tools: Bash, Read, Glob
permissionMode: default
maxTurns: 25
memory: project
mcpServers:
  - repomix
skills:
  - superpowers:verification-before-completion
  - superpowers:finishing-a-development-branch
```

**Conocimiento embebido:**
- SSH alias: `nvidia-enterprise-rag-deb106`
- Perfil de producción: `PROFILE=brev-2gpu`
- Comando de deploy: `ssh nvidia-enterprise-rag-deb106 "cd ~/rag-saldivia && git pull origin main && make deploy PROFILE=brev-2gpu"`
- Puertos del sistema: 3000 (frontend), 9000 (gateway), 8081 (RAG), 8082 (NV-Ingest)

**Preflight checks (en orden, se detiene al primer fallo):**
1. Tests Python: `uv run pytest saldivia/tests/ -v --tb=short` — si falla, reportar tests rotos
2. Build frontend: `cd /Users/enzo/rag-saldivia/services/sda-frontend && npm run build` — si falla, reportar error
3. Env vars críticas: verificar `JWT_SECRET`, `DB_PATH`, `RAG_SERVER_URL` en `config/.env.saldivia`
4. Git status: confirmar que no hay cambios sin commitear

**Failure modes conocidos:**
- `PYTHONPATH` + `set -u` → usar `${PYTHONPATH:-}` en scripts bash
- `docker network connect --alias` falla silencioso → disconnect primero
- Port conflict → verificar con `ss -tlnp | grep -E '3000|9000|8081|8082'`

**Output esperado:** tabla de preflight ✅/❌ + estado del deploy + URL de verificación

**Firecrawl:** Si un step falla con error desconocido, buscar solución en docs/GitHub Issues antes de rendirse.

---

### 2. `status`

**Propósito:** Estado completo de todos los servicios
**Triggers:** "está funcionando?", "cómo están los servicios?", "ver logs", "status", "está caído"

**Frontmatter:**
```yaml
model: haiku
tools: Bash, Read
permissionMode: default
maxTurns: 15
memory: project
```

**Servicios a verificar:**

| Puerto | Servicio | Health check |
|--------|----------|-------------|
| 3000 | SDA Frontend | `curl -sf http://localhost:3000/health` |
| 9000 | Auth Gateway | `curl -sf http://localhost:9000/health` |
| 8081 | RAG Server | `curl -sf http://localhost:8081/health` |
| 8082 | NV-Ingest | `curl -sf http://localhost:8082/health` |

**Checks adicionales:**
- `docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"` filtrado por servicios relevantes
- Últimas 50 líneas de logs del gateway: `docker logs saldivia-gateway --tail=50`

**Output esperado:** tabla con 🟢 UP / 🔴 DOWN / 🟡 DEGRADED + últimos errores si hay caídas + comando para reiniciar el servicio caído.

**Memoria:** acumula historial de caídas para detectar problemas recurrentes.

---

### 3. `ingest`

**Propósito:** Pipeline completo de ingesta de documentos
**Triggers:** "ingestar", "agregar documentos", "nueva colección", "indexar docs"

**Frontmatter:**
```yaml
model: sonnet
tools: Bash, Read, Glob
permissionMode: default
maxTurns: 20
memory: project
mcpServers:
  - repomix
  - CodeGraphContext
skills:
  - superpowers:verification-before-completion
```

**Conocimiento embebido:**
- Comando base: `make ingest DOCS=/path/to/docs COLLECTION=nombre`
- Script avanzado: `scripts/smart_ingest.py`
- Tier system de smart_ingest: tiny (<5 pág), small (5-20), medium (20-100), large (100+)
- Deadlock detection y adaptive timeout ya implementados
- Resume de ingestas interrumpidas disponible

**Firecrawl:** Consultar docs de NVIDIA NV-Ingest para errores de parsing de formatos específicos (PDF dañado, pptx con macros, etc).

**Memoria:** recuerda colecciones existentes, errores previos de ingesta, formatos problemáticos.

---

### 4. `gateway-reviewer`

**Propósito:** Code review especializado en `gateway.py` y el sistema de auth/RBAC
**Triggers:** cambios en `gateway.py`, `auth/`, nueva ruta FastAPI, "revisar gateway", "review del backend"

**Frontmatter:**
```yaml
model: sonnet
tools: Read, Grep, Glob
permissionMode: plan
effort: high
maxTurns: 30
memory: project
mcpServers:
  - CodeGraphContext
  - repomix
skills:
  - superpowers:receiving-code-review
  - error-handling-patterns    # skill global (sin namespace superpowers)
```

**Checklist de revisión (verificar para cada endpoint nuevo o modificado):**
1. ¿Tiene `require_role()` o guard de auth equivalente?
2. ¿El RBAC es correcto para el recurso que expone?
3. ¿Los errores internos no se filtran al cliente (sin stack traces en 500s)?
4. ¿JWT validation incluye todos los campos necesarios, incluyendo `name`?
5. ¿Los endpoints admin-only tienen guard server-side?
6. ¿SSE responses verifican HTTP status antes de hacer yield?
7. ¿Hay queries SQL dinámicas sin parametrizar?
8. ¿Los logs no incluyen tokens o secrets?

**Firecrawl:** OWASP Top 10 API Security, FastAPI security best practices, CVEs de dependencias detectadas.

**Memoria:** acumula patrones problemáticos encontrados, decisiones de diseño del sistema de auth.

---

### 5. `frontend-reviewer`

**Propósito:** Code review especializado en SvelteKit 5 BFF
**Triggers:** cambios en `services/sda-frontend/`, archivos `.svelte`, `.ts` del frontend, "revisar frontend"

**Frontmatter:**
```yaml
model: sonnet
tools: Read, Grep, Glob
permissionMode: plan
effort: high
maxTurns: 30
memory: project
mcpServers:
  - CodeGraphContext
  - repomix
skills:
  - superpowers:receiving-code-review
```

**Checklist de revisión:**
1. ¿El límite server/client está bien definido en SvelteKit 5?
2. ¿Las cookies de auth nunca llegan al cliente como datos?
3. ¿El BFF no filtra tokens JWT ni secrets en responses al browser?
4. ¿SSE streaming manejado correctamente (error handling, cleanup)?
5. ¿`hooks.server.ts` valida JWT en todas las rutas protegidas?
6. ¿Las rutas admin tienen guard server-side (no solo client-side)?
7. ¿Los `load` functions no exponen datos sensibles en `$page.data`?
8. ¿Form actions usan `fail()` correctamente para errores de validación?

**Firecrawl:** docs oficiales de SvelteKit 5 (hooks, load functions, form actions, server routes) para verificar patterns correctos.

---

### 6. `security-auditor`

**Propósito:** Auditoría de seguridad completa del sistema
**Triggers:** "revisar seguridad", "security audit", "¿es seguro esto?", antes de releases importantes

**Frontmatter:**
```yaml
model: opus
tools: Read, Grep, Glob
permissionMode: plan
effort: max
maxTurns: 40
memory: project
mcpServers:
  - CodeGraphContext
  - repomix
```

**Áreas de auditoría:**

**JWT/Auth:**
- Todos los campos requeridos presentes en el token (`sub`, `name`, `role`, `exp`)
- Algoritmo de firma seguro (no `none`, no HS256 con secret débil)
- Expiración configurada y validada
- Refresh tokens no reutilizables

**RBAC:**
- Mapa completo de endpoints vs roles permitidos (usando CGC para encontrar todos los endpoints)
- Sin endpoints sin guard de auth
- Sin privilege escalation posible

**SQLite/DB:**
- Queries parametrizadas (no f-strings con input de usuario)
- Paths de DB no expuestos en responses
- `detect_types=PARSE_DECLTYPES` no usado (causa crash con timestamps date-only — usar helper `_ts()`)

**Exposición de información:**
- Stack traces no en responses 500
- Tokens/secrets no en logs
- Variables de entorno no en responses
- Headers de seguridad en responses HTTP

**Dependencias:**
- Buscar CVEs conocidos en NVD para versiones de FastAPI, httpx, python-jose, SvelteKit usadas

**Firecrawl:** NVD CVE database, OWASP JWT Security Cheat Sheet, OWASP API Security Top 10, Python security advisories.

**Memoria:** acumula hallazgos históricos, decisiones de seguridad tomadas, CVEs ya evaluados.

---

### 7. `test-writer`

**Propósito:** Escribir tests pytest y Playwright siguiendo los patrones del proyecto
**Triggers:** "escribir tests", "agregar tests para X", "coverage de Y", nueva funcionalidad sin tests

**Frontmatter:**
```yaml
model: sonnet
tools: Read, Write, Edit, Grep, Glob
permissionMode: acceptEdits
isolation: worktree
maxTurns: 35
memory: project
mcpServers:
  - CodeGraphContext
  - repomix
skills:
  - superpowers:test-driven-development
  - superpowers:verification-before-completion
```

**Conocimiento embebido:**

*Patrones pytest del proyecto:*
- Fixtures en `saldivia/tests/conftest.py`
- `AuthDB` en memoria (`:memory:`) para tests
- Mock del RAG Server con `respx` o `unittest.mock`
- Edge cases documentados: JWT sin campo `name`, SSE que devuelve HTTP 200 aunque falle, `detect_types` SQLite

*Patrones Playwright del proyecto:*
- Tests en `services/sda-frontend/tests/`
- Base URL: `http://localhost:3000`
- Auth flow: POST `/api/auth/login` → cookie → rutas protegidas

*Qué NO hacer:*
- No mockear lo que puede testearse con una DB real en memoria
- No ignorar el caso de JWT expirado
- No asumir que HTTP 200 en SSE significa éxito

**Firecrawl:** pytest docs, Playwright API docs para patterns específicos de testing async/streaming.

**Memoria:** recuerda fixtures existentes, patterns de mock establecidos en el proyecto, edge cases ya cubiertos.

---

### 8. `debugger`

**Propósito:** Debugging sistemático de problemas en el sistema
**Triggers:** "no funciona", "está roto", "falla X", "error en Y", cualquier traceback o error log

**Frontmatter:**
```yaml
model: sonnet
tools: Bash, Read, Grep, Glob
permissionMode: acceptEdits
effort: high
maxTurns: 40
memory: project
mcpServers:
  - CodeGraphContext
  - repomix
skills:
  - superpowers:systematic-debugging
```

**Protocolo de debugging (en orden):**
1. **Logs:** capturar logs completos del servicio afectado
2. **Config:** verificar variables de entorno, puertos, profiles
3. **Red:** verificar conectividad entre servicios (docker network, ports)
4. **Código:** trazar el flujo de ejecución con CGC hasta el punto de falla

**Failure modes documentados (verificar primero siempre):**

| Síntoma | Causa probable | Fix |
|---------|---------------|-----|
| `PYTHONPATH: unbound variable` | `set -u` + PYTHONPATH no definida | Usar `${PYTHONPATH:-}` |
| SSE siempre HTTP 200 aunque falle | httpx `StreamingResponse` no propaga status | Verificar status antes de hacer yield |
| JWT `name` undefined en UI | JWT generado sin campo `name` | Agregar `name` al payload del token |
| `detect_types` SQLite crash | Timestamps date-only incompatibles | Usar helper `_ts()` |
| `docker network connect` falla silencioso | Container ya en la red | Disconnect primero, luego connect |
| Gateway no accesible desde frontend | docker network incorrecto | Verificar alias y network compartida |

**Firecrawl (clave):** cuando el error tiene mensaje específico, buscarlo en GitHub Issues, Stack Overflow, docs de FastAPI/SvelteKit/httpx/Milvus/Docker. Copiar el mensaje exacto del log para buscar.

**Memoria:** acumula soluciones a problemas ya resueltos, patrones de error recurrentes, configuraciones que funcionaron.

---

### 9. `doc-writer`

**Propósito:** Mantener documentación sincronizada con el código
**Triggers:** "documentar X", "actualizar README", "update docs", "CLAUDE.md desactualizado", tras cambios estructurales

**Frontmatter:**
```yaml
model: sonnet
tools: Read, Write, Edit, Glob
permissionMode: acceptEdits
maxTurns: 30
memory: project
mcpServers:
  - repomix
  - CodeGraphContext
```

**Documentos que mantiene:**
- `saldivia/README.md` — SDK Python
- `services/sda-frontend/README.md` — Frontend SvelteKit
- `CLAUDE.md` del proyecto — cuando cambian patrones o failure modes
- `docs/superpowers/specs/*.md` — specs de features implementadas
- `config/profiles/*.yaml` — comentarios de configuración

**Principio fundamental:** nunca inventar funcionalidad. Siempre leer el código actual (con repomix o Read) antes de documentar. La documentación describe lo que existe, no lo que debería existir.

**Firecrawl:** cuando documenta integraciones externas (NVIDIA Blueprint, Milvus, Brev), consulta la doc oficial para incluir links y versiones correctas.

**Memoria:** recuerda qué docs existen, su estado de actualización, decisiones de documentación previas.

---

### 10. `plan-writer`

**Propósito:** Escribir planes de implementación para features nuevas
**Triggers:** "planear X", "escribir plan para Y", "quiero implementar Z", antes de features no triviales

**Frontmatter:**
```yaml
model: sonnet
tools: Read, Write, Glob
permissionMode: acceptEdits
effort: high
maxTurns: 30
memory: project
mcpServers:
  - repomix
  - CodeGraphContext
skills:
  - superpowers:writing-plans
  - superpowers:brainstorming
```

**Conocimiento embebido:**
- Fases aprobadas y completadas (NO tocar):
  - Fase 1: `docs/superpowers/plans/2026-03-18-fase1-fundacion.md` — completada
  - Fase 2: `docs/superpowers/specs/2026-03-19-fase2-chat-pro-design.md` — diseñada y aprobada
- Fases 3+ están pendientes, numeración continúa desde ahí
- Formato de planes: `docs/superpowers/plans/YYYY-MM-DD-<topic>.md`
- Convención de naming de fases: `fase<N>-<descripcion-corta>`

**Proceso obligatorio antes de escribir un plan:**
1. Usar repomix para entender el estado actual del codebase
2. Usar CGC para analizar qué código existe vs qué hay que crear
3. Verificar que el plan no pisa fases ya completadas
4. Estructurar en fases pequeñas con entregables claros

**Firecrawl:** investiga patterns de implementación en proyectos similares cuando planea features nuevas.

**Memoria:** recuerda las fases aprobadas, numeración actual, decisiones arquitectónicas tomadas en planes anteriores.

---

## Resumen de configuración

| Agente | Modelo | Permission | Effort | Isolation | CGC | Repomix | Skills superpowers |
|--------|--------|------------|--------|-----------|-----|---------|-------------------|
| deploy | sonnet | default | — | — | — | ✅ | verification + finishing-branch |
| status | haiku | default | — | — | — | — | — |
| ingest | sonnet | default | — | — | ✅ | ✅ | verification |
| gateway-reviewer | sonnet | plan | high | — | ✅ | ✅ | receiving-review |
| frontend-reviewer | sonnet | plan | high | — | ✅ | ✅ | receiving-review |
| security-auditor | opus | plan | max | — | ✅ | ✅ | — |
| test-writer | sonnet | acceptEdits | — | worktree | ✅ | ✅ | TDD + verification |
| debugger | sonnet | acceptEdits | high | — | ✅ | ✅ | systematic-debugging |
| doc-writer | sonnet | acceptEdits | — | — | ✅ | ✅ | — |
| plan-writer | sonnet | acceptEdits | high | — | ✅ | ✅ | writing-plans + brainstorming |

---

## Notas de implementación

- Los agentes se cargan al iniciar la sesión de Claude Code. Si se crea un archivo manualmente, usar `/agents` para cargarlo sin reiniciar.
- `memory: project` guarda el knowledge en `.claude/agent-memory/<name>/MEMORY.md` — commitear junto con los agentes.
- El `security-auditor` con `model: opus` tiene mayor costo por uso — invocar deliberadamente, no automáticamente.
- `test-writer` con `isolation: worktree` crea un worktree temporal — si no hace cambios, se limpia automáticamente.
- Firecrawl no aparece en el frontmatter porque está disponible vía Bash (CLI instalado globalmente). Los system prompts de cada agente incluyen instrucciones explícitas de cuándo y cómo usarlo.

# Next session — Phase 0 item 4 (tool capabilities) + kick off Phase 1

Arrancás en `2.0.7` (10 commits ahead de main). Último commit:
`cfadc619 feat(deploy-ops): --healthcheck self-probe + drop dead tenant-template secret`.
Workstation corre el monolito `app` + `erp` + infra, todos `(healthy)`,
login de `admin@sda.local` emite JWT real.

## Final goal (ADR 026 — no se pierde de vista)

SDA reemplaza Histrix. El empleado abre SDA y:

1. Tiene UI moderna cubriendo **todo** lo que Histrix hacía (1:1 parity,
   mejor UX).
2. Tiene chat donde el agente es su representante — cap parity chat ↔ UI.
3. Arma su dashboard personal (no hay dashboard global).
4. Arma sus rutinas personales.
5. Detrás, agentes hoardean data: mail ingest, WhatsApp interno,
   tree-RAG con ACL por colección.

La vara de parity: `.intranet-scrape/` — 676 tablas + ~4,500 XML-forms.
Progresión top-down de ADR 027; Phase 0 gana siempre.

## Estado Phase 0 (ADR 027)

| # | Item | Estado |
|---|---|---|
| 1 | Migration integrity | ✅ shipped (2.0.6) |
| 2 | No-op migrators | ✅ shipped (2.0.6) |
| 3 | Orphan tables | ✅ shipped (2.0.6) |
| 4 | Tool capabilities | ⏸️ **esta sesión** |
| 5 | Workstation drift | ✅ shipped (2.0.7) — drift-check target + shape target + full prod wire |

## Paso 0 — cerrar drift antes de nada

Branch 2.0.7 tiene 10 commits ahead de main; `make check-prod-drift`
está rojo. Opciones:

- `gh pr create --base main --head 2.0.7 --title "2.0.7 — SDA monolito vivo + ultrareview follow-ups + prod hardening" --body "..."` → review → merge.
- O fast-forward directo si te quedás tranquilo con los 10 commits.

Post-merge: en el workstation `cd /opt/saldivia/repo && git checkout main && git pull` para que drift cierre verde.

## Tarea principal — Phase 0 item 4: tool capabilities

**Skill:** `agent-tools`.

### Estado actual (leído en 2.0.7)

Tools en `services/app/internal/rag/agent/tools/`:

- `Definition` hoy tiene `Name`, `Service`, `Type` (`read`/`action`),
  `RequiresConfirmation`, `Description`, `Parameters`. **No hay `Capability`
  ni chequeo de perms.**
- `Executor` despacha sin consultar perms del usuario.
- El endpoint `/v1/agent` sí está autenticado con `sdamw.AuthWithConfig`
  (FailOpen=false). El JWT trae `perms []string` pero **no se propaga**
  al executor.

Ese mismatch es la superficie del bug: un usuario con perms reducidos
puede disparar cualquier tool vía chat que la UI le bloquearía.

### Diseño

1. Agregar `Capability string` a `Definition` (nombre tipo
   `"ingest.create"`, `"search.read"`, `"erp.invoices.write"`).
2. Backfill **toda** definición existente: las 3 core (search/ingest) y
   todas las módulo (`modules/*.yaml`). Para tools legítimamente
   open-to-authed, usar `"authed"` como sentinel (siempre granted para
   tokens válidos — explícito mejor que ausencia).
3. Plumb `userPerms []string` por el context del request al executor;
   `Executor.Execute(ctx, call, perms)` firma cambia.
4. Check **antes de dispatch**:
   - Si `tool.Capability == "authed"` → allow (any authed user).
   - Si `tool.Capability ∈ perms` → allow.
   - Sino → return tool result `{"error":"forbidden","capability":"..."}`
     al LLM (el agente explica al usuario por qué no se ejecutó).
5. Audit log de cada denial en `audit_log` (writer ya existe —
   `pkg/audit.NewWriter`).
6. Integration tests por rol (admin, user, e2e-test):
   - admin → erp write allowed.
   - user → erp write denied + no side effect + audit entry.
   - unauth → 401 antes de llegar al executor.

### Trampas conocidas

- Module tools cargan desde YAML vía `agenttools.LoadModuleTools`. El
  schema YAML tiene que tener `capability` nuevo — cada módulo declara
  su cap. Sin cap declarada → tool **rechazado al load** (fail closed,
  no silent-default).
- `check_job_status` + `search_documents` son read y cualquier authed
  puede usarlos → `"authed"`.
- `create_ingest_job` tiene `RequiresConfirmation=true` hoy —
  capability + confirmation son independientes; ambos viven.
- RBAC real del usuario: el JWT `perms` viene del auth service y lo
  persiste en `user_roles` + `roles`. Para el test matrix necesitás
  fixtures con roles distintos.
- Dispatch parcial: el agente puede llamar **varias** tools en un turno.
  Una denegada no debe cortar las otras — cada tool call se evalúa
  independientemente.

## Tarea secundaria — si item 4 cierra con tiempo, kick off Phase 1

**Skills:** `migration-health` + `htx-parity`.

Phase 1 §Data migration primera fila de ADR 027:

> Every legacy Histrix table in `.intranet-scrape/db-tables.txt` is
> either migrated into an `erp_*` SDA table, or has a waiver ADR
> stating it's dead data.

Paso concreto:

1. Hacer diff entre `.intranet-scrape/db-tables.txt` (676 nombres) y el
   set de `erp_*` tables en `sda_tenant_saldivia_bench` después de la
   última run de migración.
2. Listar tablas Histrix **sin** contraparte SDA.
3. Para cada una: migrator nuevo, o entrada de waiver en
   `docs/parity/waivers.md` (ya existe una con W-001/W-002/W-003 de
   2.0.6).
4. Empezás por las que tienen más filas en Histrix.

**No ataques UI parity (Phase 1 §UI parity) hasta que §Data migration
esté verde** — las pages vacías no valen.

## Fuera de scope

- Phase 2+ (chat, prompts jerárquicos, tree-RAG, ACL) — bloqueado por
  item 4 (tool capabilities) + Phase 1 data.
- Correr el prod stack en el workstation — está todo wireado pero
  operar prod requiere secrets populados en `deploy/secrets/`.
- Toques al frontend — Phase 1 §UI parity es su propia sección y tiene
  la sub-order `Data migration → UI parity`.

## Cierre esperado

- **Mínimo**: drift cerrado (merge 2.0.7→main) + item 4 shipped →
  Phase 0 5/5 completo.
- **Ideal**: lo anterior + primera fila Phase 1 Data migration (diff
  tablas Histrix vs SDA + waivers para las muertas).
- **Stretch**: lo anterior + 1-2 migrators nuevos para las tablas
  grandes sin cobertura.

Post-item-4: **Phase 0 al 100%** y toda Phase 2 desbloqueada
(hierarchical prompts, memory curator, tree-RAG con ACL). La sesión
siguiente puede atacar Phase 1 con la libertad de que el gate de
seguridad del agente ya cerró.

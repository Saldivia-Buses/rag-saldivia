# Next session — 2.0.22: foundations for ADR 028 + ADR 029

**Pivote crítico** (post 2026-04-20): el cutover prod a tenant
`saldivia` con DB migrada del Histrix bench expuso **7 bugs en
cascada**. La causa raíz: doble tenancy (silo + en código). Decisión:
**eliminar tenancy del código** (ADR 028) y **usar la MySQL de
Histrix dentro del container** como backing del ERP de SDA (ADR 029).

Histrix server **NO se apaga** todavía — sigue corriendo como
fallback operacional. Lo que hacemos es **copiar su DB de manera
segura** (como backup) y conectar SDA a esa copia. Eventualmente,
cuando SDA cubra 100% de UX, vendrá el power-off — pero en otra ADR.

Esta sesión NO toca la prod actual de saldivia (sigue con Postgres
+ tenant_id renombrado a `saldivia` que el rename masivo dejó OK).
Esta sesión empieza el **camino de migración estructural**: Phase A
de ADR 028 + Phase A de ADR 029, ambas non-breaking.

## Estado al cierre 2.0.21 (2026-04-20)

**Branches**:
- `main` con 2.0.21 mergeado + tag `v2.0.21` + GitHub release.
- `2.0.22` con ADR 028 + ADR 029 + este next-session.

**Prod en `srv-ia-01`**:
- Tenant `saldivia` apuntando a `sda_tenant_saldivia_bench`.
- Data: 23M rows ERP, todas con `tenant_id='saldivia'` (rename masivo
  ejecutado al cierre, eliminó la divergencia `saldivia_bench`).
- Login: `enzosaldivia@gmail.com / saldivia-prod-2026!`.
- Acceso: `http://srv-ia-01` (vía Tailscale + Traefik).
- Sidebar: muestra "Sugerencias y bugs" + módulos top-level
  (Tesorería, Clientes, Estadísticas, etc.).

**Bugs documentados en `docs/parity/`**:
- Phase 0: 15 FK orphans en bench (migrator) — irrelevante post-029.
- Phase 4: E2E suite specs requieren rewrite per-cluster (htmlFor a11y).
- Phase 5: type-vs-schema audit clean.
- Phase 6: RequirePerm gating per-cluster pendiente.
- Phase 7: PrefetchLink + Lighthouse CI per-cluster pendiente.
- Phase 9: error handling clean.

## Goal de esta sesión

Foundation work: ADR 028 (eliminate tenancy) Phase A + ADR 029
(Histrix MySQL as backing) Phase A + Phase B. **No breaking changes
en prod**. Prod sigue funcionando con Postgres durante toda la
sesión; el camino de MySQL se construye en paralelo.

## Phase A — ADR 028 foundations

Non-breaking. Sólo prepara el terreno.

1. **Code-review rule**: agregar a `.claude/skills/backend-go/SKILL.md`:
   *"NUEVAS sqlc queries NO incluyen `tenant_id` en SELECT/INSERT/
   WHERE. NUEVAS tablas NO declaran columna `tenant_id`. Refactor de
   queries existentes va en su propio commit por área."*
2. **Listar queries afectadas**: script que enumere todas las
   `WHERE tenant_id` en `services/*/db/queries/*.sql` y produzca un
   tracking sheet en `docs/parity/2026-04-XX-adr028-query-inventory.md`.
3. **Refactor PRIMERA query (entities)**: PR que drop `tenant_id` de
   `services/erp/db/queries/entities.sql`. Regenerar sqlc. Update
   handler para no pasar tenantID. Tests verdes. Smoke en prod local.
   Pattern: cada cluster siguiente sigue este shape.

## Phase A — ADR 029 foundations (paralelo)

Infra. NO toca código backend todavía. **Histrix server se queda
intacto, solo copiamos su DB como backup.**

1. **Agregar `histrix-mysql` (mysql:8)** a `deploy/docker-compose.dev.yml`:
   ```yaml
   histrix-mysql:
     image: mysql:8
     environment:
       MYSQL_ROOT_PASSWORD: <generated>
       MYSQL_DATABASE: histrix
     volumes:
       - histrix_mysql_data:/var/lib/mysql
     healthcheck:
       test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
   ```
2. **Backup-grade dump desde Histrix legacy**:
   - VPN + SSH al server Histrix (`sistemas@172.22.100.99`).
   - Memorias relevantes: `reference_db_saldivia` y
     `reference_histrix_access`.
   - `mysqldump --single-transaction --routines --triggers --events
     --hex-blob` del schema `saldivia` entero.
   - scp a workstation `srv-ia-01`.
   - **Histrix server queda intacto** — no se modifica nada del
     server legacy; es read-only safe (`--single-transaction`).
   - Documentar el procedimiento exacto en
     `docs/runbook/histrix-mysql-backup.md` para que sea repetible.
3. **Restore en `histrix-mysql`** del workstation. Verifica que
   `mysql -e "SELECT COUNT(*) FROM <tabla>"` devuelve los counts
   esperados.
4. **Conectividad**: el container `erp` puede `mysql -h histrix-mysql
   -u sda` y leer. Test simple. NO refactor de queries todavía.

## Phase B — ADR 029 first read path

Una vez Phase A está estable.

1. **sqlc engine = mysql** para una query específica de prueba (la
   más simple, ej. listar marcas de chasis).
2. **ERP service**: agregar conexión MySQL paralela (sin reemplazar
   Postgres). Una query nueva hits MySQL.
3. **Frontend**: la página correspondiente (ej.
   `/ingenieria/producto/chasis-marcas`) recibe data desde MySQL.
4. **Validar parity**: la misma data via UI Histrix vs SDA → idéntica.
5. **Documentar el patrón** para que cada cluster siguiente sea
   refactor mecánico.

## Out of scope esta sesión

- **Apagar Histrix server**. Sigue corriendo intacto.
- **Drop columnas `tenant_id` en data**. Eso es Phase C de ADR 028,
  va en cycle posterior.
- **Refactor masivo de queries**. Sólo entities (Phase A 028) +
  chasis-marcas (Phase B 029) como pilots.
- **Delete del migrator**. Esa es Phase F de ADR 029, sólo cuando
  TODO el ERP esté en MySQL.
- **Phase 8 (a11y), 10 (polish), 12 (trampas)** del plan 2.0.21.
  Esos quedan para sesiones de UI dedicadas — no son foundation.
- **Sync continuo Histrix legacy ↔ histrix-mysql container**. Por
  ahora dump+restore one-shot. Sync strategy queda como open
  question hasta que SDA empiece a escribir.

## Trampas heredadas (recordar)

- **Cookie `sda_refresh` con flag `Secure`**: en HTTP el browser la
  rechaza → loop logout en boot. Fix: `Secure` condicional al
  `SDA_ENV != production`.
- **Sidebar6 tiene navigation hardcoded** — `MODULE_REGISTRY` no es
  fuente de verdad. Pendiente de unificar.
- **Migration changes verifican `read = written + skipped`**: para
  ADR 029 esto se elimina porque no hay migrator. Pero dejar la
  regla en `migration-health` skill apuntando a "obsoleto post-029".
- **Tag + release por ciclo**: cerrar 2.0.22 con `v2.0.22`.
- **Linter mods stasheados**: pop al final si quedaron pendientes.

## Done checklist

- ADR 028 Phase A merged: code-review rule + query inventory + entities
  refactor (sin tenant_id, tests verdes).
- ADR 029 Phase A merged: `histrix-mysql` container + dump+restore
  documentado + erp container puede leer MySQL.
- ADR 029 Phase B merged: una query SDA hits MySQL, una página
  frontend muestra data desde ahí.
- Documentos de runbook escritos.
- Prod sigue funcionando (Postgres + saldivia, intacta).
- Tag `v2.0.22` + GitHub release + deploy a workstation.

## Candidatos sesiones futuras (2.0.23+)

| Orden | Tema | Pre-req |
|---:|---|---|
| 1 | ADR 028 Phase B: refactor de queries por área (ERP) | 028 Phase A |
| 2 | ADR 029 Phase B-D: bulk read paths + write paths | 029 Phase A |
| 3 | ADR 028 Phase C: drop columnas tenant_id | 028 Phase B |
| 4 | ADR 028 Phase D-E: JWT cleanup + frontend cleanup | 028 Phase C |
| 5 | ADR 029 Phase E-F: drop SDA Postgres ERP tables + delete migrator | 029 Phase D |
| 6 | Power-off Histrix legacy server | 029 Phase F + parity 100% |
| 7 | Phase 4 follow-up: E2E selector rewrite per-cluster | independiente |
| 8 | Phase 8 (a11y htmlFor sweep) | independiente |
| 9 | Phase 10 (polish loop por cluster) | Phase 8/9 done |
| 10 | Lighthouse CI ≥95 | infra ready |

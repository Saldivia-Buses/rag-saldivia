# Next session — 2.0.22: ADR 028 + ADR 029, big bang

**Pivote crítico** (post 2026-04-20): el cutover prod a tenant
`saldivia` con DB migrada del Histrix bench expuso **7 bugs en
cascada**. Causa raíz: doble tenancy (silo + en código).

**Decisión tomada**: eliminar tenancy del código (ADR 028) Y usar la
MySQL de Histrix dentro del container como backing del ERP de SDA
(ADR 029). Power-off de Histrix legacy server NO en este ciclo —
solo copiamos su DB como backup, server queda intacto.

**Habilitador clave**: prod en realidad es dev-grade. Cero usuarios
reales operando, cero data operacional ingresada. Eso permite
**cambios destructivos sin coexistencia ni backward-compat**. Los
ADRs ya reflejan esto: rollout big bang, no phased.

## Estado al cierre 2.0.21 (2026-04-20)

**Branches**:
- `main` con 2.0.21 mergeado + tag `v2.0.21` + GitHub release.
- `2.0.22` con ADR 028 + ADR 029 + este next-session.

**Prod en `srv-ia-01`**:
- Tenant `saldivia` apuntando a `sda_tenant_saldivia_bench`
  (Postgres con data renombrada de `saldivia_bench` → `saldivia`).
- Login: `enzosaldivia@gmail.com / saldivia-prod-2026!`.
- URL: `http://srv-ia-01` (Tailscale + Traefik).

**Bugs documentados en `docs/parity/`** (resueltos por 028+029 o
diferidos a sesiones de UI):
- Phase 0: 15 FK orphans en bench → muere con ADR 029 (no migrator).
- Phase 4: E2E selectors per-cluster → diferido.
- Phase 6: RequirePerm gating per-cluster → diferido.
- Phase 7: PrefetchLink + Lighthouse CI → diferido.
- Cookie `sda_refresh` Secure en HTTP → fix junto con cleanup auth.

## Goal de esta sesión

**Mergear ADR 028 + ADR 029 enteros, big bang.** Sin coexistencia,
sin migrations cuidadosas. Si algo se rompe, fix forward.

Al cierre:
- Cero columnas `tenant_id` en el schema.
- Cero `WHERE tenant_id = $1` en queries sqlc.
- JWT sin `tid`/`slug`. Middleware sin cross-validation. Traefik sin
  header injection. Frontend sin `NEXT_PUBLIC_TENANT_SLUG` ni
  `getTenantSlug()`.
- ERP entero contra `histrix-mysql` (mysql:8 en docker-compose),
  sqlc engine=mysql, driver go-sql-driver/mysql.
- Postgres del silo solo para platform/chat/collections/suggestions/
  audit_log.
- Migrator borrado (`tools/cli/internal/migration/*`).
- Histrix legacy server intacto (fallback operacional, power-off en
  ADR futura).

## Plan de trabajo

### Bloque 1 — ADR 028 (eliminate tenancy)

1. **Drop columns** + drop FK constraints derived from `tenant_id`:
   single migration `db/tenant/migrations/NNN_drop_tenant_id.up.sql`.
2. **sqlc queries**: bulk find+replace de `WHERE tenant_id = $1`
   (y los `, $2`/`$3` siguientes que se descalcan). sqlc regen.
3. **Handlers**: drop el primer arg `tenantID` de cada llamada
   service.
4. **JWT**: editar `pkg/jwt/jwt.go` para no firmar/leer `tid`/`slug`.
   Tests del paquete.
5. **Middleware**: drop líneas 95-118 de `pkg/middleware/auth.go`
   (cross-validation, X-Tenant-* headers, tenant.WithInfo).
6. **Traefik**: borrar el headers middleware injecting X-Tenant-Slug
   de `deploy/traefik/dynamic/{dev,prod}.yml`.
7. **Frontend**: drop `NEXT_PUBLIC_TENANT_SLUG`, `getTenantSlug()`,
   `tenantId/tenantSlug` de `AuthUser` y AuthStore. Drop el build-arg
   en `apps/web/Dockerfile` y compose.
8. **Tests**: actualizar fixtures (sqlc-generated test models pierden
   TenantID). Update auth integration tests.
9. **Smoke**: login en prod → ver entities → hacer una mutation
   simple. Verificar todo OK.

### Bloque 2 — ADR 029 (MySQL Histrix inside container)

1. **Backup-grade dump del Histrix legacy** (server queda intacto):
   - Memorias `reference_db_saldivia` + `reference_histrix_access`.
   - VPN + SSH `sistemas@172.22.100.99`.
   - `mysqldump --single-transaction --routines --triggers --events
     --hex-blob saldivia > histrix-saldivia.sql`.
   - scp a `srv-ia-01:/opt/saldivia/dumps/`.
   - Documento `docs/runbook/histrix-mysql-backup.md` con el
     procedimiento exacto + frecuencia recomendada.
2. **`histrix-mysql` container**: agregar a
   `deploy/docker-compose.dev.yml`:
   ```yaml
   histrix-mysql:
     image: mysql:8
     environment:
       MYSQL_ROOT_PASSWORD: ${HISTRIX_MYSQL_ROOT_PASSWORD}
       MYSQL_DATABASE: histrix
     volumes:
       - histrix_mysql_data:/var/lib/mysql
     healthcheck:
       test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
   ```
3. **Restore** el dump en `histrix-mysql`.
4. **sqlc**: cambiar engine a `mysql` para `services/erp/db/queries/`.
   Regenerar bindings (van a romper compilación — esperado).
5. **Driver**: reemplazar `pgx` con `go-sql-driver/mysql` en el
   ERP service. SQL syntax adjustments (LIMIT/OFFSET, RETURNING,
   etc.).
6. **Drop SDA Postgres ERP tables** en single migration. Migration
   trivial (DROP TABLE x N).
7. **Delete** `tools/cli/internal/migration/*` y el comando
   relacionado en `tools/cli/sda`.
8. **Smoke**: login → ver entities (ahora desde MySQL) → match con
   data Histrix.

### Cierre

1. PR a main (admin merge si la branch protection sigue activa).
2. Tag `v2.0.22` + GitHub release con nota destacando que el cycle
   eliminó la tenancy interna y movió ERP a MySQL.
3. Deploy workstation: `git pull && docker compose up -d --build`.
4. Smoke test del user.

## Out of scope

- **Power-off Histrix legacy server**. Sigue corriendo. ADR
  separada cuando SDA cubra 100% UX.
- **Sync continuo Histrix legacy ↔ histrix-mysql container**.
  Por ahora dump+restore one-shot. Ongoing sync es problema futuro
  (cuando empecemos a escribir a MySQL desde SDA, hay que decidir
  si Histrix legacy sigue siendo source of truth o si SDA lo es).
- **Phase 4/8/10 del plan 2.0.21** (E2E selectors, a11y htmlFor,
  polish loop). Esos son sesiones de UI dedicadas — no foundation.
- **Refactor del sidebar6 hardcoded** → `MODULE_REGISTRY` como
  fuente de verdad. Pendiente, no bloquea 028+029.

## Trampas heredadas (recordar)

- **Cookie `sda_refresh` con flag `Secure`**: en HTTP el browser la
  rechaza → loop logout en boot. Fix de paso al editar auth handler:
  `Secure` condicional al `SDA_ENV == production`.
- **Tag + release por ciclo**: cerrar 2.0.22 con `v2.0.22`.
- **Linter mods stasheados**: pop al final si quedaron pendientes.
- **`docker-compose.dev.yml` POSTGRES_TENANT_URL**: ya es
  env-driven (`${SDA_POSTGRES_TENANT_URL:-...}`) — al eliminar
  tenancy esa env queda obsoleta.

## Done checklist

- ADR 028 mergeado: cero `tenant_id` en code/schema/JWT/middleware/
  Traefik/frontend.
- ADR 029 mergeado: `histrix-mysql` container + ERP contra MySQL +
  Postgres ERP tables dropeadas + migrator deleted.
- Histrix legacy server intacto y reachable (fallback).
- Runbook `docs/runbook/histrix-mysql-backup.md` escrito.
- Prod operable end-to-end con la nueva arquitectura.
- Tag `v2.0.22` + release + deploy workstation.

## Candidatos sesiones futuras (2.0.23+)

| Orden | Tema | Pre-req |
|---:|---|---|
| 1 | Sync strategy histrix-mysql ↔ Histrix legacy (continuous?) | 029 mergeado |
| 2 | Cookie Secure + auth hardening cross-protocol | 028 done |
| 3 | Refactor sidebar6 → usar MODULE_REGISTRY | independiente |
| 4 | Phase 4 follow-up: E2E selector rewrite per-cluster | independiente |
| 5 | Phase 8 (a11y htmlFor sweep) | independiente |
| 6 | Phase 10 (polish loop por cluster) | Phase 4/8 |
| 7 | Lighthouse CI ≥95 | infra ready |
| 8 | Apply RequirePerm + PrefetchLink en pages restantes | 028 done |
| 9 | Power-off Histrix legacy server | parity 100% + sync stable |

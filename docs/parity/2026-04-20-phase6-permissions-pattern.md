# Phase 6 pattern — RequirePerm gating en mutation buttons

**Date:** 2026-04-20.

## Status

**Infra completa, aplicación pendiente cluster por cluster.**

Hooks shipped en commit `feat(web): hook useHasPermission + perms en
auth store`. Componente `<RequirePerm>` shipped junto con la
demostración en `tesoreria/page.tsx`.

Se necesita aplicar el patrón en ~20 pages que ya disparan
`api.post / api.patch / api.delete`. Cada cambio es una edición
quirúrgica, ideal para commits chicos por cluster.

## Pattern

```tsx
import { RequirePerm } from "@/components/auth/require-perm";

// Default (mode="hide"): el botón desaparece para users sin perm.
<RequirePerm perm="erp.treasury.write">
  <Button onClick={...}>Nuevo movimiento</Button>
</RequirePerm>

// Alternativa: mostrar disabled con tooltip (mode="disable").
<RequirePerm perm="erp.purchasing.approve" mode="disable">
  <Button onClick={...}>Aprobar</Button>
</RequirePerm>
```

Para gates más complejos:

```ts
import {
  useHasPermission,
  useHasAnyPermission,
  useHasAllPermissions,
} from "@/lib/auth/use-has-permission";

const canEdit = useHasPermission("erp.stock.write");
const canViewOrEdit = useHasAnyPermission("erp.stock.read", "erp.stock.write");
const canDoBoth = useHasAllPermissions("erp.invoicing.write", "erp.invoicing.post");
```

## Mapping permission → page

Permisos vienen del JWT del backend (61 claims, ej.
`erp.treasury.read`, `erp.treasury.write`, `erp.invoicing.post`,
`erp.manufacturing.certify`). El mapping aproximado:

| Path | Permission |
|---|---|
| `/tesoreria/*` | `erp.treasury.write` (✅ aplicado en `/tesoreria`) |
| `/administracion/almacen/*` | `erp.stock.write` |
| `/administracion/contable/*` | `erp.accounts.write` |
| `/administracion/facturacion` | `erp.invoicing.write` (post → `erp.invoicing.post`) |
| `/administracion/sugerencias` | `erp.suggestions.write` |
| `/administracion/reclamos` | `erp.invoicing.write` |
| `/manufactura/unidades` | `erp.manufacturing.write` |
| `/manufactura/certificaciones` | `erp.manufacturing.certify` |
| `/seguridad/incidentes` | `erp.safety.write` |
| `/seguridad/inspecciones` | `erp.safety.write` |
| `/calidad/no-conformidades` | `erp.quality.write` |
| `/calidad/inspecciones` | `erp.quality.write` |
| `/compras/ordenes` | `erp.purchasing.write` (approve → `erp.purchasing.approve`) |
| `/ingenieria/legal` | `erp.admin.write` |
| `/ingenieria/desarrollo` | `erp.admin.write` |
| `/ingenieria/producto` | `erp.admin.write` |
| `(core)/admin/dlq` | `bigbrother.admin` |
| `(core)/chat` | `chat.write` |
| `(core)/collections` | `collections.write` |

## Why hide vs disable

- **`mode="hide"`** (default): cuando el botón no tiene sentido para
  un user que NUNCA va a obtener el permiso (ej. operario en una
  page de admin). Reduce noise visual y previene clicks fallidos.
- **`mode="disable"`**: cuando es esperable que el user pueda
  obtener el permiso (ej. action que requiere supervisor approval),
  o cuando el botón es central a la UX y esconderlo confunde
  ("¿dónde está el botón aprobar?"). El tooltip explica.

Default es `hide` salvo decisión explícita en el cluster.

## Tests

Spec por permiso (un test que loga un user sin perm, navega a la
page, y assertcs que el botón NO está visible) está pendiente —
depende del E2E selector rewrite (Phase 4 follow-up). Se difiere a
2.0.22 cuando los specs estén estables.

Mientras tanto, smoke test manual:

```bash
# 1. login con e2e@sda.local (admin: tiene los 61 perms)
# 2. navegar a /tesoreria → "Nuevo movimiento" visible
# 3. (en una sesión paralela) crear un user de test SIN
#    erp.treasury.write y comprobar que el botón no aparece
```

## Next session

- Aplicar el patrón en cada page de la tabla (1 commit por cluster
  cuando se tope el cluster en otro work).
- Una vez E2E estable (Phase 4 follow-up), agregar 1 spec por
  permiso.

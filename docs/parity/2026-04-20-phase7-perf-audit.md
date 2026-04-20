# Phase 7 audit — Performance (McMaster-grade)

**Date:** 2026-04-20.

## Findings — better than expected

### Already in good shape

- **`page_size >= 100` in API queries: zero hits.** No client-side
  big-list filtering. Every list endpoint paginates server-side at
  ≤50.
- **Detail pages (`[id]/page.tsx`) with multiple `useQuery`: 11.**
  All 11 use parallel queries (zero `enabled: !!parent.field` chains
  detected). No cascading waterfalls.
- **Bundle sizes per `page.tsx`: ≤24KB.** The largest are
  `administracion/contable` (24K) and `tesoreria` (24K). All under
  the threshold where shipping more JS becomes the bottleneck.

### The actual gap

Zero pages implement **prefetch on hover** for list-row → detail
navigation. Click → cold detail fetch every time. This is the
single biggest McMaster-velocity win available right now (their UX
prefetches aggressively on row hover so detail pages feel instant).

## What shipped this cycle

`<PrefetchLink>` component at
`apps/web/src/components/erp/prefetch-link.tsx`:

```tsx
import { PrefetchLink } from "@/components/erp/prefetch-link";
import { erpKeys } from "@/lib/erp/queries";
import { api } from "@/lib/api/client";

<PrefetchLink
  href={`/tesoreria/cajas/${row.id}`}
  prefetch={() => ({
    queryKey: erpKeys.cashRegister(row.id),
    queryFn: () => api.get(`/v1/erp/treasury/cash-registers/${row.id}`),
  })}
>
  {row.name}
</PrefetchLink>
```

Wraps Next.js `<Link>` adding `onMouseEnter` / `onFocus` /
`onTouchStart` handlers that call `queryClient.prefetchQuery` once
per link. The detail page picks up the cache hit instantly on click.

## What's pending — apply per-cluster

The 11 detail pages should swap their list-row `<Link>` for
`<PrefetchLink>`:

| List page | Detail page |
|---|---|
| `/compras/proveedores` | `[id]` |
| `/manufactura/unidades` | `[id]` |
| `/rrhh/legajos` | `[id]` |
| `/ingenieria/producto/carroceria-modelos` | `[id]` |
| `/mantenimiento/equipos` | `[id]` |
| `/mantenimiento/taller/vehiculos` | `[id]` |
| `/tesoreria/cuentas-bancarias` | `[id]` |
| `/tesoreria/cajas` | `[id]` |
| `/administracion/contable/centros-costo` | `[id]` |
| `/administracion/almacen/articulos` | `[id]` |
| `/administracion/almacen/herramientas` | `[id]` |

Each is a 2-3 line change in the list page. Ideal for incremental
commits when the cluster comes up.

## Pendiente — Lighthouse run

Lighthouse against each `(modules)/**` page requires a real
browser+CI setup. Out of scope for this sandbox session. Plan for
2.0.22:

1. Set up a Lighthouse CI job (`@lhci/cli`) that hits the local dev
   server with `e2e@sda.local` JWT preloaded.
2. Threshold: Performance ≥ 95 on every module page.
3. Fail the build on regression.

## Other McMaster patterns still pending

- **Optimistic UI** in clusters 14, 15, 16 (notes, contacts,
  movements). Mutation `onMutate` updates cache before response,
  `onError` reverts.
- **Keyboard shortcuts** (`/` focus search, `Esc` close, arrows
  navigate, Enter open detail). Global handler.
- **Drill-down preserves filters/scroll** via URL params (`useSearchParams`).
- **`tabular-nums` + sticky header** in dense tables. Visual.

All belong to a dedicated UX polish session post-2.0.21.

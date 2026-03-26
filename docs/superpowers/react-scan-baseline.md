# react-scan Baseline — Pre Plan 7

**Fecha:** 2026-03-26  
**Estado:** Pre-rediseño (antes de Plan 7 — design system)  
**Herramienta:** react-scan 0.5.3  
**Cómo activar:** `bun run dev` en `apps/web` → overlay visible en el browser

> Este documento captura el estado de performance de renderizado PREVIO al rediseño visual.
> Sirve para medir el impacto del Plan 7 y priorizar memoizaciones.

---

## Cómo usar react-scan

1. `cd apps/web && bun run dev`
2. Abrí el browser en `http://localhost:3000`
3. Los componentes que se re-renderizan se iluminan en amarillo/rojo
4. La consola loguea cada re-render con el componente y la causa probable

---

## Páginas recorridas

### /chat

**Componentes con re-renders innecesarios:**

| Componente | Trigger | Frecuencia | Causa probable | Prioridad |
|---|---|---|---|---|
| _completar_ | _completar_ | — | — | — |

**Observaciones:**
- _completar tras recorrer la página_

---

### /chat/[id]

**Componentes con re-renders innecesarios:**

| Componente | Trigger | Frecuencia | Causa probable | Prioridad |
|---|---|---|---|---|
| _completar_ | _completar_ | — | — | — |

**Observaciones:**
- _completar_

---

### /admin/users

**Componentes con re-renders innecesarios:**

| Componente | Trigger | Frecuencia | Causa probable | Prioridad |
|---|---|---|---|---|
| _completar_ | _completar_ | — | — | — |

**Observaciones:**
- _completar_

---

### /admin/analytics

**Componentes con re-renders innecesarios:**

| Componente | Trigger | Frecuencia | Causa probable | Prioridad |
|---|---|---|---|---|
| _completar_ | _completar_ | — | — | — |

**Observaciones:**
- _completar_

---

### /settings

**Componentes con re-renders innecesarios:**

| Componente | Trigger | Frecuencia | Causa probable | Prioridad |
|---|---|---|---|---|
| _completar_ | _completar_ | — | — | — |

**Observaciones:**
- _completar_

---

### /collections

**Componentes con re-renders innecesarios:**

| Componente | Trigger | Frecuencia | Causa probable | Prioridad |
|---|---|---|---|---|
| _completar_ | _completar_ | — | — | — |

**Observaciones:**
- _completar_

---

## Resumen de acciones recomendadas para Plan 7

| Prioridad | Componente | Acción |
|---|---|---|
| Alta | _completar_ | React.memo / useMemo / separar estado |
| Media | _completar_ | — |
| Baja | _completar_ | — |

---

## Notas

- react-scan solo está activo en `NODE_ENV=development`
- No se incluye en el bundle de producción (tree-shaken por webpack via guard `process.env.NODE_ENV`)
- Para desactivar temporalmente: comentar `<ReactScanInit />` en `layout.tsx`

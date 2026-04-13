---
name: Convenciones registry.ts
description: Cómo se estructura una entrada en MODULE_REGISTRY — subnav no navItems, sin firstRoute
type: project
---

En `apps/web/src/lib/modules/registry.ts`:
- El campo de sub-rutas se llama `subnav` (tipo `SubRoute[]`), NO `navItems`.
- El campo `firstRoute` NO existe en la interface `ModuleManifest` — no agregar.
- Los íconos deben estar importados en el bloque de imports de lucide-react al tope del archivo.
- El módulo `manufactura` usa `BusFront` como ícono y tiene position 20.

**Why:** La interface está definida explícitamente y agregar campos no existentes rompe TypeScript.
**How to apply:** Siempre leer la interface antes de agregar una entrada nueva al registry.

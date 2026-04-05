# Plan 6 — UI Testing Suite: Pirámide Completa

> **Estado:** COMPLETADO — 2026-03-26/27
> **Branch:** `experimental/ultra-optimize`
> **Líneas originales:** ~1152 → comprimido a resumen post-ejecución
> **Spec:** `docs/superpowers/specs/2026-03-26-ui-testing-design.md`

---

## Qué se hizo

Se construyó la pirámide completa de testing de UI: react-scan baseline, component tests con @testing-library/react, visual regression con Playwright+Storybook, a11y audit con axe-playwright.

### Stack de testing agregado

| Herramienta | Rol |
|---|---|
| react-scan | Performance baseline (re-renders) |
| @testing-library/react | Component tests |
| happy-dom | DOM en Bun test |
| Playwright | Visual regression + a11y |
| axe-playwright | Auditoría WCAG AA |
| Storybook 8 | Entorno de aislamiento |

### Fases ejecutadas

| Fase | Qué | Resultado |
|------|-----|-----------|
| 1 | react-scan baseline | Instalado, inicializado en dev con dynamic import. Reporte template creado |
| 2 | Component tests | @testing-library/react + happy-dom. ~154 tests de componentes |
| 3 | Visual regression | 22 snapshots PNG baseline sobre Storybook |
| 4 | A11y audit | axe-playwright en 5 páginas clave (login, chat, collections, admin, settings) |
| 5 | CI integration | Scripts `test:components`, `test:visual`, `test:a11y` integrados |

### Convenciones establecidas

```typescript
// Patrón obligatorio para component tests:
import { afterEach } from "bun:test"
import { cleanup, render, fireEvent } from "@testing-library/react"
afterEach(cleanup)  // OBLIGATORIO
// Queries escopadas (NO screen global), fireEvent (NO userEvent — happy-dom compat)
```

### Resultado

- ~154 component tests nuevos
- 22 visual regression baselines
- 5 páginas auditadas con WCAG AA
- Preload setup: `--preload ./src/lib/component-test-setup.ts`

### Commits

| Fase | Commit | Descripción |
|------|--------|-------------|
| F1 | `dd9e308` | chore(web): instalar react-scan para baseline de performance |
| F2 | `20f6e21` | chore(web): setup @testing-library/react con happy-dom separado de lib tests |
| F3 | `4be4f54` | test(web): 56 component tests para 8 primitivos ui |
| F3 | `69ab46a` | test(web): 147 component tests para 20 componentes — f3 completo |
| F3 | `00f0676` | test(web): baseline de visual regression — 22 snapshots light/dark |
| Docs | `fba20cb` | docs(plans): plan 6 (ui testing suite) y plan 7 (design system warm intelligence) |
| Docs | `0e086bc` | docs: documentacion completa actualizada — claude.md, design-system.md, testing.md, workflows.md |
| Cierre | `f2d887d` | feat(ci): plan 6 completo — visual regression, maestro e2e flows, a11y, ci integration |

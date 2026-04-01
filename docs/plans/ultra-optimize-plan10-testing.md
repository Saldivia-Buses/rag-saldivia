# Plan 10 — Testing Completo: Visual Regression, A11y, E2E y Cobertura

> **Estado:** COMPLETADO — 2026-03-27
> **Branch:** `experimental/ultra-optimize`
> **Líneas originales:** ~574 → comprimido a resumen post-ejecución

---

## Qué se hizo

Cerrar los 4 gaps de testing que faltaban para un release público: visual regression post-upgrades, a11y post-design-system, E2E de flujos críticos, y reporte de cobertura.

### Contexto

Plan 8 actualizó Next.js 15→16, Lucide 0.x→1.7, Drizzle 0.38→0.45, Zod 3→4. Los 22 snapshots visuales y los tests de a11y no se habían corrido desde esos upgrades.

### Fases ejecutadas

| Fase | Qué | Resultado |
|------|-----|-----------|
| F10.1 | Visual regression post-upgrades | 22 snapshots actualizados. Diferencias por Lucide 1.7 (paths SVG) — cambio intencional |
| F10.2 | A11y audit completo | 5 páginas auditadas con axe-playwright. Fix: `--fg-subtle` contraste borderline corregido |
| F10.3 | Cobertura con reporte | Badge de cobertura, threshold enforcement |
| F10.4 | E2E Playwright | Flujos: login, chat, crear usuario, upload |
| F10.5 | Smoke tests Redis | Verificación de conexión Redis en CI |

### Resultado

- 22/22 visual baselines actualizados y versionados
- 5 páginas WCAG AA compliant
- E2E de flujos críticos funcionando
- Redis smoke test en CI

### Commits

| Fase | Commit | Descripción |
|------|--------|-------------|
| F10.1 | `448c073` | test(web): baseline regresión visual storybook (22 capturas png) |
| F10.1 | `df12c0d` | chore: versionar snapshots visuales e ignorar sqlite en apps/web/data |
| F10.2 | `00fb651` | fix(web): a11y token fg-subtle, main login, aria y opt-out react-scan |
| F10.2 | `93061d3` | test(web): playwright a11y con dev:webpack y env redis/db/jwt |
| F10.2 | `063394f` | fix(auth): logout revoca jwt en redis y borra cookie auth_token |
| F10.2 | `b386aa1` | fix(web): eliminar middleware.ts duplicado (solo proxy en next 16) |
| F10.4 | `3201357` | test(web): e2e playwright flujos críticos y smoke redis |
| F10.5 | `b1f9965` | ci: umbral cobertura db, jobs e2e y a11y con redis |
| Docs | `a63eef6` | docs(plans): lista de commits plan 10 |
| Docs | `361a1bd` | docs: readme badges, changelog plan 10 y checklists planes 8/10 |

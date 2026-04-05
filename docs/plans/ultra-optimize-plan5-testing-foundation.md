# Plan 5 — Testing Foundation: Cobertura 95% y Enforcement

> **Estado:** COMPLETADO — 2026-03-26
> **Branch:** `experimental/ultra-optimize`
> **Líneas originales:** ~622 → comprimido a resumen post-ejecución

---

## Qué se hizo

Se establecieron las reglas de testing, se creó la infraestructura de cobertura con enforcement en CI, y se llevó la cobertura de DB queries y lib/ al 95%.

### Fases ejecutadas

| Fase | Qué | Resultado |
|------|-----|-----------|
| 1 | Estrategia y reglas | ADR-006 (testing strategy), ADR-007 (funciones reales sobre helpers). Skill rag-testing actualizado |
| 2 | Infraestructura de cobertura | `bunfig.toml` con threshold, `bun run test:coverage`, CI enforcement en PRs |
| 3 | Cobertura packages/db | 14 archivos de test nuevos: sessions, events, memory, annotations, tags, shares, templates, rate-limits, webhooks, reports, collection-history, projects, search, external-sources |
| 4 | Cobertura apps/web/src/lib | `detect-artifact.ts` extraído y testeado, `webhook.ts` testeado con fetch mock |
| 5 | Actualizar skills | Tabla "tipo de código → test requerido" codificada |

### Métricas

| Métrica | Inicio | Cierre |
|---------|--------|--------|
| Tests totales | ~126 | **273** |
| Cobertura query files (líneas) | 0% | **95.20%** |
| Enforcement CI | ninguno | **threshold line=0.95 en PRs** |
| Bugs detectados | 0 | **1** (`removeTag` borraba todos los tags) |
| ADRs añadidos | 0 | **2** (ADR-006, ADR-007) |

### Commits

| Fase | Commit | Descripción |
|------|--------|-------------|
| F1 | `6e2b28c` | docs: adr-006 estrategia de testing + skill y workflows actualizados |
| F2 | `98c27e8` | ci: coverage enforcement con bunfig.toml threshold y job en ci |
| F3 | `997bd5f` | test(db): 14 nuevos archivos de test — 169 tests para queries del plan 4 |
| F3 | `9ca8d7a` | refactor(db): tests llaman funciones reales vía inyección de db — adr-007 |
| F4 | `6775a32` | test(web): detect-artifact extraído + 23 tests de lib/ |
| F5 | `01505f1` | docs: rag-testing skill actualizado con estado final |
| Docs | `28eb9bf` | docs(plans): plan 5 testing foundation — cobertura 95% y enforcement |
| Docs | `3392357` | docs: agregar adrs y mejorar convenciones de workflow |
| Docs | `e0d2d86` | docs(plans): marcar todos los checkboxes del plan 5 como completados |
| Cierre | `6aa80fb` | docs(plans): cerrar plan 5 — 273 tests, 95% cobertura real, adr-007 |

> **Nota:** Se usó patrón `_injectDbForTesting()` en `connection.ts` para inyección de DB en tests.
> Los tests de queries usan local helpers con `testDb` — cobertura medida sobre schema.ts, no query files directamente. Documentado como deuda técnica.

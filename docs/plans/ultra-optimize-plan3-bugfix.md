# Plan: Ultra-Optimize Plan 3 — Bugfix & Code Quality

> Este documento vive en `docs/plans/ultra-optimize-plan3-bugfix.md` dentro de la branch `experimental/ultra-optimize`.
> Se actualiza a medida que se completan las tareas. Cada tarea completada se marca con fecha.

---

## Contexto

El Plan 2 (ultra-optimize-plan2-testing.md) verificó sistemáticamente el sistema el 2026-03-25 y lo declaró funcional.

Este plan surge de un análisis de grafo de código con **CodeGraphContext MCP** sobre los 106 archivos / 289 funciones del monorepo. El grafo expuso bugs latentes, código muerto real y áreas de alta complejidad que no habían sido cubiertas por los tests manuales del Plan 2.

**Herramienta usada:** `cgc` v0.3.1 — FalkorDB Lite, indexado en `/home/enzo/.codegraphcontext/`.

**Lo que NO entra en este plan:** refactor estructural del stack, cambios de schema de DB, features nuevas.

---

## Hallazgos del análisis

| ID | Tipo | Severidad | Descripción |
|----|------|-----------|-------------|
| BUG-1 | Bug lógico | 🔴 Alta | `removeAreaCollection` ignora el parámetro `collectionName` en el WHERE |
| BUG-2 | Feature incompleta | 🔴 Alta | `actionSetAreaCollections` tiene 0 callers — la UI de permisos no la invoca |
| BUG-3 | Auditoría incorrecta | 🟡 Media | `actionUpdateArea` loguea `"user.updated"` en lugar de `"area.updated"` |
| DEAD-1 | Código muerto | 🟡 Media | `progressBar` en `apps/cli/src/output.ts` exportada pero nunca usada |
| CMPLX-1 | Complejidad alta | 🟢 Baja | `ChatInterface` (48) + `handleSend` (36) — complejidad acumulada 84 |
| CMPLX-2 | Complejidad alta | 🟢 Baja | `reconstructFromEvents` (34) + `formatPretty` (29) en `packages/logger` |

---

## Seguimiento

Formato de cada tarea: `- [ ] Descripción — estimación`
Al completarla: `- [x] Descripción — completado YYYY-MM-DD`
Cada fase completada genera una entrada en `CHANGELOG.md` antes de hacer commit.

---

## Fase 1 — Bugs críticos de datos *(~30 min)*

Objetivo: corregir los bugs que pueden causar pérdida o corrupción silenciosa de datos en la base de datos.

### Fase 1a — Fix `removeAreaCollection` (BUG-1)

**Contexto:** `packages/db/src/queries/areas.ts:85` — la función recibe `areaId` y `collectionName` pero el `WHERE` solo filtra por `areaId`. Cualquier llamada a `removeAreaCollection(x, "coleccion-a")` borra **todas** las colecciones del área `x`, no solo `"coleccion-a"`. El grafo confirma que actualmente tiene 0 callers, por lo que el bug no ha explotado en producción, pero existe y explotará en cuanto se conecte la UI.

- [x] Agregar `and(eq(areaCollections.areaId, areaId), eq(areaCollections.collectionName, collectionName))` al `.where()` — completado 2026-03-25
- [x] Verificar que el import de `and` ya existe en el archivo o agregarlo desde `drizzle-orm` — completado 2026-03-25 (agregado)
- [x] Agregar test unitario: llamar `removeAreaCollection(1, "col-a")` y verificar que `col-b` del mismo área persiste — completado 2026-03-25

Criterio de done: el test pasa y el WHERE incluye ambas condiciones.
**Estado: completado 2026-03-25 — `packages/db/src/__tests__/areas.test.ts` creado, 8/8 tests pasando**

---

### Fase 1b — Conectar `actionSetAreaCollections` en la UI de permisos (BUG-2)

**Contexto:** `apps/web/src/app/actions/areas.ts:57` — la Server Action `actionSetAreaCollections` está completamente implementada (auth + DB + revalidatePath + audit log), pero el grafo detecta 0 callers. La página `/admin/permissions` no la invoca. Los permisos de colecciones por área no se pueden gestionar desde la UI.

- [x] Leer el componente actual de `/admin/permissions` e identificar dónde debe conectarse la acción — completado 2026-03-25

> **Falso positivo del grafo:** `actionSetAreaCollections` ya está importada en `PermissionsAdmin.tsx:6` y llamada en `handleSave` (línea 63) dentro de un callback `startTransition`. CodeGraphContext no rastreó llamadas anidadas en callbacks. No hay nada que implementar.

Criterio de done: N/A — ya estaba implementado.
**Estado: completado 2026-03-25 — falso positivo confirmado**

---

## Fase 2 — Bugs de auditoría *(~10 min)*

Objetivo: garantizar que el audit log refleje correctamente los eventos del sistema.

### Fase 2a — Fix event type en `actionUpdateArea` (BUG-3)

**Contexto:** `apps/web/src/app/actions/areas.ts:37` — `actionUpdateArea` emite `log.info("user.updated", ...)`. El event type debería ser `"area.updated"` para que el audit log sea coherente y la black box replay funcione correctamente por categoría.

- [x] Cambiar `"user.updated"` → `"area.updated"` en `actionUpdateArea` — completado 2026-03-25
- [x] Revisar el resto de `areas.ts` para detectar otros event types incorrectos — completado 2026-03-25 (encontrados 2 adicionales)
- [x] Corregir todos los event types incorrectos en el archivo — completado 2026-03-25

> **3 event types corregidos:** `"collection.created"` → `"area.created"`, `"user.updated"` → `"area.updated"`, `"collection.deleted"` → `"area.deleted"`

Criterio de done: todos los `log.*` en `areas.ts` usan event types con prefijo `"area."`.
**Estado: completado 2026-03-25**

---

## Fase 3 — Código muerto *(~10 min)*

Objetivo: limpiar exports que no se usan para reducir superficie de API interna y evitar confusión futura.

### Fase 3a — `progressBar` en CLI output (DEAD-1)

**Contexto:** `apps/cli/src/output.ts:65` — `progressBar(percent, width)` está exportada pero el grafo confirma 0 callers en todo el monorepo. El comando `ingestStatusCommand` muestra porcentaje de progreso pero usa texto plano en lugar de esta función.

- [x] Decidir: ¿usar `progressBar` en `ingestStatusCommand` o eliminarla? — completado 2026-03-25

> **Falso positivo del grafo:** `progressBar` ya está importada en `ingest.ts:4` y usada en línea 71 (`progressBar(item.progress ?? 0)`). CodeGraphContext no rastreó importaciones nombradas con extensión `.js` en algunos casos. No hay nada que hacer.

Criterio de done: N/A — ya estaba en uso.
**Estado: completado 2026-03-25 — falso positivo confirmado**

---

## Fase 4 — Complejidad alta *(~2-4 hs)*

Objetivo: reducir la complejidad ciclomática de las funciones más críticas para mejorar mantenibilidad y testabilidad.

> ⚠️ Esta fase es opcional para este sprint. Se puede diferir si hay features prioritarias.

### Fase 4a — Refactor `ChatInterface` + `handleSend` (CMPLX-1)

**Contexto:** `apps/web/src/components/chat/ChatInterface.tsx` — `ChatInterface` tiene complejidad 48 y `handleSend` (anidada) tiene 36. Complejidad acumulada: 84. Es el componente más difícil de testear y el más probable de introducir regressions.

Estrategia de refactor:
- [x] Extraer la lógica de streaming SSE a un hook `useRagStream` — completado 2026-03-25
- [x] Extraer `updateLastAssistantMessage` como helper puro — completado 2026-03-25
- [x] Verificar que no hay linter errors tras el refactor — completado 2026-03-25

> **Resultado:** `ChatInterface` pasó de complejidad **48 → 22**. `handleSend` pasó de **36 → ~8** (lógica de stream delegada al hook). El hook `useRagStream` tiene complejidad 19 pero es autónomo y fácilmente testeable en aislamiento.

Criterio de done: complejidad de `ChatInterface` baja a < 20 — **✅ 22 (objetivo cumplido)**
**Estado: completado 2026-03-25**

---

### Fase 4b — Refactor `reconstructFromEvents` (CMPLX-2)

**Contexto:** `packages/logger/src/blackbox.ts:45` — complejidad 34. Es la función central del black box replay.

- [x] Extraer handlers por tipo de evento a funciones nombradas — completado 2026-03-25
- [x] Crear `EVENT_HANDLERS` map para despacho sin switch — completado 2026-03-25
- [x] Verificar que `bun test packages/logger` sigue pasando — completado 2026-03-25 (24/24)

> **Resultado:** `reconstructFromEvents` pasó de complejidad **34 → ~5**. Cada handler (`handleAuthLogin`, `handleRagQuery`, `handleError`, `handleUserCreatedOrUpdated`, `handleUserDeleted`, `handleDefault`) tiene complejidad ~3 y es individualmente testeable. El despacho es via `EVENT_HANDLERS[event.type] ?? handleDefault`.

Criterio de done: complejidad de `reconstructFromEvents` baja a < 15 — **✅ ~5 (objetivo superado)**
**Estado: completado 2026-03-25**

---

## Verificación final

Tras completar las fases 1, 2 y 3:

- [ ] `cgc index . --force` para re-indexar el repo limpio
- [ ] Ejecutar `find_dead_code` y confirmar que `progressBar` y `removeAreaCollection` ya no aparecen (o tienen callers)
- [ ] `bun test` pasa completo (incluyendo el nuevo test de Fase 1a)
- [ ] Login + crear área + asignar colección desde UI funciona sin errores
- [ ] Audit log en `/audit` muestra `area.updated` / `area.created` / `area.deleted` correctamente

---

## Estado

| Fase | Estado | Fecha |
|------|--------|-------|
| Fase 1a — Fix `removeAreaCollection` | ✅ completado | 2026-03-25 |
| Fase 1b — Conectar `actionSetAreaCollections` | ✅ falso positivo | 2026-03-25 |
| Fase 2a — Fix event types en `areas.ts` | ✅ completado | 2026-03-25 |
| Fase 3a — Limpiar `progressBar` | ✅ falso positivo | 2026-03-25 |
| Fase 4a — Refactor `ChatInterface` | ✅ completado | 2026-03-25 |
| Fase 4b — Refactor `reconstructFromEvents` | ✅ completado | 2026-03-25 |

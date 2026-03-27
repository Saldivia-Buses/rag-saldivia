# ADR-009: Server Components por defecto, Server Actions para mutaciones

**Estado:** Aceptado  
**Fecha:** 2026-03-27  
**Contexto:** Plan 8 — Fase 2 (Refactoring de arquitectura React)

---

## Contexto

La velocidad de construcción del stack dejó antipatrones acumulados en la capa React/Next.js que el análisis con repomix reveló sistemáticamente:

1. **`settings/memory/page.tsx` era la única página de `(app)/` con `"use client"` y raw `fetch()`** — violaba el patrón Server Component + Server Actions del resto de la app. La página hacía `useEffect + fetch("/api/memory")` para cargar datos y `fetch` directo para mutaciones.

2. **10 componentes Client hacían `useEffect + fetch`** para cargar datos que podían venir como props desde un Server Component padre (`CollectionSelector`, `IngestionKanban`, `KnowledgeGapsClient`, `ReportsAdmin`, `WebhooksAdmin`, `ExternalSourcesAdmin`, `AnalyticsDashboard`, `IntegrationsAdmin`, `WhatsNewPanel`, `ProjectsClient`).

3. **7 componentes hacían `fetch()` directo para mutaciones** en lugar de usar Server Actions (`areasAdmin`, `usersAdmin`, `ragConfigAdmin`, `permissionsAdmin`, `settingsClient`, `settings/memory`).

4. **`ChatInterface` recreaba 5 event handlers en cada render** sin memoización — `useCallback` era la solución directa.

5. **`d3` (~450 KB) y `react-pdf` (~600 KB) entraban al bundle inicial** aunque solo se usaban en rutas específicas (`/collections/[name]/graph` y la feature de export).

6. **Formularios admin sin `react-hook-form`** — validación manual duplicada en cada formulario.

7. **Estado de filtros de `AuditTable` en `useState`** en lugar de la URL — el estado se perdía al navegar.

8. **Sin tipado en Server Actions** — cualquier llamada podía pasar valores incorrectos sin error en compilación.

---

## Decisión

### 1. Server Components por defecto

Todo componente que carga datos de DB o del RAG server se convierte a Server Component. Los datos se pasan como props a los Client Components hijos que necesitan interactividad.

```
Server Component (page.tsx)
  → query DB / RAG
  → <ClientComponent data={data} />  ← solo recibe datos ya cargados
```

### 2. Server Actions para todas las mutaciones

Todas las mutaciones (crear, actualizar, eliminar) usan Server Actions en `apps/web/src/app/actions/`. Los Client Components llaman a las actions directamente — sin `fetch()` manual, sin URL hard-codeada, con tipado end-to-end.

**Nuevas actions creadas:** `actionLogout`, `actionCreateProject`, `actionDeleteProject`, `actionCreateWebhook`, `actionDeleteWebhook`, `actionCreateReport`, `actionDeleteReport`, `actionAddExternalSource`, `actionDeleteExternalSource`, `actionCreateShare`, `actionAddMemory`, `actionDeleteMemory`.

### 3. `next-safe-action` para tipado de Server Actions

```typescript
// lib/safe-action.ts
import { createSafeActionClient } from "next-safe-action"
export const action = createSafeActionClient()
export const authAction = createSafeActionClient().use(authMiddleware)
```

Todas las actions del admin usan `authAction` — el middleware de autenticación corre automáticamente sin olvidar `getCurrentUser()` en cada action.

### 4. `react-hook-form` para formularios admin

Los formularios de CRUD (`UsersAdmin`, `AreasAdmin`, `RagConfigAdmin`, `PermissionsAdmin`) usan `react-hook-form` con `zodResolver` — validación en el cliente antes de llamar la Server Action, sin duplicar la validación del servidor.

### 5. `nuqs` para estado de filtros en URL

`AuditTable` y otros componentes con filtros usan `useQueryState` de `nuqs` — los filtros viven en los query params de la URL, sobreviven navegación y son compartibles.

### 6. `useCallback` para handlers en `ChatInterface`

Los 5 event handlers de `ChatInterface` (`handleSubmit`, `handleStop`, `handleFeedback`, `handleSave`, `handleExport`) se envuelven en `useCallback` con las dependencias correctas. Evita recreación en cada keystroke.

### 7. `next/dynamic` con `ssr: false` para dependencias pesadas

```typescript
const DocumentGraph = dynamic(() => import("@/components/collections/DocumentGraph"), { ssr: false })
const PdfExport = dynamic(() => import("@/components/chat/PdfExport"), { ssr: false })
```

`d3` y `react-pdf` solo entran al bundle de la ruta que los necesita.

---

## Por qué no `useOptimistic` para todas las mutaciones

`useOptimistic` es adecuado para UI que necesita feedback instantáneo antes de que el servidor confirme (ej. like/dislike). Para las operaciones admin (crear usuario, eliminar área) la latencia del servidor es aceptable y el estado optimista agregaría complejidad sin beneficio perceptible. Solo se aplica `useOptimistic` donde la UX lo justifica (feedback de mensajes en chat).

---

## Consecuencias

**Positivo:**
- Cero `useEffect + fetch` en Client Components para carga de datos — el trabajo ocurre en el servidor
- Las 9 rutas API que quedaron sin callers se eliminan (reducción de superficie de ataque)
- `ChatInterface` deja de recrear handlers en cada keystroke
- Bundle de `/chat` y `/collections/graph` reducido significativamente al excluir d3 y react-pdf del chunk inicial
- Validación de Server Actions con tipos Zod compartidos — errores detectados en compilación

**Negativo / Trade-offs:**
- `next-safe-action` agrega una dependencia externa. Si el proyecto abandona next-safe-action en el futuro, hay que migrar todas las actions. Se acepta porque la alternativa (validar manualmente en cada action) es peor.
- Los tests de componentes que mockean Server Actions necesitan actualizar el mock: `mock(() => Promise.resolve({ data: undefined }))` en lugar de `mock(() => Promise.resolve())` — `next-safe-action` envuelve el resultado en `{ data }`.

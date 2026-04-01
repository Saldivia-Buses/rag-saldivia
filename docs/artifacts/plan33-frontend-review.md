# Frontend Review — Plan 33 Conectores Externos

**Fecha:** 2026-04-01
**Tipo:** review
**Intensity:** standard

## Resultado
**CAMBIOS REQUERIDOS** (2 bloqueantes, 4 debe corregirse, 5 sugerencias)

---

## Hallazgos

### Bloqueantes

1. **[actions/connectors.ts:84-94] `actionSyncNow` no verifica ownership del conector**

   La action `actionSyncNow` recibe `id`, `provider` y `collectionDest` del cliente y los encola directamente en BullMQ sin verificar que el conector pertenezca al usuario autenticado (`ctx.user.id`). Un admin podria pasar el ID de un conector de otro admin y enqueuar un sync.

   Peor aun: el `provider` y `collectionDest` vienen del cliente, no de la DB. Un atacante podria pasar cualquier provider/collection y el worker los procesaria tal cual.

   **Fix:** Leer el conector de la DB por `id` + `ctx.user.id`, verificar que existe y esta activo, y usar los valores de la DB (no del input) para encolar el job:

   ```typescript
   .action(async ({ parsedInput: { id }, ctx }) => {
     const source = await getExternalSourceById(id, ctx.user.id)
     if (!source) throw new Error("Conector no encontrado")
     if (!source.active) throw new Error("Conector inactivo")
     await externalSyncQueue.add("manual-sync", {
       sourceId: source.id,
       provider: source.provider,
       collectionDest: source.collectionDest,
       fullSync: false,
     })
     return { ok: true }
   })
   ```

2. **[actions/connectors.ts:69-82] `actionToggleConnector` usa raw Drizzle con dynamic imports en vez de una query function**

   Tres dynamic `await import()` en serie para hacer un update simple. Esto rompe el patron del proyecto donde toda logica de DB vive en `packages/db/src/queries/`. Ademas, `externalSources` se importa dinamicamente de `@rag-saldivia/db` y `eq`/`and` de `drizzle-orm` — innecesario y fragil.

   **Fix:** Crear `toggleExternalSource(id, userId, active)` en `packages/db/src/queries/external-sources.ts` e importar estáticamente como las demas funciones:

   ```typescript
   // En packages/db/src/queries/external-sources.ts
   export async function toggleExternalSource(id: string, userId: number, active: boolean) {
     const db = getDb()
     await db.update(externalSources).set({ active }).where(and(eq(externalSources.id, id), eq(externalSources.userId, userId)))
   }

   // En connectors.ts
   import { toggleExternalSource } from "@rag-saldivia/db"
   // ...
   .action(async ({ parsedInput: { id, active }, ctx }) => {
     await toggleExternalSource(id, ctx.user.id, active)
     revalidatePath("/admin/connectors")
     return { ok: true }
   })
   ```

---

### Debe corregirse

3. **[AdminConnectors.tsx] Labels sin `htmlFor` — accesibilidad rota**

   Todos los `<label>` del form (lineas 230, 237, 249, 254, 280, 285, 289, 294, 303, 309) no tienen `htmlFor` y los inputs no tienen `id`. Los `<select>` nativos tampoco tienen `id`. Esto rompe la asociacion label-input para lectores de pantalla.

   **Fix:** Agregar `id` a cada input/select y `htmlFor` correspondiente en cada label:

   ```tsx
   <label htmlFor="connector-type" className="text-sm text-fg-muted">Tipo</label>
   <select id="connector-type" value={provider} onChange={...} ...>
   ```

4. **[AdminConnectors.tsx:61-68] `handleDelete` muestra toast ANTES de verificar resultado**

   `handleDelete` hace `await actionDeleteConnector(...)` pero no chequea si el resultado fue exitoso. Si la action falla, igual se muestra el toast de exito. Lo mismo aplica a `handleToggle` (linea 70-76) y `handleSync` (78-83).

   **Fix:** Verificar `result?.data` antes del toast, y mostrar `toast.error()` si falla:

   ```typescript
   const result = await actionDeleteConnector({ id: c.id })
   if (result?.data) {
     toast.success(`Conector "${c.name}" eliminado`)
     refresh()
   } else {
     toast.error("Error al eliminar el conector")
   }
   ```

5. **[AdminConnectors.tsx:95-101] Empty state no usa `data-density="compact"` — el admin deberia ser compact**

   Segun el design system, admin usa `data-density="compact"`. Ninguna parte del componente aplica este atributo. Verificar que el layout padre (admin layout o page) lo tiene, o agregarlo al contenedor raiz del componente.

   **Fix:** Verificar que `AdminLayout` o el admin `layout.tsx` ya aplica `data-density="compact"`. Si no lo hace, agregar en el contenedor principal de `AdminConnectors`.

6. **[AdminConnectors.tsx:170-319] `ConnectorForm` no es exportado pero deberia tener su propio test**

   Con 150+ lineas, multiples states, dynamic fields por provider, y logic de `buildCredentials()`, es un sub-componente complejo que merece testing aislado. Actualmente no hay ningun test para `AdminConnectors`.

   **Fix:** Crear `apps/web/src/components/admin/__tests__/AdminConnectors.test.tsx` con al menos:
   - Render de empty state
   - Render de lista con connectors
   - Render de form con cada provider (verifica campos dinamicos)
   - Submit del form llama la action correcta
   - Delete confirmation dialog aparece

---

### Sugerencias

7. **[AdminConnectors.tsx:106, 141, 312] Mix de inline styles y Tailwind utilities**

   Hay uso de `style={{ padding: "16px 20px" }}`, `style={{ display: "flex", gap: "4px" }}`, `style={{ marginTop: "16px" }}`, etc. mezclado con clases Tailwind. Esto es un patron aceptado en el proyecto (Tailwind v4 spacing issue) pero las nuevas instancias deberian usar `className="flex gap-1"` / `className="p-5"` donde funcione, reservando inline styles solo para los espaciados que Tailwind v4 rompe (como `space-y`).

   **Ejemplo:** `style={{ display: "flex", gap: "4px" }}` deberia ser `className="flex gap-1"`.

8. **[AdminConnectors.tsx:39-46] `timeAgo` no internacionaliza**

   Usa strings como "Hace segundos", "Hace Xh", etc. hardcodeadas. Esto esta bien para la v1 (UI en espanol), pero considerar extraer a un helper compartido si se usa en mas de un componente.

9. **[AdminConnectors.tsx:219] `needsOAuth` variable pero no hay flow para conectar el OAuth**

   El form muestra un aviso informativo para Google/SharePoint pero los archivos de callback (`/api/connectors/callback/route.ts`) no existen. El conector se crea con `credentials: "{}"`. Asegurar que el plan o un plan posterior cubre el OAuth flow completo, o agregar un disclaimer mas explicito en la UI de que Google/SharePoint todavia no es funcional.

10. **[page.tsx:7-19] N+1 query en el Server Component**

    `Promise.all(sources.map(async (s) => ({ ...s, docCount: await countSyncDocuments(s.id) })))` hace una query por fuente. Con pocos conectores esto no importa, pero si crece deberia haber un `countSyncDocumentsBatch(sourceIds)` que haga un solo query agrupado. Mismo patron en `actionListConnectors`.

11. **[AdminConnectors.tsx:232-240] Select nativo sin focus ring**

    Los `<select>` nativos (lineas 232 y 256) usan classes de Tailwind para border y bg pero no tienen `focus:ring` ni `focus:outline` como los `<Input>` de shadcn. Esto genera inconsistencia visual en focus states.

    **Fix:** Agregar `focus:ring-1 focus:ring-accent focus:outline-none` a las clases del select.

---

### Lo que esta bien

- **Server component page.tsx** es genuinamente server: hace data fetching, pasa datos como props, no tiene `"use client"`.
- **Loading skeleton** correcto con `loading.tsx` para Suspense boundary.
- **Credentials nunca llegan al browser** — tanto el server component como `actionListConnectors` mapean campos explicitamente y excluyen `credentials`. Bien hecho.
- **`adminAction` middleware** en todas las actions — auth verificada server-side antes de cualquier mutacion.
- **`ConfirmDialog` para delete** — patron correcto del proyecto, no hay delete sin confirmacion.
- **`useTransition` para mutations** — patron correcto de Next.js, el UI no se bloquea.
- **Design system tokens** aplicados correctamente: `text-fg`, `text-fg-muted`, `text-fg-subtle`, `bg-surface`, `bg-bg`, `border-border`, `text-destructive`, `Badge variant="success"/"secondary"`. No hay colores hardcodeados.
- **`revalidatePath`** en las acciones de mutacion — los datos se refrescan del server.
- **BullMQ queue** configurada correctamente con retry exponential y cleanup.
- **Zod validation** en todos los inputs de las actions via `ConnectorProviderSchema` y schemas inline.
- **AdminLayout** correctamente agrega la tab "Conectores" con icono `Plug`, consistente con las demas tabs.

---

## Archivos revisados

| Archivo | Veredicto |
|---|---|
| `apps/web/src/components/admin/AdminConnectors.tsx` | Cambios requeridos |
| `apps/web/src/components/admin/AdminLayout.tsx` | OK |
| `apps/web/src/app/(app)/admin/connectors/page.tsx` | OK |
| `apps/web/src/app/(app)/admin/connectors/loading.tsx` | OK |
| `apps/web/src/app/actions/connectors.ts` | Cambios requeridos |

## Tests faltantes

- `apps/web/src/components/admin/__tests__/AdminConnectors.test.tsx` — no existe
- `apps/web/src/app/actions/__tests__/connectors.test.ts` — no existe (testear que credentials no se leak, que ownership se verifica)

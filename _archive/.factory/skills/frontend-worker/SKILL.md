---
name: frontend-worker
description: Worker para cambios en el frontend SvelteKit 5 — componentes, rutas, BFF endpoints, TypeScript types. Verifica con agent-browser.
---

# Frontend Worker

NOTE: Startup and cleanup are handled by `worker-base`. This skill defines the WORK PROCEDURE.

## When to Use This Skill

Features que modifican `services/sda-frontend/`. Incluye componentes `.svelte`, rutas `+page.svelte`/`+page.server.ts`, BFF endpoints `+server.ts`, y types en `$lib/server/gateway.ts`.

## Required Skills

- **agent-browser**: Verificar flujos de usuario en el browser. Invocar para cada verificación interactiva listada en `verificationSteps` de la feature.

## Work Procedure

### Preparación

1. Leer los archivos existentes que se van a modificar: `+page.svelte`, `+page.server.ts`, `gateway.ts` (types relevantes).
2. Revisar los componentes existentes similares para entender patrones de código (Svelte 5 runes, design tokens, naming conventions).
3. Verificar que los tipos TypeScript del backend (en `gateway.ts`) están actualizados antes de usarlos en componentes.

### Implementación

4. **Actualizar tipos primero** — si el backend agrega campos nuevos (e.g., `username`, `success`), actualizar `AuditEntry` y `AuditParams` en `gateway.ts` antes de escribir los componentes.

5. **Crear/modificar componentes** — seguir los patrones existentes:
   - Svelte 5 runes: `$state()`, `$effect()`, `$props()`, `$bindable()`
   - Design tokens: `--bg-surface`, `--border`, `--accent`, `--text-faint`
   - Accesibilidad: `role="switch"` + `aria-checked` para toggles; labels en todos los inputs
   - Deep-linking con `goto(..., { replaceState: true, keepFocus: true })`

6. **BFF endpoint** — crear `+server.ts` que proxy al gateway con guard de auth:
   ```typescript
   // Patrón estándar para BFF admin endpoint
   export const GET: RequestHandler = async ({ locals, url }) => {
     if (!locals.user || locals.user.role !== 'admin') {
       return json({ error: 'Forbidden' }, { status: 403 });
     }
     // forward params y proxy
   };
   ```

7. **Verificar TypeScript** — `npm run check` en `services/sda-frontend/` debe pasar sin errores antes de verificación interactiva.

### Reglas críticas para este codebase

- **Auto-refresh:** Usar `setInterval` con `$effect` cleanup (`return () => clearInterval(id)`) — NO `invalidateAll()` para refresh de componentes individuales.
- **CSV export:** Client-side con Blob URL; nunca hacer request adicional al backend para export.
- **Null safety en templates:** Usar `{entry.collection ?? '—'}` para campos nullable; nunca acceder a `.property` de null directamente.
- **Filtros + URL:** `$effect(() => { const params = ...; goto(`/audit?${params}`, { replaceState: true, keepFocus: true }); })` — sincronizar reactivamente.
- **Paginación:** offset numérico (no cursor opaco); entries se REEMPLAZAN en la tabla (no append).
- **Imports:** Verificar que los componentes nuevos se exportan/importan correctamente. No crear barrel files a menos que ya existan.

### Verificación final

8. Correr `npm run check` en `services/sda-frontend/`. **Debe pasar con 0 errores.**

9. **Verificar con agent-browser** — invocar el skill `agent-browser` para cada flujo interactivo. Para cada flujo verificado, registrar en `interactiveChecks`:
   - Navegación a la ruta
   - Cada interacción del usuario (filtrar, paginar, expandir, exportar)
   - Estado final observado (screenshot mental, DOM state, network calls)

10. Para cada flujo de agent-browser: verificar que la consola del browser está limpia (sin errores rojos).

## Example Handoff

```json
{
  "salientSummary": "Implementados 5 componentes audit (AuditFilters, AuditTable, AuditRow, AuditStats, ExportButton) y reemplazada la página /audit. BFF /api/audit creado. svelte-check: 0 errores. Verificados con agent-browser: filtros deep-link, paginación offset, export CSV, auto-refresh, expand rows.",
  "whatWasImplemented": "5 nuevos componentes en src/lib/components/audit/. +page.svelte reemplazado con filtros reactivos, tabla paginada, stats y auto-refresh. +page.server.ts actualizado con offset/limit. BFF src/routes/api/audit/+server.ts creado. AuditEntry y AuditParams en gateway.ts actualizados con username, success, offset.",
  "whatWasLeftUndone": "",
  "verification": {
    "commandsRun": [
      {
        "command": "cd /Users/enzo/rag-saldivia/services/sda-frontend && npm run check 2>&1 | tail -10",
        "exitCode": 0,
        "observation": "svelte-check: 0 errors, 0 warnings"
      }
    ],
    "interactiveChecks": [
      {
        "action": "Navegar a http://localhost:5173, login como admin, click en 'Auditoría' en sidebar",
        "observed": "Página /audit carga con tabla de entries, stats cards muestran 4 métricas, toggle auto-refresh OFF"
      },
      {
        "action": "Escribir 'chat' en filtro Acción",
        "observed": "URL actualiza a /audit?action=chat (replaceState); tabla recarga con solo entries de action=chat; stats actualizan el count"
      },
      {
        "action": "Con 100+ entries, verificar que tabla muestra exactamente 50 y click 'Siguiente'",
        "observed": "tbody tiene 50 filas; click 'Siguiente' → Network muestra GET /api/audit?offset=50; primera fila cambia"
      },
      {
        "action": "Click en primera fila de la tabla",
        "observed": "Panel expandido aparece con ip_address, ID completo, query_preview completo; Network sin requests adicionales"
      },
      {
        "action": "Click en ExportButton con filtro action=chat activo",
        "observed": "Archivo audit-1234567890.csv descargado; línea 1 es header; todas las rows tienen action=chat; sin requests en Network"
      },
      {
        "action": "Activar toggle auto-refresh, esperar 35 segundos",
        "observed": "Network muestra GET /api/audit?action=chat a los ~30s; tabla actualizada; filtro preservado"
      },
      {
        "action": "Con auto-refresh ON, navegar a /chat",
        "observed": "En Network tab de /chat: sin requests a /api/audit durante 60s; consola sin errores"
      },
      {
        "action": "Acceder directamente a /audit?action=ingest&userId=1",
        "observed": "Inputs pre-populados con 'ingest' y '1'; tabla filtrada; gateway recibió ?action=ingest&user_id=1"
      }
    ]
  },
  "tests": {
    "added": []
  },
  "discoveredIssues": []
}
```

## When to Return to Orchestrator

- `npm run check` falla con errores en tipos del backend que aún no están implementados
- El gateway en `localhost:9000` no está corriendo y no es posible verificar con agent-browser
- Se descubre que un componente reutilizable existente no puede extenderse sin refactoring mayor
- Los design tokens o patrones de componentes son inconsistentes y requieren decisión de arquitectura

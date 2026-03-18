# SDA Frontend — Field Testing Findings

## Metodología
Enzo prueba el frontend en vivo y reporta. Claude documenta acá.
Cuando hay suficientes hallazgos, se arma un batch de implementación.

---

## Sesión 2026-03-18

### Bugs encontrados

#### B1 — Gateway inestable: cuelga después de pocos requests ⚠️ CRÍTICO
- **Síntoma:** Fetch requests quedan "Pending" indefinidamente. Requiere restart manual.
- **Causa raíz:** Llamadas síncronas a SQLite dentro de `async def` en gateway.py bloquean el event loop de uvicorn cuando hay requests concurrentes.
- **Fix aplicado:** Gateway corre ahora con 2 workers + `restart: unless-stopped`.
- **Fix pendiente:** Wrappear todas las llamadas DB en `asyncio.to_thread()` para no bloquear el event loop.

#### B2 — Navegación sidebar no funciona cuando gateway está caído
- **Síntoma:** Clickear "Chat" desde Settings no navega — se queda cargando.
- **Causa raíz:** Derivado de B1. Los server-side loads de SvelteKit llaman al gateway; si este no responde, la navegación cuelga.
- **Fix pendiente:** Agregar timeout a todas las llamadas de gateway en el BFF, con fallback graceful.

#### B3 — Reload de página rompe la sesión
- **Síntoma:** Al recargar la página después de estar logueado, el frontend parece no reconocer la sesión y hay que abrir ventana incógnito nueva.
- **Causa raíz:** Probablemente relacionado con B1 (gateway caído al momento del reload). Posible también: cookie `Secure` + HTTP.
- **Fix pendiente:** Investigar si es estabilidad del gateway o un bug de manejo de cookie.

#### B4 — Logout / volver al login no funciona
- **Síntoma:** No hay forma visible de cerrar sesión o volver al login desde la app.
- **Fix pendiente:** Implementar botón de logout en sidebar.

#### B5 — Hover en historial genera storm de requests al gateway ✅ RESUELTO
- **Síntoma:** Al pasar el mouse por el panel de historial, SvelteKit prefetchea cada link disparando 3 requests al gateway por cada hover.
- **Fix aplicado:** `data-sveltekit-preload-data="false"` en links de historial.

#### B6 — `settings?/refresh_key` da 500 cuando el gateway está caído ✅ RESUELTO
- **Síntoma:** El botón de regenerar API key da 500, luego SvelteKit lanza 20+ requests de invalidación que toman 30-50s cada uno.
- **Fix aplicado:** Try/catch en action → devuelve `fail(503, {...})` en lugar de explotar.

### Ideas / Mejoras

<!-- Se irán agregando acá -->

---

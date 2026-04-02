# SDA Frontend — Field Testing Findings

## Metodologia
Enzo prueba el frontend en vivo y reporta. Claude documenta aca.
Cuando hay suficientes hallazgos, se arma un batch de implementacion.

---

## Sesion 2026-03-18

### Bugs encontrados

#### B1 — Gateway inestable: cuelga despues de pocos requests -- RESUELTO (parcial)
- **Sintoma:** Fetch requests quedan "Pending" indefinidamente. Requiere restart manual.
- **Causa raiz:** Llamadas sincronas a SQLite dentro de `async def` en gateway.py bloquean el event loop de uvicorn cuando hay requests concurrentes.
- **Fix aplicado:** Gateway corre ahora con 2 workers + `restart: unless-stopped`. Handlers convertidos a `def` (sync).
- **Pendiente:** Wrappear todas las llamadas DB en `asyncio.to_thread()` para no bloquear el event loop.

#### B2 — Navegacion sidebar no funciona cuando gateway esta caido -- RESUELTO
- **Sintoma:** Clickear "Chat" desde Settings no navega — se queda cargando.
- **Causa raiz:** Los server-side loads de SvelteKit llaman al gateway; si este no responde, la navegacion cuelga indefinidamente porque no habia timeout.
- **Fix aplicado (2026-03-18 batch):**
  - Timeout de 10s en TODAS las llamadas BFF->Gateway via `AbortController` en `gateway.ts`.
  - Error handling con `try/catch` en todas las load functions: `chat/+page.server.ts`, `chat/[id]/+page.server.ts`, `collections/+page.server.ts`, `collections/[name]/+page.server.ts`, `audit/+page.server.ts`, `admin/users/+page.server.ts`.
  - Las load functions ahora lanzan `error(503, ...)` en lugar de colgar, y SvelteKit muestra la pagina de error.
  - Error pages nuevas: `+error.svelte` (root) y `(app)/+error.svelte` (dentro del app layout con sidebar).

#### B3 — Reload de pagina rompe la sesion -- RESUELTO
- **Sintoma:** Al recargar la pagina despues de estar logueado, el frontend parece no reconocer la sesion y hay que abrir ventana incognito nueva.
- **Causa raiz:** Derivado de B1 (gateway caido al momento del reload). El BFF intentaba validar la sesion contra el gateway, que no respondia, y colgaba.
- **Fix aplicado (2026-03-18 batch):**
  - Timeout de 10s en gateway.ts evita que la validacion cuelgue.
  - Si el gateway responde 401 (JWT expirado o key rotada), `hooks.server.ts` limpia la cookie y redirige a `/login` automaticamente.
  - Cookie `sda_session` no tiene flag `Secure` cuando `NODE_ENV !== 'production'`, asi que funciona en HTTP local.

#### B4 — Logout / volver al login no funciona -- RESUELTO
- **Sintoma:** No hay forma visible de cerrar sesion o volver al login desde la app.
- **Fix aplicado (2026-03-18 batch):**
  - Boton de logout agregado al sidebar (icono LogOut en la parte inferior, al lado de Settings).
  - El boton llama a `DELETE /api/auth/session` (que limpia la cookie) y redirige a `/login`.
  - Si la llamada falla (gateway caido), igual redirige a `/login` para no dejar al usuario bloqueado.
  - El boton de logout que ya existia en Settings tambien fue mejorado con el mismo patron.

#### B5 — Hover en historial genera storm de requests al gateway -- RESUELTO (previo)
- **Sintoma:** Al pasar el mouse por el panel de historial, SvelteKit prefetchea cada link disparando 3 requests al gateway por cada hover.
- **Fix aplicado:** `data-sveltekit-preload-data="false"` en links de historial.

#### B6 — `settings?/refresh_key` da 500 cuando el gateway esta caido -- RESUELTO (previo)
- **Sintoma:** El boton de regenerar API key da 500, luego SvelteKit lanza 20+ requests de invalidacion que toman 30-50s cada uno.
- **Fix aplicado:** Try/catch en action -> devuelve `fail(503, {...})` en lugar de explotar.

#### B7 — Sin timeout en llamadas BFF->Gateway -- RESUELTO
- **Sintoma:** Cualquier llamada del BFF al gateway podia colgar indefinidamente si el gateway no respondia.
- **Causa raiz:** `gateway.ts` usaba `fetch()` nativo sin `AbortController` ni timeout.
- **Fix aplicado (2026-03-18 batch):**
  - Clase `GatewayError` con `status` y `detail` para distinguir errores de gateway de errores de red.
  - Timeout de 10s (configurable) via `AbortController` en la funcion `gw()` de `gateway.ts`.
  - Si el timeout se dispara, lanza `GatewayError(504, ...)`. Si la red falla, lanza `GatewayError(502, ...)`.

#### B8 — Sin timeout en streaming RAG -- RESUELTO
- **Sintoma:** La llamada de streaming a `/v1/generate` podia colgar indefinidamente.
- **Fix aplicado (2026-03-18 batch):**
  - Timeout de 120s en la conexion inicial al endpoint de streaming (`api/chat/stream/[id]/+server.ts`).
  - Timeout de 10s en llamadas de persistencia de mensajes.

#### B9 — API routes sin error handling -- RESUELTO
- **Sintoma:** Errores del gateway en API routes (`/api/chat/sessions`, `/api/chat/sessions/[id]`, `/api/auth/session`) se propagaban como 500 sin mensaje util.
- **Fix aplicado (2026-03-18 batch):**
  - Try/catch en todas las API routes con `GatewayError` handling.
  - Mensajes de error descriptivos en espanol.

#### B10 — Botones sin loading state -- RESUELTO
- **Sintoma:** Botones de login, refresh key, crear usuario y desactivar usuario no dan feedback visual durante la operacion.
- **Fix aplicado (2026-03-18 batch):**
  - Login: boton muestra "Ingresando..." y inputs se deshabilitan durante submit.
  - Settings (refresh key): boton muestra "Regenerando..." durante submit. Error se muestra en UI.
  - Admin (crear usuario): boton muestra "Creando...", todo el form se deshabilita.
  - Admin (desactivar): boton muestra "Desactivando..." para el usuario especifico.

#### B11 — Sin pagina de error -- RESUELTO
- **Sintoma:** Cuando una load function falla, SvelteKit muestra una pagina en blanco o el error default del framework.
- **Fix aplicado (2026-03-18 batch):**
  - `+error.svelte` en root: pagina fullscreen con codigo de error, mensaje, y botones "Ir al Chat" / "Reintentar".
  - `(app)/+error.svelte` en el layout de app: misma pagina pero dentro del layout con sidebar visible.

#### B12 — JWT expiry no redirige a login -- RESUELTO
- **Sintoma:** Si el JWT expira durante la sesion, las proximas navegaciones fallan silenciosamente con error 401 del gateway.
- **Fix aplicado (2026-03-18 batch):**
  - `hooks.server.ts` intercepta `GatewayError` con status 401.
  - Limpia la cookie `sda_session` y redirige a `/login`.
  - Esto cubre tanto expiry natural del JWT como rotacion de API key en el gateway.

#### B13 — Action `delete` en admin/users sin try/catch -- RESUELTO
- **Sintoma:** Si el gateway no responde al desactivar un usuario, SvelteKit lanza 500.
- **Fix aplicado (2026-03-18 batch):**
  - Try/catch en la action `delete` de `admin/users/+page.server.ts`.
  - Devuelve `fail(503, ...)` con mensaje descriptivo.

#### B14 — Login no diferencia gateway caido vs credenciales invalidas -- RESUELTO
- **Sintoma:** Tanto credenciales invalidas como gateway caido mostraban el mismo mensaje generico.
- **Fix aplicado (2026-03-18 batch):**
  - Login action distingue `GatewayError` 401/403 (credenciales) vs otros errores (servidor no responde).
  - Mensajes distintos: "Email o contrasena incorrectos" vs "El servidor no responde."

### Ideas / Mejoras

- **Circuit breaker:** Si el gateway falla N veces seguidas, dejar de intentar por X segundos en vez de hacer timeout en cada request.
- **Retry con backoff:** Para operaciones idempotentes (GET), reintentar 1-2 veces con backoff exponencial.
- **Health check endpoint:** Polling periodico a `/health` para mostrar banner "servidor no disponible" antes de que el usuario intente navegar.
- **Toast notifications:** Reemplazar mensajes de error inline con toasts para feedback mas visible.

---

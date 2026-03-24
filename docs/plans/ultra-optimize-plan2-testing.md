# Plan: Ultra-Optimize Plan 2 — Testing

> Este documento vive en `docs/plans/ultra-optimize-plan2-testing.md` dentro de la branch `experimental/ultra-optimize`.
> Se actualiza a medida que se completan los tests. Cada tarea completada se marca con fecha.

---

## Contexto

El Plan 1 (ultra-optimize.md) construyó el stack completo: monorepo Turborepo, servidor único Next.js 15, DB Drizzle + SQLite, CLI TypeScript, logger + black box, y GitHub Actions. Todo fue marcado como completado el 2026-03-24.

Este plan verifica que lo que se construyó realmente funciona. No es un test de regresión automatizado completo — es una verificación sistemática y granular de cada capa del sistema, desde los tests unitarios más rápidos hasta los flujos E2E más complejos.

**Lo que se puede probar sin workstation física:** todo lo que está en este plan. Se usa `MOCK_RAG=true` para eliminar la dependencia de los contenedores NVIDIA.

---

## Entorno requerido

```bash
# En WSL2
cd ~/rag-saldivia && git pull
cp .env.example .env.local   # si no existe aún
# Verificar que MOCK_RAG=true en .env.local
node_modules/.bin/next dev /home/enzo/rag-saldivia/apps/web --port 3000

# Abrir http://localhost:3000 → admin@localhost / changeme

# CLI disponible globalmente
cd apps/cli && bun link
```

---

## Seguimiento

Formato de cada tarea: `- [ ] Descripción — estimación`
Al completarla: `- [x] Descripción — completado YYYY-MM-DD`
Cada fase completada genera una entrada en `CHANGELOG.md` antes de hacer commit.

---

## Fase 0 — Preparación del entorno *(15 min)*

Objetivo: confirmar que el entorno está en el estado correcto antes de ejecutar cualquier test. Si algo falla aquí, no tiene sentido avanzar.

- [x] Servidor arranca sin errores en puerto 3000 y `GET http://localhost:3000/api/health` retorna 200 — completado 2026-03-24
- [x] DB tiene seed aplicado: `admin@localhost` existe en la tabla `users` con `role = admin` — completado 2026-03-24
- [x] `rag --version` responde desde cualquier directorio (bun link aplicado en `apps/cli`) — completado 2026-03-24
- [x] `.env.local` tiene `MOCK_RAG=true` y `JWT_SECRET` definido — completado 2026-03-24

Criterio de done: los 4 checks pasan. El servidor está corriendo y la CLI responde.

---

## Fase 1 — Tests unitarios con `bun test` *(30-60 min)*

Objetivo: verificar la lógica pura del sistema sin depender del servidor ni del browser. Son los tests más rápidos y con mayor cobertura de casos edge.

### Fase 1a — Auth *(10 min)*

- [x] `bun test apps/web/src/lib/auth/__tests__/jwt.test.ts` pasa sin errores — completado 2026-03-24 (fix: await import movido al nivel del módulo)
- [x] Verificar que el test de `makeAuthCookie` incluye `Secure` cuando `NODE_ENV=production` — completado 2026-03-24
- [x] Verificar que `verifyJwt` retorna null para token con `exp` en el pasado — completado 2026-03-24

### Fase 1b — RBAC *(10 min)*

- [x] Tests de RBAC en el mismo archivo: admin, area_manager y user — los 3 pasan — completado 2026-03-24
- [x] Agregar test: `getRequiredRole("/api/admin/users")` retorna `"admin"` — completado 2026-03-24
- [x] Agregar test: `getRequiredRole("/chat")` retorna `null` — completado 2026-03-24
- [x] Agregar test: `canAccessRoute` con `area_manager` en ruta `/audit` retorna `true` — completado 2026-03-24

### Fase 1c — DB queries *(15 min)*

- [x] Crear `packages/db/src/__tests__/users.test.ts`: `createUser`, `verifyPassword`, `listUsers`, `updateUser`, `deleteUser` — completado 2026-03-24
- [x] Verificar que `verifyPassword` retorna `null` para password incorrecta — completado 2026-03-24
- [x] Verificar que `createUser` con email duplicado lanza error con mensaje descriptivo — completado 2026-03-24
- [x] Verificar que `deleteUser` elimina también las filas en `user_areas` — completado 2026-03-24

### Fase 1d — Config loader *(5 min)*

- [ ] `packages/config` carga los YAMLs de `config/` sin errores
- [ ] Acceder a un campo inexistente retorna el default tipado (no `undefined`)

### Fase 1e — Logger + Black box *(15 min)*

- [x] `packages/logger/backend.ts`: `log.info`, `log.warn`, `log.error` no lanzan excepciones — completado 2026-03-24
- [x] En `NODE_ENV=production` el output es JSON válido con campos `level`, `event`, `ts` — completado 2026-03-24 (testeado via formatJson directo; logger captura isDev al import)
- [x] `packages/logger/blackbox.ts`: `reconstructFromEvents([])` retorna array vacío sin error — completado 2026-03-24
- [x] `reconstructFromEvents` con 5 eventos los ordena por timestamp correctamente — completado 2026-03-24
- [x] `formatTimeline` produce un string no vacío con los eventos — completado 2026-03-24

Criterio de done: `bun test` corre todas las suites sin fallos. 0 errores, 0 tests skipped no intencionalmente.
**Estado: 57/57 tests pasan — Fase 1 completada 2026-03-24 (salvo Fase 1d pendiente)**

---

## Fase 2 — Tests de API con HTTP *(1-2 hs)*

Objetivo: verificar cada endpoint REST directamente, sin browser. Se usa `curl` o el script `scripts/test-api.sh`. Se prueba el happy path y los casos de error más importantes.

### Fase 2a — Auth endpoints

- [ ] `POST /api/auth/login` con credenciales válidas retorna 200, body `{ ok: true }` y cookie `auth_token` en `Set-Cookie`
- [ ] `POST /api/auth/login` con password incorrecta retorna 401 con `{ ok: false, error: "..." }`
- [ ] `POST /api/auth/login` con body malformado retorna 400 con detalles de validación Zod
- [ ] `POST /api/auth/login` con usuario inactivo retorna 403 con mensaje descriptivo
- [ ] `POST /api/auth/logout` con token válido retorna 200 y `Set-Cookie` con `Max-Age=0`
- [ ] `POST /api/auth/refresh` con cookie válida retorna nuevo token

### Fase 2b — RAG endpoints

- [ ] `GET /api/rag/collections` sin token retorna 401
- [ ] `GET /api/rag/collections` con token válido retorna 200 con lista de colecciones mock
- [ ] `POST /api/rag/generate` con query válida inicia stream SSE (Content-Type: `text/event-stream`)
- [ ] `POST /api/rag/generate` sin body retorna 400

### Fase 2c — Upload e ingesta

- [ ] `POST /api/upload` sin token retorna 401
- [ ] `POST /api/upload` con token de usuario normal y archivo PDF retorna 200 y `jobId`
- [ ] `GET /api/admin/ingestion` sin token retorna 401
- [ ] `GET /api/admin/ingestion` con token de usuario normal retorna 403
- [ ] `GET /api/admin/ingestion` con token de admin retorna 200 con lista de jobs
- [ ] `DELETE /api/admin/ingestion/999` con token de admin y ID inexistente retorna 404

### Fase 2d — Audit y black box

- [ ] `GET /api/audit` con token de usuario normal retorna 403
- [ ] `GET /api/audit` con token de area_manager retorna 200 con array de eventos
- [ ] `GET /api/audit?type=auth.login&limit=5` retorna máximo 5 eventos del tipo indicado
- [ ] `GET /api/audit/replay` con token de admin retorna timeline reconstruido
- [ ] `GET /api/audit/export` con token de admin retorna JSON descargable

### Fase 2e — Infraestructura

- [ ] `GET /api/health` retorna 200 con status de todos los servicios
- [ ] `POST /api/log` sin token retorna 200 (ruta pública para frontend logs)

Criterio de done: todos los endpoints responden con el código HTTP correcto. Los errores de autenticación retornan 401, los de permisos 403, los de validación 400.

---

## Fase 3 — Tests de UI manual en browser *(2-3 hs)*

Objetivo: verificar cada flujo de usuario en `http://localhost:3000`. Se testea el happy path y los casos de error visibles al usuario.

### Fase 3a — Login y Logout *(15 min)*

- [ ] Navegar a `http://localhost:3000` sin sesión redirige a `/login`
- [ ] Login con `admin@localhost / changeme` redirige a `/` y muestra nombre del usuario
- [ ] La cookie `auth_token` aparece en DevTools → Application → Cookies con flag `HttpOnly`
- [ ] Logout redirige a `/login` y la cookie desaparece
- [ ] Intentar acceder a `/admin/users` sin sesión redirige a `/login?from=/admin/users`

### Fase 3b — Chat *(45 min)*

- [ ] Crear nueva sesión desde la página principal — aparece en la lista
- [ ] Enviar un mensaje de texto — respuesta mock aparece en streaming (texto animado)
- [ ] El historial de mensajes persiste al recargar la página
- [ ] Crear una segunda sesión — las dos aparecen en la lista con nombres distintos
- [ ] Renombrar una sesión — el nombre actualiza en la lista inmediatamente
- [ ] Eliminar una sesión — desaparece de la lista con confirmación
- [ ] Dar feedback positivo a un mensaje (like) — el ícono cambia de estado
- [ ] Dar feedback negativo a un mensaje — el ícono cambia de estado

### Fase 3c — Upload *(20 min)*

- [ ] Navegar a `/upload` — aparece zona de drag & drop
- [ ] Arrastrar un PDF a la zona — se muestra nombre y tamaño del archivo
- [ ] Confirmar upload — el job aparece en `/admin/ingestion` con status `pending`
- [ ] Intentar subir un archivo que no sea PDF — error de validación visible

### Fase 3d — Admin: Usuarios *(30 min)*

- [ ] `/admin/users` lista todos los usuarios del seed con email, rol y estado
- [ ] Crear usuario con wizard: email, nombre, password, rol `user`, sin áreas — aparece en la lista
- [ ] Crear usuario con áreas asignadas — la relación persiste al recargar
- [ ] Cambiar el rol de un usuario de `user` a `area_manager` — el cambio persiste
- [ ] Desactivar un usuario — aparece como inactivo en la lista
- [ ] Activar el mismo usuario — vuelve a activo
- [ ] Eliminar el usuario creado — desaparece de la lista con confirmación

### Fase 3e — Admin: Áreas *(20 min)*

- [ ] `/admin/areas` lista las áreas del seed
- [ ] Crear un área nueva — aparece en la lista
- [ ] Asignar una colección al área con nivel `read` — aparece en la lista de colecciones del área
- [ ] Cambiar nivel a `write` — persiste al recargar
- [ ] Eliminar el área creada — pide confirmación si tiene usuarios asignados

### Fase 3f — Admin: Config RAG *(15 min)*

- [ ] `/admin/config` muestra los sliders con sus valores actuales
- [ ] Mover el slider de `top_k` — el valor actualiza en tiempo real
- [ ] Guardar — el cambio persiste al recargar la página
- [ ] Toggle `reranker` de on a off — persiste
- [ ] Reset a defaults — todos los valores vuelven al estado inicial

### Fase 3g — Admin: Stats del sistema *(10 min)*

- [ ] `/admin/stats` muestra cards con: usuarios totales, sesiones activas, jobs en cola, colecciones
- [ ] Botón "Refresh" actualiza los números sin recargar la página completa

### Fase 3h — Settings *(20 min)*

- [ ] `/settings` muestra el nombre y email del usuario actual
- [ ] Cambiar el nombre — persiste al recargar y aparece en el header
- [ ] Cambiar la contraseña — el nuevo password funciona en el próximo login
- [ ] Intentar cambiar la contraseña con el actual incorrecto — error visible
- [ ] Cambiar preferencias RAG (temperatura, top_k) — persiste al recargar

### Fase 3i — Audit Log UI *(10 min)*

- [ ] `/audit` muestra tabla de eventos con columnas: timestamp, tipo, usuario, detalles
- [ ] Filtrar por tipo `auth.login` — solo muestra eventos de login
- [ ] Los eventos generados en los tests anteriores (crear usuario, login, etc.) aparecen en la tabla

Criterio de done: todos los flujos visibles al usuario funcionan sin errores en consola del browser. El historial persiste entre recargas.

---

## Fase 4 — Tests de CLI *(1-2 hs)*

Objetivo: verificar que todos los comandos del CLI producen output correcto con el servidor corriendo.

### Fase 4a — Sistema

- [ ] `rag status` muestra semáforo con colores: Next.js verde, RAG en estado mock, DB verde
- [ ] `rag status` incluye latencias en ms para cada servicio

### Fase 4b — Usuarios

- [ ] `rag users list` muestra tabla con ID, email, nombre, rol, activo
- [ ] `rag users create` abre wizard interactivo y crea el usuario correctamente
- [ ] `rag users delete <id>` pide confirmación y elimina el usuario
- [ ] `rag users delete <id>` con ID inexistente muestra error descriptivo

### Fase 4c — Colecciones

- [ ] `rag collections list` muestra tabla con colecciones del RAG mock
- [ ] `rag collections create` crea una colección nueva
- [ ] `rag collections delete <name>` pide confirmación y elimina

### Fase 4d — Ingesta

- [ ] `rag ingest status` muestra tabla de jobs con ID, archivo, status, creado
- [ ] `rag ingest status` con `--status pending` filtra correctamente
- [ ] `rag ingest cancel <id>` cancela el job con confirmación

### Fase 4e — Config

- [ ] `rag config get` muestra todos los parámetros de config con sus valores actuales
- [ ] `rag config get top_k` muestra solo ese parámetro
- [ ] `rag config set top_k 10` actualiza el valor y lo confirma en output
- [ ] `rag config reset` vuelve todos los valores a defaults con confirmación

### Fase 4f — Audit

- [ ] `rag audit log` muestra tabla de eventos recientes
- [ ] `rag audit log --type auth.login --limit 5` filtra y limita correctamente
- [ ] `rag audit replay` muestra timeline reconstruido en texto
- [ ] `rag audit export` genera un archivo `.json` con todos los eventos

### Fase 4g — DB y setup

- [ ] `rag db seed` re-aplica el seed sin romper datos existentes
- [ ] `rag db migrate` corre migraciones pendientes (o dice que no hay ninguna)

### Fase 4h — REPL interactivo

- [ ] `rag` sin argumentos abre el selector interactivo con @clack/prompts
- [ ] Seleccionar "Status" desde el REPL ejecuta `rag status`
- [ ] Seleccionar "Salir" cierra el REPL limpiamente

Criterio de done: todos los comandos producen output con colores y tablas. Ningún comando lanza stack trace sin manejar.

---

## Fase 5 — Tests del black box *(30-45 min)*

Objetivo: verificar que el sistema de logging y replay funciona correctamente y puede reconstruir lo que pasó.

- [ ] Ejecutar 10 acciones variadas: login, crear usuario, subir archivo, cambiar config, chat, logout — en ese orden
- [ ] Verificar en la tabla `events` de SQLite que existen 10+ registros con los tipos correctos
- [ ] `rag audit log --limit 10` muestra esas acciones en orden cronológico inverso
- [ ] `rag audit replay` reconstruye el timeline mostrando las 10 acciones en orden
- [ ] El replay incluye el `userId` en cada evento relevante
- [ ] `rag audit export` genera un archivo JSON con todos los eventos y es válido (`JSON.parse` no lanza error)
- [ ] Forzar un error 500: enviar query con body inválido a `/api/rag/generate` y verificar que el evento `system.error` aparece en el audit log
- [ ] `packages/logger/rotation.ts`: crear un log de más de 10MB no rompe el servidor — el archivo rota correctamente

Criterio de done: después de simular las acciones, `rag audit replay` reconstruye el timeline completo. Cada acción tiene su evento correspondiente.

---

## Fase 6 — Tests de seguridad y RBAC *(1 hs)*

Objetivo: verificar que ningún bypass de autenticación o autorización es posible.

### Fase 6a — Sin autenticación

- [ ] `GET /chat` sin cookie → redirect a `/login?from=/chat`
- [ ] `GET /admin/users` sin cookie → redirect a `/login?from=/admin/users`
- [ ] `GET /api/rag/collections` sin token → 401 `{ ok: false, error: "No autenticado" }`
- [ ] `GET /api/admin/ingestion` sin token → 401

### Fase 6b — Token inválido

- [ ] Cookie `auth_token=garbage` → 401 en API, redirect en página
- [ ] Token bien formado pero firmado con secret distinto → 401
- [ ] Token bien formado pero con `exp` en el pasado → 401

### Fase 6c — Rol insuficiente (usuario con rol `user`)

- [ ] `GET /api/admin/ingestion` con token de usuario normal → 403 `{ ok: false, error: "Acceso denegado — se requiere rol admin" }`
- [ ] `GET /api/audit` con token de usuario normal → 403
- [ ] Navegar a `/admin/users` con sesión de usuario normal → redirect a `/`
- [ ] Navegar a `/audit` con sesión de usuario normal → redirect a `/`

### Fase 6d — Rol insuficiente (area_manager)

- [ ] `GET /api/admin/ingestion` con token de area_manager → 403
- [ ] `GET /api/audit` con token de area_manager → 200 (tiene acceso)
- [ ] Navegar a `/audit` con sesión de area_manager → accede correctamente
- [ ] Navegar a `/admin/users` con sesión de area_manager → redirect a `/`

### Fase 6e — Cuenta desactivada

- [ ] Login con usuario desactivado → 403 con `"Cuenta desactivada. Contactá al administrador."`
- [ ] Token válido de un usuario que se desactiva después de login → servidor debe verificar estado activo en cada request crítico

Criterio de done: ningún caso de acceso denegado retorna 200. Todos los redirects van a la ruta correcta.

---

## Fase 7 — Tests de integración E2E *(1-2 hs)*

Objetivo: verificar flujos completos de punta a punta que cruzan múltiples capas del sistema.

### Flujo 1: Nuevo colaborador

- [ ] `bun run setup` desde un directorio limpio completa sin errores
- [ ] El servidor arranca después del setup sin configuración adicional
- [ ] Login con el usuario admin del seed funciona inmediatamente
- [ ] El chat con `MOCK_RAG=true` responde sin errores

### Flujo 2: Admin crea usuario con acceso a colección

- [ ] Admin crea área "Legales" con colección `rag-collection-1` con nivel `read`
- [ ] Admin crea usuario `legales@localhost` con rol `user` asignado al área "Legales"
- [ ] Logout del admin
- [ ] Login con `legales@localhost` — accede correctamente
- [ ] En el chat, la lista de colecciones disponibles incluye `rag-collection-1`
- [ ] El usuario no puede acceder a `/admin/users` — redirect a `/`
- [ ] El usuario no puede acceder a `/audit` — redirect a `/`

### Flujo 3: Ingesta completa

- [ ] Subir un PDF desde la UI de upload
- [ ] El job aparece en `/admin/ingestion` con status `pending`
- [ ] El worker de ingesta (`apps/web/src/workers/ingestion.ts`) procesa el job
- [ ] El status cambia de `pending` a `processing` y luego a `completed`
- [ ] El evento `ingestion.completed` aparece en el audit log

### Flujo 4: Recuperación tras caída del servidor

- [ ] Subir un PDF — job queda en estado `processing` con `locked_at` seteado
- [ ] Reiniciar el servidor (Ctrl+C y volver a levantar)
- [ ] El worker detecta que `locked_at` tiene más de 5 minutos → libera el lock
- [ ] El job vuelve a estado `pending` y el worker lo retoma

### Flujo 5: Black box replay tras error

- [ ] Realizar un flujo normal: login → chat → logout
- [ ] Provocar un error deliberado (body inválido en `/api/rag/generate`)
- [ ] `rag audit replay` muestra el flujo completo incluyendo el error
- [ ] El error tiene mensaje descriptivo con sugerencia de resolución

Criterio de done: los 5 flujos completan sin intervención manual. El sistema se recupera solo de la caída del servidor.

---

## Resultado esperado al completar el plan

- Cobertura verificada de los 8 endpoints de auth, RAG, upload y audit
- 14 comandos de CLI probados con output correcto
- 9 casos de RBAC denegado verificados sin bypasses
- 5 flujos E2E completos funcionando
- El black box puede reconstruir cualquier sesión de trabajo
- Confianza para probar contra la workstation física con `MOCK_RAG=false`

## Tiempo total estimado: 8-12 horas de trabajo

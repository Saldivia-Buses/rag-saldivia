# Design Doc — Fase 8: Settings Pro

**Fecha:** 2026-03-23
**Proyecto:** RAG Saldivia — SDA Frontend
**Spec fuente:** `docs/superpowers/specs/2026-03-19-sda-super-app-design.md` §Fase 8

---

## Objetivo

Personalización completa del usuario: perfil con avatar y color, idioma UI, cambio de contraseña, preferencias RAG persistidas en DB, y notificaciones in-app de ingesta y alertas del sistema.

---

## Decisiones de diseño

| Decisión | Elección | Motivo |
|----------|----------|--------|
| Storage de preferencias | Columna JSON `preferences TEXT DEFAULT '{}'` en `users` | No se necesita query por campo; schema limpio sin tabla extra |
| Avatar | Inicial del nombre + color picker (sin upload) | Sin complejidad de file storage; suficiente para distinguir usuarios |
| Idioma | Almacenar preferencia `ui_language: "es"\|"en"` | i18n completo es out of scope; la preferencia queda en DB para Fase futura |
| Notificaciones | In-app toasts desde `+layout.server.ts` | Sin WebSocket ni email; chequeo pasivo en cada carga de página |
| Password change | Requiere contraseña actual + verificación bcrypt | Seguridad mínima; sin token de reset en esta fase |

---

## Sección 1 — Backend

### `saldivia/auth/database.py`

**1. Migration — nueva columna:**
```sql
ALTER TABLE users ADD COLUMN preferences TEXT NOT NULL DEFAULT '{}';
```

**2. Nuevos métodos:**
- `get_user_preferences(user_id: int) -> dict` — SELECT + json.loads
- `update_user_preferences(user_id: int, prefs: dict)` — merge con prefs existentes + json.dumps
- `update_user_name(user_id: int, name: str)` — UPDATE directo
- `update_user_password(user_id: int, current_pw: str, new_pw: str) -> bool` — verifica bcrypt(current) antes de actualizar

**3. Shape del JSON de preferencias:**
```json
{
  "default_collection": "",
  "default_query_mode": "standard",
  "vdb_top_k": 10,
  "reranker_top_k": 5,
  "max_sub_queries": 4,
  "follow_up_retries": true,
  "show_decomposition": false,
  "avatar_color": "#6366f1",
  "ui_language": "es",
  "notify_ingestion_done": true,
  "notify_system_alerts": true
}
```

### `saldivia/gateway.py`

- `GET /auth/me/preferences` — retorna prefs del usuario autenticado
- `PATCH /auth/me/preferences` — merge parcial de prefs (body: dict parcial)
- `PATCH /auth/me/profile` — actualiza nombre (body: `{name: str}`)
- `PATCH /auth/me/password` — cambia contraseña (body: `{current_password, new_password}`); 400 si incorrecta

---

## Sección 2 — BFF (SvelteKit)

### `src/lib/types/preferences.ts` (nuevo)

Interfaz `UserPreferences` compartida entre client y server:
```typescript
export interface UserPreferences {
  default_collection: string;
  default_query_mode: 'standard' | 'crossdoc';
  vdb_top_k: number;
  reranker_top_k: number;
  max_sub_queries: number;
  follow_up_retries: boolean;
  show_decomposition: boolean;
  avatar_color: string;
  ui_language: 'es' | 'en';
  notify_ingestion_done: boolean;
  notify_system_alerts: boolean;
}

export const DEFAULT_PREFERENCES: UserPreferences = {
  default_collection: '',
  default_query_mode: 'standard',
  vdb_top_k: 10,
  reranker_top_k: 5,
  max_sub_queries: 4,
  follow_up_retries: true,
  show_decomposition: false,
  avatar_color: '#6366f1',
  ui_language: 'es',
  notify_ingestion_done: true,
  notify_system_alerts: true,
};
```

### `src/lib/server/gateway.ts` (modificar)

Agregar:
- `gatewayGetPreferences(userId: number): Promise<UserPreferences>`
- `gatewayUpdatePreferences(userId: number, prefs: Partial<UserPreferences>): Promise<void>`
- `gatewayUpdateProfile(userId: number, name: string): Promise<void>`
- `gatewayUpdatePassword(userId: number, currentPw: string, newPw: string): Promise<void>`

### `src/routes/(app)/+layout.server.ts` (modificar)

Agregar al `load()`:
- `gatewayGetPreferences(userId)` — pasa `preferences` a todos los layouts hijos
- Chequeo de notificaciones pendientes: si `notify_ingestion_done` activo, incluir jobs completados recientes; si `notify_system_alerts` activo, incluir alertas no resueltas → array `notifications` en data

### `src/routes/(app)/settings/+page.server.ts` (modificar)

Agregar actions:
- `update_preferences` → `PATCH /auth/me/preferences`
- `update_profile` → `PATCH /auth/me/profile`
- `update_password` → `PATCH /auth/me/password`; error tipado si 400

---

## Sección 3 — Frontend

### `src/lib/stores/preferences.svelte.ts` (nuevo)

Store reactivo con las prefs del usuario. Inicializado desde `data.preferences` en el layout. Usado por:
- Avatar (color + inicial)
- Idioma UI
- CrossdocStore (hidratación inicial)

### `src/lib/stores/crossdoc.svelte.ts` (modificar)

`CrossdocStore` acepta `initialOptions?: Partial<CrossdocOptions>` en el constructor para hidratarse desde las prefs del servidor.

### `src/routes/(app)/+layout.svelte` (modificar)

- Inicializa `preferencesStore` con `data.preferences`
- Renderiza notificaciones pendientes como toasts al montar (via `toastStore`)
- Pasa `data.preferences` al `CrossdocStore`

### `src/routes/(app)/settings/+page.svelte` (reescribir)

5 secciones colapsables:

**1. Perfil**
- Avatar: círculo con inicial del nombre + `<input type="color">` para `avatar_color`
- Input nombre (editable), email + role (read-only)
- Selector idioma: radio es/en
- Botón "Guardar perfil" → action `update_profile` + `update_preferences` (color + idioma)

**2. Contraseña**
- 3 inputs: contraseña actual, nueva, confirmar nueva
- Validación client-side: nueva === confirmar
- Botón "Cambiar contraseña" → action `update_password`
- Error visible inline si 400 (contraseña actual incorrecta)

**3. Preferencias RAG**
- `default_collection`: `<select>` con colecciones disponibles (cargadas en load)
- `default_query_mode`: radio standard/crossdoc
- `vdb_top_k` / `reranker_top_k` / `max_sub_queries`: `<input type="number">` con rangos
- `follow_up_retries` / `show_decomposition`: toggles
- Botón "Guardar preferencias" → action `update_preferences`

**4. Notificaciones**
- Toggle "Notificarme cuando termine una ingesta" → `notify_ingestion_done`
- Toggle "Notificarme sobre alertas del sistema" → `notify_system_alerts`
- Botón "Guardar" → action `update_preferences`

**5. API Key + Sesión**
- Sin cambios respecto a la implementación actual

---

## Tests requeridos

| Test | Qué verifica |
|------|-------------|
| `test_get_user_preferences` | Retorna defaults para usuario sin prefs |
| `test_update_user_preferences_merge` | Merge parcial preserva campos no incluidos |
| `test_update_user_name` | UPDATE refleja en get |
| `test_update_user_password_correct` | Actualiza cuando current_pw es correcto |
| `test_update_user_password_wrong` | Retorna False cuando current_pw es incorrecto |
| `test_gateway_get_preferences` | GET 200 con prefs del usuario |
| `test_gateway_patch_preferences` | PATCH 200 actualiza prefs |
| `test_gateway_patch_profile` | PATCH 200 actualiza nombre |
| `test_gateway_patch_password_wrong` | PATCH 400 si contraseña incorrecta |
| `test_bff_update_preferences` | POST action → gateway PATCH |
| `test_bff_update_password_error` | 400 gateway → error en form |

---

## Archivos afectados

```
saldivia/auth/database.py                                      — migration + 4 métodos
saldivia/gateway.py                                            — 4 nuevos endpoints
saldivia/tests/test_fase8.py                                   — tests backend

services/sda-frontend/src/lib/types/preferences.ts             — nuevo
services/sda-frontend/src/lib/stores/preferences.svelte.ts     — nuevo
services/sda-frontend/src/lib/stores/crossdoc.svelte.ts        — hidratación desde prefs
services/sda-frontend/src/lib/server/gateway.ts                — 4 nuevas funciones
services/sda-frontend/src/routes/(app)/+layout.server.ts       — prefs + notificaciones
services/sda-frontend/src/routes/(app)/+layout.svelte          — init stores + toasts
services/sda-frontend/src/routes/(app)/settings/+page.server.ts — 3 nuevas actions
services/sda-frontend/src/routes/(app)/settings/+page.svelte   — reescritura completa
```

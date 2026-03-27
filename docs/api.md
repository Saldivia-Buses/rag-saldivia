# Referencia HTTP — RAG Saldivia

Base URL de desarrollo: `http://localhost:3000`.

Salvo rutas marcadas como **públicas**, las peticiones requieren cookie `auth_token` (JWT) o `Authorization: Bearer <jwt>`. Las rutas **admin** exigen rol `admin` o `SYSTEM_API_KEY` como token Bearer.

---

## 1. Autenticación

### `POST /api/auth/login` — **público**

| Campo | Tipo | Obligatorio |
|-------|------|-------------|
| `email` | string | sí |
| `password` | string | sí |

**Respuesta exitosa (200):** `{ ok: true, user: { … } }` y cookie `auth_token` HttpOnly.

**Errores:** `400` validación, `401` credenciales, `403` cuenta desactivada.

### `DELETE /api/auth/logout`

Invalida el JWT (Redis) y limpia la cookie.

### `POST /api/auth/refresh` — **público**

Renueva el token; ver implementación en `apps/web/src/app/api/auth/refresh/route.ts`.

---

## 2. Endpoints por grupo

### Auth (resumen)

| Método | Ruta | Auth | Descripción |
|--------|------|------|-------------|
| POST | `/api/auth/login` | ninguna | Login |
| DELETE | `/api/auth/logout` | usuario | Logout |
| POST | `/api/auth/refresh` | ninguna | Refresh |

### RAG

| Método | Ruta | Auth | Descripción |
|--------|------|------|-------------|
| POST | `/api/rag/generate` | usuario | SSE — generación RAG (stream) |
| GET | `/api/rag/collections` | usuario | Lista colecciones (filtrada por permisos) |
| POST | `/api/rag/collections` | admin | Crear colección upstream |
| DELETE | `/api/rag/collections/[name]` | admin | Eliminar colección |
| POST | `/api/rag/suggest` | usuario | Preguntas sugeridas (follow-up) |
| GET | `/api/rag/document/[name]` | usuario | Proxy a documento/PDF upstream |

### Upload

| Método | Ruta | Auth | Body | Respuesta |
|--------|------|------|------|-----------|
| POST | `/api/upload` | usuario | `multipart/form-data`: `file`, `colección` | `{ ok, jobId?, … }` |

### Search

| Método | Ruta | Auth | Query |
|--------|------|------|-------|
| GET | `/api/search` | usuario | `q` (mín. 2 caracteres útiles) |

### Admin — usuarios

| Método | Ruta | Auth |
|--------|------|------|
| GET | `/api/admin/users` | admin |
| POST | `/api/admin/users` | admin |
| PATCH | `/api/admin/users/[id]` | admin |
| DELETE | `/api/admin/users/[id]` | admin |
| GET/POST/DELETE | `/api/admin/users/[id]/areas` | admin |

### Admin — áreas

| Método | Ruta | Auth |
|--------|------|------|
| GET | `/api/admin/areas` | admin |
| POST | `/api/admin/areas` | admin |
| PATCH | `/api/admin/areas/[id]` | admin |
| DELETE | `/api/admin/areas/[id]` | admin |

### Admin — permisos (colección × área)

| Método | Ruta | Auth | Body |
|--------|------|------|------|
| POST | `/api/admin/permissions` | admin | `areaId`, `collectionName`, `permission` |
| DELETE | `/api/admin/permissions` | admin | `areaId`, `collectionName` |

### Admin — configuración RAG

| Método | Ruta | Auth |
|--------|------|------|
| GET | `/api/admin/config` | admin |
| PATCH | `/api/admin/config` | admin |
| POST | `/api/admin/config/reset` | admin |

### Admin — plantillas

| Método | Ruta | Auth |
|--------|------|------|
| GET | `/api/admin/templates` | usuario autenticado |
| POST | `/api/admin/templates` | admin |
| DELETE | `/api/admin/templates` | admin (query `?id=`) |

### Admin — ingesta

| Método | Ruta | Auth | Notas |
|--------|------|------|-------|
| GET | `/api/admin/ingestion` | usuario | Jobs BullMQ + SQLite |
| DELETE | `/api/admin/ingestion/[id]` | dueño o admin | Cancelar job |
| PATCH | `/api/admin/ingestion/[id]` | ver ruta | Actualización de estado si aplica |
| GET | `/api/admin/ingestion/stream` | usuario | **SSE** — eventos de cola |

### Admin — analytics y DB

| Método | Ruta | Auth |
|--------|------|------|
| GET | `/api/admin/analytics` | admin |
| POST | `/api/admin/db/reset` | admin / sistema |
| POST | `/api/admin/db/seed` | admin |
| POST | `/api/admin/db/migrate` | admin |

### Notificaciones

| Método | Ruta | Auth |
|--------|------|------|
| GET | `/api/notifications` | usuario |
| GET | `/api/notifications/stream` | usuario | **SSE** |

### Colecciones (app)

| Método | Ruta | Auth |
|--------|------|------|
| GET | `/api/collections/[name]/embeddings` | usuario | Grafo/similitud |
| GET | `/api/collections/[name]/history` | usuario | Historial de colección |

### Extract y health

| Método | Ruta | Auth |
|--------|------|------|
| POST | `/api/extract` | usuario | Extracción estructurada |
| GET | `/api/health` | **público** | Health check |

### Integraciones

| Método | Ruta | Auth |
|--------|------|------|
| POST | `/api/slack` | según implementación |
| POST | `/api/teams` | según implementación |

### Log y auditoría

| Método | Ruta | Auth |
|--------|------|------|
| POST | `/api/log` | **público** (eventos frontend) |
| GET | `/api/audit` | area_manager o admin |
| GET | `/api/audit/export` | según ruta |
| GET | `/api/audit/replay` | según ruta |

---

## 3. SSE (Server-Sent Events)

Rutas que devuelven stream `text/event-stream`:

- `POST /api/rag/generate`
- `GET /api/admin/ingestion/stream`
- `GET /api/notifications/stream`

**Ejemplo (browser):**

```javascript
const es = new EventSource("/api/notifications/stream", {
  withCredentials: true, // cookie HttpOnly
})
es.onmessage = (ev) => { console.log(ev.data) }
es.onerror = () => { es.close() }
```

Para `generate`, el cliente suele usar `fetch` + lectura del body como stream (ver `apps/web/src/lib/rag/stream.ts`).

---

## 4. Errores comunes

| Código | Causa típica |
|--------|----------------|
| 401 | Sin JWT o token inválido/expirado/revocado |
| 403 | Rol insuficiente o recurso de otro usuario |
| 404 | Recurso inexistente |
| 409 | Conflicto (p. ej. email duplicado) |
| 500 | Error interno o upstream |

Cuerpo JSON frecuente: `{ ok: false, error: string }`.

---

## 5. Notas RAG

Los endpoints que llaman al RAG Server (`/api/rag/*`, partes de `extract`, `upload`, etc.) requieren el servicio en `RAG_SERVER_URL` salvo **`MOCK_RAG=true`**, donde las respuestas pueden ser simuladas.

La ingesta vectorial completa en Milvus y el LLM en GPU están pensadas para el despliegue con **NVIDIA RAG Blueprint**; en desarrollo local usá mock cuando no tengas el stack.

---

## 6. Lo que no es REST en `/api`

- **Memoria del usuario** (`/settings/memory`): persistida vía Server Actions, no hay `/api/memory` en esta versión del árbol `app/api`.
- **Share de sesión**: página pública en `/share/[token]`; no hay `GET /api/share/[token]`.

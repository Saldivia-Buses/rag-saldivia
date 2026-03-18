# SDA Frontend — Diseño Completo

**Fecha:** 2026-03-17
**Estado:** APROBADO
**Proyecto:** rag-saldivia (`~/rag-saldivia/`)
**Tipo:** Overlay sobre NVIDIA RAG Blueprint v2.5.0

---

## 1. Contexto y objetivo

**SDA** es el nombre definitivo de la plataforma. Es la interfaz web empresarial para el sistema RAG de Tecpia. Permite a los usuarios de la empresa consultar documentos técnicos industriales (catálogos, manuales, specs) usando lenguaje natural, con respuestas citadas y verificables.

### Principios de diseño
- **Plataforma extensible**, no solo un frontend de RAG. Nuevos módulos se suman como rutas + layout, sin reescribir.
- **API key nunca toca el browser.** El servidor SvelteKit (BFF) hace todas las llamadas al gateway con el API key del sistema.
- **Identidad industrial.** Dark navy + indigo. Sidebar icon-only. Sensación de herramienta profesional, no de demo.
- **Usuarios técnicos primero.** Split view con fuentes visibles. Datos verificables, no respuestas mágicas.

### Qué NO es
- No es una app de cara al público — es interna.
- No requiere SEO, SSG, ni RSC.
- No reemplaza ni modifica el NVIDIA RAG Blueprint — solo lo consume.

---

## 2. Arquitectura general

```
Browser (cookie httpOnly JWT)
    │
    │  HTTPS :3000
    ▼
SvelteKit 5 — adapter-node (Docker)
  ├── (auth)/login        ← sin sidebar, sin JWT requerido
  └── (app)/              ← layout con sidebar, requiere JWT
      ├── chat/[id]
      ├── collections/[name]
      ├── admin/users
      ├── admin/areas/[id]
      ├── audit
      └── settings
      api/                ← server endpoints BFF
    │
    │  Bearer API key (server-side, invisible al browser)
    │  Internal :8090
    ▼
FastAPI Gateway (extendido)
  ├── /auth/session       ← nuevo
  ├── /admin/users        ← nuevo
  ├── /admin/areas        ← nuevo
  ├── /v1/generate        ← existente
  ├── /v1/search          ← existente
  └── /v1/collections     ← existente + stats nuevo
    │
    ├── SQLite AuthDB (users, areas, permissions, audit)
    ├── Redis (ingestion queue, job status)
    ├── RAG Server :8081 (SSE streaming)
    └── Ingestor :8082 (PDF pipeline)
```

### Flujo de autenticación

**Capas de auth separadas:**
- **Browser ↔ SvelteKit:** JWT en cookie httpOnly (identifica al usuario)
- **SvelteKit ↔ Gateway:** Bearer API key del sistema (variable de entorno `SYSTEM_API_KEY` en el container SvelteKit)

El gateway **no cambia su mecanismo de auth** para los endpoints RAG existentes. Todos los endpoints (existentes y nuevos) reciben el Bearer API key del sistema. El JWT del usuario sirve para que SvelteKit sepa *quién* hace la consulta y aplique RBAC en el BFF antes de forwarding.

**Pasos:**
1. Usuario ingresa email + password en `/login`
2. SvelteKit hace `POST /auth/session` al gateway (con el Bearer `SYSTEM_API_KEY`)
3. Gateway valida credenciales en SQLite AuthDB, devuelve JWT (payload: `user_id · email · role · area_id · exp`)
4. SvelteKit setea cookie `httpOnly; Secure; SameSite=Strict` con el JWT — expira en 8h
5. Todas las páginas protegidas leen la cookie en el `hooks.server.ts`; redirigen a `/login` si ausente o expirada
6. Las `api/` routes del BFF leen la cookie, verifican el JWT localmente (con `JWT_SECRET`), aplican RBAC, y hacen las llamadas al gateway usando `SYSTEM_API_KEY` como Bearer token

**Variable de entorno adicional:** `SYSTEM_API_KEY` — API key del sistema con permisos completos, usada por el BFF para todas las llamadas al gateway. Solo vive en el container SvelteKit.

### ERP integration (futuro)
El gateway implementa una interfaz `AuthProvider`. En v1 usa `LocalAuthProvider` (SQLite). Cuando el ERP de la empresa tenga su API REST disponible, se swapea por `ERPAuthProvider` sin tocar nada del frontend.

---

## 3. Stack de tecnología

| Capa | Tecnología | Razón |
|------|-----------|-------|
| Framework | SvelteKit 5 + Svelte 5 | 1,200 RPS vs 850 Next.js, 65% menos bundle, SSE nativo para streaming RAG |
| Componentes | Shadcn-Svelte | Accesibles, dark mode nativo, completamente customizables |
| CSS | Tailwind CSS 4 | CSS-first config, build más rápido, utility-first |
| Íconos | Lucide-Svelte | SVG tree-shakeable, consistentes, limpios |
| Dark mode | mode-watcher | SSR-safe, sin flash en carga, estándar en SvelteKit |
| Runtime | Node.js 22 + adapter-node | Server-side rendering + BFF + Docker |
| Build | Docker multi-stage | Builder (node:22-alpine) → runtime ~100 MB |

### Por qué SvelteKit sobre Next.js 15
- App interna: RSC de Next.js no aporta nada (sin SEO, sin prerenderizado)
- SSE streaming nativo en `load()` functions — ideal para respuestas RAG
- Svelte 5 compila a código mínimo (~0 KB runtime)
- Sin vendor lock-in (Cloudflare, Vercel, etc. — adapter-node es portable)

---

## 4. Estructura de la aplicación

### Rutas SvelteKit

```
src/routes/
  (auth)/
    login/                    ← página de login (sin sidebar)
      +page.svelte
      +page.server.ts         ← POST /auth/session → setea cookie
  (app)/
    +layout.svelte            ← sidebar + auth guard
    +layout.server.ts         ← valida JWT en cookie, redirige si inválido
    chat/
      +page.svelte            ← nueva conversación
      [id]/
        +page.svelte          ← conversación existente (split view)
        +page.server.ts       ← carga historial
    collections/
      +page.svelte            ← lista de colecciones con stats
      [name]/
        +page.svelte          ← detalle + botón ingestar (si tiene permisos)
    admin/
      +layout.svelte          ← guard: ADMIN o AREA_MANAGER
      users/
        +page.svelte          ← tabla de usuarios, CRUD, reset-key
      areas/
        [id]/
          +page.svelte        ← detalle del área, permisos de colecciones
    audit/
      +page.svelte            ← log con filtros (solo ADMIN)
    settings/
      +page.svelte            ← perfil, API key personal, preferencias
  api/
    auth/
      session/+server.ts      ← POST/DELETE → proxy al gateway
      me/+server.ts           ← GET /auth/me
    chat/
      sessions/+server.ts     ← GET (lista), POST (crear sesión)
      sessions/
        [id]/+server.ts       ← GET (con mensajes), DELETE
      stream/
        [id]/+server.ts       ← SSE streaming de /v1/generate + guarda en sesión
    admin/
      users/+server.ts
      areas/+server.ts
```

### Sidebar por rol

El sidebar es icon-only con tooltips. Muestra ítems según el rol del JWT.

| Módulo | USER | AREA_MANAGER | ADMIN |
|--------|------|-------------|-------|
| 💬 Chat | ✓ | ✓ | ✓ |
| 📚 Colecciones | ✓ | ✓ | ✓ |
| ⬆️ Ingestión | — | ✓ | ✓ |
| 👥 Admin área | — | ✓ (solo su área) | ✓ (global) |
| 📋 Auditoría | — | — | ✓ |
| ⚙️ Settings | ✓ | ✓ | ✓ |

---

## 5. Módulo Chat (detalle)

El módulo de chat es la pantalla principal. Layout de 3 paneles:

```
┌─────────────┬────────────────────────────┬──────────────────┐
│  Historial  │       Conversación          │    Fuentes (N)   │
│  (150px)    │      (flex: 1.4)            │    (200px)       │
│             │                             │                  │
│ • Conv #1   │  user: ¿Cuál es el torque?  │ 📄 Manual p.47   │
│ • Conv #2   │                             │ ├ "730 Nm max..." │
│ • Conv #3   │  RAG: El torque máximo es   │                  │
│             │  **730 Nm** [Manual p.47]   │ 📄 Gieck tab.4.2 │
│             │                             │ ├ "Tornillos M10" │
│             │  [─────────────────] [▶]    │                  │
└─────────────┴────────────────────────────┴──────────────────┘
```

### Features del módulo Chat
- **Selector de colección** — dropdown en el header de la conversación
- **Toggle crossdoc ON/OFF** — activa el pipeline de decomposición multi-query
- **Progreso de decomposición visible** — contador de sub-queries en tiempo real
- **SSE streaming** — la respuesta aparece token a token, no de golpe
- **Citas inline** — `[Fuente p.XX]` son links que scrollean al panel de fuentes
- **Copiar respuesta** — botón de copy en cada mensaje del RAG
- **Compartir conversación** — link permanente a `/chat/[id]`
- **Historial persistente** — guardado en SQLite vía `GET/POST /chat/sessions`. El BFF crea la sesión al primer mensaje y persiste cada par pregunta/respuesta+fuentes. Default crossdoc: OFF.

---

## 6. Backend extensions (gateway FastAPI)

Los endpoints RAG existentes (`/v1/generate`, `/v1/search`, `/v1/collections`, `/v1/ingest`, `/admin/audit`) no se modifican. Todos los endpoints (existentes y nuevos) siguen autenticando con Bearer API key — el BFF provee siempre el `SYSTEM_API_KEY`. El RBAC por rol se aplica en el BFF antes del forwarding.

Se agregan/extienden:

### Auth (nuevos)
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | `/auth/session` | Login email+password → JWT httpOnly cookie |
| DELETE | `/auth/session` | Logout → borra cookie |
| GET | `/auth/me` | Perfil del usuario actual |
| POST | `/auth/refresh-key` | Regenerar API key propia |

JWT payload: `user_id · email · role · area_id · exp`. Firmado con `SECRET_KEY`. Expira en 8h.

### Admin — Usuarios (nuevos)
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/admin/users` | Listar todos los usuarios |
| POST | `/admin/users` | Crear usuario + generar API key |
| PUT | `/admin/users/{id}` | Editar rol / área / estado |
| DELETE | `/admin/users/{id}` | Desactivar (soft delete) |
| POST | `/admin/users/{id}/reset-key` | Generar nuevo API key para ese usuario |

### Admin — Áreas y permisos (nuevos)
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/admin/areas` | Listar áreas |
| POST | `/admin/areas` | Crear área |
| PUT | `/admin/areas/{id}` | Editar nombre / descripción |
| DELETE | `/admin/areas/{id}` | Eliminar si está vacía |
| GET | `/admin/areas/{id}/collections` | Ver permisos del área |
| POST | `/admin/areas/{id}/collections` | Otorgar acceso a colección |
| DELETE | `/admin/areas/{id}/collections/{name}` | Revocar acceso |

### Colecciones (extendido)
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/v1/collections` | Ya existe |
| GET | `/v1/collections/{name}/stats` | Entidades, docs, última ingesta (nuevo) |

### Chat — Historial de conversaciones (nuevos)

El historial persiste en SQLite (tabla `chat_sessions` + `chat_messages`). El BFF agrega cada intercambio pregunta/respuesta al crear la sesión.

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/chat/sessions` | Listar sesiones del usuario actual (paginado) |
| POST | `/chat/sessions` | Crear nueva sesión (devuelve `session_id`) |
| GET | `/chat/sessions/{id}` | Sesión con todos sus mensajes (pregunta + respuesta + fuentes) |
| DELETE | `/chat/sessions/{id}` | Borrar sesión propia |

El BFF crea la sesión al primer mensaje, luego guarda cada par pregunta/respuesta con fuentes. El `session_id` mapea al `[id]` en la ruta `/chat/[id]`.

### Auditoría (extendido)
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/admin/audit` | Ya existe, se extiende con query params |

Nuevos query params: `user_id · action · from · to · collection · limit`

### RBAC por endpoint

| Endpoint | USER | AREA_MANAGER | ADMIN |
|----------|------|-------------|-------|
| `/auth/session` (POST/DELETE) | ✓ | ✓ | ✓ |
| `/auth/me` | ✓ | ✓ | ✓ |
| `/auth/refresh-key` | ✓ (propia) | ✓ (propia) | ✓ (propia) |
| `/chat/sessions` (GET/POST/DELETE) | ✓ (propias) | ✓ (propias) | ✓ (propias) |
| `/v1/generate`, `/v1/search` | ✓ (colecciones del área) | ✓ (colecciones del área) | ✓ (todas) |
| `/v1/collections` | ✓ (filtradas por área) | ✓ (filtradas por área) | ✓ (todas) |
| `/v1/collections/{name}/stats` | ✓ (si tiene acceso) | ✓ (si tiene acceso) | ✓ |
| `/v1/ingest` | ✗ | ✓ (colecciones del área) | ✓ (todas) |
| `/admin/areas` | ✗ | ✓ (solo su área) | ✓ (todas) |
| `/admin/areas/{id}/collections` | ✗ | ✓ (solo su área) | ✓ |
| `/admin/users` | ✗ | ✗ | ✓ |
| `/admin/users/{id}/reset-key` | ✗ | ✗ | ✓ |
| `/admin/audit` | ✗ | ✗ | ✓ |

**AREA_MANAGER y su área propia:** El JWT contiene `area_id`. El BFF extrae ese valor y lo usa como parámetro en las llamadas. No hay endpoint separado para "mi área" — SvelteKit simplemente usa `/admin/areas/{area_id_from_jwt}/collections`. El gateway verifica que el `area_id` del Bearer token coincida o que el rol sea ADMIN.

**Total: 20 endpoints nuevos + 2 extendidos. El gateway crece de ~6 a ~28 endpoints.**

---

## 7. Despliegue

### Dockerfile (services/sda-frontend/Dockerfile)

```dockerfile
# Stage 1: build
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Stage 2: runtime (~100 MB)
FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/build .
COPY --from=builder /app/node_modules ./node_modules
ENV NODE_ENV=production
EXPOSE 3000
CMD ["node", "index.js"]
```

### Compose (agrega a config/compose-platform-services.yaml)

```yaml
sda-frontend:
  build:
    context: ${SALDIVIA_ROOT}
    dockerfile: services/sda-frontend/Dockerfile
  environment:
    - GATEWAY_URL=http://auth-gateway:8090
    - JWT_SECRET=${JWT_SECRET}
    - SYSTEM_API_KEY=${SYSTEM_API_KEY}
    - ORIGIN=https://sda.tecpia.local
    - PUBLIC_APP_NAME=SDA
  ports:
    - "3000:3000"
  depends_on:
    - auth-gateway
  networks:
    - default   # nvidia-rag
  restart: unless-stopped
```

### Variables de entorno (agregar a .env.saldivia)
| Variable | Descripción |
|----------|-------------|
| `GATEWAY_URL` | URL interna del gateway (default: `http://auth-gateway:8090`) |
| `JWT_SECRET` | Clave compartida con el gateway para firmar/verificar JWT |
| `SYSTEM_API_KEY` | API key del sistema para llamadas BFF→gateway (mismo que el admin API key existente) |
| `ORIGIN` | URL pública de SDA para protección CSRF de SvelteKit |
| `PUBLIC_APP_NAME` | Nombre mostrado en la UI (default: `SDA`) |

### Red Docker
- Puerto `3000` es el único expuesto al host
- Gateway en `:8090` es solo accesible dentro de la red `nvidia-rag`
- `sda-frontend` resuelve `auth-gateway` por nombre dentro de la red Docker

### Perfiles de deploy
- **brev-2gpu:** Frontend + Gateway + RAG + Ingestor + NIMs — todo local en Brev
- **workstation-1gpu:** Frontend + Gateway local, RAG/Ingestor/LLM por API externa

`make deploy PROFILE=brev-2gpu` incluye el frontend automáticamente.

---

## 8. Diseño visual

### Paleta de colores
- **Background principal:** `#0f172a` (dark navy)
- **Background secundario:** `#0c1220` (más oscuro, panels laterales)
- **Bordes:** `#1e293b` (sutil)
- **Acento primario:** `#6366f1` (indigo)
- **Acento secundario:** `#4338ca` (indigo oscuro, burbujas de usuario)
- **Texto principal:** `#e2e8f0`
- **Texto secundario:** `#94a3b8`
- **Texto muted:** `#475569`

### Identidad
- Logo/badge **SDA** en sidebar — fondo indigo, texto blanco, border-radius 4px
- Tipografía: Inter (sans-serif system font)
- Sidebar: icon-only con tooltips, 40px de ancho, items de 28px × 24px
- Dark mode nativo (mode-watcher, sin flash en SSR)

---

## 9. Estructura de archivos nueva en rag-saldivia/

```
rag-saldivia/
  services/
    auth-gateway/          ← existente (FastAPI)
    mode-manager/          ← existente
    sda-frontend/          ← NUEVO
      src/
        routes/            ← ver sección 4
        lib/
          components/      ← UI components (shadcn-svelte)
          stores/          ← Svelte stores (session, chat)
          api/             ← wrappers para las routes BFF
      static/
      svelte.config.js
      tailwind.config.js
      package.json
      Dockerfile
  config/
    compose-platform-services.yaml  ← agregar sda-frontend
    profiles/
      brev-2gpu.yaml                ← incluir sda-frontend
      workstation-1gpu.yaml         ← incluir sda-frontend
```

---

## 10. Consideraciones de seguridad

- **API key del sistema** nunca llega al browser — solo vive en las variables de entorno del container SvelteKit
- **JWT en cookie httpOnly** — no accesible desde JavaScript del browser
- **CSRF protection** — SvelteKit verifica el header `Origin` automáticamente con el valor de `ORIGIN`
- **RBAC en gateway** — cada endpoint verifica el rol del JWT además de la autenticación
- **Audit log** — todas las consultas RAG se registran con `user_id`, timestamp, colección y acción

---

## 11. Fuera de scope (v1)

- Internacionalización (i18n)
- App móvil o PWA
- Notificaciones push
- Modo oscuro/claro togglable por usuario (siempre dark)
- Integración ERP activa (solo diseñada como extension point)
- Rate limiting por usuario (se puede agregar al gateway después)
- Chat multi-colección en una sola consulta

# 04 — Seguridad y Autenticacion

## 3 capas de seguridad

### Capa 1: Edge Middleware (`proxy.ts`)

**Archivo:** `apps/web/src/proxy.ts` (139 lineas)

Corre en **Edge Runtime** de Next.js — no tiene acceso a Node.js APIs (ioredis, fs, etc.).

**Que hace:**
1. Genera `x-request-id` (UUID) para correlacion de logs en CADA request
2. Deja pasar rutas publicas sin verificacion
3. Extrae JWT de cookie `auth_token` o header `Authorization: Bearer`
4. Reconoce `SYSTEM_API_KEY` para acceso servicio-a-servicio (rol admin)
5. Verifica firma JWT con `jwtVerify()` de jose (HS256)
6. Ejecuta RBAC: `canAccessRoute(claims, pathname)`
7. Propaga claims como headers para downstream:
   - `x-user-id`, `x-user-email`, `x-user-name`, `x-user-role`, `x-user-jti`

**Rutas publicas (sin auth):**
```
/login
/api/auth/login
/api/auth/refresh
/api/health
/api/log
/_next/*
/favicon.ico
```

**Comportamiento por tipo de ruta:**
- API routes sin token → 401 JSON `{ ok: false, error: "No autenticado" }`
- Pages sin token → redirect a `/login?from={pathname}`
- Token invalido en API → 401 JSON
- Token invalido en pages → redirect a login
- Sin permisos RBAC en API → 403 JSON `{ ok: false, error: "Acceso denegado" }`
- Sin permisos RBAC en pages → redirect a `/`

---

### Capa 2: Route Handlers (`extractClaims`)

**Archivo:** `apps/web/src/lib/auth/jwt.ts` (139 lineas)

Corre en **Node.js Runtime** — tiene acceso a ioredis.

**`extractClaims(request)`:**
1. Lee headers `x-user-*` propagados por proxy.ts
2. Verifica revocacion en Redis: `GET revoked:{jti}` — si existe, retorna null
3. Si no hay headers → fallback: extrae JWT de cookie/Authorization y verifica
4. Siempre verifica blacklist antes de retornar claims

**`createJwt(claims)`:**
- Algoritmo: HS256
- Expiracion: configurable via `JWT_EXPIRY` env (default 24h)
- JTI: `crypto.randomUUID()` — requerido para revocacion
- Secret: `JWT_SECRET` env (requerido, no hay fallback)

**`makeAuthCookie(token)`:**
- `HttpOnly` — no accesible desde JavaScript
- `SameSite=Lax` — proteccion CSRF basica
- `Secure` — solo en produccion (`NODE_ENV=production`)
- `Path=/` — disponible en todas las rutas

---

### Capa 3: Server Actions (`safe-action.ts`)

**Archivo:** `apps/web/src/lib/safe-action.ts` (34 lineas)

Middleware de next-safe-action:
- **`authAction`:** inyecta `user` autenticado en `ctx`. Si no hay usuario → error.
- **`adminAction`:** inyecta `user` admin en `ctx`. Si no es admin → error.
- Input validado con Zod schemas en cada action via `.schema(z.object({...}))`
- Retorno wrapped: callers acceden a `result?.data`

**Helper `clean()`:** bridge entre Zod optional (`T | undefined`) y `exactOptionalPropertyTypes` (`T?`).

---

## RBAC — Control de acceso basado en roles

### Sistema legacy (campo role en users)

**Archivo:** `apps/web/src/lib/auth/rbac.ts`

Jerarquia simple: `admin (3) > area_manager (2) > user (1)`

Rutas protegidas por patron:
| Patron | Rol minimo |
|--------|-----------|
| `/admin/*` | admin |
| `/api/admin/*` | admin |
| `/audit/*` | area_manager |
| `/api/audit/*` | area_manager |
| Todo lo demas | solo autenticacion |

### Sistema granular (Plan 21)

Tablas: `roles`, `permissions`, `role_permissions`, `user_role_assignments`

- Permisos por clave: `"users.manage"`, `"collections.admin"`, etc.
- Roles con nivel numerico (mayor = mas poderoso)
- Roles de sistema (`isSystem=true`) — no se pueden eliminar
- Cada rol tiene color e icono para badges en la UI

### Permisos por coleccion

Tabla `area_collections`: `read | write | admin` por area.
Los usuarios heredan permisos de sus areas (`user_areas`).

---

## Flujo de login completo

```
POST /api/auth/login
  Body: { email: string, password: string }
  |
  +-- Buscar usuario por email en DB
  |   Si no existe → 401
  |
  +-- Verificar password con bcrypt
  |   Si no coincide → 401
  |
  +-- Verificar usuario activo
  |   Si active=false → 403
  |
  +-- createJwt({ sub: id, email, name, role })
  |   Genera JWT con jti=UUID, exp=24h
  |
  +-- Response:
      Set-Cookie: auth_token=<JWT>; HttpOnly; SameSite=Lax; Max-Age=86400
      Body: { ok: true, user: { id, email, name, role } }
```

## Flujo de logout completo

```
DELETE /api/auth/logout
  Cookie: auth_token=<JWT>
  |
  +-- extractClaims(request) → obtener jti
  |
  +-- Redis: SET revoked:{jti} 1 EX {segundos_restantes_del_token}
  |   El TTL en Redis = expiracion del JWT - ahora
  |   Asi la key se auto-limpia cuando el JWT expiraria naturalmente
  |
  +-- Response:
      Set-Cookie: auth_token=; Max-Age=0; Path=/; HttpOnly; SameSite=Lax
      Body: { ok: true }
```

## Flujo de refresh

```
POST /api/auth/refresh
  Cookie: auth_token=<JWT_viejo>
  |
  +-- extractClaims(request) → claims del token actual
  |   Si invalido o revocado → 401
  |
  +-- Revocar token viejo en Redis
  |
  +-- createJwt(claims_actualizados)
  |   Nuevo jti, nueva expiracion
  |
  +-- Response:
      Set-Cookie: auth_token=<JWT_nuevo>; HttpOnly; ...
```

---

## SYSTEM_API_KEY

- Acceso servicio-a-servicio con rol admin imploto
- Se configura como env var
- proxy.ts detecta si el bearer token es la API key
- Headers seteados: `x-user-id=0`, `x-user-email=system@internal`, `x-user-role=admin`
- Uso: CLI, workers de background, health checks automatizados

---

## Guard rules (operaciones prohibidas sin OK de Enzo)

| Categoria | Bloqueado |
|-----------|-----------|
| Destruccion masiva | `rm -rf /`, wildcard removes |
| Historia Git | `--force` push, `reset --hard`, borrar branches remotas |
| Base de datos | `DROP TABLE`, `DROP DATABASE`, `TRUNCATE` |
| Produccion | Deploy sin verificacion |
| Seguridad | `chmod 777`, `--no-verify` |
| Codigo remoto | `curl ... | sh` |

---

## Variables de entorno de seguridad

| Variable | Proposito | Generacion |
|----------|-----------|-----------|
| `JWT_SECRET` | Firma de JWTs (HS256) | `openssl rand -base64 32` |
| `SYSTEM_API_KEY` | Acceso S2S | `openssl rand -hex 32` |
| `JWT_EXPIRY` | Duracion del token | Default "24h" |
| `REDIS_URL` | Conexion a Redis | `redis://localhost:6379` |

**Nota de seguridad:** `external_sources.credentials` almacena credenciales como JSON en texto plano. En produccion deberia cifrarse con `SYSTEM_API_KEY`.

---

## Superficie de ataque — analisis profundo

### SQL Injection — BAJO RIESGO

Drizzle ORM parametriza queries automaticamente. El unico punto de atencion es `packages/db/src/queries/search.ts` (lineas 66, 90, 109) que usa concatenacion para LIKE:

```typescript
sql`... LIKE ${"%" + q + "%"} ...`
```

**Esto es SEGURO** — el template `sql` de Drizzle bindea valores como parametros. Pero FTS5 MATCH con input no sanitizado podria tener edge cases con caracteres especiales (`*`, `"`, `OR`, `NOT`).

### XSS — PROTEGIDO

- `ArtifactPanel.tsx`: usa `DOMPurify.sanitize()` para SVG/HTML del LLM
- `MarkdownMessage.tsx`: renderer React que escapa HTML (no usa dangerouslySetInnerHTML)
- Shiki (syntax highlighting): HTML generado marcado como trusted

### CSRF — PROTEGIDO

- Server Actions: cookies HttpOnly + SameSite=Lax (proteccion implicita)
- API routes: JWT en cookie HttpOnly o Authorization header

### Headers de seguridad FALTANTES (agregar en next.config para deploy)

| Header | Proposito | Estado |
|--------|-----------|--------|
| `X-Content-Type-Options: nosniff` | Previene MIME sniffing | No configurado |
| `X-Frame-Options: DENY` | Previene clickjacking | No configurado |
| `Strict-Transport-Security` | Fuerza HTTPS | No configurado |
| `Content-Security-Policy` | Previene XSS/injection | No configurado |

### Resumen de hallazgos

| Hallazgo | Severidad |
|----------|-----------|
| Input validation completa (Zod everywhere) | OK |
| SQL injection protegido (Drizzle parametrizado) | OK |
| XSS protegido (DOMPurify) | OK |
| CSRF protegido (SameSite + HttpOnly) | OK |
| FTS5 MATCH sin sanitizar chars especiales | Baja |
| `external_sources.credentials` en texto plano | Media |
| Headers de seguridad faltantes (CSP, HSTS) | Media |
| Revocacion JWT no verificada en Edge | Media (documentado, mitigado) |

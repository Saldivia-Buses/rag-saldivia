---
name: rag-auth
description: Implement and debug authentication, JWT, RBAC, and permissions in RAG Saldivia. Use when adding auth to a new route, checking permissions, working with roles, or when the user asks about "autenticación", "permisos", "RBAC", "roles", "middleware", "proteger una ruta", "quién puede acceder a X", or "JWT".
---

# RAG Saldivia — Auth & RBAC

## Archivos críticos — leer antes de modificar

- `apps/web/src/middleware.ts` — controla auth y RBAC en cada request
- `apps/web/src/lib/auth/jwt.ts` — `createJwt`, `verifyJwt`, lectura de claims desde headers
- `apps/web/src/lib/auth/rbac.ts` — `hasPermission`, matriz de permisos por rol

## Roles

| Rol | Permisos |
|-----|----------|
| `admin` | Todo — usuarios, áreas, colecciones, config |
| `area_manager` | CRUD en sus áreas asignadas y sus colecciones |
| `user` | Solo lectura — chat, colecciones de sus áreas |

## Flujo de autenticación (resumen)

```
Request → middleware.ts
  ↓ Lee cookie "auth_token" (usuarios web)
  ↓ O header "X-API-Key" (CLI / service-to-service)
  ↓ Verifica JWT con JWT_SECRET
  ↓ Inyecta headers x-user-id, x-user-role, x-user-areas en la request
  ↓ Server Components / API routes leen via getCurrentUser() o extractClaims()
  ↓ Si inválido → 401 / redirect /login
```

## Service-to-service (CLI → API)

La CLI usa `Authorization: Bearer <SYSTEM_API_KEY>` (no JWT).  
El middleware detecta esto e inyecta headers con `role: admin`.  
`extractClaims()` lee los `x-user-*` headers directamente — nunca intenta verificar el JWT en este flujo.

## Permisos de colección (multi-tenant)

Un usuario puede acceder a una colección si:
1. Su rol es `admin`, **o**
2. Pertenece a un área con permiso `read|write|admin` sobre esa colección

La verificación se hace en DB — ver `packages/db/src/queries/users.ts`.

## Proteger una ruta nueva

**Server Component / Page:**  
Usar `requireAuth({ role: "admin" })` de `@/lib/auth/current-user`.  
Lanza redirect a `/login` automáticamente si no está autenticado.

**API Route:**  
Llamar `getAuthFromRequest(req)` de `@/lib/auth/jwt`.  
Si retorna `null` → 401. Verificar rol manualmente → 403.

**Server Action:**  
Llamar `getCurrentUser()` (usa `React.cache()` — una sola query por request).  
Validar con `hasPermission(user.role, "permiso:requerido")`.

## Debugging de auth

```bash
rag audit log --type auth.login
rag audit log --type auth.failed
```

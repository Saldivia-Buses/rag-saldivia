# Arquitectura — RAG Saldivia

> Branch: `1.0.x`
> Última actualización: 2026-03-29 (Plan 13)

---

## Diagrama

```
Usuario --> Next.js :3000 ────────────────────--> RAG Server :8081
             (UI + auth + proxy)                        |
                    |                            Milvus + NIMs
               Redis :6379                              |
         (JWT · cache · BullMQ)             Nemotron-Super-49B
```

Un único proceso Next.js reemplaza el gateway Python + frontend SvelteKit
del stack original (branch `main`).

---

## Flujo de una request

```
1. Browser --> middleware (proxy.ts)
   - Verifica JWT en edge
   - Evalúa RBAC (admin, area_manager, user)
   - Agrega x-request-id, x-user-jti

2. Si es página --> Server Component
   - Data fetching server-side
   - Renderiza HTML

3. Si es API --> Route Handler
   - /api/auth/* --> JWT create/verify/refresh
   - /api/rag/* --> Proxy al RAG Server :8081
   - /api/health --> Health check

4. Si es mutación --> Server Action
   - actions/chat.ts --> CRUD sesiones/mensajes
   - actions/settings.ts --> Perfil/password
```

---

## Paquetes

```
apps/web/          --> Next.js (consume packages/)
  src/app/         --> 5 páginas + 9 API routes + 5 server actions
  src/components/  --> 18 UI primitives + 3 chat + 3 layout + 2 collections/settings
  src/hooks/       --> useRagStream (SSE streaming)
  src/lib/         --> auth (jwt, rbac), rag (client, stream, cache), utils

packages/db/       --> Drizzle ORM + libsql + Redis
  src/schema.ts    --> Schema SQLite completo
  src/queries/     --> 17 archivos de queries por dominio
  src/redis.ts     --> Singleton Redis + BullMQ connection

packages/shared/   --> Zod schemas + tipos compartidos
packages/config/   --> Config loader YAML
packages/logger/   --> Logger estructurado + rotación
```

---

## Auth

```
Login --> /api/auth/login --> verifyPassword (bcrypt)
                          --> createJwt (jose, HS256)
                          --> Set-Cookie: HttpOnly, Secure, SameSite=Strict

Request --> proxy.ts (edge) --> verifyJwt
                            --> RBAC check (role vs route pattern)
                            --> x-user-jti header

Revocación --> extractClaims() --> Redis: check jti blacklist
              (Node.js runtime, NOT edge — no ioredis en edge)
```

---

## SSE Streaming

```
Browser --> /api/rag/generate (POST)
            --> Verify auth
            --> fetch RAG :8081/generate (SSE)
            --> CHECK HTTP STATUS FIRST (critical pattern)
            --> ReadableStream pipe to client
            --> useRagStream hook reads tokens
```

**Patrón crítico:** verificar status HTTP del RAG **antes** de streamear.
El gateway Python original siempre retornaba 200 aunque el RAG fallara.


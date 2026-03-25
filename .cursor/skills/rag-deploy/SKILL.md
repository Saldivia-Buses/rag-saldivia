---
name: rag-deploy
description: Set up, start, and manage the RAG Saldivia development stack. Use when setting up the project for the first time, starting the dev server, checking service health, asking about environment variables, or when the user says "arrancar el servidor", "cómo seteo el proyecto", "rag status", "health check", or "variables de entorno".
---

# RAG Saldivia — Setup & Dev

Reference: `docs/onboarding.md` para setup completo.

## Setup inicial

```bash
bun run setup   # onboarding completo (instala deps, crea DB, seed)
bun run dev     # Next.js en http://localhost:3000
```

Credenciales de desarrollo: `admin@localhost` / `changeme`

## WSL2 — precaución crítica

En WSL2 el repo **debe estar en el filesystem de Linux** (`~/rag-saldivia`), no en `/mnt/c/`.  
`bun install` en `/mnt/` no crea symlinks correctamente.  
Además: `DATABASE_PATH` debe ser **ruta absoluta** (ej: `/home/enzo/rag-saldivia/data/app.db`).

## Variables de entorno críticas

El `.env.local` vive en `apps/web/.env.local`, no en la raíz.  
Ver `.env.example` para la lista completa documentada.

| Variable | Descripción | Cómo generar |
|----------|-------------|--------------|
| `JWT_SECRET` | Firma JWT — requerida | `openssl rand -base64 32` |
| `SYSTEM_API_KEY` | Auth CLI/service-to-service — requerida | `openssl rand -hex 32` |
| `RAG_SERVER_URL` | `http://localhost:8081` por defecto | — |
| `MOCK_RAG` | `true` para dev sin Docker | — |
| `DATABASE_PATH` | Ruta al archivo SQLite | Absoluta en WSL2 |

## Sin Docker

```bash
# Agregar MOCK_RAG=true al .env.local
bun run dev   # simula respuestas RAG con streaming
```

## Health check

```bash
rag status                             # semáforo de todos los servicios con latencias
curl http://localhost:3000/api/health  # endpoint directo
```

## Base de datos

```bash
rag db migrate   # aplicar migraciones pendientes
rag db seed      # crear datos de prueba
rag db reset     # BORRAR DB + rehacer seed (con confirmación)
```

## Build de producción

```bash
bun run build    # build completo via Turborepo
```

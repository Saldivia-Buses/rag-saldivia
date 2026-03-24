# RAG Saldivia

Overlay sobre el **NVIDIA RAG Blueprint v2.5.0** con autenticación JWT, RBAC, multi-colección, frontend Next.js 15 y CLI TypeScript.

> **Branch activa:** `experimental/ultra-optimize` — reescritura completa en TypeScript.
> La branch `main` contiene el stack Python + SvelteKit original.

## Arquitectura

```
Usuario → Next.js :3000 ——————————————————→ RAG Server :8081
           (UI + auth + proxy)                      ↓
                                           Milvus + NIMs
                                                    ↓
                                       Nemotron-Super-49B
```

Un único proceso reemplaza el gateway Python (9000) y el frontend SvelteKit (3000) del stack original.

## Quick Start

```bash
git clone https://github.com/Camionerou/rag-saldivia
cd rag-saldivia
git checkout experimental/ultra-optimize
bun run setup
bun run dev
```

Abrí http://localhost:3000 — credenciales de desarrollo: `admin@localhost` / `changeme`

Sin Docker (solo UI): agregá `MOCK_RAG=true` en `.env.local`

## Stack

| Componente | Tecnología |
|---|---|
| Servidor | Next.js 15 App Router (TypeScript 6.0) |
| Base de datos | SQLite — Drizzle ORM + better-sqlite3 |
| Auth | JWT (jose) en cookie HttpOnly |
| Validación | Zod compartido entre frontend y backend |
| Build | Turborepo + Bun workspaces |
| CLI | Commander + @clack/prompts + chalk |

## CLI

```bash
# Instalar globalmente
cd apps/cli && bun link

rag status                    # estado de todos los servicios
rag users list                # gestión de usuarios
rag collections list          # colecciones disponibles
rag ingest start              # subir documentos
rag audit log                 # eventos del sistema
rag audit replay 2026-03-24   # reconstruir estado desde fecha
```

## Estructura

```
apps/
  web/          → servidor único (Next.js 15): UI + auth + proxy RAG + admin
  cli/          → CLI TypeScript
packages/
  shared/       → Zod schemas + tipos compartidos
  db/           → Drizzle ORM (14 tablas, reemplaza SQLite auth + Redis)
  config/       → config loader (reemplaza config.py)
  logger/       → logging estructurado + black box replay
config/         → YAMLs de configuración (sin cambios)
patches/        → patches del blueprint NVIDIA (sin cambios)
legacy/         → stack original Python + SvelteKit (referencia)
docs/           → arquitectura, CLI, blackbox, onboarding
scripts/        → setup.ts, health-check.ts
```

## Documentación

| Doc | Descripción |
|---|---|
| [Architecture](docs/architecture.md) | Arquitectura del nuevo stack, flujos de auth y RAG |
| [Onboarding](docs/onboarding.md) | Guía de 5 minutos para arrancar |
| [CLI](docs/cli.md) | Referencia completa de comandos |
| [Black Box](docs/blackbox.md) | Sistema de logging y replay de eventos |
| [Plan](docs/plans/ultra-optimize.md) | Seguimiento diario del plan de trabajo |

## License

MIT

# RAG Saldivia

[![CI](https://github.com/Camionerou/rag-saldivia/actions/workflows/ci.yml/badge.svg?branch=experimental/ultra-optimize)](https://github.com/Camionerou/rag-saldivia/actions/workflows/ci.yml) ![Version](https://img.shields.io/badge/version-1.0.0-blue) ![License](https://img.shields.io/badge/license-MIT-green) ![Bun](https://img.shields.io/badge/bun-1.3%2B-orange)

> Overlay sobre NVIDIA RAG Blueprint v2.5.0 — autenticación JWT, RBAC,
> multi-colección, frontend Next.js 16, en un único proceso.

## Para agentes de IA

Leer `docs/bible.md` primero. Contiene las reglas permanentes del proyecto
(workflow, stack, protocolos, naming). Después `docs/plans/1.0.x-plan-maestro.md`
para el roadmap actual.

## Quick Start

```bash
git clone https://github.com/Camionerou/rag-saldivia
cd rag-saldivia && git checkout 1.0.x
cp .env.example .env.local
# Editar .env.local: agregar JWT_SECRET y verificar REDIS_URL

bun run setup
MOCK_RAG=true bun run dev
```

Abrir http://localhost:3000 — login: `admin@localhost` / `changeme`

## Requisitos

- **Bun** >= 1.3 — `curl -fsSL https://bun.sh/install | bash`
- **Redis** 7+ — `docker run -d -p 6379:6379 redis:alpine`
- **RAG Server** (opcional en dev — usar `MOCK_RAG=true`)

## Arquitectura

```
Usuario --> Next.js :3000 (UI + auth + proxy) --> RAG Server :8081
                     |                                  |
                Redis :6379                       Milvus + NIMs
            (JWT · cache · BullMQ)               Nemotron-Super-49B
```

## Estructura del repo

| Path | Descripción |
|------|-------------|
| `apps/web/` | Next.js 16 — UI + auth + proxy RAG |
| `packages/db/` | Drizzle ORM + libsql — schema SQLite + Redis |
| `packages/logger/` | Logger estructurado |
| `packages/shared/` | Schemas Zod compartidos |
| `packages/config/` | Config loader |
| `docs/` | Bible, plans, ADRs, templates, toolbox |
| `_archive/` | Código aspiracional (admin, CLI, upload, etc.) |

## Stack técnico

| Componente | Tecnología |
|---|---|
| Runtime | Bun 1.3+ |
| Framework | Next.js 16 App Router |
| Base de datos | SQLite (Drizzle ORM + libsql) |
| Auth | JWT (jose) + Redis blacklist |
| Queue | BullMQ + Redis |
| CSS | Tailwind v4 + shadcn/ui + Radix |
| Testing | bun:test + happy-dom + Playwright |

## Páginas activas

| Ruta | Descripción |
|------|-------------|
| `/login` | Autenticación |
| `/chat` | Lista de sesiones |
| `/chat/[id]` | Conversación con RAG |
| `/collections` | Colecciones disponibles |
| `/settings` | Perfil y preferencias |

## Comandos principales

```bash
bun run dev              # Dev server :3000
bun run test             # Unit tests (~92)
bun run test:components  # Component tests (~99) — desde apps/web/
bun run test:visual      # Visual regression
bun run test:a11y        # WCAG AA
bun run test:e2e         # E2E Playwright
bun run storybook        # Storybook :6006
bun run lint             # Lint
```

## Variables de entorno

| Variable | Rol |
|----------|-----|
| `JWT_SECRET` | Firma JWT (obligatorio) |
| `REDIS_URL` | Conexión Redis (obligatorio) |
| `DATABASE_PATH` | Ruta al SQLite |
| `RAG_SERVER_URL` | URL del RAG Server |
| `MOCK_RAG` | `true` para dev sin GPU |

Ver `.env.example` para la lista completa.

## Documentación

| Documento | Contenido |
|-----------|-----------|
| `docs/bible.md` | Reglas permanentes del proyecto |
| `docs/plans/1.0.x-plan-maestro.md` | Roadmap y planes |
| `docs/decisions/` | ADRs (12 decisiones) |
| `docs/toolbox.md` | Herramientas externas |
| `CLAUDE.md` | Contexto para agentes de IA |
| `CONTRIBUTING.md` | Cómo contribuir |

## Contributing

Ver [CONTRIBUTING.md](CONTRIBUTING.md). TL;DR: leer `docs/bible.md` primero.

## Licencia

MIT — ver [LICENSE](LICENSE).

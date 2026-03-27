# RAG Saldivia

[![CI](https://github.com/Camionerou/rag-saldivia/actions/workflows/ci.yml/badge.svg?branch=experimental/ultra-optimize)](https://github.com/Camionerou/rag-saldivia/actions/workflows/ci.yml) ![Version](https://img.shields.io/badge/version-1.0.0-blue) ![License](https://img.shields.io/badge/license-MIT-green) ![Bun](https://img.shields.io/badge/bun-1.3%2B-orange)

> Overlay sobre NVIDIA RAG Blueprint v2.5.0 — autenticación JWT, RBAC, multi-colección, admin y CLI TypeScript, en un único proceso Next.js 16.

## Descripción

RAG Saldivia es un **overlay** sobre el [NVIDIA RAG Blueprint](https://github.com/NVIDIA-AI-Blueprints/rag) v2.5.0: **no es un fork** del blueprint; el motor RAG y los contenedores oficiales se usan como dependencia (submódulo en `vendor/rag-blueprint/`). El proyecto añade autenticación JWT con revocación en Redis, RBAC por roles y áreas, consola de administración completa, CLI TypeScript, design system “Warm Intelligence” y una sola aplicación Next.js que sirve UI, API y proxy hacia el RAG Server.

La arquitectura objetivo es **un solo proceso** en el servidor de aplicación: Next.js reemplaza el gateway Python y el frontend SvelteKit del stack original, reduciendo saltos de red y unificando el modelo de despliegue. **Redis** es obligatorio: blacklist de JWT, caché de listas de colecciones, pub/sub para notificaciones, contadores de secuencia para eventos y **BullMQ** para la cola de ingesta de documentos.

**Estado actual:** la autenticación, el panel de administración, la gestión de usuarios y áreas, el chat con sesiones, proyectos, subida de archivos con cola de ingesta y las notificaciones están implementados y probados en este monorepo. Las capacidades que dependen del **RAG Server en GPU**, Milvus y modelos NIM (respuestas en streaming contra el LLM de producción, ingesta masiva en vectores, etc.) requieren la workstation con hardware NVIDIA y se documentan como evolución en versiones futuras; en desarrollo local podés usar `MOCK_RAG=true` sin levantar el stack completo.

## Arquitectura

```
Usuario ──→ Next.js :3000 ──────────────────────────→ RAG Server :8081
             (UI + auth + proxy + admin)                      ↓
                      ↓                             Milvus + NIMs (GPU)
                 Redis :6379                                  ↓
           (JWT · cache · pub/sub · BullMQ)        Nemotron-Super-49B
```

El RAG Server y Milvus son componentes del NVIDIA RAG Blueprint. Requieren hardware GPU. La app puede ejecutarse con `MOCK_RAG=true` sin esos componentes para desarrollo de UI y flujos de aplicación.

## Requisitos

- **Bun** ≥ 1.3 — `curl -fsSL https://bun.sh/install | bash`
- **Redis** 7+ — `docker run -d -p 6379:6379 redis:alpine`
- **Node.js** ≥ 22 (opcional, para herramientas de CI o editores)
- **RAG Server** (opcional en desarrollo — ver `MOCK_RAG` en `.env.example`)

## Quick Start

```bash
# 1. Clonar y configurar
git clone https://github.com/Camionerou/rag-saldivia
cd rag-saldivia
git checkout experimental/ultra-optimize
cp .env.example .env.local
# Editar .env.local: agregar JWT_SECRET y verificar REDIS_URL

# 2. Instalar y configurar la DB
bun run setup

# 3. Iniciar el servidor de desarrollo
MOCK_RAG=true bun run dev
```

Abrí http://localhost:3000

Credenciales de desarrollo: `admin@localhost` / `changeme`

## Estructura del monorepo

| Path | Descripción |
|------|-------------|
| `apps/web/` | Next.js 16 — UI + auth + proxy RAG + admin |
| `apps/cli/` | CLI TypeScript — `rag users/collections/ingest/audit/config/db/status` |
| `packages/db/` | Drizzle ORM + libsql — esquema SQLite + Redis |
| `packages/logger/` | Logger estructurado + blackbox replay |
| `packages/shared/` | Schemas Zod compartidos entre web y CLI |
| `packages/config/` | Config loader con validación |

## Comandos principales

| Comando | Descripción |
|---------|-------------|
| `bun run dev` | Servidor de desarrollo en :3000 |
| `bun run test` | Tests de lógica en paquetes (~311 tests) |
| `bun run test:components` | Tests de componentes React (~153 tests) |
| `bun run test:visual` | Regresión visual Playwright (22 tests) |
| `bun run test:a11y` | Auditoría WCAG AA (páginas clave) |
| `bun run test:e2e` | E2E Playwright (flujos críticos) |
| `bun run storybook` | Catálogo de componentes en :6006 |
| `rag status` | Health check del sistema |

## Stack técnico

| Componente | Tecnología | Versión |
|------------|------------|---------|
| Runtime | Bun | 1.3+ |
| Framework | Next.js App Router | 16.x |
| Base de datos | SQLite (Drizzle ORM + libsql) | — |
| Cola de tareas | BullMQ + Redis | — |
| Auth | JWT (jose) + Redis blacklist | — |
| Validación | Zod | 4.x |
| Build | Turborepo + Bun workspaces | — |
| CSS | Tailwind v4 | — |
| Componentes | shadcn/ui + Radix | — |
| Testing | bun:test + Playwright | — |

## Features en v1.0.0

- Autenticación JWT con refresh y revocación inmediata (Redis)
- RBAC por roles (`admin` / `area_manager` / `user`) y áreas
- Páginas de aplicación con design system “Warm Intelligence”
- Admin: usuarios, áreas, permisos, configuración RAG, ingesta, analíticas
- CLI con comandos de estado, usuarios, colecciones, ingesta, auditoría y DB
- Subida de documentos con cola de ingesta (BullMQ)
- Notificaciones en tiempo real (Redis Pub/Sub + SSE)
- Sesiones de chat con historial, respuestas guardadas y etiquetas
- Proyectos con contexto persistente
- Dark mode y auditoría de accesibilidad (WCAG AA en flujos cubiertos)

## Features en versiones futuras

> Las siguientes features requieren la workstation con GPU y estarán disponibles en versiones futuras:

- Streaming de respuestas del LLM (Nemotron-Super-49B) en producción
- Ingestión real de documentos en Milvus a escala
- Consulta multi-colección (crossdoc) de extremo a extremo
- Grafo de similitud entre documentos con datos reales del vector store
- Vista dividida para comparación de respuestas
- SSO Google / Azure AD

## Contributing

Ver [CONTRIBUTING.md](CONTRIBUTING.md) para instrucciones de setup y convenciones de código.

## Documentación adicional

| Documento | Contenido |
|-----------|-----------|
| [docs/architecture.md](docs/architecture.md) | Arquitectura del stack |
| [docs/onboarding.md](docs/onboarding.md) | Onboarding rápido |
| [docs/api.md](docs/api.md) | Referencia de la API HTTP |
| [docs/cli.md](docs/cli.md) | Referencia de la CLI |
| [docs/testing.md](docs/testing.md) | Guía de testing |
| [docs/design-system.md](docs/design-system.md) | Tokens y componentes UI |
| [docs/blackbox.md](docs/blackbox.md) | Eventos, logs y replay |
| [docs/workflows.md](docs/workflows.md) | Git, PRs y convenciones |
| [CLAUDE.md](CLAUDE.md) | Contexto para asistentes (arquitectura y archivos críticos) |

## Variables de entorno esenciales

| Variable | Rol |
|----------|-----|
| `JWT_SECRET` | Firma de JWT (obligatorio en producción) |
| `REDIS_URL` | Conexión Redis (obligatorio — ADR-010) |
| `DATABASE_PATH` | Ruta al SQLite de aplicación |
| `RAG_SERVER_URL` | URL del RAG Server (default `http://localhost:8081`) |
| `MOCK_RAG` | `true` para no llamar al RAG real en desarrollo |
| `SYSTEM_API_KEY` | Autenticación service-to-service (CLI, integraciones) |

La lista completa está documentada en `.env.example`.

## CLI — instalación rápida

```bash
cd apps/cli && bun link
rag status
rag users list
rag collections list
rag ingest status
rag audit log
```

Ver [apps/cli/README.md](apps/cli/README.md) y [docs/cli.md](docs/cli.md) para la referencia completa.

## Calidad y CI

- **Lint:** `bun run lint` (monorepo vía Turborepo)
- **Type-check:** `bun run type-check`
- **Cobertura:** `bun run test:coverage` (paquetes que la exponen)
- **CI:** GitHub Actions en `.github/workflows/ci.yml` (tests, lint, type-check, componentes, visual, a11y, E2E según configuración)

Reportes de vulnerabilidad: ver [SECURITY.md](SECURITY.md).

## Licencia

MIT — ver [LICENSE](LICENSE).

---

## Apéndice — mapa de paquetes

### `packages/db`

Esquema Drizzle + `@libsql/client`, queries por dominio, Redis para secuencias, blacklist y cachés. Ver [packages/db/README.md](packages/db/README.md).

### `packages/logger`

Logger estructurado, escritura a tabla `events`, rotación de archivos y blackbox replay. Ver [packages/logger/README.md](packages/logger/README.md).

### `packages/shared`

Zod schemas y tipos compartidos (`User`, `RagParams`, `EventType`, roles, etc.). Ver [packages/shared/README.md](packages/shared/README.md).

### `packages/config`

Carga de YAML desde `config/` con overrides por entorno y validación Zod. Ver [packages/config/README.md](packages/config/README.md).

### Blueprint NVIDIA

El submódulo `vendor/rag-blueprint/` contiene el blueprint oficial. No modificarlo desde este repo salvo parches documentados en `patches/`.

---

## Apéndice — decisiones de arquitectura (ADRs)

En `docs/decisions/` hay registros de decisiones (libsql, CJS en packages, proceso único Next.js, Redis obligatorio, etc.). Antes de cambiar áreas cubiertas por un ADR, leé el documento correspondiente.

---

## Apéndice — branches

| Branch | Contenido |
|--------|-----------|
| `experimental/ultra-optimize` | Stack TypeScript (este README) |
| `main` | Historial con stack Python + SvelteKit en producción |

---

## Apéndice — rendimiento y modo desarrollo

- **Turbopack** es el bundler por defecto en Next.js 16; el proyecto incluye `dev:webpack` en `apps/web` si necesitás el modo webpack para depuración.
- **react-scan** puede activarse en desarrollo para perfilar renders; en CI de accesibilidad se usa `NEXT_PUBLIC_DISABLE_REACT_SCAN` para evitar ruido.

---

## Apéndice — soporte y comunidad

- **Issues:** usá las plantillas en `.github/ISSUE_TEMPLATE/` para bugs y features.
- **Contribuciones:** fork, branch desde `experimental/ultra-optimize`, PR con CI en verde.

---

## Apéndice — verificación local rápida

```bash
cd apps/web
bun run test              # lib/
bun run test:components   # componentes
```

En la raíz del monorepo, `bun run test` ejecuta todos los paquetes vía Turborepo.

---

## Apéndice — estructura de `apps/web` (resumen)

| Ruta | Uso |
|------|-----|
| `src/app/(app)/` | Páginas autenticadas (chat, admin, settings, …) |
| `src/app/(auth)/` | Login |
| `src/app/(public)/` | Rutas públicas (p. ej. share) |
| `src/app/api/` | Route Handlers HTTP |
| `src/proxy.ts` | Middleware de Next.js (auth en el edge, RBAC) |
| `src/workers/` | Workers de ingesta y sync |

---

## Apéndice — requisitos de red (desarrollo)

| Puerto | Servicio |
|--------|----------|
| 3000 | Next.js |
| 6379 | Redis |
| 8081 | RAG Server (si no usás `MOCK_RAG`) |

---

## Apéndice — troubleshooting

| Síntoma | Qué revisar |
|---------|-------------|
| Error `REDIS_URL no configurado` | Variable en `.env.local` y Redis en ejecución |
| 401 en rutas API | Cookie `auth_token` o `Authorization: Bearer` + JWT válido |
| Ingesta no avanza | Worker BullMQ, logs en `apps/web/logs/`, `rag ingest status` |
| Tests de visual fallan | `bun run visual:update` si el cambio de UI es intencional |

---

## Apéndice — releases

Las versiones siguen [Semantic Versioning](https://semver.org/). El changelog vive en [CHANGELOG.md](CHANGELOG.md).

---

## Apéndice — créditos

Basado en **NVIDIA RAG Blueprint** y herramientas de la comunidad (Bun, Next.js, Drizzle, Radix, Tailwind, Playwright).

---

## Apéndice — objetivos del producto

1. Un solo binario de aplicación fácil de desplegar (Next.js + SQLite + Redis).
2. Seguridad por defecto: JWT, RBAC, auditoría y sin secretos en el repositorio.
3. Extensibilidad: nuevas rutas API y páginas siguiendo [CONTRIBUTING.md](CONTRIBUTING.md).
4. Observabilidad: logs estructurados, eventos en base de datos y CLI de diagnóstico.

---

## Apéndice — glosario breve

| Término | Significado |
|---------|-------------|
| RAG | Retrieval-Augmented Generation |
| BFF | Backend-for-frontend (API en el mismo proceso que la UI) |
| SSE | Server-Sent Events (streaming HTTP unidireccional) |
| RBAC | Control de acceso basado en roles |


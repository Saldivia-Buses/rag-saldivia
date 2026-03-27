# Onboarding — Arrancar el proyecto en 5 minutos

> Última actualización: 2026-03-26

---

## Prerrequisitos

- [Bun](https://bun.sh) ≥ 1.3 — `curl -fsSL https://bun.sh/install | bash`
- **Redis 7+** (obligatorio — ADR-010): `docker run -d -p 6379:6379 redis:alpine`
- [Docker](https://docs.docker.com/get-docker/) con Docker Compose v2 (opcional, solo para RAG real)
- Git
- Node.js ≥ 22 (recomendado para herramientas del monorepo / CI)

---

## Setup rápido (Ubuntu / WSL2)

```bash
# 1. Clonar en el filesystem Linux nativo (en WSL2: ~/rag-saldivia, NO en /mnt/c/)
cd ~
git clone https://github.com/Camionerou/rag-saldivia
cd rag-saldivia
git checkout experimental/ultra-optimize

# 2. Setup completo (instala deps, configura .env, migra DB, seed)
bun run setup

# 3. Arrancar
bun run dev
# → http://localhost:3000
# → admin@localhost / changeme
```

> **WSL2:** Si `bun install` falla con errores de symlinks, asegurate de estar en `~/` (filesystem Linux) y no en `/mnt/c/`.

---

## Sin Docker — modo mock (desarrollo de UI)

Para trabajar en el frontend sin levantar Docker ni el stack NVIDIA:

```bash
echo "MOCK_RAG=true" >> apps/web/.env.local
bun run dev
```

El servidor simula respuestas del RAG con streaming real. Ideal para:
- Desarrollo de componentes UI
- Tests de integración
- Visual regression con Storybook

---

## Variables de entorno

El `.env.local` vive en `apps/web/.env.local`. Ver `.env.example` para la lista completa.

Las más importantes:

```env
JWT_SECRET=...          # openssl rand -base64 32
REDIS_URL=redis://localhost:6379
SYSTEM_API_KEY=...      # openssl rand -hex 32
RAG_SERVER_URL=http://localhost:8081
DATABASE_PATH=./data/app.db   # en WSL2: usar ruta absoluta
MOCK_RAG=false          # true = desarrollo sin Docker
LOG_LEVEL=INFO
```

---

## Credenciales de desarrollo

El `bun run setup` crea estos usuarios automáticamente:

| Email | Contraseña | Rol |
|---|---|---|
| `admin@localhost` | `changeme` | admin — acceso completo |
| `user@localhost` | `test1234` | user — acceso básico |

---

## Comandos esenciales

### Desarrollo

```bash
bun run dev              # Next.js en :3000 (con hot reload y react-scan activo)
bun run storybook        # Catálogo de componentes en :6006
```

### Tests

```bash
# Suite completa de lógica (270+ tests)
bun run test

# Tests de componentes React (147 tests con happy-dom)
bun run test:components

# Visual regression — comparar componentes contra baseline
bun run test:visual

# Regenerar baseline de visual regression (primera vez o cambio intencional)
bun run visual:update

# Auditoría a11y WCAG AA en páginas críticas
bun run test:a11y

# Todo en un comando
bun run test:ui          # test:components + test:visual
```

### Base de datos

```bash
bun packages/db/src/migrate.ts   # Aplicar migraciones
bun packages/db/src/seed.ts      # Crear datos de prueba
# o via CLI:
rag db migrate
rag db seed
rag db reset
```

### CLI (instalar: `cd apps/cli && bun link`)

```bash
rag status              # Estado del sistema + health check
rag users list          # Lista usuarios
rag collections list    # Lista colecciones
rag ingest status       # Estado de la cola de ingesta
rag audit log           # Últimos eventos
```

---

## Lo que vas a ver en el browser

Al abrir `http://localhost:3000`:

1. **`/login`** — formulario crema cálido con logo "R" navy. Credenciales arriba.
2. **`/chat`** — sidebar crema con lista de sesiones, área central con EmptyPlaceholder
3. **`/admin/users`** — tabla de usuarios con badges de rol (navy/verde/rojo)
4. El NavRail izquierda tiene íconos en crema — toggle ☀️/🌙 abajo para cambiar tema

> **Sistema dark mode:** next-themes usa class-based (`.dark` en `<html>`). El toggle lo cambia. La preferencia del sistema se respeta automáticamente.

---

## Flujo de trabajo típico

```bash
# 1. Arrancar
bun run dev                    # dev server
bun run storybook              # (opcional) para trabajo en componentes

# 2. Desarrollar
# ... editar componentes, páginas, queries ...

# 3. Testear
bun run test                   # lógica pura
bun run test:components        # componentes React

# 4. Antes de commitear
bun run type-check             # tipos TypeScript
# El pre-push hook también corre type-check automáticamente

# 5. Commitear con el formato del proyecto
git commit -m "feat(web): descripción del cambio"
# Tipos válidos: feat fix refactor chore docs test ci perf revert
# Scopes válidos: web cli db config logger shared auth rag chat admin...
```

---

## Referencias

| Doc | Descripción |
|---|---|
| `docs/architecture.md` | Diagrama de arquitectura, flujos de auth y RAG |
| `docs/design-system.md` | Tokens CSS, componentes UI, Storybook |
| `docs/testing.md` | Guía completa de testing por capa |
| `docs/workflows.md` | Git, commits, CI/CD, features |
| `docs/cli.md` | Referencia completa de la CLI |
| `docs/blackbox.md` | Sistema de logging y replay |
| `docs/decisions/` | ADRs — por qué tomamos las decisiones de arquitectura |
| `CLAUDE.md` | Contexto rápido para agentes de IA |

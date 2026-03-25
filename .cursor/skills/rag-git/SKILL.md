---
name: rag-git
description: Git workflow for RAG Saldivia — conventional commits, CHANGELOG, PRs, and pre-push hooks. Use when committing, writing a commit message, updating the CHANGELOG, creating a PR, asking about git conventions, or when the user says "voy a commitear", "cómo escribo el commit", "¿actualizo el changelog?", or "crear PR".
---

# RAG Saldivia — Git Workflow

Reference: `docs/workflows.md` (sección Git)

## Commit Format

```
<type>(<scope>): <description in lowercase>
```

Subject ≤ 100 chars. Full header ≤ 120 chars. Validated automatically by Husky `commit-msg` hook.

### Types permitidos

`feat` `fix` `refactor` `chore` `docs` `test` `ci` `perf` `revert`

### Scopes del proyecto

| Scope | Área |
|-------|------|
| `web` | apps/web |
| `cli` | apps/cli |
| `db` | packages/db |
| `config` | packages/config |
| `logger` | packages/logger |
| `shared` | packages/shared |
| `auth` | lógica de autenticación |
| `rag` | cliente/proxy RAG |
| `chat` | UI de chat |
| `admin` | panel de administración |
| `collections` | gestión de colecciones |
| `ingestion` | worker/queue de ingesta |
| `blackbox` | sistema de eventos/audit |
| `deps` | dependencias |
| `ci` | GitHub Actions |
| `docs` | documentación |
| `changelog` | CHANGELOG.md |
| `plans` | docs/plans/ |
| `setup` | scripts/setup.ts |

## Regla crítica: CHANGELOG antes del commit

**Actualizar `CHANGELOG.md` ANTES de hacer commit**, no después.

Formato bajo `## [Unreleased]`:

```
### Added
- `ruta/archivo.ts`: descripción de qué se agregó — YYYY-MM-DD

### Fixed
- `ruta/archivo.ts`: descripción del bug y cómo se corrigió — YYYY-MM-DD

### Changed
- `ruta/archivo.ts`: qué cambió y por qué — YYYY-MM-DD
```

Categorías: `Added` | `Changed` | `Deprecated` | `Removed` | `Fixed` | `Security`

## Pre-push hook

El hook `pre-push` ejecuta `bun run type-check` automáticamente antes de cada push. Si falla, el push se cancela. No usar `--no-verify`.

## Crear un PR

1. CHANGELOG tiene la entrada correspondiente
2. Branch pusheada al remote
3. `gh pr create` con sección CHANGELOG obligatoria en el body
4. El CI valida: commitlint + changelog check + type-check + tests + lint

## Ejemplos de commits correctos

```
feat(web): agregar paginación en /admin/users
fix(db): corregir removeAreaCollection para filtrar por collectionName
test(logger): agregar tests de reconstructFromEvents
docs: actualizar architecture.md con diagrama de auth
refactor(web): extraer lógica SSE de ChatInterface a useRagStream
chore(deps): actualizar drizzle-orm
```

# Convenciones de commits

## Formato

```
tipo(scope): descripción en inglés — planN fN
```

Enforceado por commitlint (`.commitlintrc.json`). El subject debe ser
**todo lowercase** — no `TypeScript` sino `typescript`.

## Tipos (level 2 — error si no es válido)

| Tipo | Cuándo |
|------|--------|
| `feat` | Nueva funcionalidad |
| `fix` | Corrección de bug |
| `refactor` | Cambio sin cambio de comportamiento |
| `style` | CSS/UI sin cambio de lógica |
| `chore` | Mantenimiento (deps, config, CI) |
| `docs` | Documentación |
| `test` | Agregar o modificar tests |
| `ci` | Cambios de CI/CD |
| `perf` | Mejora de performance |
| `revert` | Revertir un commit anterior |

## Scopes (level 1 — warning si no es válido)

| Scope | Área |
|-------|------|
| `web` | `apps/web/` en general |
| `ui` | componentes UI (`components/ui/`) |
| `chat` | componentes de chat |
| `auth` | autenticación/JWT |
| `rag` | proxy RAG/streaming |
| `admin` | páginas admin |
| `collections` | colecciones |
| `ingestion` | pipeline de ingesta |
| `db` | `packages/db/` |
| `shared` | `packages/shared/` |
| `config` | `packages/config/` |
| `logger` | `packages/logger/` |
| `agents` | `.claude/agents/` |
| `docs` | documentación general |
| `plans` | `docs/plans/` |
| `changelog` | CHANGELOG.md |
| `release` | version bumps, tags |
| `deps` | dependencias |
| `ci` | GitHub Actions, CI/CD |
| `setup` | onboarding, scripts |

## Ejemplos reales del proyecto

```
feat(chat): integrate vercel ai sdk for streaming — plan14 f1
style(ui): apply claude azure tokens to globals.css — plan15 f2
refactor(web): archive aspirational components — plan13 f3
docs(plans): update master plan with roadmap — plan13 f2
test(chat): add component tests for sourcespanel — plan14 f3
chore(agents): rewrite all agents for typescript stack — plan13 f1
chore(release): v1.1.0
```

## Reglas

- Subject en inglés, todo lowercase
- Siempre referenciar plan y fase (excepto releases y hotfixes)
- Sin punto final
- Max 100 chars en subject, 120 en header, 150 en body lines

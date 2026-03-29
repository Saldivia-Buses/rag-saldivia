# Convenciones de versionado

## Formato: semver (MAJOR.MINOR.PATCH)

- **MAJOR** (2.0.0): cambios breaking, nueva arquitectura
- **MINOR** (1.1.0): features nuevas, backwards-compatible
- **PATCH** (1.0.1): bugfixes

## Checklist de release

```bash
# 1. Todos los tests en verde
bunx tsc --noEmit
bun run test
bun run test:components
bun run lint

# 2. Security audit (agent security-auditor)

# 3. CHANGELOG actualizado
# Mover [Unreleased] a [X.Y.Z] — YYYY-MM-DD

# 4. Version bump
# Actualizar version en package.json raíz + apps/web + packages/*

# 5. Commit + tag
git add -A
git commit -m "chore(release): vX.Y.Z"
git tag -a vX.Y.Z -m "release X.Y.Z — [descripción corta]"

# 6. Push
git push origin 1.0.x --tags
```

## Nota para la serie 1.0.x

La serie 1.0.x es **una versión grande** construida por múltiples planes.
No se hacen releases granulares (1.0.1, 1.0.2). Cuando Plans 13-19 estén
listos, se libera como una versión significativa.

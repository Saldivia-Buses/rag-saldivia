# Convenciones de versionado

## Formato: semver (MAJOR.MINOR.PATCH)

| Tipo | Cuándo |
|------|--------|
| MAJOR (2.0.0) | Cambios breaking, nueva arquitectura |
| MINOR (1.1.0) | Features nuevas, backwards-compatible |
| PATCH (1.0.1) | Bugfixes |

## Proceso de release

1. Todos los planes del milestone completados
2. Todos los tests en verde (`tsc`, `test`, `test:components`, `lint`)
3. Security audit aprobado
4. CHANGELOG actualizado con sección `[X.Y.Z]`
5. Version bump en todos los `package.json`
6. Commit: `chore(changelog): vX.Y.Z`
7. Tag: `git tag -a vX.Y.Z -m "Release X.Y.Z — [descripción]"`
8. Push tag → GitHub Release automática

## CHANGELOG format (Keep a Changelog)

```markdown
## [X.Y.Z] — YYYY-MM-DD

### Added
- [feature] (Plan N)

### Changed
- [change] (Plan N)

### Fixed
- [fix] (Plan N)

### Removed
- [removal] (Plan N)
```

## Nota para la serie 1.0.x

La serie 1.0.x NO es una secuencia de releases granulares. Es una versión
grande construida por múltiples planes. Cuando el conjunto esté listo,
se libera como una versión significativa.

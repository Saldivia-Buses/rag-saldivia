# Pull Request Template

> **Nota:** hoy (serie 1.0.x, Enzo solo) no se usan PRs. Se trabaja directo
> en la branch `1.0.x` con un commit por fase. Este template es para cuando
> haya más gente y se active el flujo de feature branches + PRs.

## Summary

<!-- 1-3 bullet points -->

## Plan reference

Plan N — Fases X a Y

## Changes

<!-- Lista de cambios concretos -->

## Scope drift

- [ ] No scope drift — solo archivos planeados

## Testing

- [ ] `bunx tsc --noEmit` → 0 errors
- [ ] `bun run test` → green
- [ ] `bun run test:components` → green (si hay cambios UI)
- [ ] `bun run test:visual` → green (si hay cambios visuales)

## Screenshots

<!-- Si hay cambios visuales, antes/después -->

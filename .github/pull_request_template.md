## Descripción

<!-- Qué hace este PR? Por qué es necesario? -->

## Tipo de cambio

- [ ] `feat` — Nueva funcionalidad
- [ ] `fix` — Corrección de bug
- [ ] `refactor` — Refactorización sin cambio de comportamiento
- [ ] `chore` — Mantenimiento, deps, configuración
- [ ] `docs` — Solo documentación
- [ ] `test` — Solo tests
- [ ] `ci` — Cambios en CI/CD
- [ ] `perf` — Mejora de performance

## Testing

- [ ] Los tests existentes siguen pasando (`make test`)
- [ ] Agregué tests para la funcionalidad nueva (si aplica)
- [ ] Probé manualmente en desarrollo local

## Checklist

- [ ] El código sigue las convenciones del proyecto
- [ ] Los commits siguen Conventional Commits (`type(scope): description`)
- [ ] No hay secrets ni credenciales hardcodeadas

## Traces freeze (Plan 26 → 28)

`services/traces/` está congelado a bug fixes mientras se migra a `pkg/spine` (Plan 28).

- [ ] Este PR **NO** agrega funcionalidad nueva a `services/traces/`, O
- [ ] Agrega funcionalidad nueva justificada — label `traces-exception` aplicado + razón en "Notas para el reviewer"

## Notas para el reviewer

<!-- Algo que el reviewer deba saber? Decisiones de diseño, trade-offs, etc. -->

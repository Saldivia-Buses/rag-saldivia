# scripts/

Scripts de utilidad para setup, operaciones y health check del nuevo stack TypeScript.

## Archivos

| Archivo | Qué hace |
|---|---|
| `setup.ts` | Onboarding completo: preflight check, instalar deps, crear `.env.local`, migrar DB, seed, resumen de estado |
| `health-check.ts` | Verifica el estado de todos los servicios con latencias |

## Uso

```bash
# Setup completo (primera vez)
bun run setup
# o equivalente:
bun scripts/setup.ts

# Solo preflight check (sin instalar nada)
bun scripts/setup.ts --check

# Resetear DB a estado inicial
bun scripts/setup.ts --reset

# Health check
bun scripts/health-check.ts
bun scripts/health-check.ts --json   # output en JSON
```

## Scripts del stack anterior

Los scripts Python/bash del stack original (deploy.sh, setup.sh, smart_ingest.py, etc.)
están en `legacy/scripts/` para referencia.

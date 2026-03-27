# scripts/

Scripts de utilidad para setup, operaciones y health check del nuevo stack TypeScript.

## Archivos

| Archivo | Qué hace |
|---|---|
| `setup.ts` | Onboarding completo: preflight check, instalar deps, crear `.env.local`, migrar DB, seed, resumen de estado |

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
```

## Health del stack

Usá `rag status` o la ruta `GET /api/health` del servidor web.

## Scripts del stack anterior

Los scripts Python/bash del stack original (deploy.sh, setup.sh, smart_ingest.py, etc.)
están en `legacy/scripts/` para referencia.

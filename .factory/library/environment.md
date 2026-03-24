# Environment

Environment variables, external dependencies, and setup notes.

**What belongs here:** Required env vars, external API keys/services, dependency quirks, platform-specific notes.
**What does NOT belong here:** Service ports/commands (use `.factory/services.yaml`).

---

## Required Environment Variables

### Auth Gateway (saldivia/gateway.py)
- `JWT_SECRET` — secret para firmar tokens JWT (dev: cualquier string)
- `SYSTEM_API_KEY` — API key del sistema para llamadas internas
- `BYPASS_AUTH` — si `"true"`, omite verificación de JWT (solo dev/tests)
- `DATABASE_URL` — path a la SQLite DB (default: `config/saldivia.db`)
- `RAG_SERVER_URL` — URL del RAG server (default: `http://localhost:8081`)

### Frontend SvelteKit (services/sda-frontend/.env.local)
- `JWT_SECRET` — mismo secreto que el gateway (para verificar cookies)
- `SYSTEM_API_KEY` — mismo que el gateway
- `GATEWAY_URL` — URL del auth gateway (dev: `http://localhost:9000`)

El archivo `.env.local` ya existe en `services/sda-frontend/` con valores para desarrollo local.

## Dependencies

### Python
- `uv` como gestor de dependencias y runner
- `uv sync` instala desde `pyproject.toml` y `uv.lock`
- Python 3.14.3 disponible en `/opt/homebrew/bin/python3`

### Node.js
- Node 24.13.1
- `npm install` en `services/sda-frontend/` (node_modules ya presente)

## Platform Notes

- macOS darwin 25.3.0, Apple Silicon (10 cores, 16GB RAM)
- No hay Docker corriendo en dev local
- El deploy de producción es en workstation Ubuntu 24.04 (1x RTX PRO 6000 Blackwell)
- Puerto 3000 es producción Docker — NO usar en desarrollo/tests

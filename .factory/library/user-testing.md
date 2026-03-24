# User Testing

Testing surface, required tools, setup requirements, and resource cost classification.

---

## Validation Surface

### Surface 1: Python Backend (pytest)
- **Tool:** `uv run pytest saldivia/tests/ -v`
- **What it tests:** Gateway endpoints, database methods, config loading, auth
- **Setup:** Solo `uv sync` (ya disponible). No requiere servicios externos corriendo.
- **Auth/bootstrap:** Tests usan `BYPASS_AUTH=true` implícito via mocks — no necesitan gateway real.

### Surface 2: TypeScript/Svelte Frontend (svelte-check)
- **Tool:** `cd services/sda-frontend && npm run check`
- **What it tests:** Type correctness, Svelte 5 runes usage, import resolution
- **Setup:** `npm install` (node_modules ya presente)
- **No requiere:** Gateway ni servidor corriendo

### Surface 3: Browser (agent-browser)
- **Tool:** `agent-browser` skill
- **URL:** `http://localhost:5173` (SvelteKit dev server)
- **Prerequisitos:**
  - Usuario levanta `cd services/sda-frontend && npm run dev` manualmente antes de validación
  - Auth Gateway corriendo en `localhost:9000` (usuario confirma)
  - Credenciales admin: las del `.env.local` (usuario: admin, password: ver config/saldivia.db)
- **Auth flow:** Login via `/login` con credenciales admin antes de verificar rutas `/admin/*` o `/audit`
- **Rutas verificadas en esta mission:**
  - `/(app)/audit` — página principal de audit log
  - `/(app)/admin/rag-config` — admin RAG config (Fase 10, verificar regresiones)

## Validation Concurrency

### Surface 1 (pytest)
- Recursos por instancia: ~100MB RAM, 1 CPU
- Baseline sistema: ~10.7GB RAM usada de 16GB, ~5.3GB disponible
- Headroom efectivo: 5.3GB × 0.7 = 3.7GB
- **Max concurrent: 5** (uso mínimo de recursos)

### Surface 2 (svelte-check)
- Recursos por instancia: ~200MB RAM
- **Max concurrent: 5**

### Surface 3 (agent-browser)
- Chrome ~300MB RAM por instancia + SvelteKit dev server ~200MB (compartido)
- Headroom: 3.7GB. 5 instancias = ~1.7GB. Bien dentro del budget.
- **PERO:** SvelteKit dev server es uno solo para todos los validators — no escala N instancias.
- Un solo dev server puede manejar múltiples tabs del browser.
- Cada validator abre su propia pestaña/ventana; el dev server es compartido.
- **Max concurrent: 2** (conservador dado que cada validator necesita estado de auth propio y el gateway es un single process)

## Testing Infrastructure Notes

- No hay framework de unit test configurado para componentes Svelte — validación funcional es via agent-browser
- Los tests de Python usan `httpx.AsyncClient` con `TestClient` de FastAPI
- `BYPASS_AUTH=true` en env de tests evita necesidad de gateway real para tests unitarios
- El archivo `.env.local` en el frontend contiene `JWT_SECRET=dev-secret-local` para desarrollo

## Known Constraints

- El gateway en localhost:9000 debe estar corriendo para que el frontend funcione en modo dev
- Sin el gateway, las páginas que hacen fetch al BFF retornan errores pero no crashean el servidor de dev
- Auto-refresh tests requieren esperar ~30s — los validators deben presupuestar ese tiempo

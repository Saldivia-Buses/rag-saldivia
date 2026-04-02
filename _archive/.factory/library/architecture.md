# Architecture

Architectural decisions, patterns discovered, and conventions to follow.

---

## Stack

- **Backend:** FastAPI (`saldivia/gateway.py`) + SQLite (`saldivia/auth/database.py`)
- **Frontend:** SvelteKit 5 BFF (`services/sda-frontend/`)
- **Auth:** JWT cookies; gateway como único punto de autenticación
- **Testing:** pytest (backend), svelte-check + agent-browser (frontend)

## Key Patterns

### SQLite
- Usar `_ts()` helper para timestamps — `detect_types=PARSE_DECLTYPES` crashea con date-only strings
- Migrations: `ALTER TABLE ... ADD COLUMN` en `try/except sqlite3.OperationalError` (solo `duplicate column name`)
- Usar **LEFT JOIN** para queries que unen tablas donde pueden existir filas huérfanas
- Parametrized queries siempre — nunca interpolación de strings

### Gateway (FastAPI)
- Auth: `get_user_from_token` → None para no-autenticado (retorna 403, debería ser 401 en algunos endpoints — known inconsistency)
- Admin check: `if user.role != Role.ADMIN: raise HTTPException(403)`
- `log_action()`: siempre pasar `success=` explícitamente
- Instancia de módulo: `rag_config = ConfigLoader()` a nivel de módulo

### Frontend (SvelteKit 5)
- Svelte 5 runes: `$state()`, `$effect()`, `$props()`, `$bindable()`
- Deep-linking: `goto(url, { replaceState: true, keepFocus: true })` para filtros
- Auto-refresh: `setInterval` dentro de `$effect` con cleanup (`return () => clearInterval(id)`)
- BFF pattern: `+server.ts` como proxy delgado al gateway; guard de auth antes de forward
- Design tokens: `--bg-surface`, `--border`, `--accent`, `--text-primary`, `--text-faint`
- Null safety: siempre `entry.field ?? '—'` para campos nullable

### CSV Export
- Client-side siempre (Blob URL + `<a>` click programático)
- Nunca hacer request adicional al backend para export
- Null fields → `''` empty string (no "null" ni "undefined")

## Module Map

| Módulo | Responsabilidad |
|--------|----------------|
| `saldivia/gateway.py` | FastAPI app, todos los endpoints, auth, proxy RAG |
| `saldivia/auth/database.py` | SQLite AuthDB, todos los métodos de persistencia |
| `saldivia/auth/models.py` | Dataclasses: User, Area, AuditEntry, etc. |
| `saldivia/config.py` | ConfigLoader: YAML profiles + env merge + persistencia |
| `services/sda-frontend/src/lib/server/gateway.ts` | BFF client al gateway, types TypeScript |

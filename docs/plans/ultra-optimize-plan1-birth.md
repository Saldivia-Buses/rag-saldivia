# Plan 1 — Ultra-Optimización: Nacimiento del Monorepo

> **Estado:** COMPLETADO — 2026-03-24
> **Branch:** `experimental/ultra-optimize`
> **Líneas originales:** ~260 → comprimido a resumen post-ejecución

---

## Qué se hizo

Reescritura completa del proyecto. Se reemplazó el stack original (gateway Python + frontend SvelteKit + Redis + 2 procesos) por un servidor único Next.js con TypeScript puro.

### Antes → Después

```
ANTES:  Usuario → SvelteKit :3000 → gateway.py :9000 → RAG :8081
DESPUÉS: Usuario → Next.js :3000 ————————————————————→ RAG :8081
```

- **2 procesos → 1** (Next.js único)
- **2 lenguajes → 1** (Python + TS → solo TypeScript)
- **~6.4k líneas Python eliminadas**
- **Redis eliminado** como dependencia (reemplazado por SQLite queue)

### Fases ejecutadas

| Fase | Qué | Resultado |
|------|-----|-----------|
| 0 | Onboarding cero-fricción | `scripts/setup.ts`, `.env.example`, modo `MOCK_RAG` |
| 1 | Git workflow | Commitlint + Husky + 3 GitHub Actions (CI, CD, release) |
| 2 | Monorepo + packages | Turborepo + `packages/{shared,db,config,logger}` |
| 3a | Auth core | JWT en cookie HttpOnly, middleware RBAC, login/logout/refresh |
| 3b | DB layer | Drizzle ORM + SQLite, queries tipadas reemplazando 875 líneas Python |
| 3c | RAG proxy + SSE | Proxy streaming hacia RAG :8081, verificación pre-stream |
| 3d | Collections + ingestion | Upload drag & drop, ingestion queue en SQLite, worker TS |
| 3e | Chat UI | Sesiones, streaming SSE, crossdoc, Server Actions |
| 3f | Admin UI | Usuarios, áreas, permisos, config RAG, audit log |
| 3g | Settings | Perfil, contraseña, preferencias |
| 4 | CLI | ~15 comandos con Commander + @clack/prompts, modo REPL |
| 5 | Black box | Logger backend/frontend, suggestions, replay, rotación logs |
| 6 | Docs + limpieza | CHANGELOG, CLAUDE.md, architecture.md, onboarding.md |

### Stack resultante

| Componente | Tecnología |
|---|---|
| Framework | Next.js 15 App Router |
| Runtime | Bun + Turborepo |
| DB | Drizzle ORM + SQLite |
| Auth | JWT (jose) + cookie HttpOnly |
| Validación | Zod compartido |
| CLI | Commander + @clack/prompts |
| Logging | packages/logger (4 módulos) |

### Commits (branch `experimental/ultra-optimize`)

| Fase | Commit | Descripción |
|------|--------|-------------|
| Setup | `621b661` | chore(plans): init experimental/ultra-optimize branch |
| F0 | `bda0f1f` | feat(setup): Fase 0 — onboarding cero-friccion |
| F1 | `ea24557` | feat(ci): Fase 1 — git workflow profesional |
| F2 | `d987491` | feat(monorepo): Fase 2 — monorepo Turborepo + paquetes compartidos |
| F3a | `8e83e51` | feat(auth): Fase 3a — auth core (JWT, RBAC, middleware, endpoints) |
| F3b-g | `c9b5ba3` | feat(web): Fases 3b-3g — servidor unico apps/web completo |
| F4 | `050451c` | feat(cli): Fase 4 — CLI de clase mundial |
| F5 | `7eba734` | feat(blackbox): Fase 5 — audit endpoints + UI + black box replay |
| F6 | `c1fa0fe` | docs: Fase 6 — documentacion completa + gitignore |
| Cierre | `fbab789` | feat: completar todas las tareas pendientes del plan |

**Post-plan: Migración legacy + docs:**

| Commit | Descripción |
|--------|-------------|
| `1bd0efe` | refactor: mover stack legacy a carpeta legacy/ |
| `124c35d` | refactor: mover scripts y archivos Python a legacy/ |
| `d616e0f` | refactor: mover docs del stack viejo a legacy/docs/ |
| `9446911` | docs: actualizar README.md y scripts/README.md para el nuevo stack |
| `f77c61b` | docs(changelog): registrar todos los fixes del troubleshooting WSL2 |
| `b7cf470` | docs: aclarar que link-libsql.sh es solo para WSL2, no para Ubuntu nativo |
| `faeaa24` | docs(plans): renombrar ultra-optimize.md a ultra-optimize-plan1-birth.md |
| `7ade2ea` | chore(deps): mover legacy/ a gitignore — stack anterior fuera del repo |
| `951d59e` | docs: solidificar documentacion y corregir pipeline de tests |

**Post-plan: Saga WSL2 / DB driver (entre Plan 1 y Plan 2):**

Troubleshooting de deployment — migración `bun:sqlite → better-sqlite3 → @libsql/client` y fixes de bundling webpack/Next.js.

| Commit | Descripción |
|--------|-------------|
| `06a0e57` | fix(db): migrar de better-sqlite3 a bun:sqlite (nativo, sin compilacion) |
| `09282cd` | fix(db): bun:sqlite nativo, init.ts SQL puro, packageManager field |
| `ddb4306` | fix(web): configurar turbopack.root y outputFileTracingRoot para WSL2 |
| `1a0324e` | fix(web): corregir outputFileTracingRoot a __dirname |
| `2b83a51` | fix(web): marcar paquetes internos como serverExternalPackages para bun:sqlite |
| `e6f83b1` | fix(web): quitar transpilePackages conflictivo con serverExternalPackages |
| `060a19a` | fix(db,web): volver a better-sqlite3 para Next.js + transpilePackages correcto |
| `06f1eb7` | fix(db): usar @libsql/client (JS puro, sin compilacion) — **solución final** |
| `16a018c` | fix(web): extensionAlias .js→.ts para workspace packages en webpack |
| `8aaa15b` | fix(web): excluir @libsql/client y drizzle-orm del bundling webpack |
| `e5f6924` | fix(packages): quitar extensiones .js de imports + quitar type:module para webpack CJS |
| `4b17a6b` | fix(web): webpack externals para excluir libsql chain nativa del bundling |
| `c74165a` | fix(packages): eliminar todas las extensiones .js de imports relativos |
| `5401753` | fix(web): quitar drizzle-orm de externals para evitar conflicto de instancias |
| `711055d` | fix(db): agregar relaciones Drizzle para habilitar queries con with |
| `96912d3` | fix(shared): login email acepta admin@localhost (sin TLD) |
| `aebc026` | fix(wsl2): scripts para setup en WSL2 + bun.lock |

> **Nota:** Este plan fue un día maratónico donde se construyó todo el stack desde cero.
> El código original se archivó en `_archive/`. Los planes 2-12 refinaron y pulieron lo que este plan creó.

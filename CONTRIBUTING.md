# Contribuir a RAG Saldivia

Gracias por contribuir. Este documento describe cómo preparar el entorno, ejecutar tests, el flujo de PR y patrones de código.

## 1. Setup del entorno de desarrollo

1. **Instalar Bun 1.3+**  
   ```bash
   curl -fsSL https://bun.sh/install | bash
   export PATH="$HOME/.bun/bin:$PATH"
   ```

2. **Clonar y entrar al branch de trabajo**  
   ```bash
   git clone https://github.com/Camionerou/rag-saldivia
   cd rag-saldivia
   git checkout 2.0.5
   ```

3. **Redis** (obligatorio)  
   ```bash
   docker run -d -p 6379:6379 redis:alpine
   ```

4. **Variables de entorno**  
   ```bash
   cp .env.example .env.local
   ```
   Editar `.env.local` como mínimo:
   - `JWT_SECRET` — cadena larga (p. ej. `openssl rand -base64 32`)
   - `REDIS_URL=redis://localhost:6379`
   - `SYSTEM_API_KEY` — si vas a usar la CLI contra APIs admin

5. **Instalar dependencias y base de datos**  
   ```bash
   bun install
   bun run setup
   ```

6. **Arrancar la app**  
   ```bash
   MOCK_RAG=true bun run dev
   ```
   Abrir http://localhost:3000 — credenciales seed: `admin@localhost` / `changeme`.

## 2. Cómo correr los tests

| Comando | Descripción |
|---------|-------------|
| `bun run test` | Todos los tests de lógica (paquetes + `apps/web`) vía Turborepo |
| `cd apps/web && bun run test:components` | Tests de componentes React (happy-dom) |
| `cd apps/web && bun run test:visual` | Regresión visual Playwright |
| `cd apps/web && bun run test:a11y` | Auditoría WCAG con axe-playwright |
| `cd apps/web && bun run test:e2e` | E2E Playwright (requiere dev server y Redis según spec) |

Si un test visual falla por un cambio de UI intencional:

```bash
cd apps/web && bun run visual:update
```

Y commitear los PNG actualizados en `tests/visual/snapshots/`.

## 3. Convenciones de commit (Conventional Commits)

Formato: `type(scope): descripción en minúscula`

**Tipos:** `feat`, `fix`, `refactor`, `chore`, `docs`, `test`, `ci`, `perf`

El hook **commitlint** (en `husky` / `commit-msg`) rechaza mensajes que no sigan el formato.

**Buenos ejemplos:**

- `feat(web): agregar filtro por área en tabla de usuarios`
- `fix(db): corregir query de sesiones con borde de fecha`
- `docs: actualizar api.md con endpoint de notificaciones`

**Malos ejemplos:**

- `WIP` (sin tipo ni descripción)
- `Fixed stuff` (no convencional)
- `FEAT: BIG UPDATE` (mayúsculas y sin scope claro)

## 4. Flujo de PR

1. Crear branch desde `2.0.5`.
2. Nombre sugerido: `feat/descripcion-corta` o `fix/descripcion-corta`.
3. Abrir PR **hacia** `2.0.5`.
4. El **CI** debe pasar (tests, lint, type-check según workflow).
5. Completar el **PR template** (`.github/pull_request_template.md`).

## 5. Cómo agregar una página nueva

1. Crear el archivo en `apps/web/src/app/(app)/tu-ruta/page.tsx` (o `(auth)` / `(public)` según corresponda).
2. Por defecto es **Server Component**; añadir `"use client"` solo si necesitás estado, efectos o APIs del navegador.
3. Si debe aparecer en la navegación, actualizar el componente de navegación (p. ej. `NavRail`).
4. Las rutas bajo `(app)` no suelen ser públicas: la autenticación la aplica el middleware en `apps/web/src/proxy.ts` (rutas públicas están en `PUBLIC_ROUTES` dentro de ese archivo).
5. Agregar tests de componente en `apps/web/src/components/__tests__/` cuando aplique.

## 6. Cómo agregar una ruta API nueva

1. Crear `apps/web/src/app/api/tu-ruta/route.ts` (o segmentos dinámicos `[id]`).
2. En handlers, usar `extractClaims()` de `@/lib/auth/jwt` o helpers `requireUser` / `requireAdmin` si existen en el código.
3. Document the endpoint in the service README (`services/{name}/README.md`) and the corresponding `docs/services/{name}.md`.
4. Para flujos críticos, considerar cobertura E2E en `apps/web/tests/e2e-playwright/`.

## 7. Cómo agregar una tabla a la base de datos

1. Editar `packages/db/src/schema.ts`.
2. Añadir queries en `packages/db/src/queries/` (nuevo archivo o existente).
3. Aplicar cambios (según flujo del proyecto: `bunx drizzle-kit push` o migraciones si están en uso).
4. Añadir tests en `packages/db/src/__tests__/`.
5. Actualizar el diagrama ER en `packages/db/README.md`.

## 8. Architecture Decision Records (ADRs)

- Crear un ADR cuando una decisión técnica tenga **trade-offs** no obvios y deba quedar fijada para el equipo.
- Plantilla: `docs/decisions/000-template.md`.
- Numeración: el siguiente número libre es **011** (revisar `docs/decisions/` antes de crear).
- Los ADRs **no se reescriben** en silencio: si cambia la decisión, crear un ADR nuevo que cite y reemplace al anterior.

## 9. Debugging y troubleshooting

| Problema | Acción |
|----------|--------|
| Tests fallan tras un pull | `bun install`, `bun run setup`, Redis arriba |
| `REDIS_URL no configurado` | Variable en `.env.local` y contenedor Redis |
| Errores de tipos | `bun run type-check` y revisar cambios en `packages/shared` |
| CLI recibe 401 | `SYSTEM_API_KEY` alineada con el servidor; rutas admin requieren rol o API key |
| Lint en commit | `bun run lint` y corregir; `lint-staged` solo en archivos tocados |

---

Preguntas de seguridad: ver [SECURITY.md](SECURITY.md).

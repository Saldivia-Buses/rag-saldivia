# ADR-002: CJS sobre ESM en paquetes del monorepo

**Fecha:** 2026-03-24
**Estado:** Aceptado

---

## Contexto

El monorepo tiene cuatro paquetes compartidos (`packages/db`, `packages/shared`, `packages/config`, `packages/logger`) que son importados tanto desde `apps/web` (Next.js, webpack) como desde `apps/cli` (Bun directo).

TypeScript 6 + Bun soportan ESM nativamente. La tendencia del ecosistema es ESM. Sin embargo, Next.js y webpack tienen comportamientos específicos con módulos ES que generaron problemas durante la construcción inicial.

## Opciones consideradas

- **ESM (`"type": "module"` en package.json):** estándar moderno, soporte nativo en Bun. Contras: webpack (usado por Next.js) tiene manejo inconsistente de paquetes workspace ESM; los imports relativos requieren extensión `.js` explícita en TypeScript ESM, pero webpack los resuelve mal cuando los paquetes tienen `"type": "module"`; varios errores de resolución encontrados en la práctica (`Cannot find module`, `ERR_REQUIRE_ESM`).
- **CJS (sin `"type": "module"`):** formato CommonJS, el default de Node.js. webpack lo maneja sin problemas; Bun también lo soporta transparentemente. Los imports relativos en TypeScript no requieren extensión explícita en el output.
- **Dual package (CJS + ESM):** publicar ambos formatos con `exports` condicionales. Overkill para un monorepo privado; agrega complejidad de build sin beneficio real.

## Decisión

Elegimos **CJS** (omitir `"type": "module"` en todos los `packages/*/package.json`) porque elimina una clase entera de errores de resolución de módulos en webpack/Next.js sin costo funcional.

Adicionalmente: eliminar extensiones `.js` de todos los imports relativos en TypeScript (`import { foo } from "./bar"` en lugar de `import { foo } from "./bar.js"`) — las extensiones `.js` son requeridas para ESM puro pero son incompatibles con webpack cuando el paquete es CJS.

## Consecuencias

**Positivas:**
- Cero errores de resolución de módulos en `apps/web` al importar paquetes workspace.
- `bun test` y `bun run` funcionan sin configuración adicional.
- TypeScript type-check (`tsc`) pasa sin errors de módulos.

**Negativas / trade-offs:**
- No se puede usar `import.meta.url` ni `import.meta.dirname` en los paquetes. Usar `__dirname` (CJS) o `path.resolve()` en su lugar.
- Si en el futuro Next.js mejora el soporte de ESM workspace packages, este ADR podría revisarse.
- Los paquetes no son "publishables" como ESM sin una etapa de build adicional — aceptable para un monorepo privado.

## Referencias

- `packages/*/package.json` — ausencia de `"type": "module"`
- CHANGELOG.md: "eliminadas extensiones `.js` de todos los imports relativos"
- `CLAUDE.md` — sección "Patrones importantes": "CJS sobre ESM"

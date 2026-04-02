# 17 — Riesgos y Deuda Tecnica

## Deuda tecnica conocida

### Alta prioridad

#### 1. JWT revocacion no verificada en Edge
**Archivo:** `proxy.ts`
**Problema:** El middleware Edge verifica la firma del JWT pero NO verifica revocacion en Redis (ioredis no corre en Edge). La revocacion se verifica en `extractClaims()` dentro de route handlers (Node.js).
**Impacto:** Con access tokens de 15min (Plan 26), la ventana de ataque se redujo de 24h a 15min.
**Mitigacion actual:** Access token corto (15min) + refresh rotation + operaciones sensibles via route handlers.
**Solucion futura:** Upstash Redis REST en Edge para verificacion de revocacion.

#### ~~2. Credenciales de external_sources en texto plano~~ RESUELTO (Plan 28)
~~**Problema:** Credenciales en texto plano.~~
**Estado:** Cifrado AES-256-GCM implementado con lazy migration. Las credenciales se cifran automaticamente al guardar y se descifran al leer.

#### ~~3. Componentes de messaging sin tests~~ RESUELTO (Plan 29)
~~**Problema:** 19 componentes sin tests.~~
**Estado:** 19/19 componentes con tests (79 tests). Cobertura 100% de componentes de messaging.

### Media prioridad

#### 4. formatPretty con complejidad ciclomatica 29
**Archivo:** `packages/logger/src/backend.ts`
**Problema:** La funcion mas compleja del proyecto. Dificil de mantener y testear.
**Impacto:** Riesgo de bugs en logging que pasen desapercibidos.
**Solucion:** Refactorizar en funciones mas chicas.

#### ~~5. Dark mode sin refinar~~ RESUELTO (Plan 30)
~~**Problema:** Tokens dark no refinados.~~
**Estado:** Tokens corregidos (accent-fg, text-white→semantic), Mermaid theme-aware. Pendiente: visual contrast audit (F3-F4).

#### ~~6. ChatInterface complejidad 22~~ RESUELTO (Plan 28)
~~**Problema:** 643 lineas, dificil de mantener.~~
**Estado:** Descompuesto a ~360 lineas (-44%). Extraidos ChatEmptyState + ChatMessages. Con tests.

#### ~~7. Tests E2E limitados~~ PARCIALMENTE RESUELTO (Plan 29)
~~**Problema:** Pocos flujos E2E.~~
**Estado:** E2E expandido a auth flows, chat interaction, admin access control (~22 tests). A11y en 8 paginas.

### Baja prioridad

#### 8. CLI archivada
**Problema:** La CLI esta en `_archive/` — no se actualizo para el stack nuevo.
**Impacto:** Sin herramienta CLI para operaciones de administracion desde terminal.
**Solucion:** Re-escribir CLI cuando haya necesidad real.

#### 9. Workers en estado aspiracional
**Archivo:** `apps/web/src/workers/`
**Problema:** `ingestion.ts` y `external-sync.ts` existen pero pueden no estar completamente integrados.
**Impacto:** Ingesta de documentos podria no funcionar end-to-end.
**Solucion:** Verificar integracion con BullMQ al deployar.

#### 10. Storybook coverage parcial
**Problema:** No todos los componentes tienen stories.
**Impacto:** Documentacion visual incompleta.
**Solucion:** Agregar stories incrementalmente.

---

## Riesgos arquitecturales

### SQLite en single-tenant
**Riesgo:** Si un cliente necesita >50 usuarios simultaneos, SQLite podria ser bottleneck.
**Mitigacion:** Drizzle soporta Postgres. Migracion estimada: 1 dia.
**Probabilidad:** Baja (el producto es single-tenant by deployment).

### Next.js como proceso unico
**Riesgo:** Si la carga crece, no se puede escalar API y frontend independientemente.
**Mitigacion:** La logica de negocio esta en `packages/` — extraible a un API server separado.
**Probabilidad:** Baja a mediano plazo.

### Redis como dependencia requerida
**Riesgo:** Redis caido = JWT revocation no funciona + BullMQ no funciona.
**Mitigacion:** `isRevoked()` retorna false si Redis no esta disponible (fail-open para auth). BullMQ tiene retry.
**Probabilidad:** Baja (Redis es estable en workstation local).

### Dependencia de NVIDIA RAG Blueprint
**Riesgo:** Si NVIDIA cambia la API del blueprint, el adapter se rompe.
**Mitigacion:** `ai-stream.ts` aisla la transformacion. Cambiar el adapter no afecta el resto.
**Probabilidad:** Media (NVIDIA actualiza el blueprint periodicamente).

---

## Archivos de maxima fragilidad

Modificar estos archivos sin entender completamente su impacto puede causar fallas en cascada:

| Archivo | Riesgo | Tests protectores |
|---------|--------|------------------|
| `proxy.ts` | Rompe auth para toda la app | Unit tests de proxy |
| `lib/auth/jwt.ts` | Rompe login/logout/refresh | jwt.test.ts |
| `lib/safe-action.ts` | Rompe TODAS las server actions | Varios action tests |
| `globals.css` | Cambia look de toda la app | Visual regression |
| `schema/core.ts` | Requiere migracion de DB | 198 DB tests |
| `schema/chat.ts` | Requiere migracion de DB | DB tests |
| `schema/messaging.ts` | Requiere migracion de DB | messaging tests |
| `component-test-setup.ts` | Rompe 158 component tests | — (meta-test) |
| `connection.ts` | Rompe toda la DB | DB tests |
| `redis.ts` | Rompe auth + queue | redis.test.ts |

---

## Metricas de salud del codebase (actualizado post Plans 26-30)

| Metrica | Valor | Estado |
|---------|-------|--------|
| Tests pasando | ~1,059 | OK (x2.8 vs anterior) |
| Type errors | 0 (verificado Plan 29) | OK |
| Lint warnings | 0 (verificado Plan 29) | OK |
| Complejidad maxima | 29 (formatPretty) | Atencion |
| Componentes sin test | ~28 (de 75) | Mejoro (era ~44 de 68) |
| Component coverage | 63% (era 35%) | Mejoro |
| Hooks coverage | 86% (6/7) | Mejoro (era 43%) |
| ADRs superadas | 2 (008, 011) | OK — marcadas |
| Archivos archivados | 123 | OK — en _archive/ |
| Security headers | 5/6 (falta CSP) | Mejoro (era 0/6) |
| JWT rotation | Access 15min + refresh 7d | Resuelto (era 24h unico) |
| Credentials cifrado | AES-256-GCM | Resuelto (era texto plano) |
| SQLite PRAGMAs | WAL + foreign_keys + busy_timeout | Resuelto (era sin PRAGMAs) |
| Next.js output | standalone + compress | Resuelto (era default) |

---

## Recomendaciones priorizadas (actualizado)

1. ~~Cifrar credentials~~ HECHO (Plan 28)
2. ~~Tests messaging~~ HECHO (Plan 29)
3. ~~ChatInterface decomposition~~ HECHO (Plan 28)
4. **Antes de deployar (Plan 19):** Verificar end-to-end en workstation fisica.
5. **Siguiente:** Self-healing Error UX (Plan 32 en escritura) — mejorar experiencia de errores para testers.
6. **Mediano plazo:** Conectores externos (Plan 33) + SSO (Plan 34).
7. **Pendiente:** Refactorizar formatPretty (complejidad 29), CSP header.
8. **Antes de multi-tenant:** Evaluar migracion a Postgres y Redis cluster.

---

## Dead code y candidatos a limpieza

### Confirmado: `usePresence` hook sin consumidor

**Archivo:** `apps/web/src/hooks/usePresence.ts` — exportado pero no importado en ningun componente, page ni archivo del proyecto. Verificar si `PresenceIndicator.tsx` lo usa internamente; si no, es dead code puro.

### Confirmado: `events-cleanup.ts` sin consumidor

**Archivo:** `packages/db/src/queries/events-cleanup.ts` (24 lineas) — no referenciado en ningun worker, cron ni route. Posiblemente dead code.

### Tablas sin UI activa (intencional, no eliminar)

| Tabla | Feature archivada en |
|-------|---------------------|
| `projects`, `project_sessions`, `project_collections` | `_archive/app/projects/` |
| `saved_responses` | `_archive/app/saved/` |
| `session_shares` | `_archive/app/share/` |
| `external_sources` | `_archive/admin/external-sources/` |
| `bot_user_mappings` | Feature futura (Slack/Teams) |

Estas tablas existen en el schema y sus queries estan testeadas. Mantenerlas tiene costo cero. Eliminarlas requeriria migracion destructiva.

### Queries con UI archivada

`reports.ts`, `webhooks.ts`, `external-sources.ts`, `projects.ts`, `shares.ts` — funcionalidad completa pero sin pagina activa. Son preparacion futura, no dead code.

### Export que podria ser privado

`getRequiredRole()` en `lib/auth/rbac.ts` — solo lo usa `canAccessRoute()` en el mismo archivo. Exportarlo expande la surface area sin necesidad.

---

## Internacionalizacion (i18n) — deuda estructural

### Estado actual

Toda la UI esta hardcodeada en espanol. No hay sistema i18n. Los strings estan directos en JSX:

```tsx
// Ejemplo real — string directo en componente
<EmptyPlaceholder title="Sin sesiones" description="Crea una nueva sesion para empezar" />
<Button>Guardar cambios</Button>
<p className="text-fg-muted">No hay colecciones disponibles</p>
```

**Esto es intencional para v1** — el producto es para empresas argentinas y los primeros testers hablan espanol.

### Impacto si se necesita otro idioma

- Hay que extraer CADA string de CADA componente (68 componentes activos)
- No hay archivos de traduccion, no hay `t()` function, no hay namespaces
- Los mensajes de error de server actions estan en espanol
- Los mensajes de la API tambien (`"No autenticado"`, `"Acceso denegado"`)

### Recomendacion

No actuar ahora — no es necesario para v1. Pero si en algun momento se necesita ingles o portugues:

1. Adoptar `next-intl` (la mas natural para App Router)
2. Extraer strings con un codemod automatico
3. Esfuerzo estimado: 2-3 dias para los 68 componentes + API routes + actions

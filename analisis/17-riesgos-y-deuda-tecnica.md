# 17 — Riesgos y Deuda Tecnica

## Deuda tecnica conocida

### Alta prioridad

#### 1. JWT revocacion no verificada en Edge
**Archivo:** `proxy.ts`
**Problema:** El middleware Edge verifica la firma del JWT pero NO verifica revocacion en Redis (ioredis no corre en Edge). La revocacion se verifica en `extractClaims()` dentro de route handlers (Node.js).
**Impacto:** Entre el logout y el proximo request a un route handler, un JWT revocado podria acceder a Server Components que solo dependen del middleware.
**Mitigacion actual:** Los Server Components leen headers `x-user-*` que el middleware ya valido. Las operaciones sensibles pasan por route handlers que si verifican revocacion.
**Solucion futura:** Migrar a middleware que corra en Node.js (Next.js 16 lo soporta), o usar Redis compatible con Edge (Upstash).

#### 2. Credenciales de external_sources en texto plano
**Archivo:** `packages/db/src/schema/core.ts`
**Problema:** `external_sources.credentials` almacena JSON en texto plano. El comentario dice "cifrar con SYSTEM_API_KEY en prod" pero no esta implementado.
**Impacto:** Si alguien accede al archivo SQLite, las credenciales de Google Drive/SharePoint/Confluence estan expuestas.
**Solucion:** Implementar cifrado simetrico antes de persistir.

#### 3. Componentes de messaging sin tests
**Archivos:** 19 componentes en `components/messaging/`
**Problema:** Plan 25 agrego el sistema de messaging completo sin component tests.
**Impacto:** Regresiones no detectadas en la UI de messaging.
**Solucion:** Escribir tests con el patron existente (happy-dom + afterEach cleanup).

### Media prioridad

#### 4. formatPretty con complejidad ciclomatica 29
**Archivo:** `packages/logger/src/backend.ts`
**Problema:** La funcion mas compleja del proyecto. Dificil de mantener y testear.
**Impacto:** Riesgo de bugs en logging que pasen desapercibidos.
**Solucion:** Refactorizar en funciones mas chicas.

#### 5. Dark mode sin refinar (Plan 20 pendiente)
**Problema:** Los tokens dark existen pero no estan refinados para todos los componentes.
**Impacto:** Contraste pobre en algunos componentes en dark mode.
**Solucion:** Plan 20 pendiente.

#### 6. ChatInterface complejidad 22
**Archivo:** `components/chat/ChatInterface.tsx`
**Problema:** El componente mas complejo de la UI. Dificil de mantener.
**Impacto:** Cambios en el chat tienen alto riesgo de regresion.
**Solucion:** Descomponer en sub-componentes mas chicos.

#### 7. Tests E2E limitados
**Problema:** Los tests E2E cubren pocos flujos.
**Impacto:** Bugs de integracion entre paginas no detectados.
**Solucion:** Expandir cobertura E2E con Playwright.

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

## Metricas de salud del codebase

| Metrica | Valor | Estado |
|---------|-------|--------|
| Tests pasando | ~380 | OK |
| Type errors | 0 (asumido, ultima verificacion en Plan 23) | OK |
| Lint warnings | 0 (asumido) | OK |
| Complejidad maxima | 29 (formatPretty) | Atention |
| Componentes sin test | ~30 (messaging + chat avanzado) | Deuda |
| ADRs superadas | 2 (008, 011) | OK — marcadas |
| Archivos archivados | 123 | OK — en _archive/ |
| Dependencias outdated | No verificado | Verificar |

---

## Recomendaciones priorizadas

1. **Antes de deployar (Plan 19):** Verificar que la revocacion JWT funciona end-to-end. Cifrar credenciales de external_sources.
2. **Siguiente sprint:** Tests para componentes de messaging (19 componentes sin cobertura).
3. **Mediano plazo:** Refactorizar formatPretty y ChatInterface para reducir complejidad.
4. **Cuando haya testers:** Expandir tests E2E con flujos reales de usuario.
5. **Antes de multi-tenant:** Evaluar migracion a Postgres y Redis cluster.

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

# Plan 08 Review — Backend Hardening

**Fecha:** 2026-04-05
**Resultado:** CAMBIOS REQUERIDOS
**Plan:** `docs/plans/2.0.x-plan08-backend-hardening.md`

---

## Resumen ejecutivo

Plan solido, bien estructurado, con hallazgos reales verificados contra el
codigo. La gran mayoria de los 52 hallazgos son correctos y las soluciones
propuestas son razonables. Sin embargo, hay errores facticos, hallazgos
mal clasificados, omisiones significativas, y la Fase 4 (gRPC) es un
scope creep que deberia separarse. Lo que sigue es el analisis detallado.

---

## 1. Errores facticos (cosas que el plan dice mal)

### C1. Audit logging — parcialmente incorrecto

**El plan dice:** "ningun servicio llama `audit.NewWriter`"

**La realidad:** Cuatro servicios YA usan audit logging:
- `services/auth/internal/service/auth.go:69` -- `audit.NewWriter(db)` + 7 Write calls
- `services/chat/internal/service/chat.go:69` -- `audit.NewWriter(db)` + 3 Write calls
- `services/search/cmd/main.go:61` + `handler/search.go:77` -- audit en queries
- `services/notification/internal/service/notification.go:61` -- audit en ReadAll

**Servicios que SI faltan:** ingest, platform, agent, feedback, traces

**Impacto:** El hallazgo sigue siendo valido pero subestima el trabajo hecho.
Reducir el scope a los servicios que realmente faltan y reconocer los existentes.
El esfuerzo real es menor al estimado.

### C2. Rate limiting — omite Traefik rate limit existente

**El plan dice:** "No existe ninguna implementacion"

**La realidad:** `deploy/traefik/dynamic/prod.yml:110-114` tiene:
```yaml
rate-limit:
  rateLimit:
    average: 100
    burst: 200
    period: 1s
```

Esto ya cubre el "Global: X req/s por IP". Lo que falta es rate limiting
**granular a nivel aplicacion** (por usuario, por endpoint, brute-force
por cuenta). El plan deberia reconocer que existe un rate limit global
en Traefik y que C2 agrega la capa aplicativa encima.

**Reclasificacion sugerida:** Bajar de Critical a High. Traefik ya protege
contra floods basicos. Lo critico era "zero rate limiting" -- eso no es
cierto.

### H10. Feedback HealthScore — parcialmente incorrecto

**El plan dice:** "retorna placeholder"

**La realidad parcial:** El endpoint **tenant-level** (`/v1/feedback/health-score`)
SI retorna un placeholder (linea 231 de feedback.go). Pero el endpoint
**platform-level** (`/v1/platform/feedback/tenants`) ya consulta la tabla
`tenant_health_scores` con queries reales (platform_feedback.go:42-49).

El fix correcto es que el tenant-level HealthScore debe leer de la tabla
`tenant_health_scores` del platform DB filtrando por el tenant del request.
Pero esto requiere que el feedback handler tenga acceso al platformDB
(actualmente solo tiene tenantDB para el endpoint regular).

---

## 2. Priorizacion incorrecta

### M9 (Security scans) esta en Fase 2 pero deberia ser Fase 7

M9 es un cambio de CI config (gosec/trivy exit codes). Esta agrupado con
fixes de seguridad en handlers/JWT (Fase 2, "Security Depth"), pero es
puramente build/CI. Moverlo a Fase 7 donde ya hay H6 (.dockerignore),
M5 (OpenAPI), M13 (go.mod cleanup), y otros items de build.

### L4 (writeJSONError) esta en Fase 2 pero deberia ser Fase 8

L4 es un refactor cosmético de string concatenation a json.Marshal en
`pkg/middleware/auth.go:122`. No es security depth -- el unico riesgo
es si `msg` contiene caracteres JSON especiales, y los mensajes son
todos hardcoded strings en el middleware. Moverlo a Fase 8 con los
demas cleanups.

### L5 (DeleteJobByID ownership) esta en Fase 2 pero es menor

DeleteJobByID sin `user_id` solo se usa **internamente** en
`services/ingest/internal/service/ingest.go:168` para cleanup cuando
NATS publish falla -- en el mismo request que creo el job (mismo user
context). El handler externo usa `DeleteJob` que SI filtra por `user_id`.
Esto es defense-in-depth, no un bug activo. Bajar a low y mover a Fase 8.

### H5 (Cache service structs) esta en Fase 8 pero deberia ser mas alto

Crear un nuevo `service.Auth` struct por request en multi-tenant mode
(auth/handler/auth.go:84) implica que cada login:
1. Resuelve pool (OK, cacheado)
2. Crea service.NewAuth con audit.NewWriter + repository.New
3. Descarta todo al final del request

Esto es una allocation innecesaria que escala linealmente con requests.
No es critico pero es High (performance bajo carga). Deberia estar en
Fase 5 con el feature wiring, no en Fase 8 con los cleanups.

---

## 3. Omisiones significativas (cosas que faltan del plan)

### MISSING-1. Traefik configs desactualizados — CRITICO

**Problema:** Tanto `deploy/traefik/dynamic/dev.yml` como `prod.yml`
referencian un servicio `rag` en puerto 8004 que ya no existe (deprecated,
reemplazado por `agent`). Los siguientes servicios no tienen routes:
- **agent** (8004) -- reemplaza a rag, necesita route
- **search** (8010) -- endpoint REST, necesita route
- **traces** (8009) -- endpoint REST, necesita route

Ademas, `deploy/docker-compose.prod.yml` todavia define un servicio `rag`
(linea 239-267) con `Dockerfile: services/rag/Dockerfile` que no existe.
Faltan definiciones para agent, search, y traces en el compose prod.

**Severidad:** Critical. Sin esto, estos servicios son inaccesibles en dev
y produccion.

### MISSING-2. docker-compose.prod.yml dual routing conflict

`docker-compose.prod.yml` define routes via Docker labels (linea 168-171 etc.)
Y `deploy/traefik/dynamic/prod.yml` define routes via static file config.
Si ambos estan activos, hay conflicto de routing. Necesita decision:
eliminar uno o documentar que son mutuamente excluyivos.

### MISSING-3. NATS `nc.Close()` vs `nc.Drain()` inconsistencia

El plan menciona M1 (estandarizar NATS connections con `natspub.Connect()`)
pero no menciona que 6 servicios usan `defer nc.Close()` mientras 3 usan
`defer nc.Drain()`. Close descarta mensajes en flight; Drain espera a que
se completen. Todos deberian usar Drain. Esto deberia ser parte de M1.

Servicios con `nc.Close()`: auth, chat, ws, feedback, notification, ingest
Servicios con `nc.Drain()`: agent, traces, platform

### MISSING-4. Plan 07 Phase 4 nunca se ejecuto

Plan 07 Phase 4 decia: "todos los `cmd/main.go` usan `natspub.Connect()`".
Plan 08 M1 lo incluye de nuevo. Esto significa que Phase 4 de Plan 07 no
se completo. El plan deberia reconocer esto explicitamente para evitar
confusion sobre que se supone que ya esta hecho vs que se esta repitiendo.

### MISSING-5. Feedback handler raw SQL sin sqlc

`services/feedback/internal/handler/platform_feedback.go` tiene queries
SQL inline (raw strings) en los handlers en vez de usar sqlc. Lineas 41-49,
103-112, 172-180. Esto es inconsistente con el patron del resto del sistema
y deberia generarse con sqlc como todo lo demas. Si el plan tiene C3
(fix sqlc configs), deberia incluir "migrar feedback raw SQL a sqlc".

### MISSING-6. Feedback handler error swallowing

`services/feedback/internal/handler/feedback.go` lineas 51-71 llaman queries
y si fallan, logean el error pero siguen ejecutando y retornan un response
parcial con zero-values. Esto devuelve datos enganiosos al frontend en vez
de un 500.

### MISSING-7. Shutdown error handling inconsistente

Algunos servicios checkean el error de `srv.Shutdown()` (auth, ws, feedback)
mientras otros lo ignoran (chat, traces, search, agent, notification,
platform, ingest). Menor, pero parte de la estandarizacion que el plan busca.

### MISSING-8. Auth handler `nc.Close()` en vez de `nc.Drain()`

Auth service usa `defer nc.Close()` (linea 73) pero PUBLICA eventos a NATS
(login events, security events). Si el servicio se cierra justo despues
de publicar un evento, Close puede descartar el mensaje. Deberia usar Drain.

---

## 4. Fase 4 (gRPC) — Scope creep

### El problema

Fase 4 es un proyecto de 16-20h que implementa gRPC desde cero:
generar codigo, implementar servers, migrar WS Hub, inter-service calls.
Esto NO es hardening. Es una **feature nueva** que agrega un protocolo
de comunicacion que hoy no existe.

El plan dice "52 hallazgos de una auditoria de seguridad/reliability" pero
gRPC no es un hallazgo de auditoria -- es una decision arquitectonica
que el spec menciona como deseable. El plan 08 se contamina con scope
de un plan de features.

### Recomendacion

Extraer Fase 4 a un plan propio (Plan 09: gRPC inter-service). Reordenar
las fases restantes. El plan 08 queda en ~65-85h en vez de ~85-105h,
y el scope es limpio: solo hardening, cero features nuevas.

### Nota sobre protos

Los proto files existen (1182 lineas) pero cero codigo generado y cero
uso. No hay ni `buf.yaml` ni `buf.gen.yaml`. El `make proto` target no
funciona. Esto es un proyecto real, no un tweak.

---

## 5. Viabilidad tecnica — notas por hallazgo

### C2. Rate limiting

La propuesta de usar `golang.org/x/time/rate` para in-memory es correcta
para un solo nodo. Pero en multi-tenant, el rate limit por usuario necesita
ser compartido entre instancias del mismo servicio (si se escala). Para MVP
single-node, in-memory esta bien. Para produccion con replicas, Redis-backed
con sliding window es lo correcto. La propuesta menciona ambos pero no
decide. Decidir ahora: **in-memory para MVP, migration path a Redis
documentado**.

### C3. sqlc schema paths

El fix es correcto. Una cosa a verificar: las queries de feedback no estan
en sqlc (ver MISSING-5). El plan deberia agregar un sqlc.yaml para feedback
que apunte al schema correcto.

### C4. PostgreSQL SSL

La propuesta de appendear `?sslmode=require` en el resolver es peligrosa si
la URL ya tiene query params (se rompe el parsing). Mejor:
```go
u, _ := url.Parse(pgURL)
q := u.Query()
if q.Get("sslmode") == "" {
    q.Set("sslmode", "require")
}
u.RawQuery = q.Encode()
```

### H4. Pagination

La propuesta es correcta. Considerar usar cursor-based pagination
(`WHERE created_at > $cursor`) en vez de OFFSET para mensajes (que pueden
tener miles de rows). OFFSET con pagination profunda es O(N). Cursor-based
es O(1) para acceso secuencial.

### H9. ListActiveUsers LATERAL JOIN

La reescritura es correcta y significativamente mejor que la correlated
subquery. Aprobado.

### M8. Dead Letter Queue

La propuesta es correcta pero incompleta. Necesita definir:
1. Retention del DLQ (cuanto tiempo guardar mensajes fallidos?)
2. Alerting (como avisa cuando hay mensajes en DLQ?)
3. Replay (como re-procesar mensajes del DLQ?)
El plan dice "loggea + alerta via notification service" pero si
notification service esta caido, el DLQ tambien falla. Necesita un
fallback independiente (al menos slog).

### M14. API key caching

La propuesta de un struct `CachedModelConfig` sin APIKey es la correcta.
La alternativa de encriptar con `pkg/crypto` agrega complejidad innecesaria
-- si Redis se compromete y tiene la key encriptada, el atacante
probablemente tambien tiene acceso al encryption key (esta en el mismo
host). Ir con el struct reducido.

### L21. Resolver mutex panic protection

La descripcion del plan es confusa. Dice "usar `defer r.mu.Lock()` despues
del Unlock" lo cual no tiene sentido literal. El patron correcto es:

```go
r.mu.Unlock()
defer r.mu.Lock() // re-acquires even if the next line panics
pool, err := pgxpool.NewWithConfig(ctx, config)
// r.mu.Lock() happens via defer
```

Pero esto cambia la semantica: el defer se ejecuta al final de la
funcion, no inmediatamente despues de la linea. El patron real que
necesitan es:

```go
r.mu.Unlock()
pool, createErr := func() (p *pgxpool.Pool, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("pool creation panicked: %v", r)
        }
    }()
    return pgxpool.NewWithConfig(ctx, config)
}()
r.mu.Lock()
```

O mas simple: `pgxpool.NewWithConfig` no deberia panic en condiciones
normales. Si panics, chi Recoverer lo atrapa. El mutex queda en estado
inconsistente pero el proceso va a crashear de todas formas. Esto es un
fix de extrema paranoia. L (low) es la clasificacion correcta.

---

## 6. Dependencias entre fases — correcciones

### El grafo dice Fase 2 depende de Fase 1 (audit + rate limiting)

Incorrecto para la mayoria de items. H8 (GetSession ownership), H11
(JWT JTI), M11 (WriteTimeout), L4 (writeJSONError) no dependen de
audit logging ni rate limiting. Solo M9 (CI scans) podria argumentarse
que depende de C3 (sqlc fix) para que el build sea verde.

**Correccion:** Fase 2 puede ejecutarse en paralelo con Fase 1. La unica
dependencia real es que Fase 3 depende de C3 (sqlc fix) porque las queries
deben compilar contra el schema correcto.

### El grafo dice Fase 4 depende de Fase 2 ("security hardened antes de gRPC")

Esto es artificial. gRPC es un protocolo de comunicacion interna. No
necesita que H8 o H11 esten resueltos para empezar. Si gRPC se extrae
a su propio plan (recomendado), esta dependencia desaparece.

### El grafo dice Fase 5 depende de Fase 3 Y Fase 4

Fase 5 items (EnabledModules, NATS standardization, guardrails from config)
NO dependen de gRPC (Fase 4) ni de pagination (Fase 3). H2 (EnabledModules)
necesita platform DB queries que no tienen nada que ver con pagination.
La unica dependencia real: M7 (JetStream API migration) podria hacerse
independientemente.

### Grafo corregido (sin gRPC)

```
Fase 1 (critical)     ←── bloquea todo
  ├─► Fase 2 (parallel OK) -- security depth
  └─► Fase 3 (parallel OK) -- DB hardening
Fase 2 + 3 completadas
  └─► Fase 5 (feature wiring + NATS)
Fase 5
  └─► Fase 6 (testing)
Fase 6
  └─► Fase 7 (build/CI)
Fase 1-7 completadas
  └─► Fase 8 (infra + cleanup)
```

Fases 2 y 3 son completamente independientes entre si y pueden
ejecutarse en paralelo.

---

## 7. Estimaciones de tiempo

| Fase | Plan dice | Mi estimacion | Razon |
|---|---|---|---|
| 1 | 12-15h | 10-12h | C1 es ~50% menos trabajo del que dice (4 servicios ya tienen audit) |
| 2 | 4-5h | 3-4h | Items son puntuales, 30min cada uno es correcto |
| 3 | 10-12h | 10-12h | Correcto. Pagination + LATERAL JOIN + batch insert es trabajo real |
| 4 | 16-20h | 20-30h | Subestimado. gRPC from scratch + migration es mas de lo que dice |
| 5 | 14-16h | 12-14h | Razonable, algunos items son mas simples de lo que parece |
| 6 | 8-10h | 10-14h | Tests de integracion siempre toman mas de lo estimado |
| 7 | 8-10h | 6-8h | La mayoria son cambios de config, no codigo |
| 8 | 12-15h | 14-18h | Backup + CrowdSec + restore testing es subestimado |

**Total sin gRPC:** ~65-82h (vs ~85-105h con gRPC)

---

## 8. Riesgos no mencionados

### R1. Pagination breaks frontend

Agregar LIMIT/OFFSET a queries que antes retornaban todo (H4) va a romper
frontends que asumen que reciben la lista completa. Necesita coordinacion
con el frontend: agregar parametros de pagination a los API calls y
manejar responses paginadas.

### R2. sqlc schema change puede invalidar generated code

C3 cambia los schema paths. Si las migraciones centralizadas tienen
diferencias sutiles con las migraciones viejas (que ya se borraron),
sqlc puede generar structs diferentes. Verificar con `git diff` despues
de regenerar.

### R3. Rate limiting puede bloquear usuarios legitimos

Si el rate limit de AI es 30 req/min y un usuario power usa el agent
intensivamente, va a recibir 429s. Necesita mecanismo de override por
tenant/usuario (config resolver ya lo soporta).

### R4. NATS per-service auth puede romper dev workflow

C5 cambia de un token compartido a per-service auth. Si dev sigue usando
un token simple, los tests manuales siguen funcionando. Pero si alguien
corre docker-compose.dev.yml con el nuevo config, los servicios no van a
poder conectar. Documentar que dev mantiene auth simple.

### R5. gRPC duplica surface area sin beneficio inmediato

Si se mantiene Fase 4 en este plan: agregar gRPC ademas de REST duplica
la surface area de cada servicio (mas codigo, mas tests, mas bugs
potenciales) sin beneficio inmediato porque el WS Hub todavia funciona
via NATS. El beneficio de gRPC se materializa solo cuando hay suficiente
trafico inter-service. Para MVP, REST entre servicios es suficiente.

---

## 9. Lo que esta bien

- **Estructura general:** fases ordenadas por severidad, cada una con
  verificacion explicita. Formato claro y actionable.
- **C3 (sqlc fix):** hallazgo correcto, solucion limpia, verificacion trivial.
- **C4 (SSL):** hallazgo correcto, solucion pragmatica (dev sin SSL, prod
  con SSL). Bien que no fuerza SSL en dev.
- **C5 (NATS per-service auth):** hallazgo critico real. La config de NATS
  existe pero no se usa. Completar y deployar es correcto.
- **H4 + H9 (pagination + query optimization):** hallazgos reales que
  escalaran con usuarios. La reescritura LATERAL JOIN es correcta.
- **H11 (JWT JTI auto-generation):** hallazgo correcto. Sin JTI, el
  blacklist check no funciona (no hay ID para blacklistear).
- **M10 (batch insert):** `pgx.CopyFrom` es la recomendacion correcta.
- **M14 (API key caching):** hallazgo de seguridad real. La solucion
  de struct reducido es la correcta.
- **H7 (parent contexts en consumers):** hallazgo correcto, verificado
  en notification/consumer.go:149 y feedback/consumer.go:168.
- **L6 (notification purge):** pragmatico y necesario.
- **L11 (indice redundante):** correcto, el UNIQUE ya crea indice.

---

## 10. Recomendaciones finales

1. **Extraer Fase 4 (gRPC) a su propio plan.** Mantiene Plan 08 limpio como
   hardening puro. gRPC es un feature, no un fix.

2. **Corregir errores facticos** (C1 audit ya existe parcialmente, C2 rate
   limit ya existe en Traefik, H10 platform HealthScore ya funciona).

3. **Agregar hallazgos faltantes:** Traefik configs desactualizados
   (MISSING-1, critical), prod compose con rag fantasma, feedback raw SQL.

4. **Reclasificar** C2 a High (Traefik ya cubre lo basico), mover M9/L4/L5
   a fases correctas, subir H5 de Fase 8 a Fase 5.

5. **Corregir el grafo de dependencias.** Fases 2 y 3 son paralelas. La
   mayoria de dependencias inter-fase son artificiales.

6. **Agregar notas de riesgo** para pagination (frontend coordination),
   rate limiting (power users), y NATS auth (dev workflow).

7. **Decidir M14:** usar struct reducido (no encripcion).
   **Decidir C2:** in-memory para MVP, Redis-backed documentado para scale.
   **Decidir L3:** mantener single Redis (ya decidido, bien).

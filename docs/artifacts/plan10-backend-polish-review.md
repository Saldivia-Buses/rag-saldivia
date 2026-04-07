# Plan 10 Review -- Backend Polish

**Fecha:** 2026-04-05
**Resultado:** CAMBIOS REQUERIDOS

---

## 1. Priorizacion

El orden de fases es correcto. Security first (Fases 1-2), luego performance (3), DX (4-5), gRPC expansion (6-7), y polish final (8-9). El grafo de dependencias es coherente.

Unica observacion: Fase 7 es extremadamente liviana (agregar 4 lineas de health check a search). Podria fusionarse con Fase 6 para evitar un PR de 5 lineas.

---

## 2. Bloqueantes

### B1. Fase 5: `llm.ChatClient` interface tiene signature incorrecta

**Archivo referido:** `pkg/llm/client.go`

El plan propone:
```go
type ChatClient interface {
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    SimplePrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}
```

El codigo real tiene:
```go
func (c *Client) Chat(ctx context.Context, messages []Message, tools []ToolSchema, temperature float64, maxTokens int) (*ChatResponse, error)
func (c *Client) SimplePrompt(ctx context.Context, prompt string, temperature float64, maxTokens ...int) (string, error)
```

Problemas:
1. `ChatRequest` no existe como tipo. La firma real tiene 4 parametros individuales.
2. `SimplePrompt` real recibe `(ctx, prompt, temperature, maxTokens...)` -- no `(ctx, systemPrompt, userPrompt)`.

**Fix:** la interface debe reflejar las signatures reales:
```go
type ChatClient interface {
    Chat(ctx context.Context, messages []Message, tools []ToolSchema, temperature float64, maxTokens int) (*ChatResponse, error)
    SimplePrompt(ctx context.Context, prompt string, temperature float64, maxTokens ...int) (string, error)
    Model() string
    Endpoint() string
}
```

Alternativamente, si se quiere simplificar la interface, crear un `ChatRequest` struct y refactorear `Client.Chat()` para aceptarlo. Pero eso rompe todos los call sites (agent, search, ingest). Mucho mas scope del que dice el plan.

### B2. Fase 6: `sdamw.UserIDFromContext(ctx)` referido como tal en gRPC handler, pero no existe con ese nombre

El plan dice:
```go
userID := sdamw.UserIDFromContext(ctx)
```

El codigo real en `pkg/middleware/rbac.go:45` tiene exactamente `UserIDFromContext` -- esto es correcto. Retiro el bloqueante. Pero hay un problema diferente: el plan dice que el gRPC handler delegara al mismo `ChatService` interface, pero `ChatService` requiere `userID string` como parametro explícito en cada metodo (e.g., `CreateSession(ctx, userID, title, collection)`). El gRPC handler necesita extraerlo del context y pasarlo. Esto funciona, pero el plan no lo explicita -- podria causar confusion durante implementacion. No es bloqueante pero si ambiguo.

### B3. Fase 6: WS Hub guarda JWT raw en Client struct -- riesgo de seguridad

El plan propone:
```go
JWT string // raw JWT for forwarding to gRPC services
```

Esto significa que el JWT se mantiene en memoria indefinidamente (todo el tiempo que dure la conexion WebSocket). Si el token expira o se revoca, el WS Hub seguira enviando el JWT original al Chat service via gRPC, y el gRPC interceptor lo rechazara.

Esto no es un bug -- es una limitacion arquitectural que necesita documentarse. El frontend tendra que reconectar el WebSocket cuando renueve el token, o el Hub necesita un mecanismo para actualizar el JWT.

**No es bloqueante** pero si se implementa sin esa nota, las mutations dejaran de funcionar a los 15 minutos (access token expiry).

**Fix propuesto:** documentar en el plan que:
1. Los mutations via WS funcionan solo mientras el access token es valido (15 min).
2. El frontend debe reconectar el WS despues de un token refresh.
3. Alternativamente, agregar un message type `authenticate` para actualizar el JWT del client sin reconectar.

---

## 3. Debe corregirse

### M1. Fase 1: `FailOpen: true` para servicios non-auth es una decision de seguridad que necesita justificacion explicita

El plan dice: "disponibilidad > seguridad" para servicios non-auth. Pero esto significa que si Redis cae, un token revocado (e.g., despues de logout o compromise) seguira siendo aceptado por 7 servicios. Solo auth lo rechazara.

Para el MVP de 3 usuarios, esto es aceptable. Pero el plan deberia documentar:
1. POR QUE es aceptable ahora (MVP, 3 users, Redis uptime esperado ~100%)
2. CUANDO cambiar a `FailOpen: false` (production con mas usuarios)

### M2. Fase 2: `requirePlatformAdmin` no verifica blacklist

El plan correctamente identifica la falta de header stripping. Pero `requirePlatformAdmin` (linea 370-392 de platform.go) verifica JWT directamente con `sdajwt.Verify()`, NO con el middleware `sdamw.Auth()`. Esto significa que NO chequea el token blacklist.

Si un platform admin hace logout (token se blacklistea), puede seguir haciendo requests al platform service hasta que el token expire.

**Fix:** agregar blacklist check en `requirePlatformAdmin`:
```go
if blacklist != nil {
    if claims.ID == "" || revoked, _ := blacklist.IsRevoked(ctx, claims.ID); revoked {
        writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "token revoked"})
        return
    }
}
```

O migrar platform a usar `sdamw.AuthWithConfig()` como middleware (mas limpio pero mas trabajo).

### M3. Fase 2: MFA encryption key no se propaga en multi-tenant mode

El plan dice:
> "En multi-tenant mode (`handler.NewMultiTenantAuth`), pasar la key al handler y propagarla cuando se crea un service per-tenant."

El codigo actual de `resolveService` (auth handler.go:93):
```go
svc := service.NewAuth(pool, h.jwtCfg, tenantID, slug, h.publisher)
```

Esto no pasa blacklist NI encryption key. El plan solo habla de pasar encryptionKey, pero tambien falta pasar blacklist en multi-tenant mode. La Fase 2 deberia arreglar ambos o documentar que el blacklist multi-tenant queda para otro PR.

### M4. Fase 3: batch insert pierde la semantica de partial failure

El codigo actual maneja partial failure: si 5 de 200 paginas fallan, guarda las 195 que funcionaron y reporta `partial failure`. El approach propuesto de `DELETE + CopyFrom` es todo-o-nada dentro de una transaccion.

Esto cambia la semantica: un unico page con datos invalidos (e.g., JSON column demasiado grande) causa que las 200 paginas fallen. El plan lo reconoce en R5 ("TX con DELETE + COPY es idempotente") pero no menciona la perdida de partial failure.

**Fix:** documentar que la semantica ahora es all-or-nothing y que un retry via NATS (MaxDeliver: 3) mitiga el fallo.

### M5. Fase 4: Auth `PublicRoutes()` incluye `Logout` como ruta publica

El plan propone:
```go
func (h *Auth) PublicRoutes(...) chi.Router {
    r.Post("/v1/auth/logout", h.Logout)
    ...
}
```

Pero `Logout` necesita un JWT valido para saber cual token revocar. Actualmente en `main.go:150`, `Logout` esta FUERA del grupo protegido, lo cual es correcto (acepta tokens expirados para logout). Pero el plan la pone en `PublicRoutes` sin auth middleware, que es el estado actual -- correcto.

Sin embargo, mirando el handler `Logout` mas de cerca: necesita el refresh token del body Y opcionalmente el access token JTI. Si no tiene el auth middleware, no puede obtener el JTI del access token via headers. El estado actual funciona porque Logout parsea el access token directamente.

Esto NO es un problema del plan -- es el estado actual. Pero vale la pena verificar que `Logout` sigue funcionando sin el auth middleware.

### M6. Fase 5: `natspub.EventPublisher` interface ya existe como duplicacion

El plan dice "auth y ingest ya definen `EventPublisher` localmente con el mismo signature. Migrar a usar `natspub.EventPublisher` elimina duplicacion." Correcto. Pero:

- `services/auth/internal/service/auth.go:48-50` define `EventPublisher` con `Notify` solamente.
- `services/auth/internal/handler/auth.go:36-39` define `EventPublisher` con `Notify + Broadcast`.
- La interface propuesta en `natspub` tiene `Notify + Broadcast`.

El service layer de auth no usa `Broadcast`. Si se cambia el service layer a usar `natspub.EventPublisher`, se le exige `Broadcast` que no necesita.

**Fix:** dejar la interface en `natspub` como esta (Notify + Broadcast), pero NO forzar a los service layers a importarla. Los service layers pueden seguir con sus interfaces locales (mas estrechas). Solo migrar donde ambas operaciones se necesitan.

### M7. Fase 8: OpenAPI con swaggo agrega una dependencia pesada al go.mod de 3 servicios

`swaggo/swag` + `swaggo/http-swagger` traen un arbol de dependencias significativo. El plan no evalua alternativas mas livianas:
- `huma` (integrado con chi, genera OpenAPI sin annotations)
- Escribir los specs YAML a mano (mas tedioso pero zero deps)
- `ogen` (genera server + client desde OpenAPI spec)

Para un MVP de 3 usuarios, agregar annotations swaggo a handlers es trabajo manual que va a bitrot rapidamente (R4 en riesgos lo reconoce). Dado que la audiencia de la doc es "agentes de IA" (bible.md), un YAML estatico escrito a mano podria ser mas util que annotations auto-generadas.

**Sugerencia:** reconsiderar si OpenAPI vale la pena en este momento. Si se mantiene, limitar a 1 servicio (auth) como prueba antes de expandir a 3.

---

## 4. Sugerencias

### S1. Fase 1: centralizar la inicializacion de Redis + blacklist en un helper

7 servicios van a tener el mismo bloque de 12 lineas copiado. Crear un helper en `pkg/security`:
```go
func InitBlacklist(ctx context.Context, redisURL string) (*security.TokenBlacklist, func()) {
    // returns (blacklist, cleanup) -- blacklist is nil if Redis unavailable
}
```

### S2. Fase 6: considerar que `go func()` para mutations es un goroutine leak risk

```go
go func() {
    resp := h.mutations.HandleMutation(client, msg)
    client.SendMessage(resp)
}()
```

Si el gRPC call a Chat service se cuelga (timeout, network partition), esta goroutine queda bloqueada indefinidamente. Usar `context.WithTimeout` derivado del client context.

### S3. Fase 9: `.air.toml` en la raiz no funciona para servicios individuales

El `cmd` en `.air.toml` es `go build -o ./tmp/main ./cmd/main.go`. Pero cada servicio tiene su propio `go.mod` y `cmd/main.go` en `services/{name}/`. El `make dev-%` target hace `cd services/$* && air`, lo cual funcionaria solo si cada servicio tiene su propio `.air.toml` o si `air` detecta el `go.mod` local.

El plan crea un `.air.toml` en la raiz Y en `.scaffold/`. El de la raiz no se usa para nada si `make dev-%` hace `cd services/$*`. Deberia crear un `.air.toml` en cada servicio, o hacer que `make dev-%` pase `-c` con una config centralizada.

### S4. Fase 9: pool tuning env var deberia ir en Fase 1 (no Fase 9)

Si se va a agregar `POOL_MAX_CONNS` al resolver, tiene mas sentido hacerlo cuando se agrega `REDIS_URL` a todos los servicios (Fase 1), no al final. Es un cambio de 5 lineas y evita tocar el resolver dos veces.

---

## 5. Scope creep

### SC1. Fase 8 (OpenAPI) es scope creep

OpenAPI annotations son una feature de documentacion, no de "polish". El plan dice "no agrega features nuevas" pero OpenAPI es una feature nueva con dependencias nuevas, trabajo de mantenimiento continuo, y un endpoint nuevo (`/swagger/*`). Si se incluye, deberia estar en un plan aparte o reducirse a "evaluar swaggo vs alternativas" sin implementar.

### SC2. WS Hub preloading se menciona en el inventario pero no tiene fase

El inventario describe WS preloading como un performance gap, pero ninguna fase lo aborda. Correcto (no todo gap necesita solucion), pero deberia estar en "Fuera de scope" explicitamente para evitar que alguien intente meterlo.

---

## 6. Estimaciones

Las estimaciones son realistas para un developer experimentado en este codebase.

| Fase | Estimacion plan | Mi estimacion | Nota |
|------|----------------|---------------|------|
| 1 | 2-3h | 2-3h | OK, mecanico |
| 2 | 1-2h | 2-3h | MFA encryption wiring en multi-tenant mode es mas complejo |
| 3 | 2-3h | 2-3h | OK |
| 4 | 2-3h | 2-3h | OK |
| 5 | 2-3h | 1-2h | Solo agregar interfaces, no refactorear call sites |
| 6 | 4-6h | 6-8h | gRPC + WS mutations + JWT forwarding es complejo |
| 7 | 1-2h | 30min | Son 5 lineas |
| 8 | 4-6h | 6-8h | Annotations en 12+ handlers son tediosas |
| 9 | 1-2h | 1h | OK |

Total estimado: 22-33h (plan dice 20-30h). La diferencia esta en Fases 2, 6 y 8.

---

## 7. Missing

### MISS1. Blacklist en multi-tenant auth service

En multi-tenant mode, `resolveService()` crea servicios auth sin blacklist (`service.NewAuth()` sin `SetBlacklist()`). Esto deberia estar en Fase 2 junto con el MFA key wiring.

### MISS2. Graceful shutdown del gRPC server en Chat

Fase 6 muestra el gRPC server de Chat pero no incluye `grpcSrv.GracefulStop()` en el shutdown section, como si hace search (line 126). Deberia estar en la lista de cambios.

### MISS3. No hay tests unitarios propuestos

Ninguna fase propone tests nuevos. Fase 5 (interfaces) habilita mocking pero no escribe los mocks. Fase 6 (gRPC Chat server + mutations) deberia tener al menos un test de integracion.

---

## 8. Lo que esta bien

1. El inventario es excelente. Numeros de linea, archivos exactos, conteo de servicios afectados. Se nota que se leyo el codigo.
2. Security first es correcto y la justificacion es clara.
3. `FailOpen: true` con fallback graceful es el approach correcto para non-auth services en un MVP.
4. La decision de NO cambiar `PoolMaxConns = 4` es correcta y bien justificada.
5. El approach de `DELETE + CopyFrom` para batch insert es correcto tecnicamente.
6. La separacion de `PublicRoutes()` y `ProtectedRoutes()` para auth es el patron correcto.
7. El grafo de dependencias entre fases es preciso.
8. "Fuera de scope" esta bien definido (excepto SC2).
9. El checklist de scope drift es un buen mecanismo de control.
10. El plan es honesto sobre riesgos (R1-R6) y los mitiga correctamente.

# Security Audit -- Astro Service (services/astro/) -- 2026-04-06

## Resumen ejecutivo

Full audit of the Astro service (v0.1.0), a new Go microservice implemented
across 19 PRs. The service computes astrological charts using CGO (swephgo)
and optionally calls an LLM for narration via SSE. Overall architecture is
solid: Ed25519 JWT auth, sqlc-generated queries with proper tenant+user scoping,
MaxBytesReader on all POST bodies, and generic error messages. Two medium and
several low-severity issues found. No critical or high-severity vulnerabilities.

**Veredicto: APTO para produccion** (con las correcciones MEDIUM recomendadas).

---

## CRITICOS (bloquean deploy)

Ninguno encontrado.

---

## ALTOS (corregir antes de produccion)

Ninguno encontrado.

---

## MEDIOS (backlog prioritario)

### M1. FailOpen:true en auth middleware permite bypass si Redis cae

**Archivo:** `services/astro/cmd/main.go:91`

```go
authMw := sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{
    Blacklist: blacklist,
    FailOpen:  true,  // <-- si Redis no responde, se acepta el token
})
```

**Impacto:** Si Redis cae, la blacklist de tokens revocados deja de funcionar.
Un token revocado (logout, cambio de password) seguiria siendo valido mientras
Redis este caido. Esto es consistente con otros servicios no-criticos del
framework (chat, search, ingest, agent, traces, notification todos usan
`FailOpen: true`) -- solo auth y feedback-admin usan `FailOpen: false`.

**Fix:** Evaluar si para produccion conviene `FailOpen: false` para este
servicio. Si la disponibilidad es prioridad sobre la seguridad estricta de
revocacion, el actual `FailOpen: true` es aceptable con monitoreo de Redis.
Documentar la decision como ADR.

---

### M2. Query SSE: no validation de year range

**Archivo:** `services/astro/internal/handler/astro.go:339-410`

El handler `Query` (SSE endpoint) parsea `year` del body pero solo hace
`if req.Year == 0 { req.Year = time.Now().Year() }` sin validar rango.
Esto contrasta con `parseRequest()` (linea 108) que valida `-5000 < year < 5000`.

Un atacante podria enviar `year: 999999999` al endpoint `/v1/astro/query`,
lo cual causaria que `astrocontext.Build()` ejecute calculos astronomicos
para un anio extremo, potencialmente causando:
- Newton iterations que no convergen (SolcrossUT, moonCrossUT)
- Loops muy largos en CalcTransits (sampling every 5 days for extreme ranges)
- CPU exhaustion

**Fix:** Aplicar la misma validacion de rango que `parseRequest()`:

```go
if req.Year < -5000 || req.Year > 5000 {
    jsonError(w, "year out of range", http.StatusBadRequest)
    return
}
```

---

### M3. CreateContact: sin validacion de campos del body

**Archivo:** `services/astro/internal/handler/astro.go:434-461`

`CreateContact` decodifica directamente a `repository.CreateContactParams` sin
validar:
- `Name` puede estar vacio (string zero value)
- `Lat` / `Lon` sin rango valido (-90..90 / -180..180)
- `BirthDate` sin rango razonable
- `Kind` sin whitelist de valores validos

Aunque la DB probablemente tiene constraints, confiar solo en la DB para
validacion expone mensajes de error de pgx al cliente (ver M4).

**Fix:** Agregar validacion antes del INSERT:

```go
if req.Name == "" {
    jsonError(w, "name is required", http.StatusBadRequest)
    return
}
if req.Lat < -90 || req.Lat > 90 || req.Lon < -180 || req.Lon > 180 {
    jsonError(w, "invalid coordinates", http.StatusBadRequest)
    return
}
```

---

## BAJOS (nice to have)

### L1. Error message leaks contact_name in resolveContact response

**Archivo:** `services/astro/internal/handler/astro.go:67`

```go
return nil, http.StatusNotFound, fmt.Errorf("contact %q not found", contactName)
```

El `%q` incluye el valor que envio el usuario (contactName). Este error se
propaga a las respuestas JSON via `jsonError(w, err.Error(), code)` (lineas
135, 154, 173, etc). Esto refleja input del usuario en la respuesta, lo cual
es un XSS vector si alguna vez se renderiza como HTML, y expone informacion
sobre la existencia de contactos.

**Fix:** Usar mensaje generico: `"contact not found"` sin incluir el nombre.

---

### L2. SearchContacts ILIKE wildcard injection (funcional, no SQLi)

**Archivo:** `services/astro/internal/repository/queries.sql:11`

```sql
SELECT * FROM contacts WHERE tenant_id = $1 AND user_id = $2
AND name ILIKE '%' || @query::text || '%' ORDER BY name LIMIT $3 OFFSET $4;
```

Aunque sqlc lo parametriza correctamente (no hay SQL injection), el usuario
puede enviar wildcards ILIKE como `%`, `_`, `[` en el query parameter. Esto
permite pattern matching intencional pero no es un riesgo de seguridad real
porque la query ya esta filtrada por tenant_id + user_id. Solo verian sus
propios contactos.

**Fix (optional):** Escapear `%` y `_` en el input antes de pasarlo al query,
o documentar que wildcards son intencionales.

---

### L3. LLM prompt injection via user query

**Archivo:** `services/astro/internal/handler/astro.go:389`

```go
prompt := fmt.Sprintf("Eres un astrologo profesional. Analiza el siguiente
brief y responde la consulta del usuario.\n\n%s\n\nConsulta: %s",
ctx.Brief, req.Query)
```

El campo `req.Query` del usuario se inyecta directamente en el prompt del LLM
sin sanitizacion. Un usuario podria intentar prompt injection ("Ignora las
instrucciones anteriores y..."). Sin embargo:
- El LLM solo tiene acceso al brief (texto leible, no datos sensibles)
- No hay tool calling habilitado (solo `SimplePrompt`)
- El output va directo al usuario que hizo la request (no a otros usuarios)
- El scope de dano es limitado: el peor caso es que el LLM responde algo
  fuera de contexto astrologico

**Impacto:** Bajo. No hay escalation path ni cross-tenant exposure.

**Fix (recomendado para V2):** Usar separadores claros o formato XML para
delimitar el brief del query del usuario. Considerar `pkg/guardrails/`
cuando se implemente.

---

### L4. SSE event name injection theoretical

**Archivo:** `services/astro/internal/handler/astro.go:472-478`

```go
func sseEvent(w http.ResponseWriter, f http.Flusher, event string, data any) {
    fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
}
```

Los event names son hardcoded ("contact_recognized", "calc_context", etc) asi
que no hay injection real. El `data` pasa por `json.Marshal` que escapa
correctamente. No hay vector de ataque actual.

---

### L5. Rate limit es in-memory, no distribuido

**Archivo:** `pkg/middleware/ratelimit.go`

El rate limiter usa un mapa in-memory con `golang.org/x/time/rate`. Si el
servicio corre en multiples instancias (scale-out), cada instancia tiene su
propio bucket. Un atacante podria distribuir requests entre instancias para
multiplicar el rate limit.

**Impacto:** Bajo para v0.1.0 (single instance). Para produccion multi-instance,
migrar a Redis-backed sliding window (ya documentado en el comentario del codigo).

---

### L6. health endpoint sin auth devuelve version info

**Archivo:** `services/astro/cmd/main.go:84-87`

```go
r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"status":"ok","service":"astro"}`))
})
```

El health endpoint esta fuera del auth middleware (correcto para health checks).
Solo devuelve nombre del servicio, no version ni estado de DB. Riesgo minimo.

---

## Tenant isolation audit

**RESULTADO: CORRECTO.** No hay vectores de cross-tenant leak.

### SQL queries

Todas las queries en `queries.sql` incluyen `WHERE tenant_id = $1 AND user_id = $2`:
- `GetContact`: `WHERE tenant_id = $1 AND user_id = $2 AND id = $3`
- `GetContactByName`: `WHERE tenant_id = $1 AND user_id = $2 AND lower(name) = lower($3)`
- `ListContacts`: `WHERE tenant_id = $1 AND user_id = $2`
- `SearchContacts`: `WHERE tenant_id = $1 AND user_id = $2`
- `CreateContact`: `VALUES ($1, $2, ...)` -- tenant_id y user_id como primeros params
- `UpdateContact`: `WHERE tenant_id = $1 AND user_id = $2 AND id = $3`
- `DeleteContact`: `WHERE tenant_id = $1 AND user_id = $2 AND id = $3`

### Tenant ID source

`tenantAndUser()` (handler/astro.go:40-53) obtiene tenant info de:
1. `tenant.FromContext(r.Context())` -- inyectado por el auth middleware desde JWT claims
2. `sdamw.UserIDFromContext(r.Context())` -- inyectado por el auth middleware desde JWT claims

Nunca del request body ni de query params. Correcto.

El `CreateContact` handler (linea 450-451) sobreescribe `req.TenantID` y `req.UserID`
con los valores del JWT, previniendo spoofing via body:

```go
req.TenantID = tid  // del JWT, no del body
req.UserID = uid    // del JWT, no del body
```

### Cross-validation

El auth middleware (`pkg/middleware/auth.go:115-118`) cross-validates:
```go
if traefikSlug != "" && claims.Slug != traefikSlug {
    writeJSONError(w, http.StatusForbidden, "tenant mismatch")
    return
}
```

---

## JWT audit

**RESULTADO: CORRECTO.**

- **Algoritmo:** Ed25519 (EdDSA) -- asimetrico, mucho mejor que HS256.
  Auth Service firma con clave privada, todos los demas servicios solo verifican
  con clave publica. Un servicio comprometido NO puede forjar tokens.
- **Verificacion de algoritmo:** `pkg/jwt/jwt.go:125` verifica que sea
  `*gojwt.SigningMethodEd25519`, rechaza otros algoritmos (previene alg:none y alg:HS256 attacks).
- **Claims requeridos:** `UserID != "" && TenantID != "" && Slug != ""`
  (jwt.go:139). Correcto.
- **Expiry:** Access 15min, Refresh 7 dias (jwt.go:55-56).
- **JTI:** Auto-generado si vacio (`uuid.NewString()`, jwt.go:79).
- **Blacklist:** Redis-backed (`pkg/security/blacklist.go`), TTL = remaining token lifetime.
- **Header stripping:** Auth middleware strips `X-User-ID`, `X-User-Email`,
  `X-User-Role`, `X-Tenant-ID`, `X-Tenant-Slug` ANTES de procesar
  (auth.go:45-49). Previene header spoofing.
- **MFA-pending rejection:** Tokens con role "mfa_pending" son rechazados
  (auth.go:90-92).

---

## CGO safety audit (swephgo)

**RESULTADO: SIN ISSUES CRITICOS.**

### Buffer sizes

- `CalcPlanet`: `xx := make([]float64, 6)`, `serr := make([]byte, 256)` -- Swiss Ephemeris
  espera exactamente estos tamanos. Correcto.
- `CalcHouses`: `cusps := make([]float64, 13)`, `ascmc := make([]float64, 10)` -- correcto.
- `FixstarUT`: Buffer minimo 41 bytes (`fixstarBufSize`), con extension si el nombre
  es mas largo (sweph.go:205-208). Correcto.
- `RevJul`: Single-element slices `make([]int, 1)`, `make([]float64, 1)`. Correcto.
- `SolEclipseWhenGlob`, `LunEclipseWhen`: `tret := make([]float64, 10)`. Correcto.
- `RiseTrans`: `tret := make([]float64, 1)`. Correcto.

No hay buffers suballocados. No hay `unsafe.Pointer` manual. swephgo maneja
la interfaz CGO internamente.

### cstr helper

```go
func cstr(b []byte) string {
    if i := bytes.IndexByte(b, 0); i >= 0 {
        return string(b[:i])
    }
    return string(b)
}
```

Correcto: busca el null terminator y trunca. Si no hay null, devuelve todo
el buffer (defensivo).

---

## Thread safety audit

**RESULTADO: CORRECTO.**

### CalcMu mutex pattern

- `CalcMu` es un `sync.Mutex` global (ephemeris/sweph.go:64)
- `BuildNatal()` lock/unlock correcto (natal/chart.go:44-45):
  `ephemeris.CalcMu.Lock(); defer ephemeris.CalcMu.Unlock()`
- `CalcSolarReturn()` lock/unlock correcto (solar_return.go:38-39)
- `CalcProgressions()` lock/unlock correcto (progressions.go:60-61)
- `CalcPlanetFull()` locks solo si topocentric flag is set (sweph.go:139-140)

### Potential issue: non-atomic CalcTransits

`CalcTransits()` (transits.go:73) llama a `CalcPlanetFull()` en un loop sin
lock externo. Cada llamada individual es thread-safe (locks internally if
needed), pero no hay flag topocentric en transits (solo `FlagSwieph | FlagSpeed`),
asi que no necesita CalcMu. Correcto.

### No deadlock risk

Todos los lock paths usan `defer Unlock()` y no hay nested locks. Un goroutine
nunca intenta adquirir CalcMu mientras ya lo tiene. Correcto.

---

## SQL injection audit

**RESULTADO: SIN VULNERABILIDADES.**

- Todas las queries son generadas por sqlc v1.30.0 con parametros `$N`.
- No hay `fmt.Sprintf` con SQL en ningun archivo del servicio.
- No hay raw SQL construction.
- El ILIKE pattern en SearchContacts usa concatenacion server-side SQL
  (`'%' || $5::text || '%'`), pero el valor viene como parametro bind.
  No hay injection posible.

---

## Input validation audit

| Endpoint | MaxBytesReader | Year validation | Contact validation |
|----------|:-:|:-:|:-:|
| POST /natal | 1MB | -5000..5000 | via resolveContact (tenant+user) |
| POST /transits | 1MB | -5000..5000 | via resolveContact |
| POST /solar-arc | 1MB | -5000..5000 | via resolveContact |
| POST /directions | 1MB | -5000..5000 | via resolveContact |
| POST /progressions | 1MB | -5000..5000 | via resolveContact |
| POST /returns | 1MB | -5000..5000 | via resolveContact |
| POST /profections | 1MB | -5000..5000 | via resolveContact |
| POST /firdaria | 1MB | -5000..5000 | via resolveContact |
| POST /fixed-stars | 1MB | -5000..5000 | via resolveContact |
| POST /brief | 1MB | -5000..5000 | via resolveContact |
| POST /query (SSE) | 1MB | **NO** (M2) | via resolveContact |
| GET /contacts | N/A | N/A | tenant+user from JWT |
| POST /contacts | 1MB | N/A | **PARTIAL** (M3) |

---

## Error leakage audit

**RESULTADO: BUENO CON EXCEPCION MENOR.**

Todos los errores internos usan mensajes genericos:
- "chart calculation failed" (no expone errores de swephgo)
- "chart calculation failed" (no expone errores de ephemeris)
- "context build failed" (no expone detalles)
- "narration unavailable" (no expone errores del LLM)
- "list failed" / "create failed" (no expone errores de pgx)

La unica excepcion es `resolveContact` que incluye el contactName en el error
(ver L1). No es un riesgo serio porque el contactName viene del propio request
del usuario, pero es mejor practica no reflejarlo.

El `slog.Error("llm call failed", "error", err)` (linea 392) logea el error
del LLM en el server log, pero NO lo envia al cliente. Correcto.

---

## Docker audit

**RESULTADO: BUENO.**

```dockerfile
FROM golang:1.25-alpine AS builder       # multi-stage build
RUN CGO_ENABLED=1 GOOS=linux go build    # CGO required for swephgo
FROM alpine:3.22                          # minimal runtime image
USER nobody:nobody                        # non-root
EXPOSE 8011
ENTRYPOINT ["/astro"]
```

- Multi-stage build: build dependencies no estan en la imagen final
- Non-root: `USER nobody:nobody`
- No secrets en la imagen
- Ephemeris data via volume mount (`/ephe:ro` -- read-only)
- No usa distroless (necesita alpine para libc por CGO), pero alpine:3.22 es
  aceptable

---

## Dependency audit

| Package | Version | CVEs conocidos |
|---------|---------|----------------|
| golang-jwt/jwt/v5 | v5.3.1 | Ninguno conocido |
| go-chi/chi/v5 | v5.2.5 | Ninguno conocido |
| jackc/pgx/v5 | v5.9.1 | Ninguno conocido |
| mshafiee/swephgo | v1.1.0 | Ninguno conocido (wrapper C) |
| redis/go-redis/v9 | v9.18.0 | Ninguno conocido |
| otelhttp | v0.67.0 | Ninguno conocido |
| golang.org/x/time | v0.15.0 | Ninguno conocido |
| google/uuid | v1.6.0 | Ninguno conocido |

**Nota sobre swephgo v1.1.0:** Este es un wrapper CGO sobre Swiss Ephemeris.
La ultima actividad del repo mshafiee/swephgo fue 2023. Swiss Ephemeris itself
es mantenida activamente. Monitorear el upstream para vulnerabilidades en la
libreria C subyacente.

---

## Security headers

El servicio usa `sdamw.SecureHeaders()` (main.go:82) que setea:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 0` (correct -- CSP es la proteccion adecuada)
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy: camera=(), microphone=(), geolocation=()`
- `Strict-Transport-Security` (solo con TLS)

Correcto.

---

## Rate limiting

- Aplicado a TODOS los endpoints protegidos (tanto JSON group como SSE group)
- Config: 10 requests/minute per user (ByUser key function)
- In-memory token bucket (aceptable para single instance)
- SSE endpoint incluido en rate limiting (main.go:122-124)

Correcto.

---

## Faltantes de seguridad (vs spec)

1. **Audit log:** No hay audit logging de acciones del usuario (crear/borrar
   contactos). El modelo `AuditLog` existe en la DB pero el servicio no escribe.
   Recomendado para v0.2.0.

2. **RBAC permissions:** No hay `RequirePermission()` middleware en los endpoints.
   Cualquier usuario autenticado puede usar todas las funciones del servicio.
   Para v0.1.0 esto es aceptable si el modulo astro es opt-in por tenant.
   Para v1.0.0 agregar permisos como `astro.read`, `astro.write`.

3. **CORS:** Manejado por Traefik middleware (`dev-cors`, `dev-tenant`).
   Verificar que produccion tenga origins restrictivos.

4. **Query length validation:** El campo `query` en `/v1/astro/query` no tiene
   limite de longitud mas alla del MaxBytesReader de 1MB. Considerar un limite
   de ~2000 caracteres para el prompt del usuario.

---

## Veredicto: APTO para produccion

El servicio tiene una postura de seguridad solida para v0.1.0:
- JWT Ed25519 asimetrico con verificacion de algoritmo
- Tenant isolation perfecta (todas las queries scoped)
- Sin SQL injection (sqlc parametrizado)
- Sin error leakage significativa
- Rate limiting en todos los endpoints
- Docker non-root con imagen minimal
- MaxBytesReader en todos los POST bodies
- Header stripping contra spoofing

Las correcciones M2 (year validation en Query) y M3 (input validation en
CreateContact) son recomendadas antes del primer release publico pero no
bloquean un deploy interno/beta.

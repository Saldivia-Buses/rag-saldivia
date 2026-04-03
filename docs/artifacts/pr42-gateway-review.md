# Gateway Review -- PR #42 Complete Test Coverage

**Fecha:** 2026-04-02
**Tipo:** review
**Branch:** `feat/complete-test-coverage`
**Reviewer:** gateway-reviewer (Opus)
**Intensity:** thorough

## Resultado

**APROBADO** (con sugerencias)

No hay bloqueantes ni issues de seguridad. Los tests son correctos, cubren los
paths criticos, y el refactor de la interfaz en auth compila sin romper main.go.

---

## Hallazgos

### Bloqueantes

Ninguno.

### Debe corregirse

Ninguno.

---

### Sugerencias

#### S1. Auth mock no valida campos de LoginRequest

`services/auth/internal/handler/auth_test.go:22-27`

El mock ignora completamente `req.LoginRequest` -- no valida que `Email`,
`Password`, `IP` y `UserAgent` se propaguen correctamente desde el handler al
service. Si el handler dejara de pasar `r.RemoteAddr` o `r.UserAgent()`, los
tests seguirian pasando.

**Fix sugerido:** agregar un campo `capturedReq` al mock y validar en
`TestLogin_Success` que los campos fueron propagados:

```go
type mockAuthService struct {
    tokens      *service.TokenPair
    err         error
    capturedReq *service.LoginRequest
}

func (m *mockAuthService) Login(_ context.Context, req service.LoginRequest) (*service.TokenPair, error) {
    m.capturedReq = &req
    if m.err != nil {
        return nil, m.err
    }
    return m.tokens, nil
}
```

Luego en `TestLogin_Success`:
```go
if mock.capturedReq.Email != "admin@test.com" {
    t.Errorf("expected email propagation, got %q", mock.capturedReq.Email)
}
if mock.capturedReq.IP == "" {
    t.Error("expected IP to be propagated from RemoteAddr")
}
```

#### S2. ws_test.go no tiene test del Upgrade path (JWT rejection)

`services/ws/internal/handler/ws_test.go`

Los tests cubren las funciones helper (`parseOrigins`, `extractBearerToken`)
pero no cubren el handler `Upgrade` en absoluto. Los dos paths mas importantes
para seguridad -- "missing token returns 401" y "invalid token returns 401" --
no estan testeados. Entiendo que testear WebSocket upgrade es mas complejo (se
necesita `httptest.Server` + un client real o al menos un mock de la hub), pero
al menos el JWT rejection path se puede testear con httptest solo:

```go
func TestUpgrade_MissingToken_Returns401(t *testing.T) {
    h := &WS{jwtSecret: "test-secret-that-is-32-bytes-ok!"}
    req := httptest.NewRequest(http.MethodGet, "/ws", nil)
    rec := httptest.NewRecorder()
    h.Upgrade(rec, req)
    if rec.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401, got %d", rec.Code)
    }
}
```

Esto no requiere un hub real ni websocket client, y cubre el path critico de auth.

#### S3. Platform integration tests no cubren UpdateTenant ni DisableTenant/EnableTenant

`services/platform/internal/service/platform_integration_test.go`

El service tiene 6 operaciones de tenant: Create, List, Get, Update, Disable,
Enable. Los tests cubren 4 de 6 (Create, List, Get, y implicitamente el error
paths). `UpdateTenant` y `DisableTenant`/`EnableTenant` no estan testeados.
Dado que el comment del PR es "completes all remaining backend test coverage",
estos gaps son notables.

#### S4. Platform integration test schema drift risk

`services/platform/internal/service/platform_integration_test.go:51-120`

La migracion inline en el test es un subconjunto simplificado de
`services/platform/db/migrations/001_init.up.sql`. El schema real tiene tablas
adicionales (`rag_models`, `deploy_log`, `connectors`) y mas seed data. Hoy
esto es irrelevante porque los tests no tocan esas tablas, pero si un futuro
metodo las usa, el test fallaria de forma confusa.

**Opcion a considerar:** reuse del archivo de migracion real en vez de inline SQL:
```go
migration, _ := os.ReadFile("../../db/migrations/001_init.up.sql")
pool.Exec(ctx, string(migration))
```

Esto elimina el riesgo de drift permanentemente. Mismo aplica para ingest.

#### S5. Ingest integration tests no cubren el Submit path

`services/ingest/internal/service/ingest_integration_test.go`

Submit es el metodo mas critico del servicio -- stages file, crea job en DB,
publica a NATS. El test file lo documenta explicitamente ("Blueprint proxy is
not tested here") pero hay una middle ground: se podria testear que Submit crea
el job en DB y stages el file, usando un NATS testcontainer o un mock. Si NATS
es el blocker, al menos un test que valide la creacion del job y cleanup en
caso de error de NATS seria valioso.

#### S6. Consider table-driven tests for auth error mapping

`services/auth/internal/handler/auth_test.go:106-164`

Los tests `TestLogin_InvalidCredentials_Returns401`,
`TestLogin_AccountLocked_Returns429`, y `TestLogin_InternalError_Returns500_Generic`
siguen el mismo patron exacto pero como tests separados. La convencion Go de
este proyecto (bible.md: "Table-driven, archivo `_test.go` junto al codigo")
sugiere consolidar:

```go
func TestLogin_ErrorMapping(t *testing.T) {
    tests := []struct {
        name       string
        err        error
        wantStatus int
        wantMsg    string
    }{
        {"invalid credentials", service.ErrInvalidCredentials, 401, "invalid email or password"},
        {"account locked", service.ErrAccountLocked, 429, "too many attempts, try again later"},
        {"internal error", errors.New("boom"), 500, "internal error"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

Esto no cambia coverage pero alinea con la convencion del proyecto.

---

### Lo que esta bien

1. **Interface extraction en auth handler es limpia.** El `AuthService` interface
   en `handler/auth.go:17-19` tiene exactamente un metodo, es el minimum viable
   seam para testing, y `cmd/main.go` compila sin cambios porque `*service.Auth`
   satisface la interfaz implicitamente. Este es el patron correcto para Go.

2. **Error mapping no filtra internals.** `TestLogin_InternalError_Returns500_Generic`
   (auth_test.go:144) valida explicitamente que "database exploded" NO aparece
   en la response y solo se devuelve "internal error". Esto es exactamente lo
   que el security checklist requiere.

3. **Ingest ownership model esta correctamente testeado.**
   `TestGetJob_Ownership_Integration` y `TestDeleteJob_NonOwner_Integration`
   validan que un user no puede leer ni borrar jobs de otro user, y que ambos
   retornan `ErrJobNotFound` (no "access denied"), previniendo enumeracion.

4. **Platform duplicate slug detection.** `TestCreateTenant_DuplicateSlug_Integration`
   verifica que el servicio detecta `23505` (unique violation) y devuelve
   `ErrSlugTaken` en vez de un error generico de DB. Bien mapeado.

5. **Platform invalid slug validation exhaustiva.** El test prueba uppercase,
   espacios, dash al inicio, slug de un caracter, y path traversal (`../etc`).
   Buen rango de inputs.

6. **Testcontainers-go en integration tests.** Usar PostgreSQL real en vez de
   mocks para integration tests es la decision correcta, especialmente para
   validar constraints, indexes y ON CONFLICT behavior. Alineado con
   bible.md: "testify + testcontainers".

7. **ws_test.go table-driven tests.** `parseOrigins` y `extractBearerToken`
   usan table-driven approach con buenos edge cases (whitespace, empty commas,
   "Basic" auth prefix, lowercase "bearer").

8. **MaxBytesReader en login handler.** `auth.go:43` limita body a 1MB. No es
   nuevo en este PR pero es bueno que el refactor no lo haya tocado.

---

## Compile Correctness

| Check | Status |
|---|---|
| `handler.AuthService` interface matches `*service.Auth` | OK -- Login(ctx, LoginRequest) (*TokenPair, error) matches |
| `cmd/main.go` compiles (NewAuth accepts AuthService) | OK -- *service.Auth satisfies interface implicitly |
| `service.TokenPair` struct used in test has correct fields | OK -- AccessToken, RefreshToken, ExpiresIn all present |
| `db.CreateTenantParams` fields match test usage | OK -- Slug, Name, PlanID, PostgresUrl, RedisUrl, Settings |
| `db.EnableModuleForTenantParams` fields match test usage | OK -- TenantID, ModuleID, Config, EnabledBy |
| `Ingest.New(pool, nil, nil, Config{})` compiles | OK -- nil *nats.Conn and nil EventPublisher are valid |
| Build tags (`//go:build integration`) correct | OK -- both integration test files use correct syntax |

## Coverage Summary

| Service | Test File | Tests | Paths Covered | Gaps |
|---|---|---|---|---|
| auth handler | auth_test.go | 8 | success, validation (3), error mapping (3), health | Request field propagation not validated |
| ws handler | ws_test.go | 7 (2 funcs) | parseOrigins (5 cases), extractBearerToken (5 cases) | Upgrade handler not tested at all |
| platform service | platform_integration_test.go | 11 | CRUD tenant, modules, config, error paths | UpdateTenant, Disable/EnableTenant missing |
| ingest service | ingest_integration_test.go | 7 | List, Get, Delete, Update status, ownership, limit | Submit not tested |

## Security Checklist

| Check | Status |
|---|---|
| Error messages don't leak internals | PASS -- auth 500 returns "internal error", ingest returns ErrJobNotFound for non-owner |
| Ownership verified in ingest | PASS -- GetJob and DeleteJob filter by user_id |
| Account lockout tested | PASS -- 429 for ErrAccountLocked |
| Invalid slug with path traversal tested | PASS -- "../etc" rejected |
| Bearer token extraction is case-sensitive | PASS -- "bearer" (lowercase) returns empty string |

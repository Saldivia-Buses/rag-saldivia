---
title: Convention: Go Code Style
audience: ai
last_reviewed: 2026-04-15
related:
  - ./error-handling.md
  - ./logging.md
  - ./testing.md
  - ./sqlc.md
  - ../architecture/overview.md
---

Rules for writing Go in `services/` and `pkg/`. Enforced where possible by `make lint` (golangci-lint) and `bash .claude/hooks/check-invariants.sh`.

## Naming

| Element | Rule | Example |
|---|---|---|
| Packages | lowercase, single word, no underscores | `handler`, `service`, `repository` |
| Files | snake_case `.go` | `auth_handler.go`, `service_token_test.go` |
| Structs | PascalCase | `UserService`, `TokenPair` |
| Interfaces | PascalCase, `-er` suffix when describing behaviour | `EventPublisher`, `AuthService` |
| Exported funcs | PascalCase | `CreateUser` |
| Internal funcs | camelCase | `hashPassword` |
| Errors (sentinel) | `Err` prefix, package-level `var` | `ErrInvalidCredentials` |
| Constants | PascalCase (exported), camelCase (internal); group with `const ( ... )` | `DefaultTimeout` |

DO mirror existing examples — `services/auth/internal/handler/auth.go:22` is the reference for handler interfaces.

DON'T use `Get` prefix on simple accessors (`User()`, not `GetUser()`). DO use `Get*` only for sqlc-generated names.

## Function signatures

DO put `ctx context.Context` as the first parameter of any function that performs I/O, calls a repository, publishes NATS, or makes outbound HTTP calls.
- See `pkg/jwt/jwt.go:73` and every method on `AuthService` (`services/auth/internal/handler/auth.go:23-33`).

DON'T pass `*http.Request` deeper than the handler layer. Extract what the service needs (user ID, tenant slug, IP, user agent) and pass values.

DO accept interfaces, return concrete types. The handler defines the interface it depends on (`AuthService` lives next to the handler, not next to the implementation).

## Error wrapping

DO wrap every returned error with context using `%w`:
- `return nil, fmt.Errorf("get user by email: %w", err)`

See `pkg/jwt/jwt.go:75` for sentinel + wrap pattern.

DON'T return raw infrastructure errors to handlers — wrap into a `httperr.Error` (see [error-handling](./error-handling.md)).

DON'T panic in request paths. Panics are reserved for `init()` invariant failures (missing config, invalid key material).

## Service structure

Every Go service under `services/{name}/` has this shape:

```
cmd/main.go              entry point
internal/
  handler/               HTTP/gRPC handlers (chi router)
  service/               business logic + interfaces
  repository/            sqlc-generated DB access
db/
  queries/               sqlc .sql sources
  sqlc.yaml
VERSION                  semver, one line
Dockerfile               multi-stage, non-root
README.md                purpose, endpoints, NATS events
```

Migrations live under `db/{platform|tenant}/migrations/` at the repo root, not per-service. Service registration in `go.work` is mandatory — `make build` ignores unregistered services.

## Imports

DO group imports in three blocks separated by blank lines: stdlib, third-party, internal (`github.com/Camionerou/rag-saldivia/...`).

DON'T import `services/X/internal/...` from another service. Cross-service contracts go through `pkg/`, NATS events, or HTTP/gRPC.

## Concurrency

DO use `sync.Map` for caches keyed by tenant slug (`services/auth/internal/handler/auth.go:48`).

DO pass `ctx` to every goroutine and respect cancellation. Long-lived goroutines started at boot must shut down on `ctx.Done()` — see `pkg/server/`.

DON'T spawn unbounded goroutines from handlers. Use a worker pool or a buffered channel.

## Tests

Tests live next to the code as `*_test.go`. Patterns and tooling are in [testing](./testing.md).

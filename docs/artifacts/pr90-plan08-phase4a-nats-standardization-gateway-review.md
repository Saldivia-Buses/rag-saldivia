# Gateway Review — PR #90 Plan 08 Phase 4a NATS Standardization

**Fecha:** 2026-04-05
**Resultado:** APROBADO CON OBSERVACION MENOR

## Bloqueantes

Ninguno.

## Debe corregirse

### 1. `ws/cmd/main.go` y `notification/cmd/main.go` retienen `nats.DefaultURL` via import directo de `nats.go`

Ambos servicios hacen:

```go
import "github.com/nats-io/nats.go"
natsURL := config.Env("NATS_URL", nats.DefaultURL)
```

El PR declara que los imports de `nats.go` fueron eliminados de agent y platform, pero no de ws ni notification (ni de auth, chat, ingest, traces, feedback). Esto es correcto comportamiento — esos servicios sí necesitan el import para `nats.DefaultURL` o `nats.StreamConfig`. No es un bug, pero el commit message dice "unused nats.go imports removed from agent and platform" y eso es todo lo que se removió. Verificado: agent y platform ya no importan `nats.go` directamente, correcto.

No hay acción requerida aquí — solo documentando que el scope es el declarado.

## Sugerencias

### 1. `platform/cmd/main.go` linea 68 — `natspub.New(nc)` cuando `nc` es nil

En platform, NATS falla con `slog.Warn` (no `os.Exit`) y continúa. La siguiente línea es:

```go
publisher := natspub.New(nc)  // nc puede ser nil aquí
```

`natspub.New` solo asigna el campo `nc *nats.Conn`, no falla. Pero cuando `publisher.Notify()` o `publisher.Broadcast()` se llamen internamente y `nc` sea nil, habrá un nil pointer dereference en `p.nc.Publish(...)`. Este patrón ya existía antes de este PR — el PR no lo empeoró. Aun así, si se quiere hardening futuro: agregar nil check en `Publisher.Notify` y `Publisher.Broadcast` o hacer `natspub.New` aceptar nil y retornar un no-op publisher.

### 2. `agent/cmd/main.go` linea 55 — `defer nc.Drain()` dentro del bloque `else`

```go
if err != nil {
    slog.Warn("nats connect failed, trace publishing disabled", ...)
} else {
    defer nc.Drain()
}
```

Este patron es correcto — Drain solo se llama si la conexion tuvo exito. Sin observacion adicional.

## Lo que está bien

- **Migración completa a `natspub.Connect()`**: los 9 servicios (auth, ws, chat, agent, search, ingest, feedback, traces, notification, platform) usan `natspub.Connect()`. Confirmado línea por línea.

- **6 servicios migrados de `nc.Close()` a `nc.Drain()`**: auth, ws, chat, ingest, feedback, notification, traces — todos usan `defer nc.Drain()`. Agent usa `defer nc.Drain()` condicionalmente dentro del `else`. Platform usa `defer nc.Drain()` condicionalmente. Correcto.

- **Context en consumers**: `notification/internal/service/consumer.go` linea 149 usa `ctx := c.ctx` (struct field) y `feedback/internal/service/consumer.go` linea 168 hace lo mismo. `feedback/internal/service/aggregator.go` linea 71 usa `ctx := a.ctx`. Los tres usaban `context.Background()` antes — ahora propagan el contexto de cancelación del proceso correctamente, lo que permite shutdown limpio.

- **`agent.Query()` firma actualizada**: `service.Agent.Query()` acepta `(ctx, jwt, userID, userMessage, history)`. El handler en `handler/agent.go` linea 58-59 lee `userID := r.Header.Get("X-User-ID")` — que el middleware ya verificó e inyectó — y lo pasa correctamente. No hay otros callers de `Query()` en el repo (grep confirmado: único caller es el handler).

- **Imports `nats.go` removidos de agent y platform**: agent/cmd/main.go no importa `nats.go` directamente. platform/cmd/main.go tampoco. Confirmado.

- **`natspub.Connect()` incluye las opciones críticas**: `RetryOnFailedConnect`, `MaxReconnects(-1)`, `ReconnectWait(2s)`, disconnect/reconnect handlers con slog. Ningún servicio perdió opciones al migrar desde `nats.Connect()` directo.

- **Search no usa NATS**: search/cmd/main.go correctamente no tiene conexión NATS (solo lee DB + llama SGLang). No faltó migración aquí.

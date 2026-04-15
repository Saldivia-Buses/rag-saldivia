---
name: Plan 24 — Test Reinforcement
description: Test reinforcement plan written 2026-04-14, 10 phases, 10 services, unit + integration
type: project
---

Plan 24 escrito 2026-04-14. Branch: `feat/2.0.5-test-reinforcement`.

Servicios: auth (ext), chat, platform, agent, search, ingest, ws, notification, traces, feedback.

Estructura: un commit por servicio, un PR grande al final (~11 commits). No squash.

Puntos clave:
- Patron de referencia: `services/auth/internal/handler/auth_test.go` + `service/auth_integration_test.go`
- Integration tests con `//go:build integration` + testcontainers postgres
- NATS embebido (`nats-server/v2/server` InProcessServer) para consumers
- Coverage es proxy, no meta: el foco real es paths criticos de tenant isolation y ownership
- `make test-integration` target nuevo en Makefile (fase de infra)

**Why:** coverage muy desparejo — servicios con 1 test para 10 archivos, paths criticos sin verificacion en CI.

**How to apply:** al ejecutar este plan, leer cada handler y service ANTES de escribir el test para conocer la interfaz exacta y los errores exportados.

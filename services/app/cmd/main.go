// Command app is the consolidated monolith per ADR 025.
//
// The target shape is 5 domain modules in a single Go binary:
//
//	internal/core/     auth + platform + feedback
//	internal/rag/      ingest + search + agent
//	internal/realtime/ chat + ws + notification
//	internal/ops/      bigbrother + healthwatch + traces
//	internal/erp/      erp (isolated, last to fold)
//
// Fusions land one domain at a time. This binary grows accordingly; any
// module not yet folded still runs as a standalone service under its
// existing services/<svc>/cmd/main.go until its fusion session.
package main

import (
	"github.com/Camionerou/rag-saldivia/pkg/health"
	"github.com/Camionerou/rag-saldivia/pkg/server"
)

func main() {
	app := server.New("sda-app", server.WithPort("APP_PORT", "8020"))

	hc := health.New("app")
	r := app.Router()
	r.Get("/health", hc.Handler())

	app.Run()
}

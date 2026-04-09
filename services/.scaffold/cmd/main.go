package main

import (
	"github.com/Camionerou/rag-saldivia/pkg/server"
)

func main() {
	app := server.New("sda-scaffold", server.WithPort("SCAFFOLD_PORT", "8099"))
	_ = app.Context()

	// TODO: initialize dependencies (DB, NATS, Redis)
	// TODO: create handlers, register routes on app.Router()
	// TODO: add health checks: r.Get("/health", hc.Handler())

	app.Run()
}

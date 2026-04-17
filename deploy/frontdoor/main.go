// Command frontdoor is the per-tenant container's :80 entrypoint.
//
// In the all-in-one image (ADR 023 + 024) this owns public HTTP: it serves
// /readyz and /healthz in-process and will later reverse-proxy /v1/* + /ws/*
// to the Go services and the rest to the Next.js standalone server on :3000.
//
// This is a stub scoped to commit 1 — only the liveness surface. Reverse-proxy
// wiring lands in a later commit once the Go services and Next.js are baked
// into the image.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	addr := os.Getenv("FRONTDOOR_ADDR")
	if addr == "" {
		addr = ":80"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /readyz", ready)
	mux.HandleFunc("GET /healthz", ready)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("frontdoor listening", "addr", addr, "tenant", os.Getenv("SDA_TENANT"))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen failed", "err", err.Error())
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("frontdoor shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Warn("shutdown error", "err", err.Error())
	}
}

func ready(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "ok",
		"tenant":   os.Getenv("SDA_TENANT"),
		"hostname": hostname(),
	})
}

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

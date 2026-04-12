package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/Camionerou/rag-saldivia/pkg/database"
	"github.com/Camionerou/rag-saldivia/pkg/health"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	natspub "github.com/Camionerou/rag-saldivia/pkg/nats"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/server"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/handler"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

func main() {
	app := server.New("sda-erp", server.WithPort("ERP_PORT", "8013"))
	ctx := app.Context()

	tenantDBURL := config.Env("POSTGRES_TENANT_URL", "")
	if tenantDBURL == "" {
		slog.Error("POSTGRES_TENANT_URL is required")
		os.Exit(1)
	}

	publicKey := sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")
	blacklist := security.InitBlacklist(ctx, config.Env("REDIS_URL", "localhost:6379"))

	pool, err := database.NewPool(ctx, tenantDBURL)
	if err != nil {
		slog.Error("failed to connect to tenant database", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(pool.Close)

	nc, err := natspub.Connect(config.Env("NATS_URL", nats.DefaultURL))
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	app.OnShutdown(func() { nc.Drain() })
	slog.Info("connected to NATS", "url", config.RedactURL(config.Env("NATS_URL", "")))

	// Core dependencies
	auditWriter := audit.NewWriter(pool)
	publisher := traces.NewPublisher(nc)

	// Repository + Service + Handler
	repo := repository.New(pool)
	suggestionsSvc := service.NewSuggestions(repo, auditWriter, publisher)
	suggestionsHandler := handler.NewSuggestions(suggestionsSvc)
	catalogsSvc := service.NewCatalogs(repo, auditWriter, publisher)
	catalogsHandler := handler.NewCatalogs(catalogsSvc)
	entitiesSvc := service.NewEntities(repo, auditWriter, publisher)
	entitiesHandler := handler.NewEntities(entitiesSvc)
	stockSvc := service.NewStock(repo, pool, auditWriter, publisher)
	stockHandler := handler.NewStock(stockSvc)
	accountingSvc := service.NewAccounting(repo, pool, auditWriter, publisher)
	accountingHandler := handler.NewAccounting(accountingSvc)
	treasurySvc := service.NewTreasury(repo, pool, auditWriter, publisher)
	treasuryHandler := handler.NewTreasury(treasurySvc)
	purchasingSvc := service.NewPurchasing(repo, pool, auditWriter, publisher)
	purchasingHandler := handler.NewPurchasing(purchasingSvc)
	salesSvc := service.NewSales(repo, pool, auditWriter, publisher)
	salesHandler := handler.NewSales(salesSvc)
	invoicingSvc := service.NewInvoicing(repo, pool, auditWriter, publisher)
	invoicingHandler := handler.NewInvoicing(invoicingSvc)
	currentAccountsSvc := service.NewCurrentAccounts(repo, pool, auditWriter, publisher)
	currentAccountsHandler := handler.NewCurrentAccounts(currentAccountsSvc)
	productionSvc := service.NewProduction(repo, pool, auditWriter, publisher)
	productionHandler := handler.NewProduction(productionSvc)
	hrSvc := service.NewHR(repo, auditWriter, publisher)
	hrHandler := handler.NewHR(hrSvc)
	qualitySvc := service.NewQuality(repo, auditWriter, publisher)
	qualityHandler := handler.NewQuality(qualitySvc)
	maintenanceSvc := service.NewMaintenance(repo, auditWriter, publisher)
	maintenanceHandler := handler.NewMaintenance(maintenanceSvc)
	adminSvc := service.NewAdmin(repo, auditWriter, publisher)
	adminHandler := handler.NewAdmin(adminSvc)
	analyticsSvc := service.NewAnalytics(repo)
	analyticsHandler := handler.NewAnalytics(analyticsSvc)

	// Health
	hc := health.New("erp")
	hc.Add("postgres", func(ctx context.Context) error { return pool.Ping(ctx) })
	hc.Add("nats", func(ctx context.Context) error {
		if !nc.IsConnected() {
			return fmt.Errorf("nats disconnected")
		}
		return nil
	})
	if blacklist != nil {
		hc.Add("redis", func(ctx context.Context) error { return blacklist.Ping(ctx) })
	}

	// Routes
	r := app.Router()
	r.Get("/health", hc.Handler())

	// FailOpen=false for ERP — all data is sensitive (financial, PII)
	authRead := sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: false})
	authWrite := sdamw.AuthWithConfig(publicKey, sdamw.AuthConfig{Blacklist: blacklist, FailOpen: false})

	r.Group(func(r chi.Router) {
		r.Use(authRead) // default for mount, write endpoints override below
		r.Mount("/v1/erp/suggestions", suggestionsHandler.Routes(authWrite))
		r.Mount("/v1/erp/catalogs", catalogsHandler.Routes(authWrite))
		r.Mount("/v1/erp/entities", entitiesHandler.Routes(authWrite))
		r.Mount("/v1/erp/stock", stockHandler.Routes(authWrite))
		r.Mount("/v1/erp/accounting", accountingHandler.Routes(authWrite))
		r.Mount("/v1/erp/treasury", treasuryHandler.Routes(authWrite))
		r.Mount("/v1/erp/purchasing", purchasingHandler.Routes(authWrite))
		r.Mount("/v1/erp/sales", salesHandler.Routes(authWrite))
		r.Mount("/v1/erp/invoicing", invoicingHandler.Routes(authWrite))
		r.Mount("/v1/erp/accounts", currentAccountsHandler.Routes(authWrite))
		r.Mount("/v1/erp/production", productionHandler.Routes(authWrite))
		r.Mount("/v1/erp/hr", hrHandler.Routes(authWrite))
		r.Mount("/v1/erp/quality", qualityHandler.Routes(authWrite))
		r.Mount("/v1/erp/maintenance", maintenanceHandler.Routes(authWrite))
		r.Mount("/v1/erp/admin", adminHandler.Routes(authWrite))
		r.Mount("/v1/erp/analytics", analyticsHandler.Routes(authWrite))
	})

	app.Run()
}

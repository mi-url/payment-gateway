// Package main is the entry point for the payment gateway HTTP server.
// It initializes all dependencies in order, wires them together, and starts
// the server. Dependencies flow downwards: config → DB → stores → crypto →
// bank adapters → services → handlers → middleware → server.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/faloppa/payment-gateway/internal/bank"
	"github.com/faloppa/payment-gateway/internal/bank/bnc"
	"github.com/faloppa/payment-gateway/internal/config"
	"github.com/faloppa/payment-gateway/internal/crypto"
	"github.com/faloppa/payment-gateway/internal/handler"
	"github.com/faloppa/payment-gateway/internal/middleware"
	"github.com/faloppa/payment-gateway/internal/service"
	"github.com/faloppa/payment-gateway/internal/store"
)

func main() {
	// Structured JSON logger for Cloud Run log aggregation.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load and validate configuration at boot. Missing env vars cause immediate exit.
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Connect to PostgreSQL (Supabase).
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to open database connection", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		logger.Error("failed to ping database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("database connected")

	// Initialize stores (data access layer).
	merchantStore := store.NewMerchantStore(db)
	txnStore := store.NewTransactionStore(db)
	cfgStore := store.NewBankConfigStore(db)

	// Initialize KMS client.
	// In development mode, use MockKMS. In production, use GCP Cloud KMS.
	var kmsClient crypto.KMSClient
	if cfg.IsProduction() {
		logger.Warn("production KMS not yet implemented, using MockKMS — DO NOT USE WITH REAL CREDENTIALS")
		kmsClient = crypto.NewMockKMS()
	} else {
		kmsClient = crypto.NewMockKMS()
	}
	encryptor := crypto.NewEnvelopeEncryptor(kmsClient, cfg.KMSKeyResourceName)

	// Initialize bank adapter registry.
	registry := bank.NewRegistry()
	bncAdapter := bnc.NewAdapter(cfg.BNCBaseURL, logger, !cfg.IsProduction())
	registry.Register(bncAdapter)
	logger.Info("bank adapters registered", slog.Any("banks", registry.RegisteredBanks()))

	// Initialize services (business logic layer).
	chargeService := service.NewChargeService(txnStore, cfgStore, registry, encryptor, logger)

	// Initialize reconciliation worker.
	reconService := service.NewReconciliationService(txnStore, cfgStore, registry, encryptor, logger)
	reconCtx, reconCancel := context.WithCancel(context.Background())
	defer reconCancel()
	go reconService.Start(reconCtx)

	// Initialize handlers (HTTP layer).
	chargeHandler := handler.NewChargeHandler(chargeService, logger)
	txnHandler := handler.NewTransactionHandler(txnStore, logger)
	webhookBNCHandler := handler.NewWebhookBNCHandler(logger)
	bankConfigHandler := handler.NewBankConfigHandler(cfgStore, encryptor, logger)

	// Merchant lookup function for auth middleware.
	authLookup := func(ctx context.Context, apiKeyHash string) (string, bool) {
		m, found, err := merchantStore.FindByAPIKeyHash(ctx, apiKeyHash)
		if err != nil || !found {
			return "", false
		}
		return m.ID.String(), true
	}

	// Build router with middleware stack.
	mux := http.NewServeMux()

	// Public endpoints (no auth required).
	mux.HandleFunc("GET /health", handler.Health)
	mux.Handle("POST /v1/webhooks/bnc", webhookBNCHandler)

	// Authenticated endpoints.
	authed := http.NewServeMux()
	authed.Handle("POST /v1/charges/c2p", chargeHandler)
	authed.Handle("GET /v1/transactions/{id}", txnHandler)
	authed.Handle("POST /v1/config/bank", bankConfigHandler)

	// Apply auth + idempotency middleware to authenticated routes.
	idempotencyStore := middleware.NewMemoryIdempotencyStore()
	var authedHandler http.Handler = authed
	authedHandler = middleware.Idempotency(idempotencyStore)(authedHandler)
	authedHandler = middleware.Auth(authLookup)(authedHandler)
	mux.Handle("/v1/", authedHandler)

	// Apply global middleware.
	var root http.Handler = mux
	root = middleware.RateLimit(100, time.Minute)(root)
	root = middleware.Logging(logger)(root)

	// Start server with graceful shutdown.
	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      root,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Listen for shutdown signals.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("server starting",
			slog.String("addr", addr),
			slog.String("env", cfg.Env),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal, then gracefully drain connections.
	sig := <-shutdown
	logger.Info("shutdown signal received", slog.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reconCancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}

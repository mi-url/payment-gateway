// Package main is the entry point for the payment gateway HTTP server.
// It initializes configuration, sets up the router with middleware,
// registers bank adapters, and starts the server.
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/faloppa/payment-gateway/internal/bank"
	"github.com/faloppa/payment-gateway/internal/config"
	"github.com/faloppa/payment-gateway/internal/handler"
	"github.com/faloppa/payment-gateway/internal/middleware"
)

func main() {
	// Structured JSON logger for Cloud Run compatibility.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load and validate configuration at boot.
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize bank adapter registry.
	registry := bank.NewRegistry()
	logger.Info("bank adapters registered", slog.Any("banks", registry.RegisteredBanks()))

	// Initialize handlers.
	chargeHandler := handler.NewChargeHandler(logger)
	webhookBNCHandler := handler.NewWebhookBNCHandler(logger)

	// Build router with middleware and routes.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.Health)
	mux.Handle("POST /v1/charges/c2p", chargeHandler)
	mux.Handle("POST /v1/webhooks/bnc", webhookBNCHandler)

	// Apply global middleware.
	var root http.Handler = mux
	root = middleware.Logging(logger)(root)

	// Start server.
	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("server starting",
		slog.String("addr", addr),
		slog.String("env", cfg.Env),
	)

	if err := http.ListenAndServe(addr, root); err != nil {
		logger.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

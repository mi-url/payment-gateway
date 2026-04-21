// Package handler implements the HTTP handlers for the payment gateway API.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ChargeHandler handles POST /v1/charges/c2p requests.
// It validates the request, checks idempotency, routes to the appropriate
// bank adapter, and returns a standardized response.
type ChargeHandler struct {
	// Dependencies will be injected here as they are implemented:
	// chargeService *service.ChargeService
	logger *slog.Logger
}

// NewChargeHandler creates a ChargeHandler with its dependencies.
func NewChargeHandler(logger *slog.Logger) *ChargeHandler {
	return &ChargeHandler{
		logger: logger,
	}
}

// ServeHTTP handles the C2P charge request.
func (h *ChargeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Placeholder: will be implemented with the charge service in Phase 1B.
	// For now, returns a structured response showing the API contract.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "error",
		"message": "charge endpoint not yet implemented",
	})
}

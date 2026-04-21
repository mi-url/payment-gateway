// Package handler implements the HTTP handlers for the payment gateway API.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// WebhookBNCHandler handles POST /v1/webhooks/bnc requests.
// BNC sends push notifications here when payments are received.
// Critical: Must respond HTTP 200 IMMEDIATELY before any processing.
type WebhookBNCHandler struct {
	logger *slog.Logger
}

// NewWebhookBNCHandler creates a WebhookBNCHandler with its dependencies.
func NewWebhookBNCHandler(logger *slog.Logger) *WebhookBNCHandler {
	return &WebhookBNCHandler{
		logger: logger,
	}
}

// ServeHTTP handles BNC notification push webhooks.
func (h *WebhookBNCHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Placeholder: will be implemented with the BNC webhook parser in Phase 1B.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

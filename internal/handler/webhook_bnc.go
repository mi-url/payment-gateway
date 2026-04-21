// Package handler implements the HTTP handlers for the payment gateway API.
package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/faloppa/payment-gateway/internal/bank/bnc"
)

// WebhookBNCHandler handles POST /v1/webhooks/bnc requests.
// BNC sends push notifications here when payments are received (P2P, TRF, DEP).
//
// CRITICAL: Must respond HTTP 200 IMMEDIATELY before any business logic processing.
// BNC will disable the webhook if it receives errors or timeouts repeatedly.
type WebhookBNCHandler struct {
	logger *slog.Logger
	// txnStore will be used for matching notifications to transactions.
}

// NewWebhookBNCHandler creates a WebhookBNCHandler with its dependencies.
func NewWebhookBNCHandler(logger *slog.Logger) *WebhookBNCHandler {
	return &WebhookBNCHandler{
		logger: logger,
	}
}

// ServeHTTP handles BNC notification push webhooks.
func (h *WebhookBNCHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Read the body immediately.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("webhook: failed to read body",
			slog.String("error", err.Error()),
		)
		// Still respond 200 to prevent BNC from disabling the webhook.
		w.WriteHeader(http.StatusOK)
		return
	}

	// Respond 200 IMMEDIATELY — BNC requires this before any processing.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})

	// Process asynchronously after responding.
	go h.processNotification(body)
}

// processNotification handles the BNC notification in a background goroutine.
// Any error here is logged but never sent back to BNC.
func (h *WebhookBNCHandler) processNotification(body []byte) {
	notif, err := bnc.ParseWebhookNotification(body)
	if err != nil {
		h.logger.Error("webhook: invalid notification",
			slog.String("error", err.Error()),
			slog.String("raw_body", string(body)),
		)
		return
	}

	h.logger.Info("webhook: notification received",
		slog.String("type", notif.PaymentType),
		slog.String("origin_ref", notif.OriginBankReference),
		slog.String("amount", notif.Amount),
		slog.String("commerce_id", notif.CommerceID),
		slog.String("origin_bank", notif.OriginBankCode),
	)

	// TODO: Match notification to existing transaction and update status.
	// This will be wired when the store is connected to a live database.
}

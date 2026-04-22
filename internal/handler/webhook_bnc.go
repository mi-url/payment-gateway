// Package handler implements the HTTP handlers for the payment gateway API.
package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/faloppa/payment-gateway/internal/bank/bnc"
)

// maxWebhookBodySize limits webhook request bodies to 64KB (BNC notifications are ~500 bytes).
const maxWebhookBodySize = 65_536

// maxProcessedRefs is the maximum number of dedup references kept in memory.
// When exceeded, the oldest half is evicted. In production, replace with DB-backed store.
const maxProcessedRefs = 100_000

// WebhookBNCHandler handles POST /v1/webhooks/bnc requests.
// BNC sends push notifications here when payments are received (P2P, TRF, DEP).
//
// CRITICAL: Must respond HTTP 200 IMMEDIATELY before any business logic processing.
// BNC will disable the webhook if it receives errors or timeouts repeatedly.
//
// Authentication: Validates x-api-key header against the configured webhook secret.
// Deduplication: Tracks processed references to prevent double-accrual.
type WebhookBNCHandler struct {
	logger        *slog.Logger
	webhookAPIKey string
	// processedRefs tracks BNC references already handled to prevent
	// duplicate processing. In production, replace with DB-backed store.
	processedRefs map[string]time.Time
	mu            sync.Mutex
}

// NewWebhookBNCHandler creates a WebhookBNCHandler with its dependencies.
// webhookAPIKey is the secret shared with BNC during onboarding; it will
// arrive in the x-api-key header of every notification.
func NewWebhookBNCHandler(logger *slog.Logger, webhookAPIKey string) *WebhookBNCHandler {
	h := &WebhookBNCHandler{
		logger:        logger,
		webhookAPIKey: webhookAPIKey,
		processedRefs: make(map[string]time.Time),
	}
	// Start background cleanup goroutine to evict refs older than 24 hours.
	go h.cleanupLoop()
	return h
}

// ServeHTTP handles BNC notification push webhooks.
func (h *WebhookBNCHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// ── H1 FIX: Authenticate the request from BNC ──────────────────────
	// BNC sends the shared API key in the x-api-key header.
	// Reject unauthenticated requests with 401 before reading the body.
	incomingKey := r.Header.Get("x-api-key")
	if h.webhookAPIKey != "" && incomingKey != h.webhookAPIKey {
		h.logger.Warn("webhook: rejected unauthenticated request",
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("got_key", maskKey(incomingKey)),
		)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Limit request body to 64KB to prevent DoS.
	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBodySize)

	// Read the body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("webhook: failed to read body",
			slog.String("error", err.Error()),
		)
		// Still respond 200 to prevent BNC from disabling the webhook.
		w.WriteHeader(http.StatusOK)
		return
	}

	// ── H2 FIX: Parse and deduplicate BEFORE responding ────────────────
	// BNC docs: "Es obligatorio que el sistema receptor realice una
	// validación rigurosa de las referencias... para evitar el
	// procesamiento duplicado de notificaciones."
	notif, err := bnc.ParseWebhookNotification(body)
	if err != nil {
		h.logger.Error("webhook: invalid notification",
			slog.String("error", err.Error()),
			slog.String("raw_body", string(body)),
		)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Build a dedup key from the origin bank reference + amount.
	dedupKey := notif.OriginBankReference + ":" + notif.DestinyBankReference + ":" + notif.Amount
	h.mu.Lock()
	_, alreadyProcessed := h.processedRefs[dedupKey]
	if !alreadyProcessed {
		h.processedRefs[dedupKey] = time.Now()
	}
	h.mu.Unlock()

	if alreadyProcessed {
		h.logger.Info("webhook: duplicate notification ignored",
			slog.String("origin_ref", notif.OriginBankReference),
			slog.String("amount", notif.Amount),
		)
		// Respond 200 so BNC stops retrying this notification.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "already_processed"})
		return
	}

	// Respond 200 IMMEDIATELY — BNC requires this before any processing.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})

	// Process asynchronously after responding.
	go h.processNotification(notif)
}

// processNotification handles the BNC notification in a background goroutine.
// Any error here is logged but never sent back to BNC.
func (h *WebhookBNCHandler) processNotification(notif *bnc.WebhookNotification) {
	h.logger.Info("webhook: processing notification",
		slog.String("type", notif.PaymentType),
		slog.String("origin_ref", notif.OriginBankReference),
		slog.String("destiny_ref", notif.DestinyBankReference),
		slog.String("amount", notif.Amount),
		slog.String("commerce_id", notif.CommerceID),
		slog.String("origin_bank", notif.OriginBankCode),
	)

	// TODO: Match notification to existing transaction and update status.
	// 1. Query transactions by bank_reference = notif.DestinyBankReference
	// 2. Verify amount matches
	// 3. Update status to SUCCESS
	// 4. Notify merchant via webhook_url
}

// cleanupLoop runs every hour to evict dedup references older than 24 hours.
// This prevents the processedRefs map from growing infinitely in memory.
func (h *WebhookBNCHandler) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.Lock()
		now := time.Now()
		evicted := 0
		for key, ts := range h.processedRefs {
			if now.Sub(ts) > 24*time.Hour {
				delete(h.processedRefs, key)
				evicted++
			}
		}
		h.mu.Unlock()

		if evicted > 0 {
			h.logger.Info("webhook: evicted stale dedup references",
				slog.Int("evicted", evicted),
				slog.Int("remaining", len(h.processedRefs)),
			)
		}
	}
}

// maskKey returns a masked version of an API key for safe logging.
func maskKey(key string) string {
	if len(key) <= 4 {
		return "***"
	}
	return key[:4] + "***"
}

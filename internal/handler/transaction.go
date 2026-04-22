// Package handler implements the HTTP handlers for the payment gateway API.
package handler

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/faloppa/payment-gateway/internal/middleware"
	"github.com/faloppa/payment-gateway/internal/store"
	"github.com/google/uuid"
)

// TransactionHandler handles GET /v1/transactions/{id} requests.
// Allows merchants to query the status of their transactions.
type TransactionHandler struct {
	txnStore *store.TransactionStore
	logger   *slog.Logger
}

// NewTransactionHandler creates a TransactionHandler with its dependencies.
func NewTransactionHandler(txnStore *store.TransactionStore, logger *slog.Logger) *TransactionHandler {
	return &TransactionHandler{
		txnStore: txnStore,
		logger:   logger,
	}
}

// ServeHTTP handles transaction query requests.
func (h *TransactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract authenticated merchant ID.
	merchantIDStr, ok := r.Context().Value(middleware.MerchantIDKey).(string)
	if !ok || merchantIDStr == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "merchant not authenticated",
		})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid merchant ID",
		})
		return
	}

	// Extract transaction ID from URL path using Go 1.22+ ServeMux pattern.
	txnIDStr := r.PathValue("id")
	if txnIDStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "transaction ID is required in URL path",
		})
		return
	}

	txnID, err := uuid.Parse(txnIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid transaction ID format",
		})
		return
	}

	// Fetch the transaction.
	txn, err := h.txnStore.FindByID(r.Context(), txnID)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "transaction not found",
		})
		return
	}
	if err != nil {
		h.logger.Error("failed to fetch transaction",
			slog.String("txn_id", txnID.String()),
			slog.String("error", err.Error()),
		)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	// Verify the transaction belongs to the authenticated merchant.
	if txn.MerchantID != merchantID {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "transaction not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, txn)
}

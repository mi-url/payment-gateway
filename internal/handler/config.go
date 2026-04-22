// Package handler provides HTTP handlers for the payment gateway.
// This file implements the bank configuration endpoint for storing
// encrypted bank credentials via Envelope Encryption.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/faloppa/payment-gateway/internal/crypto"
	"github.com/faloppa/payment-gateway/internal/middleware"
	"github.com/faloppa/payment-gateway/internal/store"
	"github.com/google/uuid"
)

// bankConfigRequest is the payload for POST /v1/config/bank.
// Credentials arrive as plaintext from the dashboard, are encrypted
// in this handler, and only ciphertext reaches the database.
type bankConfigRequest struct {
	BankCode string          `json:"bank_code"`
	Credentials json.RawMessage `json:"credentials"`
}

// BankConfigHandler handles bank credential management endpoints.
type BankConfigHandler struct {
	store     *store.BankConfigStore
	encryptor *crypto.EnvelopeEncryptor
	logger    *slog.Logger
}

// NewBankConfigHandler creates a handler for bank configuration endpoints.
func NewBankConfigHandler(s *store.BankConfigStore, enc *crypto.EnvelopeEncryptor, logger *slog.Logger) *BankConfigHandler {
	return &BankConfigHandler{store: s, encryptor: enc, logger: logger}
}

// ServeHTTP handles POST requests to store bank credentials.
// Flow: receive plaintext credentials → envelope encrypt → store ciphertext → respond.
func (h *BankConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract merchant ID from auth middleware context.
	merchantIDStr, ok := r.Context().Value(middleware.MerchantIDKey).(string)
	if !ok || merchantIDStr == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid merchant ID"}`, http.StatusBadRequest)
		return
	}

	// Parse request body.
	var req bankConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.BankCode == "" {
		http.Error(w, `{"error":"bank_code is required"}`, http.StatusBadRequest)
		return
	}
	if len(req.Credentials) == 0 {
		http.Error(w, `{"error":"credentials are required"}`, http.StatusBadRequest)
		return
	}

	// Envelope encrypt the credentials.
	// After this call, the plaintext credentials only exist in memory
	// and will be zeroed when this function returns.
	encrypted, err := h.encryptor.Encrypt(req.Credentials)
	if err != nil {
		h.logger.Error("failed to encrypt bank credentials",
			slog.String("merchant_id", merchantID.String()),
			slog.String("bank_code", req.BankCode),
			slog.String("error", err.Error()),
		)
		http.Error(w, `{"error":"encryption failed"}`, http.StatusInternalServerError)
		return
	}

	// Store encrypted credentials in the database.
	err = h.store.Upsert(r.Context(), merchantID, req.BankCode,
		encrypted.Ciphertext, encrypted.SealedDEK, encrypted.IV,
	)
	if err != nil {
		h.logger.Error("failed to store bank config",
			slog.String("merchant_id", merchantID.String()),
			slog.String("bank_code", req.BankCode),
			slog.String("error", err.Error()),
		)
		http.Error(w, `{"error":"failed to save configuration"}`, http.StatusInternalServerError)
		return
	}

	// Zero the plaintext credentials from memory.
	crypto.ZeroBytes(req.Credentials)

	h.logger.Info("bank credentials stored",
		slog.String("merchant_id", merchantID.String()),
		slog.String("bank_code", req.BankCode),
		slog.Int("ciphertext_bytes", len(encrypted.Ciphertext)),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "configured",
		"bank_code": req.BankCode,
		"encrypted": true,
	})
}

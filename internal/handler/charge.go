// Package handler implements the HTTP handlers for the payment gateway API.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/faloppa/payment-gateway/internal/middleware"
	"github.com/faloppa/payment-gateway/internal/model"
	"github.com/faloppa/payment-gateway/internal/service"
	"github.com/google/uuid"
)

// ChargeHandler handles POST /v1/charges/c2p requests.
// It validates the request, extracts the authenticated merchant from context,
// and delegates to the ChargeService for processing.
type ChargeHandler struct {
	chargeService *service.ChargeService
	logger        *slog.Logger
}

// NewChargeHandler creates a ChargeHandler with its dependencies.
func NewChargeHandler(chargeService *service.ChargeService, logger *slog.Logger) *ChargeHandler {
	return &ChargeHandler{
		chargeService: chargeService,
		logger:        logger,
	}
}

// ServeHTTP handles the C2P charge request.
func (h *ChargeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract authenticated merchant ID from context (set by auth middleware).
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

	// Limit request body to 1MB to prevent DoS via oversized payloads.
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)

	// Parse and validate the charge request.
	var req model.C2PChargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON body",
		})
		return
	}

	if err := validateChargeRequest(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Ensure the idempotency key in the body matches the header.
	// The middleware uses the header for dedup; the service uses the body for DB storage.
	// A mismatch would cause silent inconsistencies.
	if headerKey := r.Header.Get("Idempotency-Key"); headerKey != "" && headerKey != req.IdempotencyKey {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "idempotency_key in body must match Idempotency-Key header",
		})
		return
	}

	// Process the charge.
	resp, err := h.chargeService.ProcessC2P(r.Context(), merchantID, &req)
	if err != nil {
		h.logger.Error("charge processing failed",
			slog.String("merchant_id", merchantID.String()),
			slog.String("error", err.Error()),
		)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	// Determine HTTP status code from transaction status.
	statusCode := http.StatusOK
	if resp.Status == model.StatusDeclined {
		statusCode = http.StatusUnprocessableEntity
	} else if resp.Status == model.StatusPendingReconciliation {
		statusCode = http.StatusAccepted
	}

	writeJSON(w, statusCode, resp)
}

// validateChargeRequest performs basic validation on the charge request fields.
func validateChargeRequest(req *model.C2PChargeRequest) error {
	if req.Amount.IsZero() || req.Amount.IsNegative() {
		return errValidation("amount must be greater than 0")
	}
	if req.Currency == "" {
		req.Currency = "VES"
	}
	if len(req.Currency) != 3 {
		return errValidation("currency must be a 3-letter code")
	}
	if req.Payer.BankCode == "" || len(req.Payer.BankCode) != 4 {
		return errValidation("payer.bank_code must be a 4-digit code")
	}
	if req.Payer.Phone == "" {
		return errValidation("payer.phone is required")
	}
	if req.Payer.IDDocument == "" {
		return errValidation("payer.id_document is required")
	}
	if req.Payer.OTP == "" {
		return errValidation("payer.otp is required")
	}
	if req.IdempotencyKey == "" {
		return errValidation("idempotency_key is required")
	}
	return nil
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// validationError is a simple error type for request validation failures.
type validationError struct {
	message string
}

func (e *validationError) Error() string { return e.message }

func errValidation(msg string) error { return &validationError{message: msg} }

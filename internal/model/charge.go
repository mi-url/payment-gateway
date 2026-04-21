// Package model defines the core domain types for the payment gateway.
package model

// C2PChargeRequest is the unified request format that all clients (SDKs, plugins,
// direct API calls) send to initiate a Cobro a Persona (C2P) charge.
// The gateway translates this into bank-specific formats via adapters.
type C2PChargeRequest struct {
	Amount         float64           `json:"amount" validate:"required,gt=0"`
	Currency       string            `json:"currency" validate:"required,len=3"`
	Payer          PayerInfo         `json:"payer" validate:"required"`
	IdempotencyKey string            `json:"idempotency_key" validate:"required,max=255"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// PayerInfo contains the payer's identification for a C2P charge.
// All fields are required for the Venezuelan Pago Móvil protocol.
type PayerInfo struct {
	BankCode   string `json:"bank_code" validate:"required,len=4"`
	Phone      string `json:"phone" validate:"required"`
	IDDocument string `json:"id_document" validate:"required"`
	OTP        string `json:"otp" validate:"required"`
}

// C2PChargeResponse is the standardized response returned to all clients
// after a charge attempt, regardless of which bank processed it.
type C2PChargeResponse struct {
	ID            string            `json:"id"`
	Status        TransactionStatus `json:"status"`
	BankReference string            `json:"bank_reference,omitempty"`
	Amount        float64           `json:"amount"`
	Currency      string            `json:"currency"`
	Error         *ChargeError      `json:"error,omitempty"`
	CreatedAt     string            `json:"created_at"`
}

// ChargeError provides structured error information when a charge fails.
// The Code field is a gateway-standard error code (e.g., INCORRECT_OTP),
// while BankCode is the raw code from the bank (e.g., G55 for BNC).
type ChargeError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	BankCode string `json:"bank_code,omitempty"`
}

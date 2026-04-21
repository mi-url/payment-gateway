// Package bank defines the adapter interface for bank integrations.
// Each bank (BNC, Mercantil, Banesco, etc.) implements this interface.
// The gateway routes transactions to the correct adapter based on bank_code.
package bank

import (
	"context"
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// Adapter defines the contract every bank integration must fulfill.
// Adding a new bank means implementing this interface and registering it
// in the Registry. No other code in the gateway needs to change.
type Adapter interface {
	// ProcessC2P executes a Cobro a Persona (C2P) charge against the bank API.
	// The creds parameter contains the bank-specific credentials (already decrypted).
	// Returns a BankResponse with the result or an error for infrastructure failures.
	ProcessC2P(ctx context.Context, req *C2PRequest, creds json.RawMessage) (*Response, error)

	// QueryTransaction checks whether a transaction was processed by the bank.
	// Used by the reconciliation worker to resolve PENDING_RECONCILIATION transactions.
	QueryTransaction(ctx context.Context, ref string, amount decimal.Decimal, date time.Time, creds json.RawMessage) (*QueryResult, error)

	// BankCode returns the 4-digit code identifying this bank (e.g., "0191" for BNC).
	BankCode() string

	// ValidateCredentials verifies that the provided credentials are valid
	// for this bank. Called during merchant onboarding.
	ValidateCredentials(ctx context.Context, creds json.RawMessage) error
}

// C2PRequest is the bank-adapter-facing charge request.
// It contains only the fields needed by the bank, not merchant metadata.
type C2PRequest struct {
	Amount         decimal.Decimal `json:"amount"`
	PayerBankCode  string          `json:"payer_bank_code"`
	PayerPhone     string          `json:"payer_phone"`
	PayerIDDoc     string          `json:"payer_id_document"`
	OTP            string          `json:"otp"`
	Reference      string          `json:"reference"`
	ChildClientID  string          `json:"child_client_id,omitempty"`
	BranchID       string          `json:"branch_id,omitempty"`
}

// Response is the standardized result from a bank adapter after processing.
type Response struct {
	Success       bool   `json:"success"`
	BankReference string `json:"bank_reference,omitempty"`
	ErrorCode     string `json:"error_code,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
	RawResponse   string `json:"raw_response,omitempty"`
}

// QueryResult contains the outcome of a transaction query for reconciliation.
type QueryResult struct {
	Exists bool `json:"exists"`
}

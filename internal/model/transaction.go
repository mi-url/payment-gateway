// Package model defines the core domain types for the payment gateway.
// These types represent the business entities that flow through the system:
// merchants, transactions, and bank configurations.
package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransactionStatus represents the state machine for payment processing.
// Valid transitions:
//
//	INITIATED → PROCESSING (charge submitted to bank adapter)
//	PROCESSING → SUCCESS (bank confirmed the charge)
//	PROCESSING → DECLINED (bank rejected: bad OTP, insufficient funds, etc.)
//	PROCESSING → BANK_NETWORK_ERROR (pre-TLS failure, safe to retry)
//	PROCESSING → PENDING_RECONCILIATION (timeout or ambiguous response)
type TransactionStatus string

const (
	StatusInitiated             TransactionStatus = "INITIATED"
	StatusProcessing            TransactionStatus = "PROCESSING"
	StatusSuccess               TransactionStatus = "SUCCESS"
	StatusDeclined              TransactionStatus = "DECLINED"
	StatusBankNetworkError      TransactionStatus = "BANK_NETWORK_ERROR"
	StatusPendingReconciliation TransactionStatus = "PENDING_RECONCILIATION"
)

// validTransitions defines the allowed state machine transitions.
// Any transition not listed here is a programming error.
var validTransitions = map[TransactionStatus][]TransactionStatus{
	StatusInitiated:  {StatusProcessing, StatusDeclined},
	StatusProcessing: {StatusSuccess, StatusDeclined, StatusBankNetworkError, StatusPendingReconciliation},
	// Terminal states: no further transitions allowed.
	StatusSuccess:               {},
	StatusDeclined:              {},
	StatusBankNetworkError:      {},
	StatusPendingReconciliation: {StatusSuccess, StatusDeclined},
}

// CanTransitionTo checks if a state transition is valid.
func (s TransactionStatus) CanTransitionTo(target TransactionStatus) bool {
	allowed, exists := validTransitions[s]
	if !exists {
		return false
	}
	for _, t := range allowed {
		if t == target {
			return true
		}
	}
	return false
}

// Transaction represents a single payment attempt in the gateway ledger.
type Transaction struct {
	ID             uuid.UUID         `json:"id" db:"id"`
	MerchantID     uuid.UUID         `json:"merchant_id" db:"merchant_id"`
	IdempotencyKey string            `json:"idempotency_key" db:"idempotency_key"`
	BankCode       string            `json:"bank_code" db:"bank_code"`
	Amount         decimal.Decimal   `json:"amount" db:"amount"`
	Currency       string            `json:"currency" db:"currency"`
	PayerPhone     string            `json:"payer_phone,omitempty" db:"payer_phone"`
	PayerIDDoc     string            `json:"payer_id_document,omitempty" db:"payer_id_document"`
	PayerBankCode  string            `json:"payer_bank_code,omitempty" db:"payer_bank_code"`
	BankReference  string            `json:"bank_reference,omitempty" db:"bank_reference"`
	Status         TransactionStatus `json:"status" db:"status"`
	ErrorCode      string            `json:"error_code,omitempty" db:"error_code"`
	ErrorMessage   string            `json:"error_message,omitempty" db:"error_message"`
	Metadata       map[string]string `json:"metadata,omitempty" db:"metadata"`
	InitiatedAt    time.Time         `json:"initiated_at" db:"initiated_at"`
	ProcessedAt    *time.Time        `json:"processed_at,omitempty" db:"processed_at"`
	CompletedAt    *time.Time        `json:"completed_at,omitempty" db:"completed_at"`
}

// Transition moves the transaction to a new status if the transition is valid.
// Returns an error if the transition violates the state machine rules.
func (t *Transaction) Transition(target TransactionStatus) error {
	if !t.Status.CanTransitionTo(target) {
		return fmt.Errorf("invalid transition: %s → %s", t.Status, target)
	}
	t.Status = target
	now := time.Now().UTC()
	switch target {
	case StatusProcessing:
		t.ProcessedAt = &now
	case StatusSuccess, StatusDeclined, StatusBankNetworkError, StatusPendingReconciliation:
		t.CompletedAt = &now
	}
	return nil
}

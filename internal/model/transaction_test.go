package model

import (
	"testing"
)

// TestValidTransitions verifies that all documented state machine transitions
// are allowed according to the transition rules.
func TestValidTransitions(t *testing.T) {
	valid := []struct {
		from TransactionStatus
		to   TransactionStatus
	}{
		{StatusInitiated, StatusProcessing},
		{StatusProcessing, StatusSuccess},
		{StatusProcessing, StatusDeclined},
		{StatusProcessing, StatusBankNetworkError},
		{StatusProcessing, StatusPendingReconciliation},
		{StatusPendingReconciliation, StatusSuccess},
		{StatusPendingReconciliation, StatusDeclined},
	}

	for _, tc := range valid {
		if !tc.from.CanTransitionTo(tc.to) {
			t.Errorf("expected %s → %s to be valid, but it was rejected", tc.from, tc.to)
		}
	}
}

// TestInvalidTransitions verifies that illegal state transitions are rejected.
func TestInvalidTransitions(t *testing.T) {
	invalid := []struct {
		from TransactionStatus
		to   TransactionStatus
	}{
		// Cannot go backwards.
		{StatusProcessing, StatusInitiated},
		{StatusSuccess, StatusInitiated},
		{StatusSuccess, StatusProcessing},

		// Terminal states cannot transition.
		{StatusSuccess, StatusDeclined},
		{StatusDeclined, StatusSuccess},
		{StatusDeclined, StatusProcessing},
		{StatusBankNetworkError, StatusProcessing},
		{StatusBankNetworkError, StatusSuccess},

		// Cannot skip states.
		{StatusInitiated, StatusSuccess},
		{StatusInitiated, StatusDeclined},
		{StatusInitiated, StatusPendingReconciliation},
	}

	for _, tc := range invalid {
		if tc.from.CanTransitionTo(tc.to) {
			t.Errorf("expected %s → %s to be INVALID, but it was allowed", tc.from, tc.to)
		}
	}
}

// TestTransactionTransition verifies the Transaction.Transition method
// updates status and timestamps correctly.
func TestTransactionTransition(t *testing.T) {
	txn := &Transaction{
		Status: StatusInitiated,
	}

	// Valid transition: INITIATED → PROCESSING
	if err := txn.Transition(StatusProcessing); err != nil {
		t.Fatalf("expected valid transition, got error: %v", err)
	}
	if txn.Status != StatusProcessing {
		t.Errorf("status = %s, want PROCESSING", txn.Status)
	}
	if txn.ProcessedAt == nil {
		t.Error("ProcessedAt should be set after transitioning to PROCESSING")
	}

	// Valid transition: PROCESSING → SUCCESS
	if err := txn.Transition(StatusSuccess); err != nil {
		t.Fatalf("expected valid transition, got error: %v", err)
	}
	if txn.Status != StatusSuccess {
		t.Errorf("status = %s, want SUCCESS", txn.Status)
	}
	if txn.CompletedAt == nil {
		t.Error("CompletedAt should be set after transitioning to SUCCESS")
	}

	// Invalid transition: SUCCESS → PROCESSING (terminal state)
	if err := txn.Transition(StatusProcessing); err == nil {
		t.Error("expected error for invalid transition SUCCESS → PROCESSING, got nil")
	}
}

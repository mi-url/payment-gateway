// Package store implements database access for the payment gateway.
package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/faloppa/payment-gateway/internal/model"
	"github.com/google/uuid"
)

// TransactionStore handles database operations for transactions.
type TransactionStore struct {
	db *sql.DB
}

// NewTransactionStore creates a TransactionStore with the given database connection.
func NewTransactionStore(db *sql.DB) *TransactionStore {
	return &TransactionStore{db: db}
}

// Create inserts a new transaction record with INITIATED status.
// The idempotency_key unique constraint prevents duplicate inserts.
func (s *TransactionStore) Create(ctx context.Context, txn *model.Transaction) error {
	query := `
		INSERT INTO transactions (
			id, merchant_id, idempotency_key, bank_code, amount, currency,
			payer_phone, payer_id_document, payer_bank_code, status, metadata, initiated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := s.db.ExecContext(ctx, query,
		txn.ID, txn.MerchantID, txn.IdempotencyKey, txn.BankCode,
		txn.Amount, txn.Currency, txn.PayerPhone, txn.PayerIDDoc,
		txn.PayerBankCode, txn.Status, txn.Metadata, txn.InitiatedAt,
	)
	return err
}

// UpdateStatus transitions a transaction to a new status with optional bank reference
// and error details. Enforces the state machine at the SQL level with a WHERE clause.
func (s *TransactionStore) UpdateStatus(ctx context.Context, id uuid.UUID, from, to model.TransactionStatus, bankRef, errCode, errMsg string) error {
	now := time.Now().UTC()

	query := `
		UPDATE transactions
		SET status = $1, bank_reference = $2, error_code = $3, error_message = $4,
		    processed_at = CASE WHEN $1 = 'PROCESSING' THEN $5 ELSE processed_at END,
		    completed_at = CASE WHEN $1 IN ('SUCCESS','DECLINED','BANK_NETWORK_ERROR','PENDING_RECONCILIATION') THEN $5 ELSE completed_at END
		WHERE id = $6 AND status = $7
	`

	result, err := s.db.ExecContext(ctx, query, to, nullString(bankRef), nullString(errCode), nullString(errMsg), now, id, from)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrStaleTransaction
	}
	return nil
}

// FindByID retrieves a transaction by its UUID.
func (s *TransactionStore) FindByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	query := `
		SELECT id, merchant_id, idempotency_key, bank_code, amount, currency,
		       payer_phone, payer_id_document, payer_bank_code, bank_reference,
		       status, error_code, error_message, metadata,
		       initiated_at, processed_at, completed_at
		FROM transactions
		WHERE id = $1
	`

	var txn model.Transaction
	var bankRef, errCode, errMsg, payerPhone, payerIDDoc, payerBankCode sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&txn.ID, &txn.MerchantID, &txn.IdempotencyKey, &txn.BankCode,
		&txn.Amount, &txn.Currency, &payerPhone, &payerIDDoc,
		&payerBankCode, &bankRef, &txn.Status, &errCode, &errMsg,
		&txn.Metadata, &txn.InitiatedAt, &txn.ProcessedAt, &txn.CompletedAt,
	)
	if err != nil {
		return nil, err
	}

	txn.PayerPhone = payerPhone.String
	txn.PayerIDDoc = payerIDDoc.String
	txn.PayerBankCode = payerBankCode.String
	txn.BankReference = bankRef.String
	txn.ErrorCode = errCode.String
	txn.ErrorMessage = errMsg.String

	return &txn, nil
}

// FindPendingReconciliation returns all transactions stuck in PENDING_RECONCILIATION
// status, ordered by oldest first. Used by the reconciliation background worker.
func (s *TransactionStore) FindPendingReconciliation(ctx context.Context, limit int) ([]*model.Transaction, error) {
	query := `
		SELECT id, merchant_id, idempotency_key, bank_code, amount, currency,
		       bank_reference, status, initiated_at
		FROM transactions
		WHERE status = 'PENDING_RECONCILIATION'
		ORDER BY initiated_at ASC
		LIMIT $1
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.Transaction
	for rows.Next() {
		var txn model.Transaction
		var bankRef sql.NullString
		if err := rows.Scan(
			&txn.ID, &txn.MerchantID, &txn.IdempotencyKey, &txn.BankCode,
			&txn.Amount, &txn.Currency, &bankRef, &txn.Status, &txn.InitiatedAt,
		); err != nil {
			return nil, err
		}
		txn.BankReference = bankRef.String
		result = append(result, &txn)
	}
	return result, rows.Err()
}

// nullString converts an empty string to sql.NullString for nullable DB columns.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

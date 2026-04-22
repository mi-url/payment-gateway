// Package store implements database access for the payment gateway.
package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/faloppa/payment-gateway/internal/model"
	"github.com/google/uuid"
)

// ErrStaleTransaction indicates a concurrent modification where the expected
// status no longer matches. This prevents double-processing of transactions.
var ErrStaleTransaction = errors.New("transaction status has already changed (stale update)")

// BankConfigStore handles database operations for merchant bank configurations.
type BankConfigStore struct {
	db *sql.DB
}

// NewBankConfigStore creates a BankConfigStore with the given database connection.
func NewBankConfigStore(db *sql.DB) *BankConfigStore {
	return &BankConfigStore{db: db}
}

// FindActive retrieves the active bank configuration for a merchant and bank code.
// Returns the encrypted credentials — decryption happens in the crypto layer.
func (s *BankConfigStore) FindActive(ctx context.Context, merchantID uuid.UUID, bankCode string) (*model.BankConfig, error) {
	query := `
		SELECT id, merchant_id, bank_code, encrypted_credentials,
		       kms_data_key_ciphertext, encryption_iv, is_active, created_at
		FROM merchant_bank_configs
		WHERE merchant_id = $1 AND bank_code = $2 AND is_active = true
	`

	var cfg model.BankConfig
	err := s.db.QueryRowContext(ctx, query, merchantID, bankCode).Scan(
		&cfg.ID, &cfg.MerchantID, &cfg.BankCode, &cfg.EncryptedCredentials,
		&cfg.KMSDataKeyCiphertext, &cfg.EncryptionIV, &cfg.IsActive, &cfg.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrBankConfigNotFound
	}
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ErrBankConfigNotFound indicates no active bank configuration exists
// for the requested merchant and bank code combination.
var ErrBankConfigNotFound = errors.New("no active bank configuration found for this merchant and bank code")

// Upsert stores encrypted bank credentials for a merchant. If an active config
// already exists for this merchant+bank, it is deactivated first (soft-replace).
// The caller is responsible for encrypting credentials before calling this.
func (s *BankConfigStore) Upsert(ctx context.Context, merchantID uuid.UUID, bankCode string, encCreds, sealedDEK, iv []byte) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Deactivate any existing active config for this merchant+bank.
	_, err = tx.ExecContext(ctx,
		`UPDATE merchant_bank_configs SET is_active = false
		 WHERE merchant_id = $1 AND bank_code = $2 AND is_active = true`,
		merchantID, bankCode,
	)
	if err != nil {
		return err
	}

	// Insert the new active config.
	_, err = tx.ExecContext(ctx,
		`INSERT INTO merchant_bank_configs
		 (merchant_id, bank_code, encrypted_credentials, kms_data_key_ciphertext, encryption_iv)
		 VALUES ($1, $2, $3, $4, $5)`,
		merchantID, bankCode, encCreds, sealedDEK, iv,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}


// Package store implements database access for the payment gateway.
// All queries are parameterized to prevent SQL injection.
// This package is the only layer that interacts with PostgreSQL directly.
package store

import (
	"context"
	"database/sql"

	"github.com/faloppa/payment-gateway/internal/model"
	"github.com/google/uuid"
)

// MerchantStore handles database operations for merchants.
type MerchantStore struct {
	db *sql.DB
}

// NewMerchantStore creates a MerchantStore with the given database connection.
func NewMerchantStore(db *sql.DB) *MerchantStore {
	return &MerchantStore{db: db}
}

// FindByAPIKeyHash retrieves a merchant by the SHA-256 hash of their API key.
// Returns the merchant and true if found, or nil and false if not.
func (s *MerchantStore) FindByAPIKeyHash(ctx context.Context, hash string) (*model.Merchant, bool, error) {
	query := `
		SELECT id, name, api_key_hash, saas_subscription_status, webhook_url, created_at, updated_at
		FROM merchants
		WHERE api_key_hash = $1
	`

	var m model.Merchant
	var webhookURL sql.NullString

	err := s.db.QueryRowContext(ctx, query, hash).Scan(
		&m.ID, &m.Name, &m.APIKeyHash, &m.SaaSSubscriptionStatus,
		&webhookURL, &m.CreatedAt, &m.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	if webhookURL.Valid {
		m.WebhookURL = webhookURL.String
	}

	return &m, true, nil
}

// FindByID retrieves a merchant by their UUID.
func (s *MerchantStore) FindByID(ctx context.Context, id uuid.UUID) (*model.Merchant, error) {
	query := `
		SELECT id, name, api_key_hash, saas_subscription_status, webhook_url, created_at, updated_at
		FROM merchants
		WHERE id = $1
	`

	var m model.Merchant
	var webhookURL sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.Name, &m.APIKeyHash, &m.SaaSSubscriptionStatus,
		&webhookURL, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if webhookURL.Valid {
		m.WebhookURL = webhookURL.String
	}

	return &m, nil
}

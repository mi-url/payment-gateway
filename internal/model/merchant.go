// Package model defines the core domain types for the payment gateway.
package model

import (
	"time"

	"github.com/google/uuid"
)

// Merchant represents a B2B client who uses the gateway to process payments.
// Merchants authenticate via API keys (stored as SHA-256 hashes) and receive
// transaction notifications at their configured webhook URL.
type Merchant struct {
	ID                     uuid.UUID `json:"id" db:"id"`
	Name                   string    `json:"name" db:"name"`
	APIKeyHash             string    `json:"-" db:"api_key_hash"` // Never serialized to JSON
	SaaSSubscriptionStatus string    `json:"saas_subscription_status" db:"saas_subscription_status"`
	WebhookURL             string    `json:"webhook_url,omitempty" db:"webhook_url"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

// BankConfig holds a merchant's encrypted bank credentials for a specific bank.
// Credentials are protected via Envelope Encryption:
//   - EncryptedCredentials: AES-256-GCM ciphertext of the bank credential JSON
//   - KMSDataKeyCiphertext: The Data Encryption Key (DEK), sealed by Cloud KMS
//   - EncryptionIV: The nonce/IV used for AES-GCM encryption
//
// To decrypt: send KMSDataKeyCiphertext to KMS Decrypt → get DEK plaintext →
// use DEK + EncryptionIV to AES-GCM decrypt EncryptedCredentials → zero memory.
type BankConfig struct {
	ID                    uuid.UUID `json:"id" db:"id"`
	MerchantID            uuid.UUID `json:"merchant_id" db:"merchant_id"`
	BankCode              string    `json:"bank_code" db:"bank_code"`
	EncryptedCredentials  []byte    `json:"-" db:"encrypted_credentials"`
	KMSDataKeyCiphertext  []byte    `json:"-" db:"kms_data_key_ciphertext"`
	EncryptionIV          []byte    `json:"-" db:"encryption_iv"`
	IsActive              bool      `json:"is_active" db:"is_active"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
}

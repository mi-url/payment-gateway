// Package crypto implements Envelope Encryption using GCP Cloud KMS.
//
// Envelope Encryption works as follows:
//  1. Generate a Data Encryption Key (DEK) via KMS GenerateDataKey.
//  2. Encrypt the merchant's bank credentials with the DEK using AES-256-GCM.
//  3. Store: encrypted credentials + KMS-sealed DEK + IV in the database.
//  4. To decrypt: send sealed DEK to KMS Decrypt → get plaintext DEK → decrypt credentials → zero memory.
//
// This ensures that even if the database is fully compromised, credentials
// remain unreadable without KMS access (which is restricted to the Cloud Run
// service account via IAM).
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// KMSClient defines the interface for interacting with a Key Management Service.
// In production, this wraps GCP Cloud KMS. In tests, it uses a mock.
type KMSClient interface {
	// GenerateDataKey creates a new DEK and returns both the plaintext and
	// KMS-encrypted (sealed) versions.
	GenerateDataKey(keyResourceName string) (plaintext, ciphertext []byte, err error)

	// Decrypt sends a sealed DEK to KMS and returns the plaintext DEK.
	Decrypt(keyResourceName string, ciphertext []byte) ([]byte, error)
}

// EnvelopeEncryptor performs Envelope Encryption operations.
type EnvelopeEncryptor struct {
	kms             KMSClient
	keyResourceName string
}

// NewEnvelopeEncryptor creates an encryptor bound to a specific KMS key.
func NewEnvelopeEncryptor(kms KMSClient, keyResourceName string) *EnvelopeEncryptor {
	return &EnvelopeEncryptor{
		kms:             kms,
		keyResourceName: keyResourceName,
	}
}

// EncryptedData holds the result of an envelope encryption operation.
// All three fields must be stored together in the database.
type EncryptedData struct {
	Ciphertext         []byte // AES-256-GCM encrypted data
	SealedDEK          []byte // KMS-encrypted Data Encryption Key
	IV                 []byte // AES-GCM nonce/IV
}

// Encrypt generates a new DEK, encrypts the plaintext with AES-256-GCM,
// and returns the ciphertext along with the KMS-sealed DEK and IV.
func (e *EnvelopeEncryptor) Encrypt(plaintext []byte) (*EncryptedData, error) {
	dekPlaintext, dekCiphertext, err := e.kms.GenerateDataKey(e.keyResourceName)
	if err != nil {
		return nil, fmt.Errorf("envelope encrypt: failed to generate DEK: %w", err)
	}
	defer ZeroBytes(dekPlaintext)

	block, err := aes.NewCipher(dekPlaintext)
	if err != nil {
		return nil, fmt.Errorf("envelope encrypt: failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("envelope encrypt: failed to create GCM: %w", err)
	}

	iv := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("envelope encrypt: failed to generate IV: %w", err)
	}

	ciphertext := gcm.Seal(nil, iv, plaintext, nil)

	return &EncryptedData{
		Ciphertext: ciphertext,
		SealedDEK:  dekCiphertext,
		IV:         iv,
	}, nil
}

// Decrypt unseals the DEK via KMS, then decrypts the ciphertext with AES-256-GCM.
// The plaintext DEK is zeroed from memory immediately after use.
func (e *EnvelopeEncryptor) Decrypt(data *EncryptedData) ([]byte, error) {
	dekPlaintext, err := e.kms.Decrypt(e.keyResourceName, data.SealedDEK)
	if err != nil {
		return nil, fmt.Errorf("envelope decrypt: failed to unseal DEK: %w", err)
	}
	defer ZeroBytes(dekPlaintext)

	block, err := aes.NewCipher(dekPlaintext)
	if err != nil {
		return nil, fmt.Errorf("envelope decrypt: failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("envelope decrypt: failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, data.IV, data.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("envelope decrypt: failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// ZeroBytes overwrites a byte slice with zeros to remove sensitive
// material from memory. Called via defer after using plaintext DEKs.
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

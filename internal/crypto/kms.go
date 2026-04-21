// Package crypto implements Envelope Encryption using GCP Cloud KMS.
package crypto

import (
	"crypto/rand"
	"fmt"
)

// MockKMS is a test/development implementation of KMSClient that generates
// random DEKs locally instead of calling GCP Cloud KMS. It provides the
// same encryption behavior but without external dependencies.
//
// WARNING: This must NEVER be used in production. It exists only for
// local development and testing when GCP credentials are not available.
type MockKMS struct{}

// NewMockKMS creates a mock KMS client for development and testing.
func NewMockKMS() *MockKMS {
	return &MockKMS{}
}

// GenerateDataKey creates a random 32-byte DEK. In mock mode, the "sealed"
// version is identical to the plaintext (no real KMS encryption).
func (m *MockKMS) GenerateDataKey(keyResourceName string) (plaintext, ciphertext []byte, err error) {
	dek := make([]byte, 32)
	if _, err := rand.Read(dek); err != nil {
		return nil, nil, fmt.Errorf("mock kms: failed to generate random key: %w", err)
	}
	// In mock mode, sealed == plaintext (no real KMS wrapping).
	sealed := make([]byte, len(dek))
	copy(sealed, dek)
	return dek, sealed, nil
}

// Decrypt returns the ciphertext unchanged (since MockKMS doesn't actually seal).
func (m *MockKMS) Decrypt(keyResourceName string, ciphertext []byte) ([]byte, error) {
	plaintext := make([]byte, len(ciphertext))
	copy(plaintext, ciphertext)
	return plaintext, nil
}

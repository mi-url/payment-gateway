// Package crypto implements Envelope Encryption using GCP Cloud KMS.
package crypto

import (
	"context"
	"crypto/rand"
	"fmt"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
)

// GCPCloudKMS implements KMSClient using Google Cloud KMS.
// This is the production implementation that provides real key management.
// The MockKMS should NEVER be used in production.
type GCPCloudKMS struct {
	client *kms.KeyManagementClient
}

// NewGCPCloudKMS creates a GCP Cloud KMS client.
// It uses Application Default Credentials (ADC), which are automatically
// provided by Cloud Run via the service account.
func NewGCPCloudKMS(ctx context.Context) (*GCPCloudKMS, error) {
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcp kms: failed to create client: %w", err)
	}
	return &GCPCloudKMS{client: client}, nil
}

// Close releases the KMS client resources. Should be called via defer.
func (g *GCPCloudKMS) Close() error {
	return g.client.Close()
}

// GenerateDataKey creates a random 32-byte DEK and encrypts (seals) it
// using Cloud KMS. Returns both the plaintext DEK (for immediate use)
// and the sealed DEK (for database storage).
func (g *GCPCloudKMS) GenerateDataKey(keyResourceName string) (plaintext, ciphertext []byte, err error) {
	// Generate a random 32-byte DEK locally.
	dek := make([]byte, 32)
	if _, err := rand.Read(dek); err != nil {
		return nil, nil, fmt.Errorf("gcp kms: failed to generate DEK: %w", err)
	}

	// Seal (encrypt) the DEK using Cloud KMS.
	resp, err := g.client.Encrypt(context.Background(), &kmspb.EncryptRequest{
		Name:      keyResourceName,
		Plaintext: dek,
	})
	if err != nil {
		ZeroBytes(dek)
		return nil, nil, fmt.Errorf("gcp kms: failed to seal DEK: %w", err)
	}

	return dek, resp.Ciphertext, nil
}

// Decrypt unseals a DEK by sending it to Cloud KMS for decryption.
// Returns the plaintext DEK for use in AES-256-GCM decryption.
func (g *GCPCloudKMS) Decrypt(keyResourceName string, ciphertext []byte) ([]byte, error) {
	resp, err := g.client.Decrypt(context.Background(), &kmspb.DecryptRequest{
		Name:       keyResourceName,
		Ciphertext: ciphertext,
	})
	if err != nil {
		return nil, fmt.Errorf("gcp kms: failed to unseal DEK: %w", err)
	}
	return resp.Plaintext, nil
}

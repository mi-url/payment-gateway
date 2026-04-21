package crypto

import (
	"bytes"
	"testing"
)

// TestEnvelopeEncryptDecryptRoundtrip verifies that Envelope Encryption
// correctly encrypts and decrypts data using the mock KMS.
func TestEnvelopeEncryptDecryptRoundtrip(t *testing.T) {
	kms := NewMockKMS()
	encryptor := NewEnvelopeEncryptor(kms, "test-key-resource")

	original := []byte(`{"client_guid":"4A074C46","master_key":"secret123"}`)

	encrypted, err := encryptor.Encrypt(original)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if bytes.Equal(encrypted.Ciphertext, original) {
		t.Error("ciphertext should not equal plaintext")
	}

	if len(encrypted.SealedDEK) == 0 {
		t.Error("SealedDEK should not be empty")
	}

	if len(encrypted.IV) == 0 {
		t.Error("IV should not be empty")
	}

	decrypted, err := encryptor.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, original) {
		t.Errorf("roundtrip mismatch:\n  got:  %q\n  want: %q", decrypted, original)
	}
}

// TestZeroBytes verifies that ZeroBytes actually zeroes the memory.
func TestZeroBytes(t *testing.T) {
	secret := []byte("sensitive-data-here")
	ZeroBytes(secret)

	for i, b := range secret {
		if b != 0 {
			t.Errorf("byte at index %d is %d, expected 0", i, b)
		}
	}
}

// TestEncryptProducesDifferentCiphertextEachTime verifies that successive
// encryptions of the same plaintext produce different ciphertext (due to
// random IV and random DEK). This is a critical security property.
func TestEncryptProducesDifferentCiphertextEachTime(t *testing.T) {
	kms := NewMockKMS()
	encryptor := NewEnvelopeEncryptor(kms, "test-key")

	plaintext := []byte("same-data")

	enc1, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("first Encrypt failed: %v", err)
	}

	enc2, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("second Encrypt failed: %v", err)
	}

	if bytes.Equal(enc1.Ciphertext, enc2.Ciphertext) {
		t.Error("two encryptions of the same plaintext produced identical ciphertext — IV or DEK randomness failure")
	}
}

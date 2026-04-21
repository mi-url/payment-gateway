package bnc

import (
	"testing"
)

// TestEncryptDecryptRoundtrip verifies that encrypting and decrypting a
// payload produces the original plaintext. This is the most critical test
// for the BNC integration — if encryption is wrong, all transactions fail.
func TestEncryptDecryptRoundtrip(t *testing.T) {
	testCases := []struct {
		name string
		key  string
		text string
	}{
		{
			name: "simple JSON payload",
			key:  "TestMasterKey123456",
			text: `{"ClientGUID":"4A074C46-DD4E-4E54-8010-B80A6A8758F4"}`,
		},
		{
			name: "C2P payload with amount",
			key:  "WorkingKey9876543210ABCDEF",
			text: `{"DebtorBankCode":"0191","DebtorCellPhone":"584141234567","DebtorID":"V12345678","Amount":250.00,"Token":"A4H9B2","Terminal":"TERM001"}`,
		},
		{
			name: "empty object",
			key:  "AnotherTestKey",
			text: `{}`,
		},
		{
			name: "payload with special characters",
			key:  "KeyWithSpecial!@#",
			text: `{"Description":"Pago de reserva — cancha #3, 8:00 PM"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			crypto := NewCrypto(tc.key)

			encrypted, err := crypto.Encrypt(tc.text)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			if encrypted == "" {
				t.Fatal("Encrypt returned empty string")
			}

			decrypted, err := crypto.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if decrypted != tc.text {
				t.Errorf("roundtrip mismatch:\n  got:  %q\n  want: %q", decrypted, tc.text)
			}
		})
	}
}

// TestHashSHA256 verifies the SHA-256 hash function used for the Validation field.
func TestHashSHA256(t *testing.T) {
	// Known SHA-256 hash of "test" (verified externally).
	input := "test"
	expected := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"

	result := HashSHA256(input)
	if result != expected {
		t.Errorf("HashSHA256(%q) = %q, want %q", input, result, expected)
	}
}

// TestEncodeUTF16LE verifies that UTF-8 text is correctly encoded to UTF-16LE.
// BNC requires UTF-16LE encoding before AES encryption.
func TestEncodeUTF16LE(t *testing.T) {
	// "A" in UTF-16LE is [0x41, 0x00]
	result := encodeUTF16LE("A")
	if len(result) != 2 || result[0] != 0x41 || result[1] != 0x00 {
		t.Errorf("encodeUTF16LE(\"A\") = %v, want [0x41, 0x00]", result)
	}

	// "AB" in UTF-16LE is [0x41, 0x00, 0x42, 0x00]
	result = encodeUTF16LE("AB")
	if len(result) != 4 {
		t.Errorf("encodeUTF16LE(\"AB\") length = %d, want 4", len(result))
	}
}

// TestPKCS7Padding verifies PKCS#7 padding and unpadding.
func TestPKCS7Padding(t *testing.T) {
	original := []byte("hello")
	padded := pkcs7Pad(original, 16)

	// Padded length should be a multiple of 16.
	if len(padded)%16 != 0 {
		t.Errorf("padded length %d is not a multiple of 16", len(padded))
	}

	unpadded, err := pkcs7Unpad(padded, 16)
	if err != nil {
		t.Fatalf("pkcs7Unpad failed: %v", err)
	}

	if string(unpadded) != string(original) {
		t.Errorf("roundtrip mismatch: got %q, want %q", unpadded, original)
	}
}

// TestDecryptInvalidInput verifies that decryption fails gracefully on bad input.
func TestDecryptInvalidInput(t *testing.T) {
	crypto := NewCrypto("test-key")

	_, err := crypto.Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64, got nil")
	}

	_, err = crypto.Decrypt("")
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

// TestDifferentKeysProduceDifferentCiphertext verifies that the same plaintext
// encrypted with different keys produces different ciphertext.
func TestDifferentKeysProduceDifferentCiphertext(t *testing.T) {
	plaintext := `{"test":"data"}`
	crypto1 := NewCrypto("key-one")
	crypto2 := NewCrypto("key-two")

	enc1, err := crypto1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt with key1 failed: %v", err)
	}

	enc2, err := crypto2.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt with key2 failed: %v", err)
	}

	if enc1 == enc2 {
		t.Error("different keys produced identical ciphertext — this is a security issue")
	}
}

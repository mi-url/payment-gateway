// Package bnc implements the bank adapter for Banco Nacional de Crédito (BNC).
// Bank code: 0191.
//
// BNC uses a custom encryption scheme for all API requests:
//   - PBKDF2 key derivation with SHA-1 (1000 iterations, fixed salt)
//   - AES-CBC encryption with the derived key
//   - Text must be encoded as UTF-16LE before encryption
//   - SHA-256 hash of the plaintext for integrity validation
//
// Reference: BNC ESolutions API v4.1 documentation.
package bnc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"unicode/utf16"

	"golang.org/x/crypto/pbkdf2"
)

// bncSalt is the fixed salt used by BNC for PBKDF2 key derivation.
// It represents the ASCII string "Ivan Medvedev" and is hardcoded
// in the BNC SDK documentation. It must not be changed.
var bncSalt = []byte{0x49, 0x76, 0x61, 0x6e, 0x20, 0x4d, 0x65, 0x64, 0x76, 0x65, 0x64, 0x65, 0x76}

const (
	pbkdf2Iterations = 1000
	pbkdf2KeyLen     = 48 // 32 bytes key + 16 bytes IV
)

// Crypto handles BNC-specific encryption and decryption operations.
// It derives the AES key and IV from the provided encryption key using PBKDF2.
type Crypto struct {
	key []byte // 32-byte AES-256 key
	iv  []byte // 16-byte initialization vector
}

// NewCrypto creates a Crypto instance with key material derived from encryptionKey.
// For Logon requests, encryptionKey is the MasterKey.
// For all other requests, encryptionKey is the daily WorkingKey.
func NewCrypto(encryptionKey string) *Crypto {
	derived := pbkdf2.Key([]byte(encryptionKey), bncSalt, pbkdf2Iterations, pbkdf2KeyLen, sha1.New)
	return &Crypto{
		key: derived[:32],
		iv:  derived[32:48],
	}
}

// Encrypt takes a plaintext JSON string, encodes it as UTF-16LE, encrypts it
// with AES-CBC, and returns the Base64-encoded ciphertext (the "Value" field).
func (c *Crypto) Encrypt(plaintext string) (string, error) {
	utf16LEBytes := encodeUTF16LE(plaintext)
	padded := pkcs7Pad(utf16LEBytes, aes.BlockSize)

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("bnc crypto: failed to create AES cipher: %w", err)
	}

	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, c.iv)
	mode.CryptBlocks(ciphertext, padded)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt takes a Base64-encoded ciphertext (from the "value" response field),
// decrypts it with AES-CBC, removes padding, and returns the UTF-8 plaintext.
func (c *Crypto) Decrypt(encoded string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("bnc crypto: failed to decode base64: %w", err)
	}

	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("bnc crypto: ciphertext length %d is not a multiple of block size %d", len(ciphertext), aes.BlockSize)
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("bnc crypto: failed to create AES cipher: %w", err)
	}

	plainBytes := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, c.iv)
	mode.CryptBlocks(plainBytes, ciphertext)

	unpadded, err := pkcs7Unpad(plainBytes, aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf("bnc crypto: %w", err)
	}

	return decodeUTF16LE(unpadded), nil
}

// HashSHA256 computes the SHA-256 hex digest of plaintext.
// This is the "Validation" field in BNC request envelopes.
func HashSHA256(plaintext string) string {
	h := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(h[:])
}

// encodeUTF16LE converts a UTF-8 string to UTF-16LE byte representation.
// BNC requires this encoding before AES encryption.
func encodeUTF16LE(s string) []byte {
	runes := utf16.Encode([]rune(s))
	buf := make([]byte, len(runes)*2)
	for i, r := range runes {
		binary.LittleEndian.PutUint16(buf[i*2:], r)
	}
	return buf
}

// decodeUTF16LE converts UTF-16LE bytes back to a UTF-8 string.
func decodeUTF16LE(b []byte) string {
	if len(b)%2 != 0 {
		b = b[:len(b)-1] // Drop incomplete code unit.
	}
	u16 := make([]uint16, len(b)/2)
	for i := range u16 {
		u16[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	return string(utf16.Decode(u16))
}

// pkcs7Pad adds PKCS#7 padding to reach a multiple of blockSize.
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padBytes := make([]byte, padding)
	for i := range padBytes {
		padBytes[i] = byte(padding)
	}
	return append(data, padBytes...)
}

// pkcs7Unpad removes PKCS#7 padding and validates it.
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("pkcs7: empty data")
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize || padding > len(data) {
		return nil, fmt.Errorf("pkcs7: invalid padding value %d", padding)
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("pkcs7: inconsistent padding bytes")
		}
	}
	return data[:len(data)-padding], nil
}

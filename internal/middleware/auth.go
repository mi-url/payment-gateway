// Package middleware provides HTTP middleware for the payment gateway.
package middleware

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
)

// contextKey is an unexported type for context keys in this package.
type contextKey string

// MerchantIDKey is the context key for the authenticated merchant ID.
const MerchantIDKey contextKey = "merchant_id"

// MerchantLookupFunc retrieves a merchant ID by the SHA-256 hash of their API key.
// Returns the merchant ID and true if found, or empty string and false if not.
type MerchantLookupFunc func(ctx context.Context, apiKeyHash string) (string, bool)

// Auth validates the API key from the Authorization header and injects the
// merchant ID into the request context. Returns 401 if the key is missing
// or invalid. The API key is never stored — only its SHA-256 hash is compared.
func Auth(lookup MerchantLookupFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
				return
			}

			apiKey := strings.TrimPrefix(authHeader, "Bearer ")
			if apiKey == authHeader {
				http.Error(w, `{"error":"Authorization header must use Bearer scheme"}`, http.StatusUnauthorized)
				return
			}

			hash := hashAPIKey(apiKey)
			merchantID, found := lookup(r.Context(), hash)
			if !found {
				http.Error(w, `{"error":"invalid API key"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), MerchantIDKey, merchantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// hashAPIKey computes the SHA-256 hash of an API key.
// This is the only form in which API keys exist in the database.
func hashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// ConstantTimeCompare performs a constant-time comparison of two strings
// to prevent timing attacks on API key validation.
func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

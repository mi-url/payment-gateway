package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAuthMissingHeader verifies that requests without Authorization return 401.
func TestAuthMissingHeader(t *testing.T) {
	handler := Auth(func(ctx context.Context, hash string) (string, bool) {
		return "", false
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without auth")
	}))

	req := httptest.NewRequest("POST", "/v1/charges/c2p", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

// TestAuthInvalidScheme verifies that non-Bearer auth schemes are rejected.
func TestAuthInvalidScheme(t *testing.T) {
	handler := Auth(func(ctx context.Context, hash string) (string, bool) {
		return "", false
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid scheme")
	}))

	req := httptest.NewRequest("POST", "/v1/charges/c2p", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

// TestAuthValidKey verifies that a valid API key passes through and sets merchant context.
func TestAuthValidKey(t *testing.T) {
	handler := Auth(func(ctx context.Context, hash string) (string, bool) {
		return "merchant-uuid-123", true
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		merchantID, ok := r.Context().Value(MerchantIDKey).(string)
		if !ok || merchantID != "merchant-uuid-123" {
			t.Errorf("merchant ID = %q, want merchant-uuid-123", merchantID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/v1/charges/c2p", nil)
	req.Header.Set("Authorization", "Bearer test-api-key-12345")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

// TestAuthInvalidKey verifies that an unknown API key returns 401.
func TestAuthInvalidKey(t *testing.T) {
	handler := Auth(func(ctx context.Context, hash string) (string, bool) {
		return "", false
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid key")
	}))

	req := httptest.NewRequest("POST", "/v1/charges/c2p", nil)
	req.Header.Set("Authorization", "Bearer invalid-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

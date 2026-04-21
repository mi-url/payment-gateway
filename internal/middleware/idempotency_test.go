package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestIdempotencyMissingKey verifies that requests without Idempotency-Key return 400.
func TestIdempotencyMissingKey(t *testing.T) {
	store := NewMemoryIdempotencyStore()
	handler := Idempotency(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without idempotency key")
	}))

	req := httptest.NewRequest("POST", "/v1/charges/c2p", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

// TestIdempotencyFirstRequest verifies that the first request passes through normally.
func TestIdempotencyFirstRequest(t *testing.T) {
	store := NewMemoryIdempotencyStore()
	called := false
	handler := Idempotency(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))

	req := httptest.NewRequest("POST", "/v1/charges/c2p", nil)
	req.Header.Set("Idempotency-Key", "unique-key-1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("handler should be called on first request")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

// TestIdempotencyDuplicateRequest verifies that a duplicate request returns the cached response.
func TestIdempotencyDuplicateRequest(t *testing.T) {
	store := NewMemoryIdempotencyStore()
	callCount := 0
	handler := Idempotency(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"txn-123","status":"SUCCESS"}`))
	}))

	// First request.
	req1 := httptest.NewRequest("POST", "/v1/charges/c2p", nil)
	req1.Header.Set("Idempotency-Key", "dup-key")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	// Second request with same key.
	req2 := httptest.NewRequest("POST", "/v1/charges/c2p", nil)
	req2.Header.Set("Idempotency-Key", "dup-key")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if callCount != 1 {
		t.Errorf("handler called %d times, want 1 (duplicate should be cached)", callCount)
	}

	if rec2.Header().Get("X-Idempotent-Replay") != "true" {
		t.Error("duplicate response should have X-Idempotent-Replay: true header")
	}

	if rec2.Body.String() != rec1.Body.String() {
		t.Errorf("duplicate response body mismatch:\n  first:  %s\n  second: %s", rec1.Body.String(), rec2.Body.String())
	}
}

// TestIdempotencyDifferentKeys verifies that different keys are processed independently.
func TestIdempotencyDifferentKeys(t *testing.T) {
	store := NewMemoryIdempotencyStore()
	callCount := 0
	handler := Idempotency(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))

	for _, key := range []string{"key-1", "key-2", "key-3"} {
		req := httptest.NewRequest("POST", "/v1/charges/c2p", nil)
		req.Header.Set("Idempotency-Key", key)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	if callCount != 3 {
		t.Errorf("handler called %d times, want 3 (different keys)", callCount)
	}
}

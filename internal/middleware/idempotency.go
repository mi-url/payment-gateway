// Package middleware provides HTTP middleware for the payment gateway.
package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
)

// IdempotencyStore defines how idempotency results are persisted.
// In production this is backed by PostgreSQL; in tests, by an in-memory map.
type IdempotencyStore interface {
	// Get retrieves a cached response for the given merchant + idempotency key.
	// Returns nil if no cached response exists.
	Get(merchantID, idempotencyKey string) *CachedResponse

	// Set stores a response for the given merchant + idempotency key.
	Set(merchantID, idempotencyKey string, response *CachedResponse)
}

// CachedResponse holds a previously returned response for an idempotent request.
type CachedResponse struct {
	StatusCode int
	Body       []byte
}

// Idempotency prevents duplicate charge processing by checking if a request
// with the same idempotency_key has already been processed for a merchant.
// If a cached response exists, it is returned immediately without re-processing.
func Idempotency(store IdempotencyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			merchantID, _ := r.Context().Value(MerchantIDKey).(string)
			idempotencyKey := r.Header.Get("Idempotency-Key")

			if idempotencyKey == "" {
				http.Error(w, `{"error":"Idempotency-Key header is required"}`, http.StatusBadRequest)
				return
			}

			// Check for cached response.
			if cached := store.Get(merchantID, idempotencyKey); cached != nil {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Idempotent-Replay", "true")
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
				return
			}

			// Capture the response for caching.
			recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(recorder, r)

			// Cache the response for future duplicate requests.
			store.Set(merchantID, idempotencyKey, &CachedResponse{
				StatusCode: recorder.statusCode,
				Body:       recorder.body,
			})
		})
	}
}

// responseRecorder captures the response body and status code for caching.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}

// MemoryIdempotencyStore is an in-memory implementation for development and testing.
// Production uses PostgreSQL via the store package.
type MemoryIdempotencyStore struct {
	mu    sync.RWMutex
	cache map[string]*CachedResponse
}

// NewMemoryIdempotencyStore creates an in-memory idempotency store.
func NewMemoryIdempotencyStore() *MemoryIdempotencyStore {
	return &MemoryIdempotencyStore{
		cache: make(map[string]*CachedResponse),
	}
}

func (s *MemoryIdempotencyStore) key(merchantID, idempotencyKey string) string {
	b, _ := json.Marshal([]string{merchantID, idempotencyKey})
	return string(b)
}

// Get retrieves a cached response.
func (s *MemoryIdempotencyStore) Get(merchantID, idempotencyKey string) *CachedResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache[s.key(merchantID, idempotencyKey)]
}

// Set stores a response for future duplicate detection.
func (s *MemoryIdempotencyStore) Set(merchantID, idempotencyKey string, response *CachedResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[s.key(merchantID, idempotencyKey)] = response
}

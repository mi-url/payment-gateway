// Package middleware provides HTTP middleware for the payment gateway.
package middleware

import (
	"net/http"
	"sync"
	"time"
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

// idempotencyEntry holds a cached response and its creation timestamp for TTL eviction.
type idempotencyEntry struct {
	response  *CachedResponse
	createdAt time.Time
}

// MemoryIdempotencyStore is an in-memory implementation for development and testing.
// Production uses PostgreSQL via the store package.
// Entries are automatically evicted after 24 hours to prevent unbounded memory growth.
type MemoryIdempotencyStore struct {
	mu    sync.RWMutex
	cache map[string]*idempotencyEntry
}

// NewMemoryIdempotencyStore creates an in-memory idempotency store
// with automatic background cleanup of entries older than 24 hours.
func NewMemoryIdempotencyStore() *MemoryIdempotencyStore {
	s := &MemoryIdempotencyStore{
		cache: make(map[string]*idempotencyEntry),
	}
	go s.cleanupLoop()
	return s
}

func (s *MemoryIdempotencyStore) compositeKey(merchantID, idempotencyKey string) string {
	return merchantID + ":" + idempotencyKey
}

// Get retrieves a cached response.
func (s *MemoryIdempotencyStore) Get(merchantID, idempotencyKey string) *CachedResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry := s.cache[s.compositeKey(merchantID, idempotencyKey)]
	if entry == nil {
		return nil
	}
	return entry.response
}

// Set stores a response for future duplicate detection.
func (s *MemoryIdempotencyStore) Set(merchantID, idempotencyKey string, response *CachedResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[s.compositeKey(merchantID, idempotencyKey)] = &idempotencyEntry{
		response:  response,
		createdAt: time.Now(),
	}
}

// cleanupLoop evicts entries older than 24 hours every hour.
func (s *MemoryIdempotencyStore) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		cutoff := time.Now().Add(-24 * time.Hour)
		for key, entry := range s.cache {
			if entry.createdAt.Before(cutoff) {
				delete(s.cache, key)
			}
		}
		s.mu.Unlock()
	}
}

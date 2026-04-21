// Package middleware provides HTTP middleware for the payment gateway.
package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimit applies a simple per-IP token bucket rate limiter.
// Limits each IP to maxRequests per window duration.
// Returns 429 Too Many Requests when the limit is exceeded.
func RateLimit(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	type client struct {
		tokens    int
		lastReset time.Time
	}

	var mu sync.Mutex
	clients := make(map[string]*client)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			mu.Lock()
			c, exists := clients[ip]
			if !exists {
				c = &client{tokens: maxRequests, lastReset: time.Now()}
				clients[ip] = c
			}

			// Reset tokens if the window has elapsed.
			if time.Since(c.lastReset) > window {
				c.tokens = maxRequests
				c.lastReset = time.Now()
			}

			if c.tokens <= 0 {
				mu.Unlock()
				w.Header().Set("Retry-After", window.String())
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}

			c.tokens--
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

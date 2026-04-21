// Package bnc implements the bank adapter for Banco Nacional de Crédito (BNC).
package bnc

import (
	"fmt"
	"sync"
	"time"
)

// WorkingKeyManager caches BNC daily WorkingKeys per merchant.
// The WorkingKey is obtained via the Logon endpoint using the MasterKey,
// and expires at midnight (Caracas time, UTC-4) or when BNC returns EPIRWK.
type WorkingKeyManager struct {
	mu    sync.RWMutex
	cache map[string]*cachedKey // keyed by merchant ClientGUID
}

// cachedKey holds a WorkingKey and its expiration timestamp.
type cachedKey struct {
	key       string
	expiresAt time.Time
}

// NewWorkingKeyManager creates an empty WorkingKey cache.
func NewWorkingKeyManager() *WorkingKeyManager {
	return &WorkingKeyManager{
		cache: make(map[string]*cachedKey),
	}
}

// Get retrieves the cached WorkingKey for a merchant's ClientGUID.
// Returns the key and true if valid, or empty string and false if expired/missing.
func (m *WorkingKeyManager) Get(clientGUID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.cache[clientGUID]
	if !exists || time.Now().After(entry.expiresAt) {
		return "", false
	}
	return entry.key, true
}

// Set stores a WorkingKey for a merchant's ClientGUID.
// The key expires at the next midnight in Caracas time (UTC-4).
func (m *WorkingKeyManager) Set(clientGUID, workingKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache[clientGUID] = &cachedKey{
		key:       workingKey,
		expiresAt: nextMidnightCaracas(),
	}
}

// Invalidate removes the cached WorkingKey for a merchant's ClientGUID.
// Called when BNC returns the EPIRWK error code, forcing re-authentication.
func (m *WorkingKeyManager) Invalidate(clientGUID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.cache, clientGUID)
}

// nextMidnightCaracas returns the next midnight in Venezuela time (UTC-4).
func nextMidnightCaracas() time.Time {
	caracas := time.FixedZone("VET", -4*60*60)
	now := time.Now().In(caracas)
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, caracas)
	return midnight
}

// FormatReference generates a unique daily reference for BNC transactions.
// BNC requires that the Reference field is unique per day per ClientGUID.
func FormatReference(txnID string) string {
	// Use the first 20 chars of the transaction UUID (without dashes) for uniqueness.
	clean := ""
	for _, c := range txnID {
		if c != '-' {
			clean += string(c)
		}
	}
	if len(clean) > 20 {
		clean = clean[:20]
	}
	return fmt.Sprintf("GW%s", clean)
}

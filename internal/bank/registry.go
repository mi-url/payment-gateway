// Package bank defines the adapter interface for bank integrations.
package bank

import (
	"fmt"
	"sync"
)

// Registry maps bank codes to their adapter implementations.
// It is safe for concurrent reads after initialization.
type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

// NewRegistry creates an empty adapter registry.
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[string]Adapter),
	}
}

// Register adds a bank adapter to the registry.
// Panics if the bank code is already registered (configuration error).
func (r *Registry) Register(adapter Adapter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	code := adapter.BankCode()
	if _, exists := r.adapters[code]; exists {
		panic(fmt.Sprintf("bank adapter already registered for code: %s", code))
	}
	r.adapters[code] = adapter
}

// Get returns the adapter for the given bank code, or an error if not found.
func (r *Registry) Get(bankCode string) (Adapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, exists := r.adapters[bankCode]
	if !exists {
		return nil, fmt.Errorf("no adapter registered for bank code: %s", bankCode)
	}
	return adapter, nil
}

// RegisteredBanks returns a list of all registered bank codes.
func (r *Registry) RegisteredBanks() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	codes := make([]string, 0, len(r.adapters))
	for code := range r.adapters {
		codes = append(codes, code)
	}
	return codes
}

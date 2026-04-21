package bnc

import (
	"testing"
	"time"
)

// TestWorkingKeySetAndGet verifies basic cache storage and retrieval.
func TestWorkingKeySetAndGet(t *testing.T) {
	mgr := NewWorkingKeyManager()

	mgr.Set("client-abc", "working-key-123")

	key, ok := mgr.Get("client-abc")
	if !ok {
		t.Fatal("expected key to be found")
	}
	if key != "working-key-123" {
		t.Errorf("got %q, want %q", key, "working-key-123")
	}
}

// TestWorkingKeyMiss verifies that a missing key returns false.
func TestWorkingKeyMiss(t *testing.T) {
	mgr := NewWorkingKeyManager()

	_, ok := mgr.Get("nonexistent")
	if ok {
		t.Error("expected key to not be found")
	}
}

// TestWorkingKeyInvalidate verifies that invalidation removes the key.
func TestWorkingKeyInvalidate(t *testing.T) {
	mgr := NewWorkingKeyManager()

	mgr.Set("client-abc", "working-key-123")
	mgr.Invalidate("client-abc")

	_, ok := mgr.Get("client-abc")
	if ok {
		t.Error("expected key to be invalidated")
	}
}

// TestWorkingKeyDifferentClients verifies key isolation between merchants.
func TestWorkingKeyDifferentClients(t *testing.T) {
	mgr := NewWorkingKeyManager()

	mgr.Set("client-a", "key-a")
	mgr.Set("client-b", "key-b")

	keyA, _ := mgr.Get("client-a")
	keyB, _ := mgr.Get("client-b")

	if keyA == keyB {
		t.Error("different clients should have different keys")
	}
}

// TestFormatReference verifies reference generation format.
func TestFormatReference(t *testing.T) {
	ref := FormatReference("550e8400-e29b-41d4-a716-446655440000")
	if ref[:2] != "GW" {
		t.Errorf("reference should start with GW, got %q", ref)
	}
	if len(ref) > 22 {
		t.Errorf("reference too long: %d chars", len(ref))
	}
}

// TestNextMidnightCaracas verifies the midnight calculation is in the future.
func TestNextMidnightCaracas(t *testing.T) {
	midnight := nextMidnightCaracas()
	if midnight.Before(time.Now()) {
		t.Error("next midnight should be in the future")
	}
}

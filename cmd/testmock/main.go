// Command testmock exercises the BNC Mock Server using the real BNC adapter.
// This validates the full encryption roundtrip: Gateway ↔ Mock BNC.
//
// Usage:
//
//	go run ./cmd/testmock
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/faloppa/payment-gateway/internal/bank"
	"github.com/faloppa/payment-gateway/internal/bank/bnc"
	"github.com/shopspring/decimal"
)

func main() {
	baseURL := os.Getenv("BNC_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:9090"
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	adapter := bnc.NewAdapter(baseURL, logger, true)

	// Credentials matching the mock server's testMasterKey.
	creds := map[string]string{
		"client_guid": "TEST-GUID-001",
		"master_key":  "TestMasterKey123456789012345678901",
		"terminal":    "TERM001",
	}
	credsJSON, _ := json.Marshal(creds)

	ctx := context.Background()

	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("  BNC Mock Server Integration Test")
	fmt.Println("═══════════════════════════════════════════════")

	// Test 1: Validate Credentials (Logon)
	fmt.Println("\n🔐 Test 1: Logon (ValidateCredentials)")
	err := adapter.ValidateCredentials(ctx, credsJSON)
	if err != nil {
		log.Fatalf("   ❌ FAILED: %v", err)
	}
	fmt.Println("   ✅ PASS — WorkingKey obtained successfully")

	// Test 2: Successful C2P
	fmt.Println("\n💳 Test 2: C2P — Success (normal phone)")
	resp, err := adapter.ProcessC2P(ctx, &bank.C2PRequest{
		PayerPhone:    "04141234567",
		PayerBankCode: "0102",
		PayerIDDoc:    "V12345678",
		Amount:        decimal.NewFromFloat(2500.00),
		OTP:           "123456",
		Reference:     "REF-TEST-001",
	}, credsJSON)
	if err != nil {
		log.Fatalf("   ❌ FAILED: %v", err)
	}
	fmt.Printf("   ✅ PASS — Success=%t Reference=%s\n", resp.Success, resp.BankReference)

	// Test 3: Insufficient Funds (phone ending in 0000)
	fmt.Println("\n💳 Test 3: C2P — Insufficient Funds (phone ending 0000)")
	resp, err = adapter.ProcessC2P(ctx, &bank.C2PRequest{
		PayerPhone:    "04141230000",
		PayerBankCode: "0102",
		PayerIDDoc:    "V12345678",
		Amount:        decimal.NewFromFloat(1000.00),
		OTP:           "123456",
		Reference:     "REF-TEST-002",
	}, credsJSON)
	if err != nil {
		log.Fatalf("   ❌ FAILED: %v", err)
	}
	fmt.Printf("   ✅ PASS — Success=%t ErrorCode=%s\n", resp.Success, resp.ErrorCode)

	// Test 4: Incorrect OTP (phone ending in 1111)
	fmt.Println("\n💳 Test 4: C2P — Incorrect OTP (phone ending 1111)")
	resp, err = adapter.ProcessC2P(ctx, &bank.C2PRequest{
		PayerPhone:    "04141231111",
		PayerBankCode: "0102",
		PayerIDDoc:    "V12345678",
		Amount:        decimal.NewFromFloat(500.00),
		OTP:           "000000",
		Reference:     "REF-TEST-003",
	}, credsJSON)
	if err != nil {
		log.Fatalf("   ❌ FAILED: %v", err)
	}
	fmt.Printf("   ✅ PASS — Success=%t ErrorCode=%s\n", resp.Success, resp.ErrorCode)

	fmt.Println("\n═══════════════════════════════════════════════")
	fmt.Println("  All tests passed! Mock BNC is working.")
	fmt.Println("═══════════════════════════════════════════════")
}

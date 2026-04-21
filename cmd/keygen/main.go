// Command keygen generates API keys for merchants.
// Usage: go run ./cmd/keygen -name "Mi Empresa"
//
// It outputs:
//   - The raw API key (give this to the merchant, ONCE)
//   - The SHA-256 hash (store this in the database)
//   - An SQL INSERT statement ready to run in Supabase
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
)

func main() {
	name := flag.String("name", "", "Merchant name (required)")
	flag.Parse()

	if *name == "" {
		fmt.Fprintln(os.Stderr, "Error: -name flag is required")
		fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/keygen -name \"Mi Empresa\"")
		os.Exit(1)
	}

	// Generate a 32-byte random API key.
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating key: %v\n", err)
		os.Exit(1)
	}

	apiKey := "gw_live_" + hex.EncodeToString(keyBytes)
	hash := sha256.Sum256([]byte(apiKey))
	hashHex := hex.EncodeToString(hash[:])

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  PAYMENT GATEWAY — API KEY GENERATED")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  Merchant:  %s\n", *name)
	fmt.Println()
	fmt.Println("  ⚠️  SAVE THIS KEY — IT WILL NOT BE SHOWN AGAIN:")
	fmt.Printf("  API Key:   %s\n", apiKey)
	fmt.Println()
	fmt.Printf("  SHA-256:   %s\n", hashHex)
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("  SQL (run in Supabase SQL Editor):")
	fmt.Println()
	fmt.Printf("  INSERT INTO merchants (name, api_key_hash)\n")
	fmt.Printf("  VALUES ('%s', '%s');\n", *name, hashHex)
	fmt.Println()
}

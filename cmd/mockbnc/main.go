// Command mockbnc runs a local server that simulates BNC's ESolutions API.
//
// It implements the same envelope encryption protocol (PBKDF2 + AES-CBC + UTF-16LE)
// so the real BNC adapter can be tested end-to-end without bank credentials.
//
// Usage:
//
//	go run ./cmd/mockbnc
//
// The server starts on :9090 by default and responds to:
//   - POST /Auth/LogOn           → Returns an encrypted WorkingKey
//   - POST /MobPayment/SendC2P   → Simulates a C2P charge (success/failure)
//   - POST /Validation/TransactionQuery → Simulates a transaction query
//   - POST /webhook/trigger      → Sends a webhook notification to the gateway
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf16"

	"golang.org/x/crypto/pbkdf2"
)

// ----- BNC Crypto (mirrors internal/bank/bnc/crypto.go) -----

var bncSalt = []byte{0x49, 0x76, 0x61, 0x6e, 0x20, 0x4d, 0x65, 0x64, 0x76, 0x65, 0x64, 0x65, 0x76}

type bncCrypto struct {
	key []byte
	iv  []byte
}

func newBNCCrypto(encryptionKey string) *bncCrypto {
	derived := pbkdf2.Key([]byte(encryptionKey), bncSalt, 1000, 48, sha1.New)
	return &bncCrypto{key: derived[:32], iv: derived[32:48]}
}

func (c *bncCrypto) encrypt(plaintext string) string {
	utf16LEBytes := encodeUTF16LE(plaintext)
	padded := pkcs7Pad(utf16LEBytes, aes.BlockSize)
	block, _ := aes.NewCipher(c.key)
	ct := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, c.iv).CryptBlocks(ct, padded)
	return base64.StdEncoding.EncodeToString(ct)
}

func (c *bncCrypto) decrypt(encoded string) (string, error) {
	ct, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	if len(ct) == 0 || len(ct)%aes.BlockSize != 0 {
		return "", fmt.Errorf("invalid ciphertext length %d", len(ct))
	}
	block, _ := aes.NewCipher(c.key)
	pt := make([]byte, len(ct))
	cipher.NewCBCDecrypter(block, c.iv).CryptBlocks(pt, ct)
	unpadded, err := pkcs7Unpad(pt, aes.BlockSize)
	if err != nil {
		return "", err
	}
	return decodeUTF16LE(unpadded), nil
}

func hashSHA256(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func encodeUTF16LE(s string) []byte {
	runes := utf16.Encode([]rune(s))
	buf := make([]byte, len(runes)*2)
	for i, r := range runes {
		binary.LittleEndian.PutUint16(buf[i*2:], r)
	}
	return buf
}

func decodeUTF16LE(b []byte) string {
	if len(b)%2 != 0 {
		b = b[:len(b)-1]
	}
	u16 := make([]uint16, len(b)/2)
	for i := range u16 {
		u16[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	return string(utf16.Decode(u16))
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padBytes := make([]byte, padding)
	for i := range padBytes {
		padBytes[i] = byte(padding)
	}
	return append(data, padBytes...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize || padding > len(data) {
		return nil, fmt.Errorf("invalid padding %d", padding)
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("inconsistent padding")
		}
	}
	return data[:len(data)-padding], nil
}

// ----- Mock Server State -----

const (
	// testMasterKey is the MasterKey used for dev/test.
	// Matches the value in supabase/seed.sql.
	testMasterKey = "TestMasterKey123456789012345678901"

	// testWorkingKey is the daily WorkingKey returned by the mock Logon.
	testWorkingKey = "MockWorkingKey12345678901234567890"
)

// Envelope is the BNC request format.
type Envelope struct {
	ClientGUID      string `json:"ClientGUID"`
	Reference       string `json:"Reference"`
	Value           string `json:"Value"`
	Validation      string `json:"Validation"`
	SwTestOperation bool   `json:"swTestOperation"`
}

// EnvelopeResponse is the BNC response format.
type EnvelopeResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	Value      string `json:"value"`
	Validation string `json:"validation"`
}

func main() {
	port := os.Getenv("MOCK_PORT")
	if port == "" {
		port = "9090"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/Auth/LogOn", handleLogon)
	mux.HandleFunc("/MobPayment/SendC2P", handleC2P)
	mux.HandleFunc("/Validation/TransactionQuery", handleQuery)
	mux.HandleFunc("/webhook/trigger", handleTriggerWebhook)

	log.Printf("🏦 BNC Mock Server running on :%s", port)
	log.Printf("   Master Key: %s", testMasterKey)
	log.Printf("   Working Key: %s (rotates on Logon)", testWorkingKey)
	log.Printf("")
	log.Printf("   POST /Auth/LogOn              → Returns WorkingKey")
	log.Printf("   POST /MobPayment/SendC2P      → Simulates C2P charge")
	log.Printf("   POST /Validation/TransactionQuery → Simulates query")
	log.Printf("   POST /webhook/trigger         → Sends webhook to gateway")

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Failed to start mock server: %v", err)
	}
}

// handleLogon simulates BNC's Logon endpoint.
// Decrypts the request with the MasterKey, returns an encrypted WorkingKey.
func handleLogon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	env, err := readEnvelope(r)
	if err != nil {
		writeError(w, "000001 Invalid request: "+err.Error())
		return
	}

	// Decrypt with MasterKey to validate the request.
	crypto := newBNCCrypto(testMasterKey)
	plaintext, err := crypto.decrypt(env.Value)
	if err != nil {
		writeError(w, "000002 Decryption failed: "+err.Error())
		return
	}

	log.Printf("✅ LogOn from ClientGUID=%s | Decrypted: %s", env.ClientGUID, plaintext)

	// Respond with encrypted WorkingKey.
	responseJSON := fmt.Sprintf(`{"WorkingKey":"%s"}`, testWorkingKey)
	encrypted := crypto.encrypt(responseJSON)

	writeSuccess(w, encrypted, hashSHA256(responseJSON))
}

// handleC2P simulates BNC's C2P charge endpoint.
// Supports special phone numbers to trigger different responses.
func handleC2P(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	env, err := readEnvelope(r)
	if err != nil {
		writeError(w, "000001 Invalid request: "+err.Error())
		return
	}

	crypto := newBNCCrypto(testWorkingKey)
	plaintext, err := crypto.decrypt(env.Value)
	if err != nil {
		writeError(w, "000003 WorkingKey decryption failed")
		return
	}

	// Parse the C2P payload.
	var payload struct {
		DebtorCellPhone string  `json:"DebtorCellPhone"`
		Amount          float64 `json:"Amount"`
		Token           string  `json:"Token"`
	}
	if err := json.Unmarshal([]byte(plaintext), &payload); err != nil {
		writeError(w, "000004 Invalid C2P payload")
		return
	}

	log.Printf("💳 C2P Request: Phone=%s Amount=%.2f OTP=%s Ref=%s",
		payload.DebtorCellPhone, payload.Amount, payload.Token, env.Reference)

	// Simulate different outcomes based on phone number suffix.
	phone := payload.DebtorCellPhone
	switch {
	case strings.HasSuffix(phone, "0000"): // Phone ending in 0000 → insufficient funds
		log.Printf("   ❌ Simulating INSUFFICIENT_FUNDS")
		writeError(w, "000051 Fondos insuficientes")
		return

	case strings.HasSuffix(phone, "1111"): // Phone ending in 1111 → incorrect OTP
		log.Printf("   ❌ Simulating INCORRECT_OTP")
		writeError(w, "000055 Clave temporal invalida")
		return

	case strings.HasSuffix(phone, "2222"): // Phone ending in 2222 → timeout
		log.Printf("   ⏳ Simulating TIMEOUT (10s delay)")
		time.Sleep(10 * time.Second)
		writeError(w, "000099 Timeout del sistema")
		return

	case strings.HasSuffix(phone, "3333"): // Phone ending in 3333 → expired working key
		log.Printf("   🔑 Simulating EXPIRED_WORKING_KEY")
		writeError(w, "000012 Working key expirada")
		return
	}

	// Default: SUCCESS
	ref := fmt.Sprintf("%012d", rand.Int63n(999999999999))
	responseJSON := fmt.Sprintf(`{"IdTransaction":"TXN%s","Reference":"%s"}`, ref[:8], ref)
	encrypted := crypto.encrypt(responseJSON)

	log.Printf("   ✅ Success → Reference=%s", ref)
	writeSuccess(w, encrypted, hashSHA256(responseJSON))
}

// handleQuery simulates BNC's TransactionQuery endpoint.
func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	env, err := readEnvelope(r)
	if err != nil {
		writeError(w, "000001 Invalid request: "+err.Error())
		return
	}

	crypto := newBNCCrypto(testWorkingKey)
	plaintext, err := crypto.decrypt(env.Value)
	if err != nil {
		writeError(w, "000003 Decryption failed")
		return
	}

	log.Printf("🔍 Query: %s", plaintext)

	// Always respond that the movement exists (for dev convenience).
	responseJSON := `{"MovementExists":true}`
	encrypted := crypto.encrypt(responseJSON)

	writeSuccess(w, encrypted, hashSHA256(responseJSON))
}

// handleTriggerWebhook is a convenience endpoint to manually fire a
// webhook notification to the gateway, simulating BNC NotificationPush.
func handleTriggerWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	gatewayURL := os.Getenv("GATEWAY_WEBHOOK_URL")
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8080/v1/webhooks/bnc"
	}

	notification := map[string]string{
		"PaymentType":          "P2P",
		"OriginBankReference":  fmt.Sprintf("%012d", rand.Int63n(999999999999)),
		"DestinyBankReference": fmt.Sprintf("%012d", rand.Int63n(999999999999)),
		"OriginBankCode":       "0102",
		"ClientID":             "V12345678",
		"Hour":                 time.Now().Format("15:04:05"),
		"CurrencyCode":         "VES",
		"Amount":               "2500.00",
		"Date":                 time.Now().Format("02/01/2006"),
		"CommerceID":           "J123456789",
		"CommercePhone":        "584141234567",
		"ClientPhone":          "584121234567",
		"Concept":              "Mock payment notification",
	}

	body, _ := json.Marshal(notification)
	resp, err := http.Post(gatewayURL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		log.Printf("❌ Failed to send webhook to %s: %v", gatewayURL, err)
		http.Error(w, "Failed to reach gateway: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	log.Printf("📨 Webhook sent to %s → Status: %d", gatewayURL, resp.StatusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"sent_to": gatewayURL,
		"status":  resp.StatusCode,
		"payload": notification,
	})
}

// ----- Helpers -----

func readEnvelope(r *http.Request) (*Envelope, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var env Envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

func writeSuccess(w http.ResponseWriter, encryptedValue, validation string) {
	resp := EnvelopeResponse{
		Status:     "OK",
		Message:    "000000 Successful",
		Value:      encryptedValue,
		Validation: validation,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, message string) {
	resp := EnvelopeResponse{
		Status:  "KO",
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

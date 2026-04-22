// Package bnc implements the bank adapter for Banco Nacional de Crédito (BNC).
package bnc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/faloppa/payment-gateway/internal/bank"
	"github.com/shopspring/decimal"
)

// Adapter implements bank.Adapter for Banco Nacional de Crédito (code 0191).
type Adapter struct {
	baseURL    string
	httpClient *http.Client
	keyMgr     *WorkingKeyManager
	logger     *slog.Logger
	testMode   bool // When true, sets swTestOperation=true for dry-run validation
}

// NewAdapter creates a BNC adapter configured for the given base URL.
func NewAdapter(baseURL string, logger *slog.Logger, testMode bool) *Adapter {
	return &Adapter{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		keyMgr:   NewWorkingKeyManager(),
		logger:   logger,
		testMode: testMode,
	}
}

// BankCode returns "0191" (BNC's SUDEBAN bank code).
func (a *Adapter) BankCode() string {
	return "0191"
}

// ProcessC2P executes a Cobro a Persona charge against BNC's API.
// It handles WorkingKey management, encryption, and error code mapping.
func (a *Adapter) ProcessC2P(ctx context.Context, req *bank.C2PRequest, credsJSON json.RawMessage) (*bank.Response, error) {
	var creds BNCCredentials
	if err := json.Unmarshal(credsJSON, &creds); err != nil {
		return nil, fmt.Errorf("bnc: failed to parse credentials: %w", err)
	}

	// Ensure we have a valid WorkingKey.
	workingKey, err := a.ensureWorkingKey(ctx, &creds)
	if err != nil {
		return nil, fmt.Errorf("bnc: failed to obtain working key: %w", err)
	}

	// Build the C2P payload.
	amount, _ := req.Amount.Float64()
	payload := C2PPayload{
		DebtorBankCode:  req.PayerBankCode,
		DebtorCellPhone: normalizePhone(req.PayerPhone),
		DebtorID:        req.PayerIDDoc,
		Amount:          amount,
		Token:           req.OTP,
		Terminal:        creds.Terminal,
		ChildClientID:   creds.ChildClientID,
		BranchID:        creds.BranchID,
	}

	// Encrypt and send.
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("bnc: failed to marshal C2P payload: %w", err)
	}

	resp, err := a.sendEncrypted(ctx, "/MobPayment/SendC2P", creds.ClientGUID, req.Reference, workingKey, string(payloadJSON))
	if err != nil {
		return &bank.Response{
			Success:      false,
			ErrorCode:    "NETWORK_ERROR",
			ErrorMessage: err.Error(),
		}, nil
	}

	// Check for WorkingKey expiration.
	if resp.Status == "KO" && len(resp.Message) >= 6 {
		code := resp.Message[:6]
		if IsWorkingKeyExpired(strings.TrimSpace(code)) {
			a.keyMgr.Invalidate(creds.ClientGUID)
			a.logger.Warn("bnc: working key expired, retrying with fresh key",
				slog.String("client_guid", creds.ClientGUID),
			)
			// Retry once with a fresh WorkingKey.
			return a.ProcessC2P(ctx, req, credsJSON)
		}
	}

	return a.parseC2PResponse(resp, workingKey)
}

// QueryTransaction checks whether a transaction was processed at BNC.
// Used by the reconciliation worker to resolve PENDING_RECONCILIATION transactions.
func (a *Adapter) QueryTransaction(ctx context.Context, ref string, amount decimal.Decimal, date time.Time, credsJSON json.RawMessage) (*bank.QueryResult, error) {
	var creds BNCCredentials
	if err := json.Unmarshal(credsJSON, &creds); err != nil {
		return nil, fmt.Errorf("bnc: failed to parse credentials: %w", err)
	}

	workingKey, err := a.ensureWorkingKey(ctx, &creds)
	if err != nil {
		return nil, fmt.Errorf("bnc: failed to obtain working key: %w", err)
	}

	amtF, _ := amount.Float64()
	payload := TransactionQueryPayload{
		Reference: ref,
		Amount:    amtF,
		Date:      date.Format("02/01/2006"),
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("bnc: failed to marshal query payload: %w", err)
	}

	resp, err := a.sendEncrypted(ctx, "/Validation/TransactionQuery", creds.ClientGUID, FormatReference(ref), workingKey, string(payloadJSON))
	if err != nil {
		return nil, fmt.Errorf("bnc: query request failed: %w", err)
	}

	if resp.Status != "OK" {
		return nil, fmt.Errorf("bnc: query returned status KO: %s", resp.Message)
	}

	crypto := NewCrypto(workingKey)
	decrypted, err := crypto.Decrypt(resp.Value)
	if err != nil {
		return nil, fmt.Errorf("bnc: failed to decrypt query response: %w", err)
	}

	var qr TransactionQueryResponse
	if err := json.Unmarshal([]byte(decrypted), &qr); err != nil {
		return nil, fmt.Errorf("bnc: failed to parse query response: %w", err)
	}

	return &bank.QueryResult{Exists: qr.MovementExists}, nil
}

// ValidateCredentials verifies BNC credentials by attempting a Logon.
func (a *Adapter) ValidateCredentials(ctx context.Context, credsJSON json.RawMessage) error {
	var creds BNCCredentials
	if err := json.Unmarshal(credsJSON, &creds); err != nil {
		return fmt.Errorf("bnc: failed to parse credentials: %w", err)
	}
	_, err := a.logon(ctx, &creds)
	return err
}

// ensureWorkingKey returns a valid WorkingKey, performing Logon if needed.
func (a *Adapter) ensureWorkingKey(ctx context.Context, creds *BNCCredentials) (string, error) {
	if key, ok := a.keyMgr.Get(creds.ClientGUID); ok {
		return key, nil
	}

	key, err := a.logon(ctx, creds)
	if err != nil {
		return "", err
	}

	a.keyMgr.Set(creds.ClientGUID, key)
	a.logger.Info("bnc: obtained new working key",
		slog.String("client_guid", creds.ClientGUID),
	)
	return key, nil
}

// logon authenticates with BNC to obtain a daily WorkingKey.
func (a *Adapter) logon(ctx context.Context, creds *BNCCredentials) (string, error) {
	payload := LogonPayload{ClientGUID: creds.ClientGUID}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("bnc logon: failed to marshal payload: %w", err)
	}

	resp, err := a.sendEncrypted(ctx, "/Auth/LogOn", creds.ClientGUID, FormatReference("logon"), creds.MasterKey, string(payloadJSON))
	if err != nil {
		return "", fmt.Errorf("bnc logon: request failed: %w", err)
	}

	if resp.Status != "OK" {
		return "", fmt.Errorf("bnc logon: failed with status KO: %s", resp.Message)
	}

	crypto := NewCrypto(creds.MasterKey)
	decrypted, err := crypto.Decrypt(resp.Value)
	if err != nil {
		return "", fmt.Errorf("bnc logon: failed to decrypt response: %w", err)
	}

	var logonResp LogonResponse
	if err := json.Unmarshal([]byte(decrypted), &logonResp); err != nil {
		return "", fmt.Errorf("bnc logon: failed to parse response: %w", err)
	}

	if logonResp.WorkingKey == "" {
		return "", fmt.Errorf("bnc logon: empty WorkingKey in response")
	}

	return logonResp.WorkingKey, nil
}

// sendEncrypted builds an encrypted BNC envelope and sends it to the specified endpoint.
func (a *Adapter) sendEncrypted(ctx context.Context, endpoint, clientGUID, reference, encryptionKey, payloadJSON string) (*EnvelopeResponse, error) {
	crypto := NewCrypto(encryptionKey)

	encrypted, err := crypto.Encrypt(payloadJSON)
	if err != nil {
		return nil, fmt.Errorf("bnc: encryption failed: %w", err)
	}

	envelope := Envelope{
		ClientGUID:      clientGUID,
		Reference:       reference,
		Value:           encrypted,
		Validation:      HashSHA256(payloadJSON),
		SwTestOperation: a.testMode,
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("bnc: failed to marshal envelope: %w", err)
	}

	url := a.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("bnc: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bnc: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("bnc: failed to read response body: %w", err)
	}

	var envResp EnvelopeResponse
	if err := json.Unmarshal(respBody, &envResp); err != nil {
		return nil, fmt.Errorf("bnc: failed to parse response (status %d): %w", resp.StatusCode, err)
	}

	// H3 FIX: Verify SHA-256 integrity of the response.
	// The Validation field is the SHA-256 hex digest of the decrypted Value.
	// If it doesn't match, the response may have been tampered with (MITM).
	if envResp.Status == "OK" && envResp.Value != "" && envResp.Validation != "" {
		decrypted, decErr := crypto.Decrypt(envResp.Value)
		if decErr == nil {
			expectedHash := HashSHA256(decrypted)
			if expectedHash != envResp.Validation {
				a.logger.Error("bnc: INTEGRITY CHECK FAILED — response Validation mismatch",
					slog.String("endpoint", endpoint),
					slog.String("expected", expectedHash),
					slog.String("received", envResp.Validation),
				)
				return nil, fmt.Errorf("bnc: response integrity check failed (SHA-256 mismatch)")
			}
		}
	}

	return &envResp, nil
}

// parseC2PResponse interprets the BNC C2P response and maps it to the gateway format.
func (a *Adapter) parseC2PResponse(resp *EnvelopeResponse, workingKey string) (*bank.Response, error) {
	if resp.Status == "OK" {
		crypto := NewCrypto(workingKey)
		decrypted, err := crypto.Decrypt(resp.Value)
		if err != nil {
			return nil, fmt.Errorf("bnc: failed to decrypt C2P response: %w", err)
		}

		var c2pResp C2PResponse
		if err := json.Unmarshal([]byte(decrypted), &c2pResp); err != nil {
			return nil, fmt.Errorf("bnc: failed to parse C2P response: %w", err)
		}

		return &bank.Response{
			Success:       true,
			BankReference: c2pResp.Reference,
			RawResponse:   decrypted,
		}, nil
	}

	// Extract error code from message (first 6 chars per BNC format).
	errCode := ""
	if len(resp.Message) >= 3 {
		errCode = strings.TrimSpace(resp.Message[:6])
	}

	gatewayErr := MapC2PError(errCode)

	return &bank.Response{
		Success:      false,
		ErrorCode:    gatewayErr.Code,
		ErrorMessage: gatewayErr.Message,
		RawResponse:  resp.Message,
	}, nil
}

// normalizePhone ensures the phone number is in BNC format: 584XXXXXXXXX (no +).
func normalizePhone(phone string) string {
	phone = strings.TrimPrefix(phone, "+")
	phone = strings.TrimPrefix(phone, "0")
	if !strings.HasPrefix(phone, "58") {
		phone = "58" + phone
	}
	return phone
}

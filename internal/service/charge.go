// Package service implements the business logic layer for the payment gateway.
// Services orchestrate handlers, stores, bank adapters, and crypto modules.
// They contain the core payment processing logic.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/faloppa/payment-gateway/internal/bank"
	"github.com/shopspring/decimal"
	"github.com/faloppa/payment-gateway/internal/crypto"
	"github.com/faloppa/payment-gateway/internal/model"
	"github.com/faloppa/payment-gateway/internal/store"
	"github.com/google/uuid"
)

// ChargeService orchestrates the full C2P charge flow:
// validate → create transaction → decrypt credentials → call bank → update status.
type ChargeService struct {
	txnStore    *store.TransactionStore
	cfgStore    *store.BankConfigStore
	registry    *bank.Registry
	encryptor   *crypto.EnvelopeEncryptor
	logger      *slog.Logger
}

// NewChargeService creates a ChargeService with all required dependencies.
func NewChargeService(
	txnStore *store.TransactionStore,
	cfgStore *store.BankConfigStore,
	registry *bank.Registry,
	encryptor *crypto.EnvelopeEncryptor,
	logger *slog.Logger,
) *ChargeService {
	return &ChargeService{
		txnStore:  txnStore,
		cfgStore:  cfgStore,
		registry:  registry,
		encryptor: encryptor,
		logger:    logger,
	}
}

// ProcessC2P executes a Cobro a Persona charge through the appropriate bank adapter.
// It follows this sequence:
//  1. Resolve the bank adapter from the registry.
//  2. Create a transaction record with INITIATED status.
//  3. Load and decrypt the merchant's bank credentials.
//  4. Transition to PROCESSING and call the bank adapter.
//  5. Map the bank response to a final status (SUCCESS/DECLINED/etc).
//  6. Return a standardized response to the client.
func (s *ChargeService) ProcessC2P(ctx context.Context, merchantID uuid.UUID, req *model.C2PChargeRequest) (*model.C2PChargeResponse, error) {
	// Step 1: Resolve bank adapter.
	adapter, err := s.registry.Get(req.Payer.BankCode)
	if err != nil {
		return nil, fmt.Errorf("unsupported bank code %s: %w", req.Payer.BankCode, err)
	}

	// Step 2: Create transaction record.
	txn := &model.Transaction{
		ID:             uuid.New(),
		MerchantID:     merchantID,
		IdempotencyKey: req.IdempotencyKey,
		BankCode:       req.Payer.BankCode,
		Amount:         req.Amount,
		Currency:       req.Currency,
		PayerPhone:     obfuscatePhone(req.Payer.Phone),
		PayerIDDoc:     obfuscateID(req.Payer.IDDocument),
		PayerBankCode:  req.Payer.BankCode,
		Status:         model.StatusInitiated,
		Metadata:       req.Metadata,
		InitiatedAt:    time.Now().UTC(),
	}

	if err := s.txnStore.Create(ctx, txn); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	s.logger.Info("transaction created",
		slog.String("txn_id", txn.ID.String()),
		slog.String("bank_code", txn.BankCode),
		slog.Float64("amount", txn.Amount),
	)

	// Step 3: Load and decrypt merchant bank credentials.
	bankCfg, err := s.cfgStore.FindActive(ctx, merchantID, req.Payer.BankCode)
	if err != nil {
		return s.failTransaction(ctx, txn, "CONFIG_ERROR", "No active bank configuration found", "")
	}

	credsPlaintext, err := s.encryptor.Decrypt(&crypto.EncryptedData{
		Ciphertext: bankCfg.EncryptedCredentials,
		SealedDEK:  bankCfg.KMSDataKeyCiphertext,
		IV:         bankCfg.EncryptionIV,
	})
	if err != nil {
		return s.failTransaction(ctx, txn, "DECRYPT_ERROR", "Failed to decrypt bank credentials", "")
	}
	defer crypto.ZeroBytes(credsPlaintext)

	// Step 4: Transition to PROCESSING and call bank adapter.
	if err := s.txnStore.UpdateStatus(ctx, txn.ID, model.StatusInitiated, model.StatusProcessing, "", "", ""); err != nil {
		return nil, fmt.Errorf("failed to transition to PROCESSING: %w", err)
	}

	bankReq := &bank.C2PRequest{
		Amount:        fromFloat(req.Amount),
		PayerBankCode: req.Payer.BankCode,
		PayerPhone:    req.Payer.Phone,
		PayerIDDoc:    req.Payer.IDDocument,
		OTP:           req.Payer.OTP,
		Reference:     txn.ID.String(),
	}

	bankResp, err := adapter.ProcessC2P(ctx, bankReq, json.RawMessage(credsPlaintext))

	// Step 5: Map response to final status.
	if err != nil {
		s.logger.Error("bank adapter error",
			slog.String("txn_id", txn.ID.String()),
			slog.String("error", err.Error()),
		)
		return s.pendingTransaction(ctx, txn)
	}

	if bankResp.Success {
		return s.succeedTransaction(ctx, txn, bankResp.BankReference)
	}

	return s.declineTransaction(ctx, txn, bankResp.ErrorCode, bankResp.ErrorMessage, bankResp.ErrorCode)
}

// succeedTransaction marks a transaction as SUCCESS with the bank reference.
func (s *ChargeService) succeedTransaction(ctx context.Context, txn *model.Transaction, bankRef string) (*model.C2PChargeResponse, error) {
	s.txnStore.UpdateStatus(ctx, txn.ID, model.StatusProcessing, model.StatusSuccess, bankRef, "", "")
	return &model.C2PChargeResponse{
		ID:            txn.ID.String(),
		Status:        model.StatusSuccess,
		BankReference: bankRef,
		Amount:        txn.Amount,
		Currency:      txn.Currency,
		CreatedAt:     txn.InitiatedAt.Format(time.RFC3339),
	}, nil
}

// declineTransaction marks a transaction as DECLINED with error details.
func (s *ChargeService) declineTransaction(ctx context.Context, txn *model.Transaction, errCode, errMsg, bankCode string) (*model.C2PChargeResponse, error) {
	s.txnStore.UpdateStatus(ctx, txn.ID, model.StatusProcessing, model.StatusDeclined, "", errCode, errMsg)
	return &model.C2PChargeResponse{
		ID:        txn.ID.String(),
		Status:    model.StatusDeclined,
		Amount:    txn.Amount,
		Currency:  txn.Currency,
		Error:     &model.ChargeError{Code: errCode, Message: errMsg, BankCode: bankCode},
		CreatedAt: txn.InitiatedAt.Format(time.RFC3339),
	}, nil
}

// pendingTransaction marks a transaction as PENDING_RECONCILIATION for background resolution.
func (s *ChargeService) pendingTransaction(ctx context.Context, txn *model.Transaction) (*model.C2PChargeResponse, error) {
	s.txnStore.UpdateStatus(ctx, txn.ID, model.StatusProcessing, model.StatusPendingReconciliation, "", "", "")
	return &model.C2PChargeResponse{
		ID:        txn.ID.String(),
		Status:    model.StatusPendingReconciliation,
		Amount:    txn.Amount,
		Currency:  txn.Currency,
		CreatedAt: txn.InitiatedAt.Format(time.RFC3339),
	}, nil
}

// failTransaction marks a transaction as DECLINED due to an internal error
// (e.g., missing config, decryption failure) before reaching the bank.
func (s *ChargeService) failTransaction(ctx context.Context, txn *model.Transaction, errCode, errMsg, bankCode string) (*model.C2PChargeResponse, error) {
	s.txnStore.UpdateStatus(ctx, txn.ID, model.StatusInitiated, model.StatusDeclined, "", errCode, errMsg)
	return &model.C2PChargeResponse{
		ID:       txn.ID.String(),
		Status:   model.StatusDeclined,
		Amount:   txn.Amount,
		Currency: txn.Currency,
		Error:    &model.ChargeError{Code: errCode, Message: errMsg, BankCode: bankCode},
		CreatedAt: txn.InitiatedAt.Format(time.RFC3339),
	}, nil
}

// obfuscatePhone partially masks a phone number for storage.
// Example: "04141234567" → "0414***4567"
func obfuscatePhone(phone string) string {
	if len(phone) <= 7 {
		return phone
	}
	return phone[:4] + "***" + phone[len(phone)-4:]
}

// obfuscateID partially masks an ID document for storage.
// Example: "V12345678" → "V***5678"
func obfuscateID(id string) string {
	if len(id) <= 5 {
		return id
	}
	return id[:1] + "***" + id[len(id)-4:]
}

// fromFloat converts float64 to decimal for the bank adapter.
func fromFloat(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

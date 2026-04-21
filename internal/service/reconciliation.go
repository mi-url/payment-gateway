// Package service implements the business logic layer for the payment gateway.
package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/faloppa/payment-gateway/internal/bank"
	"github.com/faloppa/payment-gateway/internal/crypto"
	"github.com/faloppa/payment-gateway/internal/model"
	"github.com/faloppa/payment-gateway/internal/store"
	"github.com/shopspring/decimal"
)

// ReconciliationService resolves transactions stuck in PENDING_RECONCILIATION.
// It runs as a background worker, periodically querying the bank for the
// outcome of ambiguous transactions (timeouts, network errors, etc).
type ReconciliationService struct {
	txnStore  *store.TransactionStore
	cfgStore  *store.BankConfigStore
	registry  *bank.Registry
	encryptor *crypto.EnvelopeEncryptor
	logger    *slog.Logger
	interval  time.Duration
	batchSize int
}

// NewReconciliationService creates a reconciliation worker.
func NewReconciliationService(
	txnStore *store.TransactionStore,
	cfgStore *store.BankConfigStore,
	registry *bank.Registry,
	encryptor *crypto.EnvelopeEncryptor,
	logger *slog.Logger,
) *ReconciliationService {
	return &ReconciliationService{
		txnStore:  txnStore,
		cfgStore:  cfgStore,
		registry:  registry,
		encryptor: encryptor,
		logger:    logger,
		interval:  5 * time.Minute,
		batchSize: 50,
	}
}

// Start begins the reconciliation loop. It runs until the context is cancelled.
// Each iteration fetches pending transactions and queries the bank for resolution.
func (s *ReconciliationService) Start(ctx context.Context) {
	s.logger.Info("reconciliation worker started",
		slog.Duration("interval", s.interval),
		slog.Int("batch_size", s.batchSize),
	)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("reconciliation worker stopped")
			return
		case <-ticker.C:
			s.reconcileBatch(ctx)
		}
	}
}

// reconcileBatch processes one batch of pending transactions.
func (s *ReconciliationService) reconcileBatch(ctx context.Context) {
	pending, err := s.txnStore.FindPendingReconciliation(ctx, s.batchSize)
	if err != nil {
		s.logger.Error("reconciliation: failed to fetch pending transactions",
			slog.String("error", err.Error()),
		)
		return
	}

	if len(pending) == 0 {
		return
	}

	s.logger.Info("reconciliation: processing batch",
		slog.Int("count", len(pending)),
	)

	for _, txn := range pending {
		s.reconcileOne(ctx, txn)
	}
}

// reconcileOne queries the bank for a single pending transaction.
func (s *ReconciliationService) reconcileOne(ctx context.Context, txn *model.Transaction) {
	adapter, err := s.registry.Get(txn.BankCode)
	if err != nil {
		s.logger.Error("reconciliation: no adapter for bank",
			slog.String("txn_id", txn.ID.String()),
			slog.String("bank_code", txn.BankCode),
		)
		return
	}

	bankCfg, err := s.cfgStore.FindActive(ctx, txn.MerchantID, txn.BankCode)
	if err != nil {
		s.logger.Error("reconciliation: failed to load bank config",
			slog.String("txn_id", txn.ID.String()),
			slog.String("error", err.Error()),
		)
		return
	}

	credsPlaintext, err := s.encryptor.Decrypt(&crypto.EncryptedData{
		Ciphertext: bankCfg.EncryptedCredentials,
		SealedDEK:  bankCfg.KMSDataKeyCiphertext,
		IV:         bankCfg.EncryptionIV,
	})
	if err != nil {
		s.logger.Error("reconciliation: failed to decrypt credentials",
			slog.String("txn_id", txn.ID.String()),
			slog.String("error", err.Error()),
		)
		return
	}
	defer crypto.ZeroBytes(credsPlaintext)

	result, err := adapter.QueryTransaction(
		ctx,
		txn.IdempotencyKey,
		decimal.NewFromFloat(txn.Amount),
		txn.InitiatedAt,
		json.RawMessage(credsPlaintext),
	)
	if err != nil {
		s.logger.Warn("reconciliation: bank query failed, will retry",
			slog.String("txn_id", txn.ID.String()),
			slog.String("error", err.Error()),
		)
		return
	}

	if result.Exists {
		s.txnStore.UpdateStatus(ctx, txn.ID, model.StatusPendingReconciliation, model.StatusSuccess, "", "", "")
		s.logger.Info("reconciliation: transaction confirmed as SUCCESS",
			slog.String("txn_id", txn.ID.String()),
		)
	} else {
		s.txnStore.UpdateStatus(ctx, txn.ID, model.StatusPendingReconciliation, model.StatusDeclined, "", "NOT_FOUND", "Transaction not found at bank after reconciliation")
		s.logger.Info("reconciliation: transaction confirmed as DECLINED",
			slog.String("txn_id", txn.ID.String()),
		)
	}
}

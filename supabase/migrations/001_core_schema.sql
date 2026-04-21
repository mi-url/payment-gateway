-- Migration: 001_core_schema
-- Description: Creates the fundamental tables for the payment gateway:
--   merchants, merchant_bank_configs, and transactions.
-- Author: Payment Gateway Team
-- Date: 2026-04-21

-- transaction_status defines the state machine for payment processing.
-- Transitions: INITIATED → PROCESSING → SUCCESS | DECLINED | BANK_NETWORK_ERROR | PENDING_RECONCILIATION
CREATE TYPE transaction_status AS ENUM (
  'INITIATED',
  'PROCESSING',
  'SUCCESS',
  'DECLINED',
  'BANK_NETWORK_ERROR',
  'PENDING_RECONCILIATION'
);

-- merchants holds B2B clients who use the gateway to process payments.
-- Each merchant authenticates via a hashed API key and receives
-- transaction notifications at their configured webhook URL.
CREATE TABLE merchants (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  api_key_hash VARCHAR(64) NOT NULL,
  saas_subscription_status VARCHAR(20) NOT NULL DEFAULT 'active'
    CHECK (saas_subscription_status IN ('active', 'past_due', 'canceled')),
  webhook_url TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- merchant_bank_configs stores encrypted bank credentials under the BYOC model.
-- Credentials are protected via Envelope Encryption: the encrypted_credentials
-- column holds AES-256-GCM ciphertext, and kms_data_key_ciphertext holds the
-- KMS-sealed Data Encryption Key (DEK). Plaintext credentials never touch disk.
CREATE TABLE merchant_bank_configs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
  bank_code VARCHAR(4) NOT NULL,
  encrypted_credentials BYTEA NOT NULL,
  kms_data_key_ciphertext BYTEA NOT NULL,
  encryption_iv BYTEA NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(merchant_id, bank_code)
);

-- transactions is the core ledger. Every C2P charge attempt creates exactly
-- one row. The idempotency_key constraint prevents duplicate charges.
-- PII fields (payer_phone, payer_id_document) are stored partially obfuscated.
CREATE TABLE transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  merchant_id UUID NOT NULL REFERENCES merchants(id),
  idempotency_key VARCHAR(255) NOT NULL,
  bank_code VARCHAR(4) NOT NULL,
  amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
  currency VARCHAR(3) NOT NULL DEFAULT 'VES',
  payer_phone VARCHAR(20),
  payer_id_document VARCHAR(20),
  payer_bank_code VARCHAR(4),
  bank_reference VARCHAR(20),
  status transaction_status NOT NULL DEFAULT 'INITIATED',
  error_code VARCHAR(20),
  error_message TEXT,
  metadata JSONB,
  initiated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  processed_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  UNIQUE(merchant_id, idempotency_key)
);

-- Performance indexes for common query patterns.
CREATE INDEX idx_transactions_merchant_status
  ON transactions(merchant_id, status);

CREATE INDEX idx_transactions_bank_reference
  ON transactions(bank_reference)
  WHERE bank_reference IS NOT NULL;

-- Partial index: only pending transactions need background reconciliation.
CREATE INDEX idx_transactions_pending
  ON transactions(status)
  WHERE status = 'PENDING_RECONCILIATION';

-- Partial index: active bank configs for fast credential lookup.
CREATE INDEX idx_merchant_bank_configs_lookup
  ON merchant_bank_configs(merchant_id, bank_code)
  WHERE is_active = true;

-- Row-Level Security: prevents cross-merchant data access at the DB layer.
ALTER TABLE transactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE merchant_bank_configs ENABLE ROW LEVEL SECURITY;

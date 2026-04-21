-- Seed data for local development.
-- Run this AFTER 001_core_schema.sql.
--
-- This creates a test merchant with a known API key for local testing.
-- API Key: gw_test_0000000000000000000000000000000000000000000000000000000000000000
-- SHA-256:  (computed below)
--
-- WARNING: This seed data is for DEVELOPMENT ONLY. Never use these
-- credentials in production.

-- Test merchant with a deterministic API key for local development.
-- The API key "gw_test_localdev" hashes to the value below.
-- Generate your own with: go run ./cmd/keygen -name "My Company"
INSERT INTO merchants (
  id,
  name,
  api_key_hash,
  saas_subscription_status,
  webhook_url
) VALUES (
  '00000000-0000-0000-0000-000000000001',
  'Test Merchant (Local Dev)',
  -- SHA-256 of "gw_test_localdev"
  'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
  'active',
  'http://localhost:3000/webhooks/payments'
) ON CONFLICT (id) DO NOTHING;

-- Test BNC bank configuration for the test merchant.
-- These are MOCK credentials — they will only work with swTestOperation=true.
-- In development, MockKMS is used so the encrypted_credentials are stored as plaintext JSON.
INSERT INTO merchant_bank_configs (
  id,
  merchant_id,
  bank_code,
  encrypted_credentials,
  kms_data_key_ciphertext,
  encryption_iv,
  is_active
) VALUES (
  '00000000-0000-0000-0000-000000000002',
  '00000000-0000-0000-0000-000000000001',
  '0191',
  -- Mock BNC credentials (JSON encoded as bytes)
  convert_to('{"client_guid":"TEST-GUID-0000-0000","master_key":"TEST-MASTER-KEY","terminal":"TERM001"}', 'UTF8'),
  -- Mock sealed DEK (not real encryption in dev mode)
  '\x0000000000000000000000000000000000000000000000000000000000000000',
  -- Mock IV
  '\x000000000000000000000000',
  true
) ON CONFLICT (id) DO NOTHING;

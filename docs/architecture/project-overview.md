# Faloppa Payments Gateway — Project Overview

## What Is This?

A **BYOC (Bring Your Own Credentials) Payment Gateway** — a technology switch/middleware (SaaS) that routes C2P (Cobro a Persona) transactions between merchant apps and Venezuelan bank APIs.

**Core principle:** The gateway **NEVER touches, holds, or settles funds**. Money flows directly from payer → merchant's bank account via interbank rails.

## Legal Model

- **Switch Tecnológico** — NOT an aggregador financiero
- Exempt from SUDEBAN Article 16 licensing (no fund custody)
- Corporation: **US LLC** for IP, billing (Stripe), and infrastructure
- Compliant with OFAC general licenses for VZ bank interactions

## Tech Stack

| Layer | Technology | Notes |
|---|---|---|
| Core Payment Engine | **Go (Golang)** | High-concurrency, crypto-heavy workloads |
| Dashboard B2B | **Next.js + TypeScript** | Merchant portal, credential management |
| Database/Auth | **Supabase (PostgreSQL)** | RLS, JWT, managed DB |
| Infrastructure | **GCP Cloud Run** | Serverless, scale-to-zero, ~$40/mo MVP |
| Security | **GCP Cloud KMS** | Envelope Encryption (AES-256-GCM) |
| CI/CD | **GitHub Actions** | Auto-build, containerize (Alpine), deploy |
| Billing | **Stripe** | SaaS subscription in USD |

## Database Schema

### `merchants`
- `id` (UUID PK), `name`, `saas_subscription_status` (active/past_due/canceled), `created_at`

### `merchant_bank_configs`
- `id` (UUID PK), `merchant_id` (FK), `bank_code` (e.g., '0191')
- `encrypted_credentials` (BYTEA — AES-256-GCM encrypted)
- `kms_data_key_ciphertext` (BYTEA — KMS-sealed DEK)
- `is_active` (BOOLEAN)

### `transactions`
- `id` (UUID PK), `merchant_id` (FK), `idempotency_key` (UNIQUE)
- `bank_code`, `amount` (DECIMAL 12,2), `currency` ('VES')
- `payer_phone`, `payer_id_document` (partially obfuscated)
- `bank_reference` (12-digit, post-confirmation)
- `status` (ENUM): INITIATED → PROCESSING → SUCCESS | DECLINED | BANK_NETWORK_ERROR | PENDING_RECONCILIATION

## Transaction State Machine

```
INITIATED → PROCESSING → SUCCESS (bank_reference captured)
                       → DECLINED (bank rejected: bad OTP, no funds, etc.)
                       → BANK_NETWORK_ERROR (pre-TLS failure, safe to retry)
                       → PENDING_RECONCILIATION (timeout/ambiguous response, background worker resolves)
```

## Envelope Encryption (Zero-Trust)

1. **Onboarding:** Gateway calls KMS `GenerateDataKey` → gets DEK (plaintext + ciphertext)
2. **Storage:** Encrypts merchant credentials with DEK (AES-256-GCM) → stores ciphertext + encrypted DEK in DB
3. **Transaction:** Reads DB → sends encrypted DEK to KMS `Decrypt` → gets DEK plaintext → decrypts credentials in RAM → calls bank API → zeroes memory
4. **Isolation:** Only Cloud Run service account has IAM permission to call KMS Decrypt

## Bank Integration Status

| Bank | Code | API Documentation | Status |
|---|---|---|---|
| **BNC** | 0191 | ✅ ESolutions API v4.1 + NotificationPush V2.0 | **CONFIRMED — Ready to implement** |
| Mercantil | 0105 | ❌ Not obtained | Assumptions from Master Plan only |
| Banesco | 0134 | ❌ Not obtained | Assumptions from Master Plan only |
| Bancamiga | — | ❌ Not obtained | Minimal info |

## Phase 1 (MVP)

- Target: **Padel App** (Flutter) with embedded C2P payments
- Single merchant initially, expandable to multi-tenant
- BNC as primary bank adapter (only confirmed documentation)

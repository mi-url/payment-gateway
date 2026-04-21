# Payment Gateway

> **BYOC (Bring Your Own Credentials) Payment Gateway for Venezuelan banks.**

A production-grade middleware that abstracts fragmented Venezuelan bank APIs into a unified REST API. Acts as a pure transaction router — **zero funds retention**.

## Architecture

```
Client (SDK/Plugin/API) → Payment Gateway → Bank API (BNC, Mercantil, ...)
                              ↓                         ↓
                         PostgreSQL              Push Notifications
                         (Supabase)                 (Webhooks)
```

**Core Engine:** Go | **Database:** PostgreSQL (Supabase) | **Hosting:** GCP Cloud Run | **Security:** AES-256-GCM + Cloud KMS

## API

### Create a C2P Charge

```bash
POST /v1/charges/c2p
Authorization: Bearer <api_key>
Idempotency-Key: unique-key-123

{
  "amount": 250.00,
  "currency": "VES",
  "payer": {
    "bank_code": "0191",
    "phone": "04141234567",
    "id_document": "V12345678",
    "otp": "A4H9B2"
  },
  "idempotency_key": "unique-key-123"
}
```

### Response

```json
{
  "id": "txn_uuid",
  "status": "SUCCESS",
  "bank_reference": "123456789012",
  "amount": 250.00,
  "currency": "VES",
  "created_at": "2026-04-21T00:00:00Z"
}
```

### Query a Transaction

```bash
GET /v1/transactions/{id}
Authorization: Bearer <api_key>
```

## Transaction States

```
INITIATED → PROCESSING → SUCCESS
                       → DECLINED
                       → BANK_NETWORK_ERROR
                       → PENDING_RECONCILIATION → SUCCESS / DECLINED
```

## Supported Banks

| Bank | Code | Status |
|---|---|---|
| BNC (Banco Nacional de Crédito) | 0191 | ✅ Implemented |
| Mercantil | 0105 | 🔒 Pending documentation |
| Banesco | 0134 | 🔒 Pending documentation |
| Banco Plaza | 0138 | 🔒 Pending documentation |

Adding a new bank = implementing the `BankAdapter` interface. No other code changes required.

## Development

### Prerequisites

- Go 1.22+
- PostgreSQL (via Supabase)
- GCP Cloud KMS (or MockKMS for development)

### Run locally

```bash
export DATABASE_URL="postgres://..."
export KMS_KEY_RESOURCE_NAME="projects/..."
go run ./cmd/gateway
```

### Run tests

```bash
go test -v -race ./...
```

### Build

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o gateway ./cmd/gateway
```

## Security

- **Envelope Encryption:** Bank credentials encrypted with AES-256-GCM, DEKs sealed by Cloud KMS (HSM-backed, FIPS 140-2 Level 3)
- **Zero plaintext storage:** Credentials exist in plaintext only in volatile memory during transaction processing, then zeroed
- **API key hashing:** SHA-256 — raw keys never stored
- **PII obfuscation:** Phone and ID document partially masked before storage
- **Row-Level Security:** PostgreSQL RLS prevents cross-merchant data access

## Documentation

- [`docs/engineering-standards.md`](docs/engineering-standards.md) — Non-negotiable audit-ready code standards
- [`docs/architecture/project-overview.md`](docs/architecture/project-overview.md) — Technical architecture overview
- [`docs/banks/bnc/api-reference.md`](docs/banks/bnc/api-reference.md) — BNC API technical reference
- [`docs/banks/integration-status.md`](docs/banks/integration-status.md) — Venezuelan banking ecosystem research

## License

Proprietary. All rights reserved.

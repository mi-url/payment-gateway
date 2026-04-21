# BNC (Banco Nacional de Crédito) — API Reference

> **Bank Code:** `0191`
> **Status:** ✅ Documentation confirmed and verified
> **Sources:** ESolutions API v4.1 + NotificationPush V2.0

---

## 1. General Architecture

**Base URL (QA):** `https://servicios.bncenlinea.com:16500/api`
**Health Check:** `https://servicios.bncenlinea.com:16500/api/welcome/home`
**API Version:** v4.1 (as of Aug 2024)

### Credentials Required from Merchant
- `ClientGUID` — 36-char UUID assigned by BNC
- `MasterKey` — Encryption key for Logon only
- `WorkingKey` — Daily session key (obtained via Logon, expires at midnight)

### All Requests Are POST with Encrypted Envelope

```json
{
  "ClientGUID": "4A074C46-DD4E-4E54-8010-B80A6A8758F4",
  "Reference": "UniqueAlphanumericIdForTheDay",
  "Value": "<AES_ENCRYPTED_PAYLOAD>",
  "Validation": "<SHA256_HASH_OF_PLAINTEXT_PAYLOAD>",
  "swTestOperation": false
}
```

### All Responses

```json
{
  "status": "OK | KO",
  "message": "6-char-code + message text",
  "value": "<AES_ENCRYPTED_RESPONSE>",
  "validation": "<SHA256_HASH>"
}
```

---

## 2. Encryption Scheme

**Algorithm:** AES (Rijndael) / CBC mode
**Key Derivation:** PBKDF2 with SHA-1

### Parameters
| Parameter | Value |
|---|---|
| Salt | Fixed: `"Ivan Medvedev"` → `[0x49,0x76,0x61,0x6e,0x20,0x4d,0x65,0x64,0x76,0x65,0x64,0x65,0x76]` |
| Iterations | 1000 |
| Hash | SHA-1 |
| Output | 48 bytes total: 32 bytes key + 16 bytes IV |
| Text Encoding | **UTF-16LE** (critical — not UTF-8!) |

### Key Usage
| Operation | Encryption Key |
|---|---|
| Logon | `MasterKey` |
| All other operations | `WorkingKey` (from Logon response) |

### Go Implementation Notes
```go
// Pseudocode for BNC encryption in Go:
// 1. Derive key/IV from encryptionKey using PBKDF2(SHA1, salt, 1000 iterations, 48 bytes)
// 2. Split: key = derived[0:32], iv = derived[32:48]
// 3. Encode plaintext JSON as UTF-16LE bytes
// 4. AES-CBC encrypt with PKCS7 padding
// 5. Base64 encode result → "Value" field
// 6. SHA-256 hash of plaintext JSON string → "Validation" field
```

---

## 3. Endpoints

### 3.1 Auth/Logon

**Endpoint:** `POST /api/Auth/LogOn`
**Encrypt with:** `MasterKey`
**Frequency:** Once daily (or when `EPIRWK` error received)

**Value Payload:**
```json
{"ClientGUID": "4A074C46-DD4E-4E54-8010-B80A6A8758F4"}
```

**Decrypted Response:**
```json
{"WorkingKey": "<hex_string>"}
```

> ⚠️ If any subsequent operation returns code `EPIRWK` in `message`, the WorkingKey has been invalidated. Must re-authenticate.

---

### 3.2 C2P — Cobro a Persona ⭐ (Primary for MVP)

**Endpoint:** `POST /api/MobPayment/SendC2P`
**Encrypt with:** `WorkingKey`

**Value Payload:**
```json
{
  "DebtorBankCode": "0191",
  "DebtorCellPhone": "584141234567",
  "DebtorID": "V12345678",
  "Amount": 250.00,
  "Token": "A4H9B2",
  "Terminal": "TERM001",
  "ChildClientID": "",
  "BranchID": ""
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| DebtorBankCode | string | ✅ | Payer's bank code (4 digits) |
| DebtorCellPhone | string | ✅ | Payer phone: `584XXXXXXXXX` |
| DebtorID | string | ✅ | Payer document: `V12345678` |
| Amount | decimal | ✅ | Amount in VES |
| Token | string | ✅ | OTP from payer's bank |
| Terminal | string | ✅ | Terminal identifier |
| ChildClientID | string | ❌ | Sub-merchant RIF |
| BranchID | string | ❌ | Branch ID |

**C2P Error Codes:**
| Code | Description | Gateway Action |
|---|---|---|
| G56 | Card/Phone not registered | → DECLINED |
| G13 | Invalid | → DECLINED |
| G55 | Incorrect PIN/OTP | → DECLINED |
| G91 | Issuer inoperative | → BANK_NETWORK_ERROR |
| G41 | Lost card | → DECLINED |
| G61 | Withdrawal limit exceeded | → DECLINED |

---

### 3.3 P2P — Pago Móvil (Emission)

**Endpoint:** `POST /api/P2P/Payment`
**Encrypt with:** `WorkingKey`

**Value Payload:**
```json
{
  "Amount": 10.01,
  "BeneficiaryBankCode": 191,
  "BeneficiaryCellPhone": "584242207524",
  "BeneficiaryEmail": "",
  "BeneficiaryID": "V23000760",
  "BeneficiaryName": "NombreBeneficiario",
  "Description": "Descripcion del pago",
  "OperationRef": "referencia_unica"
}
```

**P2P Error Codes:**
| Code | Description |
|---|---|
| G51 | Insufficient funds |
| G52 | Beneficiary not enrolled in Pago Móvil |
| G61 | Daily amount limit exceeded |
| G65 | Daily transaction count exceeded |
| G80 | Invalid beneficiary ID |

---

### 3.4 VPOS — Virtual Point of Sale

**Endpoint:** `POST /api/Transaction/Send`
**Encrypt with:** `WorkingKey`

**Value Payload:**
```json
{
  "TransactionIdentifier": "unique_txn_id",
  "Amount": 100.50,
  "idCardType": 1,
  "CardNumber": "4111111111111111",
  "dtExpiration": "122025",
  "CVV": "123",
  "CardPIN": "1234",
  "AffiliationNumber": "AFF001"
}
```

| idCardType | Meaning |
|---|---|
| 1 | Debit |
| 3 | Credit |

---

### 3.5 Transaction Query / Validation

**Endpoint:** `POST /api/Validation/TransactionQuery`
**Purpose:** Verify transaction outcome (for `PENDING_RECONCILIATION` resolution)

**Value Payload:**
```json
{
  "Reference": "original_reference",
  "Amount": 250.00,
  "Date": "20/04/2026"
}
```

**Response:** `{"MovementExists": true|false}`

---

### 3.6 Crédito Inmediato (Push Payment)

**Endpoint:** `POST /api/ImmediateCredit/Payment`

---

### 3.7 Débito Inmediato (Pull Payment)

**Step 1:** `POST /api/ImmediateDebit/RequestToken` → sends SMS to debtor
**Step 2:** `POST /api/SIMF/DebitBeginner` → executes debit with token

---

## 4. Multi-Tenant Support

```
Parent (ClientGUID only)
├── Child A (+ ChildClientID: "J00000001")
│   ├── Branch A1 (+ BranchID: "CS400")
│   └── Branch A2 (+ BranchID: "CS401")
└── Child B (+ ChildClientID: "J00000002")
```

- Parent generates daily WorkingKey; all children/branches share it
- Children add `ChildClientID` to transaction payloads
- Branches add both `ChildClientID` + `BranchID`

---

## 5. NotificationPush V2.0 (Webhooks)

BNC pushes payment confirmations to merchant's webhook endpoint.

### Authentication Options

**Option A — API Key:**
```
Header: x-api-key: <merchant_provided_key>
```

**Option B — JWT:**
1. BNC calls merchant auth URL with `{"Login":"...","Password":"..."}`
2. Merchant returns JWT token as plain text
3. BNC calls notification URL with `Authorization: Bearer <token>`

### Webhook Payload — P2P
```json
{
  "PaymentType": "P2P",
  "OriginBankReference": "ref123",
  "DestinyBankReference": "bncref456",
  "OriginBankCode": "0105",
  "ClientID": "V12345678",
  "Hour": "1430",
  "CurrencyCode": "0928",
  "Amount": "250.00",
  "Date": "20260420",
  "CommerceID": "J123456789",
  "CommercePhone": "00584141234567",
  "ClientPhone": "00584241234567",
  "Concept": "Pago reserva padel"
}
```

### Webhook Payload — Transfer/Deposit
```json
{
  "PaymentType": "DEP",
  "OriginBankReference": "ref123",
  "DestinyBankReference": "bncref456",
  "OriginBankCode": "0105",
  "Hour": "1430",
  "CurrencyCode": "0928",
  "Amount": "250.00",
  "Date": "20260420",
  "CommerceID": "J123456789",
  "CommercePhone": "00584141234567",
  "DebtorAccount": "01050000000000000000",
  "DebtorID": "V012345678",
  "CreditorAccount": "01910000000000000000"
}
```

### Critical Webhook Rules
1. **Respond HTTP 200 IMMEDIATELY** before any processing
2. BNC retries indefinitely until 200 received
3. Validate references to prevent duplicate processing
4. Service must be available 24/7
5. BNC can deactivate webhook if too many errors

### Certification Requirements
- Provide separate URLs for Development and Production
- BNC runs automatic ping test during registration
- No production testing allowed
- IP whitelist may need BNC's IP added

---

## 6. General Error Codes

| Code | Description | Action |
|---|---|---|
| EPIRWK | Security — Refresh WorkingKey | Re-authenticate immediately |
| EPICNF | Client not found or inactive | Check ClientGUID |
| EPIIMS | Model validation failed | Check payload structure |
| EPIHV | Invalid request hash | Check SHA-256 validation |
| EPIECP | Exception inserting C2P | Retry / investigate |
| EPIONA | No permissions | Check merchant config |
| EPIMC1 | No ChildClient permissions | Check hierarchy |
| EPIMC2 | Commerce not found | Check ChildClientID |

### Débito Inmediato ISO Codes
| Code | Description |
|---|---|
| ACCP | Accepted |
| AB01 | Timeout (120s max) |
| AM04 | Insufficient funds |
| AM05 | Duplicate operation |
| CUST | Debtor declined (cancelled OTP) |
| TKCM | Invalid debit token |
| RJCT | Rejected |

// Package bnc implements the bank adapter for Banco Nacional de Crédito (BNC).
package bnc

// BNC API request and response types.
// These match the BNC ESolutions API v4.1 envelope format exactly.

// Envelope is the outer request structure sent to all BNC API endpoints.
// The Value field contains the AES-encrypted payload, and Validation
// contains the SHA-256 hash of the plaintext payload for integrity.
type Envelope struct {
	ClientGUID      string `json:"ClientGUID"`
	Reference       string `json:"Reference"`
	Value           string `json:"Value"`
	Validation      string `json:"Validation"`
	SwTestOperation bool   `json:"swTestOperation"`
}

// EnvelopeResponse is the outer response structure from all BNC API endpoints.
type EnvelopeResponse struct {
	Status     string `json:"status"`     // "OK" or "KO"
	Message    string `json:"message"`    // 6-char code + message text
	Value      string `json:"value"`      // AES-encrypted response payload
	Validation string `json:"validation"` // SHA-256 hash of response payload
}

// LogonPayload is encrypted and sent as the Value in a Logon request.
type LogonPayload struct {
	ClientGUID string `json:"ClientGUID"`
}

// LogonResponse is decrypted from the Value in a Logon response.
type LogonResponse struct {
	WorkingKey string `json:"WorkingKey"`
}

// C2PPayload is encrypted and sent as the Value in a C2P charge request.
type C2PPayload struct {
	DebtorBankCode  string  `json:"DebtorBankCode"`
	DebtorCellPhone string  `json:"DebtorCellPhone"`
	DebtorID        string  `json:"DebtorID"`
	Amount          float64 `json:"Amount"`
	Token           string  `json:"Token"`
	Terminal        string  `json:"Terminal"`
	ChildClientID   string  `json:"ChildClientID,omitempty"`
	BranchID        string  `json:"BranchID,omitempty"`
}

// C2PResponse is decrypted from the Value in a successful C2P response.
type C2PResponse struct {
	IDTransaction string `json:"IdTransaction"`
	Reference     string `json:"Reference"`
}

// TransactionQueryPayload is encrypted and sent for transaction verification.
type TransactionQueryPayload struct {
	Reference string  `json:"Reference"`
	Amount    float64 `json:"Amount"`
	Date      string  `json:"Date"` // Format: dd/MM/yyyy
}

// TransactionQueryResponse is decrypted from a transaction query response.
type TransactionQueryResponse struct {
	MovementExists bool `json:"MovementExists"`
}

// BNCCredentials represents the decrypted bank credentials for a BNC merchant.
// These are stored encrypted in the database and only exist in plaintext
// in volatile memory during transaction processing.
type BNCCredentials struct {
	ClientGUID    string `json:"client_guid"`
	MasterKey     string `json:"master_key"`
	ChildClientID string `json:"child_client_id,omitempty"`
	BranchID      string `json:"branch_id,omitempty"`
	Terminal      string `json:"terminal"`
}

// WebhookNotification represents an incoming BNC NotificationPush payload.
type WebhookNotification struct {
	PaymentType         string `json:"PaymentType"` // P2P, TRF, DEP
	OriginBankReference string `json:"OriginBankReference"`
	DestinyBankReference string `json:"DestinyBankReference"`
	OriginBankCode      string `json:"OriginBankCode"`
	ClientID            string `json:"ClientID,omitempty"`
	Hour                string `json:"Hour"`
	CurrencyCode        string `json:"CurrencyCode"`
	Amount              string `json:"Amount"`
	Date                string `json:"Date"`
	CommerceID          string `json:"CommerceID"`
	CommercePhone       string `json:"CommercePhone"`
	ClientPhone         string `json:"ClientPhone,omitempty"` // P2P only
	Concept             string `json:"Concept,omitempty"`     // P2P only
	DebtorAccount       string `json:"DebtorAccount,omitempty"` // TRF/DEP only
	DebtorID            string `json:"DebtorID,omitempty"`       // TRF/DEP only
	CreditorAccount     string `json:"CreditorAccount,omitempty"` // TRF/DEP only
}

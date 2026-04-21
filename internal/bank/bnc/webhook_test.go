package bnc

import (
	"testing"
)

// TestParseWebhookNotificationP2P verifies parsing of a P2P notification.
func TestParseWebhookNotificationP2P(t *testing.T) {
	body := []byte(`{
		"PaymentType": "P2P",
		"OriginBankReference": "123456789012",
		"DestinyBankReference": "987654321098",
		"OriginBankCode": "0102",
		"Amount": "250.00",
		"Date": "21/04/2026",
		"Hour": "08:30:00",
		"CommerceID": "J123456789",
		"CommercePhone": "00584141234567",
		"ClientPhone": "00584142345678",
		"Concept": "Pago reserva cancha"
	}`)

	notif, err := ParseWebhookNotification(body)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if notif.PaymentType != "P2P" {
		t.Errorf("PaymentType = %q, want P2P", notif.PaymentType)
	}
	if notif.OriginBankReference != "123456789012" {
		t.Errorf("OriginBankReference = %q, want 123456789012", notif.OriginBankReference)
	}
	if notif.Amount != "250.00" {
		t.Errorf("Amount = %q, want 250.00", notif.Amount)
	}
}

// TestParseWebhookNotificationTRF verifies parsing of a TRF (transfer) notification.
func TestParseWebhookNotificationTRF(t *testing.T) {
	body := []byte(`{
		"PaymentType": "TRF",
		"OriginBankReference": "111111111111",
		"DestinyBankReference": "222222222222",
		"OriginBankCode": "0105",
		"Amount": "1500.50",
		"Date": "21/04/2026",
		"Hour": "10:00:00",
		"CommerceID": "J987654321",
		"CommercePhone": "00584149876543",
		"DebtorAccount": "01050012345678901234",
		"DebtorID": "V12345678",
		"CreditorAccount": "01910098765432109876"
	}`)

	notif, err := ParseWebhookNotification(body)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if notif.PaymentType != "TRF" {
		t.Errorf("PaymentType = %q, want TRF", notif.PaymentType)
	}
	if notif.DebtorID != "V12345678" {
		t.Errorf("DebtorID = %q, want V12345678", notif.DebtorID)
	}
}

// TestParseWebhookNotificationInvalidJSON verifies graceful handling of bad JSON.
func TestParseWebhookNotificationInvalidJSON(t *testing.T) {
	_, err := ParseWebhookNotification([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// TestParseWebhookNotificationMissingType verifies rejection of missing PaymentType.
func TestParseWebhookNotificationMissingType(t *testing.T) {
	body := []byte(`{"Amount": "100.00"}`)
	_, err := ParseWebhookNotification(body)
	if err == nil {
		t.Error("expected error for missing PaymentType, got nil")
	}
}

// TestParseWebhookNotificationUnknownType verifies rejection of invalid PaymentType.
func TestParseWebhookNotificationUnknownType(t *testing.T) {
	body := []byte(`{"PaymentType": "INVALID"}`)
	_, err := ParseWebhookNotification(body)
	if err == nil {
		t.Error("expected error for unknown PaymentType, got nil")
	}
}

// TestParseAmount verifies amount string parsing.
func TestParseAmount(t *testing.T) {
	tests := []struct {
		input string
		want  float64
		err   bool
	}{
		{"250.00", 250.00, false},
		{"0.01", 0.01, false},
		{"1500.50", 1500.50, false},
		{"", 0, true},
		{"abc", 0, true},
		{"-100.00", 0, true},
	}

	for _, tc := range tests {
		amount, err := ParseAmount(tc.input)
		if tc.err && err == nil {
			t.Errorf("ParseAmount(%q): expected error, got %f", tc.input, amount)
		}
		if !tc.err && err != nil {
			t.Errorf("ParseAmount(%q): unexpected error: %v", tc.input, err)
		}
		if !tc.err && amount != tc.want {
			t.Errorf("ParseAmount(%q) = %f, want %f", tc.input, amount, tc.want)
		}
	}
}

// TestNormalizeWebhookPhone verifies phone normalization from BNC format.
func TestNormalizeWebhookPhone(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"00584141234567", "04141234567"},
		{"584141234567", "04141234567"},
		{"4141234567", "04141234567"},
		{"04141234567", "04141234567"},
	}

	for _, tc := range tests {
		got := NormalizeWebhookPhone(tc.input)
		if got != tc.want {
			t.Errorf("NormalizeWebhookPhone(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

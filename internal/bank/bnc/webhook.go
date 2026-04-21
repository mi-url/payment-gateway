// Package bnc implements the bank adapter for Banco Nacional de Crédito (BNC).
package bnc

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ParseWebhookNotification parses a raw BNC NotificationPush JSON payload
// into a structured WebhookNotification. Returns an error if the JSON is invalid.
func ParseWebhookNotification(body []byte) (*WebhookNotification, error) {
	var notif WebhookNotification
	if err := json.Unmarshal(body, &notif); err != nil {
		return nil, fmt.Errorf("bnc webhook: failed to parse notification: %w", err)
	}

	if notif.PaymentType == "" {
		return nil, fmt.Errorf("bnc webhook: missing PaymentType field")
	}

	validTypes := map[string]bool{"P2P": true, "TRF": true, "DEP": true}
	if !validTypes[notif.PaymentType] {
		return nil, fmt.Errorf("bnc webhook: unknown PaymentType: %s", notif.PaymentType)
	}

	return &notif, nil
}

// ParseAmount converts BNC's amount string (format "15+2" with "." decimal)
// to a float64. Example: "250.00" → 250.00
func ParseAmount(amountStr string) (float64, error) {
	amountStr = strings.TrimSpace(amountStr)
	if amountStr == "" {
		return 0, fmt.Errorf("bnc webhook: empty amount string")
	}
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, fmt.Errorf("bnc webhook: invalid amount %q: %w", amountStr, err)
	}
	if amount < 0 {
		return 0, fmt.Errorf("bnc webhook: negative amount: %f", amount)
	}
	return amount, nil
}

// NormalizeWebhookPhone converts BNC webhook phone format (00584XXXXXXXXX)
// to the standard gateway format (04XXXXXXXXX) for matching.
func NormalizeWebhookPhone(phone string) string {
	phone = strings.TrimPrefix(phone, "00")
	phone = strings.TrimPrefix(phone, "58")
	if !strings.HasPrefix(phone, "0") {
		phone = "0" + phone
	}
	return phone
}

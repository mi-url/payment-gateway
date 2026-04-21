// Package bnc implements the bank adapter for Banco Nacional de Crédito (BNC).
package bnc

// ErrorMapping maps BNC error codes to gateway-standard error codes and messages.
// These mappings are derived from BNC ESolutions API v4.1 documentation.

// GatewayError represents a standardized error returned to clients.
type GatewayError struct {
	Code    string // Gateway-standard error code
	Message string // Human-readable message
}

// c2pErrors maps BNC C2P response codes to gateway errors.
var c2pErrors = map[string]GatewayError{
	"G13": {Code: "INVALID_REQUEST", Message: "The request contains invalid data"},
	"G41": {Code: "CARD_BLOCKED", Message: "The payer's card has been reported as lost"},
	"G55": {Code: "INCORRECT_OTP", Message: "The OTP provided is incorrect or expired"},
	"G56": {Code: "PAYER_NOT_ENROLLED", Message: "The payer's phone or card is not registered for mobile payments"},
	"G61": {Code: "LIMIT_EXCEEDED", Message: "The payer has exceeded their withdrawal limit"},
	"G91": {Code: "BANK_UNAVAILABLE", Message: "The issuing bank is temporarily unavailable"},
}

// p2pErrors maps BNC P2P response codes to gateway errors.
var p2pErrors = map[string]GatewayError{
	"G05": {Code: "BANK_COMM_ERROR", Message: "Communication problem with the bank, try again"},
	"G12": {Code: "BANK_COMM_ERROR", Message: "Communication problem with the bank, try again"},
	"G14": {Code: "BENEFICIARY_NOT_ENROLLED", Message: "The beneficiary may not be enrolled in mobile payments"},
	"G41": {Code: "BENEFICIARY_NOT_ENROLLED", Message: "The beneficiary may not be enrolled in mobile payments"},
	"G43": {Code: "BENEFICIARY_NOT_ENROLLED", Message: "The beneficiary may not be enrolled in mobile payments"},
	"G51": {Code: "INSUFFICIENT_FUNDS", Message: "Insufficient funds in the merchant account"},
	"G52": {Code: "BENEFICIARY_NOT_ENROLLED", Message: "The beneficiary is not enrolled in Pago Móvil"},
	"G61": {Code: "DAILY_LIMIT_EXCEEDED", Message: "Daily amount limit has been exceeded"},
	"G62": {Code: "BENEFICIARY_RESTRICTED", Message: "The beneficiary is restricted at their bank"},
	"G65": {Code: "DAILY_COUNT_EXCEEDED", Message: "Daily transaction count limit has been exceeded"},
	"G80": {Code: "INVALID_BENEFICIARY_ID", Message: "The beneficiary ID is invalid"},
	"G91": {Code: "BANK_UNAVAILABLE", Message: "Communication problem with the bank, try again"},
	"G96": {Code: "BANK_COMM_ERROR", Message: "Communication problem with the bank, try again"},
}

// apiErrors maps BNC general API error codes to gateway errors.
var apiErrors = map[string]GatewayError{
	"EPIRWK":  {Code: "SESSION_EXPIRED", Message: "Bank session expired, will retry automatically"},
	"EPICNF":  {Code: "MERCHANT_NOT_FOUND", Message: "Merchant not found or inactive at the bank"},
	"EPIIMS":  {Code: "VALIDATION_FAILED", Message: "Request validation failed at the bank"},
	"EPIHV":   {Code: "INTEGRITY_ERROR", Message: "Request integrity check failed"},
	"EPIECP":  {Code: "BANK_INTERNAL_ERROR", Message: "Internal error at the bank processing C2P"},
	"EPIONA":  {Code: "NO_PERMISSION", Message: "Merchant does not have permission for this operation"},
	"EPIMC1":  {Code: "NO_CHILD_PERMISSION", Message: "Merchant does not have child client permissions"},
	"EPIMC2":  {Code: "CHILD_NOT_FOUND", Message: "Child commerce not found at the bank"},
	"EPIANF":  {Code: "ACCOUNTS_NOT_FOUND", Message: "Merchant accounts not found at the bank"},
}

// MapC2PError converts a BNC C2P error code to a gateway-standard error.
// Returns a generic error if the code is not recognized.
func MapC2PError(bncCode string) GatewayError {
	if e, ok := c2pErrors[bncCode]; ok {
		return e
	}
	return GatewayError{Code: "UNKNOWN_BANK_ERROR", Message: "An unexpected error occurred at the bank"}
}

// MapAPIError converts a BNC API error code to a gateway-standard error.
func MapAPIError(bncCode string) GatewayError {
	if e, ok := apiErrors[bncCode]; ok {
		return e
	}
	return GatewayError{Code: "UNKNOWN_BANK_ERROR", Message: "An unexpected error occurred at the bank"}
}

// IsWorkingKeyExpired checks if the BNC error code indicates the WorkingKey
// needs to be refreshed. Returns true for EPIRWK.
func IsWorkingKeyExpired(bncCode string) bool {
	return bncCode == "EPIRWK"
}

// IsNetworkError checks if the BNC error code indicates a transient
// network/communication issue that is safe to retry.
func IsNetworkError(bncCode string) bool {
	return bncCode == "G05" || bncCode == "G12" || bncCode == "G91" || bncCode == "G96" || bncCode == "EPIEDD"
}

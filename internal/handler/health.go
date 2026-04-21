// Package handler implements the HTTP handlers for the payment gateway API.
// Each handler is responsible for request parsing, validation, and response
// formatting. Business logic lives in the service layer.
package handler

import (
	"encoding/json"
	"net/http"
)

// Health handles GET /health requests.
// Returns 200 with a JSON body indicating the service is operational.
// Used by Cloud Run health checks and load balancers.
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "payment-gateway",
	})
}

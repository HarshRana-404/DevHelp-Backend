// Package dto defines shared Data Transfer Objects used across all DevHelp tool handlers.
// DTOs are pure data structures with no business logic; they form the API contract boundary.
package dto

// APIResponse is the standard envelope returned by every DevHelp endpoint.
// Success responses set Success=true and populate Data.
// Error responses set Success=false and populate Error.
type APIResponse struct {
	Success   bool        `json:"success"`
	RequestID string      `json:"request_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
}

// APIError carries structured error information.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// HealthResponse is returned by the /api/v1/health endpoint.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Env     string `json:"env"`
}

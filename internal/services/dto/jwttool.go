package dto

import "time"

// JWTGenerateRequest is the input payload for generating a JWT.
type JWTGenerateRequest struct {
	// Secret is the HMAC-SHA256 signing key.
	Secret string `json:"secret" binding:"required" example:"my-super-secret"`
	// Claims are the custom payload claims to embed in the token.
	Claims map[string]interface{} `json:"claims" binding:"required"`
	// ExpiresInSeconds is the token TTL in seconds. Defaults to 3600 if zero.
	ExpiresInSeconds int64 `json:"expires_in_seconds,omitempty" example:"3600"`
}

// JWTDecodeRequest is the input payload for decoding/inspecting a JWT.
type JWTDecodeRequest struct {
	// Token is the full JWT string (header.payload.signature).
	Token string `json:"token" binding:"required"`
	// Secret is optional. When provided the signature is verified.
	Secret string `json:"secret,omitempty"`
}

// JWTGenerateResponse is returned after successfully generating a token.
type JWTGenerateResponse struct {
	Token string `json:"token"`
}

// JWTDecodeResponse is the detailed breakdown of a JWT.
type JWTDecodeResponse struct {
	Header    map[string]interface{} `json:"header"`
	Payload   map[string]interface{} `json:"payload"`
	Signature string                 `json:"signature"`
	IssuedAt  *time.Time             `json:"issued_at,omitempty"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty"`
	NotBefore *time.Time             `json:"not_before,omitempty"`
	Valid     bool                   `json:"valid"`
	Error     string                 `json:"error,omitempty"`
}

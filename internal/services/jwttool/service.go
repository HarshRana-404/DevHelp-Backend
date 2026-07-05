// Package jwttool implements the JWT Toolkit.
// It provides local JWT generation and decoding using HS256.
// No remote key verification is performed — everything is done in-memory.
package jwttool

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"devhelp/internal/services/dto"

	"github.com/golang-jwt/jwt/v5"
)

// Service defines the JWT Toolkit contract.
type Service interface {
	// Generate signs and returns a new JWT using the provided claims and secret.
	Generate(req dto.JWTGenerateRequest) (*dto.JWTGenerateResponse, error)
	// Decode parses and decodes a JWT, returning its structural breakdown.
	Decode(req dto.JWTDecodeRequest) (*dto.JWTDecodeResponse, error)
}

type service struct{}

// NewService constructs and returns a new JWT Toolkit Service.
func NewService() Service {
	return &service{}
}

// Generate creates a signed HS256 JWT with the provided claims.
func (s *service) Generate(req dto.JWTGenerateRequest) (*dto.JWTGenerateResponse, error) {
	ttl := req.ExpiresInSeconds
	if ttl <= 0 {
		ttl = 3600
	}

	now := time.Now()
	claims := jwt.MapClaims{}

	// Copy user-provided claims first, then stamp standard temporal claims.
	for k, v := range req.Claims {
		claims[k] = v
	}
	claims["iat"] = now.Unix()
	claims["exp"] = now.Add(time.Duration(ttl) * time.Second).Unix()
	claims["nbf"] = now.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(req.Secret))
	if err != nil {
		return nil, fmt.Errorf("jwttool: signing token: %w", err)
	}

	return &dto.JWTGenerateResponse{Token: signed}, nil
}

// Decode parses the JWT string and returns a structured breakdown.
// When a secret is provided the signature is verified; otherwise only the structure is decoded.
func (s *service) Decode(req dto.JWTDecodeRequest) (*dto.JWTDecodeResponse, error) {
	parts := strings.Split(req.Token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("jwttool: token must have three dot-separated parts")
	}

	resp := &dto.JWTDecodeResponse{
		Signature: parts[2],
	}

	// Decode header.
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("jwttool: decoding header: %w", err)
	}
	if err := json.Unmarshal(headerJSON, &resp.Header); err != nil {
		return nil, fmt.Errorf("jwttool: parsing header JSON: %w", err)
	}

	// Decode payload.
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("jwttool: decoding payload: %w", err)
	}
	if err := json.Unmarshal(payloadJSON, &resp.Payload); err != nil {
		return nil, fmt.Errorf("jwttool: parsing payload JSON: %w", err)
	}

	// Extract standard time claims.
	resp.IssuedAt = extractTime(resp.Payload, "iat")
	resp.ExpiresAt = extractTime(resp.Payload, "exp")
	resp.NotBefore = extractTime(resp.Payload, "nbf")

	// Optionally verify signature.
	if req.Secret != "" {
		_, err := jwt.Parse(req.Token,
			func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(req.Secret), nil
			},
		)
		if err != nil {
			resp.Valid = false
			resp.Error = err.Error()
		} else {
			resp.Valid = true
		}
	}

	return resp, nil
}

// extractTime reads a numeric Unix timestamp from a claims map and returns a *time.Time.
func extractTime(payload map[string]interface{}, key string) *time.Time {
	v, ok := payload[key]
	if !ok {
		return nil
	}
	switch t := v.(type) {
	case float64:
		ts := time.Unix(int64(t), 0).UTC()
		return &ts
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			return nil
		}
		ts := time.Unix(n, 0).UTC()
		return &ts
	}
	return nil
}

package tests

import (
	"testing"

	"devhelp/internal/services/dto"
	"devhelp/internal/services/jwttool"
)

func TestJWT_GenerateAndDecode(t *testing.T) {
	svc := jwttool.NewService()

	genResp, err := svc.Generate(dto.JWTGenerateRequest{
		Secret:           "test-secret",
		Claims:           map[string]interface{}{"sub": "user123", "role": "admin"},
		ExpiresInSeconds: 3600,
	})
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}
	if genResp.Token == "" {
		t.Fatal("expected non-empty token")
	}

	decResp, err := svc.Decode(dto.JWTDecodeRequest{
		Token:  genResp.Token,
		Secret: "test-secret",
	})
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if !decResp.Valid {
		t.Errorf("expected valid token, got error: %s", decResp.Error)
	}
	if decResp.Payload["sub"] != "user123" {
		t.Errorf("expected sub=user123, got %v", decResp.Payload["sub"])
	}
	if decResp.ExpiresAt == nil {
		t.Error("expected ExpiresAt to be set")
	}
	if decResp.IssuedAt == nil {
		t.Error("expected IssuedAt to be set")
	}
}

func TestJWT_DecodeWithWrongSecret(t *testing.T) {
	svc := jwttool.NewService()

	genResp, err := svc.Generate(dto.JWTGenerateRequest{
		Secret: "real-secret",
		Claims: map[string]interface{}{"sub": "user1"},
	})
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}

	decResp, err := svc.Decode(dto.JWTDecodeRequest{
		Token:  genResp.Token,
		Secret: "wrong-secret",
	})
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if decResp.Valid {
		t.Error("expected token to be invalid with wrong secret")
	}
}

func TestJWT_DecodeWithoutVerification(t *testing.T) {
	svc := jwttool.NewService()

	// Decode without providing a secret — should succeed structurally.
	genResp, _ := svc.Generate(dto.JWTGenerateRequest{
		Secret: "any-secret",
		Claims: map[string]interface{}{"sub": "user2"},
	})

	decResp, err := svc.Decode(dto.JWTDecodeRequest{Token: genResp.Token})
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if decResp.Header == nil {
		t.Error("expected header to be decoded")
	}
	if decResp.Payload["sub"] != "user2" {
		t.Errorf("expected sub=user2, got %v", decResp.Payload["sub"])
	}
}

func TestJWT_InvalidToken(t *testing.T) {
	svc := jwttool.NewService()

	_, err := svc.Decode(dto.JWTDecodeRequest{Token: "not.a.valid.jwt.token"})
	if err == nil {
		t.Error("expected error for malformed token")
	}
}

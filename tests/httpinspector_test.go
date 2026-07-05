package tests

import (
	"testing"

	"devhelp/internal/services/dto"
	"devhelp/internal/services/httpinspector"
)

func TestHTTPInspector_RawHTTP(t *testing.T) {
	svc := httpinspector.NewService()

	raw := "GET /search?q=golang&page=2 HTTP/1.1\r\nHost: example.com\r\nAuthorization: Bearer token123\r\nCookie: session=abc; user=john\r\n\r\n"
	result, err := svc.Inspect(dto.HTTPInspectorRequest{RawHTTP: raw})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Method != "GET" {
		t.Errorf("expected method GET, got %q", result.Method)
	}
	if result.Path != "/search" {
		t.Errorf("expected path /search, got %q", result.Path)
	}
	if result.QueryParams["q"] != "golang" {
		t.Errorf("expected query param q=golang, got %q", result.QueryParams["q"])
	}
	if result.Authorization.Type != "Bearer" {
		t.Errorf("expected auth type Bearer, got %q", result.Authorization.Type)
	}
	if result.Cookies["session"] != "abc" {
		t.Errorf("expected cookie session=abc, got %q", result.Cookies["session"])
	}
}

func TestHTTPInspector_CurlCommand(t *testing.T) {
	svc := httpinspector.NewService()

	curl := `curl -X POST https://api.example.com/users -H 'Content-Type: application/json' -d '{"name":"alice"}'`
	result, err := svc.Inspect(dto.HTTPInspectorRequest{CurlCommand: curl})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Method != "POST" {
		t.Errorf("expected method POST, got %q", result.Method)
	}
	if result.ParsedJSON == nil {
		t.Error("expected ParsedJSON to be non-nil for JSON body")
	}
}

func TestHTTPInspector_EmptyInput(t *testing.T) {
	svc := httpinspector.NewService()

	_, err := svc.Inspect(dto.HTTPInspectorRequest{})
	if err == nil {
		t.Error("expected error for empty input")
	}
}

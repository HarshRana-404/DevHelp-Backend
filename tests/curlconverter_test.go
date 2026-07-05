package tests

import (
	"strings"
	"testing"

	"devhelp/internal/services/curlconverter"
	"devhelp/internal/services/dto"
)

func TestCurlConverter_AllLanguages(t *testing.T) {
	svc := curlconverter.NewService()

	req := dto.CurlConverterRequest{
		CurlCommand: `curl -X POST https://api.example.com/data -H 'Content-Type: application/json' -d '{"key":"value"}'`,
	}

	result, err := svc.Convert(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedLangs := []string{"go", "python", "java", "c", "c++", "ruby", "javascript", "kotlin"}
	for _, lang := range expectedLangs {
		snippet, ok := result.Snippets[lang]
		if !ok {
			t.Errorf("missing snippet for language %q", lang)
			continue
		}
		if len(snippet) == 0 {
			t.Errorf("empty snippet for language %q", lang)
		}
	}
}

func TestCurlConverter_GoSnippet(t *testing.T) {
	svc := curlconverter.NewService()

	req := dto.CurlConverterRequest{
		CurlCommand: `curl https://example.com/api`,
		Languages:   []string{"go"},
	}

	result, err := svc.Convert(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	goCode := result.Snippets["go"]
	if !strings.Contains(goCode, "http.NewRequest") {
		t.Errorf("Go snippet should contain http.NewRequest, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "example.com/api") {
		t.Errorf("Go snippet should contain the URL")
	}
}

func TestCurlConverter_JSONRepresentation(t *testing.T) {
	svc := curlconverter.NewService()

	req := dto.CurlConverterRequest{
		CurlCommand: `curl -X DELETE https://api.example.com/items/1`,
	}

	result, err := svc.Convert(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.JSON == nil {
		t.Fatal("expected JSON representation to be non-nil")
	}
	if result.JSON.Method != "DELETE" {
		t.Errorf("expected method DELETE, got %q", result.JSON.Method)
	}
}

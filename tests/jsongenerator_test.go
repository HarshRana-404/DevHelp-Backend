package tests

import (
	"strings"
	"testing"

	"devhelp/internal/services/dto"
	"devhelp/internal/services/jsongenerator"
)

func TestJSONGenerator_SimpleObject(t *testing.T) {
	svc := jsongenerator.NewService()

	result, err := svc.Generate(dto.JSONGeneratorRequest{
		JSON:     `{"name":"Alice","age":30,"active":true}`,
		RootName: "User",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.GoStruct, "type User struct") {
		t.Errorf("expected Go struct named User\n%s", result.GoStruct)
	}
	if !strings.Contains(result.GoStruct, "Name string") {
		t.Errorf("expected Name string field\n%s", result.GoStruct)
	}
	if !strings.Contains(result.TypeScriptInterface, "interface User") {
		t.Errorf("expected TypeScript interface named User\n%s", result.TypeScriptInterface)
	}
}

func TestJSONGenerator_NestedObject(t *testing.T) {
	svc := jsongenerator.NewService()

	result, err := svc.Generate(dto.JSONGeneratorRequest{
		JSON:     `{"user":{"name":"Bob","address":{"city":"NYC","zip":"10001"}}}`,
		RootName: "Response",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should produce nested struct definitions.
	if !strings.Contains(result.GoStruct, "type Response struct") {
		t.Errorf("missing root struct\n%s", result.GoStruct)
	}
	if !strings.Contains(result.GoStruct, "type User struct") {
		t.Errorf("missing nested User struct\n%s", result.GoStruct)
	}
}

func TestJSONGenerator_ArrayField(t *testing.T) {
	svc := jsongenerator.NewService()

	result, err := svc.Generate(dto.JSONGeneratorRequest{
		JSON: `{"tags":["go","api","rest"],"scores":[1,2,3]}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.GoStruct, "[]string") {
		t.Errorf("expected []string for tags field\n%s", result.GoStruct)
	}
}

func TestJSONGenerator_InvalidJSON(t *testing.T) {
	svc := jsongenerator.NewService()

	_, err := svc.Generate(dto.JSONGeneratorRequest{JSON: `{invalid json`})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestJSONGenerator_NullableField(t *testing.T) {
	svc := jsongenerator.NewService()

	result, err := svc.Generate(dto.JSONGeneratorRequest{
		JSON: `{"name":"Alice","nickname":null}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.GoStruct, "*interface{}") {
		t.Errorf("expected nullable field as *interface{}\n%s", result.GoStruct)
	}
}

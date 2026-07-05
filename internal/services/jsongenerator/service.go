// Package jsongenerator implements the JSON → Code Generator tool.
// Given a JSON input it generates Go structs and TypeScript interfaces,
// correctly handling nested objects, arrays, and nullable values.
package jsongenerator

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"devhelp/internal/services/dto"
)

// Service defines the JSON Generator contract.
type Service interface {
	// Generate parses the JSON string and returns generated Go and TypeScript type definitions.
	Generate(req dto.JSONGeneratorRequest) (*dto.JSONGeneratorResponse, error)
}

type service struct{}

// NewService constructs and returns a new JSON Generator Service.
func NewService() Service {
	return &service{}
}

// Generate is the main entry point. It unmarshals the JSON, builds a type tree,
// and emits Go struct and TypeScript interface source code.
func (s *service) Generate(req dto.JSONGeneratorRequest) (*dto.JSONGeneratorResponse, error) {
	var raw interface{}
	if err := json.Unmarshal([]byte(req.JSON), &raw); err != nil {
		return nil, fmt.Errorf("jsongenerator: invalid JSON: %w", err)
	}

	rootName := strings.TrimSpace(req.RootName)
	if rootName == "" {
		rootName = "Root"
	}
	rootName = toPascalCase(rootName)

	var goStructs []string
	var tsInterfaces []string

	buildGoStructs(raw, rootName, &goStructs)
	buildTSInterfaces(raw, rootName, &tsInterfaces)

	// Reverse so the root type is first in the output.
	reverseStrings(goStructs)
	reverseStrings(tsInterfaces)

	return &dto.JSONGeneratorResponse{
		GoStruct:            strings.Join(goStructs, "\n\n"),
		TypeScriptInterface: strings.Join(tsInterfaces, "\n\n"),
	}, nil
}

// ── Go Struct Generation ─────────────────────────────────────────────────────

func buildGoStructs(v interface{}, name string, out *[]string) {
	obj, ok := v.(map[string]interface{})
	if !ok {
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s struct {\n", name))

	keys := sortedKeys(obj)
	for _, k := range keys {
		val := obj[k]
		fieldName := toPascalCase(k)
		goType, nested := resolveGoType(val, fieldName)
		sb.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", fieldName, goType, k))
		if nested != nil {
			buildGoStructs(nested, fieldName, out)
		}
	}
	sb.WriteString("}")
	*out = append(*out, sb.String())
}

func resolveGoType(v interface{}, typeName string) (string, interface{}) {
	if v == nil {
		return "*interface{}", nil
	}
	switch val := v.(type) {
	case bool:
		return "bool", nil
	case float64:
		if val == float64(int64(val)) {
			return "int64", nil
		}
		return "float64", nil
	case string:
		return "string", nil
	case map[string]interface{}:
		return typeName, val
	case []interface{}:
		if len(val) == 0 {
			return "[]interface{}", nil
		}
		elemType, nested := resolveGoType(val[0], typeName+"Item")
		return "[]" + elemType, nested
	default:
		return "interface{}", nil
	}
}

// ── TypeScript Interface Generation ─────────────────────────────────────────

func buildTSInterfaces(v interface{}, name string, out *[]string) {
	obj, ok := v.(map[string]interface{})
	if !ok {
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("interface %s {\n", name))

	keys := sortedKeys(obj)
	for _, k := range keys {
		val := obj[k]
		fieldName := toCamelCase(k)
		nullable := val == nil
		tsType, nested := resolveTSType(val, toPascalCase(k))
		if nullable {
			sb.WriteString(fmt.Sprintf("  %s?: %s | null;\n", fieldName, tsType))
		} else {
			sb.WriteString(fmt.Sprintf("  %s: %s;\n", fieldName, tsType))
		}
		if nested != nil {
			buildTSInterfaces(nested, toPascalCase(k), out)
		}
	}
	sb.WriteString("}")
	*out = append(*out, sb.String())
}

func resolveTSType(v interface{}, typeName string) (string, interface{}) {
	if v == nil {
		return "unknown", nil
	}
	switch val := v.(type) {
	case bool:
		return "boolean", nil
	case float64:
		return "number", nil
	case string:
		return "string", nil
	case map[string]interface{}:
		return typeName, val
	case []interface{}:
		if len(val) == 0 {
			return "unknown[]", nil
		}
		elemType, nested := resolveTSType(val[0], typeName+"Item")
		return elemType + "[]", nested
	default:
		return "unknown", nil
	}
}

// ── Utilities ────────────────────────────────────────────────────────────────

// toPascalCase converts "some_field_name" or "someFieldName" to "SomeFieldName".
func toPascalCase(s string) string {
	words := splitWords(s)
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		runes := []rune(w)
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, "")
}

// toCamelCase converts "some_field_name" or "SomeFieldName" to "someFieldName".
func toCamelCase(s string) string {
	words := splitWords(s)
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		runes := []rune(w)
		if i == 0 {
			runes[0] = unicode.ToLower(runes[0])
		} else {
			runes[0] = unicode.ToUpper(runes[0])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, "")
}

// splitWords splits a string on underscores, hyphens, and camelCase boundaries.
func splitWords(s string) []string {
	var parts []string
	s = strings.ReplaceAll(s, "-", "_")
	for _, part := range strings.Split(s, "_") {
		// Further split on camelCase transitions.
		var current strings.Builder
		runes := []rune(part)
		for i, r := range runes {
			if i > 0 && unicode.IsUpper(r) && !unicode.IsUpper(runes[i-1]) {
				parts = append(parts, current.String())
				current.Reset()
			}
			current.WriteRune(unicode.ToLower(r))
		}
		if current.Len() > 0 {
			parts = append(parts, current.String())
		}
	}
	return parts
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func reverseStrings(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

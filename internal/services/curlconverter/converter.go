// Package curlconverter implements the cURL ↔ Language Converter tool.
// The converter interface allows any number of language converters to be registered
// without modifying the core service — adding a new language is one file.
package curlconverter

import "devhelp/internal/services/dto"

// Converter is the interface every language generator must implement.
// Name returns the canonical lowercase language key (e.g. "go", "python").
// Convert accepts a parsed cURL and returns the generated code string.
type Converter interface {
	Name() string
	Convert(parsed *dto.ParsedCurl) string
}

package curlconverter

import (
	"fmt"
	"strings"

	"devhelp/internal/services/dto"
	"devhelp/internal/utils"
)

// Service defines the cURL Converter contract.
type Service interface {
	// Convert parses the cURL command and generates code snippets for requested languages.
	Convert(req dto.CurlConverterRequest) (*dto.CurlConverterResponse, error)
	// SupportedLanguages returns the list of language keys the service can generate.
	SupportedLanguages() []string
}

type service struct {
	converters map[string]Converter
}

// NewService builds a Service pre-loaded with all built-in language converters.
// Additional converters can be injected via the converters variadic parameter.
func NewService(extra ...Converter) Service {
	builtin := []Converter{
		&goConverter{},
		&pythonConverter{},
		&javaConverter{},
		&cConverter{},
		&cppConverter{},
		&rubyConverter{},
		&jsConverter{},
		&kotlinConverter{},
	}

	m := make(map[string]Converter, len(builtin)+len(extra))
	for _, c := range append(builtin, extra...) {
		m[strings.ToLower(c.Name())] = c
	}

	return &service{converters: m}
}

// SupportedLanguages returns sorted language keys.
func (s *service) SupportedLanguages() []string {
	keys := make([]string, 0, len(s.converters))
	for k := range s.converters {
		keys = append(keys, k)
	}
	return keys
}

// Convert parses the cURL command and generates code for each requested language.
func (s *service) Convert(req dto.CurlConverterRequest) (*dto.CurlConverterResponse, error) {
	parsed, err := utils.ParseCurlCommand(req.CurlCommand)
	if err != nil {
		return nil, fmt.Errorf("curlconverter: %w", err)
	}

	targets := req.Languages
	if len(targets) == 0 {
		targets = s.SupportedLanguages()
	}

	snippets := make(map[string]string, len(targets))
	for _, lang := range targets {
		key := strings.ToLower(lang)
		c, ok := s.converters[key]
		if !ok {
			snippets[key] = fmt.Sprintf("// Language %q is not supported", lang)
			continue
		}
		snippets[key] = c.Convert(parsed)
	}

	return &dto.CurlConverterResponse{
		Snippets: snippets,
		JSON:     parsed,
	}, nil
}

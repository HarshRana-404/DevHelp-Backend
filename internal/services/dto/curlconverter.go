package dto

// CurlConverterRequest is the input payload for the cURL Converter tool.
type CurlConverterRequest struct {
	// CurlCommand is the full cURL command string to convert.
	CurlCommand string `json:"curl_command" binding:"required" example:"curl -X POST https://api.example.com/data -H 'Content-Type: application/json' -d '{\"key\":\"value\"}'"`
	// Languages lists the target languages to generate code for.
	// If empty, all supported languages are returned.
	Languages []string `json:"languages,omitempty" example:"[\"go\",\"python\"]"`
}

// CurlConverterResponse contains generated code snippets for each requested language
// plus a JSON representation of the parsed cURL command.
type CurlConverterResponse struct {
	// Snippets maps language name to generated code string.
	Snippets map[string]string `json:"snippets"`
	// JSON is a structured representation of the parsed cURL command.
	JSON *ParsedCurl `json:"json"`
}

// ParsedCurl is a language-agnostic, serialisable representation of a cURL command.
type ParsedCurl struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body,omitempty"`
}

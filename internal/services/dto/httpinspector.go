package dto

// HTTPInspectorRequest is the input payload for the HTTP Inspector tool.
// Callers supply either a raw HTTP request string or a cURL command string.
type HTTPInspectorRequest struct {
	// RawHTTP is an optional raw HTTP/1.1 request (e.g. "GET /path HTTP/1.1\r\nHost: …").
	RawHTTP string `json:"raw_http" example:"GET /ping?foo=bar HTTP/1.1\r\nHost: example.com\r\n\r\n"`
	// CurlCommand is an optional cURL command string to be parsed and inspected.
	CurlCommand string `json:"curl_command" example:"curl -X POST https://api.example.com/data -H 'Content-Type: application/json' -d '{\"key\":\"value\"}'"`
}

// HTTPInspectorResponse is the structured result of HTTP inspection.
type HTTPInspectorResponse struct {
	Method        string            `json:"method"`
	URL           string            `json:"url"`
	Path          string            `json:"path"`
	QueryParams   map[string]string `json:"query_params"`
	Headers       map[string]string `json:"headers"`
	Cookies       map[string]string `json:"cookies"`
	Authorization AuthInfo          `json:"authorization"`
	Body          string            `json:"body"`
	ContentType   string            `json:"content_type"`
	ParsedJSON    interface{}       `json:"parsed_json,omitempty"`
	ParsedForm    map[string]string `json:"parsed_form,omitempty"`
}

// AuthInfo holds parsed authorization header details.
type AuthInfo struct {
	Type        string `json:"type"`
	Credentials string `json:"credentials"`
}

// Package httpinspector implements the HTTP Request Inspector tool.
// It accepts either a raw HTTP/1.1 request or a cURL command and returns a structured
// breakdown: method, URL, headers, cookies, auth, body, and parsed body variants.
package httpinspector

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"devhelp/internal/services/dto"
	"devhelp/internal/utils"
)

// Service defines the HTTP Inspector contract.
type Service interface {
	// Inspect parses the input and returns a structured HTTP inspection result.
	Inspect(req dto.HTTPInspectorRequest) (*dto.HTTPInspectorResponse, error)
}

type service struct{}

// NewService constructs and returns a new HTTP Inspector Service.
func NewService() Service {
	return &service{}
}

// Inspect dispatches to the appropriate parser based on which input field is populated.
func (s *service) Inspect(req dto.HTTPInspectorRequest) (*dto.HTTPInspectorResponse, error) {
	switch {
	case req.RawHTTP != "":
		return s.parseRawHTTP(req.RawHTTP)
	case req.CurlCommand != "":
		return s.parseCurl(req.CurlCommand)
	default:
		return nil, fmt.Errorf("either raw_http or curl_command must be provided")
	}
}

// parseRawHTTP parses a raw HTTP/1.1 request string.
func (s *service) parseRawHTTP(raw string) (*dto.HTTPInspectorResponse, error) {
	// Normalise CRLF / LF to CRLF as the net/http reader expects.
	normalised := strings.ReplaceAll(raw, "\r\n", "\n")
	normalised = strings.ReplaceAll(normalised, "\n", "\r\n")

	reader := bufio.NewReader(strings.NewReader(normalised))
	r, err := http.ReadRequest(reader)
	if err != nil {
		return nil, fmt.Errorf("parsing raw HTTP request: %w", err)
	}
	defer r.Body.Close()

	path, queryParams := utils.ExtractURLParts(r.RequestURI)

	headers := make(map[string]string)
	for k, v := range r.Header {
		headers[k] = strings.Join(v, ", ")
	}

	cookies := make(map[string]string)
	for _, c := range r.Cookies() {
		cookies[c.Name] = c.Value
	}

	auth := parseAuthHeader(r.Header.Get("Authorization"))
	contentType := r.Header.Get("Content-Type")

	var bodyStr string
	if r.Body != nil {
		bodyBytes, readErr := io.ReadAll(r.Body)
		if readErr == nil {
			bodyStr = string(bodyBytes)
		}
	}

	rawURL := r.Host + r.RequestURI
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "http://" + rawURL
	}

	resp := &dto.HTTPInspectorResponse{
		Method:        r.Method,
		URL:           rawURL,
		Path:          path,
		QueryParams:   queryParams,
		Headers:       headers,
		Cookies:       cookies,
		Authorization: auth,
		Body:          bodyStr,
		ContentType:   contentType,
	}

	enrichBodyParsing(resp, bodyStr, contentType)

	return resp, nil
}

// parseCurl parses a cURL command string via the shared curl parser utility.
func (s *service) parseCurl(curlCmd string) (*dto.HTTPInspectorResponse, error) {
	parsed, err := utils.ParseCurlCommand(curlCmd)
	if err != nil {
		return nil, fmt.Errorf("parsing curl command: %w", err)
	}

	path, queryParams := utils.ExtractURLParts(parsed.URL)

	auth := parseAuthHeader(parsed.Headers["Authorization"])
	contentType := parsed.Headers["Content-Type"]

	resp := &dto.HTTPInspectorResponse{
		Method:        parsed.Method,
		URL:           parsed.URL,
		Path:          path,
		QueryParams:   queryParams,
		Headers:       parsed.Headers,
		Cookies:       extractCookiesFromHeader(parsed.Headers),
		Authorization: auth,
		Body:          parsed.Body,
		ContentType:   contentType,
	}

	enrichBodyParsing(resp, parsed.Body, contentType)

	return resp, nil
}

// enrichBodyParsing attempts to parse the body as JSON or form-encoded data.
func enrichBodyParsing(resp *dto.HTTPInspectorResponse, body, contentType string) {
	if body == "" {
		return
	}

	ct := strings.ToLower(contentType)

	switch {
	case strings.Contains(ct, "application/json") || isJSON(body):
		var parsed interface{}
		if err := json.Unmarshal([]byte(body), &parsed); err == nil {
			resp.ParsedJSON = parsed
		}

	case strings.Contains(ct, "application/x-www-form-urlencoded"):
		values, err := url.ParseQuery(body)
		if err == nil {
			form := make(map[string]string, len(values))
			for k, v := range values {
				form[k] = strings.Join(v, ", ")
			}
			resp.ParsedForm = form
		}
	}
}

// isJSON returns true if the string appears to be valid JSON.
func isJSON(s string) bool {
	s = strings.TrimSpace(s)
	return (strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[")) &&
		json.Valid([]byte(s))
}

// parseAuthHeader splits an Authorization header into type and credentials.
func parseAuthHeader(header string) dto.AuthInfo {
	if header == "" {
		return dto.AuthInfo{}
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 {
		return dto.AuthInfo{Type: parts[0], Credentials: parts[1]}
	}
	return dto.AuthInfo{Type: header}
}

// extractCookiesFromHeader parses the Cookie header from a headers map.
func extractCookiesFromHeader(headers map[string]string) map[string]string {
	cookies := make(map[string]string)
	cookieHeader, ok := headers["Cookie"]
	if !ok {
		return cookies
	}
	for _, pair := range strings.Split(cookieHeader, ";") {
		k, v, ok := strings.Cut(strings.TrimSpace(pair), "=")
		if ok {
			cookies[k] = v
		}
	}
	return cookies
}

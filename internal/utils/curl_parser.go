package utils

import (
	"devhelp/internal/services/dto"
	"net/url"
	"strings"
)

// ParseCurlCommand parses a cURL command string into a ParsedCurl DTO.
// It handles -X/--request, -H/--header, -d/--data, and the URL positional argument.
func ParseCurlCommand(cmd string) (*dto.ParsedCurl, error) {
	tokens := tokenizeCurl(cmd)

	parsed := &dto.ParsedCurl{
		Method:  "GET",
		Headers: make(map[string]string),
	}

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		switch {
		case tok == "curl":
			continue

		case tok == "-X" || tok == "--request":
			if i+1 < len(tokens) {
				i++
				parsed.Method = strings.ToUpper(tokens[i])
			}

		case tok == "-H" || tok == "--header":
			if i+1 < len(tokens) {
				i++
				key, val, ok := strings.Cut(tokens[i], ":")
				if ok {
					parsed.Headers[strings.TrimSpace(key)] = strings.TrimSpace(val)
				}
			}

		case tok == "-d" || tok == "--data" || tok == "--data-raw" || tok == "--data-binary":
			if i+1 < len(tokens) {
				i++
				parsed.Body = tokens[i]
				if parsed.Method == "GET" {
					parsed.Method = "POST"
				}
			}

		case !strings.HasPrefix(tok, "-"):
			// Treat first non-flag token as the URL.
			if parsed.URL == "" {
				parsed.URL = tok
			}
		}
	}

	return parsed, nil
}

// tokenizeCurl splits a cURL command string into tokens, respecting single and double quotes.
func tokenizeCurl(cmd string) []string {
	var tokens []string
	var current strings.Builder
	inSingle := false
	inDouble := false

	for i := 0; i < len(cmd); i++ {
		ch := cmd[i]
		switch {
		case ch == '\'' && !inDouble:
			inSingle = !inSingle
		case ch == '"' && !inSingle:
			inDouble = !inDouble
		case (ch == ' ' || ch == '\t' || ch == '\n') && !inSingle && !inDouble:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			// Handle escaped characters inside double quotes.
			if inDouble && ch == '\\' && i+1 < len(cmd) {
				next := cmd[i+1]
				if next == '"' || next == '\\' || next == '\'' {
					current.WriteByte(next)
					i++
					continue
				}
			}
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

// ExtractURLParts decomposes a URL string into path and query param map.
func ExtractURLParts(rawURL string) (path string, queryParams map[string]string) {
	queryParams = make(map[string]string)

	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, queryParams
	}

	path = u.Path
	for k, v := range u.Query() {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}
	return path, queryParams
}

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"devhelp/internal/config"
	"devhelp/internal/handlers"
	"devhelp/internal/middleware"
	"devhelp/internal/services/curlconverter"
	"devhelp/internal/services/httpinspector"
	"devhelp/internal/services/jsongenerator"
	"devhelp/internal/services/jwttool"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestRouter assembles a minimal router with all middleware and handlers,
// matching production wiring without requiring a real config file.
func newTestRouter() *gin.Engine {
	cfg := &config.Config{
		App: config.AppConfig{
			Version:        "test",
			Env:            "development",
			TimeoutSeconds: 30,
		},
		RateLimiter: config.RateLimiterConfig{RequestsPerMinute: 1000},
		CORS:        config.CORSConfig{AllowedOrigins: []string{"*"}},
		Log:         config.LogConfig{Level: "debug"},
	}

	r := gin.New()
	r.Use(middleware.Recovery(newNopLogger()))
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))
	r.Use(middleware.RateLimiter(cfg.RateLimiter.RequestsPerMinute))
	r.Use(middleware.Timeout(time.Duration(cfg.App.TimeoutSeconds) * time.Second))

	healthH := handlers.NewHealthHandler(cfg)
	httpInspH := handlers.NewHTTPInspectorHandler(httpinspector.NewService())
	curlH := handlers.NewCurlConverterHandler(curlconverter.NewService())
	jwtH := handlers.NewJWTHandler(jwttool.NewService())
	jsonH := handlers.NewJSONGeneratorHandler(jsongenerator.NewService())

	v1 := r.Group("/api/v1")
	v1.GET("/health", healthH.Check)
	v1.POST("/http-inspector", httpInspH.Inspect)
	v1.POST("/curl-converter", curlH.Convert)
	v1.GET("/curl-converter/languages", curlH.Languages)
	v1.POST("/jwt/generate", jwtH.Generate)
	v1.POST("/jwt/decode", jwtH.Decode)
	v1.POST("/json-generator", jsonH.Generate)

	return r
}

// ── Health ────────────────────────────────────────────────────────────────────

func TestHandler_Health(t *testing.T) {
	r := newTestRouter()
	w := doRequest(t, r, http.MethodGet, "/api/v1/health", nil)

	assertStatus(t, w, http.StatusOK)
	body := parseBody(t, w)
	assertBool(t, "success", body["success"], true)

	data := body["data"].(map[string]interface{})
	if data["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", data["status"])
	}
}

func TestHandler_Health_RequestIDHeader(t *testing.T) {
	r := newTestRouter()
	w := doRequest(t, r, http.MethodGet, "/api/v1/health", nil)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID response header to be set")
	}
}

func TestHandler_Health_ClientRequestIDEchoed(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("X-Request-ID", "my-custom-id-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != "my-custom-id-123" {
		t.Errorf("expected echoed X-Request-ID, got %q", w.Header().Get("X-Request-ID"))
	}
}

// ── HTTP Inspector ────────────────────────────────────────────────────────────

func TestHandler_HTTPInspector_RawHTTP(t *testing.T) {
	r := newTestRouter()
	payload := map[string]interface{}{
		"raw_http": "GET /ping?env=test HTTP/1.1\r\nHost: example.com\r\n\r\n",
	}
	w := doRequest(t, r, http.MethodPost, "/api/v1/http-inspector", payload)

	assertStatus(t, w, http.StatusOK)
	body := parseBody(t, w)
	assertBool(t, "success", body["success"], true)

	data := body["data"].(map[string]interface{})
	if data["method"] != "GET" {
		t.Errorf("expected method=GET, got %v", data["method"])
	}
	qp := data["query_params"].(map[string]interface{})
	if qp["env"] != "test" {
		t.Errorf("expected query param env=test, got %v", qp["env"])
	}
}

func TestHandler_HTTPInspector_EmptyBody(t *testing.T) {
	r := newTestRouter()
	w := doRequest(t, r, http.MethodPost, "/api/v1/http-inspector", map[string]interface{}{})

	assertStatus(t, w, http.StatusBadRequest)
	body := parseBody(t, w)
	assertBool(t, "success", body["success"], false)
}

func TestHandler_HTTPInspector_CurlInput(t *testing.T) {
	r := newTestRouter()
	payload := map[string]interface{}{
		"curl_command": "curl -X DELETE https://api.example.com/items/42",
	}
	w := doRequest(t, r, http.MethodPost, "/api/v1/http-inspector", payload)

	assertStatus(t, w, http.StatusOK)
	data := parseBody(t, w)["data"].(map[string]interface{})
	if data["method"] != "DELETE" {
		t.Errorf("expected method=DELETE, got %v", data["method"])
	}
}

// ── cURL Converter ────────────────────────────────────────────────────────────

func TestHandler_CurlConverter_Convert(t *testing.T) {
	r := newTestRouter()
	payload := map[string]interface{}{
		"curl_command": "curl https://example.com/api",
		"languages":    []string{"go", "python"},
	}
	w := doRequest(t, r, http.MethodPost, "/api/v1/curl-converter", payload)

	assertStatus(t, w, http.StatusOK)
	data := parseBody(t, w)["data"].(map[string]interface{})
	snippets := data["snippets"].(map[string]interface{})
	if _, ok := snippets["go"]; !ok {
		t.Error("expected go snippet in response")
	}
	if _, ok := snippets["python"]; !ok {
		t.Error("expected python snippet in response")
	}
}

func TestHandler_CurlConverter_MissingBody(t *testing.T) {
	r := newTestRouter()
	w := doRequest(t, r, http.MethodPost, "/api/v1/curl-converter", map[string]interface{}{})

	assertStatus(t, w, http.StatusBadRequest)
}

func TestHandler_CurlConverter_Languages(t *testing.T) {
	r := newTestRouter()
	w := doRequest(t, r, http.MethodGet, "/api/v1/curl-converter/languages", nil)

	assertStatus(t, w, http.StatusOK)
	data := parseBody(t, w)["data"].(map[string]interface{})
	langs, ok := data["languages"].([]interface{})
	if !ok || len(langs) == 0 {
		t.Error("expected non-empty languages list")
	}
}

// ── JWT ───────────────────────────────────────────────────────────────────────

func TestHandler_JWT_GenerateAndDecode(t *testing.T) {
	r := newTestRouter()

	// Generate
	genPayload := map[string]interface{}{
		"secret":             "handler-test-secret",
		"claims":             map[string]interface{}{"sub": "u1"},
		"expires_in_seconds": 3600,
	}
	w := doRequest(t, r, http.MethodPost, "/api/v1/jwt/generate", genPayload)
	assertStatus(t, w, http.StatusOK)
	genData := parseBody(t, w)["data"].(map[string]interface{})
	token, ok := genData["token"].(string)
	if !ok || token == "" {
		t.Fatal("expected non-empty token from generate endpoint")
	}

	// Decode
	decPayload := map[string]interface{}{
		"token":  token,
		"secret": "handler-test-secret",
	}
	w2 := doRequest(t, r, http.MethodPost, "/api/v1/jwt/decode", decPayload)
	assertStatus(t, w2, http.StatusOK)
	decData := parseBody(t, w2)["data"].(map[string]interface{})
	if decData["valid"] != true {
		t.Errorf("expected valid=true, got %v", decData["valid"])
	}
}

func TestHandler_JWT_DecodeMalformed(t *testing.T) {
	r := newTestRouter()
	payload := map[string]interface{}{"token": "bad.token"}
	w := doRequest(t, r, http.MethodPost, "/api/v1/jwt/decode", payload)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── JSON Generator ────────────────────────────────────────────────────────────

func TestHandler_JSONGenerator(t *testing.T) {
	r := newTestRouter()
	payload := map[string]interface{}{
		"json":      `{"id":1,"name":"Alice","active":true}`,
		"root_name": "User",
	}
	w := doRequest(t, r, http.MethodPost, "/api/v1/json-generator", payload)

	assertStatus(t, w, http.StatusOK)
	data := parseBody(t, w)["data"].(map[string]interface{})
	if data["go_struct"] == "" {
		t.Error("expected non-empty go_struct")
	}
	if data["typescript_interface"] == "" {
		t.Error("expected non-empty typescript_interface")
	}
}

func TestHandler_JSONGenerator_InvalidJSON(t *testing.T) {
	r := newTestRouter()
	payload := map[string]interface{}{"json": `{bad json`}
	w := doRequest(t, r, http.MethodPost, "/api/v1/json-generator", payload)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── CORS ──────────────────────────────────────────────────────────────────────

func TestMiddleware_CORS_Headers(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("expected Access-Control-Allow-Origin header")
	}
}

func TestMiddleware_CORS_Preflight(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS preflight, got %d", w.Code)
	}
}

// ── Rate Limiter ──────────────────────────────────────────────────────────────

func TestMiddleware_RateLimiter_Blocks(t *testing.T) {
	// Use a router configured with a very low limit (2 req/min) to trigger 429.
	cfg := &config.Config{
		App:         config.AppConfig{Version: "test", Env: "test", TimeoutSeconds: 30},
		RateLimiter: config.RateLimiterConfig{RequestsPerMinute: 2},
		CORS:        config.CORSConfig{AllowedOrigins: []string{"*"}},
	}
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.RateLimiter(cfg.RateLimiter.RequestsPerMinute))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	// First two requests should succeed.
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	// Third request should be rate-limited.
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w.Code)
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func doRequest(t *testing.T, r *gin.Engine, method, path string, payload interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encoding payload: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func parseBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}
	return result
}

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Errorf("expected status %d, got %d — body: %s", expected, w.Code, w.Body.String())
	}
}

func assertBool(t *testing.T, field string, got interface{}, expected bool) {
	t.Helper()
	if got != expected {
		t.Errorf("expected %s=%v, got %v", field, expected, got)
	}
}

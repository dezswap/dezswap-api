package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dezswap/dezswap-api/configs"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
)

func TestSetMiddlewares_AllowsMCPPreflightHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := app{
		engine: gin.New(),
		config: configs.ApiConfig{
			MCP: configs.ApiMCPConfig{
				Enabled:        true,
				Path:           "/mcp",
				AllowedOrigins: []string{"https://allowed.example.com"},
			},
		},
		logger: logging.New("test", configs.LogConfig{}),
	}
	app.setMiddlewares(nil)
	app.engine.POST("/mcp", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/mcp", nil)
	req.Header.Set("Origin", "https://allowed.example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "MCP-Protocol-Version,Mcp-Session-Id,Last-Event-ID")
	rec := httptest.NewRecorder()

	app.engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected preflight status %d, got %d", http.StatusNoContent, rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://allowed.example.com" {
		t.Fatalf("unexpected allow origin: %q", got)
	}
	allowHeaders := rec.Header().Values("Access-Control-Allow-Headers")
	for _, header := range []string{"Mcp-Protocol-Version", "Mcp-Session-Id", "Last-Event-Id"} {
		if !containsHeader(allowHeaders, header) {
			t.Fatalf("expected Access-Control-Allow-Headers to include %q, got %v", header, allowHeaders)
		}
	}

	req = httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Origin", "https://allowed.example.com")
	rec = httptest.NewRecorder()

	app.engine.ServeHTTP(rec, req)

	if exposeHeaders := rec.Header().Values("Access-Control-Expose-Headers"); !containsHeader(exposeHeaders, "Mcp-Session-Id") {
		t.Fatalf("expected Access-Control-Expose-Headers to include Mcp-Session-Id, got %v", exposeHeaders)
	}
}

func containsHeader(values []string, want string) bool {
	for _, value := range values {
		for part := range strings.SplitSeq(value, ",") {
			if strings.EqualFold(strings.TrimSpace(part), want) {
				return true
			}
		}
	}
	return false
}

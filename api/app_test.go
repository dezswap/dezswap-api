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
	app.setMiddlewares()
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

// Two requests to one path, two different responses: the swagger handler reads the
// whole RequestURI to pick its asset. This is the fact that keeps the route off the
// cache, whose keys are built from the path and the parameters a route declares --
// under one of those, whichever of the two was answered first is replayed as the
// other, and the UI page comes back as the spec.
func TestMountSwagger_OnePathAnswersWithTwoAssets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	mountSwagger(engine)

	get := func(uri string) (string, string) {
		t.Helper()

		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, uri, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", uri, rec.Code)
		}
		return rec.Header().Get("Content-Type"), rec.Body.String()
	}

	specType, spec := get("/swagger/index.html?doc.json")
	pageType, page := get("/swagger/index.html")

	if spec == page {
		t.Fatal("/swagger/index.html answered with the body stored for ?doc.json")
	}
	if !strings.Contains(specType, "application/json") {
		t.Fatalf("expected the spec as json, got %q", specType)
	}
	if !strings.Contains(pageType, "text/html") {
		t.Fatalf("expected the ui page as html, got %q", pageType)
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

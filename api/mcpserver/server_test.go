package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNew_FiltersIncludedOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server, err := New(gin.New(), Config{
		IncludeOperations: []string{"get_service_version", "find_routes"},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if len(server.tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(server.tools))
	}
	if _, ok := server.toolByName["get_service_version"]; !ok {
		t.Fatal("expected get_service_version tool")
	}
	if _, ok := server.toolByName["find_routes"]; !ok {
		t.Fatal("expected find_routes tool")
	}
}

func TestNew_RejectsUnknownIncludedOperation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, err := New(gin.New(), Config{
		IncludeOperations: []string{"get_service_version", "missing_operation"},
	})
	if err == nil {
		t.Fatal("expected unknown included operation error")
	}
}

func TestValidateArgs(t *testing.T) {
	findRoutes := toolSpec{
		Name: "find_routes",
		Params: []paramSpec{
			stringQuery("from", "from"),
			stringQuery("to", "to"),
			integerQuery("hopCount", "hop count", false),
		},
	}
	if err := validateArgs(findRoutes, map[string]any{}); err == nil {
		t.Fatal("expected missing from/to error")
	}
	if err := validateArgs(findRoutes, map[string]any{"from": "axpla", "unknown": "x"}); err == nil {
		t.Fatal("expected unknown argument error")
	}
	if err := validateArgs(findRoutes, map[string]any{"from": "axpla", "hopCount": 1.2}); err == nil {
		t.Fatal("expected integer type error")
	}
	if err := validateArgs(findRoutes, map[string]any{"from": "axpla", "hopCount": float64(2)}); err != nil {
		t.Fatalf("expected valid args, got %v", err)
	}

	txs := toolSpec{
		Name: "list_recent_transactions",
		Params: []paramSpec{
			stringQuery("pool", "pool"),
			stringQuery("token", "token"),
		},
	}
	if err := validateArgs(txs, map[string]any{"pool": "pool1", "token": "token1"}); err == nil {
		t.Fatal("expected mutually exclusive pool/token error")
	}

	chart := toolSpec{
		Name: "get_market_chart",
		Params: []paramSpec{
			enumPath("type", "chart type", marketChartTypeEnum),
			durationQuery(),
		},
	}
	if err := validateArgs(chart, map[string]any{"type": "price"}); err == nil {
		t.Fatal("expected market chart enum error")
	}
	if err := validateArgs(chart, map[string]any{"type": "volume", "duration": "week"}); err == nil {
		t.Fatal("expected duration enum error")
	}
	if err := validateArgs(chart, map[string]any{"type": "volume", "duration": "all"}); err != nil {
		t.Fatalf("expected valid chart args, got %v", err)
	}
}

func TestHandleTool_DispatchesToInternalRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/v1/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": "test"})
	})
	server, err := New(engine, Config{
		IncludeOperations: []string{"get_service_version"},
		ResponseMaxBytes:  1024,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := server.handleTool(context.Background(), server.toolByName["get_service_version"], &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Arguments: json.RawMessage(`{}`),
		},
	})
	if err != nil {
		t.Fatalf("handleTool() error = %v", err)
	}
	if result.IsError {
		t.Fatalf("handleTool() returned tool error: %+v", result.Content)
	}
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	text := textContent.Text
	var body map[string]string
	if err := json.Unmarshal([]byte(text), &body); err != nil {
		t.Fatalf("tool result is not JSON: %v", err)
	}
	if body["version"] != "test" {
		t.Fatalf("expected test version, got %q", body["version"])
	}
}

func TestMCPOriginValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	if err := Mount(engine, Config{
		Enabled:           true,
		Path:              "/mcp",
		AllowedOrigins:    []string{"https://allowed.example.com"},
		IncludeOperations: []string{"get_service_version"},
	}); err != nil {
		t.Fatalf("Mount() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Origin", "https://blocked.example.com")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

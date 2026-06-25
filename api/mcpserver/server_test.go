package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNew_FiltersIncludedOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server, err := New(gin.New(), Config{
		IncludeOperations: []string{"get_service_version", "find_routes"},
	}, "test")
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
	}, "test")
	if err == nil {
		t.Fatal("expected unknown included operation error")
	}
}

func TestNew_RejectsEmptyVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, err := New(gin.New(), Config{
		IncludeOperations: []string{"get_service_version"},
	}, " ")
	if err == nil {
		t.Fatal("expected empty version error")
	}
	if err.Error() != "MCP server version is required" {
		t.Fatalf("expected version required error, got %q", err.Error())
	}
}

func TestNew_UsesServiceVersionAsMCPServerInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	server, err := New(gin.New(), Config{
		IncludeOperations: []string{"get_service_version"},
	}, "v1.2.3")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	mcpServer := server.newMCPServer()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	if _, err := mcpServer.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server.Connect() error = %v", err)
	}
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() error = %v", err)
	}
	defer clientSession.Close()

	serverInfo := clientSession.InitializeResult().ServerInfo
	if serverInfo.Name != "dezswap-api" {
		t.Fatalf("expected server name dezswap-api, got %q", serverInfo.Name)
	}
	if serverInfo.Version != "v1.2.3" {
		t.Fatalf("expected MCP server version v1.2.3, got %q", serverInfo.Version)
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
	}, "test")
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
	}, "test"); err != nil {
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

func TestMCPResources_ListAndRead(t *testing.T) {
	ctx := context.Background()
	server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "v0.0.1"}, nil)
	addResources(server)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server.Connect() error = %v", err)
	}
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() error = %v", err)
	}
	defer clientSession.Close()

	resources, err := clientSession.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("ListResources() error = %v", err)
	}
	if len(resources.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources.Resources))
	}

	got := map[string]*mcp.Resource{}
	for _, resource := range resources.Resources {
		got[resource.URI] = resource
		if resource.MIMEType != resourceMIMETypeMarkdown {
			t.Fatalf("expected resource %s MIME type %s, got %s", resource.URI, resourceMIMETypeMarkdown, resource.MIMEType)
		}
	}
	for _, uri := range []string{"dezswap://service/guide", "dezswap://dex/domain-model"} {
		if _, ok := got[uri]; !ok {
			t.Fatalf("expected resource %s", uri)
		}
	}

	for _, tc := range []struct {
		uri      string
		contains []string
	}{
		{
			uri: "dezswap://service/guide",
			contains: []string{
				"read-only Dezswap analytics and discovery",
				"cannot execute swaps",
				"cannot return amount-based quotes",
				"cannot calculate slippage",
				"cannot calculate price impact",
				"hopCount argument is a maximum path length filter",
			},
		},
		{
			uri: "dezswap://dex/domain-model",
			contains: []string{
				"ibc/<hash>",
				"ibc-<hash>",
				"decimal strings",
				"raw time series with {t, v} items",
				"not a quote",
				"public transaction action values are swap, add, and remove",
			},
		},
	} {
		result, err := clientSession.ReadResource(ctx, &mcp.ReadResourceParams{URI: tc.uri})
		if err != nil {
			t.Fatalf("ReadResource(%s) error = %v", tc.uri, err)
		}
		if len(result.Contents) != 1 {
			t.Fatalf("expected one content item for %s, got %d", tc.uri, len(result.Contents))
		}
		content := result.Contents[0]
		if content.MIMEType != resourceMIMETypeMarkdown {
			t.Fatalf("expected read MIME type %s for %s, got %s", resourceMIMETypeMarkdown, tc.uri, content.MIMEType)
		}
		for _, want := range tc.contains {
			if !strings.Contains(content.Text, want) {
				t.Fatalf("expected %s to contain %q", tc.uri, want)
			}
		}
	}
}

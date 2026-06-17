package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/dezswap/dezswap-api/api/docs"
	"github.com/dezswap/dezswap-api/configs"
	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	// DefaultPath is the default streamable HTTP MCP endpoint path.
	DefaultPath             = "/mcp"
	defaultRequestTimeout   = 10 * time.Second
	defaultResponseMaxBytes = 1 << 20
)

type Config = configs.ApiMCPConfig

type Server struct {
	engine           *gin.Engine
	config           Config
	tools            []toolSpec
	toolByName       map[string]toolSpec
	requestTimeout   time.Duration
	responseMaxBytes int64
}

type toolSpec struct {
	Name        string
	Title       string
	Description string
	Method      string
	Path        string
	Params      []paramSpec
}

type paramSpec struct {
	Name        string
	In          string
	Type        string
	Description string
	Required    bool
	Enum        []string
}

// New creates an MCP server backed by the given Gin engine.
func New(engine *gin.Engine, cfg Config) (*Server, error) {
	cfg = withDefaults(cfg)
	tools := defaultTools()
	if len(cfg.IncludeOperations) > 0 {
		allowed, err := validateIncludedOperations(tools, cfg.IncludeOperations)
		if err != nil {
			return nil, err
		}
		filtered := make([]toolSpec, 0, len(tools))
		for _, tool := range tools {
			if allowed[tool.Name] {
				filtered = append(filtered, tool)
			}
		}
		tools = filtered
	}

	if err := validateSwaggerCoverage(tools); err != nil {
		return nil, err
	}

	toolByName := make(map[string]toolSpec, len(tools))
	for _, tool := range tools {
		toolByName[tool.Name] = tool
	}

	return &Server{
		engine:           engine,
		config:           cfg,
		tools:            tools,
		toolByName:       toolByName,
		requestTimeout:   time.Duration(cfg.RequestTimeoutMs) * time.Millisecond,
		responseMaxBytes: int64(cfg.ResponseMaxBytes),
	}, nil
}

// validateIncludedOperations checks the configured operation allowlist.
func validateIncludedOperations(tools []toolSpec, operations []string) (map[string]bool, error) {
	available := make(map[string]bool, len(tools))
	for _, tool := range tools {
		available[tool.Name] = true
	}
	allowed := make(map[string]bool, len(operations))
	for _, name := range operations {
		if !available[name] {
			return nil, fmt.Errorf("unknown MCP operation %q", name)
		}
		allowed[name] = true
	}
	return allowed, nil
}

// Mount attaches the MCP endpoint when it is enabled.
func Mount(engine *gin.Engine, cfg Config) error {
	if !cfg.Enabled {
		return nil
	}
	srv, err := New(engine, cfg)
	if err != nil {
		return err
	}
	srv.Mount()
	return nil
}

// Mount registers the streamable HTTP MCP handler.
func (s *Server) Mount() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "dezswap-api",
		Version: "v1",
	}, &mcp.ServerOptions{
		Instructions: "Read-only Dezswap DEX analytics and discovery tools. These tools do not execute swaps, build transactions, or provide amount-based quotes/slippage.",
	})

	for _, spec := range s.tools {
		spec := spec
		server.AddTool(&mcp.Tool{
			Name:        spec.Name,
			Title:       spec.Title,
			Description: spec.Description,
			InputSchema: inputSchema(spec),
		}, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return s.handleTool(ctx, spec, req)
		})
	}
	addResources(server)

	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return server
	}, nil)

	path := s.config.Path
	s.engine.Any(path, func(c *gin.Context) {
		if !s.originAllowed(c.Request.Header.Get("Origin")) {
			c.JSON(http.StatusForbidden, gin.H{"error": "origin not allowed"})
			return
		}
		if c.Request.Method == http.MethodOptions {
			s.setCORSHeaders(c)
			c.Status(http.StatusNoContent)
			return
		}
		s.setCORSHeaders(c)
		handler.ServeHTTP(c.Writer, c.Request)
	})
}

// handleTool validates arguments and dispatches one MCP tool call.
func (s *Server) handleTool(ctx context.Context, spec toolSpec, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := parseArgs(req)
	if err != nil {
		return toolError(err), nil
	}
	if err := validateArgs(spec, args); err != nil {
		return toolError(err), nil
	}

	callCtx := ctx
	cancel := func() {}
	if s.requestTimeout > 0 {
		callCtx, cancel = context.WithTimeout(ctx, s.requestTimeout)
	}
	defer cancel()

	type dispatchResult struct {
		body   []byte
		status int
		err    error
	}
	resultCh := make(chan dispatchResult, 1)
	go func() {
		body, status, err := s.dispatch(callCtx, spec, args)
		resultCh <- dispatchResult{body: body, status: status, err: err}
	}()

	var result dispatchResult
	select {
	case result = <-resultCh:
	case <-callCtx.Done():
		return toolError(fmt.Errorf("tool execution timed out after %s", s.requestTimeout)), nil
	}

	if result.err != nil {
		return toolError(result.err), nil
	}
	if result.status < 200 || result.status >= 300 {
		return toolError(fmt.Errorf("internal API returned status %d: %s", result.status, string(result.body))), nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(result.body)}},
	}, nil
}

// dispatch forwards a tool call to the matching internal API route.
func (s *Server) dispatch(ctx context.Context, spec toolSpec, args map[string]any) ([]byte, int, error) {
	if spec.Name == "get_market_overview" {
		return s.dispatchMarketOverview(ctx)
	}

	path := spec.Path
	query := url.Values{}
	for _, param := range spec.Params {
		value, ok := args[param.Name]
		if !ok {
			continue
		}
		switch param.In {
		case "path":
			path = strings.ReplaceAll(path, "{"+param.Name+"}", url.PathEscape(fmt.Sprint(value)))
		case "query":
			query.Set(param.Name, fmt.Sprint(value))
		}
	}
	target := "/" + strings.TrimPrefix(path, "/")
	if queryString := query.Encode(); queryString != "" {
		target += "?" + queryString
	}

	req := httptest.NewRequestWithContext(ctx, spec.Method, target, nil)
	rec := httptest.NewRecorder()
	s.engine.ServeHTTP(rec, req)

	reader := io.LimitReader(rec.Result().Body, s.responseMaxBytes+1)
	body, err := io.ReadAll(reader)
	rec.Result().Body.Close()
	if err != nil {
		return nil, rec.Code, err
	}
	if int64(len(body)) > s.responseMaxBytes {
		return nil, rec.Code, fmt.Errorf("response exceeds max size of %d bytes", s.responseMaxBytes)
	}
	return bytes.TrimSpace(body), rec.Code, nil
}

// dispatchMarketOverview combines recent and statistics dashboard responses.
func (s *Server) dispatchMarketOverview(ctx context.Context) ([]byte, int, error) {
	recentBody, recentStatus, err := s.dispatch(ctx, toolSpec{Method: http.MethodGet, Path: "/v1/dashboard/recent"}, nil)
	if err != nil || recentStatus < 200 || recentStatus >= 300 {
		return recentBody, recentStatus, err
	}
	statisticsBody, statisticsStatus, err := s.dispatch(ctx, toolSpec{Method: http.MethodGet, Path: "/v1/dashboard/statistics"}, nil)
	if err != nil || statisticsStatus < 200 || statisticsStatus >= 300 {
		return statisticsBody, statisticsStatus, err
	}
	combined := map[string]json.RawMessage{
		"recent":     recentBody,
		"statistics": statisticsBody,
	}
	body, err := json.Marshal(combined)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if int64(len(body)) > s.responseMaxBytes {
		return nil, http.StatusOK, fmt.Errorf("response exceeds max size of %d bytes", s.responseMaxBytes)
	}
	return body, http.StatusOK, nil
}

// parseArgs decodes raw MCP tool arguments.
func parseArgs(req *mcp.CallToolRequest) (map[string]any, error) {
	args := make(map[string]any)
	if req == nil || req.Params == nil || len(req.Params.Arguments) == 0 {
		return args, nil
	}
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	return args, nil
}

// validateArgs checks argument names, types, enums, and tool-specific rules.
func validateArgs(spec toolSpec, args map[string]any) error {
	params := make(map[string]paramSpec, len(spec.Params))
	for _, param := range spec.Params {
		params[param.Name] = param
	}
	for name := range args {
		if _, ok := params[name]; !ok {
			return fmt.Errorf("unknown argument %q", name)
		}
	}
	for _, param := range spec.Params {
		value, ok := args[param.Name]
		if !ok {
			if param.Required {
				return fmt.Errorf("missing required argument %q", param.Name)
			}
			continue
		}
		if err := validateType(param, value); err != nil {
			return err
		}
		if len(param.Enum) > 0 && !enumContains(param.Enum, fmt.Sprint(value)) {
			return fmt.Errorf("argument %q must be one of: %s", param.Name, strings.Join(param.Enum, ", "))
		}
	}
	if spec.Name == "list_recent_transactions" {
		if _, hasPool := args["pool"]; hasPool {
			if _, hasToken := args["token"]; hasToken {
				return errors.New("arguments \"pool\" and \"token\" are mutually exclusive")
			}
		}
	}
	if spec.Name == "find_routes" {
		if _, hasFrom := args["from"]; !hasFrom {
			if _, hasTo := args["to"]; !hasTo {
				return errors.New("one of \"from\" or \"to\" is required")
			}
		}
	}
	return nil
}

// validateType checks a value against a tool parameter type.
func validateType(param paramSpec, value any) error {
	switch param.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("argument %q must be a string", param.Name)
		}
	case "integer":
		switch v := value.(type) {
		case float64:
			if v != float64(int64(v)) {
				return fmt.Errorf("argument %q must be an integer", param.Name)
			}
		case int, int64, uint, uint64:
		default:
			return fmt.Errorf("argument %q must be an integer", param.Name)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("argument %q must be a boolean", param.Name)
		}
	}
	return nil
}

// inputSchema builds the MCP input schema for a tool.
func inputSchema(spec toolSpec) map[string]any {
	properties := map[string]any{}
	required := make([]string, 0)
	for _, param := range spec.Params {
		prop := map[string]any{
			"type":        param.Type,
			"description": param.Description,
		}
		if len(param.Enum) > 0 {
			prop["enum"] = param.Enum
		}
		properties[param.Name] = prop
		if param.Required {
			required = append(required, param.Name)
		}
	}
	schema := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		sort.Strings(required)
		schema["required"] = required
	}
	return schema
}

// originAllowed reports whether an Origin may call the MCP endpoint.
func (s *Server) originAllowed(origin string) bool {
	if origin == "" {
		return true
	}
	for _, allowed := range s.config.AllowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

// setCORSHeaders writes MCP-specific CORS response headers.
func (s *Server) setCORSHeaders(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")
	if origin != "" && s.originAllowed(origin) {
		c.Header("Access-Control-Allow-Origin", origin)
	}
	c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Accept, MCP-Protocol-Version, Mcp-Session-Id, Last-Event-ID")
	c.Header("Access-Control-Expose-Headers", "Mcp-Session-Id")
}

// toolError converts an error into an MCP tool error result.
func toolError(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
		IsError: true,
	}
}

// withDefaults fills omitted MCP config values.
func withDefaults(cfg Config) Config {
	if cfg.Path == "" {
		cfg.Path = DefaultPath
	}
	if cfg.RequestTimeoutMs <= 0 {
		cfg.RequestTimeoutMs = int(defaultRequestTimeout / time.Millisecond)
	}
	if cfg.ResponseMaxBytes <= 0 {
		cfg.ResponseMaxBytes = defaultResponseMaxBytes
	}
	return cfg
}

// validateSwaggerCoverage ensures each tool maps to a documented API route.
func validateSwaggerCoverage(tools []toolSpec) error {
	var swagger struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal([]byte(docs.SwaggerInfo.ReadDoc()), &swagger); err != nil {
		return fmt.Errorf("read swagger document: %w", err)
	}
	for _, tool := range tools {
		path := strings.TrimPrefix(tool.Path, "/v1")
		methods, ok := swagger.Paths[path]
		if !ok {
			return fmt.Errorf("tool %s references path missing from swagger: %s", tool.Name, path)
		}
		if _, ok := methods[strings.ToLower(tool.Method)]; !ok {
			return fmt.Errorf("tool %s references method missing from swagger: %s %s", tool.Name, tool.Method, path)
		}
	}
	return nil
}

// enumContains reports whether value is in values.
func enumContains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

package configs

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestApiConfig_CorsAllowedOriginsFromConfig(t *testing.T) {
	v := newTestViper(t, `
api:
  server:
    cors_allowed_origins:
      - foo\.example\.com$
      - bar\.example\.com$
`)

	cfg := apiConfig(v)

	expected := []string{`foo\.example\.com$`, `bar\.example\.com$`}
	if !reflect.DeepEqual(cfg.Server.CorsAllowedOrigins, expected) {
		t.Fatalf("expected cors origins %v, got %v", expected, cfg.Server.CorsAllowedOrigins)
	}
}

func TestApiConfig_CorsAllowedOriginsOverrideByEnv(t *testing.T) {
	const envKey = "APP_API_SERVER_CORS_ALLOWED_ORIGINS"
	if err := os.Setenv(envKey, `foo\.example\.com$,baz\.example\.com$`); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	defer os.Unsetenv(envKey)

	v := newTestViper(t, `
api:
  server:
    cors_allowed_origins:
      - default\.example\.com$
`)

	cfg := apiConfig(v)

	expected := []string{`foo\.example\.com$`, `baz\.example\.com$`}
	if !reflect.DeepEqual(cfg.Server.CorsAllowedOrigins, expected) {
		t.Fatalf("expected cors origins %v, got %v", expected, cfg.Server.CorsAllowedOrigins)
	}
}

func TestApiConfig_MCPFromConfig(t *testing.T) {
	v := newTestViper(t, `
api:
  server:
    mcp:
      enabled: true
      path: /mcp
      base_url: https://api.example.com
      allowed_origins:
        - https://app.example.com
      include_operations:
        - get_service_version
      request_timeout_ms: 3000
      response_max_bytes: 2048
`)

	cfg := apiConfig(v)

	if !cfg.MCP.Enabled {
		t.Fatal("expected MCP enabled")
	}
	if cfg.MCP.Path != "/mcp" {
		t.Fatalf("expected MCP path /mcp, got %q", cfg.MCP.Path)
	}
	if cfg.MCP.BaseURL != "https://api.example.com" {
		t.Fatalf("expected base url, got %q", cfg.MCP.BaseURL)
	}
	if !reflect.DeepEqual(cfg.MCP.AllowedOrigins, []string{"https://app.example.com"}) {
		t.Fatalf("unexpected allowed origins: %v", cfg.MCP.AllowedOrigins)
	}
	if !reflect.DeepEqual(cfg.MCP.IncludeOperations, []string{"get_service_version"}) {
		t.Fatalf("unexpected include operations: %v", cfg.MCP.IncludeOperations)
	}
	if cfg.MCP.RequestTimeoutMs != 3000 {
		t.Fatalf("expected timeout 3000, got %d", cfg.MCP.RequestTimeoutMs)
	}
	if cfg.MCP.ResponseMaxBytes != 2048 {
		t.Fatalf("expected max bytes 2048, got %d", cfg.MCP.ResponseMaxBytes)
	}
}

func TestApiConfig_MCPOverrideByEnv(t *testing.T) {
	env := map[string]string{
		"APP_API_SERVER_MCP_ENABLED":            "true",
		"APP_API_SERVER_MCP_PATH":               "/custom-mcp",
		"APP_API_SERVER_MCP_ALLOWED_ORIGINS":    "https://one.example.com,https://two.example.com",
		"APP_API_SERVER_MCP_INCLUDE_OPERATIONS": "get_service_version,find_routes",
		"APP_API_SERVER_MCP_REQUEST_TIMEOUT_MS": "4000",
		"APP_API_SERVER_MCP_RESPONSE_MAX_BYTES": "4096",
	}
	for key, value := range env {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("failed to set env %s: %v", key, err)
		}
		defer os.Unsetenv(key)
	}

	v := newTestViper(t, `
api:
  server:
    mcp:
      path: /mcp
      request_timeout_ms: 1000
      response_max_bytes: 1024
`)

	cfg := apiConfig(v)

	if !cfg.MCP.Enabled {
		t.Fatal("expected env to enable MCP")
	}
	if cfg.MCP.Path != "/custom-mcp" {
		t.Fatalf("expected custom path, got %q", cfg.MCP.Path)
	}
	if !reflect.DeepEqual(cfg.MCP.AllowedOrigins, []string{"https://one.example.com", "https://two.example.com"}) {
		t.Fatalf("unexpected origins: %v", cfg.MCP.AllowedOrigins)
	}
	if !reflect.DeepEqual(cfg.MCP.IncludeOperations, []string{"get_service_version", "find_routes"}) {
		t.Fatalf("unexpected include operations: %v", cfg.MCP.IncludeOperations)
	}
	if cfg.MCP.RequestTimeoutMs != 4000 {
		t.Fatalf("expected timeout 4000, got %d", cfg.MCP.RequestTimeoutMs)
	}
	if cfg.MCP.ResponseMaxBytes != 4096 {
		t.Fatalf("expected max bytes 4096, got %d", cfg.MCP.ResponseMaxBytes)
	}
}

func TestApiConfig_MCPEnabledCanBeDisabledByEnv(t *testing.T) {
	const envKey = "APP_API_SERVER_MCP_ENABLED"
	if err := os.Setenv(envKey, "false"); err != nil {
		t.Fatalf("failed to set env %s: %v", envKey, err)
	}
	defer os.Unsetenv(envKey)

	v := newTestViper(t, `
api:
  server:
    mcp:
      enabled: true
      path: /mcp
`)

	cfg := apiConfig(v)

	if cfg.MCP.Enabled {
		t.Fatal("expected env to disable MCP")
	}
	if cfg.MCP.Path != "/mcp" {
		t.Fatalf("expected config path to remain /mcp, got %q", cfg.MCP.Path)
	}
}

func newTestViper(t *testing.T, config string) *viper.Viper {
	t.Helper()

	v := viper.New()
	v.SetConfigType("yaml")
	if config != "" {
		if err := v.ReadConfig(strings.NewReader(config)); err != nil {
			t.Fatalf("failed to read config: %v", err)
		}
	}
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return v
}

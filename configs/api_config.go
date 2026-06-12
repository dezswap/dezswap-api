package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func apiConfig(v *viper.Viper) ApiConfig {

	apiServerC := apiServerConfig(v.Sub("api.server"))
	envApiServerC := apiServerConfigFromEnv(v, "API_SERVER")

	apiServerC.Override(envApiServerC)

	dbC := rdbConfig(v.Sub("api.db"))
	envDbC := rdbConfigFromEnv(v, "API_DB")
	dbC.Override(envDbC)

	cacheC := cacheConfig(v.Sub("api.cache"))
	envCacheC := cacheConfigFromEnv(v, "API_CACHE")
	cacheC.Override(envCacheC)

	mcpC := apiMCPConfig(v.Sub("api.server.mcp"))
	envMCPC := apiMCPConfigFromEnv(v, "API_SERVER_MCP")
	mcpC = mcpC.Override(envMCPC)
	if v.IsSet(strings.ToUpper(fmt.Sprintf("%s_%s", "API_SERVER_MCP", "enabled"))) {
		mcpC.Enabled = envMCPC.Enabled
	}

	return ApiConfig{
		Server: apiServerC,
		MCP:    mcpC,
		DB:     dbC,
		Cache:  cacheC,
	}
}

func apiServerConfig(v *viper.Viper) ApiServerConfig {
	if v == nil {
		return ApiServerConfig{}
	}

	return ApiServerConfig{
		Name:               v.GetString("name"),
		Host:               v.GetString("host"),
		Port:               v.GetString("port"),
		Swagger:            v.GetBool("swagger"),
		Mode:               v.GetString("mode"),
		ChainId:            v.GetString("chain_id"),
		CorsAllowedOrigins: v.GetStringSlice("cors_allowed_origins"),
		CoinGeckoApiKey:    v.GetString("coingecko_api_key"),
	}
}

func apiServerConfigFromEnv(v *viper.Viper, prefix string) ApiServerConfig {
	if v == nil {
		return ApiServerConfig{}
	}
	return ApiServerConfig{
		Name:               v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "name"))),
		Host:               v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "host"))),
		Port:               v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "port"))),
		Swagger:            v.GetBool(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "swagger"))),
		Mode:               v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "mode"))),
		ChainId:            v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "chain_id"))),
		CorsAllowedOrigins: splitAndTrim(v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "cors_allowed_origins")))),
		CoinGeckoApiKey:    v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "coingecko_api_key"))),
	}
}

type ApiConfig struct {
	Server ApiServerConfig
	MCP    ApiMCPConfig
	DB     RdbConfig
	Cache  CacheConfig
}

// ApiServerConfig is config struct for app
type ApiServerConfig struct {
	Name               string
	Host               string
	Port               string
	Swagger            bool
	Mode               string
	ChainId            string
	CorsAllowedOrigins []string
	CoinGeckoApiKey    string
}

type ApiMCPConfig struct {
	Enabled           bool     // Optional; mounts the MCP endpoint.
	Path              string   // Optional; streamable HTTP path. Defaults to "/mcp".
	BaseURL           string   // Optional; public API URL.
	AllowedOrigins    []string // Optional; exact browser origins.
	IncludeOperations []string // Optional; limits exposed tools. Empty means all.
	RequestTimeoutMs  int      // Optional; limits each tool call.
	ResponseMaxBytes  int      // Optional; caps tool responses.
}

func (lhs *ApiServerConfig) Override(rhs ApiServerConfig) {
	if rhs.Name != "" {
		lhs.Name = rhs.Name
	}
	if rhs.Host != "" {
		lhs.Host = rhs.Host
	}
	if rhs.Port != "" {
		lhs.Port = rhs.Port
	}
	if rhs.Swagger {
		lhs.Swagger = rhs.Swagger
	}
	if rhs.Mode != "" {
		lhs.Mode = rhs.Mode
	}
	if rhs.ChainId != "" {
		lhs.ChainId = rhs.ChainId
	}
	if len(rhs.CorsAllowedOrigins) > 0 {
		lhs.CorsAllowedOrigins = rhs.CorsAllowedOrigins
	}
	if rhs.CoinGeckoApiKey != "" {
		lhs.CoinGeckoApiKey = rhs.CoinGeckoApiKey
	}
}

func apiMCPConfig(v *viper.Viper) ApiMCPConfig {
	if v == nil {
		return ApiMCPConfig{}
	}
	return ApiMCPConfig{
		Enabled:           v.GetBool("enabled"),
		Path:              v.GetString("path"),
		BaseURL:           v.GetString("base_url"),
		AllowedOrigins:    v.GetStringSlice("allowed_origins"),
		IncludeOperations: v.GetStringSlice("include_operations"),
		RequestTimeoutMs:  v.GetInt("request_timeout_ms"),
		ResponseMaxBytes:  v.GetInt("response_max_bytes"),
	}
}

func apiMCPConfigFromEnv(v *viper.Viper, prefix string) ApiMCPConfig {
	if v == nil {
		return ApiMCPConfig{}
	}
	return ApiMCPConfig{
		Enabled:           v.GetBool(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "enabled"))),
		Path:              v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "path"))),
		BaseURL:           v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "base_url"))),
		AllowedOrigins:    splitAndTrim(v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "allowed_origins")))),
		IncludeOperations: splitAndTrim(v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "include_operations")))),
		RequestTimeoutMs:  v.GetInt(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "request_timeout_ms"))),
		ResponseMaxBytes:  v.GetInt(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "response_max_bytes"))),
	}
}

func (lhs ApiMCPConfig) Override(rhs ApiMCPConfig) ApiMCPConfig {
	if rhs.Enabled {
		lhs.Enabled = rhs.Enabled
	}
	if rhs.Path != "" {
		lhs.Path = rhs.Path
	}
	if rhs.BaseURL != "" {
		lhs.BaseURL = rhs.BaseURL
	}
	if len(rhs.AllowedOrigins) > 0 {
		lhs.AllowedOrigins = rhs.AllowedOrigins
	}
	if len(rhs.IncludeOperations) > 0 {
		lhs.IncludeOperations = rhs.IncludeOperations
	}
	if rhs.RequestTimeoutMs > 0 {
		lhs.RequestTimeoutMs = rhs.RequestTimeoutMs
	}
	if rhs.ResponseMaxBytes > 0 {
		lhs.ResponseMaxBytes = rhs.ResponseMaxBytes
	}
	return lhs
}

func splitAndTrim(value string) []string {
	if value == "" {
		return nil
	}
	items := strings.Split(value, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

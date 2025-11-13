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

	return ApiConfig{
		Server: apiServerC,
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
	}
}

type ApiConfig struct {
	Server ApiServerConfig
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

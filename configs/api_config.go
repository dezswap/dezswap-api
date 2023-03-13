package configs

import (
	"github.com/spf13/viper"
)

type ApiConfig struct {
	Server ApiServerConfig
	DB     RdbConfig
}

// ApiServerConfig is config struct for app
type ApiServerConfig struct {
	Name    string
	Host    string
	Port    string
	Swagger bool
	Mode    string
	Version string
	ChainId string
}

func apiConfig(v *viper.Viper) ApiConfig {
	if v.Sub("api") == nil {
		return ApiConfig{}
	}
	return ApiConfig{
		Server: apiServerConfig(v.Sub("api.server")),
		DB:     rdbConfig(v.Sub("api.db")),
	}
}

func apiServerConfig(v *viper.Viper) ApiServerConfig {
	return ApiServerConfig{
		Name:    v.GetString("name"),
		Host:    v.GetString("host"),
		Port:    v.GetString("port"),
		Swagger: v.GetBool("swagger"),
		Mode:    v.GetString("mode"),
		Version: v.GetString("version"),
		ChainId: v.GetString("chain_id"),
	}
}

func (c *ApiServerConfig) Address() string {
	return c.Host + c.Port
}

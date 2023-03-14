package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type GrpcConfig struct {
	Host string
	Port string
}

func (lhs *GrpcConfig) Override(rhs GrpcConfig) {
	if rhs.Host != "" {
		lhs.Host = rhs.Host
	}
	if rhs.Port != "" {
		lhs.Port = rhs.Port
	}
}

func grpcConfig(v *viper.Viper) GrpcConfig {
	if v == nil {
		return GrpcConfig{}
	}
	return GrpcConfig{
		Host: v.GetString("host"),
		Port: v.GetString("port"),
	}
}

func grpcConfigFromEnv(v *viper.Viper, prefix string) GrpcConfig {
	if v == nil {
		return GrpcConfig{}
	}
	return GrpcConfig{
		Host: v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "host"))),
		Port: v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "port"))),
	}
}

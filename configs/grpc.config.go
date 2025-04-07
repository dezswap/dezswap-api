package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type GrpcConfig struct {
	Host   string
	Port   string
	UseTls bool
}

func (lhs *GrpcConfig) Override(rhs GrpcConfig) {
	if rhs.Host != "" {
		lhs.Host = rhs.Host
	}
	if rhs.Port != "" {
		lhs.Port = rhs.Port
	}
	if rhs.UseTls {
		lhs.UseTls = rhs.UseTls
	}
}

func grpcConfig(v *viper.Viper) GrpcConfig {
	if v == nil {
		return GrpcConfig{}
	}
	return GrpcConfig{
		Host:   v.GetString("host"),
		Port:   v.GetString("port"),
		UseTls: v.GetBool("use_tls"),
	}
}

func grpcConfigFromEnv(v *viper.Viper, prefix string) GrpcConfig {
	if v == nil {
		return GrpcConfig{}
	}
	return GrpcConfig{
		Host:   v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "host"))),
		Port:   v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "port"))),
		UseTls: v.GetBool(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "use_tls"))),
	}
}

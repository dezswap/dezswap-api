package configs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type GrpcConfig struct {
	Host            string `mapstructure:"host" json:"host"`
	Port            string `mapstructure:"port" json:"port"`
	UseTls          bool   `mapstructure:"use_tls" json:"use_tls"`
	QueryTimeoutSec int    `mapstructure:"query_timeout_sec" json:"query_timeout_sec"`
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
	if rhs.QueryTimeoutSec > 0 {
		lhs.QueryTimeoutSec = rhs.QueryTimeoutSec
	}
}

// IsZero reports whether the config names a node. QueryTimeoutSec is deliberately
// excluded: it tunes how a node is queried rather than which node is used, and callers
// treat a non-zero config as "the node was overridden".
func (lhs GrpcConfig) IsZero() bool {
	return lhs.Host == "" && lhs.Port == "" && !lhs.UseTls
}

func grpcConfig(v *viper.Viper) GrpcConfig {
	if v == nil {
		return GrpcConfig{}
	}
	return GrpcConfig{
		Host:            v.GetString("host"),
		Port:            v.GetString("port"),
		UseTls:          v.GetBool("use_tls"),
		QueryTimeoutSec: v.GetInt("query_timeout_sec"),
	}
}

func grpcConfigFromEnv(v *viper.Viper, prefix string) GrpcConfig {
	if v == nil {
		return GrpcConfig{}
	}
	return GrpcConfig{
		Host:            v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "host"))),
		Port:            v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "port"))),
		UseTls:          v.GetBool(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "use_tls"))),
		QueryTimeoutSec: v.GetInt(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "query_timeout_sec"))),
	}
}

func grpcConfigs(v *viper.Viper, key string) ([]GrpcConfig, error) {
	if v == nil {
		return nil, nil
	}

	var configs []GrpcConfig
	if err := v.UnmarshalKey(key, &configs); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", key, err)
	}
	return configs, nil
}

func grpcConfigsFromEnv(v *viper.Viper, prefix string) ([]GrpcConfig, error) {
	if v == nil {
		return nil, nil
	}

	value := v.GetString(strings.ToUpper(prefix))
	if value == "" {
		return nil, nil
	}

	var configs []GrpcConfig
	if err := json.Unmarshal([]byte(value), &configs); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", strings.ToUpper(prefix), err)
	}
	return configs, nil
}

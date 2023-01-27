package configs

import (
	"github.com/spf13/viper"
)

type GrpcConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

func grpcConfig(v *viper.Viper) GrpcConfig {
	return GrpcConfig{
		Host: v.GetString("host"),
		Port: v.GetInt("port"),
	}
}

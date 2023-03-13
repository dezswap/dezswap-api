package configs

import (
	"github.com/spf13/viper"
)

// db contains configs for other services
type RdbConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
}

func rdbConfig(v *viper.Viper) RdbConfig {
	return RdbConfig{
		Host:     v.GetString("host"),
		Port:     v.GetString("port"),
		Database: v.GetString("database"),
		Username: v.GetString("username"),
		Password: v.GetString("password"),
	}
}

func (c RdbConfig) Endpoint() string {
	return c.Host + ":" + c.Port
}

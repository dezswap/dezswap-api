package configs

import (
	"fmt"
	"strings"

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

func (lhs *RdbConfig) Override(rhs RdbConfig) {
	if rhs.Host != "" {
		lhs.Host = rhs.Host
	}
	if rhs.Port != "" {
		lhs.Port = rhs.Port
	}
	if rhs.Database != "" {
		lhs.Database = rhs.Database
	}
	if rhs.Username != "" {
		lhs.Username = rhs.Username
	}
	if rhs.Password != "" {
		lhs.Password = rhs.Password
	}
}

func rdbConfig(v *viper.Viper) RdbConfig {
	if v == nil {
		return RdbConfig{}
	}
	return RdbConfig{
		Host:     v.GetString("host"),
		Port:     v.GetString("port"),
		Database: v.GetString("database"),
		Username: v.GetString("username"),
		Password: v.GetString("password"),
	}
}

func rdbConfigFromEnv(v *viper.Viper, prefix string) RdbConfig {
	if v == nil {
		return RdbConfig{}
	}
	return RdbConfig{
		Host:     v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "host"))),
		Port:     v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "port"))),
		Database: v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "database"))),
		Username: v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "username"))),
		Password: v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "password"))),
	}
}

func (c RdbConfig) Endpoint() string {
	return c.Host + ":" + c.Port
}

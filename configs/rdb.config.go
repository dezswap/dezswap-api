package configs

import (
	"strconv"

	"github.com/spf13/viper"
)

// db contains configs for other services
type RdbConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

func rdbConfig(v *viper.Viper) RdbConfig {
	return RdbConfig{
		Host:     v.GetString("rdb.host"),
		Port:     v.GetInt("rdb.port"),
		Database: v.GetString("rdb.database"),
		Username: v.GetString("rdb.username"),
		Password: v.GetString("rdb.password"),
	}
}

func (c RdbConfig) Endpoint() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

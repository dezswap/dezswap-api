package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type RedisConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       int
	Protocol int
}

func (lhs *RedisConfig) Override(rhs RedisConfig) {
	if rhs.Host != "" {
		lhs.Host = rhs.Host
	}
	if rhs.Port != "" {
		lhs.Port = rhs.Port
	}
	if rhs.User != "" {
		lhs.User = rhs.User
	}
	if rhs.Password != "" {
		lhs.Password = rhs.Password
	}
	if rhs.DB != 0 {
		lhs.DB = rhs.DB
	}
	if rhs.Protocol != 0 {
		lhs.Protocol = rhs.Protocol
	}
}

func redisConfig(v *viper.Viper) RedisConfig {
	if v == nil {
		return RedisConfig{}
	}

	// redis default
	protocol := 3
	if !v.IsSet("protocol") {
		protocol = v.GetInt("protocol")
	}

	return RedisConfig{
		Host:     v.GetString("host"),
		Port:     v.GetString("port"),
		User:     v.GetString("user"),
		Password: v.GetString("password"),
		DB:       v.GetInt("db"),
		Protocol: protocol,
	}
}

func redisConfigFromEnv(v *viper.Viper, prefix string) RedisConfig {
	if v == nil {
		return RedisConfig{}
	}

	var protocol = 3
	if v.IsSet(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "protocol"))) {
		protocol = v.GetInt(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "protocol")))
	}

	return RedisConfig{
		Host:     v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "host"))),
		Port:     v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "port"))),
		User:     v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "user"))),
		Password: v.GetString(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "password"))),
		DB:       v.GetInt(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "db"))),
		Protocol: protocol,
	}
}

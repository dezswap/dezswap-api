package configs

import "github.com/spf13/viper"

type SentryConfig struct {
	DSN string
}

func sentryConfig(v *viper.Viper) SentryConfig {
	dsn := ""
	if sub := v.Sub("sentry"); sub != nil {
		dsn = sub.GetString("dsn")
	}
	return SentryConfig{
		DSN: dsn,
	}
}

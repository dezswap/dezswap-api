package configs

import "github.com/spf13/viper"

type SentryConfig struct {
	DSN string
}

func sentryConfig(v *viper.Viper) SentryConfig {
	return SentryConfig{
		DSN: v.GetString("sentry.dsn"),
	}
}

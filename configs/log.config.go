package configs

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type LogConfig struct {
	// Log level for the global `Logger`
	Level logrus.Level
	// Should log-messages be printed as JSON?
	FormatJSON bool
	// The environment the service is currently running in, e.g. local/development/staging
	Environment string
	ChainId     string
}

func logConfig(v *viper.Viper) LogConfig {
	level, err := logrus.ParseLevel(v.GetString("log.level"))
	if err != nil {
		panic(errors.Wrap(err, "could not parse log level"))
	}

	return LogConfig{
		Level:       level,
		FormatJSON:  v.GetBool("log.formatJson"),
		Environment: v.GetString("log.env"),
		ChainId:     v.GetString("log.chainId"),
	}
}

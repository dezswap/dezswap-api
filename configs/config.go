package configs

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	fileName  = "config"
	envPrefix = "app"
)

var envConfig Config
var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// Config aggregation
type Config struct {
	Log     LogConfig
	Sentry  SentryConfig
	Indexer IndexerConfig
}

// Init is explicit initializer for Config
func New() Config {
	v := initViper(fileName)
	envConfig = Config{
		Log:     logConfig(v),
		Sentry:  sentryConfig(v),
		Indexer: indexerConfig(v),
	}
	return envConfig
}

func NewWithFileName(fileName string) Config {
	v := initViper(fileName)
	envConfig = Config{
		Log:     logConfig(v),
		Sentry:  sentryConfig(v),
		Indexer: indexerConfig(v),
	}
	return envConfig
}

// Get returns Config object
func Get() Config {
	return envConfig
}

func initViper(configName string) *viper.Viper {
	v := viper.New()
	v.SetConfigName(configName)

	if basepath == "" {
		panic(errors.New("package root path is not initialized"))
	}
	v.AddConfigPath(".") // optionally look for config in the working directory

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// All env vars starts with APP_
	v.AutomaticEnv()
	return v
}

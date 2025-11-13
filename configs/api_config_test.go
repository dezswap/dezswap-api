package configs

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestApiConfig_CorsAllowedOriginsFromConfig(t *testing.T) {
	v := newTestViper(t, `
api:
  server:
    cors_allowed_origins:
      - foo\.example\.com$
      - bar\.example\.com$
`)

	cfg := apiConfig(v)

	expected := []string{`foo\.example\.com$`, `bar\.example\.com$`}
	if !reflect.DeepEqual(cfg.Server.CorsAllowedOrigins, expected) {
		t.Fatalf("expected cors origins %v, got %v", expected, cfg.Server.CorsAllowedOrigins)
	}
}

func TestApiConfig_CorsAllowedOriginsOverrideByEnv(t *testing.T) {
	const envKey = "APP_API_SERVER_CORS_ALLOWED_ORIGINS"
	if err := os.Setenv(envKey, `foo\.example\.com$,baz\.example\.com$`); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	defer os.Unsetenv(envKey)

	v := newTestViper(t, `
api:
  server:
    cors_allowed_origins:
      - default\.example\.com$
`)

	cfg := apiConfig(v)

	expected := []string{`foo\.example\.com$`, `baz\.example\.com$`}
	if !reflect.DeepEqual(cfg.Server.CorsAllowedOrigins, expected) {
		t.Fatalf("expected cors origins %v, got %v", expected, cfg.Server.CorsAllowedOrigins)
	}
}

func newTestViper(t *testing.T, config string) *viper.Viper {
	t.Helper()

	v := viper.New()
	v.SetConfigType("yaml")
	if config != "" {
		if err := v.ReadConfig(strings.NewReader(config)); err != nil {
			t.Fatalf("failed to read config: %v", err)
		}
	}
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return v
}

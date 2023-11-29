package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type CacheConfig struct {
	MemoryCache bool
	RedisConfig *RedisConfig
}

func (lhs *CacheConfig) Override(rhs CacheConfig) {
	if rhs.MemoryCache {
		lhs.MemoryCache = rhs.MemoryCache
	}
	if rhs.RedisConfig != nil {
		if lhs.RedisConfig == nil {
			lhs.RedisConfig = &RedisConfig{}
		}
		lhs.RedisConfig.Override(*rhs.RedisConfig)
	}
}

func cacheConfig(v *viper.Viper) CacheConfig {
	if v == nil {
		return CacheConfig{
			MemoryCache: false,
			RedisConfig: nil,
		}
	}

	var rc *RedisConfig = nil
	sub := v.Sub("redis")
	if sub != nil {
		v := redisConfig(sub)
		rc = &v
	}

	return CacheConfig{
		MemoryCache: v.GetBool("memory_cache"),
		RedisConfig: rc,
	}
}

func cacheConfigFromEnv(v *viper.Viper, prefix string) CacheConfig {
	if v == nil {
		return CacheConfig{
			MemoryCache: false,
			RedisConfig: nil,
		}
	}

	rc := redisConfigFromEnv(v, fmt.Sprintf("%s_%s", prefix, "redis"))
	return CacheConfig{
		MemoryCache: v.GetBool(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "enable_memory_cache"))),
		RedisConfig: &rc,
	}
}

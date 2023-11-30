package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type CacheConfig struct {
	MemoryCache bool
	RedisConfig RedisConfig
}

func (lhs *CacheConfig) Override(rhs CacheConfig) {
	if rhs.MemoryCache {
		lhs.MemoryCache = rhs.MemoryCache
	}
	lhs.RedisConfig.Override(rhs.RedisConfig)
}

func cacheConfig(v *viper.Viper) CacheConfig {
	if v == nil {
		return CacheConfig{
			MemoryCache: false,
			RedisConfig: RedisConfig{},
		}
	}

	return CacheConfig{
		MemoryCache: v.GetBool("memory_cache"),
		RedisConfig: redisConfig(v.Sub("redis")),
	}
}

func cacheConfigFromEnv(v *viper.Viper, prefix string) CacheConfig {
	if v == nil {
		return CacheConfig{
			MemoryCache: false,
			RedisConfig: RedisConfig{},
		}
	}

	return CacheConfig{
		MemoryCache: v.GetBool(strings.ToUpper(fmt.Sprintf("%s_%s", prefix, "enable_memory_cache"))),
		RedisConfig: redisConfigFromEnv(v, fmt.Sprintf("%s_%s", prefix, "redis")),
	}
}

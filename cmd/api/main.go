package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	cache_redis "github.com/dezswap/dezswap-api/pkg/cache/redis"
	"github.com/redis/go-redis/v9"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/dezswap/dezswap-api/api"
	"github.com/dezswap/dezswap-api/configs"
	"github.com/dezswap/dezswap-api/pkg/cache"
	"github.com/dezswap/dezswap-api/pkg/cache/memory"
)

func main() {
	c := configs.New()
	c.Log.ChainId = c.Api.Server.ChainId
	cache := cacheStore(c.Api.Cache)
	db := dbCon(c.Api.DB)
	api.RunServer(c, cache, db)
}

func dbCon(c configs.RdbConfig) *gorm.DB {

	dbDsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.Username, c.Password, c.Database,
	)
	writer := io.MultiWriter(os.Stdout)
	db, err := gorm.Open(postgres.Open(dbDsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		Logger: logger.New(
			log.New(writer, "\r\n", log.LstdFlags),
			logger.Config{
				IgnoreRecordNotFoundError: true,
				SlowThreshold:             time.Second,
				Colorful:                  false,
				LogLevel:                  logger.Warn,
			},
		),
	})
	if err != nil {
		panic(err)
	}
	return db
}

func cacheStore(c configs.CacheConfig) cache.Cache {
	if c.RedisConfig.Host != "" {
		option := redis.Options{
			Addr:     fmt.Sprintf("%s:%s", c.RedisConfig.Host, c.RedisConfig.Port),
			Username: c.RedisConfig.User,
			Password: c.RedisConfig.Password,
			DB:       c.RedisConfig.DB,
			Protocol: c.RedisConfig.Protocol,
		}

		client := redis.NewClient(&option)
		if err := client.Ping(context.Background()).Err(); err != nil {
			panic(err)
		}
		return cache_redis.New(cache.NewByteCodec(), client)
	}
	if c.MemoryCache {
		return memory.NewMemoryCache(cache.NewByteCodec())
	}

	return nil
}

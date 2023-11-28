package redis

import (
	"context"
	"time"

	"github.com/dezswap/dezswap-api/pkg/cache"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type redisClient struct {
	cache.Codable
	*redis.Client
}

var _ cache.Cache = &redisClient{}

//	option := redis.Options{
//		Addr:     fmt.Sprintf("%s:%s", c.Host, c.Port),
//		Username: c.User,
//		Password: c.Password,
//		DB:       c.DB,
//		Protocol: c.Protocol,
//	}
//
// client := redis.NewClient(&option)
// New returns a new Redis cache instance.
func New(codec cache.Codable, redis *redis.Client) cache.Cache {
	return &redisClient{codec, redis}
}

// Get implements cache.Cache.
func (r *redisClient) Get(Key string, dest interface{}) error {

	ctx := context.TODO()
	bytes, err := r.Client.Get(ctx, Key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return cache.ErrCacheMiss
		}
		return errors.Wrap(err, "redisClient.Get")
	}

	if err := r.Decode(bytes, dest); err != nil {
		return errors.Wrap(err, "redisClient.Get")
	}

	return nil
}

// Set implements cache.Cache.
func (r *redisClient) Set(Key string, value interface{}, ttl time.Duration) error {

	bytes, err := r.Encode(value)
	if err != nil {
		return errors.Wrap(err, "redisClient.Set")
	}

	ctx := context.TODO()
	if err := r.Client.Set(ctx, Key, bytes, ttl).Err(); err != nil {
		return errors.Wrap(err, "redisClient.Set")
	}

	return nil
}

// Set implements cache.Cache.
func (r *redisClient) Delete(Key string) error {

	ctx := context.TODO()
	if err := r.Client.Del(ctx, Key).Err(); err != nil {
		return errors.Wrap(err, "redisClient.Set")
	}

	return nil
}

package memory

import (
	"context"
	"sync"
	"time"

	"github.com/dezswap/dezswap-api/pkg/cache"
	"github.com/pkg/errors"
)

const defaultCleanupInterval = time.Minute

type item struct {
	Value    interface{}
	ExpireAt *time.Time
}

type memoryCacheImpl struct {
	codec cache.Codable
	*sync.RWMutex
	store map[string]item
}

func NewMemoryCache(ctx context.Context, codec cache.Codable) cache.Cache {
	c := &memoryCacheImpl{
		codec:   codec,
		RWMutex: &sync.RWMutex{},
		store:   make(map[string]item),
	}
	go c.startCleanup(ctx, defaultCleanupInterval)
	return c
}

func (r *memoryCacheImpl) Ping() error {
	return nil
}

func (c *memoryCacheImpl) Get(key string, dest interface{}) error {
	c.RLock()
	item, found := c.store[key]
	c.RUnlock()

	if !found {
		return cache.ErrCacheMiss
	}
	if item.ExpireAt != nil && time.Now().After(*item.ExpireAt) {
		c.evictIfExpired(key)
		return cache.ErrCacheMiss
	}
	if err := c.codec.Decode(item.Value.([]byte), dest); err != nil {
		return errors.Wrap(err, "memoryCacheImpl.Get")
	}
	return nil
}

func (c *memoryCacheImpl) evictIfExpired(key string) {
	c.Lock()
	defer c.Unlock()
	if v, ok := c.store[key]; ok && v.ExpireAt != nil && time.Now().After(*v.ExpireAt) {
		delete(c.store, key)
	}
}

func (c *memoryCacheImpl) Set(key string, value interface{}, ttl time.Duration) error {
	c.Lock()
	defer c.Unlock()

	item := item{
		Value: value,
	}
	if ttl > cache.CacheLifeTimeNeverExpired {
		t := time.Now().Add(ttl)
		item.ExpireAt = &t
	}
	itemBytes, err := c.codec.Encode(value)
	if err != nil {
		return errors.Wrap(err, "memoryCacheImpl.Set")
	}
	item.Value = itemBytes
	c.store[key] = item
	return nil
}

func (c *memoryCacheImpl) Delete(key string) error {
	c.Lock()
	defer c.Unlock()

	delete(c.store, key)
	return nil
}

func (c *memoryCacheImpl) startCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-ctx.Done():
			return
		}
	}
}

func (c *memoryCacheImpl) deleteExpired() {
	now := time.Now()
	c.RLock()
	var expired []string
	for key, item := range c.store {
		if item.ExpireAt != nil && now.After(*item.ExpireAt) {
			expired = append(expired, key)
		}
	}
	c.RUnlock()

	if len(expired) == 0 {
		return
	}

	c.Lock()
	for _, key := range expired {
		if v, ok := c.store[key]; ok && v.ExpireAt != nil && time.Now().After(*v.ExpireAt) {
			delete(c.store, key)
		}
	}
	c.Unlock()
}

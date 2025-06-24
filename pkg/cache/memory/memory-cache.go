package memory

import (
	"sync"
	"time"

	"github.com/dezswap/dezswap-api/pkg/cache"
	"github.com/pkg/errors"
)

type item struct {
	Value    interface{}
	ExpireAt *time.Time
}

type memoryCacheImpl struct {
	codec cache.Codable
	*sync.RWMutex
	store map[string]item
}

func NewMemoryCache(codec cache.Codable) cache.Cache {
	return &memoryCacheImpl{
		codec:   codec,
		RWMutex: &sync.RWMutex{},
		store:   make(map[string]item),
	}
}

func (r *memoryCacheImpl) Ping() error {
	return nil
}

func (c *memoryCacheImpl) Get(key string, dest interface{}) error {
	c.RLock()
	defer c.RUnlock()

	item, found := c.store[key]
	if !found {
		return cache.ErrCacheMiss
	}

	if item.ExpireAt != nil && time.Now().After(*item.ExpireAt) {
		delete(c.store, key)
		return cache.ErrCacheMiss
	}

	if err := c.codec.Decode(item.Value.([]byte), dest); err != nil {
		return errors.Wrap(err, "memoryCacheImpl.Get")
	}

	return nil
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

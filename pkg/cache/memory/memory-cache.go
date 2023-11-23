package memory

import (
	"sync"
	"time"

	"github.com/dezswap/dezswap-api/pkg/cache"
)

type item struct {
	Value    interface{}
	ExpireAt *time.Time
}

type cacheImpl struct {
	*sync.RWMutex
	store map[string]item
}

func NewMemoryCache() cache.Cache {
	return &cacheImpl{
		RWMutex: &sync.RWMutex{},
		store:   make(map[string]item),
	}
}

func (c *cacheImpl) Get(key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()

	item, found := c.store[key]
	if !found {
		return nil, false
	}

	if item.ExpireAt != nil && time.Now().After(*item.ExpireAt) {
		delete(c.store, key)
		return nil, false
	}

	return item.Value, true
}

func (c *cacheImpl) Set(key string, value interface{}, ttl time.Duration) error {
	c.Lock()
	defer c.Unlock()

	item := item{
		Value: value,
	}
	if ttl > cache.CacheLifeTimeNeverExpired {
		t := time.Now().Add(ttl)
		item.ExpireAt = &t
	}
	c.store[key] = item
	return nil
}

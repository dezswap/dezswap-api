package memory

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/dezswap/dezswap-api/pkg/cache"
	"github.com/pkg/errors"
)

const defaultCleanupInterval = time.Minute

// defaultMaxEntries bounds the store. Keys are built from the request, so a caller
// varying one writes an entry per request; expiry alone would let the process grow
// until it is killed, well inside a single TTL.
const defaultMaxEntries = 10000

type item struct {
	key      string
	value    []byte
	expireAt *time.Time
}

func (i item) expired(now time.Time) bool {
	return i.expireAt != nil && now.After(*i.expireAt)
}

type memoryCacheImpl struct {
	codec cache.Codable
	*sync.RWMutex
	store map[string]*list.Element
	// order holds the entries oldest write first, so eviction is the front of it.
	order      *list.List
	maxEntries int
}

func NewMemoryCache(ctx context.Context, codec cache.Codable) cache.Cache {
	return NewMemoryCacheWithLimit(ctx, codec, defaultMaxEntries)
}

// NewMemoryCacheWithLimit bounds the store at maxEntries, evicting the oldest write
// to stay under it. A non-positive maxEntries falls back to defaultMaxEntries.
func NewMemoryCacheWithLimit(ctx context.Context, codec cache.Codable, maxEntries int) cache.Cache {
	if maxEntries <= 0 {
		maxEntries = defaultMaxEntries
	}

	c := &memoryCacheImpl{
		codec:      codec,
		RWMutex:    &sync.RWMutex{},
		store:      make(map[string]*list.Element),
		order:      list.New(),
		maxEntries: maxEntries,
	}
	go c.startCleanup(ctx, defaultCleanupInterval)
	return c
}

func (r *memoryCacheImpl) Ping(context.Context) error {
	return nil
}

func (c *memoryCacheImpl) Get(key string, dest interface{}) error {
	c.RLock()
	element, found := c.store[key]
	var stored item
	if found {
		stored = element.Value.(item)
	}
	c.RUnlock()

	if !found {
		return cache.ErrCacheMiss
	}
	if stored.expired(time.Now()) {
		c.evictIfExpired(key)
		return cache.ErrCacheMiss
	}
	if err := c.codec.Decode(stored.value, dest); err != nil {
		return errors.Wrap(err, "memoryCacheImpl.Get")
	}
	return nil
}

func (c *memoryCacheImpl) evictIfExpired(key string) {
	c.Lock()
	defer c.Unlock()

	if element, ok := c.store[key]; ok && element.Value.(item).expired(time.Now()) {
		c.remove(element)
	}
}

func (c *memoryCacheImpl) Set(key string, value interface{}, ttl time.Duration) error {
	encoded, err := c.codec.Encode(value)
	if err != nil {
		return errors.Wrap(err, "memoryCacheImpl.Set")
	}

	stored := item{key: key, value: encoded}
	if ttl > cache.CacheLifeTimeNeverExpired {
		t := time.Now().Add(ttl)
		stored.expireAt = &t
	}

	c.Lock()
	defer c.Unlock()

	if element, found := c.store[key]; found {
		element.Value = stored
		c.order.MoveToBack(element)
		return nil
	}

	// The oldest write is also the entry closest to expiring, since the routes that
	// write here share one TTL.
	for c.order.Len() >= c.maxEntries {
		oldest := c.order.Front()
		if oldest == nil {
			break
		}
		c.remove(oldest)
	}

	c.store[key] = c.order.PushBack(stored)
	return nil
}

func (c *memoryCacheImpl) Delete(key string) error {
	c.Lock()
	defer c.Unlock()

	if element, found := c.store[key]; found {
		c.remove(element)
	}
	return nil
}

// remove drops an entry from both the map and the eviction order. Callers hold the
// write lock.
func (c *memoryCacheImpl) remove(element *list.Element) {
	delete(c.store, element.Value.(item).key)
	c.order.Remove(element)
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
	for key, element := range c.store {
		if element.Value.(item).expired(now) {
			expired = append(expired, key)
		}
	}
	c.RUnlock()

	if len(expired) == 0 {
		return
	}

	c.Lock()
	defer c.Unlock()
	for _, key := range expired {
		if element, ok := c.store[key]; ok && element.Value.(item).expired(time.Now()) {
			c.remove(element)
		}
	}
}

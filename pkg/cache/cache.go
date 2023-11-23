package cache

import "time"

// Cache is an simple cache interface
// the cache.Get return nil if key not found instead of error
// the cache.Set override value if key already exists
type Cache interface {
	Get(Key string) (interface{}, bool)
	Set(Key string, value interface{}, ttl time.Duration) error
}

type CacheLifeTime = time.Duration

const (
	CacheLifeTimeNeverExpired CacheLifeTime = 0
)

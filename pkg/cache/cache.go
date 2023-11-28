package cache

import (
	"errors"
	"time"
)

var ErrCacheMiss = errors.New("cache: key not found")

type Codable interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
}

// / Cache is an interface for cache
// / destination must be a pointer
type Cache interface {
	Get(Key string, dest interface{}) error
	Set(Key string, value interface{}, ttl time.Duration) error
}

type CacheLifeTime = time.Duration

const (
	CacheLifeTimeNeverExpired CacheLifeTime = 0
)

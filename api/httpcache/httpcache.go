// Package httpcache builds the middlewares that serve a route from a stored
// response.
package httpcache

import (
	"time"

	gin_cache "github.com/chenyahui/gin-cache"
	"github.com/dezswap/dezswap-api/api/cachekey"
	"github.com/dezswap/dezswap-api/pkg/cache"
	"github.com/gin-gonic/gin"
)

// versionedTTL is the backstop for a write the watermark cannot see; a moving
// version is what normally ends an entry's life.
const versionedTTL = 5 * time.Minute

// Handlers are the two ways a route group may have its responses stored. The
// fields are unexported so that a caller cannot assemble a half-filled one: a
// route group attaches what it is given, and gin calls a nil entry in the chain.
type Handlers struct {
	timed     gin.HandlerFunc
	versioned func(cachekey.Resource) gin.HandlerFunc
}

// Timed expires on block time and keys on the whole request URI.
func (h Handlers) Timed() gin.HandlerFunc {
	if h.timed == nil {
		return passThrough
	}
	return h.timed
}

// Versioned keys on the version of the tables the resource reads.
func (h Handlers) Versioned(r cachekey.Resource) gin.HandlerFunc {
	if h.versioned == nil {
		return passThrough
	}
	return h.versioned(r)
}

func passThrough(c *gin.Context) { c.Next() }

// New returns the handlers backed by store. A missing store, or a missing
// versioner for the versioned half, leaves that half storing nothing.
func New(store cache.Cache, versioner *cachekey.Versioner, blockTime time.Duration) Handlers {
	if store == nil {
		return Handlers{}
	}

	handlers := Handlers{timed: timed(store, blockTime)}
	if versioner != nil {
		handlers.versioned = func(r cachekey.Resource) gin.HandlerFunc {
			return versioned(store, versioner, r, blockTime)
		}
	}

	return handlers
}

// NewFrom assembles Handlers from middleware that is already built.
func NewFrom(timed gin.HandlerFunc, versioned func(cachekey.Resource) gin.HandlerFunc) Handlers {
	return Handlers{timed: timed, versioned: versioned}
}

func timed(store cache.Cache, blockTime time.Duration) gin.HandlerFunc {
	return gin_cache.Cache(store, blockTime,
		gin_cache.WithCacheStrategyByRequest(func(c *gin.Context) (bool, gin_cache.Strategy) {
			return true, gin_cache.Strategy{CacheKey: c.Request.Host + c.Request.RequestURI}
		}),
		gin_cache.WithDiscardHeaders(gin_cache.CorsHeaders()),
	)
}

func versioned(store cache.Cache, versioner *cachekey.Versioner, r cachekey.Resource, blockTime time.Duration) gin.HandlerFunc {
	return gin_cache.Cache(store, versionedTTL,
		gin_cache.WithCacheStrategyByRequest(func(c *gin.Context) (bool, gin_cache.Strategy) {
			// Undeclared parameters are dropped: a caller's cache buster would otherwise
			// split the cache into one entry per request.
			key := c.Request.Host + c.Request.URL.Path
			if query := r.CanonicalQuery(c.Request.URL.Query()); query != "" {
				key += "?" + query
			}

			version, err := versioner.Version(r)
			if err != nil {
				// Nothing to invalidate on, so fall back to a short-lived entry rather than
				// pinning the response to a key that would never move.
				return true, gin_cache.Strategy{CacheKey: key, CacheDuration: blockTime}
			}

			return true, gin_cache.Strategy{
				CacheKey:      key + cachekey.VersionSeparator + version,
				CacheDuration: versionedTTL,
			}
		}),
		gin_cache.WithDiscardHeaders(gin_cache.CorsHeaders()),
	)
}

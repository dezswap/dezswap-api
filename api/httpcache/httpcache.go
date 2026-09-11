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
	timed     func(cachekey.Query) gin.HandlerFunc
	versioned func(cachekey.Resource) gin.HandlerFunc
}

// Timed expires on block time and keys on the parameters q declares.
func (h Handlers) Timed(q cachekey.Query) gin.HandlerFunc {
	if h.timed == nil {
		return passThrough
	}
	return h.timed(q)
}

// Versioned keys on the version of the tables the resource reads.
func (h Handlers) Versioned(r cachekey.Resource) gin.HandlerFunc {
	if h.versioned == nil {
		return passThrough
	}
	return h.versioned(r)
}

func passThrough(c *gin.Context) { c.Next() }

// New returns the handlers backed by store, keying everything they hold under
// chainId. A missing store, or a missing versioner for the versioned half, leaves
// that half storing nothing.
func New(chainId string, store cache.Cache, versioner *cachekey.Versioner, blockTime time.Duration) Handlers {
	if store == nil {
		return Handlers{}
	}

	handlers := Handlers{
		timed: func(q cachekey.Query) gin.HandlerFunc { return timed(chainId, store, q, blockTime) },
	}
	if versioner != nil {
		handlers.versioned = func(r cachekey.Resource) gin.HandlerFunc {
			// r has to be a resource the versioner reads a mark for, or no version can be
			// composed for it. Called once per route at startup, not per request, so this
			// fails the boot.
			if err := versioner.Check(r); err != nil {
				panic(err)
			}

			return versioned(chainId, store, versioner, r, blockTime)
		}
	}

	return handlers
}

// NewFrom assembles Handlers from middleware that is already built.
func NewFrom(timed func(cachekey.Query) gin.HandlerFunc, versioned func(cachekey.Resource) gin.HandlerFunc) Handlers {
	return Handlers{timed: timed, versioned: versioned}
}

// key is what a response is stored under: the chain it was built from, the route
// that answered it, and the declared parameters it varies on. A parameter no route
// declares is dropped, so a caller appending a cache buster cannot mint an entry
// per request.
//
// The chain comes from config, and is what keeps two deployments sharing a store
// from reading each other's entries. The Host header would name the deployment
// too, but it is the caller's to set and nothing here checks it: keyed on that,
// every spelling a caller invents is an entry of its own for the one response.
func key(chainId string, c *gin.Context, canonicalQuery string) string {
	k := chainId + c.Request.URL.Path
	if canonicalQuery != "" {
		k += "?" + canonicalQuery
	}

	return k
}

func timed(chainId string, store cache.Cache, q cachekey.Query, blockTime time.Duration) gin.HandlerFunc {
	return gin_cache.Cache(store, blockTime,
		gin_cache.WithCacheStrategyByRequest(func(c *gin.Context) (bool, gin_cache.Strategy) {
			return true, gin_cache.Strategy{CacheKey: key(chainId, c, q.Canonical(c.Request.URL.Query()))}
		}),
		gin_cache.WithDiscardHeaders(gin_cache.CorsHeaders()),
	)
}

func versioned(chainId string, store cache.Cache, versioner *cachekey.Versioner, r cachekey.Resource, blockTime time.Duration) gin.HandlerFunc {
	return gin_cache.Cache(store, versionedTTL,
		gin_cache.WithCacheStrategyByRequest(func(c *gin.Context) (bool, gin_cache.Strategy) {
			cacheKey := key(chainId, c, r.CanonicalQuery(c.Request.URL.Query()))

			version, err := versioner.Version(r)
			if err != nil {
				// Nothing to invalidate on, so fall back to a short-lived entry rather than
				// pinning the response to a key that would never move.
				return true, gin_cache.Strategy{CacheKey: cacheKey, CacheDuration: blockTime}
			}

			return true, gin_cache.Strategy{
				CacheKey:      cacheKey + cachekey.VersionSeparator + version,
				CacheDuration: versionedTTL,
			}
		}),
		gin_cache.WithDiscardHeaders(gin_cache.CorsHeaders()),
	)
}

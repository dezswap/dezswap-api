package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dezswap/dezswap-api/api/cachekey"
	"github.com/dezswap/dezswap-api/api/httpcache"
	"github.com/dezswap/dezswap-api/configs"
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// seenPaths records which paths reached a given cache, keyed by path.
type seenPaths struct {
	timed     map[string]int
	versioned map[string]string
}

// registerWithCacheSpies wires the routes with stand-ins for the two caches that
// record which paths reach each. Handlers run against no database and are expected
// to fail; only which cache a route carries is under test.
func registerWithCacheSpies(t *testing.T) (*gin.Engine, *seenPaths) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	seen := &seenPaths{timed: map[string]int{}, versioned: map[string]string{}}
	engine := gin.New()
	engine.Use(gin.Recovery())

	cacheHandlers := httpcache.NewFrom(
		func(c *gin.Context) {
			seen.timed[c.Request.URL.Path]++
			c.Next()
		},
		func(r cachekey.Resource) gin.HandlerFunc {
			return func(c *gin.Context) {
				seen.versioned[c.Request.URL.Path] = r.String()
				c.Next()
			}
		},
	)

	RegisterRoutes(
		engine.Group("v1"),
		"test-chain",
		"",
		"test",
		pkg.NetworkMetadata{},
		nil,
		nil,
		cacheHandlers,
		logging.New("test", configs.LogConfig{}),
	)

	return engine, seen
}

func get(engine *gin.Engine, path string) {
	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
}

func TestRegisterRoutes_StatusIsNeverServedFromCache(t *testing.T) {
	engine, seen := registerWithCacheSpies(t)

	// /health reports on the database and the cache as they are at the moment it
	// is asked, and answers 200 either way. A replayed copy would keep reporting a
	// dependency that has since gone away.
	get(engine, "/v1/health")
	get(engine, "/v1/version")

	require.Empty(t, seen.timed)
	require.Empty(t, seen.versioned)
}

func TestRegisterRoutes_RoutesIsNotCached(t *testing.T) {
	engine, seen := registerWithCacheSpies(t)

	// from/to are caller-supplied and an address matching nothing still answers
	// 200, so caching would let any caller mint an entry per string it invents.
	get(engine, "/v1/routes?from=xpla1abc&to=xpla1def")
	get(engine, "/v1/routes?from=whatever-a-caller-types")

	require.Empty(t, seen.timed)
	require.Empty(t, seen.versioned)
}

func TestRegisterRoutes_TokensAndPairsAreVersioned(t *testing.T) {
	engine, seen := registerWithCacheSpies(t)

	// Only these two families have their writes tracked, and each has to be keyed on
	// its own resource rather than a shared one.
	for path, resource := range map[string]string{
		"/v1/tokens":          cachekey.Tokens.String(),
		"/v1/tokens/xpla1abc": cachekey.Tokens.String(),
		"/v1/pairs":           cachekey.Pairs.String(),
		"/v1/pairs/xpla1abc":  cachekey.Pairs.String(),
	} {
		get(engine, path)
		require.Equalf(t, resource, seen.versioned[path], "%s should be versioned on %s", path, resource)
		require.Zerof(t, seen.timed[path], "%s should not also take the timed cache", path)
	}
}

func TestRegisterRoutes_RemainingDataRoutesKeepTheTimedCache(t *testing.T) {
	engine, seen := registerWithCacheSpies(t)

	// Nothing follows when these tables change, so the clock is all that expires
	// them -- exactly what they had before versions existed.
	for _, path := range []string{
		"/v1/pools",
		"/v1/dashboard/pools",
		"/v1/notices",
		"/v1/coingecko/pairs",
	} {
		get(engine, path)
		require.Equalf(t, 1, seen.timed[path], "%s should take the timed cache", path)
		require.NotContainsf(t, seen.versioned, path, "%s has no version to be keyed on", path)
	}
}

// The previous version of this only asked for a versioned route, which carried its
// own nil check; a timed route took a nil handler straight into gin's chain.
func TestRegisterRoutes_EveryRouteServesWithoutACacheStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	// A deployment with no cache store configured still has to route.
	RegisterRoutes(
		engine.Group("v1"),
		"test-chain",
		"",
		"test",
		pkg.NetworkMetadata{},
		nil,
		nil,
		// What cmd/api hands over when neither redis nor the memory cache is set.
		httpcache.New(nil, nil, time.Second),
		logging.New("test", configs.LogConfig{}),
	)

	// The handlers themselves fail on the nil database; reaching them at all is what
	// is under test, since a nil cache handler would abort the chain before them.
	for _, path := range []string{
		"/v1/tokens",          // versioned
		"/v1/pools",           // timed
		"/v1/coingecko/pairs", // timed, nested group
		"/v1/health",          // uncached
	} {
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		require.NotEqualf(t, http.StatusNotFound, rec.Code, "%s was not registered", path)
	}
}

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
	timed map[string]int
	// timedQuery is what the timed route's declaration kept of the query it was
	// asked with.
	timedQuery map[string]string
	versioned  map[string]string
}

// registerWithCacheSpies wires the routes with stand-ins for the two caches that
// record which paths reach each. Handlers run against no database and are expected
// to fail; only which cache a route carries is under test.
func registerWithCacheSpies(t *testing.T) (*gin.Engine, *seenPaths) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	seen := &seenPaths{
		timed:      map[string]int{},
		timedQuery: map[string]string{},
		versioned:  map[string]string{},
	}
	engine := gin.New()
	engine.Use(gin.Recovery())

	cacheHandlers := httpcache.NewFrom(
		func(q cachekey.Query) gin.HandlerFunc {
			return func(c *gin.Context) {
				seen.timed[c.Request.URL.Path]++
				seen.timedQuery[c.Request.URL.Path] = q.Canonical(c.Request.URL.Query())
				c.Next()
			}
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
		"/v1/notices",
		"/v1/coingecko/pairs",
	} {
		get(engine, path)
		require.Equalf(t, 1, seen.timed[path], "%s should take the timed cache", path)
		require.NotContainsf(t, seen.versioned, path, "%s has no version to be keyed on", path)
	}
}

// A timed route that reads a query parameter has to declare it, or the entry stored
// for the first value asked for is replayed for every other value. A parameter no
// route reads must not reach the key at all: keyed on it, a caller appending one
// mints an entry per request.
func TestRegisterRoutes_TimedRoutesDeclareTheParametersTheyRead(t *testing.T) {
	engine, seen := registerWithCacheSpies(t)

	for _, tc := range []struct {
		path  string
		query string
		want  string
	}{
		{"/v1/notices", "?chain=cube&limit=3&asc=TRUE&1756771200000=", "chain=cube&limit=3&asc=true"},
		{"/v1/dashboard/txs", "?pool=xpla1abc&type=swap&1756771200000=", "pool=xpla1abc&type=swap"},
		{"/v1/dashboard/chart/tokens/xpla1abc/volume", "?duration=YEAR&1756771200000=", "duration=year"},
		{"/v1/pools", "?1756771200000=", ""},
		{"/v1/coingecko/pairs", "?1756771200000=", ""},
	} {
		get(engine, tc.path+tc.query)
		require.Equalf(t, tc.want, seen.timedQuery[tc.path], "%s keyed on the wrong query", tc.path)
	}
}

func TestRegisterRoutes_DashboardStatsRoutesAreVersioned(t *testing.T) {
	engine, seen := registerWithCacheSpies(t)

	// One resource per set of parameters read, not one for the whole family: a route
	// keyed on a parameter its handler ignores would split its cache for nothing.
	for path, resource := range map[string]string{
		"/v1/dashboard/chart/tvl":                cachekey.DashboardCharts.String(),
		"/v1/dashboard/chart/pools/xpla1abc/tvl": cachekey.DashboardCharts.String(),
		"/v1/dashboard/recent":                   cachekey.DashboardRecent.String(),
		"/v1/dashboard/pools":                    cachekey.DashboardPools.String(),
	} {
		get(engine, path)
		require.Equalf(t, resource, seen.versioned[path], "%s should be versioned on %s", path, resource)
		// A stacked timed cache would answer first and cap the version at block time.
		require.Zerof(t, seen.timed[path], "%s should not also take the timed cache", path)
	}
}

func TestRegisterRoutes_DashboardRoutesReadingUntrackedTablesStayTimed(t *testing.T) {
	engine, seen := registerWithCacheSpies(t)

	for _, path := range []string{
		"/v1/dashboard/txs",
		"/v1/dashboard/statistics",
		"/v1/dashboard/tokens",
		"/v1/dashboard/tokens/xpla1abc",
		"/v1/dashboard/pools/xpla1abc",
		"/v1/dashboard/chart/tokens/xpla1abc/volume",
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
		httpcache.New("test-chain", nil, nil, time.Second),
		logging.New("test", configs.LogConfig{}),
	)

	// The handlers themselves fail on the nil database; reaching them at all is what
	// is under test, since a nil cache handler would abort the chain before them.
	for _, path := range []string{
		"/v1/tokens",           // versioned
		"/v1/pools",            // timed
		"/v1/coingecko/pairs",  // timed, nested group
		"/v1/dashboard/recent", // versioned, attached per route
		"/v1/dashboard/txs",    // timed, attached per route
		"/v1/health",           // uncached
	} {
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		require.NotEqualf(t, http.StatusNotFound, rec.Code, "%s was not registered", path)
	}
}

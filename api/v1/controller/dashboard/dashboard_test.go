package dashboard

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dezswap/dezswap-api/api/cachekey"
	ds "github.com/dezswap/dezswap-api/api/v1/service/dashboard"
	"github.com/dezswap/dezswap-api/configs"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTokenAddrs(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []ds.Addr
	}{
		{
			name:     "Empty String",
			input:    "",
			expected: []ds.Addr(nil),
		},
		{
			name:     "Empty Tokens",
			input:    ", ",
			expected: []ds.Addr(nil),
		},
		{
			name:     "A Single Token",
			input:    "axpla",
			expected: []ds.Addr{"axpla"},
		},
		{
			name:     "Multiple Tokens",
			input:    "axpla,xpla1abcd,ibc/ABCD1234",
			expected: []ds.Addr{"axpla", "xpla1abcd", "ibc/ABCD1234"},
		},
		{
			name:     "Multiple Tokens with Whitespace",
			input:    "axpla, xpla1abcd ,ibc/ABCD1234",
			expected: []ds.Addr{"axpla", "xpla1abcd", "ibc/ABCD1234"},
		},
		{
			name:     "Multiple Tokens Including Empty One",
			input:    "axpla,,ibc/ABCD1234",
			expected: []ds.Addr{"axpla", "ibc/ABCD1234"},
		},
		{
			name:     "Multiple Tokens Including Whitespace Token",
			input:    "axpla,  ,ibc/ABCD1234",
			expected: []ds.Addr{"axpla", "ibc/ABCD1234"},
		},
		{
			name:     "Starts with Whitespace",
			input:    " axpla,  ,ibc/ABCD1234",
			expected: []ds.Addr{"axpla", "ibc/ABCD1234"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseTokenAddrs(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// stubDashboard answers the handlers under test and leaves the rest of the
// interface nil, so a request that should have been turned away panics instead of
// passing quietly.
type stubDashboard struct {
	ds.Dashboard
}

func (stubDashboard) Volumes(ds.Duration) (ds.Volumes, error)            { return nil, nil }
func (stubDashboard) Pools(...ds.Addr) (ds.Pools, error)                 { return nil, nil }
func (stubDashboard) VolumesOf(ds.Addr, ds.Duration) (ds.Volumes, error) { return nil, nil }

// PoolExists so that a well-formed address reaching the service answers 200 rather
// than the handler's own not-found, which would read the same as a rejection.
func (stubDashboard) PoolDetail(ds.Addr) (ds.PoolDetail, error) {
	return ds.PoolDetail{Recent: ds.Recent{PoolExists: true}}, nil
}

func (stubDashboard) TxsOfToken(ds.TxType, ...ds.Addr) (ds.Txs, error) {
	return nil, nil
}

func newTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	c := &dashboardController{stubDashboard{}, logging.New("test", configs.LogConfig{}), mapper{}}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.GET("/chart/:type", c.Chart)
	engine.GET("/chart/pools/:address/:type", c.ChartByPool)
	engine.GET("/pools", c.Pools)
	engine.GET("/pools/:address", c.Pool)
	engine.GET("/txs", c.Txs)

	return engine
}

func status(engine *gin.Engine, path string) int {
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec.Code
}

func TestHandlers_RejectValuesThatWouldReachTheDatabase(t *testing.T) {
	engine := newTestEngine()

	for name, tc := range map[string]struct {
		path string
		want int
	}{
		// An unknown duration misses chartCriteriaByDuration and puts an empty
		// interval into the SQL, which fails -- and a failure is not cached, so each
		// one is a database round trip.
		"unknown duration":      {"/chart/volume?duration=zzz", http.StatusBadRequest},
		"duration is case free": {"/chart/volume?duration=YEAR", http.StatusOK},
		"absent duration":       {"/chart/volume", http.StatusOK},
		"empty duration":        {"/chart/volume?duration=", http.StatusOK},
		"known duration":        {"/chart/volume?duration=year", http.StatusOK},

		// The check is a shape, not a lookup: an invented name is served like any
		// other. Only characters no address carries are turned away.
		"token nothing names":       {"/pools?token=not-an-address", http.StatusOK},
		"token with a stray symbol": {"/pools?token=nope!", http.StatusBadRequest},
		"no token":                  {"/pools", http.StatusOK},

		// The forms the indexer actually stores: it strips 0x and prepends the
		// configured prefix, so a check written around bare 0x turns real tokens away.
		"native denom":      {"/pools?token=axpla", http.StatusOK},
		"prefixed erc20":    {"/pools?token=xerc20:8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0", http.StatusOK},
		"prefixed cw20":     {"/pools?token=xcw20:xpla1j33xdql0h4kpgj2mhggy4vutw655u90z7nyj4afhxgj4v5urtadq44e3vd", http.StatusOK},
		"bare cw20":         {"/pools?token=xpla1j33xdql0h4kpgj2mhggy4vutw655u90z7nyj4afhxgj4v5urtadq44e3vd", http.StatusOK},
		"ibc denom":         {"/pools?token=ibc/A1B2C3D4E5F60718293A4B5C6D7E8F90A1B2C3D4E5F60718293A4B5C6D7E8F90", http.StatusOK},
		"tx prefixed erc20": {"/txs?token=xerc20:EFGH", http.StatusOK},

		"tx token with a stray symbol":  {"/txs?token=nope!", http.StatusBadRequest},
		"one bad token among good ones": {"/txs?token=axpla,nope!", http.StatusBadRequest},
		"tx pool with a stray symbol":   {"/txs?pool=nope!", http.StatusBadRequest},

		// A path segment lands in the cache key whole, so it is held to what a query
		// parameter is. /chart/pools is versioned, where an entry outlives a timed one.
		"path address":                   {"/chart/pools/xpla1abc/volume", http.StatusOK},
		"path address with stray symbol": {"/chart/pools/nope!/volume", http.StatusBadRequest},
		"pool detail path address":       {"/pools/xpla1abc", http.StatusOK},
		"pool detail stray symbol":       {"/pools/nope!", http.StatusBadRequest},
	} {
		require.Equalf(t, tc.want, status(engine, tc.path), "%s: %s", name, tc.path)
	}
}

// An absent duration puts ds.All in the key, which is what the handler defaults to.
// If the two drift apart, the key names a response that was never built.
func TestDashboardCharts_KeyFallbackMatchesTheHandlerDefault(t *testing.T) {
	require.Equal(t,
		"duration="+string(ds.All),
		cachekey.DashboardCharts.CanonicalQuery(url.Values{}),
	)
}

// ToDuration folds case, so the key has to fold with it. Otherwise each spelling
// gets its own entry, all holding the one response the handler built.
func TestDashboardCharts_KeyFoldsCaseWithTheHandler(t *testing.T) {
	for _, spelling := range []string{"year", "YEAR", "Year", "yEaR"} {
		duration, ok := ds.ToDuration(spelling)
		require.Truef(t, ok, "ToDuration(%q)", spelling)

		require.Equalf(t,
			"duration="+string(duration),
			cachekey.DashboardCharts.CanonicalQuery(url.Values{"duration": {spelling}}),
			"CanonicalQuery(duration=%s)", spelling,
		)
	}
}

package dashboard

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
	} {
		require.Equalf(t, tc.want, status(engine, tc.path), "%s: %s", name, tc.path)
	}
}

package httpcache

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dezswap/dezswap-api/api/cachekey"
	"github.com/dezswap/dezswap-api/pkg/cache"
	"github.com/dezswap/dezswap-api/pkg/cache/memory"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func expectTokenWatermark(mock sqlmock.Sqlmock, rowCount int64, maxMark string) {
	mock.ExpectQuery(`FROM tokens`).
		WithArgs("test-chain").
		WillReturnRows(sqlmock.NewRows([]string{"row_count", "max_mark"}).AddRow(rowCount, maxMark))
}

func TestVersioned_CollapsesCallerAddedQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	require.NoError(t, err)

	// One watermark read backs every request below.
	expectTokenWatermark(mock, 1000, "2026-09-02 00:00:00")

	store := memory.NewMemoryCache(context.Background(), cache.NewByteCodec())
	versioner := cachekey.NewVersioner(context.Background(), gormDB, "test-chain", time.Minute, nil)

	handled := 0
	engine := gin.New()
	engine.GET("/v1/tokens",
		versioned(store, versioner, cachekey.Tokens, time.Second),
		func(c *gin.Context) {
			handled++
			c.JSON(http.StatusOK, gin.H{"tokens": []string{}})
		},
	)

	// The web app appends a millisecond timestamp to every request. Keyed on the
	// whole URI these would be three separate entries and three misses.
	for _, uri := range []string{
		"/v1/tokens?1756771200000=",
		"/v1/tokens?1756771200001=",
		"/v1/tokens",
	} {
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, uri, nil))
		require.Equal(t, http.StatusOK, rec.Code, uri)
	}

	require.Equal(t, 1, handled, "the handler must run once for the three requests")
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestVersioned_AWriteServesAFreshResponse is the claim the whole package rests
// on: a response stops being replayed because the tables behind it were written,
// not because a timer ran out.
func TestVersioned_AWriteServesAFreshResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	require.NoError(t, err)

	// The memo is what holds a version still, so keep it short enough to outlive.
	const memoTTL = 30 * time.Millisecond

	store := memory.NewMemoryCache(context.Background(), cache.NewByteCodec())
	versioner := cachekey.NewVersioner(context.Background(), gormDB, "test-chain", memoTTL, nil)

	handled := 0
	engine := gin.New()
	engine.GET("/v1/tokens",
		versioned(store, versioner, cachekey.Tokens, time.Second),
		func(c *gin.Context) {
			handled++
			c.JSON(http.StatusOK, gin.H{"served": handled})
		},
	)

	get := func() string {
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/tokens", nil))
		require.Equal(t, http.StatusOK, rec.Code)
		return rec.Body.String()
	}

	expectTokenWatermark(mock, 1000, "2026-09-02 00:00:00")
	first := get()

	// Nothing was written, so this has to come back off the stored copy.
	require.Equal(t, first, get(), "the stored response must be replayed while the version holds still")
	require.Equal(t, 1, handled, "the handler must not run again while the version holds still")

	time.Sleep(2 * memoTTL)

	// A token was written: the watermark moves, and with it the key.
	expectTokenWatermark(mock, 1001, "2026-09-02 00:01:00")
	third := get()

	require.NotEqual(t, first, third, "a fresh response must be served after the write")
	require.Equal(t, 2, handled, "the handler must run again after the write")
	require.NoError(t, mock.ExpectationsWereMet())
}

// fallbackRoute is a route whose every request is served by the fallback branch of
// versioned, the one a watermark read failure leaves behind.
type fallbackRoute struct {
	get func() *httptest.ResponseRecorder
	// handled counts handler runs, so a replayed response leaves it alone.
	handled *int
	// reported is what the versioner failed with. An empty one means the read
	// succeeded and the test is no longer looking at the fallback at all.
	reported *[]error
}

// unreadableWatermark wires that route with blockTime as the lifetime the fallback
// is expected to key on. Only one failing read is queued: the failure is memoized
// for a minute, so no later request reaches the database again.
func unreadableWatermark(t *testing.T, blockTime time.Duration) fallbackRoute {
	t.Helper()
	gin.SetMode(gin.TestMode)

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	require.NoError(t, err)

	mock.ExpectQuery(`FROM tokens`).
		WithArgs("test-chain").
		WillReturnError(errors.New("connection refused"))
	t.Cleanup(func() { require.NoError(t, mock.ExpectationsWereMet()) })

	reported := &[]error{}
	store := memory.NewMemoryCache(context.Background(), cache.NewByteCodec())
	versioner := cachekey.NewVersioner(context.Background(), gormDB, "test-chain", time.Minute,
		func(err error) { *reported = append(*reported, err) })

	handled := 0
	engine := gin.New()
	engine.GET("/v1/tokens",
		versioned(store, versioner, cachekey.Tokens, blockTime),
		func(c *gin.Context) {
			handled++
			c.JSON(http.StatusOK, gin.H{"served": handled})
		},
	)

	return fallbackRoute{
		get: func() *httptest.ResponseRecorder {
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/tokens", nil))
			require.Equal(t, http.StatusOK, rec.Code)
			return rec
		},
		handled:  &handled,
		reported: reported,
	}
}

// Losing the watermark costs the version, not the cache: the fallback keys on a
// path alone, and has to go on serving from it.
func TestVersioned_FallbackIsReusedWhileItLives(t *testing.T) {
	// Wide enough that both requests land inside it without racing the clock.
	route := unreadableWatermark(t, 2*time.Second)

	first := route.get().Body.String()
	second := route.get().Body.String()

	require.NotEmpty(t, *route.reported, "the watermark read has to have failed, or this is not the fallback")
	require.Equal(t, first, second, "the fallback entry has to be replayed within blockTime")
	require.Equal(t, 1, *route.handled, "the handler has to run once while the fallback entry lives")
}

// A strategy that leaves CacheDuration unset is served gin_cache's own default,
// which this package sets to versionedTTL. Dropping the field from the fallback
// would hold an unversioned response for five minutes across an outage the version
// cannot be read through -- the one thing the branch exists to prevent -- and every
// other test here would still pass.
func TestVersioned_FallbackExpiresOnBlockTime(t *testing.T) {
	const blockTime = 150 * time.Millisecond
	route := unreadableWatermark(t, blockTime)

	first := route.get().Body.String()
	require.Equal(t, first, route.get().Body.String(), "the fallback entry has to be stored before it can expire")
	require.Equal(t, 1, *route.handled)

	time.Sleep(2 * blockTime)

	require.NotEqual(t, first, route.get().Body.String(), "the fallback entry must not outlive blockTime")
	require.Equal(t, 2, *route.handled, "the handler has to run again once the fallback entry expires")
	require.NotEmpty(t, *route.reported)
}

// New must never hand back a nil handler: a route group attaches what it is given
// without checking, and gin calls a nil entry in the chain.
func TestNew_AlwaysYieldsUsableHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := memory.NewMemoryCache(context.Background(), cache.NewByteCodec())

	for name, handlers := range map[string]Handlers{
		// The zero value is reachable from any caller, so it has to answer too.
		"zero value":             {},
		"no store, no versioner": New(nil, nil, time.Second),
		"store, no versioner":    New(store, nil, time.Second),
		"nothing assembled":      NewFrom(nil, nil),
	} {
		require.NotNil(t, handlers.Timed(), "%s: timed handler", name)
		require.NotNil(t, handlers.Versioned(cachekey.Tokens), "%s: versioned handler", name)

		// Serving through them has to work, not merely be non-nil.
		engine := gin.New()
		engine.GET("/a", handlers.Timed(), func(c *gin.Context) { c.Status(http.StatusOK) })
		engine.GET("/b", handlers.Versioned(cachekey.Tokens), func(c *gin.Context) { c.Status(http.StatusOK) })

		for _, path := range []string{"/a", "/b"} {
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			require.Equal(t, http.StatusOK, rec.Code, "%s: %s", name, path)
		}
	}
}

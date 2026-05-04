package coingecko

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTickerServiceWithMock(t *testing.T) (*tickerService, sqlmock.Sqlmock, func() error) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(
		postgres.New(postgres.Config{Conn: sqlDB}),
		&gorm.Config{},
	)
	require.NoError(t, err)

	svc := &tickerService{
		chainId: "test-chain",
		DB:      gormDB,
	}

	return svc, mock, sqlDB.Close
}

var inactivePoolColumns = []string{
	"base_currency", "target_currency",
	"base_volume", "target_volume", "last_price",
	"base_decimals", "target_decimals",
	"base_liquidity_in_price", "pool_id", "timestamp",
}

// TestQueryInactivePools_ByContractId verifies that inactivePool passes the
// contract address as a bound parameter instead of interpolating it into SQL.
func TestQueryInactivePools_ByContractId(t *testing.T) {
	svc, mock, close := setupTickerServiceWithMock(t)
	defer close()

	mock.ExpectQuery(`from pair_stats_30m`).
		WithArgs("test-chain", "xpla1pool123").
		WillReturnRows(
			sqlmock.NewRows(inactivePoolColumns).
				AddRow("tokenA", "tokenB", "0", "0", "1.5", 6, 6, "100.0", "xpla1pool123", 1000000.0),
		)

	tickers, err := svc.inactivePool("xpla1pool123")
	require.NoError(t, err)
	require.Len(t, tickers, 1)
	require.Equal(t, "xpla1pool123", tickers[0].PoolId)
	require.Equal(t, "1.5", tickers[0].LastPrice)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryInactivePools_WithExclusion verifies that inactivePools passes
// excluded pool IDs as bound parameters instead of interpolating them into SQL.
func TestQueryInactivePools_WithExclusion(t *testing.T) {
	svc, mock, close := setupTickerServiceWithMock(t)
	defer close()

	mock.ExpectQuery(`from pair_stats_30m`).
		WithArgs("test-chain", "xpla1active1", "xpla1active2").
		WillReturnRows(
			sqlmock.NewRows(inactivePoolColumns).
				AddRow("tokenA", "tokenB", "0", "0", "2.0", 6, 6, "200.0", "xpla1inactive", 1000000.0),
		)

	tickers, err := svc.inactivePools([]string{"xpla1active1", "xpla1active2"})
	require.NoError(t, err)
	require.Len(t, tickers, 1)
	require.Equal(t, "xpla1inactive", tickers[0].PoolId)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestQueryInactivePools_NoFilter verifies that inactivePools with no exclusions
// runs the base query with only the chain ID as a bound parameter.
func TestQueryInactivePools_NoFilter(t *testing.T) {
	svc, mock, close := setupTickerServiceWithMock(t)
	defer close()

	mock.ExpectQuery(`from pair_stats_30m`).
		WithArgs("test-chain").
		WillReturnRows(sqlmock.NewRows(inactivePoolColumns))

	tickers, err := svc.inactivePools(nil)
	require.NoError(t, err)
	require.Empty(t, tickers)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestCachePriceInUsdRefetchesAfterTTL verifies that a call after TTL expiry
// does issue a new HTTP request.
func TestCachePriceInUsdRefetchesAfterTTL(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"prices":[[1000000,1.0],[2000000,2.0]]}`)
	}))
	defer srv.Close()

	s := &tickerService{httpClient: srv.Client(), endpoint: srv.URL + "/", apiKey: "test-key"}

	// first call
	assert.NoError(t, s.cachePriceInUsd(priceTokenId))
	assert.Equal(t, int32(1), callCount.Load())

	// simulate TTL expiry
	s.mu.Lock()
	s.cacheExpiry = time.Now().Add(-time.Second)
	s.mu.Unlock()

	// second call after TTL — should re-fetch
	assert.NoError(t, s.cachePriceInUsd(priceTokenId))
	assert.Equal(t, int32(2), callCount.Load(), "call after TTL expiry should re-fetch")
}

// TestNoApiKeyReturnsOnePrice verifies that without an API key no HTTP request
// is made and price() returns 1.0.
func TestNoApiKeyReturnsOnePrice(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		fmt.Fprintln(w, `{"prices":[[1000000,2.5]]}`)
	}))
	defer srv.Close()

	s := NewTickerService("", nil, "").(*tickerService)
	s.httpClient = srv.Client()
	s.endpoint = srv.URL + "/"

	assert.NoError(t, s.cachePriceInUsd(priceTokenId))
	assert.Equal(t, int32(0), callCount.Load(), "no HTTP request should be made without an API key")
	assert.Equal(t, 1.0, s.price(3_000_000, true), "price should be 1.0 when no API key is set")
	assert.Equal(t, 1.0, s.price(3_000_000, false), "price(ts, false) should return 1.0 — trigger block must be skipped")
}

// TestPrice covers all branches of the price() method.
// cachedPrices entries are [timestamp_ms, price_usd]; targetTimestamp is
// converted to ms via math.Trunc(ts)*1_000 before comparison.
func TestPrice(t *testing.T) {
	entries := [][priceInfoLength]float64{
		{1_000_000, 1.0},
		{2_000_000, 2.0},
		{3_000_000, 3.0},
	}

	tests := []struct {
		name  string
		cache [][priceInfoLength]float64
		ts    float64
		force bool
		want  float64
	}{
		// empty cache
		{"empty_force", nil, 1_000, true, 0},
		{"empty_no_force", nil, 1_000, false, 0},
		// ts before first entry (loop exits immediately, price still 0)
		{"before_first_force", entries, 0, true, 0},
		{"before_first_no_force", entries, 0, false, 0},
		// ts exactly at first entry boundary (equal is not >, entry is consumed)
		{"at_first_force", entries, 1_000, true, 1.0},
		{"at_first_no_force", entries, 1_000, false, 1.0},
		// ts between first and second entry
		{"between_1_2_force", entries, 1_500, true, 1.0},
		{"between_1_2_no_force", entries, 1_500, false, 1.0},
		// ts at last entry (loop exhausts)
		{"at_last_force", entries, 3_000, true, 3.0},
		{"at_last_no_force", entries, 3_000, false, 0},
		// ts beyond all entries (loop exhausts)
		{"after_last_force", entries, 4_000, true, 3.0},
		{"after_last_no_force", entries, 4_000, false, 0},
		// fractional ts — math.Trunc strips the decimal
		{"fractional_ts", entries, 1_999.9, false, 1.0},
		// no-key sentinel: {0,1.0},{MaxFloat64,1.0} must return 1.0 without force
		{"sentinel_no_force", [][priceInfoLength]float64{{0, 1.0}, {math.MaxFloat64, 1.0}}, 3_000_000, false, 1.0},
		{"sentinel_force", [][priceInfoLength]float64{{0, 1.0}, {math.MaxFloat64, 1.0}}, 3_000_000, true, 1.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &tickerService{cachedPrices: tc.cache}
			assert.Equal(t, tc.want, s.price(tc.ts, tc.force))
		})
	}
}

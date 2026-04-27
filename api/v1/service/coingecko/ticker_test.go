package coingecko

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

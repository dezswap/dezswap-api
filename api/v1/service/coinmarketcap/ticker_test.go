package coinmarketcap

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
	"base_address", "base_name", "base_symbol",
	"quote_address", "quote_name", "quote_symbol",
	"base_volume", "quote_volume", "last_price",
	"quote_decimals", "base_decimals", "timestamp",
}

// TestInactivePools_ExcludesActivePoolIds verifies that active pool IDs are
// passed as bound parameters instead of being interpolated into SQL.
func TestInactivePools_ExcludesActivePoolIds(t *testing.T) {
	svc, mock, close := setupTickerServiceWithMock(t)
	defer close()

	mock.ExpectQuery(`from pair_stats_30m`).
		WithArgs("test-chain", "xpla1active1", "xpla1active2").
		WillReturnRows(
			sqlmock.NewRows(inactivePoolColumns).
				AddRow("addrA", "TokenA", "TKA", "addrB", "TokenB", "TKB", "0", "0", "1.0", 6, 6, 1000000.0),
		)

	tickers, err := svc.inactivePools([]string{"xpla1active1", "xpla1active2"})
	require.NoError(t, err)
	require.Len(t, tickers, 1)
	require.Equal(t, "1.0", tickers[0].LastPrice)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestInactivePools_ReturnsAllWhenNoActiveIds verifies that passing no active
// pool IDs runs the base query with only the chain ID bound.
func TestInactivePools_ReturnsAllWhenNoActiveIds(t *testing.T) {
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

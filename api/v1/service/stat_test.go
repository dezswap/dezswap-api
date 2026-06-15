package service

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupServiceWithMock(t *testing.T) (*statService, sqlmock.Sqlmock, func() error) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(
		postgres.New(postgres.Config{
			Conn: sqlDB,
		}),
		&gorm.Config{},
	)
	require.NoError(t, err)

	service := &statService{
		chainId: "test-chain",
		DB:      gormDB,
	}

	return service, mock, sqlDB.Close
}

func TestStatService_GetAll_Empty(t *testing.T) {
	service, mock, close := setupServiceWithMock(t)
	defer close()

	mock.ExpectQuery(`from pair_stats_30m`).
		WithArgs("test-chain", "1mon").
		WillReturnRows(
			sqlmock.NewRows([]string{
				"address",
				"volume0_in_price",
				"volume1_in_price",
				"commission0_in_price",
				"commission1_in_price",
				"liquidity0_in_price",
				"liquidity1_in_price",
				"timestamp",
			}),
		)

	actual, err := service.GetAll()
	require.NoError(t, err)
	require.Len(t, actual, int(CountOfPeriodType))

	require.Empty(t, actual[Period24h])
	require.Nil(t, actual[Period7d])
	require.Nil(t, actual[Period1mon])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStatService_MapToSlice_AnnualizesAprByPeriod(t *testing.T) {
	service, _, close := setupServiceWithMock(t)
	defer close()

	pairStatMap := map[string][countOfStatType]math.LegacyDec{
		"pair": {
			statVolume:     math.LegacyNewDec(1000),
			statCommission: math.LegacyNewDec(7),
			statLiquidity:  math.LegacyNewDec(100),
		},
	}

	stats24h := service.mapToSlice(pairStatMap, aprAnnualizationByPeriod[statPeriod24h])
	require.Len(t, stats24h, 1)
	require.Equal(t, "2555.000000000000000000", stats24h[0].AprInPrice)

	stats7d := service.mapToSlice(pairStatMap, aprAnnualizationByPeriod[statPeriod7d])
	require.Len(t, stats7d, 1)
	require.Equal(t, "365.000000000000000000", stats7d[0].AprInPrice)

	stats1mon := service.mapToSlice(pairStatMap, aprAnnualizationByPeriod[statPeriod1mon])
	require.Len(t, stats1mon, 1)
	require.Equal(t, "84.000000000000000000", stats1mon[0].AprInPrice)
}

func TestStatService_MapToSlice_ZeroLiquidity(t *testing.T) {
	service, _, close := setupServiceWithMock(t)
	defer close()

	pairStatMap := map[string][countOfStatType]math.LegacyDec{
		"pair": {
			statVolume:     math.LegacyNewDec(1000),
			statCommission: math.LegacyNewDec(7),
			statLiquidity:  math.LegacyNewDec(0),
		},
	}

	stats := service.mapToSlice(pairStatMap, aprAnnualizationByPeriod[statPeriod24h])
	require.Len(t, stats, 1)
	require.Equal(t, "0", stats[0].AprInPrice, "APR should be 0 when liquidity is 0")
}

package service

import (
	"testing"

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

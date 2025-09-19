package dashboard

import (
	"fmt"
	"github.com/dezswap/dezswap-api/configs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strconv"
	"testing"
	"time"
)

var (
	testChainID           = "testchain-1"
	testPairContractAddr1 = "test1abcd"

	testPairContractAddr2 = "test1efgh"
	testTokenAddr         = "xerc20:ABCD"
	tsData                = time.Time{}
)

func SetupDB(t *testing.T) *gorm.DB {
	t.Helper()

	c := configs.NewWithFileName("test_supplements/config")
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Api.DB.Host, c.Api.DB.Port, c.Api.DB.Username, c.Api.DB.Password, c.Api.DB.Database,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`
INSERT INTO tokens (chain_id, address, name, symbol, decimals) VALUES (?, ?, 'Abcd', 'ABCD', 18)
`, testChainID, testTokenAddr).Error)

	require.NoError(t, db.Exec(`
INSERT INTO tokens (chain_id, address, name, symbol, decimals, verified) VALUES (?, 'axpla', 'XPLA', 'XPLA', 18, true)
`, testChainID).Error)

	require.NoError(t, db.Exec(`
INSERT INTO pair (chain_id, contract, asset0, asset1, lp) VALUES (?, ?, 'xpla1asset0', 'xpla1asset1', 'xpla1lp1')
`, testChainID, testPairContractAddr1).Error)

	var pairID int64
	row := db.Raw(`
INSERT INTO pair (chain_id, contract, asset0, asset1, lp) VALUES (?, ?, 'axpla', ?, 'xpla1lp2') RETURNING id
`, testChainID, testPairContractAddr2, testTokenAddr).Row()
	require.NoError(t, row.Err())
	require.NoError(t, row.Scan(&pairID))

	tsData = time.Now()
	require.NoError(t, db.Exec(`
INSERT INTO pair_stats_30m (
    year_utc, month_utc, day_utc, hour_utc, minute_utc,
    pair_id, chain_id,
    volume0, volume1, volume0_in_price, volume1_in_price,
    last_swap_price,
    liquidity0, liquidity1, liquidity0_in_price, liquidity1_in_price,
    commission0, commission1, commission0_in_price, commission1_in_price,
    price_token,
    tx_cnt, provider_cnt,
    timestamp,
    created_at, modified_at
)
VALUES (
    ?, ?,?, ?, ?,
    ?, ?,
    100, 200, 123.45, 678.90,
    1.23,
    1000, 2000, 1111.11, 2222.22,
    10, 20, 12.34, 56.78,
    'ibc/ABCD',
    5, 2,
    ?,
    ?, ?
)
`, tsData.Year(), tsData.Month(), tsData.Day(), tsData.Hour(), tsData.Minute(), pairID, testChainID, tsData.Unix(), tsData.Unix(), tsData.Unix()).Error)

	return db
}

func CleanupDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	require.NoError(t, db.Exec(`DELETE FROM pair_stats_30m WHERE pair_id IN (SELECT id FROM pair WHERE chain_id = ?)`, testChainID).Error)
	require.NoError(t, db.Exec(`DELETE FROM pair WHERE chain_id = ?`, testChainID).Error)
	require.NoError(t, db.Exec(`DELETE FROM tokens WHERE chain_id = ?`, testChainID).Error)

	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

func TestRecentOf_NoDivisionByZero_WithPrevZeros(t *testing.T) {
	db := SetupDB(t)
	defer CleanupDB(t, db)

	d := &dashboard{DB: db, chainId: testChainID}
	recent, err := d.RecentOf(Addr(testPairContractAddr1))
	require.NoError(t, err)

	assert.True(t, recent.PoolExists)
	assert.Equal(t, "0", recent.Volume)
	assert.Equal(t, float32(0), recent.VolumeChangeRate)
	assert.Equal(t, "0.000000", recent.Fee)
	assert.Equal(t, float32(0), recent.FeeChangeRate)
	assert.Equal(t, "0", recent.Tvl)
	assert.Equal(t, float32(0), recent.TvlChangeRate)
	assert.Equal(t, float32(0), recent.Apr)
	assert.Equal(t, float32(0), recent.AprChangeRate)
}

func TestTestTokenVolumes(t *testing.T) {
	db := SetupDB(t)
	defer CleanupDB(t, db)

	d := &dashboard{DB: db, chainId: testChainID}
	addr := Addr(testTokenAddr)

	t.Run("Month interval", func(t *testing.T) {
		chart, err := d.TokenVolumes(addr, Month)
		require.NoError(t, err)
		require.NotNil(t, chart)
		require.True(t, len(chart) > 0)

		_, err = strconv.ParseInt(chart[0].Timestamp, 10, 64)
		require.NoError(t, err)
	})

	t.Run("Quarter interval", func(t *testing.T) {
		chart, err := d.TokenVolumes(addr, Quarter)
		require.NoError(t, err)
		require.True(t, len(chart) > 0)

		_, err = strconv.ParseInt(chart[0].Timestamp, 10, 64)
		require.NoError(t, err)
	})

	t.Run("Year interval", func(t *testing.T) {
		chart, err := d.TokenVolumes(addr, Year)
		require.NoError(t, err)
		require.True(t, len(chart) > 0)

		_, err = strconv.ParseInt(chart[0].Timestamp, 10, 64)
		require.NoError(t, err)
	})

	t.Run("All interval", func(t *testing.T) {
		chart, err := d.TokenVolumes(addr, All)
		require.NoError(t, err)
		require.True(t, len(chart) > 0)

		_, err = strconv.ParseInt(chart[0].Timestamp, 10, 64)
		require.NoError(t, err)
	})
}

func TestTestTokenTvls(t *testing.T) {
	db := SetupDB(t)
	defer CleanupDB(t, db)

	d := &dashboard{DB: db, chainId: testChainID}
	addr := Addr(testTokenAddr)

	t.Run("Month interval", func(t *testing.T) {
		chart, err := d.TokenTvls(addr, Month)
		require.NoError(t, err)
		require.NotNil(t, chart)
		require.True(t, len(chart) > 0)

		_, err = strconv.ParseInt(chart[0].Timestamp, 10, 64)
		require.NoError(t, err)
	})

	t.Run("Quarter interval", func(t *testing.T) {
		chart, err := d.TokenTvls(addr, Quarter)
		require.NoError(t, err)
		require.True(t, len(chart) > 0)

		_, err = strconv.ParseInt(chart[0].Timestamp, 10, 64)
		require.NoError(t, err)
	})

	t.Run("Year interval", func(t *testing.T) {
		chart, err := d.TokenTvls(addr, Year)
		require.NoError(t, err)
		require.True(t, len(chart) > 0)

		_, err = strconv.ParseInt(chart[0].Timestamp, 10, 64)
		require.NoError(t, err)
	})

	t.Run("Year interval", func(t *testing.T) {
		chart, err := d.TokenTvls(addr, All)
		require.NoError(t, err)
		require.True(t, len(chart) > 0)

		_, err = strconv.ParseInt(chart[0].Timestamp, 10, 64)
		require.NoError(t, err)
	})
}

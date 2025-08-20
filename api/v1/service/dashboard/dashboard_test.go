package dashboard

import (
	"fmt"
	"github.com/dezswap/dezswap-api/configs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
)

var testChainID = "testchain-1"
var testContract = "test1abcd"

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
INSERT INTO pair (chain_id, contract, asset0, asset1, lp) VALUES (?, ?, 'xpla1asset0', 'xpla1asset1', 'xpla1lp')
`, testChainID, testContract).Error)

	return db
}

func CleanupDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	require.NoError(t, db.Exec(`DELETE FROM pair_stats_30m WHERE pair_id IN (SELECT id FROM pair WHERE chain_id = ?)`, testChainID).Error)
	require.NoError(t, db.Exec(`DELETE FROM pair WHERE chain_id = ?`, testChainID).Error)

	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

func TestRecentOf_NoDivisionByZero_WithPrevZeros(t *testing.T) {
	db := SetupDB(t)
	defer CleanupDB(t, db)

	d := &dashboard{DB: db, chainId: testChainID}
	recent, err := d.RecentOf(Addr(testContract))
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

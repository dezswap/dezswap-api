package visibility

import (
	"fmt"
	"testing"

	"github.com/dezswap/dezswap-api/configs"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Temporary tables keep the visibility policy test independent of shared fixtures.
func TestVisibilityHidingAndRestoring(t *testing.T) {
	c := configs.NewWithFileName("test_supplements/config").Api.DB
	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Port, c.Username, c.Password, c.Database)), &gorm.Config{})
	require.NoError(t, err)
	conn, err := db.DB()
	require.NoError(t, err)
	defer conn.Close()
	tx := db.Begin()
	require.NoError(t, tx.Error)
	defer tx.Rollback()
	for _, stmt := range []string{
		`CREATE TEMP TABLE token_exception (chain_id text, contract text, hidden bool, skip_parse bool) ON COMMIT DROP`,
		`CREATE TEMP TABLE tokens (id int, chain_id text, address text) ON COMMIT DROP`,
		`CREATE TEMP TABLE pair (id int, chain_id text, asset0 text, asset1 text) ON COMMIT DROP`,
		`CREATE TEMP TABLE parsed_tx (id int, chain_id text, asset0 text, asset1 text) ON COMMIT DROP`,
		`CREATE TEMP TABLE route (id int, chain_id text, asset0 text, asset1 text, route text[], deleted_at float8) ON COMMIT DROP`,
		`CREATE TEMP TABLE pair_stats_30m (pair_id int, chain_id text, volume int) ON COMMIT DROP`,
		`CREATE TEMP TABLE pair_stats_recent (pair_id int, chain_id text, volume int) ON COMMIT DROP`,
		`CREATE TEMP TABLE price (id int, chain_id text, route_id int) ON COMMIT DROP`,
		`INSERT INTO tokens VALUES (1,'a','good'),(2,'a','bad'),(3,'b','bad')`,
		`INSERT INTO pair VALUES (1,'a','good','other'),(2,'a','good','bad'),(3,'b','good','bad')`,
		`INSERT INTO parsed_tx SELECT * FROM pair`,
		`INSERT INTO pair_stats_30m VALUES (1,'a',10),(2,'a',100),(3,'b',1000)`,
		`INSERT INTO pair_stats_recent SELECT * FROM pair_stats_30m`,
		`INSERT INTO route VALUES (1,'a','good','other','{}',NULL),(2,'a','good','other','{bad}',NULL),(3,'a','bad','good','{}',NULL),(4,'a','good','other','{}',1),(5,'b','good','bad','{}',NULL)`,
		`INSERT INTO price SELECT id,chain_id,id FROM route`,
		`INSERT INTO token_exception VALUES ('a','bad',false,true)`,
	} {
		require.NoError(t, tx.Exec(stmt).Error)
	}
	count := func(table string, expected int64) {
		t.Helper()
		var actual int64
		require.NoError(t, tx.Raw(Query("SELECT count(*) FROM "+table)).Scan(&actual).Error)
		require.Equal(t, expected, actual, table)
	}
	// skip_parse alone does not hide anything. A retired route is always excluded.
	count("tokens", 3)
	count("route", 4)
	require.NoError(t, tx.Exec(`UPDATE token_exception SET hidden = true, skip_parse = false`).Error)
	for table, expected := range map[string]int64{"tokens": 2, "pair": 2, "parsed_tx": 2, "route": 2, "price": 2, "pair_stats_30m": 2, "pair_stats_recent": 2} {
		count(table, expected)
	}
	var volume int
	require.NoError(t, tx.Raw(Query(`WITH totals AS (SELECT sum(volume) AS volume FROM pair_stats_30m WHERE chain_id = ?) SELECT volume FROM totals`), "a").Scan(&volume).Error)
	require.Equal(t, 10, volume)
	// Endpoint predicates agree with reporting CTEs, including intermediate hops.
	var routes int64
	require.NoError(t, tx.Table("route").Where(Route("route")).Count(&routes).Error)
	require.EqualValues(t, 2, routes)
	require.NoError(t, tx.Exec(`UPDATE token_exception SET hidden = false`).Error)
	count("tokens", 3)
	count("pair", 3)
	count("route", 4)
	count("price", 4)
}

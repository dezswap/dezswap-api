package aggregator

import (
	"github.com/dezswap/cosmwasm-etl/pkg/db/schemas"
)

// / table schemas of cosmwasm-etl
type Account = schemas.Account
type LpHistory = schemas.LpHistory
type Price = schemas.Price
type Route = schemas.Route
type ParsedTxWithPrice = schemas.ParsedTxWithPrice
type PairStatsRecent = schemas.PairStatsRecent
type PairStats30m = schemas.PairStats30m
type HAccountStats30m = schemas.HAccountStats30m

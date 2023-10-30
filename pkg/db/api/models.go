package api

import (
	"github.com/dezswap/cosmwasm-etl/pkg/db/schemas"
)

// / table schemas of cosmwasm-etl
type Account = schemas.Account
type LpHistory = schemas.LpHistory
type Price = schemas.Price
type Route = schemas.Route
type ParsedTxWithPrice = schemas.ParsedTxWithPrice
type PairStatsIn24h = schemas.PairStatsIn24h
type PairStats30m = schemas.PairStats30m
type HAccountStats30m = schemas.HAccountStats30m

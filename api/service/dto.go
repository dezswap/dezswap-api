package service

import (
	"github.com/dezswap/dezswap-api/pkg/db/indexer"
)

type Pair struct {
	ID      string
	ChainId string
	Address string
	Asset0  Token `gorm:"embedded;embeddedPrefix:asset0_"`
	Asset1  Token `gorm:"embedded;embeddedPrefix:asset1_"`
	Lp      Token `gorm:"embedded;embeddedPrefix:lp_"`
}

type Ticker struct {
	BaseCurrency         string
	TargetCurrency       string
	LastPrice            string
	BaseVolume           string
	TargetVolume         string
	BaseDecimals         int
	TargetDecimals       int
	BaseLiquidityInPrice string
	PoolId               string
	Timestamp            float64
}

type Pool = indexer.LatestPool

type Token = indexer.Token

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

type Pool = indexer.LatestPool

type Token = indexer.Token

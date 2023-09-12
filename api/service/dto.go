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

type PairStat struct {
	Address           string
	VolumeInPrice     string
	CommissionInPrice string
	AprInPrice        string
}

type PeriodTypeIdx int

const (
	Period24h PeriodTypeIdx = 0 + iota
	Period7d
	Period1mon
	CountOfPeriodType
)

type PairStats []PairStat

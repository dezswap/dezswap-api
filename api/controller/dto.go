package controller

import (
	"github.com/dezswap/dezswap-api/pkg/dezswap"
)

type PoolsRes []PoolRes

type PoolRes struct {
	Address string `json:"address"`
	*dezswap.PoolRes
}

type TokensRes []TokenRes

type TokenRes struct{}

type PairsRes []TokenRes

type PairRes struct{}

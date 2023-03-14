package controller

import (
	"github.com/dezswap/dezswap-api/api/service"
	"github.com/dezswap/dezswap-api/pkg/dezswap"
)

type PoolsRes []PoolRes

type PoolRes struct {
	Address string `json:"address"`
	*dezswap.PoolRes
}

type TokensRes []TokenRes

type TokenRes struct {
	*service.Token
}

type PairsRes []TokenRes

type PairRes struct {
	*service.Pair
}

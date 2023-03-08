package controller

import (
	"github.com/dezswap/dezswap-api/pkg/dezswap"
)

type PairRes struct {
}

type PoolsRes []PoolRes

type PoolRes struct {
	*dezswap.PoolRes
}

type TokensRes []TokenRes

type TokenRes struct {
}

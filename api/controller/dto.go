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

type TokenRes struct {
	ChainId     string `json:"chainId"`
	Token       string `json:"token"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Decimals    uint8  `json:"decimals"`
	TotalSupply string `json:"total_supply" `
	Icon        string `json:"icon"`
	Protocol    string `json:"protocol"`
	Verified    bool   `json:"verified"`
}

type PairsRes struct {
	Pairs []PairRes `json:"pairs"`
}

type PairRes struct {
	*dezswap.PairRes
}

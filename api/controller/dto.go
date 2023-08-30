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

type TickersRes []TickerRes

type TickerRes struct {
	TickerId       string  `json:"ticker_id"`
	BaseCurrency   string  `json:"base_currency"`
	TargetCurrency string  `json:"target_currency"`
	LastPrice      float64 `json:"last_price"`
	BaseVolume     float64 `json:"base_volume"`
	TargetVolume   float64 `json:"target_volume" `
	PoolId         string  `json:"pool_id"`
	LiquidityInUsd float64 `json:"liquidity_in_usd"`
}

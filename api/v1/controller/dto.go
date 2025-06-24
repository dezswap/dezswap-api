package controller

import (
	"github.com/dezswap/dezswap-api/pkg/dezswap"
)

type HealthDependency struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type HealthResponse struct {
	Status       string             `json:"status"`
	Timestamp    string             `json:"timestamp"`
	Dependencies []HealthDependency `json:"dependencies"`
}

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

type StatsRes struct {
	Stats24h  StatRes `json:"stats_24h"`
	Stats7d   StatRes `json:"stats_7d"`
	Stats1mon StatRes `json:"stats_1mon"`
}

type StatRes struct {
	Volume []StatValueRes `json:"volume"`
	Fee    []StatValueRes `json:"fee"`
	Apr    []StatValueRes `json:"apr"`
}

type StatValueRes struct {
	Address string `json:"address"`
	Value   string `json:"value"`
}

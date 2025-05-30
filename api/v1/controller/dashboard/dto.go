package dashboard

import (
	"strings"
	"time"
)

type TxType string

const (
	TX_TYPE_SWAP   TxType = "swap"
	TX_TYPE_ADD    TxType = "add"
	TX_TYPE_REMOVE TxType = "remove"
	TX_TYPE_ALL    TxType = ""
)

type RecentRes struct {
	Volume           string  `json:"volume"`
	VolumeChangeRate float32 `json:"volumeChangeRate"`
	Fee              string  `json:"fee"`
	FeeChangeRate    float32 `json:"feeChangeRate"`
	Tvl              string  `json:"tvl"`
	TvlChangeRate    float32 `json:"tvlChangeRate"`
	Apr              float32 `json:"apr"`
	AprChangeRate    float32 `json:"aprChangeRate"`
}

type StatisticRes []StatisticResItem
type StatisticResItem struct {
	AddressCount uint64    `json:"addressCount"`
	TxCount      uint64    `json:"txCount"`
	Fee          string    `json:"fee"`
	Timestamp    time.Time `json:"timestamp"`
}

type PoolsRes []PoolRes

type PoolRes struct {
	Address string `json:"address"`
	Tvl     string `json:"tvl"`
	Volume  string `json:"volume"`
	Fee     string `json:"fee"`
	Apr     string `json:"apr"`
}

type PoolDetailRes struct {
	Recent RecentRes `json:"recent"`
	Txs    TxsRes    `json:"txs"`
}

type TokensRes []TokenRes
type TokenRes struct {
	Address         string  `json:"address"`
	Price           string  `json:"price"`
	PriceChange     float32 `json:"priceChange"`
	Volume24h       string  `json:"volume24h"`
	Volume24hChange string  `json:"volume24hChange,omitempty"`
	Volume7d        string  `json:"volume7d,omitempty"`
	Volume7dChange  string  `json:"volume7dChange,omitempty"`
	Tvl             string  `json:"tvl"`
	TvlChange       string  `json:"tvlChange,omitempty"`
	Fee             string  `json:"fee,omitempty"`
}

type TxsRes []TxRes
type TxRes struct {
	Action        string `json:"action"`
	ActionDisplay string `json:"actionDisplay"`

	Address      string    `json:"address"`
	Hash         string    `json:"hash"`
	TotalValue   string    `json:"totalValue"`
	Asset0       string    `json:"asset0"`
	Asset0Amount string    `json:"asset0amount"`
	Asset1       string    `json:"asset1"`
	Asset1Amount string    `json:"asset1amount"`
	Account      string    `json:"account"`
	Timestamp    time.Time `json:"timestamp"`
}

type ChartType = string

const (
	ChartTypeVolume ChartType = "volume"
	ChartTypeTvl    ChartType = "tvl"
	ChartTypeApr    ChartType = "apr"
	ChartTypeFee    ChartType = "fee"
	ChartTypePrice  ChartType = "price"
	ChartTypeNone   ChartType = ""
)

func ToChartType(s string) ChartType {
	switch strings.ToLower(s) {
	case "volume":
		return ChartTypeVolume
	case "tvl":
		return ChartTypeTvl
	case "apr":
		return ChartTypeApr
	case "fee":
		return ChartTypeFee
	case "price":
		return ChartTypePrice
	default:
		return ChartTypeNone
	}
}

type ChartRes []ChartItem
type ChartItem struct {
	Timestamp time.Time `json:"t"`
	Value     string    `json:"v"`
}

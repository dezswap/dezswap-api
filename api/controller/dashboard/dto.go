package dashboard

import (
	"strings"
	"time"
)

type RecentRes struct {
	Volume           string  `json:"volume"`
	VolumeChangeRate float32 `json:"volumeChangeRate"`
	Fee              string  `json:"fee"`
	FeeChangeRate    float32 `json:"feeChangeRate"`
	Tvl              string  `json:"tvl"`
	TvlChangeRate    float32 `json:"tvlChangeRate"`
}

type VolumesRequest struct {
	Duration string `form:"duration"`
}
type VolumesRes = []VolumeRes
type VolumeRes struct {
	Volume    string    `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

type TvlsRes = []TvlRes
type TvlRes struct {
	Tvl       string    `json:"tvl"`
	Timestamp time.Time `json:"timestamp"`
}

type AprsRes = []AprRes
type AprRes struct {
	Apr       string    `json:"apr"`
	Timestamp time.Time `json:"timestamp"`
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

type ChartRequest struct{}

type TokensRes []TokenRes
type TokenRes struct {
	Address         string  `json:"address"`
	Price           string  `json:"price"`
	PriceChange     float32 `json:"price_change"`
	Volume24h       string  `json:"volume_24h"`
	Volume24hChange string  `json:"volume_24h_change,omitempty"`
	Volume7d        string  `json:"volume_7d,omitempty"`
	Volume7dChange  string  `json:"volume_7d_change,omitempty"`
	Tvl             string  `json:"tvl"`
	TvlChange       string  `json:"tvl_change,omitempty"`
	Fee             string  `json:"fee,omitempty"`
}

type TokenChart []TimestampValue

type TimestampValue [2]string

type TxsRes []TxRes
type TxRes struct {
	Action       string    `json:"action"`
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
	default:
		return ChartTypeNone
	}
}

type ChartRes []ChartItem
type ChartItem struct {
	Timestamp time.Time `json:"t"`
	Value     string    `json:"v"`
}

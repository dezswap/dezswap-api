package dashboard

import "time"

type Addr string

type Recent struct {
	Volume           string  `json:"volume"`
	VolumeChangeRate float32 `json:"volumeChangeRate"`
	Fee              string  `json:"fee"`
	FeeChangeRate    float32 `json:"feeChangeRate"`
	Tvl              string  `json:"tvl"`
	TvlChangeRate    float32 `json:"tvlChangeRate"`
}

type Volumes = []Volume
type Volume struct {
	Volume    string    `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

type Tvls = []Tvl
type Tvl struct {
	Tvl       string    `json:"tvl"`
	Timestamp time.Time `json:"timestamp"`
}

type Statistic struct {
	AddsCounts AddsCounts `json:"addsCounts"`
	TxCounts   TxCounts   `json:"txCounts"`
	Fees       Fees       `json:"fees"`
}

type AddsCounts = []AddsCount
type AddsCount struct {
	AddsCount uint64    `json:"addsCount"`
	Timestamp time.Time `json:"timestamp"`
}

type TxCounts = []TxCount
type TxCount struct {
	TxCount   uint64    `json:"txCount"`
	Timestamp time.Time `json:"timestamp"`
}

type Fees = []Fee
type Fee struct {
	Fee       string    `json:"fee"`
	Timestamp time.Time `json:"timestamp"`
}

type Pools []Pool

type Pool struct {
	Adds   string `json:"adds"`
	Tvl    string `json:"tvl"`
	Volume string `json:"volume"`
	Fee    string `json:"fee"`
	Apr    string `json:"apr"`
}

type Tokens []Token
type Token struct {
	Adds        string  `json:"adds"`
	Price       string  `json:"price"`
	PriceChange float32 `json:"priceChange"`
	Volume      string  `json:"volume"`
	Tvl         string  `json:"tvl"`
}

type Txs []Tx
type Tx struct {
	Action       string    `json:"action"`
	TotalValue   string    `json:"totalValue"`
	Asset0Amount string    `json:"asset0amount"`
	Asset1Amount string    `json:"asset1amount"`
	Sender       string    `json:"sender"`
	Time         time.Time `json:"time"`
}

type Prices = []Price
type Price struct {
	Price     string    `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

type Aprs = []Apr
type Apr struct {
	Apr       string    `json:"apr"`
	Timestamp time.Time `json:"timestamp"`
}

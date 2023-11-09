package dashboard

import "time"

type ActiveAccounts = []ActiveAccount
type ActiveAccount struct {
	ActiveAccount uint64
	Timestamp     time.Time
}

type Addr string

type Recent struct {
	Volume           string
	VolumeChangeRate float32
	Fee              string
	FeeChangeRate    float32
	Tvl              string
	TvlChangeRate    float32
}

type Volumes = []Volume
type Volume struct {
	Volume    string
	Timestamp time.Time
}

type Tvls = []Tvl
type Tvl struct {
	Tvl       string
	Timestamp time.Time
}

type Statistic struct {
	AddsCounts AddsCounts
	TxCounts   TxCounts
	Fees       Fees
}

type AddsCounts = []AddsCount
type AddsCount struct {
	AddressCount uint64
	Timestamp    time.Time
}

type TxCounts = []TxCount
type TxCount struct {
	TxCount   uint64
	Timestamp time.Time
}

type Fees = []Fee
type Fee struct {
	Fee       string
	Timestamp time.Time
}

type Pools []Pool

type Pool struct {
	Adds   string
	Tvl    string
	Volume string
	Fee    string
	Apr    string
}

type Tokens []Token
type Token struct {
	Addr        Addr
	Price       string
	PriceChange float32
	Volume      string
	Tvl         string
}

type Txs []Tx
type Tx struct {
	Action       string
	TotalValue   string
	Asset0Amount string
	Asset1Amount string
	Sender       string
	Time         time.Time
}

type Prices = []Price
type Price struct {
	Price     string
	Timestamp time.Time
}

type Aprs = []Apr
type Apr struct {
	Apr       string
	Timestamp time.Time
}

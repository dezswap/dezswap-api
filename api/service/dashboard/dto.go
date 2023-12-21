package dashboard

import "time"

type Addr string
type Duration string

const (
	Month   Duration = "month"
	Quarter Duration = "quarter"
	Year    Duration = "year"
	All     Duration = "all"
)

type ActiveAccounts = []ActiveAccount
type ActiveAccount struct {
	ActiveAccount uint64
	Timestamp     time.Time
}

type Recent struct {
	Volume           string  `gorm:"volume"`
	VolumeChangeRate float32 `gorm:"volume_change_rate"`
	Fee              string  `gorm:"fee"`
	FeeChangeRate    float32 `gorm:"fee_change_rate"`
	Tvl              string  `gorm:"tvl"`
	TvlChangeRate    float32 `gorm:"tvl_change_rate"`
	Apr              float32 `gorm:"apr"`
	AprChangeRate    float32 `gorm:"apr_change_rate"`
}

type Fees = []Fee
type Fee struct {
	Fee       string
	Timestamp time.Time
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

type Statistic = []StatisticItem
type StatisticItem struct {
	AddressCount uint64
	TxCount      uint64
	Fee          string
	Timestamp    time.Time
}

type Pools []struct {
	Address string
	Symbols string
	Tvl     string
	Volume  string
	Fee     string
	Apr     string
}

type PoolDetail struct {
	Recent Recent
	Txs    Txs
}

type Tokens []Token
type Token struct {
	Addr           Addr
	Price          string
	PriceChange    float32
	Volume         string
	VolumeChange   string
	Volume7d       string
	Volume7dChange string
	Tvl            string
	TvlChange      string
	Commission     string
}

type TokenValue struct {
	Timestamp string
	Value     string
}

type TokenChart []TokenValue

type TxType string

const (
	TX_TYPE_SWAP     TxType = "swap"
	TX_TYPE_PROVIDE  TxType = "provide"
	TX_TYPE_WITHDRAW TxType = "withdraw"
	TX_TYPE_ALL      TxType = ""
)

type Txs []Tx
type Tx struct {
	Action       string
	Hash         string
	Sender       string
	Address      string
	Asset0       string
	Asset0Symbol string
	Asset0Amount string
	Asset1       string
	Asset1Symbol string
	Asset1Amount string
	TotalValue   string
	Timestamp    time.Time
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

package dashboard

type Duration string

const (
	Year    Duration = "year"
	Quarter Duration = "quarter"
	Month   Duration = "month"
)

type RecentRes struct {
	Volume           string  `json:"volume"`
	VolumeChangeRate float32 `json:"volumeChangeRate"`
	Fee              string  `json:"fee"`
	FeeChangeRate    float32 `json:"feeChangeRate"`
	TVL              string  `json:"tvl"`
	TVLChangeRate    float32 `json:"tvlChangeRate"`
}

type VolumesRequest struct {
	Duration string `form:"duration"`
}
type VolumesRes = []VolumeRes
type VolumeRes struct {
	Volume    string `json:"volume"`
	Timestamp string `json:"timestamp"`
}

type TVLsRes = []TVLRes
type TVLRes struct {
	TVL       string `json:"tvl"`
	Timestamp string `json:"timestamp"`
}

type StatisticRes struct {
	AddressCountsRes AddressCountsRes `json:"addressCounts"`
	TxCountsRes      TxCountsRes      `json:"txCounts"`
	FeesRes          FeesRes          `json:"fees"`
}

type AddressCountsRes = []AddressCountRes
type AddressCountRes struct {
	AddressCount uint64 `json:"addressCount"`
	Timestamp    string `json:"timestamp"`
}

type TxCountsRes = []TxCountRes
type TxCountRes struct {
	TxCount   uint64 `json:"txCount"`
	Timestamp string `json:"timestamp"`
}

type FeesRes = []FeeRes
type FeeRes struct {
	Fee       string `json:"fee"`
	Timestamp string `json:"timestamp"`
}

type PoolsRes []PoolRes

type PoolRes struct {
	Address string `json:"address"`
	TVL     string `json:"tvl"`
	Volume  string `json:"volume"`
	Fee     string `json:"fee"`
	APR     string `json:"apr"`
}

type TokensRes []TokenRes
type TokenRes struct {
	Address     string  `json:"address"`
	Price       string  `json:"price"`
	PriceChange float32 `json:"priceChange"`
	Volume      string  `json:"volume"`
	TVL         string  `json:"tvl"`
}

type TxReses []TxRes
type TxRes struct {
	Action       string `json:"action"`
	TotalValue   string `json:"totalValue"`
	Asset0Amount string `json:"asset0amount"`
	Asset1Amount string `json:"asset1amount"`
	Sender       string `json:"sender"`
	Time         string `json:"time"`
}

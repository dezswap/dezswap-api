package coingecko

type Ticker struct {
	BaseCurrency         string
	TargetCurrency       string
	LastPrice            string
	BaseVolume           string
	TargetVolume         string
	BaseDecimals         int
	TargetDecimals       int
	BaseLiquidityInPrice string
	PoolId               string
	Timestamp            float64
}

type Pair struct {
	TickerId string
	Base     string
	Target   string
	PoolId   string
}

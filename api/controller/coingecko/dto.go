package coingecko

type PairsRes []PairRes

type PairRes struct {
	TickerId string `json:"ticker_id"`
	Base     string `json:"base"`
	Target   string `json:"target"`
	PoolId   string `json:"pool_id"`
}

type TickersRes []TickerRes

type TickerRes struct {
	TickerId       string `json:"ticker_id"`
	BaseCurrency   string `json:"base_currency"`
	TargetCurrency string `json:"target_currency"`
	LastPrice      string `json:"last_price"`
	BaseVolume     string `json:"base_volume"`
	TargetVolume   string `json:"target_volume"`
	PoolId         string `json:"pool_id"`
	LiquidityInUsd string `json:"liquidity_in_usd"`
}

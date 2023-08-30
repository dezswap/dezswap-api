package coingecko

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

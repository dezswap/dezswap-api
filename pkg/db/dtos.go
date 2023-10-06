package db

type LastIdLimitCondition struct {
	LastId    uint
	Limit     int
	DescOrder bool
}

type PairStat struct {
	Address            string  `json:"address"`
	Volume0InPrice     string  `json:"volume0_in_price"`
	Volume1InPrice     string  `json:"volume1_in_price"`
	Commission0InPrice string  `json:"commission0_in_price"`
	Commission1InPrice string  `json:"commission1_in_price"`
	Liquidity0InPrice  string  `json:"liquidity0_in_price"`
	Liquidity1InPrice  string  `json:"liquidity1_in_price"`
	Timestamp          float64 `json:"timestamp"`
}

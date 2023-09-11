package coinmarketcap

type Ticker struct {
	BaseAddress   string
	BaseName      string
	BaseSymbol    string
	QuoteAddress  string
	QuoteName     string
	QuoteSymbol   string
	LastPrice     string
	BaseVolume    string
	QuoteVolume   string
	BaseDecimals  int
	QuoteDecimals int
	PoolId        string
	Timestamp     float64
}

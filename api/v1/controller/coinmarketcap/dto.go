package coinmarketcap

type TickersRes map[string]TickerRes

type TickerRes struct {
	BaseId      string `json:"base_id"`
	BaseName    string `json:"base_name"`
	BaseSymbol  string `json:"base_symbol"`
	QuoteId     string `json:"quote_id"`
	QuoteName   string `json:"quote_name"`
	QuoteSymbol string `json:"quote_symbol"`
	LastPrice   string `json:"last_price"`
	BaseVolume  string `json:"base_volume"`
	QuoteVolume string `json:"quote_volume"`
}

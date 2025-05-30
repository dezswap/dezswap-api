package coinmarketcap

import (
	coinMarketCapService "github.com/dezswap/dezswap-api/api/v1/service/coinmarketcap"
)

type tickerMapper struct{}

func (m *tickerMapper) tickerToRes(ticker coinMarketCapService.Ticker) TickerRes {
	return TickerRes{
		BaseId:      ticker.BaseAddress,
		BaseSymbol:  ticker.BaseSymbol,
		QuoteId:     ticker.QuoteAddress,
		QuoteName:   ticker.QuoteName,
		QuoteSymbol: ticker.QuoteSymbol,
		LastPrice:   ticker.LastPrice,
		BaseVolume:  ticker.BaseVolume,
		QuoteVolume: ticker.QuoteVolume,
	}
}

func (m *tickerMapper) tickersToRes(tickers []coinMarketCapService.Ticker) TickersRes {
	res := make(map[string]TickerRes, len(tickers))
	for _, t := range tickers {
		res[t.BaseAddress+"_"+t.QuoteAddress] = m.tickerToRes(t)
	}

	return res
}

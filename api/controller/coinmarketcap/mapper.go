package coinmarketcap

import (
	coinMarketCapService "github.com/dezswap/dezswap-api/api/service/coinmarketcap"
)

type tickerMapper struct{}

func (m *tickerMapper) tickerToRes(ticker coinMarketCapService.Ticker) (TickerRes, error) {
	return TickerRes{
		BaseId:      ticker.BaseAddress,
		BaseSymbol:  ticker.BaseSymbol,
		QuoteId:     ticker.QuoteAddress,
		QuoteName:   ticker.QuoteName,
		QuoteSymbol: ticker.QuoteSymbol,
		LastPrice:   ticker.LastPrice,
		BaseVolume:  ticker.BaseVolume,
		QuoteVolume: ticker.QuoteVolume,
	}, nil
}

func (m *tickerMapper) tickersToRes(tickers []coinMarketCapService.Ticker) (TickersRes, error) {
	var err error
	res := make(map[string]TickerRes, len(tickers))
	for _, t := range tickers {
		res[t.BaseAddress+"_"+t.QuoteAddress], err = m.tickerToRes(t)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

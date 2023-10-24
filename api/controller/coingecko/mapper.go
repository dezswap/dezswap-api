package coingecko

import (
	coingeckoService "github.com/dezswap/dezswap-api/api/service/coingecko"
	"strconv"
)

type pairMapper struct{}

type tickerMapper struct{}

func (m *pairMapper) pairToRes(pair coingeckoService.Pair) PairRes {
	return PairRes{
		TickerId: pair.TickerId,
		Base:     pair.Base,
		Target:   pair.Target,
		PoolId:   pair.PoolId,
	}
}

func (m *pairMapper) pairsToRes(pairs []coingeckoService.Pair) PairsRes {
	res := make([]PairRes, len(pairs))
	for i, ticker := range pairs {
		res[i] = m.pairToRes(ticker)
	}
	return res
}

func (m *tickerMapper) tickerToRes(ticker coingeckoService.Ticker) (TickerRes, error) {
	lastPrice, err := strconv.ParseFloat(ticker.LastPrice, 64)
	if err != nil {
		return TickerRes{}, err
	}

	baseVolume, err := strconv.ParseFloat(ticker.BaseVolume, 64)
	if err != nil {
		return TickerRes{}, err
	}
	targetVolume, err := strconv.ParseFloat(ticker.TargetVolume, 64)
	if err != nil {
		return TickerRes{}, err
	}
	baseLiquidityInPrice, err := strconv.ParseFloat(ticker.BaseLiquidityInPrice, 64)
	if err != nil {
		return TickerRes{}, err
	}

	res := TickerRes{
		TickerId:       ticker.BaseCurrency + "_" + ticker.TargetCurrency,
		BaseCurrency:   ticker.BaseCurrency,
		TargetCurrency: ticker.TargetCurrency,
		LastPrice:      strconv.FormatFloat(lastPrice, 'f', -1, 64),
		BaseVolume:     strconv.FormatFloat(baseVolume, 'f', -1, 64),
		TargetVolume:   strconv.FormatFloat(targetVolume, 'f', -1, 64),
		PoolId:         ticker.PoolId,
		LiquidityInUsd: strconv.FormatFloat(baseLiquidityInPrice*2, 'f', -1, 64),
	}
	return res, nil
}

func (m *tickerMapper) tickersToRes(tickers []coingeckoService.Ticker) (TickersRes, error) {
	var err error
	res := make([]TickerRes, len(tickers))
	for i, ticker := range tickers {
		res[i], err = m.tickerToRes(ticker)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

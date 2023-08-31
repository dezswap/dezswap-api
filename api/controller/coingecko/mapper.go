package coingecko

import (
	"context"
	"encoding/json"
	coingeckoService "github.com/dezswap/dezswap-api/api/service/coingecko"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const priceTokenId = "axlusdc"
const coinGeckoEndpoint = "https://api.coingecko.com/api/v3/coins/"
const queryTimeout = 10 * time.Second

type priceInfo int

const (
	priceTimestamp priceInfo = 0 + iota
	priceValue
	priceInfoLength
)

type pairMapper struct{}

type tickerMapper struct {
	cachedPrices [][priceInfoLength]float64
}

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
	priceTokenInUsd := m.getPrice(ticker.Timestamp, false)
	if priceTokenInUsd == 0 {
		err = m.updateCachedPriceInUsd(priceTokenId)
		if err != nil {
			return TickerRes{}, err
		}
		priceTokenInUsd = m.getPrice(ticker.Timestamp, true)
	}

	res := TickerRes{
		TickerId:       ticker.BaseCurrency + "_" + ticker.TargetCurrency,
		BaseCurrency:   ticker.BaseCurrency,
		TargetCurrency: ticker.TargetCurrency,
		LastPrice:      lastPrice,
		BaseVolume:     baseVolume,
		TargetVolume:   targetVolume,
		PoolId:         ticker.PoolId,
		LiquidityInUsd: baseLiquidityInPrice * priceTokenInUsd,
	}
	return res, nil
}

// TODO: fixed endpoint and arguments
func (m *tickerMapper) updateCachedPriceInUsd(priceCoinId string) error {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	endpoint, err := url.Parse(coinGeckoEndpoint + priceCoinId + "/market_chart")
	if err != nil {
		return err
	}

	query := endpoint.Query()
	query.Add("vs_currency", "usd")
	query.Add("days", "1")
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(timeoutCtx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return err
	}

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		code := response.StatusCode
		return errors.New(strings.Join([]string{"Price endpoint returns http code [", http.StatusText(code), " ", strconv.Itoa(code), "]"}, ""))
	}

	type queryResponse struct {
		Prices [][priceInfoLength]float64
		///...
	}

	var decoded queryResponse
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&decoded)
	if err != nil {
		return err
	}

	m.cachedPrices = decoded.Prices

	return nil
}

func (m *tickerMapper) getPrice(targetTimestamp float64, force bool) float64 {
	price := float64(0)
	for _, p := range m.cachedPrices {
		if p[priceTimestamp] > targetTimestamp {
			return price
		}
		price = p[priceValue]
	}

	if force {
		return price
	}
	return 0
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

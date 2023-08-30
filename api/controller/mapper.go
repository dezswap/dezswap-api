package controller

import (
	"context"
	"encoding/json"
	"github.com/dezswap/dezswap-api/api/service"
	"github.com/dezswap/dezswap-api/pkg/dezswap"
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
type poolMapper struct{}
type tokenMapper struct{}
type tickerMapper struct {
	cachedPrices [][priceInfoLength]float64
}

func (m *poolMapper) poolToRes(pool service.Pool) PoolRes {
	res := PoolRes{
		Address: pool.Address,
		PoolRes: &dezswap.PoolRes{},
	}
	res.TotalShare = pool.LpAmount
	res.Assets = []dezswap.AssetInfoRes{
		dezswap.ToAssetInfoRes(pool.Asset0, pool.Asset0Amount),
		dezswap.ToAssetInfoRes(pool.Asset1, pool.Asset1Amount),
	}
	return res
}
func (m *poolMapper) poolsToRes(pools []service.Pool) []PoolRes {
	res := make([]PoolRes, len(pools))
	for i, pool := range pools {
		res[i] = m.poolToRes(pool)
	}
	return res
}

func (m *pairMapper) pairToRes(pair service.Pair) PairRes {
	res := PairRes{
		PairRes: &dezswap.PairRes{
			ContractAddr: pair.Address,
			AssetInfos: []dezswap.AssetInfoTokenRes{
				dezswap.ToAssetInfoTokenRes(pair.Asset0.Address),
				dezswap.ToAssetInfoTokenRes(pair.Asset1.Address),
			},
			LiquidityToken: pair.Lp.Address,
			AssetDecimals:  []uint{uint(pair.Asset0.Decimals), uint(pair.Asset1.Decimals)},
		},
	}
	return res
}

func (m *pairMapper) pairsToRes(pairs []service.Pair) PairsRes {
	res := make([]PairRes, len(pairs))
	for i, pair := range pairs {
		res[i] = m.pairToRes(pair)
	}
	return PairsRes{Pairs: res}
}

func (m *tokenMapper) tokenToRes(token service.Token) TokenRes {
	res := TokenRes{
		ChainId:  token.ChainId,
		Token:    token.Address,
		Name:     token.Name,
		Symbol:   token.Symbol,
		Decimals: token.Decimals,
		Icon:     token.Icon,
		Protocol: token.Protocol,
		Verified: token.Verified,
	}
	return res
}

func (m *tokenMapper) tokensToRes(tokens []service.Token) []TokenRes {
	res := make([]TokenRes, len(tokens))
	for i, token := range tokens {
		res[i] = m.tokenToRes(token)
	}
	return res
}

func (m *tickerMapper) tickerToRes(ticker service.Ticker) (TickerRes, error) {
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

func (m *tickerMapper) tickersToRes(tickers []service.Ticker) (TickersRes, error) {
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

package coingecko

import (
	"context"
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/dezswap/dezswap-api/api/service"
	"github.com/pkg/errors"
	"gorm.io/gorm"
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

type tickerService struct {
	chainId string
	*gorm.DB
	cachedPrices [][priceInfoLength]float64
}

func NewTickerService(chainId string, db *gorm.DB) service.Getter[Ticker] {
	return &tickerService{chainId, db, [][priceInfoLength]float64{}}
}

// Get implements Getter
func (s *tickerService) Get(key string) (*Ticker, error) {
	tickers := []Ticker{}

	tokens := strings.Split(key, "_")
	if len(tokens) < 2 {
		return nil, errors.New("unable to parse ticker: " + key)
	}

	if tx := s.Table("pair_stats_in_24h ps").Joins(
		"join pair p on ps.pair_id = p.id "+
			"join tokens t0 on p.chain_id = t0.chain_id and p.asset0 = t0.address "+
			"join tokens t1 on p.chain_id = t1.chain_id and p.asset1 = t1.address",
	).Where("ps.chain_id = ? and p.asset0 = ? and p.asset1 = ?", s.chainId, tokens[0], tokens[1]).Order("ps.timestamp asc").Select(
		"p.asset0 base_currency," +
			"p.asset1 target_currency," +
			"ps.volume0 base_volume," +
			"ps.volume1 target_volume," +
			"t0.decimals base_decimals," +
			"t1.decimals target_decimals," +
			"ps.liquidity0_in_price base_liquidity_in_price," +
			"p.contract pool_id," +
			"ps.timestamp",
	).Scan(&tickers); tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "TickerService.Get")
	}

	totalBaseVolume := types.ZeroDec()
	totalTargetVolume := types.ZeroDec()
	lastPrice := "0"
	for _, t := range tickers {
		baseVolume, err := types.NewDecFromStr(t.BaseVolume)
		if err != nil {
			return nil, errors.Wrap(err, "TickerService.Get")
		}
		targetVolume, err := types.NewDecFromStr(t.TargetVolume)
		if err != nil {
			return nil, errors.Wrap(err, "TickerService.Get")
		}

		totalBaseVolume = totalBaseVolume.Add(baseVolume)
		totalTargetVolume = totalTargetVolume.Add(targetVolume)
		if t.TargetDecimals > t.BaseDecimals {
			decimalDiff := types.NewDec(10).Power(uint64(t.TargetDecimals - t.BaseDecimals))
			lastPrice = baseVolume.Quo(targetVolume).Mul(decimalDiff).Abs().String()
		} else {
			decimalDiff := types.NewDec(10).Power(uint64(t.BaseDecimals - t.TargetDecimals))
			lastPrice = baseVolume.Quo(targetVolume).Quo(decimalDiff).Abs().String()
		}
	}

	ticker := &Ticker{}
	if len(tickers) > 0 {
		ticker = &tickers[len(tickers)-1]
		ticker.BaseVolume = totalBaseVolume.String()
		ticker.TargetVolume = totalTargetVolume.String()
		ticker.LastPrice = lastPrice
	} else {
		err := s.liquidity(tokens[0], tokens[1], ticker)
		if err != nil {
			return nil, errors.Wrap(err, "TickerService.Get")
		}
	}

	baseLiquidityInUsd, err := s.liquidityInUsd(*ticker)
	if err != nil {
		return nil, errors.Wrap(err, "TickerService.GetAll")
	}
	ticker.BaseLiquidityInPrice = baseLiquidityInUsd

	return ticker, nil
}

func (s *tickerService) liquidity(base string, target string, ticker *Ticker) error {
	if tx := s.Table("pair_stats_30m ps").Joins(
		"join (select pair_id, max(timestamp) latest_timestamp from pair_stats_30m group by pair_id) t on ps.pair_id = t.pair_id and ps.timestamp = t.latest_timestamp "+
			"join pair p on ps.pair_id = p.id "+
			"join tokens t0 on p.chain_id = t0.chain_id and p.asset0 = t0.address "+
			"join tokens t1 on p.chain_id = t1.chain_id and p.asset1 = t1.address",
	).Where("p.chain_id = ? and p.asset0 = ? and p.asset1 = ?", s.chainId, base, target).Select(
		"p.asset0 base_currency," +
			"p.asset1 target_currency," +
			"'0' base_volume," +
			"'0' target_volume," +
			"ps.last_swap_price last_price," +
			"t0.decimals base_decimals," +
			"t1.decimals target_decimals," +
			"liquidity0_in_price base_liquidity_in_price," +
			"p.contract pool_id, " +
			"extract(epoch from now()) * 1000 timestamp",
	).Find(&ticker); tx.Error != nil {
		return errors.Wrap(tx.Error, "TickerService.updateLiquidities")
	}

	return nil
}

// GetAll implements Getter
func (s *tickerService) GetAll() ([]Ticker, error) {
	tickers := []Ticker{}

	if tx := s.Table("pair_stats_in_24h ps").Joins(
		"join pair p on ps.pair_id = p.id "+
			"join tokens t0 on p.chain_id = t0.chain_id and p.asset0 = t0.address "+
			"join tokens t1 on p.chain_id = t1.chain_id and p.asset1 = t1.address",
	).Where("ps.chain_id = ?", s.chainId).Order("ps.timestamp asc").Select(
		"p.asset0 base_currency," +
			"p.asset1 target_currency," +
			"ps.volume0 base_volume," +
			"ps.volume1 target_volume," +
			"t0.decimals base_decimals," +
			"t1.decimals target_decimals," +
			"ps.liquidity0_in_price base_liquidity_in_price," +
			"p.contract pool_id, " +
			"ps.timestamp",
	).Scan(&tickers); tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "TickerService.GetAll")
	}

	type tickerWithDec struct {
		Ticker
		BaseVolume   types.Dec
		TargetVolume types.Dec
	}
	tickerMap := make(map[string]tickerWithDec)
	for _, t := range tickers {
		baseVolume, err := types.NewDecFromStr(t.BaseVolume)
		if err != nil {
			return nil, err
		}
		targetVolume, err := types.NewDecFromStr(t.TargetVolume)
		if err != nil {
			return nil, err
		}

		var lastPrice types.Dec
		if t.TargetDecimals > t.BaseDecimals {
			decimalDiff := types.NewDec(10).Power(uint64(t.TargetDecimals - t.BaseDecimals))
			lastPrice = baseVolume.Quo(targetVolume).Mul(decimalDiff).Abs()
		} else {
			decimalDiff := types.NewDec(10).Power(uint64(t.BaseDecimals - t.TargetDecimals))
			lastPrice = baseVolume.Quo(targetVolume).Quo(decimalDiff).Abs()
		}

		t.LastPrice = lastPrice.String()
		if lt, ok := tickerMap[t.PoolId]; ok {
			tickerMap[t.PoolId] = tickerWithDec{
				Ticker:       t,
				BaseVolume:   lt.BaseVolume.Add(baseVolume),
				TargetVolume: lt.TargetVolume.Add(targetVolume),
			}
		} else {
			tickerMap[t.PoolId] = tickerWithDec{
				Ticker:       t,
				BaseVolume:   baseVolume,
				TargetVolume: targetVolume,
			}
		}
	}

	tickers = make([]Ticker, 0, len(tickerMap))
	poolIds := make([]string, 0, len(tickerMap))
	for _, v := range tickerMap {
		t := v.Ticker
		t.BaseVolume = v.BaseVolume.String()
		t.TargetVolume = v.TargetVolume.String()
		tickers = append(tickers, t)
		poolIds = append(poolIds, t.PoolId)
	}

	inactiveTickers, err := s.inactivePools(poolIds)
	if err != nil {
		return nil, err
	}
	tickers = append(tickers, inactiveTickers...)

	for i, t := range tickers {
		baseLiquidityInUsd, err := s.liquidityInUsd(t)
		if err != nil {
			return nil, errors.Wrap(err, "TickerService.GetAll")
		}
		tickers[i].BaseLiquidityInPrice = baseLiquidityInUsd
	}

	return tickers, nil
}

func (s *tickerService) inactivePools(activePoolIds []string) ([]Ticker, error) {
	tickers := []Ticker{}

	tx := s.Table("pair_stats_30m ps").Joins(
		"join (select pair_id, max(timestamp) latest_timestamp from pair_stats_30m group by pair_id) t on ps.pair_id = t.pair_id and ps.timestamp = t.latest_timestamp " +
			"join pair p on ps.pair_id = p.id " +
			"join tokens t0 on p.chain_id = t0.chain_id and p.asset0 = t0.address " +
			"join tokens t1 on p.chain_id = t1.chain_id and p.asset1 = t1.address",
	).Select(
		"p.asset0 base_currency," +
			"p.asset1 target_currency," +
			"'0' base_volume," +
			"'0' target_volume," +
			"ps.last_swap_price last_price," +
			"t0.decimals base_decimals," +
			"t1.decimals target_decimals," +
			"liquidity0_in_price base_liquidity_in_price," +
			"p.contract pool_id, " +
			"extract(epoch from now()) * 1000 timestamp",
	)
	if len(activePoolIds) > 0 {
		tx = tx.Where("p.chain_id = ? and p.contract not in ?", s.chainId, activePoolIds)
	} else {
		tx = tx.Where("p.chain_id = ?", s.chainId)
	}

	if tx := tx.Find(&tickers); tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "TickerService.inactivePools")
	}

	return tickers, nil
}

func (s *tickerService) liquidityInUsd(ticker Ticker) (string, error) {
	baseLiquidityInPrice, err := strconv.ParseFloat(ticker.BaseLiquidityInPrice, 64)
	if err != nil {
		return "", err
	}
	priceTokenInUsd := s.price(ticker.Timestamp*1000, false)
	if priceTokenInUsd == 0 {
		err = s.cachePriceInUsd(priceTokenId)
		if err != nil {
			return "", err
		}
		priceTokenInUsd = s.price(ticker.Timestamp*1000, true)
	}

	return strconv.FormatFloat(baseLiquidityInPrice*priceTokenInUsd, 'f', -1, 64), nil
}

// TODO: fixed endpoint and arguments
func (s *tickerService) cachePriceInUsd(priceCoinId string) error {
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

	s.cachedPrices = decoded.Prices

	return nil
}

func (s *tickerService) price(targetTimestamp float64, force bool) float64 {
	price := float64(0)
	for _, p := range s.cachedPrices {
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

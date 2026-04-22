package coingecko

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	cmath "cosmossdk.io/math"
	"github.com/dezswap/dezswap-api/api/v1/service"
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"golang.org/x/sync/singleflight"
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
	mu           sync.RWMutex
	cachedPrices [][priceInfoLength]float64
	sfGroup      singleflight.Group
	httpClient   *http.Client
	endpoint     string
}

func NewTickerService(chainId string, db *gorm.DB) service.Getter[Ticker] {
	return &tickerService{chainId: chainId, DB: db, endpoint: coinGeckoEndpoint}
}

// Get implements Getter
func (s *tickerService) Get(key string) (*Ticker, error) {
	tokens := strings.Split(key, "_")
	if len(tokens) < 2 {
		return nil, errors.New("unable to parse ticker: " + key)
	}

	tickers, err := s.tickers(" and p.asset0 = ? and p.asset1 = ?", tokens...)
	if err != nil {
		return nil, errors.Wrap(err, "tickerService.Get")
	}

	ticker := &Ticker{}
	if len(tickers) > 0 {
		ticker = &tickers[len(tickers)-1]
		if ticker.LastPrice == "" {
			price, err := s.lastSwapPriceFromInactive(ticker.PoolId)
			if err != nil {
				return nil, errors.Wrap(err, "tickerService.Get")
			}
			ticker.LastPrice = price
		}
	} else {
		err := s.liquidity(tokens[0], tokens[1], ticker)
		if err != nil {
			return nil, errors.Wrap(err, "tickerService.Get")
		}
	}

	if p := s.price(ticker.Timestamp, false); p == 0 {
		if _, err, _ = s.sfGroup.Do(priceTokenId, func() (any, error) {
			return nil, s.cachePriceInUsd(priceTokenId)
		}); err != nil {
			return nil, err
		}
	}

	baseLiquidityInUsd, err := s.liquidityInUsd(*ticker)
	if err != nil {
		return nil, errors.Wrap(err, "tickerService.Get")
	}
	ticker.BaseLiquidityInPrice = baseLiquidityInUsd

	return ticker, nil
}

// lastSwapPriceFromInactive returns the most recent swap price
// for a pool that has no activity in the recent window.
func (s *tickerService) lastSwapPriceFromInactive(poolId string) (string, error) {
	inactiveTickers, err := s.inactivePool(poolId)
	if err != nil {
		return "", err
	}
	if len(inactiveTickers) == 0 {
		return "", errors.New("no ticker has returned")
	}

	return inactiveTickers[0].LastPrice, nil
}

func (s *tickerService) liquidity(base string, target string, ticker *Ticker) error {
	query := `
select p.asset0 base_currency,
       p.asset1 target_currency,
       '0' base_volume,
       '0' target_volume,
       ps.last_swap_price last_price,
       t0.decimals base_decimals,
       t1.decimals target_decimals,
       liquidity0_in_price base_liquidity_in_price,
       p.contract pool_id,
       extract(epoch from now()) * 1000 as timestamp
from pair_stats_30m ps
	join (select pair_id, max(timestamp) latest_timestamp
	      from pair_stats_30m
	      group by pair_id) t on ps.pair_id = t.pair_id and ps.timestamp = t.latest_timestamp
	join pair p on ps.pair_id = p.id
	join tokens t0 on p.chain_id = t0.chain_id and p.asset0 = t0.address
	join tokens t1 on p.chain_id = t1.chain_id and p.asset1 = t1.address
where p.chain_id = ? and p.asset0 = ? and p.asset1 = ?
`
	if tx := s.Raw(query, s.chainId, base, target).Find(&ticker); tx.Error != nil {
		return errors.Wrap(tx.Error, "TickerService.liquidity")
	}

	return nil
}

// GetAll implements Getter
func (s *tickerService) GetAll() ([]Ticker, error) {
	tickers, err := s.tickers("")
	if err != nil {
		return nil, errors.Wrap(err, "tickerService.GetAll")
	}

	var activePoolIds []string
	zeroPricePoolIdIdxMap := make(map[string]int)

	var latestTs float64
	if len(tickers) > 0 {
		latestTs = tickers[len(tickers)-1].Timestamp
	}
	for i, t := range tickers {
		if len(t.LastPrice) > 0 {
			activePoolIds = append(activePoolIds, t.PoolId)
		} else {
			zeroPricePoolIdIdxMap[t.PoolId] = i
		}
	}

	inactiveTickers, err := s.inactivePools(activePoolIds)
	if err != nil {
		return nil, errors.Wrap(err, "tickerService.GetAll")
	}
	for _, t := range inactiveTickers {
		if i, ok := zeroPricePoolIdIdxMap[t.PoolId]; ok {
			tickers[i].LastPrice = t.LastPrice
		} else {
			tickers = append(tickers, t)
		}
	}

	if latestTs == 0 {
		for _, t := range inactiveTickers {
			if latestTs < t.Timestamp {
				latestTs = t.Timestamp
			}
		}
	}

	if p := s.price(latestTs, false); p == 0 {
		if _, err, _ = s.sfGroup.Do(priceTokenId, func() (any, error) {
			return nil, s.cachePriceInUsd(priceTokenId)
		}); err != nil {
			return nil, err
		}
	}

	for i, t := range tickers {
		baseLiquidityInUsd, err := s.liquidityInUsd(t)
		if err != nil {
			return nil, errors.Wrap(err, "tickerService.GetAll")
		}
		tickers[i].BaseLiquidityInPrice = baseLiquidityInUsd
	}

	return tickers, nil
}

func (s *tickerService) tickers(cond string, bindings ...string) ([]Ticker, error) {
	query := `
select distinct
    p.asset0 base_currency,
    p.asset1 target_currency,
    sum(ps.volume0) over (partition by ps.pair_id) base_volume,
    sum(ps.volume1) over (partition by ps.pair_id) target_volume,
    t0.decimals base_decimals,
    t1.decimals target_decimals,
    first_value(ps.liquidity0_in_price) over (partition by ps.pair_id order by ps.timestamp desc) base_liquidity_in_price,
    p.contract pool_id,
    first_value(ps.timestamp) over (partition by ps.pair_id order by ps.timestamp desc) as timestamp
from pair_stats_recent ps
    join pair p on ps.pair_id = p.id
    join tokens t0 on p.chain_id = t0.chain_id and p.asset0 = t0.address
    join tokens t1 on p.chain_id = t1.chain_id and p.asset1 = t1.address
where ps.chain_id = ?
  and ps.timestamp >= extract(epoch from now()-interval'24h')
`
	var tickers []Ticker
	if cond != "" {
		b := make([]interface{}, len(bindings)+1)
		b[0] = s.chainId
		for i, v := range bindings {
			b[i+1] = v
		}
		if tx := s.Raw(query+cond, b...).Find(&tickers); tx.Error != nil {
			return nil, errors.Wrap(tx.Error, "tickerService.tickers")
		}
	} else {
		if tx := s.Raw(query, s.chainId).Find(&tickers); tx.Error != nil {
			return nil, errors.Wrap(tx.Error, "tickerService.tickers")
		}
	}

	for i, t := range tickers {
		var baseVolume cmath.LegacyDec
		if v, err := pkg.NewDecFromStrWithTruncate(t.BaseVolume); err != nil {
			return nil, errors.Wrap(err, "tickerService.tickers")
		} else {
			baseVolume = cmath.LegacyNewDecFromIntWithPrec(v.TruncateInt(), int64(t.BaseDecimals))
		}

		var targetVolume cmath.LegacyDec
		if v, err := pkg.NewDecFromStrWithTruncate(t.TargetVolume); err != nil {
			return nil, errors.Wrap(err, "tickerService.tickers")
		} else {
			targetVolume = cmath.LegacyNewDecFromIntWithPrec(v.TruncateInt(), int64(t.TargetDecimals))
		}

		tickers[i].BaseVolume = baseVolume.String()
		tickers[i].TargetVolume = targetVolume.String()

		targetDecimal := cmath.LegacyNewDec(10).Power(uint64(t.TargetDecimals))
		if !baseVolume.IsZero() {
			tickers[i].LastPrice = cmath.LegacyNewDecFromIntWithPrec(targetVolume.Quo(baseVolume).Mul(targetDecimal).RoundInt(), int64(t.TargetDecimals)).String()
		}
	}

	return tickers, nil
}

// inactivePool returns ticker data for a single inactive pool by contract address.
func (s *tickerService) inactivePool(contractId string) ([]Ticker, error) {
	return s.queryInactivePools(" and p.contract = ?", contractId)
}

// inactivePools returns ticker data for all inactive pools, excluding the given pool IDs.
func (s *tickerService) inactivePools(excludePoolIds []string) ([]Ticker, error) {
	if len(excludePoolIds) > 0 {
		return s.queryInactivePools(" and p.contract not in ?", excludePoolIds)
	}
	return s.queryInactivePools("", nil)
}

// queryInactivePools runs the pair_stats_30m base query with an optional extra condition and arg.
func (s *tickerService) queryInactivePools(cond string, arg any) ([]Ticker, error) {
	const query = `
select p.asset0 base_currency,
       p.asset1 target_currency,
       '0' base_volume,
       '0' target_volume,
       ps.last_swap_price last_price,
       t0.decimals base_decimals,
       t1.decimals target_decimals,
       liquidity0_in_price base_liquidity_in_price,
       p.contract pool_id,
       extract(epoch from now()) * 1000 as timestamp
from pair_stats_30m ps
	join (select pair_id, max(timestamp) latest_timestamp
		from pair_stats_30m
		group by pair_id) t on ps.pair_id = t.pair_id and ps.timestamp = t.latest_timestamp
	join pair p on ps.pair_id = p.id
	join tokens t0 on p.chain_id = t0.chain_id and p.asset0 = t0.address
	join tokens t1 on p.chain_id = t1.chain_id and p.asset1 = t1.address
where p.chain_id = ?
`
	var tickers []Ticker
	var tx *gorm.DB
	if cond != "" {
		tx = s.Raw(query+cond, s.chainId, arg).Find(&tickers)
	} else {
		tx = s.Raw(query, s.chainId).Find(&tickers)
	}
	if tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "TickerService.queryInactivePools")
	}

	return tickers, nil
}

func (s *tickerService) liquidityInUsd(ticker Ticker) (string, error) {
	baseLiquidityInPrice, err := strconv.ParseFloat(ticker.BaseLiquidityInPrice, 64)
	if err != nil {
		return "", err
	}
	priceTokenInUsd := s.price(ticker.Timestamp, true)

	return strconv.FormatFloat(baseLiquidityInPrice*priceTokenInUsd, 'f', -1, 64), nil
}

// TODO: fixed endpoint and arguments
func (s *tickerService) cachePriceInUsd(priceCoinId string) error {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	endpoint, err := url.Parse(s.endpoint + priceCoinId + "/market_chart")
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

	client := s.httpClient
	if client == nil {
		client = http.DefaultClient
	}
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

	s.mu.Lock()
	s.cachedPrices = decoded.Prices
	s.mu.Unlock()

	return nil
}

func (s *tickerService) price(targetTimestamp float64, force bool) float64 {
	s.mu.RLock()
	prices := s.cachedPrices
	s.mu.RUnlock()

	price := float64(0)
	for _, p := range prices {
		if p[priceTimestamp] > math.Trunc(targetTimestamp)*1_000 {
			return price
		}
		price = p[priceValue]
	}

	if force {
		return price
	}
	return 0
}

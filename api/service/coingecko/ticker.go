package coingecko

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/dezswap/dezswap-api/api/service"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strings"
)

type tickerService struct {
	chainId string
	*gorm.DB
}

func NewTickerService(chainId string, db *gorm.DB) service.Getter[Ticker] {
	return &tickerService{chainId, db}
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
		return nil, errors.Wrap(tx.Error, "TickerService.GetAll")
	}

	totalBaseVolume := types.ZeroDec()
	totalTargetVolume := types.ZeroDec()
	lastPrice := "0"
	for _, t := range tickers {
		baseVolume, err := types.NewDecFromStr(t.BaseVolume)
		if err != nil {
			return nil, err
		}
		targetVolume, err := types.NewDecFromStr(t.TargetVolume)
		if err != nil {
			return nil, err
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
		err := s.updateLiquidity(tokens[0], tokens[1], ticker)
		if err != nil {
			return nil, err
		}
	}

	return ticker, nil
}

func (s *tickerService) updateLiquidity(base string, target string, ticker *Ticker) error {
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

		if lt, ok := tickerMap[t.PoolId]; ok {
			lt.LastPrice = lastPrice.String()
			tickerMap[t.PoolId] = tickerWithDec{
				Ticker:       lt.Ticker,
				BaseVolume:   lt.BaseVolume.Add(baseVolume),
				TargetVolume: lt.TargetVolume.Add(targetVolume),
			}
		} else {
			t.LastPrice = lastPrice.String()
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

	inactiveTickers, err := s.updateInactivePools(poolIds)
	if err != nil {
		return nil, err
	}
	tickers = append(tickers, inactiveTickers...)

	return tickers, nil
}

func (s *tickerService) updateInactivePools(activePoolIds []string) ([]Ticker, error) {
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
		return nil, errors.Wrap(tx.Error, "TickerService.updateInactivePools")
	}

	return tickers, nil
}

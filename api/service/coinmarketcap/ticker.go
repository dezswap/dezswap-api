package coinmarketcap

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

func (s tickerService) Get(key string) (*Ticker, error) {
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
		"t0.address base_address," +
			"t0.name base_name," +
			"t0.symbol base_symbol," +
			"t1.address quote_address," +
			"t1.name quote_name," +
			"t1.symbol quote_symbol," +
			"ps.volume0 base_volume," +
			"ps.volume1 quote_volume," +
			"t0.decimals base_decimals," +
			"t1.decimals quote_decimals," +
			"p.contract pool_id," +
			"ps.timestamp",
	).Scan(&tickers); tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "TickerService.GetAll")
	}

	totalBaseVolume := types.ZeroDec()
	totalQuoteVolume := types.ZeroDec()
	lastPrice := "0"
	for _, t := range tickers {
		baseVolume, err := types.NewDecFromStr(t.BaseVolume)
		if err != nil {
			return nil, err
		}
		quoteVolume, err := types.NewDecFromStr(t.QuoteVolume)
		if err != nil {
			return nil, err
		}

		totalBaseVolume = totalBaseVolume.Add(baseVolume)
		totalQuoteVolume = totalQuoteVolume.Add(quoteVolume)
		if t.QuoteDecimals > t.BaseDecimals {
			decimalDiff := types.NewDec(10).Power(uint64(t.QuoteDecimals - t.BaseDecimals))
			lastPrice = baseVolume.Quo(quoteVolume).Mul(decimalDiff).Abs().String()
		} else {
			decimalDiff := types.NewDec(10).Power(uint64(t.BaseDecimals - t.QuoteDecimals))
			lastPrice = baseVolume.Quo(quoteVolume).Quo(decimalDiff).Abs().String()
		}
	}

	var ticker *Ticker
	if len(tickers) > 0 {
		ticker = &tickers[len(tickers)-1]
		ticker.BaseVolume = totalBaseVolume.String()
		ticker.QuoteVolume = totalQuoteVolume.String()
		ticker.LastPrice = lastPrice
	} else {
		var err error
		ticker, err = s.updateLastPrice(tokens[0], tokens[1])
		if err != nil {
			return nil, err
		}
	}

	return ticker, nil
}

func (s *tickerService) updateLastPrice(base string, quote string) (*Ticker, error) {
	var ticker Ticker
	if tx := s.Table("pair_stats_30m ps").Joins(
		"join (select pair_id, max(timestamp) latest_timestamp from pair_stats_30m group by pair_id) t on ps.pair_id = t.pair_id and ps.timestamp = t.latest_timestamp ",
	).Where("p.chain_id = ? and p.asset0 = ? and p.asset1 = ?", s.chainId, base, quote).Select(
		"ps.last_swap_price last_price," +
			"extract(epoch from now()) * 1000 timestamp",
	).Find(&ticker); tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "TickerService.updateLastPrice")
	}

	return &ticker, nil
}

func (s tickerService) GetAll() ([]Ticker, error) {
	tickers := []Ticker{}

	if tx := s.Table("pair_stats_in_24h ps").Joins(
		"join pair p on ps.pair_id = p.id "+
			"join tokens t0 on p.chain_id = t0.chain_id and p.asset0 = t0.address "+
			"join tokens t1 on p.chain_id = t1.chain_id and p.asset1 = t1.address",
	).Where("ps.chain_id = ?", s.chainId).Order("ps.timestamp asc").Select(
		"t0.address base_address," +
			"t0.name base_name," +
			"t0.symbol base_symbol," +
			"t1.address quote_address," +
			"t1.name quote_name," +
			"t1.symbol quote_symbol," +
			"ps.volume0 base_volume," +
			"ps.volume1 quote_volume," +
			"t0.decimals base_decimals," +
			"t1.decimals quote_decimals," +
			"p.contract pool_id, " +
			"ps.timestamp",
	).Scan(&tickers); tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "TickerService.GetAll")
	}

	type tickerWithDec struct {
		Ticker
		BaseVolume  types.Dec
		QuoteVolume types.Dec
	}
	tickerMap := make(map[string]tickerWithDec)
	for _, t := range tickers {
		baseVolume, err := types.NewDecFromStr(t.BaseVolume)
		if err != nil {
			return nil, err
		}
		quoteVolume, err := types.NewDecFromStr(t.QuoteVolume)
		if err != nil {
			return nil, err
		}

		var lastPrice types.Dec
		if t.QuoteDecimals > t.BaseDecimals {
			decimalDiff := types.NewDec(10).Power(uint64(t.QuoteDecimals - t.BaseDecimals))
			lastPrice = baseVolume.Quo(quoteVolume).Mul(decimalDiff).Abs()
		} else {
			decimalDiff := types.NewDec(10).Power(uint64(t.BaseDecimals - t.QuoteDecimals))
			lastPrice = baseVolume.Quo(quoteVolume).Quo(decimalDiff).Abs()
		}

		if lt, ok := tickerMap[t.PoolId]; ok {
			lt.LastPrice = lastPrice.String()
			tickerMap[t.PoolId] = tickerWithDec{
				Ticker:      lt.Ticker,
				BaseVolume:  lt.BaseVolume.Add(baseVolume),
				QuoteVolume: lt.QuoteVolume.Add(quoteVolume),
			}
		} else {
			t.LastPrice = lastPrice.String()
			tickerMap[t.PoolId] = tickerWithDec{
				Ticker:      t,
				BaseVolume:  baseVolume,
				QuoteVolume: quoteVolume,
			}
		}
	}

	tickers = make([]Ticker, 0, len(tickerMap))
	poolIds := make([]string, 0, len(tickerMap))
	for _, v := range tickerMap {
		t := v.Ticker
		t.BaseVolume = v.BaseVolume.String()
		t.QuoteVolume = v.QuoteVolume.String()
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
		"t0.address base_address," +
			"t0.name base_name," +
			"t0.symbol base_symbol," +
			"t1.address quote_address," +
			"t1.name quote_name," +
			"t1.symbol quote_symbol," +
			"'0' base_volume," +
			"'0' quote_volume," +
			"ps.last_swap_price last_price," +
			"t0.decimals base_decimals," +
			"t1.decimals quote_decimals," +
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

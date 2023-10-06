package service

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/dezswap/dezswap-api/pkg/db"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type statTypeIdx int

const (
	statVolume statTypeIdx = 0 + iota
	statCommission
	statLiquidity
	countOfStatType
)

const (
	statPeriod24h  = "24h"
	statPeriod7d   = "7d"
	statPeriod1mon = "1mon"
)

type statService struct {
	chainId string
	*gorm.DB
}

var _ Getter[PairStats] = &statService{}

func NewStatService(chainId string, db *gorm.DB) Getter[PairStats] {
	return &statService{chainId, db}
}

// Get implements Getter
func (s *statService) Get(key string) (*PairStats, error) {
	pairStatMap := make(map[string][countOfStatType]types.Dec)

	switch key {
	case statPeriod24h:
		err := s.sumPairStats24h(0, pairStatMap)
		if err != nil {
			return nil, errors.Wrap(err, "statService.GetAll")
		}
		pairStats := s.mapToSlice(pairStatMap)

		return &pairStats, nil
	case statPeriod7d, statPeriod1mon:
		stats, err := s.pairStats30m(key)
		if err != nil {
			return nil, errors.Wrap(err, "statService.Get")
		}
		latestTimestamp := float64(0)
		for _, stat := range stats {
			if latestTimestamp < stat.Timestamp {
				latestTimestamp = stat.Timestamp
			}

			err := s.sumPairStat(stat, pairStatMap)
			if err != nil {
				return nil, errors.Wrap(err, "statService.Get")
			}
		}

		err = s.sumPairStats24h(latestTimestamp, pairStatMap)
		if err != nil {
			return nil, errors.Wrap(err, "statService.GetAll")
		}
		pairStats := s.mapToSlice(pairStatMap)

		return &pairStats, nil
	default:
		return nil, errors.New("unsupported period")
	}
}

// GetAll implements Getter
func (s *statService) GetAll() ([]PairStats, error) {
	pairStatsByPeriod := make([]PairStats, CountOfPeriodType)
	pairStatMap := make(map[string][countOfStatType]types.Dec)
	now := time.Now()

	stats30m, err := s.pairStats30m("1mon")
	if err != nil {
		return nil, errors.Wrap(err, "statService.GetAll")
	}
	err = s.sumPairStats24h(stats30m[len(stats30m)-1].Timestamp, pairStatMap)
	if err != nil {
		return nil, errors.Wrap(err, "statService.GetAll")
	}

	tsBefore24h := now.AddDate(0, 0, -1).UnixMicro() * 1000
	tsBefore7d := now.AddDate(0, 0, -7).UnixMicro() * 1000

	done24h := false
	done7d := false

	for _, stat := range stats30m {
		if stat.Timestamp > float64(tsBefore24h) && !done24h {
			pairStatsByPeriod[Period24h] = s.mapToSlice(pairStatMap)
			done24h = true
		} else if stat.Timestamp > float64(tsBefore7d) && !done7d {
			pairStatsByPeriod[Period7d] = s.mapToSlice(pairStatMap)
			done7d = true
		}

		err := s.sumPairStat(stat, pairStatMap)
		if err != nil {
			return nil, errors.Wrap(err, "statService.GetAll")
		}
	}
	pairStatsByPeriod[Period1mon] = s.mapToSlice(pairStatMap)

	return pairStatsByPeriod, nil
}

func (s *statService) sumPairStats24h(minTimestamp float64, sumStatMap map[string][countOfStatType]types.Dec) error {
	stats := []db.PairStat{}
	if err := s.Table("pair_stats_in_24h ps").Joins(
		"join pair p on p.id = ps.pair_id").Where(
		"ps.chain_id = ? and ps.timestamp > ?", s.chainId, minTimestamp).Select(
		"p.contract address," +
			"ps.volume0_in_price," +
			"ps.volume1_in_price," +
			"ps.commission0_in_price," +
			"ps.commission1_in_price," +
			"ps.liquidity0_in_price," +
			"ps.liquidity1_in_price," +
			"ps.timestamp",
	).Scan(&stats).Error; err != nil {
		return err
	}
	for _, stat := range stats {
		err := s.sumPairStat(stat, sumStatMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *statService) pairStats30m(period string) ([]db.PairStat, error) {
	stats := []db.PairStat{}
	if err := s.Table("pair_stats_30m ps").Joins(
		"join pair p on p.id = ps.pair_id").Where(
		"ps.chain_id = ? and ps.timestamp > extract(epoch from now()-$2::interval)", s.chainId, period).Select(
		"p.contract address," +
			"ps.volume0_in_price," +
			"ps.volume1_in_price," +
			"ps.commission0_in_price," +
			"ps.commission1_in_price," +
			"ps.liquidity0_in_price," +
			"ps.liquidity1_in_price," +
			"ps.timestamp",
	).Scan(&stats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *statService) sumPairStat(stat db.PairStat, sumStatMap map[string][countOfStatType]types.Dec) error {
	volume0, err := types.NewDecFromStr(stat.Volume0InPrice)
	if err != nil {
		return err
	}
	volume1, err := types.NewDecFromStr(stat.Volume1InPrice)
	if err != nil {
		return err
	}
	commission0, err := types.NewDecFromStr(stat.Commission0InPrice)
	if err != nil {
		return err
	}
	commission1, err := types.NewDecFromStr(stat.Commission1InPrice)
	if err != nil {
		return err
	}
	liquidity0, err := types.NewDecFromStr(stat.Liquidity0InPrice)
	if err != nil {
		return err
	}
	liquidity1, err := types.NewDecFromStr(stat.Liquidity1InPrice)
	if err != nil {
		return err
	}

	nps := [countOfStatType]types.Dec{}
	nps[statVolume] = volume0.Add(volume1)
	nps[statCommission] = commission0.Add(commission1)
	nps[statLiquidity] = liquidity0.Add(liquidity1)

	if ps, ok := sumStatMap[stat.Address]; ok {
		ps[statVolume] = ps[statVolume].Add(nps[statVolume])
		ps[statCommission] = ps[statCommission].Add(nps[statCommission])
		ps[statLiquidity] = ps[statLiquidity].Add(nps[statLiquidity])
		sumStatMap[stat.Address] = ps
	} else {
		sumStatMap[stat.Address] = nps
	}

	return nil
}

func (s *statService) mapToSlice(pairStatMap map[string][countOfStatType]types.Dec) PairStats {
	var pairStats PairStats
	for k, v := range pairStatMap {
		pairStats = append(
			pairStats,
			PairStat{
				Address:           k,
				VolumeInPrice:     v[statVolume].String(),
				CommissionInPrice: v[statCommission].String(),
				AprInPrice:        v[statCommission].Quo(v[statLiquidity]).MulInt64(100).String(),
			})
	}

	return pairStats
}
package service

import (
	"github.com/dezswap/dezswap-api/pkg"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/dezswap/dezswap-api/pkg/db"
	"github.com/pkg/errors"
	"gorm.io/gorm"
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
		minTs := time.Now().Add(-24 * time.Hour)

		err := s.sumRecentPairStatsSince(float64(minTs.Unix()), pairStatMap)
		if err != nil {
			return nil, errors.Wrap(err, "statService.Get")
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

		err = s.sumRecentPairStatsSince(latestTimestamp+1, pairStatMap)
		if err != nil {
			return nil, errors.Wrap(err, "statService.Get")
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

	var stats30m []db.PairStat
	var err error

	// retrieve stats from db
	{
		stats30m, err = s.pairStats30m("1mon")
		if err != nil {
			return nil, errors.Wrap(err, "statService.GetAll")
		}
		err = s.sumRecentPairStatsSince(stats30m[len(stats30m)-1].Timestamp+1, pairStatMap)
		if err != nil {
			return nil, errors.Wrap(err, "statService.GetAll")
		}
	}

	if len(stats30m) == 0 {
		pairStatsByPeriod[Period24h] = s.mapToSlice(pairStatMap)
		return pairStatsByPeriod, nil
	}

	// calculate and assign stat sum by period
	{
		now := time.Now()
		tsBefore24h := now.AddDate(0, 0, -1).Unix()
		tsBefore7d := now.AddDate(0, 0, -7).Unix()

		done24h := false
		done7d := false

		for _, stat := range stats30m {
			if stat.Timestamp < float64(tsBefore24h) && !done24h {
				pairStatsByPeriod[Period24h] = s.mapToSlice(pairStatMap)
				done24h = true
			}
			if stat.Timestamp < float64(tsBefore7d) && !done7d {
				pairStatsByPeriod[Period7d] = s.mapToSlice(pairStatMap)
				done7d = true
			}

			if err := s.sumPairStat(stat, pairStatMap); err != nil {
				return nil, errors.Wrap(err, "statService.GetAll")
			}
		}
		pairStatsByPeriod[Period1mon] = s.mapToSlice(pairStatMap)
	}

	return pairStatsByPeriod, nil
}

func (s *statService) sumRecentPairStatsSince(minTimestamp float64, sumStatMap map[string][countOfStatType]types.Dec) error {
	query := `
select p.contract address,
       ps.volume0_in_price,
       ps.volume1_in_price,
       ps.commission0_in_price,
       ps.commission1_in_price,
       ps.liquidity0_in_price,
       ps.liquidity1_in_price,
       ps.timestamp
from pair_stats_recent ps
     join pair p on p.id = ps.pair_id
where ps.chain_id = ?
  and ps.timestamp >= ?
`
	stats := []db.PairStat{}
	if tx := s.Raw(query, s.chainId, minTimestamp).Find(&stats); tx.Error != nil {
		return errors.Wrap(tx.Error, "statService.sumRecentPairStatsSince")
	}
	for _, stat := range stats {
		err := s.sumPairStat(stat, sumStatMap)
		if err != nil {
			return errors.Wrap(err, "statService.sumRecentPairStatsSince")
		}
	}

	return nil
}

func (s *statService) pairStats30m(period string) ([]db.PairStat, error) {
	query := `
select p.contract address,
       ps.volume0_in_price,
       ps.volume1_in_price,
       ps.commission0_in_price,
       ps.commission1_in_price,
       ps.liquidity0_in_price,
       ps.liquidity1_in_price,
       ps.timestamp
from pair_stats_30m ps
     join pair p on p.id = ps.pair_id
where ps.chain_id = ? and ps.timestamp > extract(epoch from now()-?::interval)
order by ps.timestamp desc
`

	stats := []db.PairStat{}
	if tx := s.Raw(query, s.chainId, period).Find(&stats); tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "StatService.pairStats30m")
	}

	return stats, nil
}

func (s *statService) sumPairStat(stat db.PairStat, sumStatMap map[string][countOfStatType]types.Dec) error {
	volume0, err := pkg.NewDecFromStrWithTruncate(stat.Volume0InPrice)
	if err != nil {
		return err
	}
	volume1, err := pkg.NewDecFromStrWithTruncate(stat.Volume1InPrice)
	if err != nil {
		return err
	}
	commission0, err := pkg.NewDecFromStrWithTruncate(stat.Commission0InPrice)
	if err != nil {
		return err
	}
	commission1, err := pkg.NewDecFromStrWithTruncate(stat.Commission1InPrice)
	if err != nil {
		return err
	}
	liquidity0, err := pkg.NewDecFromStrWithTruncate(stat.Liquidity0InPrice)
	if err != nil {
		return err
	}
	liquidity1, err := pkg.NewDecFromStrWithTruncate(stat.Liquidity1InPrice)
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

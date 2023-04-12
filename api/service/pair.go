package service

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type pairService struct {
	chainId string
	*gorm.DB
}

var _ Getter[Pair] = &pairService{}

func NewPairService(chainId string, db *gorm.DB) Getter[Pair] {
	return &pairService{chainId, db}
}

// Get implements Getter
func (s *pairService) Get(key string) (*Pair, error) {
	pair := &Pair{}
	// pairs := []map[string]interface{}{}
	if err := s.Table("pair as P").Joins(
		"INNER JOIN tokens AS T0 on T0.address = P.asset0 and T0.chain_id = P.chain_id",
	).Joins(
		"INNER JOIN tokens AS T1 on T1.address = P.asset1 and T1.chain_id = P.chain_id",
	).Joins(
		"INNER JOIN tokens AS LP on LP.address = P.lp and LP.chain_id = P.chain_id",
	).Where(
		"P.chain_id = ? and P.contract = ?", s.chainId, key,
	).Select(
		"P.id as id",
		"P.chain_id as chain_id",
		"P.contract as address",
		"T0.address as asset0_address, T0.decimals as asset0_decimals, T0.verified as asset0_verified",
		"T1.address as asset1_address, T1.decimals as asset1_decimals, T1.verified as asset1_verified",
		"LP.address as lp_address, LP.decimals as lp_decimals, LP.verified as lp_verified",
	).Scan(pair).Error; err != nil {
		return nil, errors.Wrap(err, "PairService.GetAll")
	}
	if pair.Address != key {
		pair = nil
	}

	return pair, nil
}

// GetAll implements Getter
func (s *pairService) GetAll() ([]Pair, error) {
	pairs := []Pair{}
	// pairs := []map[string]interface{}{}
	if err := s.Table("pair as P").Joins(
		"INNER JOIN tokens AS T0 on T0.address = P.asset0 and T0.chain_id = P.chain_id",
	).Joins(
		"INNER JOIN tokens AS T1 on T1.address = P.asset1 and T1.chain_id = P.chain_id",
	).Joins(
		"INNER JOIN tokens AS LP on LP.address = P.lp and LP.chain_id = P.chain_id",
	).Where(
		"P.chain_id = ?", s.chainId,
	).Select(
		"P.id as id",
		"P.chain_id as chain_id",
		"P.contract as address",
		"T0.address as asset0_address, T0.decimals as asset0_decimals, T0.verified as asset0_verified",
		"T1.address as asset1_address, T1.decimals as asset1_decimals, T1.verified as asset1_verified",
		"LP.address as lp_address, LP.decimals as lp_decimals, LP.verified as lp_verified",
	).Order(
		"P.id asc",
	).Scan(&pairs).Error; err != nil {
		return nil, errors.Wrap(err, "PairService.GetAll")
	}

	return pairs, nil
}

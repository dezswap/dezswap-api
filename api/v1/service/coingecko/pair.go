package coingecko

import (
	"github.com/dezswap/dezswap-api/api/v1/service"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type pairService struct {
	chainId string
	*gorm.DB
}

func NewPairService(chainId string, db *gorm.DB) service.Getter[Pair] {
	return &pairService{chainId, db}
}

// Get implements Getter
func (s *pairService) Get(key string) (*Pair, error) {
	pair := &Pair{}

	if tx := s.Table("pair").Where("chain_id = ? and contract = ?", s.chainId, key).Select(
		"concat(asset0, '_', asset1) ticker_id," +
			"asset0 base," +
			"asset1 target," +
			"contract pool_id",
	).Scan(&pair); tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "pairService.Get")
	}

	return pair, nil
}

// GetAll implements Getter
func (s *pairService) GetAll() ([]Pair, error) {
	pairs := []Pair{}

	if tx := s.Table("pair").Where("chain_id = ?", s.chainId).Select(
		"concat(asset0, '_', asset1) ticker_id," +
			"asset0 base," +
			"asset1 target," +
			"contract pool_id",
	).Scan(&pairs); tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "pairService.GetAll")
	}

	return pairs, nil
}

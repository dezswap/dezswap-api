package service

import (
	"github.com/dezswap/dezswap-api/pkg/db/parser"
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
func (s *pairService) Get(key string) (Pair, error) {
	pair := parser.Pair{}
	if err := s.Model(&parser.Pair{}).Where("chain_id = ? and contract = ?", s.chainId, key).Omit("id,created_at,updated_at,deleted_at").Find(&pair).Error; err != nil {
		return pair, errors.Wrap(err, "PairService.Get")
	}
	return pair, nil
}

// GetAll implements Getter
func (s *pairService) GetAll() ([]Pair, error) {
	pairs := []parser.Pair{}
	if err := s.Model(&parser.Pair{}).Where("chain_id = ?", s.chainId).Omit("id,created_at,updated_at,deleted_at").Find(&pairs).Error; err != nil {
		return nil, errors.Wrap(err, "PairService.GetAll")
	}

	return pairs, nil
}

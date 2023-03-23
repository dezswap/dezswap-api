package service

import (
	"github.com/dezswap/dezswap-api/pkg/db/indexer"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type poolService struct {
	chainId string
	*gorm.DB
}

var _ Getter[Pool] = &poolService{}

func NewPoolService(chainId string, db *gorm.DB) Getter[Pool] {
	return &poolService{chainId, db}
}

// Get implements Getter
func (s *poolService) Get(key string) (*Pool, error) {
	pool := &indexer.LatestPool{}
	if err := s.Model(&indexer.LatestPool{}).Where("chain_id = ? and address = ?", s.chainId, key).Omit("id,created_at,updated_at,deleted_at").Find(pool).Error; err != nil {
		return nil, errors.Wrap(err, "PoolService.Get")
	}
	if pool.Address != key {
		pool = nil
	}
	return pool, nil
}

// GetAll implements Getter
func (s *poolService) GetAll() ([]Pool, error) {
	pools := []indexer.LatestPool{}
	if err := s.Model(&indexer.LatestPool{}).Where("chain_id = ?", s.chainId).Omit("id,created_at,updated_at,deleted_at").Find(&pools).Error; err != nil {
		return nil, errors.Wrap(err, "PoolService.GetAll")
	}
	return pools, nil
}

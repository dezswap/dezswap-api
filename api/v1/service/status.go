package service

import (
	"github.com/dezswap/dezswap-api/pkg/cache"
	"gorm.io/gorm"
)

type statusService struct {
	*gorm.DB
	cache.Cache
}

func NewStatusService(db *gorm.DB, cache cache.Cache) StatusService {
	return &statusService{db, cache}
}

func (s *statusService) CheckDB() error {
	if err := s.Exec("SELECT 1").Error; err != nil {
		return err
	}
	return nil
}

func (s *statusService) CheckCache() error {
	if err := s.Ping(); err != nil {
		return err
	}
	return nil
}

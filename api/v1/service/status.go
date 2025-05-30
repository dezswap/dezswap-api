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

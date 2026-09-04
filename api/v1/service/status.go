package service

import (
	"context"
	"time"

	"github.com/dezswap/dezswap-api/pkg/cache"
	"gorm.io/gorm"
)

// checkTimeout bounds one dependency check. The health endpoint is what a readiness
// probe reads, so a dependency that has stopped answering has to be reported as
// down well inside the probe's own timeout rather than hold the response open until
// the prober gives up and learns nothing.
const checkTimeout = 2 * time.Second

type statusService struct {
	*gorm.DB
	cache.Cache
}

func NewStatusService(db *gorm.DB, cache cache.Cache) StatusService {
	return &statusService{db, cache}
}

func (s *statusService) CheckDB(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()

	if err := s.WithContext(ctx).Exec("SELECT 1").Error; err != nil {
		return err
	}
	return nil
}

func (s *statusService) CheckCache(ctx context.Context) error {
	// A cache is optional: cacheStore hands back nil when neither redis nor the
	// in-memory store is configured, and the API serves without one. Nothing to
	// reach is not the same as failing to reach it.
	if s.Cache == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()

	if err := s.Ping(ctx); err != nil {
		return err
	}
	return nil
}

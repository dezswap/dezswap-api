package service

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockedStatusService(t *testing.T) (StatusService, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	require.NoError(t, err)

	return NewStatusService(gormDB, nil), mock
}

// A readiness probe reads this endpoint on a fixed timeout. A database that has
// stopped answering has to be reported as down inside that window, not hold the
// response open until the prober gives up and learns nothing.
func TestCheckDB_StalledQueryGivesUpOnItsOwnDeadline(t *testing.T) {
	s, mock := newMockedStatusService(t)

	mock.ExpectExec(`SELECT 1`).
		WillDelayFor(10 * checkTimeout).
		WillReturnResult(sqlmock.NewResult(0, 0))

	start := time.Now()
	err := s.CheckDB(context.Background())
	elapsed := time.Since(start)

	require.Error(t, err, "a query that outlives the deadline has to be reported as a failure")
	require.Less(t, elapsed, checkTimeout+time.Second, "the check should have given up near its own deadline")
	require.NoError(t, mock.ExpectationsWereMet(), "the query has to have reached the driver")
}

// The caller's context is what a request being abandoned travels through, so it has
// to cut the check short rather than run out the full timeout on nobody's behalf.
func TestCheckDB_HonorsACancelledCaller(t *testing.T) {
	s, mock := newMockedStatusService(t)

	mock.ExpectExec(`SELECT 1`).
		WillDelayFor(10 * checkTimeout).
		WillReturnResult(sqlmock.NewResult(0, 0))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	require.Error(t, s.CheckDB(ctx))
	require.Less(t, time.Since(start), checkTimeout, "a cancelled caller should not wait out the deadline")
}

// cacheStore hands back nil when neither redis nor the in-memory store is
// configured. Reaching through it would panic, and a panicking health endpoint
// would hold the instance out of rotation for as long as the cache stayed off.
func TestCheckCache_NoCacheConfiguredIsNotAFailure(t *testing.T) {
	s, _ := newMockedStatusService(t)

	require.NotPanics(t, func() {
		require.NoError(t, s.CheckCache(context.Background()))
	})
}

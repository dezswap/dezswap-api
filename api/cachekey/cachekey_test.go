package cachekey

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	testChainId = "test-chain"
	testMemoTTL = 5 * time.Second
)

func setupVersioner(t *testing.T, memoTTL time.Duration) (*Versioner, sqlmock.Sqlmock, *[]error, func() error) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	require.NoError(t, err)

	reported := &[]error{}
	versioner := NewVersioner(context.Background(), gormDB, testChainId, memoTTL, func(err error) {
		*reported = append(*reported, err)
	})

	return versioner, mock, reported, sqlDB.Close
}

// expire sends the next lookup back to the database, standing in for the memo TTL
// running out.
func expire(v *Versioner) {
	v.versions.Delete(Tokens.name)
	v.versions.Delete(Pairs.name)
}

func expectTokenWatermark(mock sqlmock.Sqlmock, rowCount int64, maxMark string) {
	mock.ExpectQuery(`FROM tokens WHERE chain_id = \$1 AND deleted_at IS NULL`).
		WithArgs(testChainId).
		WillReturnRows(sqlmock.NewRows([]string{"row_count", "max_mark"}).AddRow(rowCount, maxMark))
}

func expectPairWatermark(mock sqlmock.Sqlmock, rowCount int64, maxMark string) {
	mock.ExpectQuery(`FROM pair WHERE chain_id = \$1`).
		WithArgs(testChainId).
		WillReturnRows(sqlmock.NewRows([]string{"row_count", "max_mark"}).AddRow(rowCount, maxMark))
}

func TestCanonicalQuery(t *testing.T) {
	// Declaring page and limit is all it should take for a paginated endpoint to be
	// keyed correctly, and a parameter left undeclared must not reach the key.
	paged := Resource{name: "paged", sources: []source{tokenSource}, params: []string{"page", "limit"}}

	for name, tc := range map[string]struct {
		resource Resource
		query    string
		want     string
	}{
		"nothing declared drops everything": {Tokens, "page=2&1756771200000=", ""},
		"declared parameters are kept":      {paged, "page=2&limit=50", "page=2&limit=50"},
		"caller order does not matter":      {paged, "limit=50&page=2", "page=2&limit=50"},
		"undeclared parameters are dropped": {paged, "page=2&1756771200000=", "page=2"},
		"absent parameters are skipped":     {paged, "limit=50", "limit=50"},
		"repeats stay with their parameter": {paged, "page=2&page=3", "page=2,3"},
		"empty query":                       {paged, "", ""},
	} {
		q, err := url.ParseQuery(tc.query)
		require.NoError(t, err)

		require.Equalf(t, tc.want, tc.resource.CanonicalQuery(q), "%s: %q", name, tc.query)
	}
}

func TestVersion_MovesWhenTableIsWritten(t *testing.T) {
	v, mock, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	expectTokenWatermark(mock, 1000, "2026-09-02 00:00:00")
	before, err := v.Version(Tokens)
	require.NoError(t, err)

	expire(v)

	expectTokenWatermark(mock, 1001, "2026-09-02 00:01:00")
	after, err := v.Version(Tokens)
	require.NoError(t, err)

	require.NotEqual(t, before, after)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_MovesWhenRowCountDropsOnly(t *testing.T) {
	v, mock, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	expectTokenWatermark(mock, 1000, "2026-09-02 00:00:00")
	before, err := v.Version(Tokens)
	require.NoError(t, err)

	expire(v)

	// A row leaving the table does not move MAX, so the count has to carry it.
	expectTokenWatermark(mock, 999, "2026-09-02 00:00:00")
	after, err := v.Version(Tokens)
	require.NoError(t, err)

	require.NotEqual(t, before, after)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_IsMemoizedWithinTTL(t *testing.T) {
	v, mock, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	// One read is queued; a second reaching the database would be unexpected.
	expectTokenWatermark(mock, 1000, "2026-09-02 00:00:00")

	for range 10 {
		_, err := v.Version(Tokens)
		require.NoError(t, err)
	}

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_SteadyReadsDoNotKeepAVersionAliveForever(t *testing.T) {
	// ttlcache pushes an entry's expiry back on every read unless that is turned
	// off. Left on, a version under steady traffic would never be re-read and the
	// endpoint would serve one response until the traffic stopped.
	const memoTTL = 40 * time.Millisecond
	v, mock, _, close := setupVersioner(t, memoTTL)
	defer close()

	// Two reads are queued. Held alive by touch-on-hit, only the first would fire.
	expectTokenWatermark(mock, 1000, "2026-09-02 00:00:00")
	expectTokenWatermark(mock, 1001, "2026-09-02 00:01:00")

	first, err := v.Version(Tokens)
	require.NoError(t, err)

	// Reading without pause until the version moves. The read that reloads is the one
	// that returns the moved version, so stopping there is what keeps a third read --
	// which has nothing queued behind it -- from reaching the database at all.
	deadline := time.Now().Add(4 * memoTTL)
	for {
		current, err := v.Version(Tokens)
		require.NoError(t, err)
		if current != first {
			break
		}

		require.True(t, time.Now().Before(deadline),
			"steady reads held the version past its TTL: it was never re-read")
		time.Sleep(memoTTL / 8)
	}

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_PairsFollowBothSources(t *testing.T) {
	v, mock, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	expectPairWatermark(mock, 40, "pair-40")
	expectTokenWatermark(mock, 1000, "2026-09-02 00:00:00")
	before, err := v.Version(Pairs)
	require.NoError(t, err)

	expire(v)

	// Pairs are untouched, but a pair response carries joined token columns.
	expectPairWatermark(mock, 40, "pair-40")
	expectTokenWatermark(mock, 1000, "2026-09-02 00:05:00")
	after, err := v.Version(Pairs)
	require.NoError(t, err)

	require.NotEqual(t, before, after)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_TokensAndPairsDoNotShareAVersion(t *testing.T) {
	v, mock, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	expectTokenWatermark(mock, 1000, "2026-09-02 00:00:00")
	tokens, err := v.Version(Tokens)
	require.NoError(t, err)

	expectPairWatermark(mock, 40, "pair-40")
	expectTokenWatermark(mock, 1000, "2026-09-02 00:00:00")
	pairs, err := v.Version(Pairs)
	require.NoError(t, err)

	require.NotEqual(t, tokens, pairs)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_ResourceWithoutSourcesIsRejected(t *testing.T) {
	v, mock, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	// A resource with nothing to watch would yield an empty version that no write
	// could ever move, and the responses keyed on it would never be invalidated.
	_, err := v.Version(Resource{name: "unregistered"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "no sources")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_FailureIsReportedOncePerInterval(t *testing.T) {
	v, mock, reported, close := setupVersioner(t, testMemoTTL)
	defer close()

	// A database that has gone away must be read -- and reported -- once per
	// interval, not once for every request arriving during the outage.
	mock.ExpectQuery(`FROM tokens`).
		WithArgs(testChainId).
		WillReturnError(errors.New("connection refused"))

	for range 50 {
		_, err := v.Version(Tokens)
		require.Error(t, err)
	}

	require.Len(t, *reported, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_StalledReadGivesUpInsteadOfHoldingTheRoute(t *testing.T) {
	v, mock, reported, close := setupVersioner(t, testMemoTTL)
	defer close()

	// The read sits on the request path with every other caller for this resource
	// queued behind it, so a database that stops answering has to fall out.
	mock.ExpectQuery(`FROM tokens`).
		WithArgs(testChainId).
		WillDelayFor(10 * readTimeout).
		WillReturnRows(sqlmock.NewRows([]string{"row_count", "max_mark"}).AddRow(1000, ""))

	start := time.Now()
	_, err := v.Version(Tokens)
	elapsed := time.Since(start)

	require.Error(t, err)
	require.Less(t, elapsed, 5*readTimeout, "the read should have given up near its own deadline")
	require.Len(t, *reported, 1)
}

// The Versioner's context is the process lifetime, so shutting down has to stop a
// read rather than leave it running out its own deadline.
func TestVersion_ShutdownCancelsAReadInFlight(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	v := NewVersioner(ctx, gormDB, testChainId, testMemoTTL, nil)

	mock.ExpectQuery(`FROM tokens`).
		WithArgs(testChainId).
		WillDelayFor(10 * readTimeout).
		WillReturnRows(sqlmock.NewRows([]string{"row_count", "max_mark"}).AddRow(1000, ""))

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err = v.Version(Tokens)
	elapsed := time.Since(start)

	require.Error(t, err)
	require.Less(t, elapsed, readTimeout, "shutdown should not wait out the read timeout")
}

func TestVersion_RecoversOnceTheWatermarkIsReadable(t *testing.T) {
	v, mock, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	mock.ExpectQuery(`FROM tokens`).
		WithArgs(testChainId).
		WillReturnError(errors.New("connection refused"))

	_, err := v.Version(Tokens)
	require.Error(t, err)

	expire(v)

	expectTokenWatermark(mock, 1000, "2026-09-02 00:00:00")
	version, err := v.Version(Tokens)

	require.NoError(t, err)
	require.NotEmpty(t, version)
	require.NoError(t, mock.ExpectationsWereMet())
}

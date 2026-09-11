package cachekey

import (
	"context"
	"errors"
	"net/url"
	"regexp"
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
	v.marks.Delete(marksKey)
}

// expectMarks queues the single read every version is composed from. Only one is
// queued, so a second query would go unmet and fail the test -- that is what holds
// the fan-out to one.
func expectMarks(mock sqlmock.Sqlmock, marks map[string]mark) {
	rows := sqlmock.NewRows([]string{"source", "row_count", "max_mark"})
	for _, table := range []string{"tokens", "pair", "pair_stats_30m"} {
		m := marks[table]
		rows.AddRow(table, m.RowCount, m.MaxMark)
	}

	mock.ExpectQuery(`(?s)FROM "tokens".*UNION ALL.*FROM "pair" .*UNION ALL.*FROM "pair_stats_30m"`).
		WithArgs(testChainId, testChainId, testChainId).
		WillReturnRows(rows)
}

// steady is a mark set in which nothing has been written.
func steady() map[string]mark {
	return map[string]mark{
		"tokens":         {RowCount: 1000, MaxMark: "2026-09-02 00:00:00"},
		"pair":           {RowCount: 40, MaxMark: "pair-40"},
		"pair_stats_30m": {MaxMark: "1756771200"},
	}
}

func TestCanonicalQuery(t *testing.T) {
	// Declaring page and limit is all it should take for a paginated endpoint to be
	// keyed correctly, and a parameter left undeclared must not reach the key.
	paged := Resource{name: "paged", sources: []source{tokenSource}, query: Query{params: []param{{name: "page"}, {name: "limit"}}}}
	defaulted := Resource{name: "defaulted", sources: []source{tokenSource}, query: Query{params: []param{{name: "duration", fallback: "all"}}}}
	folded := Resource{name: "folded", sources: []source{tokenSource}, query: Query{params: []param{{name: "duration", fold: true}, {name: "token"}}}}

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
		"empty query":                       {paged, "", ""},

		// The handler reads gin's Query, which takes the first value. Keying on the
		// repeats too would let a caller mint an entry per repeat, each holding the
		// one response built from the first value.
		"repeats keep only the first value": {paged, "page=2&page=3", "page=2"},
		"a repeat cannot split the key":     {defaulted, "duration=year&duration=zzz", "duration=year"},

		"absent parameter takes its fallback": {defaulted, "", "duration=all"},
		"empty parameter takes its fallback":  {defaulted, "duration=", "duration=all"},
		"the fallback spelled out is one key": {defaulted, "duration=all", "duration=all"},
		"a value stands in for the fallback":  {defaulted, "duration=year", "duration=year"},

		// Folding is per parameter. A case-insensitive handler needs it, or one
		// response ends up under an entry per spelling; a token does not, because its
		// case is part of the address.
		"a folded parameter is lowercased":     {folded, "duration=YEAR", "duration=year"},
		"spellings collapse to one key":        {folded, "duration=yEaR", "duration=year"},
		"an unfolded parameter keeps its case": {folded, "token=ibc/A1B2", "token=ibc%2FA1B2"},
	} {
		q, err := url.ParseQuery(tc.query)
		require.NoError(t, err)

		require.Equalf(t, tc.want, tc.resource.CanonicalQuery(q), "%s: %q", name, tc.query)
	}
}

func TestNoticesCanonicalQuery_DelimitersCannotBecomeParameters(t *testing.T) {
	injected, err := url.ParseQuery("chain=dimension%26limit=1")
	require.NoError(t, err)
	legitimate, err := url.ParseQuery("chain=dimension&limit=1")
	require.NoError(t, err)

	require.NotEqual(t, Notices.Canonical(injected), Notices.Canonical(legitimate))
	decoded, err := url.ParseQuery(Notices.Canonical(injected))
	require.NoError(t, err)
	require.Equal(t, injected, decoded)
}

// The dashboard routes share their sources, so declaring a parameter on one of
// them used to declare it on all. A route keys on what its own handler reads: one
// it ignores would split its cache on input no response was built from.
func TestCanonicalQuery_EachDashboardRouteKeysOnWhatItsHandlerReads(t *testing.T) {
	q, err := url.ParseQuery("duration=year&token=xpla1abc")
	require.NoError(t, err)

	require.Empty(t, DashboardRecent.CanonicalQuery(q))
	require.Equal(t, "duration=year", DashboardCharts.CanonicalQuery(q))
	require.Equal(t, "token=xpla1abc", DashboardPools.CanonicalQuery(q))
}

// The identifiers are interpolated into the watermark statement rather than bound,
// so they have to stay package constants.
func TestSources_IdentifiersAreNotUserReachable(t *testing.T) {
	shape := regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

	for _, s := range distinctSources(resources) {
		require.Truef(t, shape.MatchString(s.table), "table %q is not a plain identifier", s.table)
		require.Truef(t, shape.MatchString(s.mark), "mark %q is not a plain identifier", s.mark)
	}
}

func TestVersion_EveryResourceSharesOneRead(t *testing.T) {
	v, mock, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	// One read queued for every resource. Memoizing per resource instead would
	// re-read tokens once for each of them.
	expectMarks(mock, steady())

	for _, r := range resources {
		_, err := v.Version(r)
		require.NoErrorf(t, err, "%s", r)
	}

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDistinctSources_AResourceOfKnownTablesAddsNoBranch(t *testing.T) {
	// The property the registry rests on: registering a resource built from tables
	// already watched costs no additional query.
	extra := Resource{name: "extra", sources: []source{tokenSource, pairStats30mSource}}

	require.Len(t,
		distinctSources(append(append([]Resource{}, resources...), extra)),
		len(distinctSources(resources)),
	)
}

// read maps marks by table, so a table reaching the statement under two definitions
// would have one of them silently win.
func TestDistinctSources_ATableIsReadUnderOneDefinition(t *testing.T) {
	seen := map[string]bool{}
	for _, s := range distinctSources(resources) {
		require.NotContainsf(t, seen, s.table, "%s is read under two definitions", s.table)
		seen[s.table] = true
	}
}

func TestVersion_MovesWhenTableIsWritten(t *testing.T) {
	v, mock, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	expectMarks(mock, steady())
	before, err := v.Version(Tokens)
	require.NoError(t, err)

	expire(v)

	moved := steady()
	moved["tokens"] = mark{RowCount: 1001, MaxMark: "2026-09-02 00:01:00"}
	expectMarks(mock, moved)
	after, err := v.Version(Tokens)
	require.NoError(t, err)

	require.NotEqual(t, before, after)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_MovesWhenRowCountDropsOnly(t *testing.T) {
	v, mock, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	expectMarks(mock, steady())
	before, err := v.Version(Tokens)
	require.NoError(t, err)

	expire(v)

	// A row leaving the table does not move MAX, so the count has to carry it.
	dropped := steady()
	dropped["tokens"] = mark{RowCount: 999, MaxMark: "2026-09-02 00:00:00"}
	expectMarks(mock, dropped)
	after, err := v.Version(Tokens)
	require.NoError(t, err)

	require.NotEqual(t, before, after)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_IsMemoizedWithinTTL(t *testing.T) {
	v, mock, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	// One read is queued; a second reaching the database would be unexpected.
	expectMarks(mock, steady())

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
	expectMarks(mock, steady())
	moved := steady()
	moved["tokens"] = mark{RowCount: 1001, MaxMark: "2026-09-02 00:01:00"}
	expectMarks(mock, moved)

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

	expectMarks(mock, steady())
	before, err := v.Version(Pairs)
	require.NoError(t, err)

	expire(v)

	// Pairs are untouched, but a pair response carries joined token columns.
	moved := steady()
	moved["tokens"] = mark{RowCount: 1000, MaxMark: "2026-09-02 00:05:00"}
	expectMarks(mock, moved)
	after, err := v.Version(Pairs)
	require.NoError(t, err)

	require.NotEqual(t, before, after)
	require.NoError(t, mock.ExpectationsWereMet())
}

// A version comes from marks alone, so resources reading the same tables share one.
// That is what lets them share the read, and their paths already keep the keys apart.
//
// The groups below are written out by hand rather than derived from each resource's
// sources. Derived, the test could not tell an intended sharing from a source list
// that drifted into another resource's tables -- both look identical from the
// version. Written out, a resource that took the wrong tables lands in the wrong
// group and fails.
func TestVersion_ResourcesShareAVersionExactlyWhenTheyShareSources(t *testing.T) {
	v, mock, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	expectMarks(mock, steady())

	groups := [][]Resource{
		{Tokens},
		{Pairs},
		{DashboardRecent, DashboardCharts, DashboardPools},
	}

	grouped := 0
	for _, group := range groups {
		grouped += len(group)
	}
	// A resource the registry gains has to be placed above, or it goes unchecked.
	require.Equal(t, len(resources), grouped, "every registered resource belongs to a group")

	byVersion := map[string]Resource{}
	for _, group := range groups {
		version, err := v.Version(group[0])
		require.NoError(t, err)

		for _, r := range group[1:] {
			got, err := v.Version(r)
			require.NoError(t, err)
			require.Equalf(t, version, got, "%s reads its group's tables but got another version", r)
		}

		require.NotContainsf(t, byVersion, version,
			"%s shares a version with %s, which reads other tables", group[0], byVersion[version])
		byVersion[version] = group[0]
	}

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_DashboardMovesWithTheNewestWindow(t *testing.T) {
	v, mock, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	expectMarks(mock, steady())
	before, err := v.Version(DashboardCharts)
	require.NoError(t, err)
	tokensBefore, err := v.Version(Tokens)
	require.NoError(t, err)

	expire(v)

	moved := steady()
	moved["pair_stats_30m"] = mark{MaxMark: "1756773000"}
	expectMarks(mock, moved)

	after, err := v.Version(DashboardCharts)
	require.NoError(t, err)
	tokensAfter, err := v.Version(Tokens)
	require.NoError(t, err)

	require.NotEqual(t, before, after)
	// A window closing says nothing about the token list.
	require.Equal(t, tokensBefore, tokensAfter)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_DashboardFollowsItsJoinedTables(t *testing.T) {
	v, mock, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	expectMarks(mock, steady())
	before, err := v.Version(DashboardPools)
	require.NoError(t, err)

	expire(v)

	// No new window, but a pair was listed: /dashboard/pools carries a row per pair.
	moved := steady()
	moved["pair"] = mark{RowCount: 41, MaxMark: "pair-41"}
	expectMarks(mock, moved)
	after, err := v.Version(DashboardPools)
	require.NoError(t, err)

	require.NotEqual(t, before, after)
	require.NoError(t, mock.ExpectationsWereMet())
}

// register is what puts a resource in the statement, so nothing that went through
// it can fail Check. Check exists for the resource that skipped register.
func TestCheck_PassesEveryRegisteredResource(t *testing.T) {
	v, _, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	for _, r := range resources {
		require.NoErrorf(t, v.Check(r), "%s", r)
	}
}

func TestCheck_RejectsAResourceThatSkippedTheRegistry(t *testing.T) {
	v, _, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	// Declared beside the others but never registered, so the watermark statement
	// does not read its table and no version could be composed for it.
	stray := Resource{name: "stray", sources: []source{{table: "parsed_tx", mark: "id"}}}

	require.ErrorContains(t, v.Check(stray), "not a registered source")
	require.ErrorContains(t, v.Check(Resource{name: "empty"}), "no sources")
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

func TestVersion_ResourceReadingAnUnregisteredTableIsRejected(t *testing.T) {
	v, mock, _, close := setupVersioner(t, testMemoTTL)
	defer close()

	expectMarks(mock, steady())

	// Its table is not in the watermark statement, so there is no mark to key on.
	// Serving the request anyway would pin it to a version that never moves.
	_, err := v.Version(Resource{name: "stray", sources: []source{{table: "parsed_tx", mark: "id"}}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "not a registered source")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_FailureIsReportedOncePerInterval(t *testing.T) {
	v, mock, reported, close := setupVersioner(t, testMemoTTL)
	defer close()

	// A database that has gone away must be read -- and reported -- once per
	// interval, not once for every request arriving during the outage.
	mock.ExpectQuery(`FROM "tokens"`).WillReturnError(errors.New("connection refused"))

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

	// The read sits on the request path with every other caller queued behind it, so
	// a database that stops answering has to fall out.
	mock.ExpectQuery(`FROM "tokens"`).
		WillDelayFor(10 * readTimeout).
		WillReturnRows(sqlmock.NewRows([]string{"source", "row_count", "max_mark"}))

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

	mock.ExpectQuery(`FROM "tokens"`).
		WillDelayFor(10 * readTimeout).
		WillReturnRows(sqlmock.NewRows([]string{"source", "row_count", "max_mark"}))

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

	mock.ExpectQuery(`FROM "tokens"`).WillReturnError(errors.New("connection refused"))

	_, err := v.Version(Tokens)
	require.Error(t, err)

	expire(v)

	expectMarks(mock, steady())
	version, err := v.Version(Tokens)

	require.NoError(t, err)
	require.NotEmpty(t, version)
	require.NoError(t, mock.ExpectationsWereMet())
}

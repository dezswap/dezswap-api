// Package cachekey reports a version of the tables an endpoint reads, so a cached
// response can be invalidated by a write rather than by a timer.
package cachekey

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	// VersionSeparator divides a path from the version its response was built at.
	VersionSeparator = "|v="
	// readTimeout bounds a watermark read. It sits on the request path with every
	// other caller for that resource queued behind it, so a database that has stopped
	// answering has to fall out rather than hold the route.
	readTimeout = time.Second
	// Every source is read together, so one memo entry serves them all.
	marksKey = "marks"
)

// source is a table whose writes invalidate a cached response. Both names are
// package constants and never carry request input, so they are safe to interpolate
// into the watermark statement.
type source struct {
	table string
	// hiddenDigest tracks manually edited visibility flags without timestamps.
	hiddenDigest bool
	// mark is the column MAX is taken over. Inserts are caught by the row count, so
	// it only has to move on an update.
	mark string
	// softDeleted marks a table whose rows the API reads with `deleted_at IS NULL`.
	softDeleted bool
	// unbounded drops COUNT(*), which is a full scan. A delete that leaves MAX
	// standing then waits for the entry's own expiry.
	unbounded bool
}

var (
	// Hidden flags may be edited without any timestamp changing.
	hiddenSource = source{table: "token_exception", mark: "contract", hiddenDigest: true}
	// SaveTokens assigns updated_at on the insert and the conflict branch alike.
	tokenSource = source{table: "tokens", mark: "updated_at", softDeleted: true}
	// pair carries no timestamp of its own, but is append-only.
	pairSource = source{table: "pair", mark: "id"}
	// The (chain_id, timestamp) index makes MAX(timestamp) a lookup. A replay that
	// rewrites an older window in place leaves the MAX where it was, so that change
	// waits for versionedTTL.
	pairStats30mSource = source{table: "pair_stats_30m", mark: "timestamp", unbounded: true}
)

type param struct {
	name     string
	fallback string
	// fold lowercases the value, and belongs only to a parameter its handler reads
	// case-insensitively. Folded onto a handler that does not, two spellings the
	// handler tells apart would share one entry, and one of them would be served
	// the other's response.
	fold bool
}

// Query is the set of query parameters a route's responses vary on. A parameter a
// caller sends that is not declared here never reaches the cache key: a cache
// buster would otherwise split one response into an entry per request.
type Query struct {
	params []param
}

// NoParams declares a route whose response its path alone decides.
var NoParams = Query{}

// Canonical returns the part of v that belongs in a cache key, ordered as the
// declaration is rather than as a caller sent it.
func (q Query) Canonical(v url.Values) string {
	if len(q.params) == 0 {
		return ""
	}

	parts := make([]string, 0, len(q.params))
	for _, p := range q.params {
		value := p.fallback
		// Handlers read these with gin's Query, which takes the first value.
		if values := v[p.name]; len(values) > 0 && values[0] != "" {
			value = values[0]
		}
		if p.fold {
			value = strings.ToLower(value)
		}

		if value != "" {
			parts = append(parts, url.QueryEscape(p.name)+"="+url.QueryEscape(value))
		}
	}

	return strings.Join(parts, "&")
}

type Resource struct {
	name    string
	sources []source
	// query names the parameters that decide which response a request gets. One the
	// handler reads but that is missing here would serve a single response for all
	// of its values.
	query Query
}

func (r Resource) String() string { return r.name }

// CanonicalQuery returns the part of q that belongs in this resource's cache key.
func (r Resource) CanonicalQuery(q url.Values) string { return r.query.Canonical(q) }

// resources is every resource a Versioner can resolve. Pools are deliberately not
// among them: SaveLatestPools upserts every row on each pass, so a pool version
// would move on every block and buy nothing over a plain expiry.
var resources []Resource

// register is the only way into resources. A resource declared beside the others
// but never passed through it has no mark to read, and its route would quietly fall
// back to a short-lived entry. Check turns that into a startup failure.
func register(r Resource) Resource {
	resources = append(resources, r)
	return r
}

// ToDuration reads this parameter case-insensitively and answers an absent one with
// All, so the spellings of a window are one entry.
var durationQuery = Query{params: []param{{name: "duration", fallback: "all", fold: true}}}

var (
	Tokens = register(Resource{name: "tokens", sources: []source{tokenSource, hiddenSource}})
	// A pair response carries columns joined in from tokens.
	Pairs = register(Resource{name: "pairs", sources: []source{pairSource, tokenSource, hiddenSource}})

	// dashboardSources is shared by the three resources below, so they also share a
	// version and the single read behind it. They stay separate resources because
	// params differ: no two of those handlers read the same set.
	dashboardSources = []source{pairStats30mSource, pairSource, tokenSource, hiddenSource}

	DashboardRecent = register(Resource{name: "dashboard_recent", sources: dashboardSources})
	DashboardCharts = register(Resource{
		name:    "dashboard_charts",
		sources: dashboardSources,
		query:   durationQuery,
	})
	DashboardPools = register(Resource{
		name:    "dashboard_pools",
		sources: dashboardSources,
		query:   Query{params: []param{{name: "token"}}},
	})
)

// The declarations below belong to routes that expire on a timer rather than on a
// version, so they name parameters only: there is no table to watch behind them.
var (
	// Notices pages, and its page is what a caller asks for.
	Notices = Query{params: []param{
		{name: "chain"},
		{name: "startTs"},
		{name: "after"},
		{name: "limit"},
		// gin binds this with strconv.ParseBool, which reads the word forms whatever
		// their case.
		{name: "asc", fold: true},
	}}
	// DashboardTxs filters by pool or by token, and narrows to one action.
	//
	// type is not folded: the handler matches it against lowercase constants and
	// falls back to every action, so "SWAP" and "swap" are different responses.
	DashboardTxs = Query{params: []param{{name: "pool"}, {name: "token"}, {name: "type"}}}
	// DashboardTokenCharts reads the same window as the versioned charts do.
	DashboardTokenCharts = durationQuery
)

type mark struct {
	RowCount uint64
	MaxMark  string
}

type marksResult struct {
	marks map[string]mark
	err   error
}

type Versioner struct {
	db     *gorm.DB
	stmt   string
	args   []any
	tables map[string]bool
	marks  *ttlcache.Cache[string, marksResult]
}

// NewVersioner returns a Versioner that reads no more than once per memoTTL, and
// only when asked.
//
// Reads derive from ctx rather than from the request being served: they are shared
// through the loader, so one caller giving up must not cancel the read the others
// are waiting on.
func NewVersioner(ctx context.Context, db *gorm.DB, chainId string, memoTTL time.Duration, onError func(error)) *Versioner {
	sources := distinctSources(resources)
	tables := make(map[string]bool, len(sources))
	for _, s := range sources {
		tables[s.table] = true
	}

	stmt, args := watermarkStatement(db, sources, chainId)
	v := &Versioner{db: db, stmt: stmt, args: args, tables: tables}

	load := ttlcache.LoaderFunc[string, marksResult](
		func(c *ttlcache.Cache[string, marksResult], key string) *ttlcache.Item[string, marksResult] {
			readCtx, cancel := context.WithTimeout(ctx, readTimeout)
			defer cancel()

			marks, err := v.read(readCtx)
			// onError is called here rather than from Version, so an outage costs one
			// report per memo interval instead of one per request.
			if err != nil && onError != nil {
				onError(err)
			}
			return c.Set(key, marksResult{marks: marks, err: err}, ttlcache.DefaultTTL)
		},
	)

	v.marks = ttlcache.New(
		ttlcache.WithTTL[string, marksResult](memoTTL),
		// A burst arriving on an expired entry becomes one read, not one per request.
		ttlcache.WithLoader[string, marksResult](ttlcache.NewSuppressedLoader(load, nil)),
		// Left on, an entry's expiry is pushed back on every read and a version under
		// steady traffic would never be re-read at all.
		ttlcache.WithDisableTouchOnHit[string, marksResult](),
	)

	return v
}

// distinctSources is the set of tables the watermark statement reads, so a table
// several resources share costs one branch. It deduplicates on the whole source,
// not just the table name: a table registered twice under different terms then
// yields two branches, which a test catches, instead of one silently winning.
func distinctSources(rs []Resource) []source {
	seen := make(map[source]bool)
	sources := []source{}

	for _, r := range rs {
		for _, s := range r.sources {
			if !seen[s] {
				seen[s] = true
				sources = append(sources, s)
			}
		}
	}

	return sources
}

// watermarkStatement never varies, so it is assembled once and the request path
// only binds chainId.
func watermarkStatement(db *gorm.DB, sources []source, chainId string) (string, []any) {
	var b strings.Builder
	args := make([]any, 0, len(sources))

	for i, s := range sources {
		if i > 0 {
			b.WriteString("\nUNION ALL\n")
		}

		if s.hiddenDigest {
			fmt.Fprintf(&b, "SELECT '%s' AS source, COUNT(*) AS row_count, COALESCE(md5(string_agg(contract, ',' ORDER BY contract)), '') AS max_mark FROM %s WHERE chain_id = ? AND hidden", s.table, db.Statement.Quote(s.table))
			args = append(args, chainId)
			continue
		}

		count := "COUNT(*)"
		if s.unbounded {
			count = "0::bigint"
		}

		fmt.Fprintf(&b,
			"SELECT '%s' AS source, %s AS row_count, COALESCE(MAX(%s)::text, '') AS max_mark FROM %s WHERE chain_id = ?",
			s.table, count, db.Statement.Quote(s.mark), db.Statement.Quote(s.table),
		)
		if s.softDeleted {
			b.WriteString(" AND deleted_at IS NULL")
		}

		args = append(args, chainId)
	}

	return b.String(), args
}

// Check reports whether every table r reads is one the watermark statement covers.
// Routes are attached at startup, so a resource this fails is a wiring mistake worth
// stopping the boot for: on the request path it would only show as a shorter entry.
func (v *Versioner) Check(r Resource) error {
	if len(r.sources) == 0 {
		return errors.Errorf("Versioner.Check: resource %q has no sources", r.name)
	}

	for _, s := range r.sources {
		if !v.tables[s.table] {
			return errors.Errorf("Versioner.Check: %q reads %s, which is not a registered source", r.name, s.table)
		}
	}

	return nil
}

// Version returns a token that changes once any of the resource's tables is written.
func (v *Versioner) Version(r Resource) (string, error) {
	if len(r.sources) == 0 {
		return "", errors.Errorf("Versioner.Version: resource %q has no sources", r.name)
	}

	item := v.marks.Get(marksKey)
	if item == nil {
		return "", errors.New("Versioner.Version: no marks resolved")
	}
	result := item.Value()
	if result.err != nil {
		return "", result.err
	}

	// Hashed to a fixed length, so that a mark carrying the separator -- MaxMark is a
	// text cast of whatever column the source names -- cannot break the key apart.
	h := fnv.New64a()
	for _, s := range r.sources {
		m, ok := result.marks[s.table]
		if !ok {
			return "", errors.Errorf("Versioner.Version: %q reads %s, which is not a registered source", r.name, s.table)
		}

		_, _ = h.Write(fmt.Appendf(nil, "%s:%d:%s;", s.table, m.RowCount, m.MaxMark))
	}

	return strconv.FormatUint(h.Sum64(), 36), nil
}

func (v *Versioner) read(ctx context.Context) (map[string]mark, error) {
	var rows []struct {
		Source   string
		RowCount uint64
		MaxMark  string
	}

	if err := v.db.WithContext(ctx).Raw(v.stmt, v.args...).Scan(&rows).Error; err != nil {
		return nil, errors.Wrap(err, "Versioner.read")
	}

	marks := make(map[string]mark, len(rows))
	for _, r := range rows {
		marks[r.Source] = mark{RowCount: r.RowCount, MaxMark: r.MaxMark}
	}

	return marks, nil
}

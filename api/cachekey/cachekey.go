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

// VersionSeparator divides a path from the version its response was built at.
const VersionSeparator = "|v="

// readTimeout bounds a watermark read. It sits on the request path with every
// other caller for that resource queued behind it, so a database that has stopped
// answering has to fall out rather than hold the route.
const readTimeout = time.Second

// source is a table whose writes invalidate a cached response. Both names are
// package constants and never carry request input, so they are safe to interpolate
// into the watermark statement.
type source struct {
	table string
	// mark is the column MAX is taken over. Inserts are caught by the row count, so
	// it only has to move on an update.
	mark string
	// softDeleted marks a table whose rows the API reads with `deleted_at IS NULL`.
	softDeleted bool
}

var (
	// SaveTokens assigns updated_at on the insert and the conflict branch alike.
	tokenSource = source{table: "tokens", mark: "updated_at", softDeleted: true}
	// pair carries no timestamp of its own, but is append-only.
	pairSource = source{table: "pair", mark: "id"}
)

// A Resource is an endpoint family together with the tables its responses are
// built from.
type Resource struct {
	name    string
	sources []source
	// params are the query parameters that decide which response a request gets.
	// One the handler reads but that is missing here would serve a single response
	// for all of its values.
	params []string
}

func (r Resource) String() string { return r.name }

// CanonicalQuery returns the part of q that belongs in a cache key, ordered as the
// resource declares rather than as a caller sent.
func (r Resource) CanonicalQuery(q url.Values) string {
	if len(r.params) == 0 {
		return ""
	}

	parts := make([]string, 0, len(r.params))
	for _, p := range r.params {
		if values, ok := q[p]; ok {
			parts = append(parts, p+"="+strings.Join(values, ","))
		}
	}

	return strings.Join(parts, "&")
}

var (
	Tokens = Resource{name: "tokens", sources: []source{tokenSource}}
	// A pair response carries columns joined in from tokens.
	Pairs = Resource{name: "pairs", sources: []source{pairSource, tokenSource}}

	// resources is the one list a Versioner resolves from. Pools are absent because
	// SaveLatestPools upserts every row on each pass, so their version would move on
	// every block and buy nothing over a plain expiry.
	resources = []Resource{Tokens, Pairs}
)

type versionResult struct {
	version string
	err     error
}

type Versioner struct {
	db       *gorm.DB
	chainId  string
	versions *ttlcache.Cache[string, versionResult]
}

// NewVersioner returns a Versioner that reads no more than once per memoTTL per
// resource, and only when asked. onError is called from the read itself, so an
// outage costs one report per interval rather than one per request.
//
// Reads derive from ctx rather than from the request being served: they are shared
// through the loader, so one caller giving up must not cancel the read the others
// are waiting on.
func NewVersioner(ctx context.Context, db *gorm.DB, chainId string, memoTTL time.Duration, onError func(error)) *Versioner {
	v := &Versioner{db: db, chainId: chainId}

	byName := make(map[string]Resource, len(resources))
	for _, r := range resources {
		byName[r.name] = r
	}

	load := ttlcache.LoaderFunc[string, versionResult](
		func(c *ttlcache.Cache[string, versionResult], name string) *ttlcache.Item[string, versionResult] {
			readCtx, cancel := context.WithTimeout(ctx, readTimeout)
			defer cancel()

			version, err := v.read(readCtx, byName[name])
			if err != nil && onError != nil {
				onError(err)
			}
			return c.Set(name, versionResult{version: version, err: err}, ttlcache.DefaultTTL)
		},
	)

	v.versions = ttlcache.New(
		ttlcache.WithTTL[string, versionResult](memoTTL),
		// A burst arriving on an expired entry becomes one read, not one per request.
		ttlcache.WithLoader[string, versionResult](ttlcache.NewSuppressedLoader(load, nil)),
		// Left on, an entry's expiry is pushed back on every read and a version under
		// steady traffic would never be re-read at all.
		ttlcache.WithDisableTouchOnHit[string, versionResult](),
	)

	return v
}

// Version returns a token that changes once any of the resource's tables is written.
func (v *Versioner) Version(r Resource) (string, error) {
	item := v.versions.Get(r.name)
	if item == nil {
		return "", errors.Errorf("Versioner.Version: no version resolved for %q", r.name)
	}

	result := item.Value()
	return result.version, result.err
}

func (v *Versioner) read(ctx context.Context, r Resource) (string, error) {
	// Nothing to watch would yield a version no write could move.
	if len(r.sources) == 0 {
		return "", errors.Errorf("Versioner.read: resource %q has no sources", r.name)
	}

	// Hashed for a fixed length, and so that a mark the ETL controls -- pair.id is a
	// free-form string -- cannot collide with the separator. Not for opacity: the
	// version never leaves the server, and reading it back means re-running the
	// statement below.
	h := fnv.New64a()

	for _, s := range r.sources {
		var mark struct {
			RowCount uint64
			MaxMark  string
		}

		stmt := fmt.Sprintf(
			"SELECT COUNT(*) AS row_count, COALESCE(MAX(%s)::text, '') AS max_mark FROM %s WHERE chain_id = ?",
			s.mark, s.table,
		)
		if s.softDeleted {
			stmt += " AND deleted_at IS NULL"
		}

		if err := v.db.WithContext(ctx).Raw(stmt, v.chainId).Scan(&mark).Error; err != nil {
			return "", errors.Wrapf(err, "Versioner.read: %s", s.table)
		}

		// The count travels with the mark: a row leaving the table does not move MAX.
		_, _ = h.Write(fmt.Appendf(nil, "%s:%d:%s;", s.table, mark.RowCount, mark.MaxMark))
	}

	return strconv.FormatUint(h.Sum64(), 36), nil
}

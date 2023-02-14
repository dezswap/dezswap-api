package repo

import (
	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg/db/parser"
	"github.com/pkg/errors"
)

type dbMapper interface {
	parserPairToPair(p parser.Pair) (indexer.Pair, error)
	parserPairsToPairs(pairs []parser.Pair) ([]indexer.Pair, error)
}

var _ dbMapper = &dbMapperImpl{}

type dbMapperImpl struct{}

// parserPairToPair implements dbMapper
func (d *dbMapperImpl) parserPairToPair(p parser.Pair) (indexer.Pair, error) {
	return indexer.Pair{
		ID:      p.ID,
		Address: p.Contract,
		Asset0:  p.Asset0,
		Asset1:  p.Asset1,
		Lp:      p.Lp,
	}, nil
}

// parserPairsToPairs implements dbMapper
func (d *dbMapperImpl) parserPairsToPairs(sourcePairs []parser.Pair) ([]indexer.Pair, error) {
	pairs := []indexer.Pair{}
	for _, p := range sourcePairs {
		pair, err := d.parserPairToPair(p)
		if err != nil {
			return nil, errors.Wrap(err, "dbMapperImpl.parserPairsToPairs")
		}
		pairs = append(pairs, pair)
	}
	return pairs, nil
}

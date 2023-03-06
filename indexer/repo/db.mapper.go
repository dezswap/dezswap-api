package repo

import (
	"github.com/dezswap/dezswap-api/indexer"
	indexer_db "github.com/dezswap/dezswap-api/pkg/db/indexer"
	"github.com/dezswap/dezswap-api/pkg/db/parser"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type dbMapper interface {
	parserPairToPair(p parser.Pair) (indexer.Pair, error)
	parserPairsToPairs(pairs []parser.Pair) ([]indexer.Pair, error)

	parserParsedTxToParsedTx(p parser.ParsedTx) (indexer.ParsedTx, error)
	parserParsedTxsToParsedTxs(txs []parser.ParsedTx) ([]indexer.ParsedTx, error)

	parserPoolInfoToPoolInfo(p parser.PoolInfo) (indexer.PoolInfo, error)
	parserPoolInfosToPoolInfos(p []parser.PoolInfo) ([]indexer.PoolInfo, error)

	tokenModelToToken(token indexer_db.Token) (indexer.Token, error)
	tokenModelsToTokens(tokens []indexer_db.Token) ([]indexer.Token, error)

	tokenToModel(token indexer.Token) (indexer_db.Token, error)
	tokensToModels(tokens []indexer.Token) ([]indexer_db.Token, error)

	poolToPoolModel(p indexer.PoolInfo, height uint64) (indexer_db.LatestPool, error)
	poolsToPoolModels(ps []indexer.PoolInfo, height uint64) ([]indexer_db.LatestPool, error)
}

var _ dbMapper = &dbMapperImpl{}

type dbMapperImpl struct{}

// tokenToModel implements dbMapper
func (m *dbMapperImpl) tokenToModel(token indexer.Token) (indexer_db.Token, error) {
	t := indexer_db.Token{
		Model: gorm.Model{
			ID: token.ID,
		},
		Protocol: token.Protocol,
		Symbol:   token.Symbol,
		Name:     token.Name,
		Icon:     token.Icon,
		Verified: token.Verified,
		Decimals: token.Decimals,
		ChainModel: indexer_db.ChainModel{
			ChainId: token.ChainId,
			Address: token.Address,
		},
	}
	return t, nil
}

// tokensToModels implements dbMapper
func (m *dbMapperImpl) tokensToModels(tokens []indexer.Token) ([]indexer_db.Token, error) {
	tokenModels := make([]indexer_db.Token, len(tokens))
	for idx, t := range tokens {
		tokenModel, err := (*dbMapperImpl).tokenToModel(&dbMapperImpl{}, t)
		if err != nil {
			return nil, errors.Wrap(err, "tokensToModels")
		}
		tokenModels[idx] = tokenModel
	}
	return tokenModels, nil
}

// poolToPoolModel implements dbMapper
func (m *dbMapperImpl) poolToPoolModel(p indexer.PoolInfo, height uint64) (indexer_db.LatestPool, error) {
	return indexer_db.LatestPool{
		Height: height,
		ChainModel: indexer_db.ChainModel{
			ChainId: p.ChainId,
			Address: p.Address,
		},
		Asset0Amount: p.Asset0Amount,
		Asset1Amount: p.Asset1Amount,
		LpAmount:     p.LpAmount,
	}, nil
}

// poolsToPoolModels implements dbMapper
func (m *dbMapperImpl) poolsToPoolModels(ps []indexer.PoolInfo, height uint64) ([]indexer_db.LatestPool, error) {
	poolModels := make([]indexer_db.LatestPool, len(ps))
	for idx, p := range ps {
		poolModel, err := m.poolToPoolModel(p, height)
		if err != nil {
			return nil, errors.Wrap(err, "poolsToPoolModels")
		}
		poolModels[idx] = poolModel
	}
	return poolModels, nil
}

// tokenModelToToken implements dbMapper
func (*dbMapperImpl) tokenModelToToken(token indexer_db.Token) (indexer.Token, error) {
	return indexer.Token{
		ID:       token.ID,
		Address:  token.Address,
		ChainId:  token.ChainId,
		Protocol: token.Protocol,
		Symbol:   token.Symbol,
		Name:     token.Name,
		Decimals: token.Decimals,
		Icon:     token.Icon,
		Verified: token.Verified,
	}, nil
}

// tokenModelsToTokens implements dbMapper
func (*dbMapperImpl) tokenModelsToTokens(tokens []indexer_db.Token) ([]indexer.Token, error) {
	indexerTokens := make([]indexer.Token, len(tokens))
	for idx, token := range tokens {
		indexerToken, err := (*dbMapperImpl).tokenModelToToken(&dbMapperImpl{}, token)
		if err != nil {
			return nil, errors.Wrap(err, "tokenModelsToTokens")
		}
		indexerTokens[idx] = indexerToken
	}
	return indexerTokens, nil
}

// parserPoolInfoToPoolInfo implements dbMapper
func (*dbMapperImpl) parserPoolInfoToPoolInfo(p parser.PoolInfo) (indexer.PoolInfo, error) {
	return indexer.PoolInfo{
		Height:       p.Height,
		ChainId:      p.ChainId,
		Address:      p.Contract,
		Asset0Amount: p.Asset0Amount,
		Asset1Amount: p.Asset1Amount,
		LpAmount:     p.LpAmount,
	}, nil
}

// parserPoolInfosToPoolInfos implements dbMapper
func (*dbMapperImpl) parserPoolInfosToPoolInfos(p []parser.PoolInfo) ([]indexer.PoolInfo, error) {
	poolInfos := make([]indexer.PoolInfo, len(p))
	for idx, pi := range p {
		poolInfo, err := (*dbMapperImpl).parserPoolInfoToPoolInfo(&dbMapperImpl{}, pi)
		if err != nil {
			return nil, errors.Wrap(err, "parserPoolInfosToPoolInfos")
		}
		poolInfos[idx] = poolInfo
	}
	return poolInfos, nil
}

// parserParsedTxsToParsedTxs implements dbMapper
func (m *dbMapperImpl) parserParsedTxsToParsedTxs(txs []parser.ParsedTx) ([]indexer.ParsedTx, error) {
	indexerTxs := make([]indexer.ParsedTx, len(txs))
	for idx, tx := range txs {
		indexerTx, err := m.parserParsedTxToParsedTx(tx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse parsed tx")
		}
		indexerTxs[idx] = indexerTx
	}
	return indexerTxs, nil
}

// parserParsedTxToParsedTx implements dbMapper
func (*dbMapperImpl) parserParsedTxToParsedTx(p parser.ParsedTx) (indexer.ParsedTx, error) {
	return indexer.ParsedTx{
		ID:                p.ID,
		ChainId:           p.ChainId,
		Height:            p.Height,
		Timestamp:         p.Timestamp,
		Hash:              p.Hash,
		Sender:            p.Sender,
		Type:              indexer.Action(p.Type),
		Address:           p.Contract,
		Asset0:            p.Asset0,
		Asset0Amount:      p.Asset0Amount,
		Asset1:            p.Asset1,
		Asset1Amount:      p.Asset1Amount,
		Lp:                p.Lp,
		LpAmount:          p.LpAmount,
		CommissionAmount:  p.CommissionAmount,
		Commission0Amount: p.Commission0Amount,
		Commission1Amount: p.Commission1Amount,
	}, nil
}

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

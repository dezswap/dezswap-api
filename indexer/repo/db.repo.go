package repo

import (
	"fmt"

	"github.com/dezswap/dezswap-api/configs"
	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg/db"
	indexer_db "github.com/dezswap/dezswap-api/pkg/db/indexer"
	"github.com/dezswap/dezswap-api/pkg/db/parser"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var _ indexer.DbRepo = &dbRepoImpl{}

type dbRepoImpl struct {
	dbMapper
	*gorm.DB
	chainId string
}

func New(chainId string, c configs.RdbConfig) indexer.DbRepo {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		c.Host, c.Username, c.Password, c.Database, c.Port)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{},
	})
	if err != nil {
		panic(err)
	}

	return &dbRepoImpl{&dbMapperImpl{}, gormDB, chainId}
}

// Pair implements indexer.DbRepo
func (r *dbRepoImpl) Pair(addr string) (indexer.Pair, error) {
	sourcePair := parser.Pair{}
	if err := r.Where("address = ? and chain_id= ?", addr, r.chainId).Omit("CreatedAt", "UpdatedAt", "DeletedAt").First(&sourcePair).Error; err != nil {
		return indexer.Pair{}, errors.Wrap(err, "dbRepoImpl.Pair")
	}

	pair, err := r.parserPairToPair(sourcePair)
	if err != nil {
		return indexer.Pair{}, errors.Wrap(err, "dbRepoImpl.Pair")
	}

	return pair, nil
}

// Pairs implements indexer.DbRepo
func (r *dbRepoImpl) Pairs(c db.LastIdLimitCondition) ([]indexer.Pair, error) {
	if c.Limit <= 0 {
		c.Limit = -1
	}
	orderBy := "id"
	if c.DescOrder {
		orderBy = "id DESC"
	}

	condition := r.Where("id > ? and chain_id= ?", c.LastId, r.chainId).Order(orderBy).Limit(c.Limit).Omit("CreatedAt", "UpdatedAt", "DeletedAt")
	sourcePairs := []parser.Pair{}
	if err := condition.Find(&sourcePairs).Error; err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Pairs")
	}

	pairs, err := r.parserPairsToPairs(sourcePairs)
	if err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Pairs")
	}

	return pairs, nil
}

// ParsedTxs implements indexer.DbRepo
func (r *dbRepoImpl) ParsedTxs(height uint64) ([]indexer.ParsedTx, error) {
	if height == 0 {
		return nil, nil
	}
	condition := r.Where("height = ?", height).Omit("CreatedAt", "UpdatedAt", "DeletedAt")
	sourceTxs := []parser.ParsedTx{}
	if err := condition.Find(&sourceTxs).Error; err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.ParsedTxs")
	}

	txs, err := r.parserParsedTxsToParsedTxs(sourceTxs)
	if err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.ParsedTxs")
	}

	return txs, nil
}

// Pool implements indexer.DbRepo
func (r *dbRepoImpl) Pool(addr string, height uint64) (indexer.PoolInfo, error) {
	//gorm pool
	sourcePool := parser.PoolInfo{}
	if err := r.Where("address = ? and height = ?", addr, height).Omit("CreatedAt", "UpdatedAt", "DeletedAt").First(&sourcePool).Error; err != nil {
		return indexer.PoolInfo{}, errors.Wrap(err, "dbRepoImpl.Pool")
	}

	pool, err := r.parserPoolInfoToPoolInfo(sourcePool)
	if err != nil {
		return indexer.PoolInfo{}, errors.Wrap(err, "dbRepoImpl.Pool")
	}

	return pool, nil
}

// Pools implements indexer.DbRepo
func (r *dbRepoImpl) Pools(height uint64) ([]indexer.PoolInfo, error) {
	if height == 0 {
		return nil, nil
	}
	condition := r.Where("height = ?", height).Omit("CreatedAt", "UpdatedAt", "DeletedAt")
	sourcePools := []parser.PoolInfo{}
	if err := condition.Find(&sourcePools).Error; err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Pools")
	}

	pools, err := r.parserPoolInfosToPoolInfos(sourcePools)
	if err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Pools")
	}

	return pools, nil
}

// SavePools implements indexer.DbRepo
func (r *dbRepoImpl) SavePools(pools []indexer.PoolInfo, height uint64) error {
	if len(pools) == 0 {
		return nil
	}

	poolModels, err := r.poolsToPoolModels(pools, height)
	if err != nil {
		return errors.Wrap(err, "dbRepoImpl.SavePools")
	}

	if err := r.Create(poolModels).Error; err != nil {
		return errors.Wrap(err, "dbRepoImpl.SavePools")
	}

	return nil
}

// SaveTokens implements indexer.DbRepo
func (r *dbRepoImpl) SaveTokens(tokens []indexer.Token) error {
	if len(tokens) == 0 {
		return nil
	}

	models, err := r.tokensToModels(tokens)
	if err != nil {
		return errors.Wrap(err, "dbRepoImpl.SaveTokens")
	}

	if err := r.Save(models).Error; err != nil {
		return errors.Wrap(err, "dbRepoImpl.SaveTokens")
	}

	return nil
}

// SyncedHeight implements indexer.DbRepo
func (r *dbRepoImpl) SyncedHeight() (uint64, error) {
	height := parser.SyncedHeight{}
	if err := r.FirstOrCreate(height, parser.SyncedHeight{ChainId: r.chainId}).Error; err != nil {
		return 0, errors.Wrap(err, "dbRepoImpl.SyncedHeight")
	}
	return height.Height, nil
}

// Token implements indexer.DbRepo
func (r *dbRepoImpl) Token(addr string) (indexer.Token, error) {
	tokenModel := indexer_db.Token{}
	if err := r.Where("address = ?", addr).Omit("CreatedAt", "UpdatedAt", "DeletedAt").First(&tokenModel).Error; err != nil {
		return indexer.Token{}, errors.Wrap(err, "dbRepoImpl.Token")
	}

	token, err := r.tokenModelToToken(tokenModel)
	if err != nil {
		return indexer.Token{}, errors.Wrap(err, "dbRepoImpl.Token")
	}

	return token, nil
}

// Tokens implements indexer.DbRepo
func (r *dbRepoImpl) Tokens(c db.LastIdLimitCondition) ([]indexer.Token, error) {
	if c.Limit <= 0 {
		c.Limit = -1
	}
	orderBy := "id"
	if c.DescOrder {
		orderBy = "id DESC"
	}

	condition := r.Where("id > ?", c.LastId).Order(orderBy).Limit(c.Limit).Omit("CreatedAt", "UpdatedAt", "DeletedAt")
	tokenModels := []indexer_db.Token{}
	if err := condition.Find(&tokenModels).Error; err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Tokens")
	}

	tokens, err := r.tokenModelsToTokens(tokenModels)
	if err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Tokens")
	}

	return tokens, nil
}

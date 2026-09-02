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
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

var _ indexer.DbRepo = &dbRepoImpl{}

// upsertBatchSize caps the rows per INSERT so a large call stays well under the 65535
// bind parameter limit of a single postgres statement.
const upsertBatchSize = 500

type dbRepoImpl struct {
	dbMapper
	*gorm.DB
	dest    *gorm.DB
	chainId string
}

func NewDbRepo(chainId string, srcC configs.RdbConfig, destC configs.RdbConfig) (indexer.DbRepo, error) {
	openDb := func(c configs.RdbConfig) (*gorm.DB, error) {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
			c.Host, c.Username, c.Password, c.Database, c.Port)
		if c.SSLMode != "" {
			dsn = fmt.Sprintf("%s sslmode=%s", dsn, c.SSLMode)
		}

		gormDB, err := gorm.Open(postgres.New(postgres.Config{
			DSN: dsn,
		}), &gorm.Config{
			NamingStrategy: schema.NamingStrategy{},
		})
		if err != nil {
			return nil, err
		}
		return gormDB, nil
	}
	srcDb, err := openDb(srcC)
	if err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.NewDbRepo")
	}
	destDb, err := openDb(destC)
	if err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.NewDbRepo")
	}

	return &dbRepoImpl{&dbMapperImpl{}, srcDb, destDb, chainId}, nil
}

// Pair implements indexer.DbRepo
func (r *dbRepoImpl) Pair(addr string) (*indexer.Pair, error) {
	sourcePair := parser.Pair{}
	if err := r.Where("address = ? and chain_id= ?", addr, r.chainId).Omit("CreatedAt", "UpdatedAt", "DeletedAt").First(&sourcePair).Error; err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Pair")
	}

	pair, err := r.parserPairToPair(sourcePair)
	if err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Pair")
	}

	return &pair, nil
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
	condition := r.Where("height = ? and chain_id = ?", height, r.chainId).Omit("CreatedAt", "UpdatedAt", "DeletedAt")
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
func (r *dbRepoImpl) Pool(addr string, height uint64) (*indexer.PoolInfo, error) {
	//gorm pool
	sourcePool := parser.PoolInfo{}
	if err := r.Where("address = ? and height = ? and chain_id = ?", addr, height, r.chainId).Omit("CreatedAt", "UpdatedAt", "DeletedAt").First(&sourcePool).Error; err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Pool")
	}

	pool, err := r.parserPoolInfoToPoolInfo(sourcePool)
	if err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Pool")
	}

	return &pool, nil
}

// Pools implements indexer.DbRepo
func (r *dbRepoImpl) Pools(height uint64) ([]indexer.PoolInfo, error) {
	if height == 0 {
		return nil, nil
	}
	condition := r.Where("height = ? and chain_id = ?", height, r.chainId).Omit("CreatedAt", "UpdatedAt", "DeletedAt")
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

// Pools implements indexer.DbRepo
func (r *dbRepoImpl) LatestPools() ([]indexer.PoolInfo, error) {

	condition := r.Where("chain_id = ?", r.chainId).Omit("CreatedAt", "UpdatedAt", "DeletedAt")
	sourcePools := []indexer_db.LatestPool{}
	if err := condition.Find(&sourcePools).Error; err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Pools")
	}

	pools, err := r.latestPoolInfosToPoolInfos(sourcePools)
	if err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Pools")
	}

	return pools, nil
}

// SavePools implements indexer.DbRepo
func (r *dbRepoImpl) SaveLatestPools(pools []indexer.PoolInfo, height uint64) error {
	if len(pools) == 0 {
		return nil
	}

	poolModels, err := r.poolsToPoolModels(pools, height)
	if err != nil {
		return errors.Wrap(err, "dbRepoImpl.SavePools")
	}

	// poolToPoolModel never fills the embedded gorm.Model, so no id reaches the insert
	// and idx_latest_pools_chain_id_address_key is the only arbiter available.
	if err := r.dest.Clauses(clause.OnConflict{
		UpdateAll: true,
		Columns:   []clause.Column{{Name: "address"}, {Name: "chain_id"}},
	}).CreateInBatches(&poolModels, upsertBatchSize).Error; err != nil {
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

	// The id is left to the sequence so that idx_tokens_chain_id_address_key is the only
	// index an insert can collide on. Sending an existing id would also collide with
	// tokens_pkey, which is not the declared arbiter and would raise a unique violation
	// instead of taking the update branch.
	for i := range models {
		if models[i].Model != nil {
			models[i].Model.ID = 0
		}
	}

	if err := r.dest.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "chain_id"}, {Name: "address"}},
		// created_at and deleted_at are left alone; the rest is what a token carries.
		DoUpdates: clause.AssignmentColumns([]string{
			"updated_at", "protocol", "symbol", "name", "decimals", "icon", "verified",
		}),
	}).CreateInBatches(&models, upsertBatchSize).Error; err != nil {
		return errors.Wrap(err, "dbRepoImpl.SaveTokens")
	}

	return nil
}

// SyncedHeight implements indexer.DbRepo
func (r *dbRepoImpl) SyncedHeight() (uint64, error) {
	height := parser.SyncedHeight{}
	cond := parser.SyncedHeight{}
	cond.ChainId = r.chainId
	if err := r.FirstOrCreate(&height, cond).Error; err != nil {
		return 0, errors.Wrap(err, "dbRepoImpl.SyncedHeight")
	}
	return height.Height, nil
}

// Token implements indexer.DbRepo
func (r *dbRepoImpl) Token(addr string) (*indexer.Token, error) {
	tokenModel := indexer_db.Token{}
	if err := r.Where("address = ?", addr).Omit("CreatedAt", "UpdatedAt", "DeletedAt").First(&tokenModel).Error; err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Token")
	}

	token, err := r.tokenModelToToken(tokenModel)
	if err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Token")
	}

	return &token, nil
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

	condition := r.Where("id > ? and chain_id = ?", c.LastId, r.chainId).Order(orderBy).Limit(c.Limit).Omit("CreatedAt", "UpdatedAt", "DeletedAt")
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

func (r *dbRepoImpl) TokenAddresses(c db.LastIdLimitCondition) ([]string, error) {
	if c.Limit <= 0 {
		c.Limit = -1
	}
	orderBy := "id"
	if c.DescOrder {
		orderBy = "id DESC"
	}

	condition := r.Model(&indexer_db.Token{}).Where("id > ? and chain_id = ?", c.LastId, r.chainId).Select("address").Order(orderBy).Limit(c.Limit)
	tokenAddresses := []string{}
	if err := condition.Find(&tokenAddresses).Error; err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.TokenAddresses")
	}

	return tokenAddresses, nil
}

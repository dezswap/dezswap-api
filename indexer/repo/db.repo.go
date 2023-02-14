package repo

import (
	"fmt"

	"github.com/dezswap/dezswap-api/configs"
	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg/db"
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

// Pairs implements indexer.DbRepo
func (d *dbRepoImpl) Pairs(c db.LastIdLimitCondition) ([]indexer.Pair, error) {
	if c.Limit <= 0 {
		c.Limit = -1
	}
	orderBy := "id"
	if c.DescOrder {
		orderBy = "id DESC"
	}

	stateMent := d.Where("id > ? and chain_id= ?", c.LastId, d.chainId).Order(orderBy).Limit(c.Limit).Omit("CreatedAt", "UpdatedAt", "DeletedAt")
	sourcePairs := []parser.Pair{}
	if err := stateMent.Find(&sourcePairs).Error; err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Pairs")
	}

	pairs, err := d.parserPairsToPairs(sourcePairs)
	if err != nil {
		return nil, errors.Wrap(err, "dbRepoImpl.Pairs")
	}

	return pairs, nil
}

// ParsedTxs implements indexer.DbRepo
func (*dbRepoImpl) ParsedTxs() ([]indexer.ParsedTx, error) {
	panic("unimplemented")
}

// Pools implements indexer.DbRepo
func (*dbRepoImpl) Pools(height uint64) ([]indexer.PoolInfo, error) {
	panic("unimplemented")
}

// SavePools implements indexer.DbRepo
func (*dbRepoImpl) SavePools(pools []indexer.PoolInfo, height uint64) error {
	panic("unimplemented")
}

// SaveTokens implements indexer.DbRepo
func (*dbRepoImpl) SaveTokens([]indexer.Token) error {
	panic("unimplemented")
}

// SyncedHeight implements indexer.DbRepo
func (*dbRepoImpl) SyncedHeight() (uint64, error) {
	panic("unimplemented")
}

// Tokens implements indexer.DbRepo
func (*dbRepoImpl) Tokens(c db.LastIdLimitCondition) ([]indexer.Token, error) {
	panic("unimplemented")
}

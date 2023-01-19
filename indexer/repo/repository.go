package repo

import (
	"gorm.io/gorm"
)

type Repo interface {
}

type repoImpl struct {
	mapper
	db      *gorm.DB
	chainId string
}

// func New(chainId string, c configs.RdbConfig) indexer.Repo {
// 	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
// 		c.Host, c.Username, c.Password, c.Database, c.Port)

// 	gormDB, err := gorm.Open(postgres.New(postgres.Config{
// 		DSN: dsn,
// 	}), &gorm.Config{
// 		NamingStrategy: schema.NamingStrategy{},
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	return &repoImpl{
// 		mapper:  &parserMapperImpl{},
// 		db:      gormDB,
// 		chainId: chainId,
// 	}
// }

// // GetSyncedHeight implements parser.Repo
// func (r *repoImpl) GetSyncedHeight() (uint64, error) {
// 	syncedHeight := db.SyncedHeight{}
// 	tx := r.db.FirstOrCreate(&syncedHeight, db.SyncedHeight{ChainId: r.chainId})

// 	if tx.Error != nil {
// 		return 0, errors.Wrap(tx.Error, "repo.GetSyncedHeight")
// 	}
// 	return syncedHeight.Height, nil
// }

// // // GetPairs implements parser.Repo
// // func (r *repoImpl) GetPairs() (map[string]parser.Pair, error) {
// // 	pairs := []schemas.Pair{}
// // 	tx := r.db.Where("chain_id = ?", r.chainId).Find(&pairs)
// // 	if tx.Error != nil {
// // 		return nil, errors.Wrap(tx.Error, "repo.GetPairs")
// // 	}
// // 	result := make(map[string]parser.Pair)
// // 	for _, pair := range pairs {
// // 		result[pair.Contract] = r.mapper.toPairDto(pair)
// // 	}
// // 	return result, nil
// // }

// // // Insert implements parser.Repo
// // func (r *repoImpl) Insert(height uint64, txs []parser.ParsedTx, pools []parser.PoolInfo, pairs []parser.Pair) error {
// // 	parsedTxs := []schemas.ParsedTx{}
// // 	for _, tx := range txs {
// // 		parsedTxs = append(parsedTxs, r.mapper.toParsedTxModel(r.chainId, height, tx))
// // 	}
// // 	poolInfoTxs := []schemas.PoolInfo{}
// // 	for _, pool := range pools {
// // 		poolInfoTxs = append(poolInfoTxs, r.mapper.toPoolInfoModel(r.chainId, height, pool))
// // 	}
// // 	pairTxs := []schemas.Pair{}
// // 	for _, pair := range pairs {
// // 		pairTxs = append(pairTxs, r.mapper.toPairModel(r.chainId, pair))
// // 	}

// // 	return r.db.Transaction(func(tx *gorm.DB) error {
// // 		if err := tx.Model(schemas.Pair{}).CreateInBatches(pairTxs, len(pairTxs)).Error; err != nil {
// // 			return errors.Wrap(err, "repo.Insert.Pair")
// // 		}
// // 		if err := tx.Model(schemas.ParsedTx{}).CreateInBatches(parsedTxs, len(parsedTxs)).Error; err != nil {
// // 			return errors.Wrap(err, "repo.Insert.ParsedTx")
// // 		}
// // 		if err := tx.Model(schemas.PoolInfo{}).CreateInBatches(poolInfoTxs, len(poolInfoTxs)).Error; err != nil {
// // 			return errors.Wrap(err, "repo.Insert.PoolInfo")
// // 		}
// // 		if err := tx.Model(&schemas.SyncedHeight{}).Where("chain_id = ? AND height = ?", r.chainId, height-1).Update("height", height).Error; err != nil {
// // 			return errors.Wrap(err, "repo.Insert.SyncedHeight")
// // 		}
// // 		return nil
// // 	})
// // }

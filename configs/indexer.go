package configs

import (
	"github.com/spf13/viper"
)

type IndexerConfig struct {
	ChainId string
	SrcNode GrpcConfig
	SrcDb   RdbConfig
	Db      RdbConfig
}

func indexerConfig(v *viper.Viper) IndexerConfig {
	return IndexerConfig{
		ChainId: v.GetString("indexer.chain_id"),
		SrcNode: grpcConfig(v.Sub("indexer.src_node")),
		SrcDb:   rdbConfig(v.Sub("indexer.src_db")),
		Db:      rdbConfig(v.Sub("indexer.db")),
	}
}

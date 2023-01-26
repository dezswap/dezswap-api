package configs

import (
	"github.com/spf13/viper"
)

type IndexerConfig struct {
	SrcDb RdbConfig
	Db    RdbConfig
}

func indexerConfig(v *viper.Viper) IndexerConfig {
	return IndexerConfig{
		SrcDb: rdbConfig(v.Sub("indexer.src_db")),
		Db:    rdbConfig(v.Sub("indexer.db")),
	}
}

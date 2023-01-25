package configs

import (
	"github.com/spf13/viper"
)

type IndexerConfig struct {
	SrcDb     RdbConfig
	IndexerDb RdbConfig
}

func indexerConfig(v *viper.Viper) IndexerConfig {
	return IndexerConfig{
		SrcDb:     rdbConfig(v.Sub("indexer.srcDb")),
		IndexerDb: rdbConfig(v.Sub("indexer.indexerDb")),
	}
}

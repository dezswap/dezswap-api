package configs

import (
	"github.com/spf13/viper"
)

type IndexerConfig struct {
	ChainId           string
	SrcNode           GrpcConfig
	SrcEvmRpcEndpoint string
	SrcDb             RdbConfig
	Db                RdbConfig
}

func indexerConfig(v *viper.Viper) IndexerConfig {
	chainId := v.GetString("indexer.chain_id")
	envChainId := v.GetString("INDEXER_CHAIN_ID")
	if envChainId != "" {
		chainId = envChainId
	}

	nodeC := grpcConfig(v.Sub("indexer.src_node"))
	envNodeC := grpcConfigFromEnv(v, "INDEXER_SRC_NODE")
	nodeC.Override(envNodeC)

	srcEvmRpcEndpoint := v.GetString("indexer.src_evm_rpc_endpoint")
	envSrcEvmRpcEndpoint := v.GetString("INDEXER_SRC_EVM_RPC_ENDPOINT")
	if envSrcEvmRpcEndpoint != "" {
		srcEvmRpcEndpoint = envSrcEvmRpcEndpoint
	}

	srcDbC := rdbConfig(v.Sub("indexer.src_db"))
	envSrcDbC := rdbConfigFromEnv(v, "INDEXER_SRC_DB")
	srcDbC.Override(envSrcDbC)

	dbC := rdbConfig(v.Sub("indexer.db"))
	envDbC := rdbConfigFromEnv(v, "INDEXER_DB")
	dbC.Override(envDbC)

	return IndexerConfig{
		ChainId:           chainId,
		SrcNode:           nodeC,
		SrcEvmRpcEndpoint: srcEvmRpcEndpoint,
		SrcDb:             srcDbC,
		Db:                dbC,
	}
}

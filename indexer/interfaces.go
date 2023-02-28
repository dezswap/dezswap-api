package indexer

import "github.com/dezswap/dezswap-api/pkg/db"

type AssetRepo interface {
	VerifiedTokens(chainId string) ([]Token, error)
}

type NodeRepo interface {
	LatestHeightFromNode() (uint64, error)
	TokenFromNode(addr string) (*Token, error)
	PoolFromNode(addr string, height uint64) (*PoolInfo, error)
}

type DbRepo interface {
	SyncedHeight() (uint64, error)

	Pair(addr string) (*Pair, error)
	Pairs(db.LastIdLimitCondition) ([]Pair, error)

	Token(addr string) (*Token, error)
	Tokens(db.LastIdLimitCondition) ([]Token, error)

	Pool(addr string, height uint64) (*PoolInfo, error)
	Pools(height uint64) ([]PoolInfo, error)

	ParsedTxs(height uint64) ([]ParsedTx, error)

	SavePools(pools []PoolInfo, height uint64) error
	SaveTokens([]Token) error
}

type Repo interface {
	AssetRepo
	NodeRepo
	DbRepo
}

type Indexer interface {
	UpdateVerifiedTokens() error
	UpdateTokens() error
	UpdateLatestPools() error
}

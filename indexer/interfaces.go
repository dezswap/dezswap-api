package indexer

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
	Pairs() ([]Pair, error)
	Tokens() ([]Token, error)
	Pools(height uint64) ([]PoolInfo, error)
	ParsedTxs() ([]ParsedTx, error)

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

package indexer

type Repo interface {
	SyncedHeight() (uint64, error)

	Pairs() ([]Pair, error)
	Tokens() ([]Token, error)
	Pools(height uint64) ([]PoolInfo, error)

	Pair(addr string) (*Pair, error)
	Token(addr string) (*Token, error)
	Pool(addr string, height uint64) (*PoolInfo, error)

	ParsedTxs() ([]ParsedTx, error)

	SaveLatestPools(pools []PoolInfo, height uint64) error
	SaveTokens([]Token) error
}

type Indexer interface {
	UpdateTokens() error
	UpdateLatestPools() error
}

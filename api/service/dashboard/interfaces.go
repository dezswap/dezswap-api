package dashboard

type Dashboard interface {
	Recent() (Recent, error)

	Volumes() (Volumes, error)
	Volume(of Addr) (Volumes, error)

	Tvls() (Tvls, error)
	Tvl(of Addr) (Tvls, error)

	Statistic() (Statistic, error)

	Pools() (Pools, error)
	Pool(of Addr) (Pool, error)

	Tokens() (Tokens, error)
	Token(of Addr) (Token, error)

	Txs() (Txs, error)
	Tx(of Addr) (Tx, error)

	Prices() (Prices, error)
	Price(of Addr) (Prices, error)

	Aprs() (Aprs, error)
	Apr(of Addr) (Aprs, error)
}

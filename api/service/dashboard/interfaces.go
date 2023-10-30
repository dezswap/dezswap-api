package dashboard

type Dashboard interface {
	Recent() (Recent, error)

	Volumes(duration Duration) (Volumes, error)
	Volume(of Addr, duration Duration) (Volume, error)

	Tvls(duration Duration) (Tvls, error)
	Tvl(of Addr, duration Duration) (Tvl, error)

	Statistic() (Statistic, error)

	Pools() (Pools, error)
	Pool(of Addr) (Pool, error)

	Tokens() (Tokens, error)
	Token(of Addr) (Token, error)

	Txs() (Txs, error)
	Tx(of Addr) (Tx, error)

	Prices(duration Duration) (Prices, error)
	Price(of Addr, duration Duration) (Price, error)

	Aprs(duration Duration) (Aprs, error)
	Apr(of Addr, duration Duration) (Apr, error)
}

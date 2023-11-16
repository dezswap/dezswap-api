package dashboard

type Dashboard interface {
	Recent() (Recent, error)

	Volumes(ChartDuration, ...Addr) (Volumes, error)

	Tvls(...Addr) (Tvls, error)

	Statistic(...Addr) (Statistic, error)

	Pools() (Pools, error)
	Pool(addr Addr) (Pools, error)

	Tokens(...Addr) (Tokens, error)

	Txs(...Addr) (Txs, error)

	Prices(...Addr) (Prices, error)

	Aprs(...Addr) (Aprs, error)
}

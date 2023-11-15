package dashboard

type Dashboard interface {
	Recent() (Recent, error)

	Volumes(ChartDuration, ...Addr) (Volumes, error)

	Tvls(...Addr) (Tvls, error)

	Statistic(...Addr) (Statistic, error)

	Pools(...Addr) (Pools, error)
	Pool(addr Addr) (Pools, error)

	Tokens(...Addr) (Tokens, error)

	TokenVolumes(addr Addr, itv Duration) (TokenChart, error)

	TokenTvls(addr Addr, itv Duration) (TokenChart, error)

	TokenPrices(addr Addr, itv Duration) (TokenChart, error)

	Txs(...Addr) (Txs, error)

	Prices(...Addr) (Prices, error)

	Aprs(...Addr) (Aprs, error)
}

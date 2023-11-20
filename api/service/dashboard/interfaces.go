package dashboard

type Dashboard interface {
	Recent() (Recent, error)
	RecentOf(addr Addr) (Recent, error)

	Statistic(...Addr) (Statistic, error)

	Pools(tokens ...Addr) (Pools, error)
	PoolDetail(Addr) (PoolDetail, error)

	Tokens(...Addr) (Tokens, error)
	TokenVolumes(addr Addr, itv Duration) (TokenChart, error)
	TokenTvls(addr Addr, itv Duration) (TokenChart, error)
	TokenPrices(addr Addr, itv Duration) (TokenChart, error)

	Txs(...Addr) (Txs, error)

	Volumes(Duration) (Volumes, error)
	VolumesOf(Addr, Duration) (Volumes, error)

	Tvls(Duration) (Tvls, error)
	TvlsOf(Addr, Duration) (Tvls, error)

	Aprs(Duration) (Aprs, error)
	AprsOf(Addr, Duration) (Aprs, error)
}

package dashboard

type Dashboard interface {
	Recent() (Recent, error)

	Volumes(Duration) (Volumes, error)
	VolumesOf(Addr, Duration) (Volumes, error)

	Tvls(Duration) (Tvls, error)

	Statistic(...Addr) (Statistic, error)

	Pools(...Addr) (Pools, error)
	Pool(addr Addr) (Pools, error)

	Tokens(item string, ascending bool) (Tokens, error)
	Token(addr Addr) (Token, error)

	TokenVolumes(addr Addr, itv Duration) (TokenChart, error)

	TokenTvls(addr Addr, itv Duration) (TokenChart, error)

	TokenPrices(addr Addr, itv Duration) (TokenChart, error)

	Txs(...Addr) (Txs, error)

	Aprs(...Addr) (Aprs, error)
}

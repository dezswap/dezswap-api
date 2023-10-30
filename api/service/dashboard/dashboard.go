package dashboard

import "gorm.io/gorm"

type dashboard struct {
	*gorm.DB
}

var _ Dashboard = &dashboard{}

func NewDashboardService(db *gorm.DB) Dashboard {
	return &dashboard{db}
}

// Apr implements Dashboard.
func (*dashboard) Apr(of Addr, duration Duration) (Apr, error) {
	panic("unimplemented")
}

// Aprs implements Dashboard.
func (*dashboard) Aprs(duration Duration) ([]Apr, error) {
	panic("unimplemented")
}

// Pool implements Dashboard.
func (*dashboard) Pool(of Addr) (Pool, error) {
	panic("unimplemented")
}

// Pools implements Dashboard.
func (*dashboard) Pools() (Pools, error) {
	panic("unimplemented")
}

// Price implements Dashboard.
func (*dashboard) Price(of Addr, duration Duration) (Price, error) {
	panic("unimplemented")
}

// Prices implements Dashboard.
func (*dashboard) Prices(duration Duration) ([]Price, error) {
	panic("unimplemented")
}

// Recent implements Dashboard.
func (*dashboard) Recent() (Recent, error) {
	panic("unimplemented")
}

// Statistic implements Dashboard.
func (*dashboard) Statistic() (Statistic, error) {
	panic("unimplemented")
}

// Tvl implements Dashboard.
func (*dashboard) Tvl(of Addr, duration Duration) (Tvl, error) {
	panic("unimplemented")
}

// Tvls implements Dashboard.
func (*dashboard) Tvls(duration Duration) ([]Tvl, error) {
	panic("unimplemented")
}

// Token implements Dashboard.
func (*dashboard) Token(of Addr) (Token, error) {
	panic("unimplemented")
}

// Tokens implements Dashboard.
func (*dashboard) Tokens() (Tokens, error) {
	panic("unimplemented")
}

// Tx implements Dashboard.
func (*dashboard) Tx(of Addr) (Tx, error) {
	panic("unimplemented")
}

// Txs implements Dashboard.
func (*dashboard) Txs() (Txs, error) {
	panic("unimplemented")
}

// Volume implements Dashboard.
func (*dashboard) Volume(of Addr, duration Duration) (Volume, error) {
	panic("unimplemented")
}

// Volumes implements Dashboard.
func (*dashboard) Volumes(duration Duration) ([]Volume, error) {
	panic("unimplemented")
}

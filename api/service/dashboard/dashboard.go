package dashboard

import "gorm.io/gorm"

var _ Dashboard = &dashboard{}

func NewDashboardService(chainId string, db *gorm.DB) Dashboard {
	return &dashboard{chainId, db}
}

// Apr implements Dashboard.
func (d *dashboard) Apr(of Addr, duration Duration) (Apr, error) {
	panic("unimplemented")
}

// Aprs implements Dashboard.
func (d *dashboard) Aprs(duration Duration) ([]Apr, error) {
	panic("unimplemented")
}

// Pool implements Dashboard.
func (d *dashboard) Pool(of Addr) (Pool, error) {
	panic("unimplemented")
}

// Pools implements Dashboard.
func (d *dashboard) Pools() (Pools, error) {
	panic("unimplemented")
}

// Price implements Dashboard.
func (d *dashboard) Price(of Addr, duration Duration) (Price, error) {
	panic("unimplemented")
}

// Prices implements Dashboard.
func (d *dashboard) Prices(duration Duration) ([]Price, error) {
	panic("unimplemented")
}

// Recent implements Dashboard.
func (d *dashboard) Recent() (Recent, error) {
	panic("unimplemented")
}

// Statistic implements Dashboard.
func (d *dashboard) Statistic() (Statistic, error) {
	panic("unimplemented")
}

// Tvl implements Dashboard.
func (d *dashboard) Tvl(of Addr, duration Duration) (Tvl, error) {
	panic("unimplemented")
}

// Tvls implements Dashboard.
func (d *dashboard) Tvls(duration Duration) ([]Tvl, error) {
	panic("unimplemented")
}

// Token implements Dashboard.
func (d *dashboard) Token(of Addr) (Token, error) {
	panic("unimplemented")
}

// Tokens implements Dashboard.
func (d *dashboard) Tokens() (Tokens, error) {
	panic("unimplemented")
}

// Tx implements Dashboard.
func (d *dashboard) Tx(of Addr) (Tx, error) {
	panic("unimplemented")
}

// Txs implements Dashboard.
func (d *dashboard) Txs() (Txs, error) {
	panic("unimplemented")
}

// Volume implements Dashboard.
func (d *dashboard) Volume(of Addr, duration Duration) (Volume, error) {
	panic("unimplemented")
}

// Volumes implements Dashboard.
func (d *dashboard) Volumes(duration Duration) ([]Volume, error) {
	panic("unimplemented")
}

type dashboard struct {
	chainId string
	*gorm.DB
}

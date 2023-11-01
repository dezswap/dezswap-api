package dashboard

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var _ Dashboard = &dashboard{}

func NewDashboardService(chainId string, db *gorm.DB) Dashboard {
	return &dashboard{chainId, db}
}

// Volume implements Dashboard.
func (d *dashboard) Volume(of Addr) (Volumes, error) {
	query := `
SELECT
	DATE_TRUNC('day', TO_TIMESTAMP(PS. "timestamp")) AS active_date,
	SUM(PS.volume0_in_price) AS volume
FROM
	pair_stats_30m AS PS
	LEFT JOIN pair AS P ON PS.pair_id = P.id
WHERE
	P.contract = ?
	AND
	DATE_TRUNC('day', TO_TIMESTAMP(PS. "timestamp")) >= DATE_TRUNC('day', NOW() - INTERVAL '1 year')
GROUP BY
	DATE_TRUNC('day', TO_TIMESTAMP(PS. "timestamp"))
ORDER BY
	DATE_TRUNC('day', TO_TIMESTAMP(PS. "timestamp"))
	DESC;
	`
	volumes := Volumes{}
	err := d.DB.Raw(query, of).Scan(&volumes).Error
	return volumes, errors.Wrap(err, "dashboard.Volumes")
}

// Volumes implements Dashboard.
func (d *dashboard) Volumes() ([]Volume, error) {
	query := `
SELECT
	DATE_TRUNC('day', TO_TIMESTAMP(PS. "timestamp")) AS active_date,
	SUM(PS.volume0_in_price) AS volume
FROM
	pair_stats_30m AS PS
WHERE
	DATE_TRUNC('day', TO_TIMESTAMP(PS. "timestamp")) >= DATE_TRUNC('day', NOW() - INTERVAL '1 year')
GROUP BY
	DATE_TRUNC('day', TO_TIMESTAMP(PS. "timestamp"))
ORDER BY
	DATE_TRUNC('day', TO_TIMESTAMP(PS. "timestamp"))
	DESC;
`
	volumes := []Volume{}
	if err := d.DB.Raw(query).Scan(&volumes).Error; err != nil {
		return nil, errors.Wrap(err, "dashboard.Volumes")
	}
	return volumes, nil
}

func (d *dashboard) ActiveAccounts() ([]uint, error) {
	panic("unimplemented")
}

// Apr implements Dashboard.
func (d *dashboard) Apr(of Addr) ([]Apr, error) {
	panic("unimplemented")
}

// Aprs implements Dashboard.
func (d *dashboard) Aprs() ([]Apr, error) {
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
func (d *dashboard) Price(of Addr) (Prices, error) {
	panic("unimplemented")
}

// Prices implements Dashboard.
func (d *dashboard) Prices() (Prices, error) {
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

// Token implements Dashboard.
func (d *dashboard) Token(of Addr) (Token, error) {
	panic("unimplemented")
}

// Tokens implements Dashboard.
func (d *dashboard) Tokens() (Tokens, error) {
	panic("unimplemented")
}

// Tvl implements Dashboard.
func (d *dashboard) Tvl(of Addr) (Tvls, error) {
	panic("unimplemented")
}

// Tvls implements Dashboard.
func (d *dashboard) Tvls() (Tvls, error) {
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

type dashboard struct {
	chainId string
	*gorm.DB
}

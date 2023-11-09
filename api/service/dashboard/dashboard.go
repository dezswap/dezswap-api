package dashboard

import (
	"fmt"

	"github.com/dezswap/dezswap-api/pkg/db/aggregator"
	"github.com/dezswap/dezswap-api/pkg/db/parser"
	"github.com/dezswap/dezswap-api/pkg/dezswap"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type dashboard struct {
	chainId string
	*gorm.DB
}

var _ Dashboard = &dashboard{}

func NewDashboardService(chainId string, db *gorm.DB) Dashboard {
	return &dashboard{chainId, db}
}

// Aprs implements Dashboard.
func (d *dashboard) Aprs(addr ...Addr) (Aprs, error) {
	m := aggregator.PairStats30m{}
	query := d.DB.Model(
		&m,
	).Select(
		fmt.Sprintf("SUM(volume0_in_price) * %f AS volume, DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) AS timestamp", dezswap.SWAP_FEE),
	).Where(
		fmt.Sprintf("%s.chain_id = ?", m.TableName()), d.chainId,
	).Where(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) >= DATE_TRUNC('day', NOW() - INTERVAL '1 year')",
	).Group(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp))",
	).Order(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) ASC",
	)

	if len(addr) > 0 {
		joinMsg := fmt.Sprintf("LEFT JOIN pair AS P ON %s.pair_id = P.id", m.TableName())
		query = query.Where("P.contract = ?", string(addr[0])).Joins(joinMsg)
	}
	aprs := Aprs{}
	if err := query.Scan(&aprs).Error; err != nil {
		return nil, errors.Wrap(err, "dashboard.Volumes")
	}
	return aprs, nil
}

// Pools implements Dashboard.
func (d *dashboard) Pools(...Addr) (Pools, error) {
	panic("unimplemented")
}

// Prices implements Dashboard.
func (d *dashboard) Prices(...Addr) ([]Price, error) {
	panic("unimplemented")
}

// Recent implements Dashboard.
func (d *dashboard) Recent() (Recent, error) {
	panic("unimplemented")
}

// Statistic implements Dashboard.
func (d *dashboard) Statistic(addr ...Addr) (st Statistic, err error) {
	st = Statistic{}
	st.AddsCounts, err = d.dau(addr...)
	if err != nil {
		return st, errors.Wrap(err, "dashboard.Statistic")
	}
	st.TxCounts, err = d.txCounts(addr...)
	if err != nil {
		return st, errors.Wrap(err, "dashboard.Statistic")
	}
	st.Fees, err = d.fees(addr...)
	if err != nil {
		return st, errors.Wrap(err, "dashboard.Statistic")
	}

	return st, nil
}

// Tokens implements Dashboard.
func (d *dashboard) Tokens(...Addr) (Tokens, error) {
	query := `
select t.address as addr, coalesce(p.price, 0) price,
       floor((coalesce(p.price, 0)-coalesce(p24h.price, 0))/coalesce(p.price, 1)*10000)/100 as price_change
from tokens t
    left join (
        select token_id, price
        from price p
            join (
                select max(id) id
                from price
                group by token_id) t on p.id = t.id) p on t.id = p.token_id
    left join (
        select token_id, price
        from price p
            join (
                select max(id) id
                from price
                where height <= (select coalesce(max(height), 0) from parsed_tx
                  where chain_id = ? and timestamp <= extract(epoch from now() - interval '1 day'))
                group by token_id) t on p.id = t.id) p24h on t.id = p24h.token_id
where t.chain_id = ? and t.symbol != 'uLP'
`
	var tokens []Token
	if tx := d.Raw(query, d.chainId, d.chainId).Find(&tokens); tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "dashboard.Tokens")
	}

	query = `
select address,
       sum(volume) as volume,
       sum(tvl) as tvl
from (
    select t.address,
       case when p.asset0 = t.address then sum(s.sum_vol0) else sum(s.sum_vol1) end as volume,
       case when p.asset0 = t.address then sum(s.lp0) else sum(s.lp1) end as tvl
    from (
        select distinct pair_id,
                        sum(volume0_in_price) over(partition by pair_id) sum_vol0,
                        sum(volume1_in_price) over(partition by pair_id) sum_vol1,
                        first_value(liquidity0_in_price) over(partition by pair_id order by timestamp desc) lp0,
                        first_value(liquidity1_in_price) over(partition by pair_id order by timestamp desc) lp1
        from pair_stats_recent
        where timestamp >= extract(epoch from now() - interval '1 day')
          and chain_id = ?) s
    join pair p on p.id = s.pair_id
    left join tokens t on p.chain_id = t.chain_id and (p.asset0 = t.address or p.asset1 = t.address)
    group by t.address, p.asset0) t
group by address
`
	type tokenStat struct {
		Address Addr
		Volume  string
		Tvl     string
	}
	var stats []tokenStat
	if tx := d.Raw(query, d.chainId).Find(&stats); tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "dashboard.Tokens")
	}

	statMap := make(map[Addr]tokenStat, len(stats))
	for _, s := range stats {
		statMap[s.Address] = s
	}

	for i, t := range tokens {
		if stat, ok := statMap[t.Addr]; ok {
			t.Volume = stat.Volume
			t.Tvl = stat.Tvl
		} else {
			t.Volume = "0"
			t.Tvl = "0"
		}
		tokens[i] = t
	}

	return tokens, nil
}

// Tvls implements Dashboard.
func (d *dashboard) Tvls(addr ...Addr) ([]Tvl, error) {
	panic("unimplemented")
}

// Txs implements Dashboard.
func (d *dashboard) Txs(addr ...Addr) (Txs, error) {
	m := parser.ParsedTx{}
	query := d.DB.Model(
		&m,
	).Select(
		"SUM(volume0_in_price) AS volume, DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) AS timestamp",
	).Where(
		fmt.Sprintf("%s.chain_id = ?", m.TableName()), d.chainId,
	).Where(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) >= DATE_TRUNC('day', NOW() - INTERVAL '1 year')",
	).Group(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp))",
	).Order(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) ASC",
	)

	if len(addr) > 0 {
		joinMsg := fmt.Sprintf("LEFT JOIN pair AS P ON %s.pair_id = P.id", m.TableName())
		query = query.Where("P.contract = ?", string(addr[0])).Joins(joinMsg)
	}
	txs := Txs{}
	if err := query.Scan(&txs).Error; err != nil {
		return nil, errors.Wrap(err, "dashboard.Volumes")
	}
	return txs, nil
}

// Volumes implements Dashboard.
func (d *dashboard) Volumes(addr ...Addr) ([]Volume, error) {
	m := aggregator.PairStats30m{}
	query := d.DB.Model(
		&m,
	).Select(
		"SUM(volume0_in_price) AS volume, DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) AS timestamp",
	).Where(
		fmt.Sprintf("%s.chain_id = ?", m.TableName()), d.chainId,
	).Where(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) >= DATE_TRUNC('day', NOW() - INTERVAL '1 year')",
	).Group(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp))",
	).Order(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) ASC",
	)

	if len(addr) > 0 {
		joinMsg := fmt.Sprintf("LEFT JOIN pair AS P ON %s.pair_id = P.id", m.TableName())
		query = query.Where("P.contract = ?", string(addr[0])).Joins(joinMsg)
	}
	volumes := []Volume{}
	if err := query.Scan(&volumes).Error; err != nil {
		return nil, errors.Wrap(err, "dashboard.Volumes")
	}
	return volumes, nil
}

func (d *dashboard) fees(addr ...Addr) (Fees, error) {
	m := aggregator.PairStats30m{}
	query := d.DB.Model(
		&m,
	).Select(
		"SUM(commission0_in_price + commission1_in_price) AS fee, DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) AS timestamp",
	).Where(
		fmt.Sprintf("%s.chain_id = ?", m.TableName()), d.chainId,
	).Where(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) >= DATE_TRUNC('day', NOW() - INTERVAL '1 month')",
	).Group(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp))",
	).Order(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) ASC",
	)

	if len(addr) > 0 {
		joinMsg := fmt.Sprintf("LEFT JOIN pair AS P ON %s.pair_id = P.id", m.TableName())
		query = query.Where("P.contract = ?", string(addr[0])).Joins(joinMsg)
	}

	fees := Fees{}
	if err := query.Scan(&fees).Error; err != nil {
		return nil, errors.Wrap(err, "dashboard.fees")
	}
	return fees, nil
}

func (d *dashboard) dau(addr ...Addr) (AddsCounts, error) {
	m := parser.ParsedTx{}
	query := d.DB.Model(
		&m,
	).Select(
		"COUNT(DISTINCT sender) AS address_count, DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) AS timestamp",
	).Where(
		fmt.Sprintf("%s.chain_id = ?", m.TableName()), d.chainId,
	).Where(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) >= DATE_TRUNC('day', NOW() - INTERVAL '1 month')",
	).Group(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp))",
	).Order(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) ASC",
	)

	if len(addr) > 0 {
		if len(addr) > 0 {
			query = query.Where(fmt.Sprintf("%s.contract = ?", m.TableName()), addr[0])
		}
	}

	adds := AddsCounts{}
	if err := query.Scan(&adds).Error; err != nil {
		return nil, errors.Wrap(err, "dashboard.dau")
	}
	return adds, nil
}

func (d *dashboard) txCounts(addr ...Addr) (TxCounts, error) {
	m := parser.ParsedTx{}
	query := d.DB.Model(
		&m,
	).Select(
		"COUNT(*) AS tx_count, DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) AS timestamp",
	).Where(
		fmt.Sprintf("%s.chain_id = ?", m.TableName()), d.chainId,
	).Where(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) >= DATE_TRUNC('day', NOW() - INTERVAL '1 month')",
	).Group(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp))",
	).Order(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) ASC",
	)

	if len(addr) > 0 {
		query = query.Where(fmt.Sprintf("%s.contract = ?", m.TableName()), addr[0])
	}

	txs := TxCounts{}
	if err := query.Scan(&txs).Error; err != nil {
		return nil, errors.Wrap(err, "dashboard.txCounts")
	}
	return txs, nil
}

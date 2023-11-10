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
		"SUM(volume0_in_price) over AS volume, DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) AS timestamp",
	).Where(
		fmt.Sprintf("%s.chain_id = ?", m.TableName()), d.chainId,
	).Where(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) >= DATE_TRUNC('day', NOW() - INTERVAL '1 year')",
	).Group(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp))",
	).Order(
		"DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) DESC",
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
func (d *dashboard) Pools(addr ...Addr) (Pools, error) {

	timeRangeWith := `
		WITH time_range AS (
			SELECT
			CASE WHEN EXTRACT(MINUTE FROM now()) >= 30 THEN
				DATE_TRUNC('hour',now())
			ELSE
				DATE_TRUNC('hour',now()) - INTERVAL '30 min'
			END AS end_time,
			CASE WHEN EXTRACT(MINUTE FROM now()) >= 30 THEN
				DATE_TRUNC('hour',now()) - INTERVAL '7 day'
			ELSE
				DATE_TRUNC('hour',now()) - INTERVAL '7 day' - INTERVAL '30 min'
			END AS start_time
		)
	`
	latestTvls := `
		SELECT DISTINCT ON (ps.pair_id)
			ps.pair_id AS pair_id,
			p.contract AS address,
			(ps.liquidity0_in_price + ps.liquidity1_in_price) AS tvl
		FROM
			"pair_stats_30m" AS ps
			JOIN pair AS p ON ps.pair_id = p.id
		WHERE
			p.chain_id = ?
		AND
			TO_TIMESTAMP(ps. "timestamp") <= (
				SELECT
					end_time
				FROM
					time_range)
			ORDER BY
				ps.pair_id,
				ps.timestamp DESC`
	volume1d := `
		SELECT
			ps0.pair_id AS pair_id,
			SUM(volume0_in_price) AS volume
		FROM
			pair_stats_30m AS ps0
		WHERE (
			SELECT
				end_time
			FROM
				time_range) - INTERVAL '1 day' < TO_TIMESTAMP(ps0. "timestamp")
		GROUP BY
			ps0.pair_id
		ORDER BY
			ps0.pair_id
	`
	volume7d := `
		SELECT
			v.pair_id AS pair_id,
			SUM(v.volume0_in_price) AS volume
		FROM
			pair_stats_30m AS v
		WHERE
			TO_TIMESTAMP(v.timestamp) > (SELECT start_time FROM time_range)
			GROUP BY
				v.pair_id
			ORDER BY
				v.pair_id
	`
	query := fmt.Sprintf(
		`%s
		SELECT
			t.address, t.tvl, v1.volume, v1.volume * %f as fee, v7.volume/t.tvl as apr
		FROM
			(%s) AS t
			LEFT JOIN (%s) AS v1 ON v1.pair_id = t.pair_id
			LEFT JOIN (%s) AS v7 ON v7.pair_id = t.pair_id
		`,
		timeRangeWith, dezswap.SWAP_FEE, latestTvls, volume1d, volume7d,
	)
	pools := Pools{}
	if err := d.DB.Raw(query, d.chainId).Scan(&pools).Error; err != nil {
		return nil, errors.Wrap(err, "dashboard.Pools")
	}
	return pools, nil
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
func (d *dashboard) Tokens(addrs ...Addr) (Tokens, error) {
	var addrCond string
	if len(addrs) > 0 {
		addrCond += "t.address in ('" + string(addrs[0])
		for _, a := range addrs[1:] {
			addrCond += "','" + string(a)
		}
		addrCond += "')"
	}

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
	if len(addrCond) > 0 {
		query += " and " + addrCond
	}

	var tokens []Token
	if tx := d.Raw(query, d.chainId, d.chainId).Find(&tokens); tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "dashboard.Tokens")
	}

	query = `
with s as (
    select distinct pair_id,
           sum(volume0_in_price) filter (where timestamp_before = 0) over w sum_vol0,
           sum(volume1_in_price) filter (where timestamp_before = 0) over w sum_vol1,
           sum(volume0_in_price) filter (where timestamp_before > 0) over w sum_vol0_before,
           sum(volume1_in_price) filter (where timestamp_before > 0) over w sum_vol1_before,
           first_value(liquidity0_in_price) over w_lp lp0,
           first_value(liquidity1_in_price) over w_lp lp1,
           first_value(liquidity0_in_price) over w_lp_before lp0_before,
           first_value(liquidity1_in_price) over w_lp_before lp1_before
    from (
        select *, case when timestamp < extract(epoch from now() - interval '1 day') then timestamp else 0 end as timestamp_before
        from pair_stats_recent where chain_id = ?) t
    window w as (partition by pair_id),
           w_lp as (partition by pair_id order by timestamp desc),
           w_lp_before as (partition by pair_id order by timestamp_before desc)),
    s_7d as (
    select distinct pair_id,
           sum(volume0_in_price) filter (where timestamp >= extract(epoch from now() - interval '7 day')) over w sum_vol0,
           sum(volume1_in_price) filter (where timestamp >= extract(epoch from now() - interval '7 day')) over w sum_vol1,
           sum(volume0_in_price) filter (where timestamp < extract(epoch from now() - interval '7 day')) over w sum_vol0_before,
           sum(volume1_in_price) filter (where timestamp < extract(epoch from now() - interval '7 day')) over w sum_vol1_before,
           first_value(liquidity0_in_price) over w_lp lp0,
           first_value(liquidity1_in_price) over w_lp lp1
    from pair_stats_30m
    where timestamp >= extract(epoch from now() - interval '14 day')
      and timestamp < extract(epoch from now() - interval '1 day')
      and chain_id = ?
    window w as (partition by pair_id),
           w_lp as (partition by pair_id))
select address,
       coalesce(sum(volume_24h),0) as volume,
       coalesce((sum(volume_24h)-sum(volume_24h_before))/sum(volume_24h),0) as volume_change,
       coalesce(sum(volume_7d)+sum(volume_24h),0) as volume_week,
       coalesce((sum(volume_7d)+sum(volume_24h)-sum(volume_7d_before))/(sum(volume_7d)+sum(volume_24h)),0) as volume_week_change,
       coalesce(sum(tvl),0) as tvl,
       coalesce((sum(tvl)-sum(tvl_24h_before))/sum(tvl),0) as tvl_change
from (
    select t.address,
           case when p.asset0 = t.address then sum(s.sum_vol0) else sum(s.sum_vol1) end as volume_24h,
           case when p.asset0 = t.address then sum(s.sum_vol0_before) else sum(s.sum_vol1_before) end as volume_24h_before,
           case when p.asset0 = t.address then sum(s_7d.sum_vol0) else sum(s_7d.sum_vol1) end as volume_7d,
           case when p.asset0 = t.address then sum(s_7d.sum_vol0_before) else sum(s_7d.sum_vol1_before) end as volume_7d_before,
           case when p.asset0 = t.address then sum(s.lp0) else sum(s.lp1) end as tvl,
           case when p.asset0 = t.address then sum(s.lp0_before) else sum(s.lp1_before) end as tvl_24h_before
    from s
        join pair p on p.id = s.pair_id
        join s_7d  on p.id = s_7d.pair_id
        left join tokens t on p.chain_id = t.chain_id and (p.asset0 = t.address or p.asset1 = t.address)
`
	if len(addrs) > 0 {
		query += " where " + addrCond
	}
	query += " group by t.address, p.asset0) t group by address"

	type tokenStat struct {
		Address          Addr
		Volume           string
		VolumeChange     string
		VolumeWeek       string
		VolumeWeekChange string
		Tvl              string
		TvlChange        string
	}
	var stats []tokenStat
	if tx := d.Raw(query, d.chainId, d.chainId).Find(&stats); tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "dashboard.Tokens")
	}

	statMap := make(map[Addr]tokenStat, len(stats))
	for _, s := range stats {
		statMap[s.Address] = s
	}

	for i, t := range tokens {
		if stat, ok := statMap[t.Addr]; ok {
			t.Volume = stat.Volume
			t.VolumeChange = stat.VolumeChange
			t.Volume7d = stat.VolumeWeek
			t.Volume7dChange = stat.VolumeWeekChange
			t.Tvl = stat.Tvl
			t.TvlChange = stat.TvlChange
		} else {
			t.Volume = "0"
			t.VolumeChange = "0"
			t.Volume7d = "0"
			t.Volume7dChange = "0"
			t.Tvl = "0"
			t.TvlChange = "0"
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
		return nil, errors.Wrap(err, "dashboard.Txs")
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

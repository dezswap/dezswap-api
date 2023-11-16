package dashboard

import (
	"fmt"
	"time"

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

var chartCriteriaByDuration = map[ChartDuration]struct {
	Ago     string
	TruncBy time.Duration
}{
	// a month range with everyday data
	Month: {"1 month", time.Hour * 24},
	// 3 months range with every week data
	Quarter: {"3 month", time.Hour * 24 * 7},
	// 1 year range with every 2 weeks data
	Year: {"1 year", time.Hour * 24 * 7 * 2},
	// 10 years range with every month data
	All: {"10 year", time.Hour * 24 * 31},
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
func (d *dashboard) Pools() (Pools, error) {

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
			CONCAT(t0.symbol, '-', t1.symbol) AS symbols,
			(ps.liquidity0_in_price + ps.liquidity1_in_price) AS tvl
		FROM
			"pair_stats_30m" AS ps
			JOIN pair AS p ON ps.pair_id = p.id
			JOIN tokens AS t0 ON p.asset0 = t0.address
			JOIN tokens AS t1 ON p.asset1 = t1.address
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
		WHERE
			(SELECT end_time FROM time_range) - INTERVAL '1 day' < TO_TIMESTAMP(ps0. "timestamp")
			AND
			TO_TIMESTAMP(ps0. "timestamp") <= (SELECT end_time FROM time_range)
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
			(SELECT start_time FROM time_range) < TO_TIMESTAMP(v.timestamp)
		AND
			TO_TIMESTAMP(v.timestamp) <= (SELECT end_time FROM time_range)
		GROUP BY
			v.pair_id
		ORDER BY
			v.pair_id
	`
	query := fmt.Sprintf(
		`%s
		SELECT
			t.address, t.symbols, t.tvl, v1.volume, v1.volume * %f as fee, v7.volume/t.tvl as apr
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

// Pool implements Dashboard.
func (*dashboard) Pool(addr Addr) (Pools, error) {
	panic("unimplemented")
}

// Prices implements Dashboard.
func (d *dashboard) Prices(...Addr) ([]Price, error) {
	panic("unimplemented")
}

// Recent implements Dashboard.
func (d *dashboard) Recent() (Recent, error) {
	current, mins := time.Now().Truncate(time.Hour), time.Now().Minute()
	if mins >= 30 {
		current = current.Add(time.Minute * -30)
	}
	dayAgo, twoDaysAgo := current.Add(time.Hour*-24), current.Add(time.Hour*-48)
	tvl := func(at time.Time) string {
		return fmt.Sprintf(`
		SELECT
			SUM(ps.liquidity0_in_price + ps.liquidity1_in_price) AS tvl
		FROM
			pair_stats_30m as ps
			JOIN (
				SELECT
					pair_id,
					MAX(timestamp) AS timestamp
				FROM
					pair_stats_30m
				WHERE
					chain_id = '%s'
				AND
					TO_TIMESTAMP(timestamp) <= '%s'
				GROUP BY
					pair_id) AS latests ON ps.pair_id = latests.pair_id
			AND ps.timestamp = latests.timestamp`, d.chainId, at.UTC().Format(time.DateTime),
		)
	}
	volume := func(from, to time.Time) string {
		return fmt.Sprintf(`
		SELECT
			SUM(volume0_in_price) AS volume
		FROM
			pair_stats_30m
		WHERE
			chain_id = '%s'
		AND
			'%s' < TO_TIMESTAMP("timestamp")
		AND
			TO_TIMESTAMP("timestamp") <= '%s'
		`, d.chainId, from.UTC().Format(time.DateTime), to.UTC().Format(time.DateTime))
	}

	query := fmt.Sprintf(`
		WITH tvl AS (%s),
			prev_tvl AS (%s),
			volume AS (%s),
			prev_volume AS (%s)
		SELECT
			tvl.tvl AS tvl,
			CAST((tvl.tvl / prev_tvl.tvl - 1) AS float4) AS tvl_change_rate,
			volume.volume AS volume,
			CAST((volume.volume / prev_volume.volume - 1) AS float4) AS volume_change_rate,
			volume.volume * ? as fee,
			CAST((volume.volume / prev_volume.volume - 1) AS float4) AS fee_change_rate
		FROM
			tvl, prev_tvl, volume, prev_volume;
	`, tvl(current), tvl(dayAgo), volume(dayAgo, current), volume(twoDaysAgo, dayAgo))
	recent := Recent{}
	if err := d.DB.Raw(query, dezswap.SWAP_FEE).Scan(&recent).Error; err != nil {
		return recent, errors.Wrap(err, "dashboard.Recent")
	}
	return recent, nil
}

// Statistic implements Dashboard.
func (d *dashboard) Statistic(addr ...Addr) (st Statistic, err error) {
	subDau := d.dau(addr...)
	subTxCnts := d.txCounts(addr...)
	subFees := d.fees(addr...)

	query := `
	WITH dau AS (?),
		tx_counts AS (?),
		fees AS (?)
	SELECT
		dau.address_count,
		tx_counts.tx_count,
		fees.fee,
		dau.timestamp
	FROM
		dau
		JOIN tx_counts ON dau.timestamp = tx_counts.timestamp
		JOIN fees ON dau.timestamp = fees.timestamp
	ORDER BY dau.timestamp ASC
	`

	st = Statistic{}
	if err := d.DB.Raw(query, subDau, subTxCnts, subFees).Scan(&st).Error; err != nil {
		return nil, errors.Wrap(err, "dashboard.Pools")
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
           sum(commission0_in_price) filter (where timestamp_before = 0) over w sum_com0,
           sum(commission1_in_price) filter (where timestamp_before = 0) over w sum_com1,
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
       coalesce((sum(tvl)-sum(tvl_24h_before))/sum(tvl),0) as tvl_change,
       coalesce(sum(commission),0) as commission
from (
    select t.address,
           case when p.asset0 = t.address then sum(s.sum_vol0) else sum(s.sum_vol1) end as volume_24h,
           case when p.asset0 = t.address then sum(s.sum_vol0_before) else sum(s.sum_vol1_before) end as volume_24h_before,
           case when p.asset0 = t.address then sum(s_7d.sum_vol0) else sum(s_7d.sum_vol1) end as volume_7d,
           case when p.asset0 = t.address then sum(s_7d.sum_vol0_before) else sum(s_7d.sum_vol1_before) end as volume_7d_before,
           case when p.asset0 = t.address then sum(s.lp0) else sum(s.lp1) end as tvl,
           case when p.asset0 = t.address then sum(s.lp0_before) else sum(s.lp1_before) end as tvl_24h_before,
           case when p.asset0 = t.address then sum(s.sum_com0) else sum(s.sum_com1) end as commission
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
		Commission       string
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
			t.Commission = stat.Commission
		} else {
			t.Volume = "0"
			t.VolumeChange = "0"
			t.Volume7d = "0"
			t.Volume7dChange = "0"
			t.Tvl = "0"
			t.TvlChange = "0"
			t.Commission = "0"
		}
		tokens[i] = t
	}

	return tokens, nil
}

func (d *dashboard) TokenVolumes(addr Addr, itv Duration) (TokenChart, error) {
	query := `
select cast(extract(epoch from make_date(year_utc, month_utc, 1)::timestamp) as varchar) as timestamp, coalesce(sum(volume), 0) as value
from (
    select year_utc, month_utc,
           case when p.asset0 = t.address then sum(ps.volume0_in_price) else sum(ps.volume1_in_price) end as volume
    from pair_stats_30m ps
    join pair p on p.id = ps.pair_id
    join tokens t on p.chain_id = t.chain_id and (p.asset0 = t.address or p.asset1 = t.address)
    where ps.chain_id = ?
      and t.address = ?
    group by year_utc, month_utc, day_utc, p.asset0, t.address
) t
group by year_utc, month_utc
order by year_utc, month_utc
`
	switch itv {
	case month:
		query = `
select cast(extract(epoch from make_date(year_utc, month_utc, day_utc)::timestamp) as varchar) as timestamp, coalesce(sum(volume), 0) as value
from (
    select year_utc, month_utc, day_utc,
           case when p.asset0 = t.address then sum(ps.volume0_in_price) else sum(ps.volume1_in_price) end as volume
    from pair_stats_30m ps
    join pair p on p.id = ps.pair_id
    join tokens t on p.chain_id = t.chain_id and (p.asset0 = t.address or p.asset1 = t.address)
    where ps.chain_id = ?
      and t.address = ?
      and ps.timestamp >= extract(epoch from now() - interval '1 month')
    group by year_utc, month_utc, day_utc, p.asset0, t.address
) t
group by year_utc, month_utc, day_utc
order by year_utc, month_utc, day_utc
`
	case quarter:
		query = `
select cast(extract(epoch from to_date(concat(year_utc, week), 'iyyyiw')::timestamp at time zone 'UTC' + interval '6 day') as varchar) as timestamp, coalesce(sum(volume), 0)  as value
from (
    select year_utc,
           least(ceil((extract(doy from to_timestamp(timestamp) at time zone 'UTC'))/7), 52) as week,
           case when p.asset0 = t.address then sum(ps.volume0_in_price) else sum(ps.volume1_in_price) end as volume
    from pair_stats_30m ps
    join pair p on p.id = ps.pair_id
    join tokens t on p.chain_id = t.chain_id and (p.asset0 = t.address or p.asset1 = t.address)
    where ps.chain_id = ?
      and t.address = ?
      and ps.timestamp >= extract(epoch from now() - interval '3 month')
    group by year_utc, week, p.asset0, t.address
) t
group by year_utc, week
order by year_utc, week
`
	case year:
		query = `
select cast(extract(epoch from to_date(concat(year_utc, week2), 'iyyyiw')::timestamp at time zone 'UTC' + interval '15 day') as varchar) as timestamp,
       coalesce(sum(volume), 0) as value
from (
    select year_utc,
           week-mod(cast(week+1 as bigint),2) week2,
           case when p.asset0 = t.address then sum(ps.volume0_in_price) else sum(ps.volume1_in_price) end as volume
    from (select least(ceil((extract(doy from to_timestamp(timestamp) at time zone 'UTC'))/7), 52) as week, * from pair_stats_30m) ps
    join pair p on p.id = ps.pair_id
    join tokens t on p.chain_id = t.chain_id and (p.asset0 = t.address or p.asset1 = t.address)
    where ps.chain_id = ?
      and t.address = ?
      and ps.timestamp >= extract(epoch from now() - interval '1 year')
    group by year_utc, week, p.asset0, t.address
) t
group by year_utc, week2
order by year_utc, week2
`
	}
	var chart TokenChart
	if tx := d.Raw(query, d.chainId, addr).Find(&chart); tx.Error != nil {
		return TokenChart{}, errors.Wrap(tx.Error, "dashboard.TokenVolumes")
	}

	return chart, nil
}

func (d *dashboard) TokenTvls(addr Addr, itv Duration) (TokenChart, error) {
	query := `
select cast(extract(epoch from make_date(year_utc, month_utc, 1)::timestamp) as varchar) as timestamp, sum(tvl) as value
from (select distinct pair_id, year_utc, month_utc,
           case when p.asset0 = t.address then
               first_value(ps.liquidity0_in_price) over (partition by pair_id, year_utc, month_utc order by timestamp desc)
           else
               first_value(ps.liquidity0_in_price) over (partition by pair_id, year_utc, month_utc order by timestamp desc)
           end as tvl
    from pair_stats_30m ps
    join pair p on p.id = ps.pair_id
    join tokens t on p.chain_id = t.chain_id and (p.asset0 = t.address or p.asset1 = t.address)
    where ps.chain_id = ?
      and t.address = ?) t
group by year_utc, month_utc
order by year_utc, month_utc
`
	switch itv {
	case month:
		query = `
select cast(extract(epoch from make_date(year_utc, month_utc, day_utc)::timestamp) as varchar) as timestamp, sum(tvl) as value
from (select distinct pair_id, year_utc, month_utc, day_utc,
           case when p.asset0 = t.address then
               first_value(ps.liquidity0_in_price) over (partition by pair_id, year_utc, month_utc, day_utc order by timestamp desc)
           else
               first_value(ps.liquidity0_in_price) over (partition by pair_id, year_utc, month_utc, day_utc order by timestamp desc)
           end as tvl
    from pair_stats_30m ps
    join pair p on p.id = ps.pair_id
    join tokens t on p.chain_id = t.chain_id and (p.asset0 = t.address or p.asset1 = t.address)
    where ps.chain_id = ?
      and t.address = ?
      and ps.timestamp >= extract(epoch from now() - interval '1 month')) t
group by year_utc, month_utc, day_utc
order by year_utc, month_utc, day_utc
`
	case quarter:
		query = `
select cast(extract(epoch from to_date(concat(year_utc, week), 'iyyyiw')::timestamp at time zone 'UTC' + interval '6 day') as varchar) as timestamp, sum(tvl) as value
from (select distinct pair_id, year_utc, week,
           min(timestamp) over (partition by pair_id, year_utc, week) as start,
           case when p.asset0 = t.address then
               first_value(ps.liquidity0_in_price) over (partition by pair_id, year_utc, week order by timestamp desc)
           else
               first_value(ps.liquidity0_in_price) over (partition by pair_id, year_utc, week order by timestamp desc)
           end as tvl
    from (select least(ceil((extract(doy from to_timestamp(timestamp) at time zone 'UTC'))/7), 52) as week, * from pair_stats_30m) ps
    join pair p on p.id = ps.pair_id
    join tokens t on p.chain_id = t.chain_id and (p.asset0 = t.address or p.asset1 = t.address)
    where ps.chain_id = ?
      and t.address = ?
      and ps.timestamp >= extract(epoch from now() - interval '3 month')) t
group by year_utc, week
order by year_utc, week
`
	case year:
		query = `
select cast(extract(epoch from to_date(concat(year_utc, week2), 'iyyyiw')::timestamp at time zone 'UTC' + interval '15 day') as varchar) as timestamp, sum(tvl) as value
from (select distinct on (pair_id, year_utc, week-mod(cast(week+1 as bigint),2))
           pair_id, year_utc, week-mod(cast(week+1 as bigint),2) as week2,
           case when p.asset0 = t.address then
               first_value(ps.liquidity0_in_price) over (partition by pair_id, year_utc, week-mod(cast(week as bigint),2) order by timestamp desc)
           else
               first_value(ps.liquidity0_in_price) over (partition by pair_id, year_utc, week-mod(cast(week as bigint),2) order by timestamp desc)
           end as tvl
    from (select least(ceil((extract(doy from to_timestamp(timestamp) at time zone 'UTC'))/7), 52) as week, * from pair_stats_30m) ps
    join pair p on p.id = ps.pair_id
    join tokens t on p.chain_id = t.chain_id and (p.asset0 = t.address or p.asset1 = t.address)
    where ps.chain_id = ?
      and t.address = ?
      and ps.timestamp >= extract(epoch from now() - interval '1 year')) t
group by year_utc, week2
order by year_utc, week2
`
	}
	var chart TokenChart
	if tx := d.Raw(query, d.chainId, addr).Find(&chart); tx.Error != nil {
		return TokenChart{}, errors.Wrap(tx.Error, "dashboard.TokenTvls")
	}

	return chart, nil
}

func (d *dashboard) TokenPrices(addr Addr, itv Duration) (TokenChart, error) {
	query := `
select distinct cast(extract(epoch from make_date(year_utc, month_utc, 1)::timestamp) as varchar) as timestamp,
                first_value(price) over (partition by year_utc, month_utc order by height desc) as value
from (select p.height,
             cast(extract(year from to_timestamp(pt.timestamp) at time zone 'UTC') as int) year_utc,
             cast(extract(month from to_timestamp(pt.timestamp) at time zone 'UTC') as int) month_utc,
             p.price
      from price p
          join tokens t on p.token_id = t.id
          join parsed_tx pt on p.chain_id = pt.chain_id and p.height = pt.height
      where t.chain_id = ?
        and t.address= ?) t
order by timestamp asc
`
	switch itv {
	case month:
		query = `
select distinct cast(extract(epoch from make_date(year_utc, month_utc, day_utc)::timestamp) as varchar) as timestamp,
                first_value(price) over (partition by year_utc, month_utc, day_utc order by height desc) as value
from (select p.height,
             cast(extract(year from to_timestamp(pt.timestamp) at time zone 'UTC') as int) year_utc,
             cast(extract(month from to_timestamp(pt.timestamp) at time zone 'UTC') as int) month_utc,
             cast(extract(day from to_timestamp(pt.timestamp) at time zone 'UTC') as int) day_utc,
             p.price
      from price p
          join tokens t on p.token_id = t.id
          join parsed_tx pt on p.chain_id = pt.chain_id and p.height = pt.height
      where t.chain_id = ?
        and t.address= ?
        and ps.timestamp >= extract(epoch from now() - interval '1 month')) t
order by timestamp asc
`
	case quarter:
		query = `
select distinct cast(extract(epoch from to_date(concat(year_utc, week), 'iyyyiw')::timestamp at time zone 'UTC' + interval '6 day') as varchar) as timestamp,
                first_value(price) over (partition by year_utc, week order by height desc) as value
from (select p.height,
             cast(extract(year from to_timestamp(pt.timestamp) at time zone 'UTC') as int) year_utc,
             least(ceil((extract(doy from to_timestamp(timestamp) at time zone 'UTC'))/7), 52) as week,
             p.price
      from price p
          join tokens t on p.token_id = t.id
          join parsed_tx pt on p.chain_id = pt.chain_id and p.height = pt.height
      where t.chain_id = ?
        and t.address= ?
        and ps.timestamp >= extract(epoch from now() - interval '3 month')) t
order by timestamp asc
`
	case year:
		query = `
select distinct cast(extract(epoch from to_date(concat(year_utc, week2), 'iyyyiw')::timestamp at time zone 'UTC' + interval '15 day') as varchar) as timestamp,
                first_value(price) over (partition by year_utc, week2 order by height desc) as value
from (select p.height,
             cast(extract(year from to_timestamp(pt.timestamp) at time zone 'UTC') as int) year_utc,
             week-mod(cast(week+1 as bigint),2) week2,
             p.price
      from price p
          join tokens t on p.token_id = t.id
          join (select least(ceil((extract(doy from to_timestamp(timestamp) at time zone 'UTC'))/7), 52) as week, * from parsed_tx) pt on p.chain_id = pt.chain_id and p.height = pt.height
      where t.chain_id = ?
        and t.address= ?
        and ps.timestamp >= extract(epoch from now() - interval '1 year')) t
order by timestamp asc
`
	}
	var chart TokenChart
	if tx := d.Raw(query, d.chainId, addr).Find(&chart); tx.Error != nil {
		return TokenChart{}, errors.Wrap(tx.Error, "dashboard.TokenTvls")
	}

	return chart, nil
}

// Tvls implements Dashboard.
func (d *dashboard) Tvls(addr ...Addr) ([]Tvl, error) {
	panic("unimplemented")
}

// Txs implements Dashboard.
func (d *dashboard) Txs(addr ...Addr) (Txs, error) {
	m := parser.ParsedTx{}
	subQuery := d.DB.Model(&m).Select("*").Where("chain_id = ?", d.chainId).Order("timestamp DESC").Limit(100)
	if len(addr) > 0 {
		subQuery = subQuery.Where("contract = ?", string(addr[0]))
	}

	query := d.DB.Select(
		`pt.type AS action,
		pt.hash AS hash,
		pt.contract AS address,
		pt.sender AS sender,
		pt.asset0 AS asset0,
		t0.symbol AS asset0_symbol,
		pt.asset0_amount AS asset0_amount,
		pt.asset1 AS asset1,
		t1.symbol AS asset1_symbol,
		pt.asset1_amount AS asset1_amount,
		0 as total_value,
		TO_TIMESTAMP(pt."timestamp") as timestamp`,
	).Table("(?) as pt", subQuery).Joins(`
		JOIN tokens AS t0 ON pt.asset0 = t0.address AND pt.chain_id = t0.chain_id
		JOIN tokens AS t1 ON pt.asset1 = t1.address AND pt.chain_id = t0.chain_id
		`,
	// TODO(join and find price for total_value when it is ready)
	).Order(`pt. "timestamp" DESC`)

	txs := Txs{}
	if err := query.Scan(&txs).Error; err != nil {
		return nil, errors.Wrap(err, "dashboard.Txs")
	}
	return txs, nil
}

// Volumes implements Dashboard.
func (d *dashboard) Volumes(duration ChartDuration) (Volumes, error) {
	truncBy := int64(chartCriteriaByDuration[duration].TruncBy.Truncate(time.Second).Seconds())
	intervalAgo := chartCriteriaByDuration[duration].Ago
	query := fmt.Sprintf(`
		SELECT
			SUM(volume0_in_price) AS volume,
			TO_TIMESTAMP(FLOOR("timestamp" / %d ) * %d) AT TIME ZONE 'UTC' as timestamp
		FROM
			pair_stats_30m
		WHERE
			chain_id = ?
			AND "timestamp" > EXTRACT(EPOCH FROM NOW() - INTERVAL '%s')
		GROUP BY
			FLOOR("timestamp" / %d )
		ORDER BY
			FLOOR("timestamp" / %d )
	`, truncBy, truncBy, intervalAgo, truncBy, truncBy)

	volumes := Volumes{}
	if err := d.DB.Raw(query, d.chainId).Scan(&volumes).Error; err != nil {
		return nil, errors.Wrap(err, "dashboard.Volumes")
	}
	return volumes, nil
}

// Volumes implements Dashboard.
func (d *dashboard) VolumesOf(addr Addr, duration ChartDuration) (Volumes, error) {
	truncBy := int64(chartCriteriaByDuration[duration].TruncBy.Truncate(time.Second).Seconds())
	intervalAgo := chartCriteriaByDuration[duration].Ago
	query := fmt.Sprintf(`
		SELECT
			SUM(volume0_in_price) AS volume,
			TO_TIMESTAMP(FLOOR(ps."timestamp" / %d ) * %d) AT TIME ZONE 'UTC' as timestamp
		FROM
			pair_stats_30m AS ps
			JOIN pair as p on ps.pair_id = p.id
		WHERE
			ps.chain_id = ?
			AND p.contract = ?
			AND ps."timestamp" > EXTRACT(EPOCH FROM NOW() - INTERVAL '%s')
		GROUP BY
			FLOOR(ps."timestamp" / %d )
		ORDER BY
			FLOOR(ps."timestamp" / %d )
	`, truncBy, truncBy, intervalAgo, truncBy, truncBy)

	volumes := Volumes{}
	if err := d.DB.Raw(query, d.chainId, addr).Scan(&volumes).Error; err != nil {
		return nil, errors.Wrap(err, "dashboard.Volumes")
	}
	return volumes, nil
}

func (d *dashboard) fees(addr ...Addr) *gorm.DB {
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

	return query
}

func (d *dashboard) dau(addr ...Addr) *gorm.DB {
	m := parser.ParsedTx{}
	query := d.DB.Model(
		&m,
	).Select(`
		COUNT(DISTINCT sender) AS address_count,
		DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) AT TIME ZONE 'UTC' AS timestamp`,
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
	return query
}

func (d *dashboard) txCounts(addr ...Addr) *gorm.DB {
	m := parser.ParsedTx{}
	query := d.DB.Model(
		&m,
	).Select(`
		COUNT(*) AS tx_count,
		DATE_TRUNC('day', TO_TIMESTAMP(timestamp)) AS timestamp`,
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

	return query
}

package dashboard

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	ds "github.com/dezswap/dezswap-api/api/service/dashboard"
	"github.com/pkg/errors"
)

type mapper struct{}

func (m *mapper) tokenToRes(token ds.Token) TokenRes {
	return TokenRes{
		Address:         string(token.Addr),
		Price:           token.Price,
		PriceChange:     token.PriceChange,
		Volume24h:       token.Volume,
		Volume24hChange: token.VolumeChange,
		Volume7d:        token.Volume7d,
		Volume7dChange:  token.Volume7dChange,
		Tvl:             token.Tvl,
		TvlChange:       token.TvlChange,
		Fee:             token.Commission,
	}
}

func (m *mapper) tokensToRes(tokens ds.Tokens) TokensRes {
	res := make(TokensRes, len(tokens))
	for i, t := range tokens {
		res[i] = m.tokenToRes(t)
	}

	return res
}

func (m *mapper) recentToRes(recent ds.Recent) RecentRes {
	return RecentRes{
		Volume:           recent.Volume,
		VolumeChangeRate: recent.VolumeChangeRate,
		Fee:              recent.Fee,
		FeeChangeRate:    recent.FeeChangeRate,
		Tvl:              recent.Tvl,
		TvlChangeRate:    recent.TvlChangeRate,
		Apr:              recent.Apr,
		AprChangeRate:    recent.AprChangeRate,
	}
}

func (m *mapper) statisticToRes(statistic ds.Statistic) StatisticRes {
	res := make(StatisticRes, len(statistic))
	for i, s := range statistic {
		res[i] = StatisticResItem{
			AddressCount: s.AddressCount,
			TxCount:      s.TxCount,
			Fee:          s.Fee,
			Timestamp:    s.Timestamp,
		}
	}
	return res
}

func (m *mapper) txsToRes(txs ds.Txs) TxsRes {
	actionConverter := func(action string) string {
		actionStr := m.serviceTxTypeToTxTypeString(action)
		actionStr = strings.ReplaceAll(actionStr, "_", " ")
		return fmt.Sprintf("%s%s", strings.ToUpper(actionStr[0:1]), actionStr[1:])
	}

	type asset struct {
		Asset  string
		Amount string
		Symbol string
	}
	type assets [2]asset

	actionDisplayConverter := func(action ds.TxType, orderedAssets [2]asset) string {
		switch action {
		case ds.TX_TYPE_SWAP:
			return fmt.Sprintf("%s %s for %s", actionConverter(string(action)), orderedAssets[0].Symbol, orderedAssets[1].Symbol)
		default:
			return fmt.Sprintf("%s %s and %s", actionConverter(string(action)), orderedAssets[0].Symbol, orderedAssets[1].Symbol)
		}
	}

	orderAssets := func(action ds.TxType, unorderedAssets assets) assets {
		switch action {
		case ds.TX_TYPE_SWAP:
			if strings.Contains(unorderedAssets[0].Amount, "-") {
				return [2]asset{unorderedAssets[1], unorderedAssets[0]}
			}
		}
		return unorderedAssets
	}

	res := make(TxsRes, len(txs))
	for i, tx := range txs {
		targetAssets := assets{
			{Asset: tx.Asset0, Amount: tx.Asset0Amount, Symbol: tx.Asset0Symbol},
			{Asset: tx.Asset1, Amount: tx.Asset1Amount, Symbol: tx.Asset1Symbol},
		}
		targetAssets = orderAssets(ds.TxType(tx.Action), targetAssets)
		res[i] = TxRes{
			Action:        m.serviceTxTypeToTxTypeString(tx.Action),
			ActionDisplay: actionDisplayConverter(ds.TxType(tx.Action), targetAssets),
			Hash:          tx.Hash,
			Address:       tx.Address,
			Asset0:        targetAssets[0].Asset,
			Asset0Amount:  strings.ReplaceAll(targetAssets[0].Amount, "-", ""),
			Asset1:        targetAssets[1].Asset,
			Asset1Amount:  strings.ReplaceAll(targetAssets[1].Amount, "-", ""),
			TotalValue:    tx.TotalValue,
			Account:       tx.Sender,
			Timestamp:     tx.Timestamp,
		}
	}
	return res
}

func (m *mapper) poolsToRes(pools ds.Pools) PoolsRes {
	res := make(PoolsRes, len(pools))

	for i, p := range pools {
		res[i] = PoolRes{
			Address: p.Address,
			Tvl:     p.Tvl,
			Volume:  p.Volume,
			Fee:     p.Fee,
			Apr:     p.Apr,
		}
	}

	return res
}

func (m *mapper) poolDetailToRes(pool ds.PoolDetail) PoolDetailRes {
	res := PoolDetailRes{}

	res.Recent = m.recentToRes(pool.Recent)
	res.Txs = m.txsToRes(pool.Txs)

	return res
}

func (m *mapper) volumesToChartRes(volumes ds.Volumes) ChartRes {
	res := make(ChartRes, len(volumes))

	for i, v := range volumes {
		res[i] = ChartItem{
			Value:     v.Volume,
			Timestamp: v.Timestamp,
		}
	}
	return res
}

func (m *mapper) tvlsToChartRes(tvls ds.Tvls) ChartRes {
	res := make(ChartRes, len(tvls))

	for i, v := range tvls {
		res[i] = ChartItem{
			Value:     v.Tvl,
			Timestamp: v.Timestamp,
		}
	}
	return res
}

func (m *mapper) aprsToChartRes(aprs ds.Aprs) ChartRes {
	res := make(ChartRes, len(aprs))

	for i, v := range aprs {
		res[i] = ChartItem{
			Value:     v.Apr,
			Timestamp: v.Timestamp,
		}
	}
	return res
}

func (m *mapper) feesToChartRes(aprs ds.Fees) ChartRes {
	res := make(ChartRes, len(aprs))

	for i, v := range aprs {
		res[i] = ChartItem{
			Value:     v.Fee,
			Timestamp: v.Timestamp,
		}
	}
	return res
}

func (m *mapper) tokenChartToChartRes(chart ds.TokenChart) (ChartRes, error) {
	res := make(ChartRes, len(chart))

	for i, v := range chart {
		timestamp, err := strconv.ParseInt(v.Timestamp, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "mapper.tokenChartToChartRes")
		}
		t := time.Unix(timestamp, 0).UTC()
		res[i] = ChartItem{Timestamp: t, Value: v.Value}
	}

	return res, nil
}

func (m *mapper) txTypeToServiceTxType(ty TxType) ds.TxType {
	switch ty {
	case TX_TYPE_SWAP:
		return ds.TX_TYPE_SWAP
	case TX_TYPE_ADD:
		return ds.TX_TYPE_PROVIDE
	case TX_TYPE_REMOVE:
		return ds.TX_TYPE_WITHDRAW
	}
	return ds.TX_TYPE_ALL
}

func (m *mapper) serviceTxTypeToTxTypeString(ty string) string {
	switch ds.TxType(ty) {
	case ds.TX_TYPE_SWAP:
		return string(TX_TYPE_SWAP)
	case ds.TX_TYPE_PROVIDE:
		return string(TX_TYPE_ADD)
	case ds.TX_TYPE_WITHDRAW:
		return string(TX_TYPE_REMOVE)
	}

	// return the tx type as is
	return ty
}

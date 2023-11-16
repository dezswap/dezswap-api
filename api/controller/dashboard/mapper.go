package dashboard

import (
	"fmt"
	"strings"

	dashboardService "github.com/dezswap/dezswap-api/api/service/dashboard"
)

type mapper struct{}

func (m *mapper) tokenToRes(token dashboardService.Token) TokenRes {
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

func (m *mapper) tokensToRes(tokens dashboardService.Tokens) TokensRes {
	res := make(TokensRes, len(tokens))
	for i, t := range tokens {
		res[i] = m.tokenToRes(t)
	}

	return res
}

func (m *mapper) tokenChartToRes(chart dashboardService.TokenChart) TokenChart {
	res := make(TokenChart, len(chart))

	for i, v := range chart {
		res[i] = [2]string{v.Timestamp, v.Value}
	}

	return res
}

func (m *mapper) recentToRes(recent dashboardService.Recent) RecentRes {
	return RecentRes{
		Volume:           recent.Volume,
		VolumeChangeRate: recent.VolumeChangeRate,
		Fee:              recent.Fee,
		FeeChangeRate:    recent.FeeChangeRate,
		Tvl:              recent.Tvl,
		TvlChangeRate:    recent.TvlChangeRate,
	}
}

func (m *mapper) statisticToRes(statistic dashboardService.Statistic) StatisticRes {
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

func (m *mapper) txsToRes(txs dashboardService.Txs) TxsRes {
	actionConverter := func(action string) string {
		switch action {
		case "swap":
			return "Swap"
		case "provide":
			return "Add"
		case "withdraw":
			return "Remove"
		default:
			str := strings.ReplaceAll(action, "_", " ")
			return fmt.Sprintf("%s%s", strings.ToUpper(str[0:1]), str[1:])
		}

	}
	res := make(TxsRes, len(txs))
	for i, tx := range txs {
		res[i] = TxRes{
			Action:       fmt.Sprintf("%s %s and %s", actionConverter(tx.Action), tx.Asset0Symbol, tx.Asset1Symbol),
			Hash:         tx.Hash,
			Address:      tx.Address,
			Asset0:       tx.Asset0,
			Asset0Amount: tx.Asset0Amount,
			Asset1:       tx.Asset1,
			Asset1Amount: tx.Asset1Amount,
			TotalValue:   tx.TotalValue,
			Account:      tx.Sender,
			Timestamp:    tx.Timestamp,
		}
	}
	return res
}

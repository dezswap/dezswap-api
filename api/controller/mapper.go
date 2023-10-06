package controller

import (
	"github.com/dezswap/dezswap-api/api/service"
	"github.com/dezswap/dezswap-api/pkg/dezswap"
)

type pairMapper struct{}
type poolMapper struct{}
type tokenMapper struct{}

type statMapper struct{}

func (m *poolMapper) poolToRes(pool service.Pool) PoolRes {
	res := PoolRes{
		Address: pool.Address,
		PoolRes: &dezswap.PoolRes{},
	}
	res.TotalShare = pool.LpAmount
	res.Assets = []dezswap.AssetInfoRes{
		dezswap.ToAssetInfoRes(pool.Asset0, pool.Asset0Amount),
		dezswap.ToAssetInfoRes(pool.Asset1, pool.Asset1Amount),
	}
	return res
}
func (m *poolMapper) poolsToRes(pools []service.Pool) []PoolRes {
	res := make([]PoolRes, len(pools))
	for i, pool := range pools {
		res[i] = m.poolToRes(pool)
	}
	return res
}

func (m *pairMapper) pairToRes(pair service.Pair) PairRes {
	res := PairRes{
		PairRes: &dezswap.PairRes{
			ContractAddr: pair.Address,
			AssetInfos: []dezswap.AssetInfoTokenRes{
				dezswap.ToAssetInfoTokenRes(pair.Asset0.Address),
				dezswap.ToAssetInfoTokenRes(pair.Asset1.Address),
			},
			LiquidityToken: pair.Lp.Address,
			AssetDecimals:  []uint{uint(pair.Asset0.Decimals), uint(pair.Asset1.Decimals)},
		},
	}
	return res
}

func (m *pairMapper) pairsToRes(pairs []service.Pair) PairsRes {
	res := make([]PairRes, len(pairs))
	for i, pair := range pairs {
		res[i] = m.pairToRes(pair)
	}
	return PairsRes{Pairs: res}
}

func (m *tokenMapper) tokenToRes(token service.Token) TokenRes {
	res := TokenRes{
		ChainId:  token.ChainId,
		Token:    token.Address,
		Name:     token.Name,
		Symbol:   token.Symbol,
		Decimals: token.Decimals,
		Icon:     token.Icon,
		Protocol: token.Protocol,
		Verified: token.Verified,
	}
	return res
}

func (m *tokenMapper) tokensToRes(tokens []service.Token) []TokenRes {
	res := make([]TokenRes, len(tokens))
	for i, token := range tokens {
		res[i] = m.tokenToRes(token)
	}
	return res
}

func (m *statMapper) statToRes(stat service.PairStats) StatRes {
	volumes := make([]StatValueRes, 0, len(stat))
	fees := make([]StatValueRes, 0, len(stat))
	aprs := make([]StatValueRes, 0, len(stat))

	for _, s := range stat {
		volumes = append(volumes, StatValueRes{Address: s.Address, Value: s.VolumeInPrice})
		fees = append(fees, StatValueRes{Address: s.Address, Value: s.CommissionInPrice})
		aprs = append(aprs, StatValueRes{Address: s.Address, Value: s.AprInPrice})
	}

	return StatRes{Volume: volumes, Fee: fees, Apr: aprs}
}

func (m *statMapper) statsToRes(stats []service.PairStats) StatsRes {
	return StatsRes{
		Stats24h:  m.statToRes(stats[service.Period24h]),
		Stats7d:   m.statToRes(stats[service.Period7d]),
		Stats1mon: m.statToRes(stats[service.Period1mon]),
	}
}

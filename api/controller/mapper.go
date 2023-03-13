package controller

import (
	"github.com/dezswap/dezswap-api/api/service"
	"github.com/dezswap/dezswap-api/pkg/dezswap"
)

type poolMapper struct{}

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

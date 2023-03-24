package dezswap

import (
	"github.com/dezswap/dezswap-api/pkg/xpla"
)

func ToAssetInfoRes(addr string, amount string) AssetInfoRes {
	var assetAmount *string
	if amount != "" {
		assetAmount = &amount
	}
	res := AssetInfoRes{
		Amount: assetAmount,
		Info:   AssetInfoTokenRes{},
	}
	if xpla.IsCw20(addr) {
		res.Info.Token = &TokenAssetInfoRes{
			ContractAddress: addr,
		}
	} else {
		res.Info.NativeToken = &NativeTokenAssetInfoRes{
			Denom: addr,
		}
	}
	return res
}

func ToAssetInfoTokenRes(addr string) AssetInfoTokenRes {
	res := AssetInfoTokenRes{}
	if xpla.IsCw20(addr) {
		res.Token = &TokenAssetInfoRes{
			ContractAddress: addr,
		}
	} else {
		res.NativeToken = &NativeTokenAssetInfoRes{
			Denom: addr,
		}
	}
	return res
}

package dezswap

import (
	"github.com/dezswap/dezswap-api/pkg"
)

func ToAssetInfoRes(addr string, amount string, networkMetadata pkg.NetworkMetadata) AssetInfoRes {
	var assetAmount *string
	if amount != "" {
		assetAmount = &amount
	}
	res := AssetInfoRes{
		Amount: assetAmount,
		Info:   AssetInfoTokenRes{},
	}
	if networkMetadata.IsCw20(addr) {
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

func ToAssetInfoTokenRes(addr string, networkMetadata pkg.NetworkMetadata) AssetInfoTokenRes {
	res := AssetInfoTokenRes{}
	if networkMetadata.IsCw20(addr) {
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

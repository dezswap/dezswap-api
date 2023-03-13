package dezswap

import "github.com/dezswap/dezswap-api/pkg/xpla"

func ToAssetInfoRes(addr string, amount string) AssetInfoRes {
	res := AssetInfoRes{
		Amount: amount,
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

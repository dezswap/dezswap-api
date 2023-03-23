package dezswap

type PairRes struct {
	AssetInfos     []AssetInfoRes `json:"asset_infos"`
	ContractAddr   string         `json:"contract_addr"`
	LiquidityToken string         `json:"liquidity_token"`
	AssetDecimals  []uint         `json:"asset_decimals"`
}

type TokenInfoRes struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Decimals    uint   `json:"decimals"`
	TotalSupply string `json:"total_supply"`
}

type PoolRes struct {
	Assets     []AssetInfoRes `json:"assets"`
	TotalShare string         `json:"total_share"`
}

type NativeTokenAssetInfoRes struct {
	Denom string `json:"denom"`
}

type TokenAssetInfoRes struct {
	ContractAddress string `json:"contract_addr"`
}

type AssetInfoRes struct {
	Info   AssetInfoTokenRes `json:"info"`
	Amount *string           `json:"amount,omitempty"`
}

type AssetInfoTokenRes struct {
	Token       *TokenAssetInfoRes       `json:"token,omitempty"`
	NativeToken *NativeTokenAssetInfoRes `json:"native_token,omitempty"`
}

func (p *PoolRes) GetAsset(idx uint) string {
	if p.Assets[idx].Info.Token != nil {
		return p.Assets[idx].Info.Token.ContractAddress
	}
	return p.Assets[idx].Info.NativeToken.Denom
}

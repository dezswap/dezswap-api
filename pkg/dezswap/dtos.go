package dezswap

type TokenInfoRes struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Decimals    uint   `json:"decimals"`
	TotalSupply string `json:"total_supply"`
}

type PoolRes struct {
	AssetInfos []AssetInfo
	TotalShare string `json:"total_share"`
}

type AssetInfo struct {
	Token       *TokenAsset `json:"token"`
	NativeToken *DenomAsset `json:"native_token"`
	Amount      string      `json:"amount"`
}

type TokenAsset struct {
	ContractAddress string `json:"contract_address"`
}

type DenomAsset struct {
	Denom string `json:"denom"`
}

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
	Token *struct {
		ContractAddress string `json:"contract_address"`
	} `json:"token"`
	NativeToken *struct {
		Denom string `json:"denom"`
	}
	Amount string `json:"amount"`
}

package dezswap

type TokenInfoRes struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Decimals    uint   `json:"decimals"`
	TotalSupply string `json:"total_supply"`
}

type PoolRes struct {
	Assets []struct {
		Info struct {
			Token *struct {
				ContractAddress string `json:"contract_addr"`
			} `json:"token"`
			Denom *struct {
				Denom string `json:"denom"`
			} `json:"native_token"`
		} `json:"info"`
		Amount string `json:"amount"`
	} `json:"assets"`
	TotalShare string `json:"total_share"`
}

package xpla

type unmarshalable interface {
	TokensRes | IbcsRes
}

type TokensRes struct {
	Mainnet TokenResMap `json:"mainnet"`
	Testnet TokenResMap `json:"testnet"`
}

type TokenResMap map[string]TokenRes

type TokenRes struct {
	Protocol string `json:"protocol"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Token    string `json:"token"`
	Icon     string `json:"icon"`
	Decimals uint8  `json:"decimals"`
}

type IbcsRes struct {
	Mainnet IbcResMap `json:"mainnet"`
	Testnet IbcResMap `json:"testnet"`
}

type IbcResMap map[string]IbcRes

type IbcRes struct {
	Denom     string `json:"denom"`
	Path      string `json:"path"`
	BaseDenom string `json:"base_denom"`
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
}

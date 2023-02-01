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
	Protocol *string `json:"protocol,omitempty"`
	Symbol   *string `json:"symbol,omitempty"`
	Name     *string `json:"name,omitempty"`
	Token    *string `json:"token,omitempty"`
	Icon     *string `json:"icon,omitempty"`
	Decimals *uint8  `json:"decimals,omitempty"`
}

type IbcsRes struct {
	Mainnet IbcResMap `json:"mainnet"`
	Testnet IbcResMap `json:"testnet"`
}

type IbcResMap map[string]IbcRes

type IbcRes struct {
	Denom     *string `json:"denom,omitempty"`
	Path      *string `json:"path,omitempty"`
	BaseDenom *string `json:"base_denom,omitempty"`
	Symbol    *string `json:"symbol,omitempty"`
	Name      *string `json:"name,omitempty"`
	Icon      *string `json:"icon,omitempty"`
}

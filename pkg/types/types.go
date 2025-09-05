package types

type Unmarshalable interface {
	TokensRes | IbcsRes | AssetsRes
}

type TokenType int

const (
	TokenTypeCW20 TokenType = 0 + iota
	TokenTypeERC20
)

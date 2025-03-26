package types

type Unmarshalable interface {
	TokensRes | IbcsRes | AssetsRes
}

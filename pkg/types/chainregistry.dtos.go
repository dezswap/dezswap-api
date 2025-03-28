package types

type AssetsRes struct {
	Assets AssetResList `json:"assets"`
}

type AssetResList []AssetRes

type AssetRes struct {
	DenomUnits []DenomUnit       `json:"denom_units,omitempty"`
	Base       *string           `json:"base,omitempty"`
	Name       *string           `json:"name,omitempty"`
	Symbol     *string           `json:"symbol,omitempty"`
	Address    *string           `json:"address,omitempty"`
	LogoUris   map[string]string `json:"logo_URIs,omitempty"`
	Icon       *string           `json:"icon,omitempty"`
	TypeAsset  *string           `json:"type_asset,omitempty"`
	Traces     []Trace           `json:"traces,omitempty"`
}

type DenomUnit struct {
	Denom    *string `json:"denom"`
	Exponent uint8   `json:"exponent"`
}

type Trace struct {
	Type         *string         `json:"type,omitempty"`
	CounterParty CounterPartyRes `json:"counterparty,omitempty"`
	Chain        ChainRes        `json:"chain,omitempty"`
}

type CounterPartyRes struct {
	ChainName *string `json:"chain_name,omitempty"`
	BaseDenom *string `json:"base_denom,omitempty"`
	ChannelId *string `json:"channel_id,omitempty"`
}

type ChainRes struct {
	ChannelId *string `json:"channel_id,omitempty"`
	Path      *string `json:"path,omitempty"`
}

type AssetType string

const (
	AssetTypeCw20  AssetType = "cw20"
	AssetTypeIcs20 AssetType = "ics20"
)

type TraceType string

const (
	TraceTypeIbc     TraceType = "ibc"
	TraceTypeIbcCw20 TraceType = "ibc-cw20"
)

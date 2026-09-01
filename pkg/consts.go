package pkg

import "github.com/dezswap/dezswap-api/pkg/types"

// NetworkName depends on https://github.com/cosmos/chain-registry
type NetworkName string

const (
	NetworkNameXplaChain    = "xpla"
	NetworkNameAsiAlliance  = "fetchhub"
	NetworkNameTerraClassic = "terra"
	NetworkNameTerra2       = "terra2"
)

const (
	IBC_PREFIX                 = "ibc/"
	IBC_DEFAULT_TOKEN_DECIMALS = 6
)

var (
	networkMetadataList = []NetworkMetadata{
		NewNetworkMetadata(NetworkMetadataConfig{
			NetworkName:   NetworkNameXplaChain,
			AddrPrefix:    "xpla1",
			TokenPrefixes: map[types.TokenType]string{types.TokenTypeCW20: "xcw20:", types.TokenTypeERC20: "xerc20:"},
			BlockSecond:   5,
			Mainnet:       ChainInfo{ChainIdPrefix: "dimension", FactoryAddress: "xpla1j33xdql0h4kpgj2mhggy4vutw655u90z7nyj4afhxgj4v5urtadq44e3vd"},
			Testnets: []ChainInfo{
				{ChainIdPrefix: "cube", FactoryAddress: "xpla1j4kgjl6h4rt96uddtzdxdu39h0mhn4vrtydufdrk4uxxnrpsnw2qug2yx2"},
			},
		}),
		NewNetworkMetadata(NetworkMetadataConfig{
			NetworkName:   NetworkNameAsiAlliance,
			AddrPrefix:    "fetch1",
			TokenPrefixes: map[types.TokenType]string{},
			BlockSecond:   5,
			Mainnet:       ChainInfo{ChainIdPrefix: "fetchhub", FactoryAddress: "fetch1slz6c85kxp4ek5ufmcakfhnscv9r2snlemxgwz6cjhklgh7v2hms8rgt5v"},
			Testnets: []ChainInfo{
				{ChainIdPrefix: "dorado", FactoryAddress: "fetch1kmag3937lrl6dtsv29mlfsedzngl9egv5c3apnr468q50gu04zrqea398u"},
			},
		}),
	}
)

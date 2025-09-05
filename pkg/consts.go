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
		NewNetworkMetadata(
			NetworkNameXplaChain,
			"dimension",
			"cube",
			"xpla1",
			map[types.TokenType]string{types.TokenTypeCW20: "xcw20:", types.TokenTypeERC20: "xerc20:"},
			5,
			0),
		NewNetworkMetadata(
			NetworkNameAsiAlliance,
			"fetchhub",
			"dorado",
			"fetch1",
			map[types.TokenType]string{},
			5,
			0),
	}
)

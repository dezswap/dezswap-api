package pkg

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	"strings"
)

const (
	IBC_PREFIX                 = "ibc/"
	IBC_DEFAULT_TOKEN_DECIMALS = 6
)

type NetworkMetadata struct {
	NetworkName           NetworkName
	mainnetPrefix         string
	testnetPrefix         string
	addrPrefix            string
	BlockSecond           uint8
	LatestHeightIndicator uint64
}

func NewNetworkMetadata(
	networkName NetworkName, mainnetPrefix string, testnetPrefix string, addrPrefix string, blockSecond uint8, latestHeightIndicator uint64) NetworkMetadata {
	return NetworkMetadata{
		networkName,
		mainnetPrefix,
		testnetPrefix,
		addrPrefix,
		blockSecond,
		latestHeightIndicator,
	}
}

func (i NetworkMetadata) IsMainnet(chainId string) bool {
	return strings.Contains(chainId, i.mainnetPrefix)
}

func (i NetworkMetadata) IsTestnet(chainId string) bool {
	return strings.Contains(chainId, i.testnetPrefix)
}

func (i NetworkMetadata) IsMainnetOrTestnet(chainId string) bool {
	return i.IsMainnet(chainId) || i.IsTestnet(chainId)
}

func (i NetworkMetadata) IsCw20(addr string) bool {
	return strings.HasPrefix(addr, i.addrPrefix)
}

func (i NetworkMetadata) IsIbcToken(addr string) bool {
	return strings.HasPrefix(addr, IBC_PREFIX)
}

func NewDecFromStrWithTruncate(input string) (types.Dec, error) {
	truncatedInput := truncateDecimal(input)
	dec, err := types.NewDecFromStr(truncatedInput)
	if err != nil {
		return types.Dec{}, fmt.Errorf("failed to parse decimal: %w", err)
	}

	return dec, nil
}

func truncateDecimal(input string) string {
	parts := strings.Split(input, ".")
	if len(parts) == 1 {
		return input
	}

	fractional := parts[1]
	if len(fractional) > types.Precision {
		fractional = fractional[:types.Precision]
	}

	return parts[0] + "." + fractional
}

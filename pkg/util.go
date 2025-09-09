package pkg

import (
	"cosmossdk.io/math"
	"fmt"
	"github.com/pkg/errors"
	"strings"
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

func GetNetworkMetadata(chainId string) (NetworkMetadata, error) {
	for _, nm := range networkMetadataList {
		if nm.IsMainnetOrTestnet(chainId) {
			return nm, nil
		}
	}

	return NetworkMetadata{}, errors.New("unsupported network")
}

func NewDecFromStrWithTruncate(input string) (math.LegacyDec, error) {
	truncatedInput := truncateDecimal(input)

	dec, err := math.LegacyNewDecFromStr(truncatedInput)
	if err != nil {
		return math.LegacyDec{}, fmt.Errorf("failed to parse decimal: %w", err)
	}

	return dec, nil
}

func truncateDecimal(input string) string {
	parts := strings.Split(input, ".")
	if len(parts) == 1 {
		return input
	}

	fractional := parts[1]
	if len(fractional) > math.LegacyPrecision {
		fractional = fractional[:math.LegacyPrecision]
	}

	return parts[0] + "." + fractional
}

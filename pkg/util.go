package pkg

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	"github.com/dezswap/dezswap-api/pkg/types"
	"github.com/pkg/errors"
)

var (
	ErrUnregisteredFactoryAddress = errors.New("unregistered factory address")
)

type NetworkMetadata struct {
	NetworkName           NetworkName
	mainnetPrefix         string
	testnetPrefix         string
	addrPrefix            string
	tokenPrefixes         map[types.TokenType]string
	BlockSecond           uint8
	LatestHeightIndicator uint64
	factoryAddresses      map[string]string
}

func NewNetworkMetadata(
	networkName NetworkName, mainnetPrefix string, testnetPrefix string, addrPrefix string, tokenPrefixes map[types.TokenType]string, blockSecond uint8, latestHeightIndicator uint64,
	mainnetFactoryAddress, testnetFactoryAddress string) NetworkMetadata {
	return NetworkMetadata{
		networkName,
		mainnetPrefix,
		testnetPrefix,
		addrPrefix,
		tokenPrefixes,
		blockSecond,
		latestHeightIndicator,
		map[string]string{
			mainnetPrefix: mainnetFactoryAddress,
			testnetPrefix: testnetFactoryAddress,
		},
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

func (i NetworkMetadata) GetFactoryAddress(chainId string) (string, error) {
	if i.IsMainnet(chainId) {
		return i.factoryAddresses[i.mainnetPrefix], nil
	}
	if i.IsTestnet(chainId) {
		return i.factoryAddresses[i.testnetPrefix], nil
	}

	return "", errors.New("unsupported network")
}

func (i NetworkMetadata) IsCw20(addr string) bool {
	if prefix, ok := i.tokenPrefixes[types.TokenTypeCW20]; ok {
		addr, _ = strings.CutPrefix(addr, prefix)
	}

	return strings.HasPrefix(addr, i.addrPrefix)
}

func (i NetworkMetadata) IsErc20(addr string) bool {
	if prefix, ok := i.tokenPrefixes[types.TokenTypeERC20]; ok {
		return strings.HasPrefix(addr, prefix)
	}

	return false
}

func (i NetworkMetadata) PrependErc20Prefix(addr string) string {
	if prefix, ok := i.tokenPrefixes[types.TokenTypeERC20]; ok {
		return prefix + addr
	}

	return addr
}

func (i NetworkMetadata) TrimDenomPrefix(addr string) string {
	for _, prefix := range i.tokenPrefixes {
		if addr, found := strings.CutPrefix(addr, prefix); found {
			return addr
		}
	}

	return addr
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

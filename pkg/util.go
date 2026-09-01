package pkg

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	"github.com/dezswap/dezswap-api/pkg/types"
	"github.com/pkg/errors"
)

var (
	ErrUnsupportedNetwork         = errors.New("unsupported network")
	ErrUnregisteredFactoryAddress = errors.New("unregistered factory address")
)

// ChainInfo is a single chain of a network, identified by the prefix of its chain ID.
type ChainInfo struct {
	ChainIdPrefix  string
	FactoryAddress string
}

type NetworkMetadata struct {
	NetworkName           NetworkName
	addrPrefix            string
	tokenPrefixes         map[types.TokenType]string
	BlockSecond           uint8
	LatestHeightIndicator uint64
	mainnet               ChainInfo
	testnets              []ChainInfo
}

// NetworkMetadataConfig registers the chains of a network. At most one of them, Mainnet,
// is classified as the mainnet; every other registered chain is classified as a testnet.
// A chain whose ChainIdPrefix is empty cannot be matched against any chain ID, so it is
// left unregistered.
type NetworkMetadataConfig struct {
	NetworkName           NetworkName
	AddrPrefix            string
	TokenPrefixes         map[types.TokenType]string
	BlockSecond           uint8
	LatestHeightIndicator uint64
	Mainnet               ChainInfo
	Testnets              []ChainInfo
}

func NewNetworkMetadata(c NetworkMetadataConfig) NetworkMetadata {
	metadata := NetworkMetadata{
		NetworkName:           c.NetworkName,
		addrPrefix:            c.AddrPrefix,
		tokenPrefixes:         c.TokenPrefixes,
		BlockSecond:           c.BlockSecond,
		LatestHeightIndicator: c.LatestHeightIndicator,
		mainnet:               c.Mainnet,
	}

	for _, testnet := range c.Testnets {
		if testnet.ChainIdPrefix == "" {
			continue
		}
		metadata.testnets = append(metadata.testnets, testnet)
	}

	return metadata
}

// IsMainnet reports whether chainId belongs to the network's mainnet.
func (i NetworkMetadata) IsMainnet(chainId string) bool {
	return i.mainnet.ChainIdPrefix != "" && strings.HasPrefix(chainId, i.mainnet.ChainIdPrefix)
}

// IsTestnet reports whether chainId belongs to any of the network's testnets.
func (i NetworkMetadata) IsTestnet(chainId string) bool {
	_, found := i.testnet(chainId)
	return found
}

func (i NetworkMetadata) IsMainnetOrTestnet(chainId string) bool {
	return i.IsMainnet(chainId) || i.IsTestnet(chainId)
}

func (i NetworkMetadata) GetFactoryAddress(chainId string) (string, error) {
	if i.IsMainnet(chainId) {
		return i.mainnet.FactoryAddress, nil
	}
	if testnet, found := i.testnet(chainId); found {
		return testnet.FactoryAddress, nil
	}

	return "", ErrUnsupportedNetwork
}

func (i NetworkMetadata) testnet(chainId string) (ChainInfo, bool) {
	for _, testnet := range i.testnets {
		if strings.HasPrefix(chainId, testnet.ChainIdPrefix) {
			return testnet, true
		}
	}

	return ChainInfo{}, false
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

	return NetworkMetadata{}, ErrUnsupportedNetwork
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

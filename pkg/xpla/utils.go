package xpla

import "strings"

func IsMainnet(chainId string) bool {
	return strings.Contains(chainId, MAINNET_PREFIX)
}

func IsTestnet(chainId string) bool {
	return strings.Contains(chainId, TESTNET_PREFIX)
}

func IsMainnetOrTestnet(chainId string) bool {
	return IsMainnet(chainId) || IsTestnet(chainId)
}

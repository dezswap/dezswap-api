package xpla

import "strings"

func IsMainnet(chainId string) bool {
	return strings.Contains(chainId, MAINNET_CHAIN_PREFIX)
}

func IsTestnet(chainId string) bool {
	return strings.Contains(chainId, TESTNET_CHAIN_PREFIX)
}

func IsMainnetOrTestnet(chainId string) bool {
	return IsMainnet(chainId) || IsTestnet(chainId)
}

func IsCw20(addr string) bool {
	return strings.HasPrefix(addr, ADDR_PREFIX)
}

func IsIbcToken(addr string) bool {
	return strings.HasPrefix(addr, IBC_PREFIX)
}

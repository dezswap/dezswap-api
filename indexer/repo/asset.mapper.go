package repo

import (
	"fmt"
	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/dezswap/dezswap-api/pkg/types"
)

type assetMapper interface {
	TokenResToTokens(res *types.TokenResMap, chainId string) []indexer.Token
	IbcsResToTokens(es *types.IbcResMap, chainId string) []indexer.Token
}

type assetMapperImpl struct{}

var _ assetMapper = &assetMapperImpl{}

// TokenResToTokens implements assetMapper
func (*assetMapperImpl) TokenResToTokens(res *types.TokenResMap, chainId string) []indexer.Token {
	tokens := []indexer.Token{}
	for k, v := range *res {
		token := indexer.Token{
			Address:  k,
			ChainId:  chainId,
			Verified: true,
		}
		if v.Protocol != nil {
			token.Protocol = *v.Protocol
		}
		if v.Symbol != nil {
			token.Symbol = *v.Symbol
		}
		if v.Name != nil {
			token.Name = *v.Name
		}
		if v.Decimals != nil {
			token.Decimals = *v.Decimals
		}
		if v.Icon != nil {
			token.Icon = *v.Icon
		}
		tokens = append(tokens, token)
	}
	return tokens
}

// IbcsResToTokens implements assetMapper
func (*assetMapperImpl) IbcsResToTokens(res *types.IbcResMap, chainId string) []indexer.Token {
	tokens := []indexer.Token{}
	for k, v := range *res {
		token := indexer.Token{
			Address:  fmt.Sprintf("ibc/%s", k),
			ChainId:  chainId,
			Decimals: pkg.IBC_DEFAULT_TOKEN_DECIMALS,
			Verified: true,
		}
		if v.Icon != nil {
			token.Icon = *v.Icon
		}
		if v.Name != nil {
			token.Name = *v.Name
		}
		if v.Symbol != nil {
			token.Symbol = *v.Symbol
		}
		if v.Decimals != nil {
			token.Decimals = *v.Decimals
		}
		tokens = append(tokens, token)
	}
	return tokens
}

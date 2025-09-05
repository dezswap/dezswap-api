package repo

import (
	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/dezswap/dezswap-api/pkg/chainregistry"
	"github.com/dezswap/dezswap-api/pkg/types"
	"github.com/dezswap/dezswap-api/pkg/xpla"
	"github.com/pkg/errors"
	"strings"
)

const EthHexPrefix = "0x"

type assetRepoImpl struct {
	pkg.Client
	assetMapper
	pkg.NetworkMetadata
}

var _ indexer.AssetRepo = &assetRepoImpl{}

func NewAssetRepo(networkMetadata pkg.NetworkMetadata, chainId string) (indexer.AssetRepo, error) {
	var client pkg.Client

	switch networkMetadata.NetworkName {
	case pkg.NetworkNameXplaChain:
		client = xpla.NewClient()
	case pkg.NetworkNameAsiAlliance:
		var err error
		client, err = chainregistry.NewClient(chainId)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported network")
	}

	return &assetRepoImpl{client, &assetMapperImpl{}, networkMetadata}, nil
}

// VerifiedTokens implements indexer.AssetRepo
func (r *assetRepoImpl) VerifiedTokens(chainId string) ([]indexer.Token, error) {
	if !r.IsMainnetOrTestnet(chainId) {
		return nil, errors.New("assetRepo.VerifiedTokens: invalid chainId")
	}
	isMainnet := r.IsMainnet(chainId)

	cw20s, err := r.VerifiedCw20s()
	if err != nil {
		return nil, errors.Wrap(err, "assetRepo.VerifiedTokens")
	}
	var tokens []indexer.Token
	if isMainnet {
		tokens = r.TokenResToTokens(&cw20s.Mainnet, chainId)
	} else {
		tokens = r.TokenResToTokens(&cw20s.Testnet, chainId)
	}

	ibcs, err := r.VerifiedIbcs()
	if err != nil {
		return nil, errors.Wrap(err, "assetRepo.VerifiedTokens")
	}

	if isMainnet {
		tokens = append(tokens, r.IbcsResToTokens(&ibcs.Mainnet, chainId)...)
	} else {
		tokens = append(tokens, r.IbcsResToTokens(&ibcs.Testnet, chainId)...)
	}

	erc20s, err := r.VerifiedErc20s()
	if err != nil {
		return nil, errors.Wrap(err, "assetRepo.VerifiedTokens")
	}

	var networkTokens *types.TokenResMap
	if isMainnet {
		networkTokens = r.convertErc20Addr(erc20s.Mainnet)
	} else {
		networkTokens = r.convertErc20Addr(erc20s.Testnet)
	}
	tokens = append(tokens, r.TokenResToTokens(networkTokens, chainId)...)

	return tokens, nil
}

func (r *assetRepoImpl) convertErc20Addr(tokens types.TokenResMap) *types.TokenResMap {
	convertedTokens := make(types.TokenResMap)

	for k, v := range tokens {
		k = strings.TrimPrefix(k, EthHexPrefix)
		convertedTokens["xerc20:"+k] = v
	}

	return &convertedTokens
}

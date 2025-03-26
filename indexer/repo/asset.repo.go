package repo

import (
	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/pkg/errors"
)

type assetRepoImpl struct {
	pkg.Client
	assetMapper
	pkg.NetworkMetadata
}

var _ indexer.AssetRepo = &assetRepoImpl{}

func NewAssetRepo(client pkg.Client, networkMetadata pkg.NetworkMetadata) indexer.AssetRepo {
	return &assetRepoImpl{client, &assetMapperImpl{}, networkMetadata}
}

// VerifiedTokens implements indexer.AssetRepo
func (r *assetRepoImpl) VerifiedTokens(chainId string) ([]indexer.Token, error) {
	if !r.NetworkMetadata.IsMainnetOrTestnet(chainId) {
		return nil, errors.New("assetRepo.VerifiedTokens: invalid chainId")
	}
	isMainnet := r.NetworkMetadata.IsMainnet(chainId)

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
	return tokens, nil
}

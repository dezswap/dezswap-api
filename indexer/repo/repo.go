package repo

import (
	"github.com/dezswap/dezswap-api/indexer"
)

type repoImpl struct {
	*assetRepoImpl
	*dbRepoImpl
	*nodeRepoImpl
}

var _ indexer.Repo = &repoImpl{}

func NewRepo(nodeRepo indexer.NodeRepo, dbRepo indexer.DbRepo, assetRepo indexer.AssetRepo) indexer.Repo {
	var ar *assetRepoImpl
	if assetRepo != nil {
		ar = assetRepo.(*assetRepoImpl)
	}
	return &repoImpl{ar, dbRepo.(*dbRepoImpl), nodeRepo.(*nodeRepoImpl)}
}

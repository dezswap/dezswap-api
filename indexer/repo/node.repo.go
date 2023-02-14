package repo

import (
	"github.com/dezswap/dezswap-api/configs"
	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg/dezswap"
	"github.com/dezswap/dezswap-api/pkg/xpla"
	"github.com/pkg/errors"
)

type nodeRepoImpl struct {
	xpla.GrpcClient
	nodeMapper
	chainId string
}

var _ indexer.NodeRepo = &nodeRepoImpl{}

func NewNodeRepo(client xpla.GrpcClient, c configs.IndexerConfig) (indexer.NodeRepo, error) {
	return &nodeRepoImpl{client, &nodeMapperImpl{}, c.ChainId}, nil
}

// LatestHeightFromNode implements NodeRepo
func (r *nodeRepoImpl) LatestHeightFromNode() (uint64, error) {
	height, err := r.SyncedHeight()
	if err != nil {
		return 0, errors.Wrap(err, "nodeRepoImpl.HeightFromNode")
	}
	return height, nil
}

// PoolFromNode implements NodeRepo
func (r *nodeRepoImpl) PoolFromNode(addr string, height uint64) (*indexer.PoolInfo, error) {
	res, err := r.QueryContract(addr, dezswap.QUERY_POOL, height)
	if err != nil {
		return nil, errors.Wrap(err, "nodeRepoImpl.PoolFromNode")
	}

	poolInfo, err := r.resToPoolInfo(addr, height, res)
	if err != nil {
		return nil, errors.Wrap(err, "nodeRepoImpl.PoolFromNode")
	}
	return poolInfo, nil
}

// TokenFromNode implements NodeRepo
func (r *nodeRepoImpl) TokenFromNode(addr string) (*indexer.Token, error) {
	res, err := r.QueryContract(addr, dezswap.QUERY_TOKEN, xpla.LATEST_HEIGHT_INDICATOR)
	if err != nil {
		return nil, errors.Wrap(err, "nodeRepoImpl.TokenFromNode")
	}

	token, err := r.resToToken(res)
	if err != nil {
		return nil, errors.Wrap(err, "nodeRepoImpl.TokenFromNode")
	}
	return token, nil
}

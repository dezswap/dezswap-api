package repo

import (
	"github.com/dezswap/dezswap-api/configs"
	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/dezswap/dezswap-api/pkg/dezswap"
	"github.com/dezswap/dezswap-api/pkg/xpla"
	"github.com/pkg/errors"
)

type nodeRepoImpl struct {
	pkg.GrpcClient
	nodeMapper
	chainId string
}

var _ indexer.NodeRepo = &nodeRepoImpl{}

func NewNodeRepo(client pkg.GrpcClient, c configs.IndexerConfig) (indexer.NodeRepo, error) {
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

	poolInfo, err := r.resToPoolInfo(addr, r.chainId, height, res)
	if err != nil {
		return nil, errors.Wrap(err, "nodeRepoImpl.PoolFromNode")
	}
	return poolInfo, nil
}

// TokenFromNode implements NodeRepo
func (r *nodeRepoImpl) TokenFromNode(addr string) (*indexer.Token, error) {
	var token *indexer.Token
	var err error

	if xpla.IsIbcToken(addr) {
		token, err = r.ibcFromNode(addr)
	} else if xpla.IsCw20(addr) {
		token, err = r.cw20FromNode(addr)
	} else {
		// currently, query denom is supported (no metadata)
		token, err = r.denomFromNode(addr)
	}

	if err != nil {
		return nil, errors.Wrap(err, "TokenFromNode")
	}

	return token, nil
}

func (r *nodeRepoImpl) denomFromNode(addr string) (*indexer.Token, error) {
	return nil, errors.New("metadata of denom is not supported")
}

func (r *nodeRepoImpl) ibcFromNode(addr string) (*indexer.Token, error) {
	trace, err := r.QueryIbcDenomTrace(addr)
	if err != nil {
		return nil, errors.Wrap(err, "nodeRepoImpl.cw20FromNode")
	}

	if trace == nil {
		return nil, errors.New("denom trace is nil")
	}

	token, err := r.denomTraceToToken(addr, r.chainId, trace)
	if err != nil {
		return nil, errors.Wrap(err, "nodeRepoImpl.cw20FromNode")
	}
	return token, nil
}

func (r *nodeRepoImpl) cw20FromNode(addr string) (*indexer.Token, error) {
	res, err := r.QueryContract(addr, dezswap.QUERY_TOKEN, xpla.LATEST_HEIGHT_INDICATOR)
	if err != nil {
		return nil, errors.Wrap(err, "nodeRepoImpl.cw20FromNode")
	}

	token, err := r.resToToken(addr, r.chainId, res)
	token.Address = addr
	token.ChainId = r.chainId
	if err != nil {
		return nil, errors.Wrap(err, "nodeRepoImpl.cw20FromNode")
	}
	return token, nil
}

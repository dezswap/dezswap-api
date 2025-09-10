package repo

import (
	"context"
	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/dezswap/dezswap-api/pkg/dezswap"
	"github.com/pkg/errors"
	"time"
)

const queryTimeout = 5 * time.Second

type nodeRepoImpl struct {
	pkg.EthClient
	pkg.GrpcClient
	nodeMapper
	pkg.NetworkMetadata
	chainId string
}

var _ indexer.NodeRepo = &nodeRepoImpl{}

func NewNodeRepo(grpcEndpoint, ethRpcEndpoint string, useTls bool, chainId string, networkMetadata pkg.NetworkMetadata) (indexer.NodeRepo, error) {
	grpcClient, err := pkg.NewGrpcClient(grpcEndpoint, useTls)
	if err != nil {
		return nil, err
	}

	var ethClient pkg.EthClient
	if ethRpcEndpoint != "" {
		ethClient, err = pkg.NewEthClient(ethRpcEndpoint)
		if err != nil {
			return nil, err
		}
	}

	return &nodeRepoImpl{
		ethClient,
		grpcClient,
		&nodeMapperImpl{},
		networkMetadata,
		chainId,
	}, nil
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

	if r.IsIbcToken(addr) {
		token, err = r.ibcFromNode(addr)
	} else if r.IsCw20(addr) {
		token, err = r.cw20FromNode(addr)
	} else if r.IsErc20(addr) {
		token, err = r.erc20FromNode(addr)
	} else {
		// currently, query denom is not supported (no metadata)
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
	trimmedAddr := r.TrimDenomPrefix(addr) // in case of xcw20: address

	res, err := r.QueryContract(trimmedAddr, dezswap.QUERY_TOKEN, r.LatestHeightIndicator)
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

func (r *nodeRepoImpl) erc20FromNode(addr string) (*indexer.Token, error) {
	if r.EthClient == nil {
		return nil, errors.Errorf("New ETH address %s found but no rpc client supported on indexer.", addr)
	}

	trimmedAddr := r.TrimDenomPrefix(addr)
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	erc20Meta, err := r.QueryErc20Info(ctx, trimmedAddr)
	if err != nil {
		return nil, errors.Wrap(err, "nodeRepoImpl.erc20FromNode")
	}

	return &indexer.Token{
		Address:  addr,
		ChainId:  r.chainId,
		Symbol:   erc20Meta.Symbol,
		Name:     erc20Meta.Name,
		Decimals: erc20Meta.Decimals,
	}, nil
}

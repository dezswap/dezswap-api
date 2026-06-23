package repo

import (
	"context"

	ibc_types "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/dezswap/dezswap-api/pkg/dezswap"
	"github.com/pkg/errors"
)

type nodeRepoImpl struct {
	pkg.EthClient
	grpcClients []pkg.GrpcClient
	nodeMapper
	pkg.NetworkMetadata
	chainId string
}

var _ indexer.NodeRepo = &nodeRepoImpl{}

type GrpcEndpoint struct {
	Target string
	UseTLS bool
}

func NewNodeRepo(grpcEndpoint, ethRpcEndpoint string, useTls bool, chainId string, networkMetadata pkg.NetworkMetadata) (indexer.NodeRepo, error) {
	return NewNodeRepoWithGrpcEndpoints([]GrpcEndpoint{{Target: grpcEndpoint, UseTLS: useTls}}, ethRpcEndpoint, chainId, networkMetadata)
}

func NewNodeRepoWithGrpcEndpoints(grpcEndpoints []GrpcEndpoint, ethRpcEndpoint string, chainId string, networkMetadata pkg.NetworkMetadata) (indexer.NodeRepo, error) {
	if len(grpcEndpoints) == 0 {
		return nil, errors.New("NewNodeRepoWithGrpcEndpoints: grpc endpoint is not configured")
	}

	grpcClients := make([]pkg.GrpcClient, 0, len(grpcEndpoints))
	for i, endpoint := range grpcEndpoints {
		grpcClient, err := pkg.NewGrpcClient(endpoint.Target, endpoint.UseTLS)
		if err != nil {
			return nil, errors.Wrapf(err, "NewNodeRepoWithGrpcEndpoints: failed to create grpc client endpoint[%d]=%s", i, endpoint.Target)
		}
		grpcClients = append(grpcClients, grpcClient)
	}

	var ethClient pkg.EthClient
	var err error
	if ethRpcEndpoint != "" {
		ethClient, err = pkg.NewEthClient(ethRpcEndpoint)
		if err != nil {
			return nil, errors.Wrapf(err, "NewNodeRepoWithGrpcEndpoints: failed to create eth client endpoint=%s", ethRpcEndpoint)
		}
	}

	return &nodeRepoImpl{
		EthClient:       ethClient,
		grpcClients:     grpcClients,
		nodeMapper:      &nodeMapperImpl{},
		NetworkMetadata: networkMetadata,
		chainId:         chainId,
	}, nil
}

// LatestHeightFromNode implements NodeRepo
func (r *nodeRepoImpl) LatestHeightFromNode() (uint64, error) {
	var lastErr error
	for _, client := range r.grpcClients {
		height, err := client.SyncedHeight()
		if err == nil {
			return height, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return 0, errors.Wrap(lastErr, "nodeRepoImpl.HeightFromNode")
	}
	return 0, errors.New("nodeRepoImpl.HeightFromNode: grpc client is not configured")
}

// PoolFromNode implements NodeRepo
func (r *nodeRepoImpl) PoolFromNode(addr string, height uint64) (*indexer.PoolInfo, error) {
	res, err := r.queryContractFromNode(addr, dezswap.QUERY_POOL, height)
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
	trace, err := r.ibcDenomTraceFromNode(addr)
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

	res, err := r.queryContractFromNode(trimmedAddr, dezswap.QUERY_TOKEN, r.LatestHeightIndicator)
	if err != nil {
		return nil, errors.Wrap(err, "nodeRepoImpl.cw20FromNode")
	}

	token, err := r.resToToken(addr, r.chainId, res)
	if err != nil {
		return nil, errors.Wrap(err, "nodeRepoImpl.cw20FromNode")
	}
	if token == nil {
		return nil, errors.New("token is nil")
	}
	token.Address = addr
	token.ChainId = r.chainId
	return token, nil
}

func (r *nodeRepoImpl) queryContractFromNode(addr string, query []byte, height uint64) ([]byte, error) {
	var lastErr error
	for _, client := range r.grpcClients {
		res, err := client.QueryContract(addr, query, height)
		if err == nil {
			return res, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("nodeRepoImpl.QueryContractFromNode: grpc client is not configured")
}

func (r *nodeRepoImpl) ibcDenomTraceFromNode(addr string) (*ibc_types.Denom, error) {
	var lastErr error
	for _, client := range r.grpcClients {
		trace, err := client.QueryIbcDenomTrace(addr)
		if err == nil {
			return trace, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("nodeRepoImpl.QueryIbcDenomTraceFromNode: grpc client is not configured")
}

func (r *nodeRepoImpl) erc20FromNode(addr string) (*indexer.Token, error) {
	if r.EthClient == nil {
		return nil, errors.Errorf("New ETH address %s found but no rpc client supported on indexer.", addr)
	}

	trimmedAddr := r.TrimDenomPrefix(addr)
	ctx, cancel := context.WithTimeout(context.Background(), pkg.NodeQueryTimeout)
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

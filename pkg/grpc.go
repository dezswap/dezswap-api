package pkg

import (
	"context"
	"fmt"
	"strconv"

	cosmwasm_types "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	cosmos_types "github.com/cosmos/cosmos-sdk/types/grpc"
	ibc_types "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient interface {
	SyncedHeight() (uint64, error)
	QueryContract(addr string, query []byte, height uint64) ([]byte, error)
	QueryIbcDenomTrace(hash string) (*ibc_types.DenomTrace, error)
}

type grpcClient struct {
	*grpc.ClientConn
}

var _ GrpcClient = &grpcClient{}

func NewGrpcClient(target string) (GrpcClient, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrap(err, "NewGrpcClient: failed to dial")
	}

	return &grpcClient{conn}, nil
}

// SyncedHeight implements GrpcClient
func (c *grpcClient) SyncedHeight() (uint64, error) {
	client := tmservice.NewServiceClient(c)

	// Get the latest block height
	res, err := client.GetLatestBlock(context.Background(), &tmservice.GetLatestBlockRequest{})
	if err != nil {
		fmt.Printf("failed to get latest block height: %v\n", err)
		return 0, err
	}

	return uint64(res.Block.Header.Height), nil
}

// QueryContract implements GrpcClient
func (c *grpcClient) QueryContract(addr string, query []byte, height uint64) ([]byte, error) {
	client := cosmwasm_types.NewQueryClient(c)
	ctx := context.Background()
	if height > 0 {
		//nolint:staticcheck
		ctx = context.WithValue(ctx, cosmos_types.GRPCBlockHeightHeader, strconv.FormatUint(height, 10))
	}

	res, err := client.SmartContractState(ctx, &cosmwasm_types.QuerySmartContractStateRequest{Address: addr, QueryData: query})
	if err != nil {
		return nil, errors.Wrapf(err, "QueryContract(%s)", addr)
	}

	return res.Data, nil
}

// QueryIbcDenomTrace implements GrpcClient
func (c *grpcClient) QueryIbcDenomTrace(addr string) (*ibc_types.DenomTrace, error) {
	client := ibc_types.NewQueryClient(c)
	ctx := context.Background()

	res, err := client.DenomTrace(ctx, &ibc_types.QueryDenomTraceRequest{Hash: addr})
	if err != nil {
		return nil, errors.Wrapf(err, "QueryContract(%s)", addr)
	}

	return res.GetDenomTrace(), nil
}

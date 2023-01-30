package xpla

import (
	"context"
	"fmt"
	"strconv"

	cosmwasm_types "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	cosmos_types "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient interface {
	SyncedHeight() (uint64, error)
	QueryContract(addr string, query []byte, height uint64) ([]byte, error)
}

type grpcClient struct {
	*grpc.ClientConn
}

var _ GrpcClient = &grpcClient{}

func NewGrpcClient(host string) (GrpcClient, error) {
	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		ctx = context.WithValue(ctx, cosmos_types.GRPCBlockHeightHeader, strconv.FormatUint(height, 10))
	}

	res, err := client.SmartContractState(ctx, &cosmwasm_types.QuerySmartContractStateRequest{Address: addr, QueryData: query})
	if err != nil {
		return nil, errors.Wrapf(err, "QueryContract(%s)", addr)
	}

	return res.Data, nil
}

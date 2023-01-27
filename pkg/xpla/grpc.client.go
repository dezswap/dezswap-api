package xpla

import (
	"context"
	"strconv"

	cosmwasm_types "github.com/CosmWasm/wasmd/x/wasm/types"
	cosmos_types "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient interface {
	QueryContract(addr string, query []byte, height uint64) ([]byte, error)
}

type grpcClient struct {
	*grpc.ClientConn
}

func NewGrpcClient(host string) (GrpcClient, error) {
	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrap(err, "NewGrpcClient: failed to dial")
	}

	return &grpcClient{conn}, nil
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

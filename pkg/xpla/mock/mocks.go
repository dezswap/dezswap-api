package mock

import (
	ibctypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/dezswap/dezswap-api/pkg/types"
	"github.com/stretchr/testify/mock"
)

type GrpcClientMock struct {
	*mock.Mock
}

var _ pkg.GrpcClient = &GrpcClientMock{}

func NewGrpcClientMock() *GrpcClientMock {
	return &GrpcClientMock{&mock.Mock{}}
}

// QueryContract implements pkg.GrpcClient
func (g *GrpcClientMock) QueryContract(addr string, query []byte, height uint64) ([]byte, error) {
	args := g.MethodCalled("QueryContract", addr, query, height)
	return args.Get(0).([]byte), args.Error(1)
}

// SyncedHeight implements pkg.GrpcClient
func (g *GrpcClientMock) SyncedHeight() (uint64, error) {
	args := g.MethodCalled("SyncedHeight")
	return args.Get(0).(uint64), args.Error(1)
}

// QueryIbcDenomTrace implements pkg.GrpcClient
func (g *GrpcClientMock) QueryIbcDenomTrace(hash string) (*ibctypes.Denom, error) {
	args := g.MethodCalled("QueryIbcDenomTrace")
	return args.Get(0).(*ibctypes.Denom), args.Error(1)
}

type ClientMock struct {
	*mock.Mock
}

var _ pkg.Client = &ClientMock{}

func NewClientMock() *ClientMock {
	return &ClientMock{&mock.Mock{}}
}

func (c ClientMock) VerifiedCw20s() (*types.TokensRes, error) {
	args := c.MethodCalled("VerifiedCw20s")
	return args.Get(0).(*types.TokensRes), args.Error(1)
}

func (c ClientMock) VerifiedIbcs() (*types.IbcsRes, error) {
	args := c.MethodCalled("VerifiedIbcs")
	return args.Get(0).(*types.IbcsRes), args.Error(1)
}

func (c ClientMock) VerifiedErc20s() (*types.TokensRes, error) {
	args := c.MethodCalled("VerifiedErc20s")
	return args.Get(0).(*types.TokensRes), args.Error(1)
}

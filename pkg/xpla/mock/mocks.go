package mock

import (
	ibc_types "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/dezswap/dezswap-api/pkg"
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
func (g *GrpcClientMock) QueryIbcDenomTrace(hash string) (*ibc_types.DenomTrace, error) {
	args := g.MethodCalled("QueryIbcDenomTrace")
	return args.Get(0).(*ibc_types.DenomTrace), args.Error(1)
}

package mock

import (
	"github.com/dezswap/dezswap-api/pkg/xpla"
	"github.com/stretchr/testify/mock"
)

type GrpcClientMock struct {
	*mock.Mock
}

var _ xpla.GrpcClient = &GrpcClientMock{}

func NewGrpcClientMock() *GrpcClientMock {
	return &GrpcClientMock{&mock.Mock{}}
}

// QueryContract implements xpla.GrpcClient
func (g *GrpcClientMock) QueryContract(addr string, query []byte, height uint64) ([]byte, error) {
	args := g.MethodCalled("QueryContract", addr, query, height)
	return args.Get(0).([]byte), args.Error(1)
}

// SyncedHeight implements xpla.GrpcClient
func (g *GrpcClientMock) SyncedHeight() (uint64, error) {
	args := g.MethodCalled("SyncedHeight")
	return args.Get(0).(uint64), args.Error(1)
}

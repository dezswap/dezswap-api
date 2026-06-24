package repo

import (
	"encoding/json"
	"strings"
	"testing"

	ibc_types "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/dezswap/dezswap-api/pkg/dezswap"
	"github.com/dezswap/dezswap-api/pkg/types"
	"github.com/stretchr/testify/mock"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	xpla_mock "github.com/dezswap/dezswap-api/pkg/xpla/mock"
)

type nodeMapperMock struct {
	mock.Mock
}

func (m *nodeMapperMock) resToToken(addr, chainId string, data []byte) (*indexer.Token, error) {
	args := m.Called(addr, chainId, data)
	token, _ := args.Get(0).(*indexer.Token)
	return token, args.Error(1)
}

func (m *nodeMapperMock) resToPoolInfo(addr, chainId string, height uint64, data []byte) (*indexer.PoolInfo, error) {
	args := m.Called(addr, chainId, height, data)
	poolInfo, _ := args.Get(0).(*indexer.PoolInfo)
	return poolInfo, args.Error(1)
}

func (m *nodeMapperMock) denomTraceToToken(addr, chainId string, trace *ibc_types.Denom) (*indexer.Token, error) {
	args := m.Called(addr, chainId, trace)
	token, _ := args.Get(0).(*indexer.Token)
	return token, args.Error(1)
}

type nodeRepoSuite struct {
	suite.Suite
	client          pkg.GrpcClient
	ethClient       pkg.EthClient
	chainId         string
	r               nodeRepoImpl
	networkMetadata pkg.NetworkMetadata
}

func (s *nodeRepoSuite) SetupSuite() {
	s.client = xpla_mock.NewGrpcClientMock()
	s.ethClient = xpla_mock.NewEthClientMock()
	s.chainId = "test"
	s.networkMetadata = pkg.NewNetworkMetadata(
		pkg.NetworkNameXplaChain,
		"dimension",
		"cube",
		"xpla1",
		map[types.TokenType]string{types.TokenTypeCW20: "xcw20:", types.TokenTypeERC20: "xerc20:"},
		5,
		0,
		"",
		"",
	)
	s.r = nodeRepoImpl{
		EthClient:       s.ethClient,
		grpcClients:     []pkg.GrpcClient{s.client},
		nodeMapper:      &nodeMapperImpl{},
		NetworkMetadata: s.networkMetadata,
		chainId:         s.chainId,
	}
}

func (s *nodeRepoSuite) Test_LatestHeightFromNode() {
	tcs := []struct {
		height   uint64
		expected uint64
		err      error
	}{
		{height: 1, expected: 1, err: nil},
		{height: 0, expected: 0, err: errors.New("failed to get latest block height")},
	}

	for _, tc := range tcs {
		s.client.(*xpla_mock.GrpcClientMock).On("SyncedHeight").Return(tc.height, tc.err).Once()
		actual, err := s.r.LatestHeightFromNode()
		if err != nil {
			assert.True(s.T(), strings.Contains(err.Error(), tc.err.Error()))
		} else {
			assert.Equal(s.T(), tc.expected, actual)
		}
	}
}

func (s *nodeRepoSuite) Test_LatestHeightFromNode_SelectsNextClientOnFailure() {
	firstClient := xpla_mock.NewGrpcClientMock()
	secondClient := xpla_mock.NewGrpcClientMock()
	r := nodeRepoImpl{
		grpcClients:     []pkg.GrpcClient{firstClient, secondClient},
		nodeMapper:      &nodeMapperImpl{},
		NetworkMetadata: s.networkMetadata,
		chainId:         s.chainId,
	}

	firstClient.On("SyncedHeight").Return(uint64(0), errors.New("unavailable")).Once()
	secondClient.On("SyncedHeight").Return(uint64(100), nil).Once()

	height, err := r.LatestHeightFromNode()

	s.Require().NoError(err)
	s.Equal(uint64(100), height)
	firstClient.AssertExpectations(s.T())
	secondClient.AssertExpectations(s.T())
}

func (s *nodeRepoSuite) Test_PoolFromNode_SelectsNextClientOnFailure() {
	firstClient := xpla_mock.NewGrpcClientMock()
	secondClient := xpla_mock.NewGrpcClientMock()
	mapperMock := &nodeMapperMock{}
	r := nodeRepoImpl{
		grpcClients:     []pkg.GrpcClient{firstClient, secondClient},
		nodeMapper:      mapperMock,
		NetworkMetadata: s.networkMetadata,
		chainId:         s.chainId,
	}

	const addr = "xpla1pool"
	const height = uint64(100)
	dummyRes := []byte(`{}`)
	expected := &indexer.PoolInfo{Address: addr, Height: height}

	firstClient.On("QueryContract", addr, dezswap.QUERY_POOL, height).Return([]byte(nil), errors.New("unavailable")).Once()
	secondClient.On("QueryContract", addr, dezswap.QUERY_POOL, height).Return(dummyRes, nil).Once()
	mapperMock.On("resToPoolInfo", addr, s.chainId, height, dummyRes).Return(expected, nil).Once()

	actual, err := r.PoolFromNode(addr, height)

	s.Require().NoError(err)
	s.Equal(expected, actual)
	firstClient.AssertExpectations(s.T())
	secondClient.AssertExpectations(s.T())
	mapperMock.AssertExpectations(s.T())
}

func (s *nodeRepoSuite) Test_PoolFromNode_ReturnsLastErrorWhenAllClientsFail() {
	firstClient := xpla_mock.NewGrpcClientMock()
	secondClient := xpla_mock.NewGrpcClientMock()
	mapperMock := &nodeMapperMock{}
	r := nodeRepoImpl{
		grpcClients:     []pkg.GrpcClient{firstClient, secondClient},
		nodeMapper:      mapperMock,
		NetworkMetadata: s.networkMetadata,
		chainId:         s.chainId,
	}

	const addr = "xpla1pool"
	const height = uint64(100)

	firstClient.On("QueryContract", addr, dezswap.QUERY_POOL, height).Return([]byte(nil), errors.New("unavailable")).Once()
	secondClient.On("QueryContract", addr, dezswap.QUERY_POOL, height).Return([]byte(nil), errors.New("connection reset")).Once()

	actual, err := r.PoolFromNode(addr, height)

	s.Require().Error(err)
	s.Nil(actual)
	s.Contains(err.Error(), "connection reset")
	mapperMock.AssertNotCalled(s.T(), "resToPoolInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	firstClient.AssertExpectations(s.T())
	secondClient.AssertExpectations(s.T())
}

func (s *nodeRepoSuite) Test_cw20FromNode_SelectsNextClientOnFailure() {
	firstClient := xpla_mock.NewGrpcClientMock()
	secondClient := xpla_mock.NewGrpcClientMock()
	mapperMock := &nodeMapperMock{}
	r := nodeRepoImpl{
		grpcClients:     []pkg.GrpcClient{firstClient, secondClient},
		nodeMapper:      mapperMock,
		NetworkMetadata: s.networkMetadata,
		chainId:         s.chainId,
	}

	const addr = "xpla1abc"
	dummyRes := []byte(`{}`)
	expected := &indexer.Token{Symbol: "TEST", Name: "Test Token", Decimals: 6}

	firstClient.On("QueryContract", addr, dezswap.QUERY_TOKEN, s.networkMetadata.LatestHeightIndicator).Return([]byte(nil), errors.New("unavailable")).Once()
	secondClient.On("QueryContract", addr, dezswap.QUERY_TOKEN, s.networkMetadata.LatestHeightIndicator).Return(dummyRes, nil).Once()
	mapperMock.On("resToToken", addr, s.chainId, dummyRes).Return(expected, nil).Once()

	token, err := r.cw20FromNode(addr)

	s.Require().NoError(err)
	s.Require().NotNil(token)
	s.Equal(addr, token.Address)
	s.Equal(s.chainId, token.ChainId)
	s.Equal("TEST", token.Symbol)
	firstClient.AssertExpectations(s.T())
	secondClient.AssertExpectations(s.T())
	mapperMock.AssertExpectations(s.T())
}

func (s *nodeRepoSuite) Test_ibcFromNode_SelectsNextClientOnFailure() {
	firstClient := xpla_mock.NewGrpcClientMock()
	secondClient := xpla_mock.NewGrpcClientMock()
	mapperMock := &nodeMapperMock{}
	r := nodeRepoImpl{
		grpcClients:     []pkg.GrpcClient{firstClient, secondClient},
		nodeMapper:      mapperMock,
		NetworkMetadata: s.networkMetadata,
		chainId:         s.chainId,
	}

	const addr = "ibc/ABC"
	trace := &ibc_types.Denom{Base: "uxpla"}
	expected := &indexer.Token{Address: addr, Symbol: "XPLA", Name: "XPLA", Decimals: 18}

	firstClient.On("QueryIbcDenomTrace", addr).Return((*ibc_types.Denom)(nil), errors.New("unavailable")).Once()
	secondClient.On("QueryIbcDenomTrace", addr).Return(trace, nil).Once()
	mapperMock.On("denomTraceToToken", addr, s.chainId, trace).Return(expected, nil).Once()

	token, err := r.ibcFromNode(addr)

	s.Require().NoError(err)
	s.Equal(expected, token)
	firstClient.AssertExpectations(s.T())
	secondClient.AssertExpectations(s.T())
	mapperMock.AssertExpectations(s.T())
}

func (s *nodeRepoSuite) Test_TokenFromNode() {
	tcs := []struct {
		inputAddr   string
		trimmedAddr string
		expected    indexer.Token
		err         error
	}{
		{
			inputAddr:   "xerc20:ABCD",
			trimmedAddr: "ABCD",
			expected: indexer.Token{
				Address:  "xerc20:ABCD",
				ChainId:  s.chainId,
				Symbol:   "ABCD",
				Name:     "Abcd",
				Decimals: 18,
			},
			err: nil,
		},
		{
			inputAddr:   "xpla1efgh",
			trimmedAddr: "xpla1efgh",
			expected: indexer.Token{
				Address:  "xpla1efgh",
				ChainId:  s.chainId,
				Symbol:   "EFGH",
				Name:     "Efgh",
				Decimals: 6,
			},
			err: nil,
		},
	}

	for _, tc := range tcs {
		// prepare mock response
		{
			s.ethClient.(*xpla_mock.EthClientMock).On("QueryErc20Info", mock.Anything, tc.trimmedAddr).Return(
				pkg.ERC20Meta{
					Name:     tc.expected.Name,
					Symbol:   tc.expected.Symbol,
					Decimals: tc.expected.Decimals,
				}, tc.err).Once()

			res, _ := json.Marshal(dezswap.TokenInfoRes{
				Name:     tc.expected.Name,
				Symbol:   tc.expected.Symbol,
				Decimals: uint(tc.expected.Decimals),
			})
			s.client.(*xpla_mock.GrpcClientMock).On("QueryContract", tc.trimmedAddr, dezswap.QUERY_TOKEN, mock.Anything).Return(res, tc.err).Once()
		}

		actual, err := s.r.TokenFromNode(tc.inputAddr)
		if err != nil {
			assert.True(s.T(), strings.Contains(err.Error(), tc.err.Error()))
		} else {
			assert.Equal(s.T(), tc.expected, *actual)
		}
	}
}

func (s *nodeRepoSuite) Test_cw20FromNode() {
	const addr = "xpla1abc"
	dummyRes := []byte(`{}`)

	tcs := []struct {
		name      string
		mapperErr error
		expectErr string
	}{
		{
			name:      "resToToken returns error",
			mapperErr: errors.New("parse error"),
			expectErr: "parse error",
		},
		{
			name:      "resToToken returns nil token",
			mapperErr: nil,
			expectErr: "token is nil",
		},
		{
			name:      "successful token resolution",
			mapperErr: nil,
			expectErr: "",
		},
	}

	for _, tc := range tcs {
		s.Run(tc.name, func() {
			client := xpla_mock.NewGrpcClientMock()
			mapperMock := &nodeMapperMock{}
			r := nodeRepoImpl{
				EthClient:       s.ethClient,
				grpcClients:     []pkg.GrpcClient{client},
				nodeMapper:      mapperMock,
				NetworkMetadata: s.networkMetadata,
				chainId:         s.chainId,
			}

			client.On("QueryContract", addr, dezswap.QUERY_TOKEN, s.networkMetadata.LatestHeightIndicator).Return(dummyRes, nil).Once()

			var mockToken *indexer.Token
			if tc.expectErr == "" {
				mockToken = &indexer.Token{Symbol: "TEST", Name: "Test Token", Decimals: 6}
			}
			mapperMock.On("resToToken", addr, s.chainId, dummyRes).Return(mockToken, tc.mapperErr).Once()

			token, err := r.cw20FromNode(addr)
			if tc.expectErr != "" {
				s.Require().Error(err)
				s.True(strings.Contains(err.Error(), tc.expectErr))
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(token)
				s.Equal(addr, token.Address)
				s.Equal(s.chainId, token.ChainId)
				s.Equal("TEST", token.Symbol)
			}
			client.AssertExpectations(s.T())
			mapperMock.AssertExpectations(s.T())
		})
	}
}

func Test_NodeRepo(t *testing.T) {
	suite.Run(t, new(nodeRepoSuite))
}

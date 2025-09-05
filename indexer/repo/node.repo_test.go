package repo

import (
	"encoding/json"
	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/dezswap/dezswap-api/pkg/dezswap"
	"github.com/dezswap/dezswap-api/pkg/types"
	"github.com/stretchr/testify/mock"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	xpla_mock "github.com/dezswap/dezswap-api/pkg/xpla/mock"
)

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
	)
	s.r = nodeRepoImpl{s.ethClient, s.client, &nodeMapperImpl{}, s.networkMetadata, s.chainId}
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

func Test_NodeRepo(t *testing.T) {
	suite.Run(t, new(nodeRepoSuite))
}

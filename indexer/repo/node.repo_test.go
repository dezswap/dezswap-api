package repo

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/dezswap/dezswap-api/pkg/xpla"
	xpla_mock "github.com/dezswap/dezswap-api/pkg/xpla/mock"
)

type repoSuite struct {
	suite.Suite
	client  xpla.GrpcClient
	chainId string
	r       nodeRepoImpl
}

func (s *repoSuite) SetupSuite() {
	s.client = xpla_mock.NewGrpcClientMock()
	s.chainId = "test"
	s.r = nodeRepoImpl{s.client, &nodeMapperImpl{}, s.chainId}
}

func (s *repoSuite) Test_LatestHeightFromNode() {
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

// func (s *repoSuite) Test_PoolFromNode() {
// 	const (
// 		addr = "xpla1testaddres"
// 	)
// 	tcs := []struct {
// 		poolRes  dezswap.PoolRes
// 		expected indexer.PoolInfo
// 		err      error
// 	}{
// 		{
// 			poolRes: dezswap.PoolRes{TotalShare: "1000",

// 			expected: indexer.PoolInfo{Address: addr, LpAmount: "1000", Asset0Amount: "1000", Asset1Amount: "1000"},
// 			err:      nil,
// 		},
// 		{
// 			poolRes:  dezswap.PoolRes{},
// 			expected: indexer.PoolInfo{},
// 			err:      errors.New("QueryContract"),
// 		},
// 	}

// 	for _, tc := range tcs {
// 		data, _ := json.Marshal(tc.poolRes)
// 		s.client.(*xpla_mock.GrpcClientMock).On("QueryContract", mock.Anything, mock.Anything, mock.Anything).Return(data, tc.err).Once()
// 		actual, err := s.r.PoolFromNode(addr, xpla.LATEST_HEIGHT_INDICATOR)
// 		if err != nil {
// 			assert.True(s.T(), strings.Contains(err.Error(), tc.err.Error()))
// 		} else {
// 			assert.Equal(s.T(), tc.expected, *actual)
// 		}

// 	}
// }

func TestMain(t *testing.T) {
	suite.Run(t, new(repoSuite))
}

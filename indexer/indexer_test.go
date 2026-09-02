package indexer

import (
	"fmt"
	"testing"

	"github.com/dezswap/dezswap-api/pkg"

	"github.com/dezswap/dezswap-api/pkg/db"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	Repo
	*mock.Mock
}

func (m *mockRepo) Tokens(cond db.LastIdLimitCondition) ([]Token, error) {
	args := m.Called(cond)
	return args.Get(0).([]Token), args.Error(1)
}

func (m *mockRepo) Pairs(cond db.LastIdLimitCondition) ([]Pair, error) {
	args := m.Called(cond)
	return args.Get(0).([]Pair), args.Error(1)
}

func (m *mockRepo) TokenAddresses(cond db.LastIdLimitCondition) ([]string, error) {
	args := m.Called(cond)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockRepo) TokenFromNode(addr string) (*Token, error) {
	args := m.Called(addr)
	token, _ := args.Get(0).(*Token)
	return token, args.Error(1)
}

func (m *mockRepo) VerifiedTokens(chainId string) ([]Token, error) {
	args := m.Called(chainId)
	return args.Get(0).([]Token), args.Error(1)
}

func (m *mockRepo) SaveTokens(tokens []Token) error {
	args := m.Called(tokens)
	return args.Error(0)
}

func Test_UpdateVerified(t *testing.T) {
	repo := mockRepo{nil, &mock.Mock{}}
	dexIndexer := dexIndexer{pkg.NetworkMetadata{}, &repo, "chainId"}

	type testcase struct {
		tokens                  []Token
		verifiedTokens          []Token
		expectedUpdatableTokens []Token
		err                     string
	}
	tests := []testcase{
		{
			[]Token{{Address: "0x1", ChainId: "chainId", Protocol: "protocol", Symbol: "symbol", Name: "name", Decimals: 18, Icon: "icon", Verified: false}},
			[]Token{{Address: "0x1", ChainId: "chainId", Protocol: "protocol", Symbol: "symbol", Name: "name", Decimals: 18, Icon: "icon", Verified: true}},
			[]Token{{Address: "0x1", ChainId: "chainId", Protocol: "protocol", Symbol: "symbol", Name: "name", Decimals: 18, Icon: "icon", Verified: true}},
			"",
		},
		// verified tokens must be removed if it is not in verifiedTokens
		{
			[]Token{{Address: "0x1", ChainId: "chainId", Protocol: "protocol", Symbol: "symbol", Name: "name", Decimals: 18, Icon: "icon", Verified: true}},
			[]Token{},
			[]Token{{Address: "0x1", ChainId: "chainId", Protocol: "protocol", Symbol: "symbol", Name: "name", Decimals: 18, Icon: "icon", Verified: false}},
			"",
		},
		// verified tokens must be updated if it changed the value
		{
			[]Token{{Address: "0x1", ChainId: "chainId", Protocol: "protocol", Symbol: "SYMBOL", Name: "name", Decimals: 18, Icon: "icon", Verified: true}},
			[]Token{{Address: "0x1", ChainId: "chainId", Protocol: "protocol", Symbol: "symbol", Name: "name", Decimals: 18, Icon: "icon", Verified: true}},
			[]Token{{Address: "0x1", ChainId: "chainId", Protocol: "protocol", Symbol: "symbol", Name: "name", Decimals: 18, Icon: "icon", Verified: true}},
			"",
		},
	}

	assert := assert.New(t)
	for _, test := range tests {
		repo.On("Tokens", db.LastIdLimitCondition{}).Return(test.tokens, nil).Once()
		repo.On("VerifiedTokens", "chainId").Return(test.verifiedTokens, nil).Once()
		repo.On("SaveTokens", test.expectedUpdatableTokens).Return(nil).Once()

		err := dexIndexer.UpdateVerifiedTokens()
		if test.err != "" {
			assert.NotNil(err)
			assert.Equal(test.err, err.Error())
		} else {
			assert.Nil(err)
		}
	}
}

func Test_UpdateTokens_SkipsAddressesAlreadyStored(t *testing.T) {
	repo := mockRepo{nil, &mock.Mock{}}
	dexIndexer := dexIndexer{pkg.NetworkMetadata{}, &repo, "chainId"}

	repo.On("Pairs", db.LastIdLimitCondition{}).
		Return([]Pair{{Address: "pair0", Asset0: "stored", Asset1: "new", Lp: "lp0"}}, nil).Once()
	repo.On("TokenAddresses", db.LastIdLimitCondition{}).Return([]string{"stored"}, nil).Once()
	repo.On("TokenFromNode", "new").Return(&Token{Address: "new"}, nil).Once()
	repo.On("TokenFromNode", "lp0").Return(&Token{Address: "lp0"}, nil).Once()
	repo.On("SaveTokens", []Token{{Address: "new"}, {Address: "lp0"}}).Return(nil).Once()

	require.NoError(t, dexIndexer.UpdateTokens())
	repo.AssertExpectations(t)
	// the stored address must never reach the node
	repo.AssertNotCalled(t, "TokenFromNode", "stored")
}

func Test_UpdateTokens_ContinuesWhenATokenFails(t *testing.T) {
	repo := mockRepo{nil, &mock.Mock{}}
	dexIndexer := dexIndexer{pkg.NetworkMetadata{}, &repo, "chainId"}

	repo.On("Pairs", db.LastIdLimitCondition{}).
		Return([]Pair{{Address: "pair0", Asset0: "uluna", Asset1: "ok1", Lp: "ok2"}}, nil).Once()
	repo.On("TokenAddresses", db.LastIdLimitCondition{}).Return([]string{}, nil).Once()
	repo.On("TokenFromNode", "uluna").Return(nil, errors.New("metadata of denom is not supported")).Once()
	repo.On("TokenFromNode", "ok1").Return(&Token{Address: "ok1"}, nil).Once()
	repo.On("TokenFromNode", "ok2").Return(&Token{Address: "ok2"}, nil).Once()
	// the addresses resolved after the failure are still persisted
	repo.On("SaveTokens", []Token{{Address: "ok1"}, {Address: "ok2"}}).Return(nil).Once()

	err := dexIndexer.UpdateTokens()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "1 token(s) failed")
	assert.Contains(t, err.Error(), "first failure on uluna")
	assert.Contains(t, err.Error(), "metadata of denom is not supported")
	repo.AssertExpectations(t)
}

func Test_UpdateTokens_QueriesAFailingAddressOncePerPass(t *testing.T) {
	repo := mockRepo{nil, &mock.Mock{}}
	dexIndexer := dexIndexer{pkg.NetworkMetadata{}, &repo, "chainId"}

	// a native denom paired against many assets must not be retried for every pair
	pairs := make([]Pair, 0, 3)
	for i := 0; i < 3; i++ {
		pairs = append(pairs, Pair{
			Address: fmt.Sprintf("pair%d", i),
			Asset0:  "uluna",
			Asset1:  fmt.Sprintf("cw20-%d", i),
			Lp:      fmt.Sprintf("lp%d", i),
		})
	}

	repo.On("Pairs", db.LastIdLimitCondition{}).Return(pairs, nil).Once()
	repo.On("TokenAddresses", db.LastIdLimitCondition{}).Return([]string{}, nil).Once()
	repo.On("TokenFromNode", "uluna").Return(nil, errors.New("boom")).Once()
	for i := 0; i < 3; i++ {
		for _, addr := range []string{fmt.Sprintf("cw20-%d", i), fmt.Sprintf("lp%d", i)} {
			repo.On("TokenFromNode", addr).Return(&Token{Address: addr}, nil).Once()
		}
	}
	repo.On("SaveTokens", mock.Anything).Return(nil).Once()

	err := dexIndexer.UpdateTokens()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "1 token(s) failed")
	// Once() on "uluna" already asserts a single call; make the intent explicit
	repo.AssertNumberOfCalls(t, "TokenFromNode", 7)
	repo.AssertExpectations(t)
}

func Test_UpdateTokens_SavesInBatches(t *testing.T) {
	repo := mockRepo{nil, &mock.Mock{}}
	dexIndexer := dexIndexer{pkg.NetworkMetadata{}, &repo, "chainId"}

	// 34 pairs * 3 addresses = 102 tokens, so one full batch plus a remainder
	const pairCount = 34
	totalAddrs := pairCount * 3
	require.Greater(t, totalAddrs, tokenSaveBatchSize)

	pairs := make([]Pair, 0, pairCount)
	for i := 0; i < pairCount; i++ {
		pairs = append(pairs, Pair{
			Address: fmt.Sprintf("pair%d", i),
			Asset0:  fmt.Sprintf("asset0-%d", i),
			Asset1:  fmt.Sprintf("asset1-%d", i),
			Lp:      fmt.Sprintf("lp-%d", i),
		})
	}

	repo.On("Pairs", db.LastIdLimitCondition{}).Return(pairs, nil).Once()
	repo.On("TokenAddresses", db.LastIdLimitCondition{}).Return([]string{}, nil).Once()
	repo.On("TokenFromNode", mock.Anything).Return(&Token{Address: "any"}, nil).Times(totalAddrs)

	savedBatchSizes := []int{}
	repo.On("SaveTokens", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		savedBatchSizes = append(savedBatchSizes, len(args.Get(0).([]Token)))
	}).Twice()

	require.NoError(t, dexIndexer.UpdateTokens())

	// a full batch is flushed mid-pass, the remainder at the end
	assert.Equal(t, []int{tokenSaveBatchSize, totalAddrs - tokenSaveBatchSize}, savedBatchSizes)
	repo.AssertExpectations(t)
}

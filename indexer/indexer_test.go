package indexer

import (
	"testing"

	"github.com/dezswap/dezswap-api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	Repo
	*mock.Mock
}

func (m *mockRepo) Tokens(cond db.LastIdLimitCondition) ([]Token, error) {
	args := m.Called(cond)
	return args.Get(0).([]Token), args.Error(1)
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
	dexIndexer := dexIndexer{&repo, "chainId"}

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

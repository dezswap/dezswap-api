package chainregistry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_VerifiedCw20s(t *testing.T) {
	assert := assert.New(t)

	endpoint := "https://raw.githubusercontent.com/cosmos/chain-registry/refs/heads/master/terra2/assetlist.json"
	c := NewClient(endpoint)
	res, err := c.VerifiedCw20s()
	assert.NotNil(res)
	assert.NoError(err)
}

func Test_VerifiedIbcs(t *testing.T) {
	assert := assert.New(t)

	endpoint := "https://raw.githubusercontent.com/cosmos/chain-registry/refs/heads/master/terra2/assetlist.json"
	c := NewClient(endpoint)
	res, err := c.VerifiedIbcs()
	assert.NotNil(res)
	assert.NoError(err)
}

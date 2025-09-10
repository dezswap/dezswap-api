package xpla

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_VerifiedCw20s(t *testing.T) {
	assert := assert.New(t)

	c := NewClient()
	res, err := c.VerifiedCw20s()
	assert.NotNil(res)
	assert.NoError(err)
}

func Test_VerifiedIbcs(t *testing.T) {
	assert := assert.New(t)

	c := NewClient()
	res, err := c.VerifiedIbcs()
	assert.NotNil(res)
	assert.NoError(err)
}

func Test_VerifiedErc20s(t *testing.T) {
	assert := assert.New(t)

	c := NewClient()
	res, err := c.VerifiedErc20s()
	assert.NotNil(res)
	assert.NoError(err)
}

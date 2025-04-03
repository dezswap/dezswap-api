package chainregistry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_VerifiedCw20s(t *testing.T) {
	assert := assert.New(t)

	c, err := NewClient("phoenix-1")
	assert.NoError(err)

	res, err := c.VerifiedCw20s()
	assert.NotNil(res)
	assert.NoError(err)
}

func Test_VerifiedIbcs(t *testing.T) {
	assert := assert.New(t)

	c, err := NewClient("columbus-5")
	assert.NoError(err)

	res, err := c.VerifiedIbcs()
	assert.NotNil(res)
	assert.NoError(err)
}

package pkg

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func _TestQueryErc20Info(t *testing.T) {
	rpcURL := ""       // e.g. http://127.0.0.1:8545
	contractAddr := "" // e.g. 69FD386467E3659F81e58b6EC7a12C64b32FB1E2

	c, err := NewEthClient(rpcURL)
	assert.NoError(t, err)

	erc20Meta, err := c.QueryErc20Info(context.Background(), contractAddr)
	assert.NoError(t, err)
	assert.NotEmpty(t, erc20Meta.Name)
	assert.NotEmpty(t, erc20Meta.Symbol)
	assert.NotZero(t, erc20Meta.Decimals)
}

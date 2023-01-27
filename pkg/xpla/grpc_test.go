package xpla

import (
	"fmt"
	"testing"

	"github.com/dezswap/dezswap-api/configs"

	"github.com/dezswap/dezswap-api/pkg/dezswap"
	"github.com/stretchr/testify/assert"
)

func Test_QueryContract(t *testing.T) {
	cf := configs.New().Indexer.SrcNode
	assert := assert.New(t)
	c, err := NewGrpcClient(fmt.Sprintf("%s:%d", cf.Host, cf.Port))
	assert.NoError(err)
	res, err := c.QueryContract(dezswap.TESTNET_FACTORY, []byte(`{"pairs": {}}`), LATEST_HEIGHT_INDICATOR)
	assert.NotNil(res)
	assert.NoError(err)
}

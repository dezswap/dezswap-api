package dezswap

import (
	"fmt"
	"testing"

	"github.com/dezswap/dezswap-api/configs"

	"github.com/dezswap/dezswap-api/pkg/xpla"
	"github.com/stretchr/testify/assert"
)

func _Test_QueryContract(t *testing.T) {
	cf := configs.New().Indexer.SrcNode
	assert := assert.New(t)
	c, err := xpla.NewGrpcClient(fmt.Sprintf("%s:%s", cf.Host, cf.Port))
	assert.NoError(err)
	res, err := c.QueryContract(TESTNET_FACTORY, []byte(`{"pairs": {}}`), xpla.LATEST_HEIGHT_INDICATOR)
	assert.NotNil(res)
	assert.NoError(err)
}

func _Test_SyncedHeight(t *testing.T) {
	cf := configs.New().Indexer.SrcNode
	assert := assert.New(t)
	c, err := xpla.NewGrpcClient(fmt.Sprintf("%s:%s", cf.Host, cf.Port))
	assert.NoError(err)
	res, err := c.SyncedHeight()
	assert.NotNil(res)
	assert.NoError(err)
}

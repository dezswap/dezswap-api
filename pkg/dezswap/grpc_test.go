package dezswap

import (
	"fmt"
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/dezswap/dezswap-api/pkg/xpla"
	"testing"

	"github.com/dezswap/dezswap-api/configs"

	"github.com/stretchr/testify/assert"
)

func _Test_QueryContract(t *testing.T) {
	cf := configs.New().Indexer.SrcNode
	assert := assert.New(t)
	c, err := pkg.NewGrpcClient(fmt.Sprintf("%s:%s", cf.Host, cf.Port))
	assert.NoError(err)
	res, err := c.QueryContract(TESTNET_FACTORY, []byte(`{"pairs": {}}`), xpla.NetworkMetadata.LatestHeightIndicator)
	assert.NotNil(res)
	assert.NoError(err)
}

func _Test_SyncedHeight(t *testing.T) {
	cf := configs.New().Indexer.SrcNode
	assert := assert.New(t)
	c, err := pkg.NewGrpcClient(fmt.Sprintf("%s:%s", cf.Host, cf.Port))
	assert.NoError(err)
	res, err := c.SyncedHeight()
	assert.NotNil(res)
	assert.NoError(err)
}

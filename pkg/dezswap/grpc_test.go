package dezswap

import (
	"fmt"
	"github.com/dezswap/dezswap-api/pkg"
	"testing"

	"github.com/dezswap/dezswap-api/configs"

	"github.com/stretchr/testify/assert"
)

func _Test_QueryContract(t *testing.T) {
	cf := configs.New().Indexer
	snc := configs.New().Indexer.SrcNode
	networkMetadata, _ := pkg.GetNetworkMetadata(cf.ChainId)

	assert := assert.New(t)
	c, err := pkg.NewGrpcClient(fmt.Sprintf("%s:%s", snc.Host, snc.Port))
	assert.NoError(err)
	res, err := c.QueryContract(TESTNET_FACTORY, []byte(`{"pairs": {}}`), networkMetadata.LatestHeightIndicator)
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

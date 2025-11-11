package dezswap

import (
	"fmt"
	"testing"

	"github.com/dezswap/dezswap-api/pkg"

	"github.com/dezswap/dezswap-api/configs"

	"github.com/stretchr/testify/assert"
)

// Remove _ prefix to test directly with source node
func _Test_QueryContract(t *testing.T) {
	cf := configs.New().Indexer
	snc := configs.New().Indexer.SrcNode
	networkMetadata, _ := pkg.GetNetworkMetadata(cf.ChainId)

	assert := assert.New(t)
	c, err := pkg.NewGrpcClient(fmt.Sprintf("%s:%s", snc.Host, snc.Port), false)
	assert.NoError(err)
	res, err := c.QueryContract("<<FACTORY_ADDR>>", []byte(`{"pairs": {}}`), networkMetadata.LatestHeightIndicator)
	assert.NotNil(res)
	assert.NoError(err)
}

func _Test_SyncedHeight(t *testing.T) {
	cf := configs.New().Indexer.SrcNode
	assert := assert.New(t)
	c, err := pkg.NewGrpcClient(fmt.Sprintf("%s:%s", cf.Host, cf.Port), false)
	assert.NoError(err)
	res, err := c.SyncedHeight()
	assert.NotNil(res)
	assert.NoError(err)
}

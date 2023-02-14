package repo

import (
	"strconv"
	"testing"

	"github.com/dezswap/dezswap-api/configs"
	"github.com/dezswap/dezswap-api/pkg/db"
	"github.com/stretchr/testify/assert"
)

func Test_Pairs(t *testing.T) {
	c := configs.New()
	r := New(c.Indexer.ChainId, c.Indexer.SrcDb)
	pairs, err := r.Pairs(db.LastIdLimitCondition{})
	assert.NoError(t, err)
	assert.NotEmpty(t, pairs)
	lastIdx := len(pairs) - 1
	if lastIdx < 0 {
		lastIdx = 0
	}

	lastId, _ := strconv.Atoi(pairs[len(pairs)-1].ID)
	pairs, err = r.Pairs(db.LastIdLimitCondition{
		LastId: uint(lastId),
	})
	assert.NoError(t, err)
	assert.Empty(t, pairs)
}

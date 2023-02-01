package repo

import (
	"encoding/json"

	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg/dezswap"
	"github.com/pkg/errors"
)

type nodeMapper interface {
	resToToken(data []byte) (*indexer.Token, error)
	resToPoolInfo(addr string, height uint64, data []byte) (*indexer.PoolInfo, error)
}

var _ nodeMapper = &nodeMapperImpl{}

type nodeMapperImpl struct{}

// resToPoolInfo implements nodeMapper
func (*nodeMapperImpl) resToPoolInfo(addr string, height uint64, data []byte) (*indexer.PoolInfo, error) {
	res := dezswap.PoolRes{}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, errors.Wrap(err, "nodeMapperImpl.resToPoolInfo")
	}
	return &indexer.PoolInfo{
		Height:       height,
		Address:      addr,
		Asset0Amount: res.AssetInfos[0].Amount,
		Asset1Amount: res.AssetInfos[1].Amount,
		LpAmount:     res.TotalShare,
	}, nil
}

// resToToken implements nodeMapper
func (*nodeMapperImpl) resToToken(data []byte) (*indexer.Token, error) {
	panic("unimplemented")
}

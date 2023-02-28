package repo

import (
	"encoding/json"

	ibc_types "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg/dezswap"

	"github.com/pkg/errors"
)

type nodeMapper interface {
	resToToken(data []byte) (*indexer.Token, error)
	resToPoolInfo(addr string, height uint64, data []byte) (*indexer.PoolInfo, error)

	denomTraceToToken(*ibc_types.DenomTrace) (*indexer.Token, error)
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
	res := dezswap.TokenInfoRes{}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, errors.Wrap(err, "nodeMapperImpl.resToToken")
	}
	return &indexer.Token{
		Name:     res.Name,
		Symbol:   res.Symbol,
		Decimals: uint8(res.Decimals),
	}, nil
}

// denomTraceToToken implements nodeMapper
func (*nodeMapperImpl) denomTraceToToken(trace *ibc_types.DenomTrace) (*indexer.Token, error) {
	return &indexer.Token{
		Name:   trace.BaseDenom,
		Symbol: trace.BaseDenom,

		Decimals: 18,
	}, nil
}

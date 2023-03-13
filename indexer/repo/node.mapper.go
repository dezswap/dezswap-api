package repo

import (
	"encoding/json"

	ibc_types "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg/dezswap"

	"github.com/pkg/errors"
)

type nodeMapper interface {
	resToToken(addr, chainId string, data []byte) (*indexer.Token, error)
	resToPoolInfo(addr, chainId string, height uint64, data []byte) (*indexer.PoolInfo, error)

	denomTraceToToken(addr, chainId string, trace *ibc_types.DenomTrace) (*indexer.Token, error)
}

var _ nodeMapper = &nodeMapperImpl{}

type nodeMapperImpl struct{}

// resToPoolInfo implements nodeMapper
func (*nodeMapperImpl) resToPoolInfo(addr, chainId string, height uint64, data []byte) (*indexer.PoolInfo, error) {
	res := dezswap.PoolRes{}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, errors.Wrap(err, "nodeMapperImpl.resToPoolInfo")
	}
	return &indexer.PoolInfo{
		ChainId:      chainId,
		Height:       height,
		Address:      addr,
		Asset0:       res.GetAsset(0),
		Asset0Amount: res.Assets[0].Amount,
		Asset1:       res.GetAsset(1),
		Asset1Amount: res.Assets[1].Amount,
		LpAmount:     res.TotalShare,
	}, nil
}

// resToToken implements nodeMapper
func (*nodeMapperImpl) resToToken(addr, chainId string, data []byte) (*indexer.Token, error) {
	res := dezswap.TokenInfoRes{}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, errors.Wrap(err, "nodeMapperImpl.resToToken")
	}
	return &indexer.Token{
		Name:     res.Name,
		Address:  addr,
		ChainId:  chainId,
		Symbol:   res.Symbol,
		Decimals: uint8(res.Decimals),
	}, nil
}

// denomTraceToToken implements nodeMapper
func (*nodeMapperImpl) denomTraceToToken(addr, chainId string, trace *ibc_types.DenomTrace) (*indexer.Token, error) {
	return &indexer.Token{
		Name:    trace.BaseDenom,
		Symbol:  trace.BaseDenom,
		Address: addr,
		ChainId: chainId,

		Decimals: 18,
	}, nil
}

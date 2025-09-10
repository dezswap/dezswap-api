package pkg

//go:generate abigen --abi=erc20/ERC20.abi --pkg=erc20 --out=erc20/erc20.go

import (
	"context"
	"github.com/dezswap/dezswap-api/pkg/erc20"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

type ERC20Meta struct {
	Name     string
	Symbol   string
	Decimals uint8
}

type EthClient interface {
	QueryErc20Info(ctx context.Context, contractAddr string) (ERC20Meta, error)
}

type ethClientImpl struct {
	rpcURL string
}

var _ EthClient = &ethClientImpl{}

func NewEthClient(rpcURL string) (EthClient, error) {
	return &ethClientImpl{rpcURL}, nil
}

// QueryErc20Info connects to the given EVM RPC and returns ERC-20 metadata for the contract.
func (c *ethClientImpl) QueryErc20Info(ctx context.Context, contractAddr string) (ERC20Meta, error) {
	cli, err := ethclient.DialContext(ctx, c.rpcURL)
	if err != nil {
		return ERC20Meta{}, errors.Wrapf(err, "QueryErc20Info: failed to dial EVM RPC: %s", c.rpcURL)
	}
	defer cli.Close()

	client, err := erc20.NewErc20(common.HexToAddress(contractAddr), cli)
	if err != nil {
		return ERC20Meta{}, errors.Wrapf(err, "QueryErc20Info: failed to init ERC20 binding (contract=%s)", contractAddr)
	}

	name, err := client.Name(&bind.CallOpts{Context: ctx})
	if err != nil {
		return ERC20Meta{}, errors.Wrapf(err, "QueryErc20Info: erc20.Name() call failed (contract=%s)", contractAddr)
	}

	symbol, err := client.Symbol(&bind.CallOpts{Context: ctx})
	if err != nil {
		return ERC20Meta{}, errors.Wrapf(err, "QueryErc20Info: erc20.Symbol() call failed (contract=%s)", contractAddr)
	}

	decimals, err := client.Decimals(&bind.CallOpts{Context: ctx})
	if err != nil {
		return ERC20Meta{}, errors.Wrapf(err, "QueryErc20Info: erc20.Decimals() call failed (contract=%s)", contractAddr)
	}

	return ERC20Meta{Name: name, Symbol: symbol, Decimals: decimals}, nil
}
